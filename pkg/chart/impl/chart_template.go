package impl

import (
	"bytes"
	"io"

	"go.starlark.net/starlark"
	"gopkg.in/yaml.v2"
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

func (c *chartImpl) Template(thread *starlark.Thread) (string, error) {
	t, err := starlark.Call(thread, c.templateFunction(), nil, nil)
	if err != nil {
		return "", err
	}
	return t.(starlark.String).GoString(), nil
}

func (c *chartImpl) templateFunction() starlark.Callable {
	return starlark.NewBuiltin("template", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		if err := starlark.UnpackArgs("template", args, kwargs); err != nil {
			return nil, err
		}
		var writer bytes.Buffer
		err := c.templateRecursive(thread, &writer, &HelmOptions{})
		if err != nil {
			return starlark.None, err
		}
		return starlark.String(writer.String()), nil
	})
}

func (c *chartImpl) templateRecursive(thread *starlark.Thread, writer io.Writer, options *HelmOptions) error {
	err := c.eachSubChart(func(subChart *chartImpl) error {
		return subChart.templateRecursive(thread, writer, options)
	})
	if err != nil {
		return err
	}
	return c.template(thread, writer, options)
}

func (c *chartImpl) template(thread *starlark.Thread, writer io.Writer, options *HelmOptions) error {
	h, err := NewHelmTemplater(c.path(), c.namespace)
	if err != nil {
		return err
	}
	values := toGo(c).(map[string]interface{})
	methods := make(map[string]interface{})
	for k, f := range c.methods {
		method := f
		methods[k] = func() (interface{}, error) {
			value, err := method.CallInternal(thread, nil, nil)
			return toGo(value), err
		}
	}
	err = h.Template(struct {
		Values  interface{}
		Methods map[string]interface{}
		Chart   Chart
		Release Release
		Files   files
	}{
		Values:  values,
		Methods: methods,
		Chart: Chart{
			Name:       c.Name,
			AppVersion: c.Version.String(),
			Version:    c.Version.String(),
		},
		Release: Release{Name: c.Name, Namespace: c.namespace, Service: c.Name},
		Files:   files{dir: c.dir},
	}, writer, options)
	if err != nil {
		return err
	}
	if len(c.userCredentials) == 0 {
		return nil
	}
	writer.Write([]byte("---\n"))
	enc := yaml.NewEncoder(writer)
	for _, credential := range c.userCredentials {
		err = enc.Encode(credential.secret(c.namespace))
		if err != nil {
			return err
		}
	}

	return nil
}
