package cmd

import (
	"bytes"

	"github.com/kramerul/shalm/cmd/fakes"
	"github.com/kramerul/shalm/internal/pkg/chart"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Apply Chart", func() {

	Context("apply chart", func() {
		It("produces the correct output", func() {
			writer := bytes.Buffer{}
			err := apply(&chart.LocalRepo{BaseDir: "../example"}, "cf", &fakes.K8sFake{Writer: &writer},
				&chart.Release{Name: "cf", Namespace: "namespace", Service: "cf"})
			Expect(err).ToNot(HaveOccurred())
			output := writer.String()
			Expect(output).To(ContainSubstring("CREATE OR REPLACE USER 'uaa'"))
		})
	})
})
