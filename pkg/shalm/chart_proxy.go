package shalm

import (
	"bytes"
	"io"

	corev1 "k8s.io/api/core/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	shalmv1a1 "github.com/kramerul/shalm/api/v1alpha1"

	"go.starlark.net/starlark"
)

type chartProxy struct {
	*chartImpl
	args   []interface{}
	kwargs map[string]interface{}
	url    string
}

var (
	_ ChartValue = (*chartProxy)(nil)
)

func newChartProxy(delegate *chartImpl, url string, args starlark.Tuple, kwargs []starlark.Tuple) (ChartValue, error) {
	return &chartProxy{
		chartImpl: delegate,
		args:      toGo(args).([]interface{}),
		kwargs:    kwargsToGo(kwargs),
		url:       url,
	}, nil
}

// Attr returns the value of the specified field.
func (c *chartProxy) Attr(name string) (starlark.Value, error) {
	switch name {
	case "apply":
		return c.applyFunction(), nil
	case "delete":
		return c.deleteFunction(), nil
	}
	return c.chartImpl.Attr(name)
}

func (c *chartProxy) applyFunction() starlark.Callable {
	return starlark.NewBuiltin("apply", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var k K8sValue
		if err := starlark.UnpackArgs("apply", args, kwargs, "k8s", &k); err != nil {
			return nil, err
		}
		namespace := &corev1.Namespace{
			TypeMeta: v1.TypeMeta{
				Kind:       "Namespace",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
			ObjectMeta: v1.ObjectMeta{
				Name: c.namespace,
			},
		}
		shalmSpec := shalmv1a1.ShalmChartSpec{
			Values:     shalmv1a1.ClonableMap(c.chartImpl.templateValues()),
			Args:       shalmv1a1.ClonableArray(c.args),
			KwArgs:     shalmv1a1.ClonableMap(c.kwargs),
			KubeConfig: "",
			Namespace:  c.namespace,
		}
		buffer := &bytes.Buffer{}
		if err := c.chartImpl.Package(buffer); err != nil {
			return nil, err
		}
		shalmSpec.ChartTgz = buffer.Bytes()
		shalmChart := &shalmv1a1.ShalmChart{
			TypeMeta: v1.TypeMeta{
				Kind:       "ShalmChart",
				APIVersion: shalmv1a1.GroupVersion.String(),
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      c.Name,
				Namespace: c.namespace,
			},
			Spec: shalmSpec,
		}

		encoder := json.NewSerializerWithOptions(json.DefaultMetaFactory, nil, nil, json.SerializerOptions{})

		return starlark.None, k.Apply(func(writer io.Writer) error {
			err := encoder.Encode(namespace, writer)
			if err != nil {
				return err
			}
			return encoder.Encode(shalmChart, writer)
		}, &K8sOptions{})
	})
}

func (c *chartProxy) Apply(thread *starlark.Thread, k K8s) error {
	_, err := starlark.Call(thread, c.applyFunction(), starlark.Tuple{NewK8sValue(k)}, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *chartProxy) deleteFunction() starlark.Callable {
	return starlark.NewBuiltin("delete", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var k K8sValue
		if err := starlark.UnpackArgs("delete", args, kwargs, "k8s", &k); err != nil {
			return nil, err
		}

		return starlark.None, k.DeleteObject("ShalmChart", c.Name, &K8sOptions{})
	})
}

func (c *chartProxy) Delete(thread *starlark.Thread, k K8s) error {
	_, err := starlark.Call(thread, c.deleteFunction(), starlark.Tuple{NewK8sValue(k)}, nil)
	if err != nil {
		return err
	}
	return nil

}
