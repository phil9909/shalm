package shalm

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/blang/semver"
	"github.com/kramerul/shalm/pkg/shalm/renderer"
	"go.starlark.net/starlark"
)

type chartImpl struct {
	Name            string
	Version         semver.Version
	values          map[string]starlark.Value
	methods         map[string]starlark.Callable
	dir             string
	namespace       string
	userCredentials []*userCredential
}

var (
	_ ChartValue = (*chartImpl)(nil)
)

func newChart(thread *starlark.Thread, repo Repo, dir string, namespace string, args starlark.Tuple, kwargs []starlark.Tuple) (*chartImpl, error) {
	name := strings.Split(filepath.Base(dir), ":")[0]
	c := &chartImpl{Name: name, dir: dir, namespace: namespace}
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
	return c, nil

}

func (c *chartImpl) GetName() string {
	return c.Name
}

func (c *chartImpl) GetVersion() semver.Version {
	return c.Version
}

func (c *chartImpl) walk(cb func(name string, size int64, body io.Reader, err error) error) error {
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
}

// Attr returns the value of the specified field.
func (c *chartImpl) Attr(name string) (starlark.Value, error) {
	if name == "namespace" {
		return starlark.String(c.namespace), nil
	}
	if name == "name" {
		return starlark.String(c.Name), nil
	}
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
	return wrapDict(value), nil
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
	c.values[name] = unwrapDict(val)
	return nil
}

func (c *chartImpl) applyFunction() starlark.Callable {
	return starlark.NewBuiltin("apply", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var k K8sValue
		if err := starlark.UnpackArgs("apply", args, kwargs, "k8s", &k); err != nil {
			return nil, err
		}
		return starlark.None, c.apply(thread, k)
	})
}

func (c *chartImpl) Apply(thread *starlark.Thread, k K8s) error {
	_, err := starlark.Call(thread, c.methods["apply"], starlark.Tuple{NewK8sValue(k)}, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *chartImpl) apply(thread *starlark.Thread, k K8sValue) error {
	err := c.eachSubChart(func(subChart *chartImpl) error {
		_, err := subChart.methods["apply"].CallInternal(thread, starlark.Tuple{k}, nil)
		return err
	})
	if err != nil {
		return err
	}
	return c.applyLocal(thread, k, &K8sOptions{}, &renderer.Options{})
}

func (c *chartImpl) applyLocalFunction() starlark.Callable {
	return starlark.NewBuiltin("__apply", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var k K8sValue
		parser := &kwargsParser{kwargs: kwargs}
		rendererOptionss := unpackRendererOptions(parser)
		k8sOptions := unpackK8sOptions(parser)
		if err := starlark.UnpackArgs("__apply", args, parser.Parse(), "k8s", &k); err != nil {
			return nil, err
		}
		return starlark.None, c.applyLocal(thread, k, k8sOptions, rendererOptionss)
	})
}

func (c *chartImpl) applyLocal(thread *starlark.Thread, k K8sValue, k8sOptions *K8sOptions, rendererOptions *renderer.Options) error {
	for _, credential := range c.userCredentials {
		err := credential.GetOrCreate(k)
		if err != nil {
			return err
		}
	}
	k8sOptions.Namespaced = false
	return k.Apply(func(writer io.Writer) error {
		return c.template(thread, writer, rendererOptions)
	}, k8sOptions)
}

func (c *chartImpl) Delete(thread *starlark.Thread, k K8s) error {
	_, err := starlark.Call(thread, c.methods["delete"], starlark.Tuple{NewK8sValue(k)}, nil)
	if err != nil {
		return err
	}
	return nil

}

func (c *chartImpl) deleteFunction() starlark.Callable {
	return starlark.NewBuiltin("delete", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var k K8sValue
		if err := starlark.UnpackArgs("delete", args, kwargs, "k8s", &k); err != nil {
			return nil, err
		}
		return starlark.None, c.delete(thread, k)
	})
}

func (c *chartImpl) delete(thread *starlark.Thread, k K8sValue) error {
	err := c.eachSubChart(func(subChart *chartImpl) error {
		_, err := subChart.methods["delete"].CallInternal(thread, starlark.Tuple{k}, nil)
		return err
	})
	if err != nil {
		return err
	}
	return c.deleteLocal(thread, k, &K8sOptions{}, &renderer.Options{})
}

func (c *chartImpl) deleteLocalFunction() starlark.Callable {
	return starlark.NewBuiltin("__delete", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
		var k K8sValue
		parser := &kwargsParser{kwargs: kwargs}
		rendererOptionss := unpackRendererOptions(parser)
		k8sOptions := unpackK8sOptions(parser)
		if err := starlark.UnpackArgs("__delete", args, parser.Parse(), "k8s", &k); err != nil {
			return nil, err
		}
		return starlark.None, c.deleteLocal(thread, k, k8sOptions, rendererOptionss)
	})
}

func (c *chartImpl) deleteLocal(thread *starlark.Thread, k K8sValue, k8sOptions *K8sOptions, rendererOptions *renderer.Options) error {
	rendererOptions.UninstallOrder = true
	k8sOptions.Namespaced = false
	return k.Delete(func(writer io.Writer) error {
		return c.template(thread, writer, rendererOptions)
	}, k8sOptions)
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

func (c *chartImpl) mergeValues(values map[string]interface{}) {
	for k, v := range values {
		c.values[k] = merge(c.values[k], toStarlark(v))
	}
}

func unpackRendererOptions(parser *kwargsParser) *renderer.Options {
	result := &renderer.Options{}
	parser.Arg("glob", func(value starlark.Value) {
		result.Glob = value.(starlark.String).GoString()
	})
	return result
}

func (c *chartImpl) Package(writer io.Writer) error {
	gz := gzip.NewWriter(writer)
	tw := tar.NewWriter(gz)
	defer tw.Close()
	defer gz.Close()
	return c.walk(func(file string, size int64, body io.Reader, err error) error {
		hdr := &tar.Header{
			Name: path.Join(c.Name, file),
			Mode: 0644,
			Size: size,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := io.Copy(tw, body); err != nil {
			return err
		}
		return nil
	})
}

var chartDirExpr = regexp.MustCompile("^[^/]*/")

func tarExtract(in io.Reader, dir string) error {
	reader := bufio.NewReader(in)
	testBytes, err := reader.Peek(64)
	if err != nil {
		return err
	}
	in = reader
	contentType := http.DetectContentType(testBytes)
	if strings.Contains(contentType, "x-gzip") {
		in, err = gzip.NewReader(in)
		if err != nil {
			return err
		}
	}
	tr := tar.NewReader(in)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}
		if hdr.FileInfo().IsDir() {
			continue
		}
		fn := path.Join(dir, chartDirExpr.ReplaceAllString(hdr.Name, ""))
		if err := os.MkdirAll(path.Dir(fn), 0755); err != nil {
			return err
		}
		out, err := os.Create(fn)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, tr); err != nil {
			log.Fatal(err)
		}
		out.Close()
	}
	return nil
}
