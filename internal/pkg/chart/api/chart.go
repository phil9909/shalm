package api

import (
	"io"

	"github.com/spf13/afero"
	"go.starlark.net/starlark"
)

// InstallOpts -
type InstallOpts struct {
	Namespace string
}

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

// Chart -
type Chart interface {
	GetName() string
	GetNamespace() string
	GetDir() string
	GetFs() afero.Fs
	Walk(cb func(name string, size int64, body io.Reader, err error) error) error
	Apply(thread *starlark.Thread, k K8s) error
	Delete(thread *starlark.Thread, k K8s) error
	Template(thread *starlark.Thread) (string, error)
}

// ChartValue -
type ChartValue interface {
	starlark.HasSetField
	Chart
}
