package shalm

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("chartClass", func() {

	It("behaves like starlark value", func() {
		cc := &chartClass{Name: "xxx"}
		Expect(cc.String()).To(ContainSubstring("xxx"))
		Expect(cc.Type()).To(Equal("chart_class"))
		Expect(func() { cc.Hash() }).Should(Panic())
		Expect(cc.Truth()).To(BeEquivalentTo(false))
		Expect(cc.AttrNames()).To(ConsistOf("api_version", "name", "version", "description", "keywords", "home", "sources", "icon"))
		for _, attribute := range cc.AttrNames() {
			_, err := cc.Attr(attribute)
			Expect(err).NotTo(HaveOccurred())
		}
		_, err := cc.Attr("unknown")
		Expect(err).To(HaveOccurred())
	})

})
