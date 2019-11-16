package impl

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	yaml "gopkg.in/yaml.v2"

	"github.com/Masterminds/sprig/v3"
)

type HelmTemplater struct {
	helpers string
	dir     string
}

type files struct {
	dir string
}

// NewHelmTemplater -
func NewHelmTemplater(dir string) (*HelmTemplater, error) {
	h := &HelmTemplater{
		dir: dir,
	}
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

func (h *HelmTemplater) template(name string) (result *template.Template, err error) {
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

func (h *HelmTemplater) tpl() func(stringTemplate string, values interface{}) interface{} {
	return func(stringTemplate string, values interface{}) interface{} {
		tpl, err := h.template("internal template")
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

func (f files) Glob(pattern string) map[string][]byte {
	result := make(map[string][]byte)
	matches, err := filepath.Glob(path.Join(f.dir, pattern))
	if err != nil {
		return result
	}
	for _, match := range matches {
		data, err := ioutil.ReadFile(match)
		if err == nil {
			p, err := filepath.Rel(f.dir, match)
			if err == nil {
				result[p] = data
			} else {
			}
		} else {
		}
	}
	return result
}

func (f files) Get(name string) string {
	data, err := ioutil.ReadFile(path.Join(f.dir, name))
	if err != nil {
		return err.Error()
	}
	return string(data)
}
