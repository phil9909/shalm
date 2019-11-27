package impl

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	yaml "gopkg.in/yaml.v2"

	"github.com/Masterminds/sprig/v3"
	"github.com/spf13/afero"
)

// HelmTemplater -
type HelmTemplater struct {
	helpers string
	dir     string
	fs      afero.Fs
}

type options struct {
	glob           *string
	uninstallOrder bool
}

// Option for templating
type Option func(o *options)

// WithGlob only use a subset of templates
func WithGlob(glob string) Option {
	return func(o *options) {
		o.glob = &glob
	}
}

// WithUninstallOrder sort yaml docs in reverse order
func WithUninstallOrder() Option {
	return func(o *options) {
		o.uninstallOrder = true
	}
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

type yamlDocs []map[string]interface{}

func (a yamlDocs) Len() int      { return len(a) }
func (a yamlDocs) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func kindOrdinal(kind interface{}) int {
	switch kind {
	case "Namespace":
		return 1
	case "NetworkPolicy":
		return 2
	case "ResourceQuota":
		return 3
	case "LimitRange":
		return 4
	case "PodSecurityPolicy":
		return 5
	case "PodDisruptionBudget":
		return 6
	case "Secret":
		return 7
	case "ConfigMap":
		return 8
	case "StorageClass":
		return 9
	case "PersistentVolume":
		return 10
	case "PersistentVolumeClaim":
		return 11
	case "ServiceAccount":
		return 12
	case "CustomResourceDefinition":
		return 13
	case "ClusterRole":
		return 14
	case "ClusterRoleList":
		return 15
	case "ClusterRoleBinding":
		return 16
	case "ClusterRoleBindingList":
		return 17
	case "Role":
		return 18
	case "RoleList":
		return 19
	case "RoleBinding":
		return 20
	case "RoleBindingList":
		return 21
	case "Service":
		return 22
	case "DaemonSet":
		return 23
	case "Pod":
		return 24
	case "ReplicationController":
		return 25
	case "ReplicaSet":
		return 26
	case "Deployment":
		return 27
	case "HorizontalPodAutoscaler":
		return 28
	case "StatefulSet":
		return 29
	case "Job":
		return 30
	case "CronJob":
		return 31
	case "Ingress":
		return 32
	case "APIService":
		return 33
	default:
		return 1000
	}
}

func (a yamlDocs) Less(i, j int) bool {
	return kindOrdinal(a[i]["kind"]) < kindOrdinal(a[j]["kind"])
}

// Template -
func (h *HelmTemplater) Template(value interface{}, writer io.Writer, opts ...Option) error {
	var glob string
	o := options{}
	for _, f := range opts {
		f(&o)
	}
	if o.glob != nil {
		glob = path.Join(h.dir, "templates", *o.glob)
	} else {
		glob = path.Join(h.dir, "templates", "*.yaml")
	}
	filenames, err := afero.Glob(h.fs, glob)
	if err != nil {
		return err
	}
	if len(filenames) == 0 {
		return nil
	}

	var docs yamlDocs
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
			dec := yaml.NewDecoder(&buffer)
			var doc map[string]interface{}
			for dec.Decode(&doc) == nil {
				docs = append(docs, doc)
			}
		}
	}
	if o.uninstallOrder {
		sort.Sort(sort.Reverse(docs))
	} else {
		sort.Sort(docs)
	}
	enc := yaml.NewEncoder(writer)
	for _, doc := range docs {
		enc.Encode(doc)
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
