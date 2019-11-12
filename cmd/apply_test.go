package cmd

import (
	"bytes"

	"github.com/kramerul/shalm/internal/pkg/repo"

	"github.com/kramerul/shalm/cmd/fakes"
	"github.com/kramerul/shalm/internal/pkg/chart"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Apply Chart", func() {

	Context("apply chart", func() {
		It("produces the correct output", func() {
			writer := bytes.Buffer{}
			k8s := &fakes.K8sFake{Writer: &writer}
			err := apply(&repo.LocalRepo{BaseDir: "../example"}, "cf", k8s,
				&chart.Release{Name: "cf", Namespace: "namespace", Service: "cf"})
			Expect(err).ToNot(HaveOccurred())
			output := writer.String()
			Expect(output).To(ContainSubstring("CREATE OR REPLACE USER 'uaa'"))
			Expect(k8s.Methods).To(HaveLen(1))
			Expect(k8s.Methods[0]).To(Equal("rollout_status"))
		})
	})
})
