package impl

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/kramerul/shalm/internal/pkg/chart/api"
	"go.starlark.net/starlark"
)

// NewK8sValue create new instance to interact with kubernetes
func NewK8sValue(k api.K8s) api.K8sValue {
	return &k8sValueImpl{k}
}

type k8sValueImpl struct {
	api.K8s
}

var (
	_ starlark.HasAttrs = (*k8sValueImpl)(nil)
)

// String -
func (k *k8sValueImpl) String() string { return os.Getenv("KUBECONFIG") }

// Type -
func (k *k8sValueImpl) Type() string { return "k8s" }

// Freeze -
func (k *k8sValueImpl) Freeze() {}

// Truth -
func (k *k8sValueImpl) Truth() starlark.Bool { return false }

// Hash -
func (k *k8sValueImpl) Hash() (uint32, error) { panic("implement me") }

// Attr -
func (k *k8sValueImpl) Attr(name string) (starlark.Value, error) {
	if name == "rollout_status" {
		return starlark.NewBuiltin("rollout_status", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
			var timeout = 120
			var kind string
			var name string
			if err := starlark.UnpackArgs("rollout_status", args, kwargs,
				"type", &kind, "name", &name, "timeout?", &timeout); err != nil {
				return nil, err
			}
			return starlark.None, k.RolloutStatus(kind, name, time.Duration(timeout)*time.Second)
		}), nil
	}
	if name == "delete" {
		return starlark.NewBuiltin("delete", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
			var kind string
			var name string
			if err := starlark.UnpackArgs("delete", args, kwargs,
				"type", &kind, "name?", &name); err != nil {
				return nil, err
			}
			if name == "" {
				return starlark.None, errors.New("no parameter name given")
			}
			return starlark.None, k.DeleteObject(kind, name)
		}), nil
	}
	return starlark.None, starlark.NoSuchAttrError(fmt.Sprintf("k8s has no .%s attribute", name))
}

// AttrNames -
func (k *k8sValueImpl) AttrNames() []string { return []string{"wait_crd"} }
