package shalm

import (
	"io"
	"time"

	"github.com/blang/semver"
	"go.starlark.net/starlark"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o fake_k8s.go . K8s

// Credential -
type Credential interface {
	GetOrCreate(k K8s) error
}

// CredentialValue -
type CredentialValue interface {
	starlark.HasAttrs
	Credential
}

// Chart -
type Chart interface {
	GetName() string
	GetVersion() semver.Version
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

// K8sOptions common options for calls to k8s
type K8sOptions struct {
	Namespaced bool
	Timeout    time.Duration
}

// K8s kubernetes API
type K8s interface {
	ForNamespace(namespace string) K8s
	RolloutStatus(kind string, name string, options *K8sOptions) error
	DeleteObject(kind string, name string, options *K8sOptions) error
	Apply(output func(io.Writer) error, options *K8sOptions) error
	Delete(output func(io.Writer) error, options *K8sOptions) error
	Get(kind string, name string, writer io.Writer, options *K8sOptions) error
	IsNotExist(err error) bool
}

// K8sValue -
type K8sValue interface {
	starlark.Value
	K8s
}

// Repo -
type Repo interface {
	// Get -
	Get(thread *starlark.Thread, url string, namespace string, args starlark.Tuple, kwargs []starlark.Tuple) (ChartValue, error)
}