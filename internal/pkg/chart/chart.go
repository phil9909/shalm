package chart

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"text/template"

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
	if args.Len() != 1 {
		return nil, fmt.Errorf("chart: expected paramater name")
	}

	return NewChart(thread, repo, args.Index(0).(starlark.String).GoString())
}

// NewChart -
func NewChart(thread *starlark.Thread, repo Repo, name string) (*Chart, error) {
	dir, err := repo.Directory(name)
	if err != nil {
		return nil, err
	}
	c := &Chart{Name: name, repo: repo, dir: dir}
	c.values = make(map[string]starlark.Value)
	c.methods = make(map[string]starlark.Callable)
	if err = c.init(thread); err != nil {
		return nil, err
	}
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
	c.initialized = true
	return c, nil

}

func (c *Chart) path(part ...string) string {
	return filepath.Join(append([]string{c.dir}, part...)...)
}

func (c *Chart) loadChartYaml() error {
	var helmChart HelmChart

	err := loadYamlFile(c.path("Chart.yaml"), &helmChart)
	if err != nil {
		return err
	}
	c.Version = semver.MustParse(helmChart.Version)
	c.Name = helmChart.Name
	return nil
}

func (c *Chart) loadValuesYaml() error {
	var values map[string]interface{}
	err := loadYamlFile(c.path("values.yaml"), &values)
	if err != nil {
		return err
	}
	for k, v := range values {
		c.values[k] = toStarlark(v)
	}
	return nil
}

func (c *Chart) init(thread *starlark.Thread) error {
	file := c.path("Chart.star")
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil
	}
	globals, err := starlark.ExecFile(thread, file, nil, starlark.StringDict{
		"chart": starlark.NewBuiltin("chart", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
			return LoadChart(thread, c.repo, fn, args, kwargs)
		}),
	})
	if err != nil {
		return err
	}

	init, ok := globals["init"]

	if ok {
		_, err := starlark.Call(thread, init, starlark.Tuple{c}, nil)
		if err != nil {
			return err
		}
	}

	for k, v := range globals {
		if k == "init" {
			continue
		}
		f, ok := v.(starlark.Callable)
		if ok {
			c.methods[k] = starlark.NewBuiltin(f.Name(), func(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
				allArgs := make([]starlark.Value, args.Len()+1)
				allArgs[0] = c
				for i := 0; i < args.Len(); i++ {
					allArgs[i+1] = args.Index(i)
				}
				return f.CallInternal(thread, allArgs, kwargs)
			})
		}
	}
	return nil
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
		m, ok := c.methods[name]
		if !ok {
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

// Template -
func (c *Chart) Template(release *Release) (string, error) {
	var writer bytes.Buffer
	err := c.template(release, &writer, make(map[string]bool))
	if err != nil {
		return "", err
	}
	return writer.String(), nil
}

func (c *Chart) template(release *Release, writer io.Writer, done map[string]bool) error {
	if done[c.Name] {
		return nil
	}
	done[c.Name] = true
	for _, v := range c.values {
		subChart, ok := v.(*Chart)
		if ok {
			err := subChart.template(release, writer, done)
			if err != nil {
				return err
			}
		}
	}
	glob := c.path("templates", "*.yaml")
	filenames, err := filepath.Glob(glob)
	if err != nil {
		return err
	}
	if len(filenames) == 0 {
		return nil
	}
	helpers := c.path("templates", "_helpers.tpl")
	if _, err := os.Stat(helpers); err != nil {
		helpers = ""
	}

	values := toGo(c)
	for _, filename := range filenames {
		var buffer bytes.Buffer
		tpl := template.New(filepath.Base(filename))
		tpl = addTemplateFuncs(tpl)
		if helpers == "" {
			tpl, err = tpl.ParseFiles(filename)
		} else {
			tpl, err = tpl.ParseFiles(helpers, filename)
		}

		if err != nil {
			return err
		}
		err = tpl.Execute(&buffer, struct {
			Values  interface{}
			Chart   *Chart
			Release *Release
			Files   files
		}{
			Values:  values,
			Chart:   c,
			Release: release,
			Files:   files(make(map[string][]byte)),
		})
		if err != nil {
			return err
		}
		if buffer.Len() > 0 {
			content := strings.TrimSpace(buffer.String())
			if len(content) > 0 {
				if !strings.HasPrefix(content, "---") {
					writer.Write([]byte("---\n"))
				}
				writer.Write([]byte(content))
				writer.Write([]byte("\n"))
			}
		}

	}
	return nil
}

func loadYamlFile(filename string, value interface{}) error {
	reader, err := os.Open(filename) // For read access.
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("Unable to open file %s for parsing: %s", filename, err.Error())
	}
	defer reader.Close()
	decoder := yaml.NewDecoder(reader)
	err = decoder.Decode(value)
	if err != nil {
		return fmt.Errorf("Error during parsing file %s: %s", filename, err.Error())
	}
	return nil
}
