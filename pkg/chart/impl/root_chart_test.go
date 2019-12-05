package impl

import (
	"github.com/blang/semver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("root chart", func() {

	It("behaves simple", func() {
		rc := NewRootChartForDir("namespace", "/tmp")
		Expect(rc.GetDir()).To(Equal("/tmp"))
		Expect(rc.GetNamespace()).To(Equal("namespace"))
		Expect(rc.GetName()).To(Equal("root"))
		Expect(rc.GetVersion()).To(Equal(semver.Version{}))
		Expect(rc.Apply(nil, nil)).NotTo(HaveOccurred())
		Expect(rc.Delete(nil, nil)).NotTo(HaveOccurred())
		Expect(rc.Package(nil)).NotTo(HaveOccurred())
		_, err := rc.Template(nil)
		Expect(err).NotTo(HaveOccurred())
		_, err = NewRootChart("namespace")
		Expect(err).NotTo(HaveOccurred())
	})

})
