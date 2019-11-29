package impl

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HelmTemplater", func() {

	Context("renders chart", func() {
		var h *HelmTemplater

		It("renders chart correct", func() {
			var err error
			dir := newTestDir()
			defer dir.Remove()
			dir.MkdirAll("templates", 0755)
			dir.WriteFile("templates/test.yaml", []byte("test: {{ .Value }}"), 0644)
			h, err = NewHelmTemplater(dir.Root(), "namespace")
			Expect(err).ToNot(HaveOccurred())
			writer := &bytes.Buffer{}
			err = h.Template(struct {
				Value string
			}{
				Value: "test",
			}, writer, &HelmOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(writer.String()).To(Equal("---\ntest: test\n"))
		})
		It("it loads helpers", func() {
			var err error
			dir := newTestDir()
			defer dir.Remove()
			dir.MkdirAll("templates", 0755)
			dir.WriteFile("templates/test.yaml", []byte("test: {{ template \"chart\" }}"), 0644)
			dir.WriteFile("templates/_helpers.tpl", []byte(`
{{- define "chart" -}}
{{- printf "%s-%s" "chart" "version" | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}
`), 0644)
			h, err = NewHelmTemplater(dir.Root(), "namespace")
			Expect(err).ToNot(HaveOccurred())
			writer := &bytes.Buffer{}
			err = h.Template(struct {
				Value string
			}{
				Value: "test",
			}, writer, &HelmOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(writer.String()).To(Equal("---\ntest: chart-version\n"))
		})
		It("renders multipe files", func() {
			var err error
			dir := newTestDir()
			defer dir.Remove()
			dir.MkdirAll("templates", 0755)
			dir.WriteFile("templates/test1.yaml", []byte("test: test1"), 0644)
			dir.WriteFile("templates/test2.yaml", []byte("test: test2"), 0644)
			h, err = NewHelmTemplater(dir.Root(), "namespace")
			Expect(err).ToNot(HaveOccurred())
			writer := &bytes.Buffer{}
			err = h.Template(struct {
				Value string
			}{
				Value: "test",
			}, writer, &HelmOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(writer.String()).To(Equal("---\ntest: test1\n---\ntest: test2\n"))
		})
		It("repects glob patterns", func() {
			var err error
			dir := newTestDir()
			defer dir.Remove()
			dir.MkdirAll("templates", 0755)
			dir.WriteFile("templates/test1.yaml", []byte("test: test1"), 0644)
			dir.WriteFile("templates/test3.yaml", []byte("test: test2"), 0644)
			h, err = NewHelmTemplater(dir.Root(), "namespace")
			Expect(err).ToNot(HaveOccurred())
			writer := &bytes.Buffer{}
			err = h.Template(struct {
				Value string
			}{
				Value: "test",
			}, writer, &HelmOptions{glob: "*[1-2].yaml"})
			Expect(err).ToNot(HaveOccurred())
			Expect(writer.String()).To(Equal("---\ntest: test1\n"))
		})
		FIt("sorts by kind", func() {
			var err error
			dir := newTestDir()
			defer dir.Remove()
			dir.MkdirAll("templates", 0755)
			dir.WriteFile("templates/test1.yaml", []byte("kind: Other"), 0644)
			dir.WriteFile("templates/test2.yaml", []byte("kind: StatefulSet"), 0644)
			dir.WriteFile("templates/test3.yaml", []byte("kind: Service"), 0644)
			h, err = NewHelmTemplater(dir.Root(), "namespace")
			Expect(err).ToNot(HaveOccurred())

			By("Sorts in install order")
			writer := &bytes.Buffer{}
			err = h.Template(struct {
				Value string
			}{
				Value: "test",
			}, writer, &HelmOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(writer.String()).To(Equal(`---
metadata:
  namespace: namespace
kind: Service
---
metadata:
  namespace: namespace
kind: StatefulSet
---
metadata:
  namespace: namespace
kind: Other
`))

			By("Sorts in uninstall order")
			writer = &bytes.Buffer{}
			err = h.Template(struct {
				Value string
			}{
				Value: "test",
			}, writer, &HelmOptions{uninstallOrder: true})
			Expect(err).ToNot(HaveOccurred())
			Expect(writer.String()).To(Equal(`---
metadata:
  namespace: namespace
kind: Other
---
metadata:
  namespace: namespace
kind: StatefulSet
---
metadata:
  namespace: namespace
kind: Service
`))
		})
	})
})
