package chart

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"go.starlark.net/starlark"
)

// TemplateFunction -
func (c *Chart) TemplateFunction() starlark.Callable {
	return starlark.NewBuiltin("template", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var release *Release
		if err := starlark.UnpackArgs("template", args, kwargs, "release", &release); err != nil {
			return nil, err
		}
		var writer bytes.Buffer
		err := c.templateRecursive(thread, release, &writer)
		if err != nil {
			return starlark.None, err
		}
		return starlark.String(writer.String()), nil
	})
}

func (c *Chart) templateRecursive(thread *starlark.Thread, release *Release, writer io.Writer) error {
	err := c.eachSubChart(func(subChart *Chart) error {
		return subChart.templateRecursive(thread, release, writer)
	})
	if err != nil {
		return err
	}
	return c.template(thread, release, writer)
}

func (c *Chart) template(thread *starlark.Thread, release *Release, writer io.Writer) error {
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
			Chart   *Chart
			Release *Release
			Files   files
		}{
			Values:  values,
			Methods: methods,
			Chart:   c,
			Release: release,
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
