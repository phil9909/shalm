package impl

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("files", func() {

	It("works as expected", func() {
		dir := newTestDir()
		defer dir.Remove()
		dir.WriteFile("file1.yaml", []byte("1234"), 0644)
		dir.WriteFile("file2.yaml", []byte("aaaa"), 0644)
		f := files{dir: dir.Root()}
		content := f.Glob("*.yaml")
		Expect(content).To(HaveKeyWithValue("file1.yaml", []byte("1234")))
		Expect(content).To(HaveKeyWithValue("file2.yaml", []byte("aaaa")))

		Expect(f.Get("file2.yaml")).To(Equal("aaaa"))
	})
})
