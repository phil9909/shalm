package impl

import (
	"fmt"
	"os"

	"github.com/blang/semver"
	"github.com/kramerul/shalm/internal/pkg/chart/api"
	yaml "gopkg.in/yaml.v2"

	"go.starlark.net/starlark"
)

func (c *chartImpl) loadChartYaml() error {
	var helmChart api.HelmChart

	err := c.loadYamlFile(c.path("Chart.yaml"), &helmChart)
	if err != nil {
		return err
	}
	c.Version = semver.MustParse(helmChart.Version)
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

func (c *chartImpl) init(thread *starlark.Thread, repo api.Repo, args starlark.Tuple, kwargs []starlark.Tuple) error {
	c.methods["apply"] = c.applyFunction()
	c.methods["delete"] = c.deleteFunction()
	c.methods["__apply"] = c.applyLocalFunction()
	c.methods["__delete"] = c.deleteLocalFunction()

	file := c.path("Chart.star")
	if _, err := c.fs.Stat(file); os.IsNotExist(err) {
		return nil
	}
	globals, err := starlark.ExecFile(thread, file, nil, starlark.StringDict{
		"chart": starlark.NewBuiltin("chart", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("%s: got %d arguments, want at most %d", "chart", 0, 1)
			}
			return repo.Get(thread, c, args[0].(starlark.String).GoString(), args[1:], kwargs)
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
		k, ok := args[0].(api.K8sValue)
		if !ok {
			return nil, fmt.Errorf("Invalid first argument to %s", callable.Name())
		}
		args[0] = &k8sValueImpl{k.ForNamespace(namespace)}
		return callable.CallInternal(thread, args, kwargs)
	})
}

func (c *chartImpl) loadYamlFile(filename string, value interface{}) error {
	reader, err := c.fs.Open(filename) // For read access.
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
