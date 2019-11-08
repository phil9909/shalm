package chart

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kramerul/shalm/internal/pkg/k8s"

	"github.com/blang/semver"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// Chart -
type Chart struct {
	Name        string
	Version     semver.Version
	values      map[string]starlark.Value
	methods     map[string]starlark.Callable
	frozen      bool
	repo        Repo
	dir         string
	initialized bool
}

var (
	_ starlark.HasAttrs    = (*Chart)(nil)
	_ starlark.HasSetField = (*Chart)(nil)
)

// LoadChart -
func LoadChart(thread *starlark.Thread, repo Repo, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() == 0 {
		return nil, fmt.Errorf("chart: expected paramater name")
	}

	return NewChart(thread, repo, args.Index(0).(starlark.String).GoString(), args[1:], kwargs)
}

// NewChart -
func NewChart(thread *starlark.Thread, repo Repo, name string, args starlark.Tuple, kwargs []starlark.Tuple) (*Chart, error) {
	dir, err := repo.Directory(name)
	if err != nil {
		return nil, err
	}
	c := &Chart{Name: name, repo: repo, dir: dir}
	c.values = make(map[string]starlark.Value)
	c.methods = make(map[string]starlark.Callable)
	if err = c.loadChartYaml(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	if err = c.loadValuesYaml(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	if err = c.init(thread, args, kwargs); err != nil {
		return nil, err
	}
	c.initialized = true
	return c, nil

}

func (c *Chart) path(part ...string) string {
	return filepath.Join(append([]string{c.dir}, part...)...)
}

func (c *Chart) String() string {
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
func (c *Chart) Type() string { return "chart" }

// Truth -
func (c *Chart) Truth() starlark.Bool { return true } // even when empty

// Hash -
func (c *Chart) Hash() (uint32, error) {
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
func (c *Chart) Freeze() {
	if c.frozen {
		return
	}
	c.frozen = true
	for _, e := range c.values {
		e.Freeze()
	}
}

// Attr returns the value of the specified field.
func (c *Chart) Attr(name string) (starlark.Value, error) {
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
func (c *Chart) AttrNames() []string {
	names := make([]string, 0)
	for k := range c.values {
		names = append(names, k)
	}
	names = append(names, "template")
	return names
}

// SetField -
func (c *Chart) SetField(name string, val starlark.Value) error {
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
func (c *Chart) CompareSameType(op syntax.Token, yv starlark.Value, depth int) (bool, error) {
	y := yv.(*Chart)
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

func chartEqual(x, y *Chart, depth int) (bool, error) {
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

// ApplyFunction -
func (c *Chart) ApplyFunction() starlark.Callable {
	return c.methods["apply"]
}

func (c *Chart) applyFunction() starlark.Callable {
	return starlark.NewBuiltin("apply", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var release *Release
		var k k8s.K8s
		if err := starlark.UnpackArgs("apply", args, kwargs, "k8s", &k, "release", &release); err != nil {
			return nil, err
		}
		return starlark.None, c.apply(thread, k, release)
	})
}

func (c *Chart) apply(thread *starlark.Thread, k k8s.K8s, release *Release) error {
	err := c.eachSubChart(func(subChart *Chart) error {
		_, err := subChart.ApplyFunction().CallInternal(thread, starlark.Tuple{k, release}, nil)
		return err
	})
	if err != nil {
		return err
	}
	return c.applyLocal(thread, k, release)
}

func (c *Chart) applyLocalFunction() starlark.Callable {
	return starlark.NewBuiltin("__apply", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var release *Release
		var k k8s.K8s
		if err := starlark.UnpackArgs("__apply", args, kwargs, "k8s", &k, "release", &release); err != nil {
			return nil, err
		}
		return starlark.None, c.applyLocal(thread, k, release)
	})
}

func (c *Chart) applyLocal(thread *starlark.Thread, k k8s.K8s, release *Release) error {
	return k.Apply(release.Namespace, func(writer io.Writer) error {
		return c.template(thread, release, writer)
	})
}

// DeleteFunction -
func (c *Chart) DeleteFunction() starlark.Callable {
	return c.methods["delete"]
}

func (c *Chart) deleteFunction() starlark.Callable {
	return starlark.NewBuiltin("delete", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var release *Release
		var k k8s.K8s
		if err := starlark.UnpackArgs("delete", args, kwargs, "k8s", &k, "release", &release); err != nil {
			return nil, err
		}
		return starlark.None, c.delete(thread, k, release)
	})
}

func (c *Chart) delete(thread *starlark.Thread, k k8s.K8s, release *Release) error {
	err := c.eachSubChart(func(subChart *Chart) error {
		_, err := subChart.DeleteFunction().CallInternal(thread, starlark.Tuple{k, release}, nil)
		return err
	})
	if err != nil {
		return err
	}
	return c.deleteLocal(thread, k, release)
}

func (c *Chart) deleteLocalFunction() starlark.Callable {
	return starlark.NewBuiltin("__delete", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var release *Release
		var k k8s.K8s
		if err := starlark.UnpackArgs("__delete", args, kwargs, "k8s", &k, "release", &release); err != nil {
			return nil, err
		}
		return starlark.None, c.deleteLocal(thread, k, release)
	})
}

func (c *Chart) deleteLocal(thread *starlark.Thread, k k8s.K8s, release *Release) error {
	return k.Delete(release.Namespace, func(writer io.Writer) error {
		return c.template(thread, release, writer)
	})
}

func (c *Chart) eachSubChart(block func(subChart *Chart) error) error {
	for _, v := range c.values {
		subChart, ok := v.(*Chart)
		if ok {
			err := block(subChart)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
