package impl

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/blang/semver"
	"github.com/kramerul/shalm/internal/pkg/chart/api"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

type chartImpl struct {
	Name        string
	Version     semver.Version
	values      map[string]starlark.Value
	methods     map[string]starlark.Callable
	frozen      bool
	dir         string
	initialized bool
}

var (
	_ api.ChartValue = (*chartImpl)(nil)
)

// NewChart -
func NewChart(thread *starlark.Thread, repo api.Repo, dir string, name string, args starlark.Tuple, kwargs []starlark.Tuple) (api.ChartValue, error) {
	c := &chartImpl{Name: name, dir: dir}
	c.values = make(map[string]starlark.Value)
	c.methods = make(map[string]starlark.Callable)
	if err := c.loadChartYaml(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	if err := c.loadValuesYaml(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	if err := c.init(thread, repo, args, kwargs); err != nil {
		return nil, err
	}
	c.initialized = true
	return c, nil

}

func (c *chartImpl) GetName() string {
	return c.Name
}

func (c *chartImpl) GetDir() string {
	return c.dir
}

func (c *chartImpl) Walk(cb func(name string, size int64, body io.Reader, err error) error) error {
	return filepath.Walk(c.dir, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(c.dir, file)
		if err != nil {
			return err
		}
		body, err := os.Open(file)
		if err != nil {
			return err
		}
		defer body.Close()
		return cb(rel, info.Size(), body, nil)
	})
}

func (c *chartImpl) path(part ...string) string {
	return filepath.Join(append([]string{c.dir}, part...)...)
}

func (c *chartImpl) String() string {
	buf := new(strings.Builder)
	buf.WriteString("chart")
	buf.WriteByte('(')
	s := 0
	for i, e := range c.values {
		if s > 0 {
			buf.WriteString(", ")
		}
		s++
		buf.WriteString(i)
		buf.WriteString(" = ")
		buf.WriteString(e.String())
	}
	buf.WriteByte(')')
	return buf.String()
}

// Type -
func (c *chartImpl) Type() string { return "chart" }

// Truth -
func (c *chartImpl) Truth() starlark.Bool { return true } // even when empty

// Hash -
func (c *chartImpl) Hash() (uint32, error) {
	var x, m uint32 = 8731, 9839
	for k, e := range c.values {
		namehash, _ := starlark.String(k).Hash()
		x = x ^ 3*namehash
		y, err := e.Hash()
		if err != nil {
			return 0, err
		}
		x = x ^ y*m
		m += 7349
	}
	return x, nil
}

// Freeze -
func (c *chartImpl) Freeze() {
	if c.frozen {
		return
	}
	c.frozen = true
	for _, e := range c.values {
		e.Freeze()
	}
}

// Attr returns the value of the specified field.
func (c *chartImpl) Attr(name string) (starlark.Value, error) {
	value, ok := c.values[name]
	if !ok {
		var m starlark.Value
		m, ok = c.methods[name]
		if !ok {
			m = nil
		}
		if m == nil {
			return nil, starlark.NoSuchAttrError(
				fmt.Sprintf("chart has no .%s attribute", name))

		}
		return m, nil
	}
	return value, nil
}

// AttrNames returns a new sorted list of the struct fields.
func (c *chartImpl) AttrNames() []string {
	names := make([]string, 0)
	for k := range c.values {
		names = append(names, k)
	}
	names = append(names, "template")
	return names
}

// SetField -
func (c *chartImpl) SetField(name string, val starlark.Value) error {
	if c.frozen {
		return fmt.Errorf("chart is frozen")
	}
	if c.initialized {
		_, ok := c.values[name]
		if !ok {
			return starlark.NoSuchAttrError(
				fmt.Sprintf("chart has no .%s attribute", name))
		}
	}
	c.values[name] = val
	return nil
}

// CompareSameType -
func (c *chartImpl) CompareSameType(op syntax.Token, yv starlark.Value, depth int) (bool, error) {
	y := yv.(*chartImpl)
	switch op {
	case syntax.EQL:
		return chartEqual(c, y, depth)
	case syntax.NEQ:
		eq, err := chartEqual(c, y, depth)
		return !eq, err
	default:
		return false, fmt.Errorf("%s %s %s not implemented", c.Type(), op, y.Type())
	}
}

func chartEqual(x, y *chartImpl, depth int) (bool, error) {
	if len(x.values) != len(y.values) {
		return false, nil
	}

	for k, vx := range x.values {
		vy, ok := y.values[k]
		if !ok {
			return false, nil
		} else if eq, err := starlark.EqualDepth(vx, vy, depth-1); err != nil {
			return false, err
		} else if !eq {
			return false, nil
		}
	}
	return true, nil
}

func notImplemented(_ interface{}) string {
	panic("not implemented")
}

func (c *chartImpl) applyFunction() starlark.Callable {
	return starlark.NewBuiltin("apply", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var release *releaseValue
		var k api.K8sValue
		if err := starlark.UnpackArgs("apply", args, kwargs, "k8s", &k, "release", &release); err != nil {
			return nil, err
		}
		return starlark.None, c.apply(thread, k, release)
	})
}

