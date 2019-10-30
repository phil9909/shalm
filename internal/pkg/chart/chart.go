package chart

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"text/template"

	"github.com/blang/semver"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// Chart -
type Chart struct {
	name      string
	version   semver.Version
	values    map[string]starlark.Value
	init      bool
	frozen    bool
	directory string
}

var (
	_ starlark.HasAttrs    = (*Chart)(nil)
	_ starlark.HasSetField = (*Chart)(nil)
)

// LoadChart -
func LoadChart(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return nil, fmt.Errorf("chart: expected paramater name")
	}

	return NewChart(thread, args.Index(0).(starlark.String).GoString())
}

func predeclared() starlark.StringDict {
	return starlark.StringDict{
		"chart": starlark.NewBuiltin("chart", LoadChart),
		"helm":  starlark.NewBuiltin("chart", LoadHelmChart),
	}
}

// NewChart -
func NewChart(thread *starlark.Thread, name string) (*Chart, error) {

	directory := repo.Directory(name)
	result := &Chart{directory: directory, name: name}
	file := fmt.Sprintf("%s/chart.star", directory)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return result, nil
	}
	globals, err := starlark.ExecFile(thread, file, nil, predeclared())
	if err != nil {
		panic(err)
	}

	init, ok := globals["init"]

	if ok {
		_, err := starlark.Call(thread, init, starlark.Tuple{result}, nil)
		if err != nil {
			return nil, err
		}
	}

	return result, nil

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
		return nil, starlark.NoSuchAttrError(
			fmt.Sprintf("chart has no .%s attribute", name))

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
	if c.init {
		_, ok := c.values[name]
		if !ok {
			return starlark.NoSuchAttrError(
				fmt.Sprintf("chart has no .%s attribute", name))
		}
	}
	if c.values == nil {
		c.values = make(map[string]starlark.Value)
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

// Template -
func (c *Chart) Template() (string, error) {
	glob := c.directory + "/templates/*.yaml"
	tpl, err := template.ParseGlob(glob)
	if err != nil {
		return "", err
	}
	var buffer bytes.Buffer
	err = tpl.Execute(&buffer, struct{ Values interface{} }{Values: toGo(c)})
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func toGo(v starlark.Value) interface{} {
	switch v := v.(type) {
	case starlark.NoneType:
		return nil
	case starlark.Bool:
		return v
	case starlark.Int:
		i, _ := v.Int64()
		return i
	case starlark.Float:
		return v
	case starlark.String:
		return v.GoString()
	case starlark.Indexable: // Tuple, List
		a := make([]interface{}, 0)
		for i := 0; i < starlark.Len(v); i++ {
			a = append(a, toGo(v.Index(i)))
		}
		return a
	case starlark.IterableMapping:
		d := make(map[string]interface{})

		for _, t := range v.Items() {
			key, ok := t.Index(0).(starlark.String)
			if ok {
				d[key.GoString()] = toGo(t.Index(1))
			}
		}
		return d

	case *Chart:
		d := make(map[string]interface{})

		for k, v := range v.values {
			d[k] = toGo(v)
		}
		return d
	default:
		panic(fmt.Errorf("cannot convert %s to GO", v.Type()))
	}
}
