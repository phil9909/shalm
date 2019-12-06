package shalm

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"

	"go.starlark.net/starlark"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type testDir string

func newTestDir() testDir {
	dir, err := ioutil.TempDir("", "shalm")
	if err != nil {
		panic(err)
	}
	return testDir(dir)
}

func (t testDir) Remove() error {
	return os.RemoveAll(string(t))
}

func (t testDir) Join(parts ...string) string {
	parts = append([]string{t.Root()}, parts...)
	return path.Join(parts...)
}

func (t testDir) Root() string {
	return string(t)
}

func (t testDir) MkdirAll(path string, mode os.FileMode) error {
	return os.MkdirAll(t.Join(path), mode)
}

func (t testDir) WriteFile(path string, content []byte, mode os.FileMode) error {
	return ioutil.WriteFile(t.Join(path), content, mode)
}

var _ = Describe("Chart", func() {

	Context("initialize", func() {

		It("reads Chart.yaml", func() {
			thread := &starlark.Thread{Name: "main"}
			dir := newTestDir()
			defer dir.Remove()
			repo := NewRepo()
			dir.WriteFile("Chart.yaml", []byte("name: mariadb\nversion: 6.12.2\n"), 0644)
			c, err := newChart(thread, repo, dir.Root(), "namespace", nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(c.GetName()).To(Equal("mariadb"))
		})

		It("reads values.yaml", func() {
			thread := &starlark.Thread{Name: "main"}
			dir := newTestDir()
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

		It("reads Chart.star", func() {
			thread := &starlark.Thread{Name: "main"}
			dir := newTestDir()
			defer dir.Remove()
			repo := NewRepo()
			dir.WriteFile("values.yaml", []byte("timeout: \"30s\"\n"), 0644)
			dir.WriteFile("Chart.star", []byte("def init(self):\n  self.timeout = \"60s\"\n"), 0644)
			c, err := newChart(thread, repo, dir.Root(), "namespace", nil, nil)
			Expect(err).NotTo(HaveOccurred())
			attr, err := c.Attr("timeout")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr.(starlark.String).GoString()).To(Equal("60s"))
		})

		It("templates a c ", func() {
			thread := &starlark.Thread{Name: "main"}
			dir := newTestDir()
			defer dir.Remove()
			repo := NewRepo()
			dir.MkdirAll("templates", 0755)
			dir.WriteFile("templates/deployment.yaml", []byte("namespace: {{ .Release.Namespace}}"), 0644)
			dir.WriteFile("Chart.yaml", []byte("name: mariadb\nversion: 6.12.2\n"), 0644)
			c, err := newChart(thread, repo, dir.Root(), "namespace", nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(c.GetName()).To(Equal("mariadb"))
			output, err := c.Template(thread)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal("---\nmetadata:\n  namespace: namespace\nnamespace: namespace\n"))
		})

		It("applies a c ", func() {
			thread := &starlark.Thread{Name: "main"}
			dir := newTestDir()
			defer dir.Remove()
			repo := NewRepo()
			dir.MkdirAll("templates", 0755)
			dir.WriteFile("templates/deployment.yaml", []byte("namespace: {{ .Release.Namespace}}"), 0644)
			dir.WriteFile("Chart.yaml", []byte("name: mariadb\nversion: 6.12.2\n"), 0644)
			c, err := newChart(thread, repo, dir.Root(), "namespace", nil, nil)
			Expect(err).NotTo(HaveOccurred())
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
			err = c.Apply(thread, k)
			Expect(err).NotTo(HaveOccurred())
			Expect(writer.String()).To(Equal("---\nmetadata:\n  namespace: namespace\nnamespace: namespace\n"))
		})

		It("deletes a c ", func() {
			thread := &starlark.Thread{Name: "main"}
			dir := newTestDir()
			defer dir.Remove()
			repo := NewRepo()
			dir.MkdirAll("templates", 0755)
			dir.WriteFile("templates/deployment.yaml", []byte("namespace: {{ .Release.Namespace}}"), 0644)
			dir.WriteFile("Chart.yaml", []byte("name: mariadb\nversion: 6.12.2\n"), 0644)
			c, err := newChart(thread, repo, dir.Root(), "namespace", nil, nil)
			Expect(err).NotTo(HaveOccurred())
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
			err = c.Delete(thread, k)
			Expect(err).NotTo(HaveOccurred())
			Expect(writer.String()).To(Equal("---\nmetadata:\n  namespace: namespace\nnamespace: namespace\n"))
		})

		It("applies subcharts", func() {
			thread := &starlark.Thread{Name: "main"}
			dir := newTestDir()
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

		It("behaves like starlark value", func() {
			thread := &starlark.Thread{Name: "main"}
			dir := newTestDir()
			defer dir.Remove()
			repo := NewRepo()
			dir.WriteFile("values.yaml", []byte("replicas: \"1\"\ntimeout: \"30s\"\n"), 0644)
			c, err := newChart(thread, repo, dir.Root(), "namespace", nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(c.String()).To(ContainSubstring("replicas = \"1\""))
			Expect(c.Hash()).NotTo(Equal(uint32(0)))
			Expect(c.Truth()).To(BeEquivalentTo(true))
		})

		It("applies a credentials ", func() {
			thread := &starlark.Thread{Name: "main"}
			dir := newTestDir()
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
			Expect(writer.String()).To(ContainSubstring("kind: secret"))
			Expect(writer.String()).To(ContainSubstring("type: Opaque"))
			Expect(writer.String()).To(ContainSubstring("  name: test"))
			Expect(writer.String()).To(ContainSubstring("  username: "))
			Expect(writer.String()).To(ContainSubstring("  password: "))
		})

	})
})
