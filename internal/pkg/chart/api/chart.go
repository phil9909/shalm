package api

import (
	"io"

	"go.starlark.net/starlark"
)

// Release -
type Release struct {
	Name      string
	Namespace string
	Service   string
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
	Walk(cb func(name string, size int64, body io.Reader, err error) error) error
	Apply(thread *starlark.Thread, k K8s, release *Release) error
	Delete(thread *starlark.Thread, k K8s, release *Release) error
	Template(thread *starlark.Thread, release *Release) (string, error)
}

// ChartValue -
type ChartValue interface {
	starlark.HasSetField
	Chart
}
