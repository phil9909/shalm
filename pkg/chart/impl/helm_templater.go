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

	"gopkg.in/yaml.v2"

	"github.com/Masterminds/sprig/v3"
)

// HelmTemplater -
type HelmTemplater struct {
	helpers   string
	dir       string
	namespace string
}

// HelmOptions -
type HelmOptions struct {
	glob           string
	uninstallOrder bool
}

// MetaData -
type MetaData struct {
	Namespace  string                 `yaml:"namespace,omitempty"`
	Name       string                 `yaml:"name,omitempty"`
	Additional map[string]interface{} `yaml:",inline"`
}

// Object -
type Object struct {
	MetaData   MetaData               `yaml:"metadata,omitempty"`
	Kind       string                 `yaml:"kind,omitempty"`
	Additional map[string]interface{} `yaml:",inline"`
}

func (o *Object) setDefaultNamespace(namespace string) {
	switch o.Kind {
	case "Namespace":
		return
	case "ResourceQuota":
		return
	case "CustomResourceDefinition":
		return
	case "ClusterRole":
		return
	case "ClusterRoleList":
		return
	case "ClusterRoleBinding":
		return
	case "ClusterRoleBindingList":
		return
	case "APIService":
		return
	}
	if o.MetaData.Namespace == "" {
		o.MetaData.Namespace = namespace
	}
}

func (o *Object) kindOrdinal() int {
	switch o.Kind {
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

type files struct {
	dir string
}

// NewHelmTemplater -
func NewHelmTemplater(dir string, namespace string) (*HelmTemplater, error) {
	h := &HelmTemplater{
		dir:       dir,
		namespace: namespace,
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

// Template -
func (h *HelmTemplater) Template(value interface{}, writer io.Writer, opts *HelmOptions) error {
	var filenames []string
	glob := "*.yaml"
	if opts.glob != "" {
		glob = opts.glob
	}
	err := filepath.Walk(path.Join(h.dir, "templates"), func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			match, err := filepath.Match(glob, path.Base(file))
			if err != nil {
				return err
			}
			if match {
				filenames = append(filenames, file)
			}
		}
		return nil
	})

	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	if len(filenames) == 0 {
		return nil
	}

	var docs []Object
	for _, filename := range filenames {
		var buffer bytes.Buffer
		tpl, err := h.template(filepath.Base(filename))
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
		err = tpl.Execute(&buffer, value)
		if err != nil {
			return err
		}
		if buffer.Len() > 0 {
			dec := yaml.NewDecoder(&buffer)
			var doc Object
			for dec.Decode(&doc) == nil {
				doc.setDefaultNamespace(h.namespace)
				docs = append(docs, doc)
			}
		}
	}
	if opts.uninstallOrder {
		sort.Slice(docs, func(i, j int) bool {
			return docs[i].kindOrdinal() > docs[j].kindOrdinal()
		})
	} else {
		sort.Slice(docs, func(i, j int) bool {
			return docs[i].kindOrdinal() < docs[j].kindOrdinal()
		})
	}
	writer.Write([]byte("---\n"))
	enc := yaml.NewEncoder(writer)
	for _, doc := range docs {
		err = enc.Encode(doc)
		if err != nil {
			return err
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
