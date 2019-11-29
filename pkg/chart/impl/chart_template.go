package impl

import (
	"bytes"
	"io"

	"go.starlark.net/starlark"
)

// Release -
type Release struct {
	Name      string
	Namespace string
	Service   string
}

// Chart -
type Chart struct {
	Name       string
	Version    string
	AppVersion string
	APIVersion string
}

func (c *main.chartImpl) Template(thread *starlark.Thread) (string, error) {
	t, err := starlark.Call(thread, c.templateFunction(), nil, nil)
	if err != nil {
		return "", err
	}
	return t.(starlark.String).GoString(), nil
}

func (c *main.chartImpl) templateFunction() starlark.Callable {
	return starlark.NewBuiltin("template", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		if err := starlark.UnpackArgs("template", args, kwargs); err != nil {
			return nil, err
		}
		var writer bytes.Buffer
		err := c.templateRecursive(thread, &writer, &main.HelmOptions{})
		if err != nil {
			return starlark.None, err
		}
		return starlark.String(writer.String()), nil
	})
}

func (c *main.chartImpl) templateRecursive(thread *starlark.Thread, writer io.Writer, options *main.HelmOptions) error {
	err := c.eachSubChart(func(subChart *main.chartImpl) error {
		return subChart.templateRecursive(thread, writer, options)
	})
	if err != nil {
		return err
	}
	return c.template(thread, writer, options)
}

func (c *main.chartImpl) template(thread *starlark.Thread, writer io.Writer, options *main.HelmOptions) error {
	h, err := main.NewHelmTemplater(c.path(), c.namespace)
	if err != nil {
		return err
	}
	values := main.toGo(c).(map[string]interface{})
	methods := make(map[string]interface{})
	for k, f := range c.methods {
		method := f
		methods[k] = func() (interface{}, error) {
			value, err := method.CallInternal(thread, nil, nil)
			return main.toGo(value), err
		}
	}
	return h.Template(struct {
		Values  interface{}
		Methods map[string]interface{}
		Chart   Chart
		Release Release
		Files   main.files
	}{
		Values:  values,
		Methods: methods,
		Chart: Chart{
			Name:       c.Name,
			AppVersion: c.Version.String(),
			Version:    c.Version.String(),
		},
		Release: Release{Name: c.Name, Namespace: c.namespace, Service: c.Name},
		Files:   main.files{dir: c.dir},
	}, writer, options)
}
