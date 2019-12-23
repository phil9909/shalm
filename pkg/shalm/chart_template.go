package shalm

import (
	"bytes"
	"io"
	"path"

	"github.com/kramerul/shalm/pkg/shalm/renderer"

	"go.starlark.net/starlark"
	"gopkg.in/yaml.v2"
)

type release struct {
	Name      string
	Namespace string
	Service   string
	Revision  int
	IsUpgrade bool
	IsInstall bool
}

type chart struct {
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
		err := c.templateRecursive(thread, &writer, &renderer.Options{})
		if err != nil {
			return starlark.None, err
		}
		return starlark.String(writer.String()), nil
	})
}

func (c *chartImpl) templateRecursive(thread *starlark.Thread, writer io.Writer, options *renderer.Options) error {
	err := c.eachSubChart(func(subChart *chartImpl) error {
		return subChart.templateRecursive(thread, writer, options)
	})
	if err != nil {
		return err
	}
	return c.template(thread, writer, options)
}

func (c *chartImpl) templateValues() map[string]interface{} {
	d := make(map[string]interface{})

	for k, v := range c.values {
		value := toGo(v)
		if value != nil {
			d[k] = value
		}
	}
	return d
}

func (c *chartImpl) template(thread *starlark.Thread, writer io.Writer, options *renderer.Options) error {
	values := c.templateValues()
	methods := make(map[string]interface{})
	for k, f := range c.methods {
		method := f
		methods[k] = func() (interface{}, error) {
			value, err := method.CallInternal(thread, nil, nil)
			return toGo(value), err
		}
	}
	helmFileRenderer, err := renderer.HelmFileRenderer(c.path(), struct {
		Values  interface{}
		Methods map[string]interface{}
		Chart   chart
		Release release
		Files   renderer.Files
	}{
		Values:  values,
		Methods: methods,
		Chart: chart{
			Name:       c.Name,
			AppVersion: c.Version.String(),
			Version:    c.Version.String(),
		},
		Release: release{
			Name:      c.Name,
			Namespace: c.namespace,
			Service:   c.Name,
			Revision:  1,
			IsInstall: false,
			IsUpgrade: true,
		},
		Files: renderer.Files{Dir: c.dir},
	})
	if err != nil {
		return err
	}

	err = renderer.DirRender(c.namespace, writer, options,
		renderer.DirSpec{
			Dir:          path.Join(c.dir, "templates"),
			FileRenderer: helmFileRenderer,
		},
		renderer.DirSpec{
			Dir:          path.Join(c.dir, "ytt"),
			FileRenderer: renderer.YttFileRenderer(c),
		})

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
