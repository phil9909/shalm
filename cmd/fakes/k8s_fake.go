package fakes

import (
	"io"

	"github.com/kramerul/shalm/internal/pkg/k8s"

	"go.starlark.net/starlark"
)

// K8sFake -
type K8sFake struct {
	Writer io.Writer
}

var (
	_ k8s.K8s = (*K8sFake)(nil)
)

// String -
func (k K8sFake) String() string {
	panic("implement me")
}

// Type -
func (k *K8sFake) Type() string {
	panic("implement me")
}

// Freeze -
func (k *K8sFake) Freeze() {
	panic("implement me")
}

// Truth -
func (k *K8sFake) Truth() starlark.Bool {
	panic("implement me")
}

// Hash -
func (k *K8sFake) Hash() (uint32, error) {
	panic("implement me")
}

// Attr -
func (k *K8sFake) Attr(name string) (starlark.Value, error) {
	return starlark.NewBuiltin(name, func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		return starlark.None, nil
	}), nil
}

// AttrNames -
func (k *K8sFake) AttrNames() []string {
	panic("implement me")
}

// Apply -
func (k *K8sFake) Apply(namespace string, output func(io.Writer) error) error {
	return output(k.Writer)
}

// Delete -
func (k *K8sFake) Delete(namespace string, output func(io.Writer) error) error {
	return output(k.Writer)
}
