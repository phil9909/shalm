package shalm

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/blang/semver"
	"gopkg.in/yaml.v2"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type helmChart struct {
	APIVersion  string   `json:"apiVersion,omitempty"`
	Name        string   `json:"name,omitempty"`
	Version     string   `json:"version,omitempty"`
	Description string   `json:"description,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
	Home        string   `json:"home,omitempty"`
	Sources     []string `json:"sources,omitempty"`
	Icon        string   `json:"icon,omitempty"`
}

func (c *chartImpl) loadChartYaml() error {
	var helmChart helmChart

	err := c.loadYamlFile(c.path("Chart.yaml"), &helmChart)
	if err != nil {
		return err
	}
	if strings.HasPrefix(helmChart.Version, "v") {
		c.Version, err = semver.Parse(helmChart.Version[1:])
		if err != nil {
			return errors.Wrap(err, "Invalid version in helm chart")
		}
	} else {
		c.Version, err = semver.Parse(helmChart.Version)
		if err != nil {
			return errors.Wrap(err, "Invalid version in helm chart")
		}
	}
	c.Name = helmChart.Name
	return nil
}

func (c *chartImpl) loadValuesYaml() error {
	var values map[string]interface{}
	err := c.loadYamlFile(c.path("values.yaml"), &values)
	if err != nil {
		return err
	}
	for k, v := range values {
		c.values[k] = toStarlark(v)
	}
	return nil
}

func (c *chartImpl) init(thread *starlark.Thread, repo Repo, args starlark.Tuple, kwargs []starlark.Tuple) error {
	c.methods["apply"] = c.applyFunction()
	c.methods["delete"] = c.deleteFunction()
	c.methods["__apply"] = c.applyLocalFunction()
	c.methods["__delete"] = c.deleteLocalFunction()

	file := c.path("Chart.star")
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil
	}
	globals, err := starlark.ExecFile(thread, file, nil, starlark.StringDict{
		"chart": starlark.NewBuiltin("chart", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("%s: got %d arguments, want at most %d", "chart", 0, 1)
			}
			url := args[0].(starlark.String).GoString()
			if !filepath.IsAbs(url) {
				url = path.Join(c.dir, url)
			}
			namespace := c.namespace
			parser := &kwargsParser{kwargs: kwargs}
			parser.Arg("namespace", func(value starlark.Value) {
				namespace = value.(starlark.String).GoString()
			})
			kwargs = parser.Parse()
			return repo.Get(thread, url, namespace, args[1:], kwargs)
		}),
		"user_credential": starlark.NewBuiltin("user_credential", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
			s := &userCredential{}
			s.setDefaultKeys()
			if err := starlark.UnpackArgs("user_credential", args, kwargs, "name", &s.name,
				"username_key?", &s.usernameKey, "password_key?", &s.passwordKey,
				"username?", &s.username, "password?", &s.password); err != nil {
				return nil, err
			}
			c.userCredentials = append(c.userCredentials, s)
			return s, nil
		}),
		"struct": starlark.NewBuiltin("struct", starlarkstruct.Make),
		"k8s": starlark.NewBuiltin("k8s", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
			var kubeconfig string
			if err := starlark.UnpackArgs("k8s", args, kwargs, "kubeconfig", kubeconfig); err != nil {
				return nil, err
			}
			return &k8sValueImpl{&k8sImpl{kubeconfig: kubeconfig, namespace: c.namespace}}, nil
		}),
	})
	if err != nil {
		return err
	}

	init, ok := globals["init"]

	if ok {
		_, err := starlark.Call(thread, init, append([]starlark.Value{c}, args...), kwargs)
		if err != nil {
			return err
		}
	}

	for k, v := range globals {
		if k == "init" {
			continue
		}
		f, ok := v.(starlark.Callable)
		if ok {
			c.methods[k] = starlark.NewBuiltin(f.Name(), func(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
				allArgs := make([]starlark.Value, args.Len()+1)
				allArgs[0] = c
				for i := 0; i < args.Len(); i++ {
					allArgs[i+1] = args.Index(i)
				}
				return f.CallInternal(thread, allArgs, kwargs)
			})
		}
	}
	c.methods["apply"] = wrapNamespace(c.methods["apply"], c.namespace)
	c.methods["delete"] = wrapNamespace(c.methods["delete"], c.namespace)

	return nil
}

func wrapNamespace(callable starlark.Callable, namespace string) starlark.Callable {
	return starlark.NewBuiltin(callable.Name(), func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("Missing first argument k8s")
		}
		k, ok := args[0].(K8sValue)
		if !ok {
			return nil, fmt.Errorf("Invalid first argument to %s", callable.Name())
		}
		args[0] = &k8sValueImpl{k.ForNamespace(namespace)}
		return callable.CallInternal(thread, args, kwargs)
	})
}

func (c *chartImpl) loadYamlFile(filename string, value interface{}) error {
	reader, err := os.Open(filename) // For read access.
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("Unable to open file %s for parsing: %s", filename, err.Error())
	}
	defer reader.Close()
	decoder := yaml.NewDecoder(reader)
	err = decoder.Decode(value)
	if err != nil {
		return fmt.Errorf("Error during parsing file %s: %s", filename, err.Error())
	}
	return nil
}
