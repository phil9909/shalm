package renderer

import (
	"bytes"

	. "github.com/kramerul/shalm/pkg/shalm/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("helmRenderer", func() {

	Context("renders chart", func() {

		It("renders chart correct", func() {
			var err error
			dir := NewTestDir()
			defer dir.Remove()
			dir.WriteFile("test.yaml", []byte("test: {{ .Value }}"), 0644)
			helmFileRenderer, err := HelmFileRenderer(dir.Root(), struct {
				Value string
			}{
				Value: "test",
			})
			Expect(err).ToNot(HaveOccurred())
			writer := &bytes.Buffer{}
			err = helmFileRenderer(dir.Join("test.yaml"), writer)
			Expect(err).ToNot(HaveOccurred())
			Expect(writer.String()).To(Equal("test: test"))
		})

		It("toYaml works corret", func() {
			var err error
			dir := NewTestDir()
			defer dir.Remove()
			dir.WriteFile("test.yaml", []byte("test:\n{{ .Value | toYaml | indent 2}}"), 0644)
			helmFileRenderer, err := HelmFileRenderer(dir.Root(), struct {
				Value map[string]string
			}{
				Value: map[string]string{"key": "value"},
			})
			Expect(err).ToNot(HaveOccurred())
			writer := &bytes.Buffer{}
			err = helmFileRenderer(dir.Join("test.yaml"), writer)
			Expect(err).ToNot(HaveOccurred())
			Expect(writer.String()).To(Equal("test:\n  key: value"))
		})

		It("toJson works corret", func() {
			var err error
			dir := NewTestDir()
			defer dir.Remove()
			dir.WriteFile("test.yaml", []byte("test: {{ .Value | toJson }}"), 0644)
			helmFileRenderer, err := HelmFileRenderer(dir.Root(), struct {
				Value map[string]string
			}{
				Value: map[string]string{"key": "value"},
			})
			Expect(err).ToNot(HaveOccurred())
			writer := &bytes.Buffer{}
			err = helmFileRenderer(dir.Join("test.yaml"), writer)
			Expect(err).ToNot(HaveOccurred())
			Expect(writer.String()).To(Equal("test: {\"key\":\"value\"}"))
		})

		It("it loads helpers", func() {
			var err error
			dir := NewTestDir()
			defer dir.Remove()
			dir.MkdirAll("templates", 0755)
			dir.WriteFile("templates/test.yaml", []byte("test: {{ template \"chart\" }}"), 0644)
			dir.WriteFile("templates/_helpers.tpl", []byte(`
{{- define "chart" -}}
{{- printf "%s-%s" "chart" "version" | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}
`), 0644)
			helmFileRenderer, err := HelmFileRenderer(dir.Root(), struct {
				Value string
			}{
				Value: "test",
			})
			Expect(err).ToNot(HaveOccurred())
			writer := &bytes.Buffer{}
			err = helmFileRenderer(dir.Join("templates/test.yaml"), writer)
			Expect(err).ToNot(HaveOccurred())
			Expect(writer.String()).To(Equal("test: chart-version"))
		})
	})
})
