package impl

import (
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
	})
})
