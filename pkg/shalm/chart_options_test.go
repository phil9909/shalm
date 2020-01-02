package shalm

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.starlark.net/starlark"
)

var _ = Describe("Chart Options", func() {

	It("produces the correct output", func() {
		args := ChartOptions{cmdArgs: []string{"a=b,c=d"}}
		kwargs := args.kwArgs()
		Expect(kwargs).To(HaveLen(2))
		Expect(kwargs[0]).To(HaveLen(2))
		Expect(kwargs[0][0]).To(Equal(starlark.String("a")))
		Expect(kwargs[0][1]).To(Equal(starlark.String("b")))
		Expect(kwargs[1]).To(HaveLen(2))
		Expect(kwargs[1][0]).To(Equal(starlark.String("c")))
		Expect(kwargs[1][1]).To(Equal(starlark.String("d")))
	})
})
