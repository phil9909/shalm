package api

import (
	"io"
	"time"

	"go.starlark.net/starlark"
)

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
}

// K8sValue -
type K8sValue interface {
	starlark.Value
	K8s
}
