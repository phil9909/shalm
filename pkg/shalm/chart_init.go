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

func (c *chartImpl) loadChartYaml() error {

	err := c.loadYamlFile(c.path("Chart.yaml"), &c.clazz)
	if err != nil {
		return err
	}
	if strings.HasPrefix(c.clazz.Version, "v") {
		c.Version, err = semver.Parse(c.clazz.Version[1:])
		if err != nil {
			return errors.Wrap(err, "Invalid version in helm chart")
		}
	} else {
		c.Version, err = semver.Parse(c.clazz.Version)
		if err != nil {
			return errors.Wrap(err, "Invalid version in helm chart")
		}
	}
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

	internal := starlark.StringDict{
		"chart": starlark.NewBuiltin("chart", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
			if len(args) == 0 {
				return starlark.None, fmt.Errorf("%s: got %d arguments, want at most %d", "chart", 0, 1)
			}
			url := args[0].(starlark.String).GoString()
			if !(filepath.IsAbs(url) || strings.HasPrefix(url, "http")) {
				url = path.Join(c.dir, url)
			}
			co := ChartOptions{namespace: c.namespace, suffix: c.suffix}
			parser := &kwargsParser{kwargs: kwargs}
			parser.Arg("namespace", func(value starlark.Value) {
				co.namespace = value.(starlark.String).GoString()
			})
			parser.Arg("proxy", func(value starlark.Value) {
				co.proxy = bool(value.(starlark.Bool))
			})
			parser.Arg("suffix", func(value starlark.Value) {
				co.suffix = value.(starlark.String).GoString()
			})
			co.kwargs = parser.Parse()
			return repo.Get(thread, url, co.Options())
		}),
		"user_credential": starlark.NewBuiltin("user_credential", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
			s, err := makeUserCredential(thread, fn, args, kwargs)
			if err != nil {
				return s, err
			}
			c.userCredentials = append(c.userCredentials, s.(*userCredential))
			return s, nil
		}),
		"struct": starlark.NewBuiltin("struct", starlarkstruct.Make),
		"k8s": starlark.NewBuiltin("k8s", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
			return makeK8sValue(thread, fn, args, kwargs, c.namespace)
		}),
	}
	globals, err := starlark.ExecFile(thread, file, nil, internal)
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
