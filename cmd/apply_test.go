package cmd

import (
	"bytes"
	"path"
	"path/filepath"
	"runtime"

	"github.com/kramerul/shalm/cmd/fakes"
	"github.com/kramerul/shalm/internal/pkg/chart/impl"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
	root       = path.Join(filepath.Dir(b), "..")
	example    = path.Join(root, "example")
)

var _ = Describe("Apply Chart", func() {

	Context("apply chart", func() {
		It("produces the correct output", func() {
			writer := bytes.Buffer{}
			k := &fakes.K8sFake{Writer: &writer}
			err := apply(impl.NewRepo(), impl.NewRootChartForDir("mynamespace", example, afero.NewOsFs()), "cf", impl.NewK8sValue(k))
			Expect(err).ToNot(HaveOccurred())
			output := writer.String()
			Expect(output).To(ContainSubstring("CREATE OR REPLACE USER 'uaa'"))
			Expect(k.RolloutStatusCalls).To(HaveLen(1))
			Expect(k.RolloutStatusCalls[0]).To(Equal("mariadb-master"))
			Expect(k.Namespace).To(Equal("mynamespace"))
		})
	})
})
