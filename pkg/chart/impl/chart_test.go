package impl

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/kramerul/shalm/pkg/chart/api"
	"github.com/kramerul/shalm/pkg/chart/fakes"
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
			thread := &starlark.Thread{Name: "my thread"}
			dir := newTestDir()
			defer dir.Remove()
			repo := NewRepo()
			dir.WriteFile("Chart.yaml", []byte("name: mariadb\nversion: 6.12.2\n"), 0644)
			chart, err := NewChart(thread, repo, ".", NewRootChartForDir("namespace", dir.Root()), nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(chart.GetName()).To(Equal("mariadb"))
		})

		It("reads values.yaml", func() {
			thread := &starlark.Thread{Name: "my thread"}
			dir := newTestDir()
			defer dir.Remove()
			repo := NewRepo()
			dir.WriteFile("values.yaml", []byte("replicas: \"1\"\ntimeout: \"30s\"\n"), 0644)
			chart, err := NewChart(thread, repo, ".", NewRootChartForDir("namespace", dir.Root()), nil, nil)
			Expect(err).NotTo(HaveOccurred())
			attr, err := chart.Attr("replicas")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr.(starlark.String).GoString()).To(Equal("1"))
			attr, err = chart.Attr("timeout")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr.(starlark.String).GoString()).To(Equal("30s"))
		})

		It("reads Chart.star", func() {
			thread := &starlark.Thread{Name: "my thread"}
			dir := newTestDir()
			defer dir.Remove()
			repo := NewRepo()
			dir.WriteFile("values.yaml", []byte("timeout: \"30s\"\n"), 0644)
			dir.WriteFile("Chart.star", []byte("def init(self):\n  self.timeout = \"60s\"\n"), 0644)
			chart, err := NewChart(thread, repo, ".", NewRootChartForDir("namespace", dir.Root()), nil, nil)
			Expect(err).NotTo(HaveOccurred())
			attr, err := chart.Attr("timeout")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr.(starlark.String).GoString()).To(Equal("60s"))
		})

		It("templates a chart ", func() {
			thread := &starlark.Thread{Name: "my thread"}
			dir := newTestDir()
			defer dir.Remove()
			repo := NewRepo()
			dir.MkdirAll("templates", 0755)
			dir.WriteFile("templates/deployment.yaml", []byte("namespace: {{ .Release.Namespace}}"), 0644)
			dir.WriteFile("Chart.yaml", []byte("name: mariadb\nversion: 6.12.2\n"), 0644)
			chart, err := NewChart(thread, repo, ".", NewRootChartForDir("namespace", dir.Root()), nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(chart.GetName()).To(Equal("mariadb"))
			output, err := chart.Template(thread)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal("---\nnamespace: namespace\n"))
		})

		It("applies a chart ", func() {
			thread := &starlark.Thread{Name: "my thread"}
			dir := newTestDir()
			defer dir.Remove()
			repo := NewRepo()
			dir.MkdirAll("templates", 0755)
			dir.WriteFile("templates/deployment.yaml", []byte("namespace: {{ .Release.Namespace}}"), 0644)
			dir.WriteFile("Chart.yaml", []byte("name: mariadb\nversion: 6.12.2\n"), 0644)
			chart, err := NewChart(thread, repo, ".", NewRootChartForDir("namespace", dir.Root()), nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(chart.GetName()).To(Equal("mariadb"))
			writer := bytes.Buffer{}
			k := &fakes.FakeK8s{
				ApplyStub: func(i func(io.Writer) error, options *api.K8sOptions) error {
					i(&writer)
					return nil
				},
			}
			k.ForNamespaceStub = func(s string) api.K8s {
				return k
			}
			err = chart.Apply(thread, k)
			Expect(err).NotTo(HaveOccurred())
			Expect(writer.String()).To(Equal("---\nnamespace: namespace\n"))
		})

		It("deletes a chart ", func() {
			thread := &starlark.Thread{Name: "my thread"}
			dir := newTestDir()
			defer dir.Remove()
			repo := NewRepo()
			dir.MkdirAll("templates", 0755)
			dir.WriteFile("templates/deployment.yaml", []byte("namespace: {{ .Release.Namespace}}"), 0644)
			dir.WriteFile("Chart.yaml", []byte("name: mariadb\nversion: 6.12.2\n"), 0644)
			chart, err := NewChart(thread, repo, ".", NewRootChartForDir("namespace", dir.Root()), nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(chart.GetName()).To(Equal("mariadb"))
			writer := bytes.Buffer{}
			k := &fakes.FakeK8s{
				DeleteStub: func(i func(io.Writer) error, options *api.K8sOptions) error {
					i(&writer)
					return nil
				},
			}
			k.ForNamespaceStub = func(s string) api.K8s {
				return k
			}
			err = chart.Delete(thread, k)
			Expect(err).NotTo(HaveOccurred())
			Expect(writer.String()).To(Equal("---\nnamespace: namespace\n"))
		})

		It("applies subcharts", func() {
			thread := &starlark.Thread{Name: "my thread"}
			dir := newTestDir()
			defer dir.Remove()
			repo := NewRepo()
			dir.MkdirAll("chart1/templates", 0755)
			dir.MkdirAll("chart2/templates", 0755)
			dir.WriteFile("chart1/Chart.star", []byte("def init(self):\n  self.chart2 = chart(\"../chart2\",namespace=\"chart2\")\n"), 0644)

			dir.WriteFile("chart2/templates/deployment.yaml", []byte("namespace: {{ .Release.Namespace}}"), 0644)
			chart, err := NewChart(thread, repo, "chart1", NewRootChartForDir("namespace", dir.Root()), nil, nil)
			Expect(err).NotTo(HaveOccurred())
			writer := bytes.Buffer{}
			k := &fakes.FakeK8s{
				DeleteStub: func(i func(io.Writer) error, options *api.K8sOptions) error {
					i(&writer)
					return nil
				},
			}
			k.ForNamespaceStub = func(s string) api.K8s {
				return k
			}
			err = chart.Delete(thread, k)
			Expect(err).NotTo(HaveOccurred())
			Expect(writer.String()).To(Equal("---\nnamespace: namespace\n"))
		})

		It("behaves like starlark value", func() {
			thread := &starlark.Thread{Name: "my thread"}
			dir := newTestDir()
			defer dir.Remove()
			repo := NewRepo()
			dir.WriteFile("values.yaml", []byte("replicas: \"1\"\ntimeout: \"30s\"\n"), 0644)
			chart, err := NewChart(thread, repo, ".", NewRootChartForDir("namespace", dir.Root()), nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(chart.String()).To(ContainSubstring("replicas = \"1\""))
			Expect(chart.Hash()).NotTo(Equal(uint32(0)))
			Expect(chart.Truth()).To(BeEquivalentTo(true))
		})

	})
})
