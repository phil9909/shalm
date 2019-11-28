package cmd

import (
	"bytes"
	"io"
	"path"
	"path/filepath"
	"runtime"

	"github.com/kramerul/shalm/internal/pkg/chart/api"
	"github.com/kramerul/shalm/internal/pkg/chart/fakes"

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
			k := &fakes.FakeK8s{
				ApplyStub: func(i func(io.Writer) error, options *api.K8sOptions) error {
					i(&writer)
					return nil
				},
			}
			k.ForNamespaceStub = func(s string) api.K8s {
				return k
			}

			err := apply(impl.NewRepo(), impl.NewRootChartForDir("mynamespace", example, afero.NewOsFs()), "cf", impl.NewK8sValue(k))
			Expect(err).ToNot(HaveOccurred())
			output := writer.String()
			Expect(output).To(ContainSubstring("CREATE OR REPLACE USER 'uaa'"))
			Expect(k.RolloutStatusCallCount()).To(Equal(1))
			Expect(k.ForNamespaceCallCount()).To(Equal(3))
			Expect(k.ForNamespaceArgsForCall(0)).To(Equal("mynamespace"))
			Expect(k.ForNamespaceArgsForCall(1)).To(Equal("mynamespace"))
			Expect(k.ForNamespaceArgsForCall(2)).To(Equal("uaa"))
			kind, name, _ := k.RolloutStatusArgsForCall(0)
			Expect(name).To(Equal("mariadb-master"))
			Expect(kind).To(Equal("statefulset"))
		})
	})
})