func (c *chartImpl) Apply(thread *starlark.Thread, k api.K8s, release *api.Release) error {
	_, err := starlark.Call(thread, c.methods["apply"], starlark.Tuple{NewK8sValue(k), NewReleaseValue(release)}, nil)
	if err != nil {
		return err
	}
	return nil

}

func (c *chartImpl) apply(thread *starlark.Thread, k api.K8sValue, release *releaseValue) error {
	err := c.eachSubChart(func(subChart *chartImpl) error {
		_, err := subChart.methods["apply"].CallInternal(thread, starlark.Tuple{k, release}, nil)
		return err
	})
	if err != nil {
		return err
	}
	return c.applyLocal(thread, k, release)
}

func (c *chartImpl) applyLocalFunction() starlark.Callable {
	return starlark.NewBuiltin("__apply", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var release *releaseValue
		var k api.K8sValue
		if err := starlark.UnpackArgs("__apply", args, kwargs, "k8s", &k, "release", &release); err != nil {
			return nil, err
		}
		return starlark.None, c.applyLocal(thread, k, release)
	})
}

func (c *chartImpl) applyLocal(thread *starlark.Thread, k api.K8sValue, release *releaseValue) error {
	return k.Apply(release.Namespace, func(writer io.Writer) error {
		return c.template(thread, release, writer)
	})
}

func (c *chartImpl) Delete(thread *starlark.Thread, k api.K8s, release *api.Release) error {
	_, err := starlark.Call(thread, c.methods["delete"], starlark.Tuple{NewK8sValue(k), NewReleaseValue(release)}, nil)
	if err != nil {
		return err
	}
	return nil

}

func (c *chartImpl) deleteFunction() starlark.Callable {
	return starlark.NewBuiltin("delete", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var release *releaseValue
		var k api.K8sValue
		if err := starlark.UnpackArgs("delete", args, kwargs, "k8s", &k, "release", &release); err != nil {
			return nil, err
		}
		return starlark.None, c.delete(thread, k, release)
	})
}

func (c *chartImpl) delete(thread *starlark.Thread, k api.K8sValue, release *releaseValue) error {
	err := c.eachSubChart(func(subChart *chartImpl) error {
		_, err := subChart.methods["delete"].CallInternal(thread, starlark.Tuple{k, release}, nil)
		return err
	})
	if err != nil {
		return err
	}
	return c.deleteLocal(thread, k, release)
}

func (c *chartImpl) deleteLocalFunction() starlark.Callable {
	return starlark.NewBuiltin("__delete", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var release *releaseValue
		var k api.K8sValue
		if err := starlark.UnpackArgs("__delete", args, kwargs, "k8s", &k, "release", &release); err != nil {
			return nil, err
		}
		return starlark.None, c.deleteLocal(thread, k, release)
	})
}

func (c *chartImpl) deleteLocal(thread *starlark.Thread, k api.K8sValue, release *releaseValue) error {
	return k.Delete(release.Namespace, func(writer io.Writer) error {
		return c.template(thread, release, writer)
	})
}

func (c *chartImpl) eachSubChart(block func(subChart *chartImpl) error) error {
	for _, v := range c.values {
		subChart, ok := v.(*chartImpl)
		if ok {
			err := block(subChart)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
