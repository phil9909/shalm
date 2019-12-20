package shalm

import (
	"io"

	"go.starlark.net/starlark"
)

type chartProxy struct {
	*chartImpl
}

var (
	_ ChartValue = (*chartProxy)(nil)
)

func newChartProxy(thread *starlark.Thread, repo Repo, dir string, namespace string, args starlark.Tuple, kwargs []starlark.Tuple) (ChartValue, error) {
	delegate, err := newChart(thread, repo, dir, namespace, args, kwargs)
	if err != nil {
		return nil, err
	}
	return &chartProxy{
		chartImpl: delegate.(*chartImpl),
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

		return starlark.None, k.Apply(func(writer io.Writer) error {
			return nil
		}, &K8sOptions{})
	})
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
