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

	err := loadYamlFile(c.path("Chart.yaml"), &helmChart)
	if err != nil {
		return err
	}
	c.Version = semver.MustParse(helmChart.Version)
	c.Name = helmChart.Name
	return nil
}

func (c *chartImpl) loadValuesYaml() error {
	var values map[string]interface{}
	err := loadYamlFile(c.path("values.yaml"), &values)
	if err != nil {
		return err
	}
	for k, v := range values {
		c.values[k] = toStarlark(v)
	}
	return nil
}

func (c *chartImpl) init(thread *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) error {
	file := c.path("Chart.star")
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil
	}
	globals, err := starlark.ExecFile(thread, file, nil, starlark.StringDict{
		"chart": starlark.NewBuiltin("chart", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("%s: got %d arguments, want at most %d", "chart", 0, 1)
			}
			return c.repo.Get(thread, args[0].(starlark.String).GoString(), args[1:], kwargs)
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
	_, ok = c.methods["apply"]
	if !ok {
		c.methods["apply"] = c.applyFunction()
	}
	_, ok = c.methods["delete"]
	if !ok {
		c.methods["delete"] = c.deleteFunction()
	}
	c.methods["__apply"] = c.applyLocalFunction()
	c.methods["__delete"] = c.deleteLocalFunction()
	return nil
}

func loadYamlFile(filename string, value interface{}) error {
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