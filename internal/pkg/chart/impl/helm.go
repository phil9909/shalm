package impl

import (
	"bytes"
	"strings"
	"text/template"

	yaml "gopkg.in/yaml.v2"

	"github.com/Masterminds/sprig/v3"
)

type files map[string][]byte

func addTemplateFuncs(tpl *template.Template) *template.Template {
	return tpl.
		Funcs(sprig.TxtFuncMap()).
		Funcs(map[string]interface{}{
			"toToml":   notImplemented,
			"toYaml":   toYAML,
			"fromYaml": notImplemented,
			"toJson":   notImplemented,
			"fromJson": notImplemented,
			"include": func(name string, data interface{}) (string, error) {
				var buf strings.Builder
				err := tpl.ExecuteTemplate(&buf, name, data)
				return buf.String(), err
			},
			"tpl":      templ,
			"required": notImplemented,
		})

}

func templ(stringTemplate string, values interface{}) interface{} {
	tpl, err := template.New("test").Parse(stringTemplate)
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

func toYAML(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}

func (f files) Glob(pattern string) files {
	return f
}
