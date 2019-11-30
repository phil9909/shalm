package api

import (
	"io"

	"github.com/blang/semver"
	"go.starlark.net/starlark"
)

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

// Credentials -
type Credential interface {
	GetOrCreate(k K8s) error
}

// CredentialsValue -
type CredentialValue interface {
	starlark.HasAttrs
	Credential
}

// Chart -
type Chart interface {
	GetName() string
	GetVersion() semver.Version
	GetNamespace() string
	GetDir() string
	Apply(thread *starlark.Thread, k K8s) error
	Delete(thread *starlark.Thread, k K8s) error
	Template(thread *starlark.Thread) (string, error)
	Package(writer io.Writer) error
}

// ChartValue -
type ChartValue interface {
	starlark.HasSetField
	Chart
}
