package cmd

import (
	"bytes"
	"io"
	"path"

	"github.com/kramerul/shalm/pkg/shalm"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delete Chart", func() {

	It("produces the correct output", func() {
		writer := bytes.Buffer{}
		k := &shalm.FakeK8s{
			DeleteStub: func(i func(io.Writer) error, options *shalm.K8sOptions) error {
				i(&writer)
				return nil
			},
		}
		k.ForNamespaceStub = func(s string) shalm.K8s {
			return k
		}

		err := delete(path.Join(example, "cf"), "mynamespace", shalm.NewK8sValue(k))
		Expect(err).ToNot(HaveOccurred())
		output := writer.String()
		Expect(output).To(ContainSubstring("CREATE OR REPLACE USER 'uaa'"))
		Expect(k.DeleteCallCount()).To(Equal(3))
	})
})
