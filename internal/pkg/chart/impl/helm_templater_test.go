package impl

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("HelmTemplater", func() {

	Context("renders chart", func() {
		var h *HelmTemplater
		var fs afero.Fs

		BeforeEach(func() {
			var err error
			fs = afero.NewMemMapFs()
			fs.MkdirAll("templates", 0755)
			afero.WriteFile(fs, "templates/test.yaml", []byte("test: {{ .Value }}"), 0644)
			h, err = NewHelmTemplater(fs, ".")
			Expect(err).ToNot(HaveOccurred())
		})
		It("renders chart correct", func() {
			writer := &bytes.Buffer{}
			err := h.Template("", struct {
				Value string
			}{
				Value: "test",
			}, writer)
			Expect(err).ToNot(HaveOccurred())
			Expect(writer.String()).To(Equal("---\ntest: test\n"))
		})
	})
})
