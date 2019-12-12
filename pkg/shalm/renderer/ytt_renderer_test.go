package renderer

import (
	"bytes"

	. "github.com/kramerul/shalm/pkg/shalm/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.starlark.net/starlark"
)

var _ = Describe("ytt", func() {

	It("template file is working", func() {
		dir := NewTestDir()
		defer dir.Remove()
		dir.WriteFile("ytt.yaml", []byte("test: #@ self\n"), 0644)
		out := &bytes.Buffer{}
		err := yttRenderFile(starlark.String("hello"), dir.Join("ytt.yaml"), out)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(out.Bytes())).To(Equal("test: hello\n"))
	})

	It("template is working", func() {
		in := bytes.NewBuffer([]byte("test: #@ self\n"))
		out := &bytes.Buffer{}
		err := yttRender(starlark.String("hello"), in, "stdin", out)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(out.Bytes())).To(Equal("test: hello\n"))
	})
})