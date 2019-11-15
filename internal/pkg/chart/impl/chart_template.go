package impl

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/kramerul/shalm/internal/pkg/chart/api"
	"go.starlark.net/starlark"
)

type Release struct {
	Name      string
	Namespace string
	Service   string
}

func (c *chartImpl) Template(thread *starlark.Thread, installOpts *api.InstallOpts) (string, error) {
	t, err := starlark.Call(thread, c.templateFunction(), starlark.Tuple{NewInstallOptsValue(installOpts)}, nil)
	if err != nil {
		return "", err
	}
	return t.(starlark.String).GoString(), nil
}

func (c *chartImpl) templateFunction() starlark.Callable {
	return starlark.NewBuiltin("template", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var installOpts *installOptsValue
		if err := starlark.UnpackArgs("template", args, kwargs, "installOpts", &installOpts); err != nil {
			return nil, err
		}
		var writer bytes.Buffer
		err := c.templateRecursive(thread, installOpts, &writer)
		if err != nil {
			return starlark.None, err
		}
		return starlark.String(writer.String()), nil
	})
}

func (c *chartImpl) templateRecursive(thread *starlark.Thread, installOpts *installOptsValue, writer io.Writer) error {
	err := c.eachSubChart(func(subChart *chartImpl) error {
		return subChart.templateRecursive(thread, installOpts, writer)
	})
	if err != nil {
		return err
	}
	return c.template(thread, installOpts, writer)
}

func (c *chartImpl) template(thread *starlark.Thread, installOpts *installOptsValue, writer io.Writer) error {
	glob := c.path("templates", "*.yaml")
	filenames, err := filepath.Glob(glob)
	if err != nil {
		return err
	}
	if len(filenames) == 0 {
		return nil
	}
	helpers := c.path("templates", "_helpers.tpl")
	if _, err := os.Stat(helpers); err != nil {
		helpers = ""
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
		tpl := template.New(filepath.Base(filename))
		tpl = addTemplateFuncs(tpl)
		if helpers == "" {
			tpl, err = tpl.ParseFiles(filename)
		} else {
			tpl, err = tpl.ParseFiles(helpers, filename)
		}

		if err != nil {
			return err
		}
		err = tpl.Execute(&buffer, struct {
			Values  interface{}
			Methods map[string]interface{}
			Chart   *chartImpl
			Release Release
			Files   files
		}{
			Values:  values,
			Methods: methods,
			Chart:   c,
			Release: Release{Name: c.Name, Namespace: installOpts.Namespace, Service: c.Name},
			Files:   files(make(map[string][]byte)),
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
