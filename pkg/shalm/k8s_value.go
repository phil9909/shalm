package shalm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"go.starlark.net/starlark"
)

// NewK8sValue create new instance to interact with kubernetes
func NewK8sValue(k K8s) K8sValue {
	return &k8sValueImpl{k}
}

type k8sValueImpl struct {
	K8s
}

type k8sWatcher struct {
	k8s     K8s
	kind    string
	name    string
	options *K8sOptions
}

type k8sWatcherIterator struct {
	decoder *json.Decoder
	closer  io.Closer
}

var (
	_ starlark.Iterable = (*k8sWatcher)(nil)
	_ starlark.Iterator = (*k8sWatcherIterator)(nil)
	_ K8sValue          = (*k8sValueImpl)(nil)
)

// String -
func (k *k8sValueImpl) String() string { return k.Inspect() }

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
			err = json.Unmarshal(buffer.Bytes(), &obj)
			if err != nil {
				return starlark.None, err
			}
			if obj["kind"] == "Secret" {
				var s secret
				err = json.Unmarshal(buffer.Bytes(), &s)
				if err != nil {
					return starlark.None, err
				}
				return wrapDict(toStarlark(map[string]interface{}{"data": s.Data})), nil
			}
			return wrapDict(toStarlark(obj)), nil
		}), nil
	}
	if name == "watch" {
		return starlark.NewBuiltin("watch", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
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
			return &k8sWatcher{name: name, kind: kind, options: k8sOptions, k8s: k.K8s}, nil
		}), nil
	}

	return starlark.None, starlark.NoSuchAttrError(fmt.Sprintf("k8s has no .%s attribute", name))
}

// AttrNames -
func (k *k8sValueImpl) AttrNames() []string { return []string{"rollout_status", "delete", "get"} }

func unpackK8sOptions(parser *kwargsParser) *K8sOptions {
	result := &K8sOptions{Namespaced: true}
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

func (w *k8sWatcher) Freeze()               {}
func (w *k8sWatcher) String() string        { return "k8sWatcher" }
func (w *k8sWatcher) Type() string          { return "k8sWatcher" }
func (w *k8sWatcher) Truth() starlark.Bool  { return true }
func (w *k8sWatcher) Hash() (uint32, error) { return 0, fmt.Errorf("k8sWatcher is unhashable") }
func (w *k8sWatcher) Iterate() starlark.Iterator {
	reader, err := w.k8s.Watch(w.kind, w.name, w.options)
	if err != nil {
		return &k8sWatcherIterator{decoder: json.NewDecoder(bufio.NewReader(bytes.NewReader([]byte{})))}
	}
	return &k8sWatcherIterator{decoder: json.NewDecoder(reader), closer: reader}
}

func (i *k8sWatcherIterator) Next(p *starlark.Value) bool {
	var obj map[string]interface{}
	err := i.decoder.Decode(&obj)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	*p = wrapDict(toStarlark(obj))
	return true
}

func (i *k8sWatcherIterator) Done() {
	if i.closer != nil {
		i.closer.Close()
	}
}
