package shalm

import (
	"io"
	"time"

	"github.com/blang/semver"
	"go.starlark.net/starlark"

	shalmv1a1 "github.com/kramerul/shalm/api/v1alpha1"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o fake_k8s_test.go . K8s

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
	Inspect() string
	Watch(kind string, name string, options *K8sOptions) (io.ReadCloser, error)
	RolloutStatus(kind string, name string, options *K8sOptions) error
	Wait(kind string, name string, condition string, options *K8sOptions) error
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
	Get(thread *starlark.Thread, url string, namespace string, proxy bool, args starlark.Tuple, kwargs []starlark.Tuple) (ChartValue, error)
	// GetFromSpec -
	GetFromSpec(thread *starlark.Thread, spec *shalmv1a1.ShalmChartSpec) (ChartValue, error)
}
