package renderer

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"

	"github.com/Masterminds/sprig/v3"
)

type helmRenderer struct {
	helpers string
}

// HelmFileRenderer -
func HelmFileRenderer(dir string, value interface{}) (func(filename string, writer io.Writer) error, error) {
	h, err := newHelmRenderer(dir)
	if err != nil {
		return nil, err
	}
	return h.fileTemplater(value), nil
}
func newHelmRenderer(dir string) (*helmRenderer, error) {
	h := &helmRenderer{}
	content, err := ioutil.ReadFile(path.Join(dir, "templates", "_helpers.tpl"))
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		h.helpers = ""
	} else {
		h.helpers = string(content)
	}
	return h, nil
}

func (h *helmRenderer) fileTemplater(value interface{}) func(filename string, writer io.Writer) error {
	return func(filename string, writer io.Writer) error {
		tpl, err := h.loadTemplate(filepath.Base(filename))
		if err != nil {
			return err
		}
		content, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		tpl, err = tpl.Parse(string(content))
		if err != nil {
			return err
		}
		return tpl.Execute(writer, value)
	}

}

func notImplemented(_ interface{}) string {
	panic("not implemented")
}

func (h *helmRenderer) loadTemplate(name string) (result *template.Template, err error) {
	result = template.New(name)
	result = result.Funcs(sprig.TxtFuncMap())
	result = result.Funcs(map[string]interface{}{
		"toToml":   notImplemented,
		"toYaml":   toYAML,
		"fromYaml": notImplemented,
		"toJson":   toJSON,
		"fromJson": notImplemented,
		"tpl":      h.tpl(),
		"required": notImplemented,
	})
	incResult := result
	result = result.Funcs(map[string]interface{}{
		"include": func(name string, data interface{}) (string, error) {
			var buf strings.Builder
			err := incResult.ExecuteTemplate(&buf, name, data)
			return buf.String(), err
		},
	})
	if h.helpers != "" {
		result, err = result.Parse(h.helpers)
		if err != nil {
			return
		}
	}
	return
}

func (h *helmRenderer) tpl() func(stringTemplate string, values interface{}) interface{} {
	return func(stringTemplate string, values interface{}) interface{} {
		tpl, err := h.loadTemplate("internal template")
		if err != nil {
			return err.Error()
		}
		tpl, err = tpl.Parse(stringTemplate)
		if err != nil {
			return err.Error()
		}
		var buffer bytes.Buffer
		err = tpl.Execute(&buffer, values)
		if err != nil {
			return err.Error()
		}
		return buffer.String()
	}
}

func toYAML(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}

func toJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}
