package impl

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

	yaml "gopkg.in/yaml.v2"

	"github.com/Masterminds/sprig/v3"
	"github.com/spf13/afero"
)

type HelmTemplater struct {
	helpers string
	dir     string
	fs      afero.Fs
}

type files struct {
	dir string
	fs  afero.Fs
}

// NewHelmTemplater -
func NewHelmTemplater(fs afero.Fs, dir string) (*HelmTemplater, error) {
	h := &HelmTemplater{
		dir: dir,
		fs:  fs,
	}
	content, err := afero.ReadFile(h.fs, path.Join(dir, "templates", "_helpers.tpl"))
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

func (h *HelmTemplater) Template(externGlob string, value interface{}, writer io.Writer) error {
	var glob string
	if externGlob != "" {
		glob = path.Join(h.dir, "templates", externGlob)
	} else {
		glob = path.Join(h.dir, "templates", "*.yaml")
	}
	filenames, err := filepath.Glob(glob)
	if err != nil {
		return err
	}
	if len(filenames) == 0 {
		return nil
	}

	for _, filename := range filenames {
		var buffer bytes.Buffer
		tpl, err := h.template(filepath.Base(filename))
		if err != nil {
			return err
		}
		content, err := afero.ReadFile(h.fs, filename)
		if err != nil {
			return err
		}
		tpl, err = tpl.Parse(string(content))
		if err != nil {
			return err
		}
		err = tpl.Execute(&buffer, value)
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
	matches, err := afero.Glob(f.fs, path.Join(f.dir, pattern))
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
	data, err := afero.ReadFile(f.fs, path.Join(f.dir, name))
	if err != nil {
		return err.Error()
	}
	return string(data)
}
