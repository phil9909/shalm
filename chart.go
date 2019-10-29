package main

import (
	"fmt"
	"os"
	"strings"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

type Chart struct {
	values map[string]starlark.Value
	init   bool
	frozen bool
}

var (
	_ starlark.HasAttrs    = (*Chart)(nil)
	_ starlark.HasSetField = (*Chart)(nil)
)

func LoadChart(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return nil, fmt.Errorf("chart: expected paramater name")
	}

	name := args.Index(0).String()

	predeclared := starlark.StringDict{
		"chart": starlark.NewBuiltin("chart", LoadChart),
	}

	result := &Chart{}
	file := fmt.Sprintf("example/%s.star", name)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return result, nil
	}
	globals, err := starlark.ExecFile(thread, file, nil, predeclared)
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
		s += 1
		buf.WriteString(i)
		buf.WriteString(" = ")
		buf.WriteString(e.String())
	}
	buf.WriteByte(')')
	return buf.String()
}

func (c *Chart) Type() string         { return "chart" }
func (c *Chart) Truth() starlark.Bool { return true } // even when empty
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
	for k, _ := range c.values {
		names = append(names, k)
	}
	return names
}

// SetField
func (c *Chart) SetField(name string, val starlark.Value) error {
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

func (c *Chart) CompareSameType(op syntax.Token, y_ starlark.Value, depth int) (bool, error) {
	y := y_.(*Chart)
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
