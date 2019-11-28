package impl

import (
	"bytes"
	"io"

	"github.com/kramerul/shalm/internal/pkg/chart/api"
	"github.com/kramerul/shalm/internal/pkg/chart/fakes"
	"go.starlark.net/starlark"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Chart", func() {

	Context("initialize", func() {
		var fs afero.Fs

		It("reads Chart.yaml", func() {
			thread := &starlark.Thread{Name: "my thread"}
			fs = afero.NewMemMapFs()
			repo := NewRepo()
			afero.WriteFile(fs, "Chart.yaml", []byte("name: mariadb\nversion: 6.12.2\n"), 0644)
			chart, err := NewChart(thread, repo, ".", NewRootChartForDir("namespace", ".", fs), nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(chart.GetName()).To(Equal("mariadb"))
		})

		It("reads values.yaml", func() {
			thread := &starlark.Thread{Name: "my thread"}
			fs = afero.NewMemMapFs()
			repo := NewRepo()
			afero.WriteFile(fs, "values.yaml", []byte("replicas: \"1\"\ntimeout: \"30s\"\n"), 0644)
			chart, err := NewChart(thread, repo, ".", NewRootChartForDir("namespace", ".", fs), nil, nil)
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
			fs = afero.NewOsFs()
			// starlark can only read from plain filesystem
			repo := NewRepo()
			afero.WriteFile(fs, "/tmp/values.yaml", []byte("timeout: \"30s\"\n"), 0644)
			afero.WriteFile(fs, "/tmp/Chart.star", []byte("def init(self):\n  self.timeout = \"60s\"\n"), 0644)
			chart, err := NewChart(thread, repo, "/tmp", NewRootChartForDir("namespace", "/", fs), nil, nil)
			Expect(err).NotTo(HaveOccurred())
			attr, err := chart.Attr("timeout")
			Expect(err).NotTo(HaveOccurred())
			Expect(attr.(starlark.String).GoString()).To(Equal("60s"))
		})

		It("templates a chart ", func() {
			thread := &starlark.Thread{Name: "my thread"}
			fs = afero.NewMemMapFs()
			repo := NewRepo()
			fs.MkdirAll("templates", 0755)
			afero.WriteFile(fs, "templates/deployment.yaml", []byte("namespace: {{ .Release.Namespace}}"), 0644)
			afero.WriteFile(fs, "Chart.yaml", []byte("name: mariadb\nversion: 6.12.2\n"), 0644)
			chart, err := NewChart(thread, repo, ".", NewRootChartForDir("namespace", ".", fs), nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(chart.GetName()).To(Equal("mariadb"))
			output, err := chart.Template(thread)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal("---\nnamespace: namespace\n"))
		})

		It("applies a chart ", func() {
			thread := &starlark.Thread{Name: "my thread"}
			fs = afero.NewMemMapFs()
			repo := NewRepo()
			fs.MkdirAll("templates", 0755)
			afero.WriteFile(fs, "templates/deployment.yaml", []byte("namespace: {{ .Release.Namespace}}"), 0644)
			afero.WriteFile(fs, "Chart.yaml", []byte("name: mariadb\nversion: 6.12.2\n"), 0644)
			chart, err := NewChart(thread, repo, ".", NewRootChartForDir("namespace", ".", fs), nil, nil)
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
			fs = afero.NewMemMapFs()
			repo := NewRepo()
			fs.MkdirAll("templates", 0755)
			afero.WriteFile(fs, "templates/deployment.yaml", []byte("namespace: {{ .Release.Namespace}}"), 0644)
			afero.WriteFile(fs, "Chart.yaml", []byte("name: mariadb\nversion: 6.12.2\n"), 0644)
			chart, err := NewChart(thread, repo, ".", NewRootChartForDir("namespace", ".", fs), nil, nil)
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
			fs = afero.NewOsFs()
			repo := NewRepo()
			fs.MkdirAll("/tmp/chart1/templates", 0755)
			fs.MkdirAll("/tmp/chart2/templates", 0755)
			afero.WriteFile(fs, "/tmp/chart1/Chart.star", []byte("def init(self):\n  self.chart2 = chart(\"../chart2\",namespace=\"chart2\")\n"), 0644)

			afero.WriteFile(fs, "/tmp/chart2/templates/deployment.yaml", []byte("namespace: {{ .Release.Namespace}}"), 0644)
			chart, err := NewChart(thread, repo, "/tmp/chart1", NewRootChartForDir("namespace", "/", fs), nil, nil)
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
			fs = afero.NewMemMapFs()
			repo := NewRepo()
			afero.WriteFile(fs, "values.yaml", []byte("replicas: \"1\"\ntimeout: \"30s\"\n"), 0644)
			chart, err := NewChart(thread, repo, ".", NewRootChartForDir("namespace", ".", fs), nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(chart.String()).To(ContainSubstring("replicas = \"1\""))
			Expect(chart.Hash()).NotTo(Equal(uint32(0)))
			Expect(chart.Truth()).To(BeEquivalentTo(true))
		})

	})
})
