package impl

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"

	"go.starlark.net/starlark"
)

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
		err := c.templateRecursive(thread, &writer)
		if err != nil {
			return starlark.None, err
		}
		return starlark.String(writer.String()), nil
	})
}

func (c *chartImpl) templateRecursive(thread *starlark.Thread, writer io.Writer) error {
	err := c.eachSubChart(func(subChart *chartImpl) error {
		return subChart.templateRecursive(thread, writer)
	})
	if err != nil {
		return err
	}
	return c.template(thread, "", writer)
}

func (c *chartImpl) template(thread *starlark.Thread, externGlob string, writer io.Writer) error {
	helmTemplater, err := NewHelmTemplater(c.path())
	if err != nil {
		return err
	}
	var glob string
	if externGlob != "" {
		glob = c.path("templates", externGlob)
	} else {
		glob = c.path("templates", "*.yaml")
	}
	filenames, err := filepath.Glob(glob)
	if err != nil {
		return err
	}
	if len(filenames) == 0 {
		return nil
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
	for _, filename := range filenames {
		var buffer bytes.Buffer
		tpl, err := helmTemplater.template(filepath.Base(filename))
		if err != nil {
			return err
		}
		tpl, err = tpl.ParseFiles(filename)
		if err != nil {
			return err
		}
		err = tpl.Execute(&buffer, struct {
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
			Files:   files{c.dir},
		})
		if err != nil {
			return err
		}
		if buffer.Len() > 0 {
			content := strings.TrimSpace(buffer.String())
			if len(content) > 0 {
				if !strings.HasPrefix(content, "---") {
					writer.Write([]byte("---\n"))
				}
				writer.Write([]byte(content))
				writer.Write([]byte("\n"))
			}
		}

	}
	return nil
}
