package chart

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"go.starlark.net/starlark"

	yaml "gopkg.in/yaml.v2"

	"github.com/Masterminds/sprig/v3"
)

// Release -
type Release struct {
	Name      string
	Namespace string
	Service   string
}

var (
	_ starlark.HasAttrs = (*Release)(nil)
)

// String -
func (r *Release) String() string {
	return "release"
}

// Type -
func (r *Release) Type() string {
	return "release"
}

// Freeze -
func (r *Release) Freeze() {
}

// Truth -
func (r *Release) Truth() starlark.Bool {
	return false
}

// Hash -
func (r *Release) Hash() (uint32, error) {
	panic("implement me")
}

// Attr -
func (r *Release) Attr(name string) (starlark.Value, error) {
	if name == "namespace" {
		return starlark.String(r.Namespace), nil
	}
	if name == "name" {
		return starlark.String(r.Name), nil
	}
	if name == "service" {
		return starlark.String(r.Service), nil
	}
	return starlark.None, starlark.NoSuchAttrError(fmt.Sprintf("release has no .%s attribute", name))
}

// AttrNames -
func (r *Release) AttrNames() []string {
	return []string{"namespace", "name", "service"}
}

var _ starlark.Value = &Release{}

// HelmChart -
type HelmChart struct {
	APIVersion  string   `json:"apiVersion,omitempty"`
	Name        string   `json:"name,omitempty"`
	Version     string   `json:"version,omitempty"`
	Description string   `json:"description,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
	Home        string   `json:"home,omitempty"`
	Sources     []string `json:"sources,omitempty"`
	Icon        string   `json:"icon,omitempty"`
}

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
