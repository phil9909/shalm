package impl

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/kramerul/shalm/pkg/chart"
	"go.starlark.net/starlark"
)

// NewK8sValue create new instance to interact with kubernetes
func NewK8sValue(k chart.K8s) chart.K8sValue {
	return &k8sValueImpl{k}
}

type k8sValueImpl struct {
	chart.K8s
}

var (
	_ chart.K8sValue = (*k8sValueImpl)(nil)
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
			var kind string
			var name string
			parser := &kwargsParser{kwargs: kwargs}
			k8sOptions := unpackK8sOptions(parser)
			if err := starlark.UnpackArgs("rollout_status", args, parser.Parse(),
				"kind", &kind, "name", &name); err != nil {
				return nil, err
			}
			return starlark.None, k.RolloutStatus(kind, name, k8sOptions)
		}), nil
	}
	if name == "delete" {
		return starlark.NewBuiltin("delete", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
			var kind string
			var name string
			parser := &kwargsParser{kwargs: kwargs}
			k8sOptions := unpackK8sOptions(parser)
			if err := starlark.UnpackArgs("delete", args, parser.Parse(),
				"kind", &kind, "name?", &name); err != nil {
				return nil, err
			}
			if name == "" {
				return starlark.None, errors.New("no parameter name given")
			}
			return starlark.None, k.DeleteObject(kind, name, k8sOptions)
		}), nil
	}
	if name == "get" {
		return starlark.NewBuiltin("get", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
			var kind string
			var name string
			parser := &kwargsParser{kwargs: kwargs}
			k8sOptions := unpackK8sOptions(parser)
			if err := starlark.UnpackArgs("get", args, parser.Parse(),
				"kind", &kind, "name", &name); err != nil {
				return nil, err
			}
			if name == "" {
				return starlark.None, errors.New("no parameter name given")
			}
			var buffer bytes.Buffer
			err := k.Get(kind, name, &buffer, k8sOptions)
			if err != nil {
				return starlark.None, err
			}
			var obj map[string]interface{}
			yaml.Unmarshal(buffer.Bytes(), &obj)
			return toStarlark(obj), nil
			//return starlark.String(string(buffer.Bytes())), nil
		}), nil
	}
	return starlark.None, starlark.NoSuchAttrError(fmt.Sprintf("k8s has no .%s attribute", name))
}

// AttrNames -
func (k *k8sValueImpl) AttrNames() []string { return []string{"rollout_status", "delete"} }

func unpackK8sOptions(parser *kwargsParser) *chart.K8sOptions {
	result := &chart.K8sOptions{Namespaced: true}
	parser.Arg("namespaced", func(value starlark.Value) {
		result.Namespaced = bool(value.(starlark.Bool))
	})
	parser.Arg("timeout", func(value starlark.Value) {
		timeout, ok := value.(starlark.Int).Int64()
		if ok {
			result.Timeout = time.Duration(timeout) * time.Second
		}
	})
	return result
}
