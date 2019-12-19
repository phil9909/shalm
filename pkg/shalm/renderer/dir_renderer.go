package renderer

import (
	"bytes"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v2"
)

// MetaData -
type MetaData struct {
	Namespace  string                 `yaml:"namespace,omitempty"`
	Name       string                 `yaml:"name,omitempty"`
	Additional map[string]interface{} `yaml:",inline"`
}

type object struct {
	MetaData   MetaData               `yaml:"metadata,omitempty"`
	Kind       string                 `yaml:"kind,omitempty"`
	Additional map[string]interface{} `yaml:",inline"`
}

// Options -
type Options struct {
	Glob           string
	UninstallOrder bool
}

func (o *object) setDefaultNamespace(namespace string) {
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

func (o *object) kindOrdinal() int {
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

// DirSpec -
type DirSpec struct {
	Dir          string
	FileRenderer func(filename string, writer io.Writer) error
}

// DirRender -
func DirRender(namespace string, writer io.Writer, opts *Options, specs ...DirSpec) error {
	glob := "*.y*ml"
	if opts.Glob != "" {
		glob = opts.Glob
	}
	var docs []object
	for _, r := range specs {
		var filenames []string

		err := filepath.Walk(r.Dir, func(file string, info os.FileInfo, err error) error {
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

		for _, filename := range filenames {
			var buffer bytes.Buffer
			err := r.FileRenderer(filename, &buffer)
			if err != nil {
				return err
			}
			if buffer.Len() > 0 {
				dec := yaml.NewDecoder(&buffer)
				var doc object
				for dec.Decode(&doc) == nil {
					doc.setDefaultNamespace(namespace)
					docs = append(docs, doc)
				}
			}
		}
	}
	if opts.UninstallOrder {
		sort.Slice(docs, func(i, j int) bool {
			return docs[i].kindOrdinal() > docs[j].kindOrdinal()
		})
	} else {
		sort.Slice(docs, func(i, j int) bool {
			return docs[i].kindOrdinal() < docs[j].kindOrdinal()
		})
	}
	if len(docs) > 0 {
		writer.Write([]byte("---\n"))
		enc := yaml.NewEncoder(writer)
		for _, doc := range docs {
			err := enc.Encode(doc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
