package api

import (
	"io"
	"time"

	"go.starlark.net/starlark"
)

// K8s kubernetes API
type K8s interface {
	RolloutStatus(namespace string, typ string, name string, timeout time.Duration) error
	Apply(namespace string, output func(io.Writer) error) error
	Delete(namespace string, output func(io.Writer) error) error
}

// K8sValue -
type K8sValue interface {
	starlark.Value
	K8s
}
