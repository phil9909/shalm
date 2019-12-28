package shalm

import (
	"bytes"
	"errors"
	"io"

	"go.starlark.net/starlark"

	"github.com/blang/semver"
	. "github.com/kramerul/shalm/pkg/shalm/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Chart", func() {

	Context("initialize", func() {

		It("reads Chart.yaml", func() {
			thread := &starlark.Thread{Name: "main"}
			dir := NewTestDir()
			defer dir.Remove()
			repo := NewRepo()
			dir.WriteFile("Chart.yaml", []byte("name: mariadb\nversion: 6.12.2\n"), 0644)
			c, err := newChart(thread, repo, dir.Root(), "namespace", nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(c.GetName()).To(Equal("mariadb"))
		})
		It("reads Chart.yaml 'v' prefix in version", func() {
			thread := &starlark.Thread{Name: "main"}
			dir := NewTestDir()
			defer dir.Remove()
			repo := NewRepo()
			dir.WriteFile("Chart.yaml", []byte("name: mariadb\nversion: v6.12.2\n"), 0644)
			c, err := newChart(thread, repo, dir.Root(), "namespace", nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(c.GetName()).To(Equal("mariadb"))
			Expect(c.GetVersion()).To(Equal(semver.Version{Major: 6, Minor: 12, Patch: 2}))
		})

		It("reads values.yaml", func() {
			thread := &starlark.Thread{Name: "main"}
			dir := NewTestDir()
			defer dir.Remove()
			repo := NewRepo()
			dir.WriteFile("values.yaml", []byte("replicas: \"1\"\ntimeout: \"30s\"\n"), 0644)
			c, err := newChart(thread, repo, dir.Root(), "namespace", nil, nil)
			Expect(err).NotTo(HaveOccurred())
			attr, err := c.Attr("replicas")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr.(starlark.String).GoString()).To(Equal("1"))
			attr, err = c.Attr("timeout")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr.(starlark.String).GoString()).To(Equal("30s"))
		})

	})
	Context("Chart.start", func() {
		var dir TestDir
		var c ChartValue
		thread := &starlark.Thread{Name: "main"}
		BeforeEach(func() {
			dir = NewTestDir()
			repo := NewRepo()
			dir.WriteFile("values.yaml", []byte("timeout: \"30s\"\n"), 0644)
			dir.WriteFile("Chart.star", []byte(`
def init(self):
	self.timeout = "60s"
	k8s('test')
def method(self):
	return self.namespace
def apply(self,k8s):
	return self.__apply(k8s)
def delete(self,k8s):
	return self.__delete(k8s)
`),
				0644)
			var err error
			c, err = newChart(thread, repo, dir.Root(), "namespace", nil, nil)
			Expect(err).NotTo(HaveOccurred())

		})
		AfterEach(func() {
			dir.Remove()
		})
		It("evalutes constructor", func() {
			attr, err := c.Attr("timeout")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr.(starlark.String).GoString()).To(Equal("60s"))
		})
		It("binds method to self", func() {
			attr, err := c.Attr("method")
			Expect(err).NotTo(HaveOccurred())
			value, err := starlark.Call(thread, attr.(starlark.Callable), nil, nil)
			Expect(value.(starlark.String).GoString()).To(Equal("namespace"))
		})
		It("overrides apply", func() {
			attr, err := c.Attr("apply")
			Expect(err).NotTo(HaveOccurred())
			k := &FakeK8s{}
			k.ForNamespaceStub = func(s string) K8s {
				return k
			}
			_, err = starlark.Call(thread, attr.(starlark.Callable), starlark.Tuple{NewK8sValue(k)}, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(k.ApplyCallCount()).To(Equal(1))
		})
		It("overrides delete", func() {
			attr, err := c.Attr("delete")
			Expect(err).NotTo(HaveOccurred())
			k := &FakeK8s{}
			k.ForNamespaceStub = func(s string) K8s {
				return k
			}
			_, err = starlark.Call(thread, attr.(starlark.Callable), starlark.Tuple{NewK8sValue(k)}, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(k.DeleteCallCount()).To(Equal(1))
		})
	})
	Context("methods", func() {
		var dir TestDir
		var c ChartValue
		thread := &starlark.Thread{Name: "main"}

		BeforeEach(func() {
			dir = NewTestDir()
			dir.MkdirAll("templates", 0755)
			dir.WriteFile("templates/deployment.yaml", []byte("namespace: {{ .Release.Namespace}}"), 0644)
			dir.WriteFile("Chart.yaml", []byte("name: mariadb\nversion: 6.12.2\n"), 0644)
			repo := NewRepo()
			var err error
			c, err = newChart(thread, repo, dir.Root(), "namespace", nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(c.GetName()).To(Equal("mariadb"))

		})
		AfterEach(func() {
			dir.Remove()
		})
		It("templates a chart", func() {
			defer dir.Remove()
			Expect(c.GetName()).To(Equal("mariadb"))
			output, err := c.Template(thread)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal("---\nmetadata:\n  namespace: namespace\nnamespace: namespace\n"))
		})

		It("applies a chart", func() {
			Expect(c.GetName()).To(Equal("mariadb"))
			writer := bytes.Buffer{}
			k := &FakeK8s{
				ApplyStub: func(i func(io.Writer) error, options *K8sOptions) error {
					i(&writer)
					return nil
				},
			}
			k.ForNamespaceStub = func(s string) K8s {
				return k
			}
			err := c.Apply(thread, k)
			Expect(err).NotTo(HaveOccurred())
			Expect(writer.String()).To(Equal("---\nmetadata:\n  namespace: namespace\nnamespace: namespace\n"))
		})

		It("deletes a chart", func() {
			Expect(c.GetName()).To(Equal("mariadb"))
			writer := bytes.Buffer{}
			k := &FakeK8s{
				DeleteStub: func(i func(io.Writer) error, options *K8sOptions) error {
					i(&writer)
					return nil
				},
			}
			k.ForNamespaceStub = func(s string) K8s {
				return k
			}
			err := c.Delete(thread, k)
			Expect(err).NotTo(HaveOccurred())
			Expect(writer.String()).To(Equal("---\nmetadata:\n  namespace: namespace\nnamespace: namespace\n"))
		})

		It("packages a chart", func() {
			writer := &bytes.Buffer{}
			err := c.Package(writer)
			Expect(err).NotTo(HaveOccurred())
			Expect(bytes.HasPrefix(writer.Bytes(), []byte{0x1F, 0x8B, 0x08})).To(BeTrue())
		})

		It("applies subcharts", func() {
			thread := &starlark.Thread{Name: "main"}
			dir := NewTestDir()
			defer dir.Remove()
			repo := NewRepo()
			dir.MkdirAll("chart1/templates", 0755)
			dir.MkdirAll("chart2/templates", 0755)
			dir.WriteFile("chart1/Chart.star", []byte("def init(self):\n  self.chart2 = chart(\"../chart2\",namespace=\"chart2\")\n"), 0644)

			dir.WriteFile("chart2/templates/deployment.yaml", []byte("namespace: {{ .Release.Namespace}}"), 0644)
			c, err := newChart(thread, repo, dir.Join("chart1"), "namespace", nil, nil)
			Expect(err).NotTo(HaveOccurred())
			writer := bytes.Buffer{}
			k := &FakeK8s{
				DeleteStub: func(i func(io.Writer) error, options *K8sOptions) error {
					i(&writer)
					return nil
				},
			}
			k.ForNamespaceStub = func(s string) K8s {
				return k
			}
			err = c.Delete(thread, k)
			Expect(err).NotTo(HaveOccurred())
			Expect(k.DeleteCallCount()).To(Equal(2))
			Expect(writer.String()).To(Equal("---\nmetadata:\n  namespace: chart2\nnamespace: chart2\n"))
		})

	})
	It("behaves like starlark value", func() {
		thread := &starlark.Thread{Name: "main"}
		dir := NewTestDir()
		defer dir.Remove()
		repo := NewRepo()
		dir.WriteFile("values.yaml", []byte("replicas: \"1\"\ntimeout: \"30s\"\n"), 0644)
		c, err := newChart(thread, repo, dir.Root(), "namespace", nil, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(c.String()).To(ContainSubstring("replicas = \"1\""))
		Expect(c.Hash()).NotTo(Equal(uint32(0)))
		Expect(c.Truth()).To(BeEquivalentTo(true))
		Expect(c.Type()).To(Equal("chart"))
		value, err := c.Attr("name")
		Expect(err).NotTo(HaveOccurred())
		Expect(value.(starlark.String).GoString()).To(ContainSubstring("shalm"))
		value, err = c.Attr("namespace")
		Expect(err).NotTo(HaveOccurred())
		Expect(value.(starlark.String).GoString()).To(Equal("namespace"))
		value, err = c.Attr("apply")
		Expect(err).NotTo(HaveOccurred())
		Expect(value.(starlark.Callable).Name()).To(Equal("apply"))
		value, err = c.Attr("unknown")
		Expect(err).To(HaveOccurred())
	})

	It("applies a credentials ", func() {
		thread := &starlark.Thread{Name: "main"}
		dir := NewTestDir()
		defer dir.Remove()
		repo := NewRepo()
		dir.WriteFile("Chart.star", []byte("def init(self):\n  user_credential(\"test\")\n"), 0644)
		c, err := newChart(thread, repo, dir.Root(), "namespace", nil, nil)
		Expect(err).NotTo(HaveOccurred())
		writer := bytes.Buffer{}
		k := &FakeK8s{
			ApplyStub: func(i func(io.Writer) error, options *K8sOptions) error {
				i(&writer)
				return nil
			},
			GetStub: func(kind string, name string, writer io.Writer, k8s *K8sOptions) error {
				return errors.New("NotFound")
			},
			IsNotExistStub: func(err error) bool {
				return true
			},
		}
		k.ForNamespaceStub = func(s string) K8s {
			return k
		}
		err = c.Apply(thread, k)
		Expect(err).NotTo(HaveOccurred())
		Expect(writer.String()).To(ContainSubstring("apiVersion: v1"))
		Expect(writer.String()).To(ContainSubstring("kind: Secret"))
		Expect(writer.String()).To(ContainSubstring("type: Opaque"))
		Expect(writer.String()).To(ContainSubstring("  name: test"))
		Expect(writer.String()).To(ContainSubstring("  username: "))
		Expect(writer.String()).To(ContainSubstring("=="))
		Expect(writer.String()).To(ContainSubstring("  password: "))
	})

	It("merges values ", func() {
		thread := &starlark.Thread{Name: "main"}
		dir := NewTestDir()
		defer dir.Remove()
		repo := NewRepo()
		dir.WriteFile("Chart.star", []byte("def init(self):\n  self.timeout=50\n"), 0644)
		c, err := newChart(thread, repo, dir.Root(), "namespace", nil, nil)
		Expect(err).NotTo(HaveOccurred())
		c.mergeValues(map[string]interface{}{"timeout": 60, "string": "test"})
		Expect(c.values["timeout"]).To(Equal(starlark.MakeInt(60)))
		Expect(c.values["string"]).To(Equal(starlark.String("test")))
	})

})
