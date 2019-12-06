package shalm

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.starlark.net/starlark"
)

var _ = Describe("K8sValue", func() {

	It("behaves like starlark value", func() {
		k8s := &k8sValueImpl{&FakeK8s{}}
		Expect(k8s.String()).To(ContainSubstring("KUBECONFIG = "))
		Expect(k8s.Type()).To(Equal("k8s"))
		Expect(func() { k8s.Hash() }).Should(Panic())
		Expect(k8s.Truth()).To(BeEquivalentTo(false))
		for _, method := range []string{"rollout_status", "delete", "get"} {
			value, err := k8s.Attr(method)
			Expect(err).NotTo(HaveOccurred())
			_, ok := value.(starlark.Callable)
			Expect(ok).To(BeTrue())
		}
		Expect(k8s.AttrNames()).To(ConsistOf("rollout_status", "delete", "get"))
	})

	It("methods", func() {
		fake := &FakeK8s{}
		k8s := &k8sValueImpl{fake}
		thread := &starlark.Thread{}
		for _, method := range []string{"rollout_status", "delete", "get"} {
			value, err := k8s.Attr(method)
			_, err = starlark.Call(thread, value, starlark.Tuple{starlark.String("kind"), starlark.String("object")},
				[]starlark.Tuple{{starlark.String("timeout"), starlark.MakeInt(10)},
					{starlark.String("namespaced"), starlark.Bool(true)}})
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(fake.RolloutStatusCallCount()).To(Equal(1))
		kind, name, options := fake.RolloutStatusArgsForCall(0)
		Expect(kind).To(Equal("kind"))
		Expect(name).To(Equal("object"))
		Expect(options.Timeout).To(Equal(10 * time.Second))
		Expect(options.Namespaced).To(BeTrue())
		Expect(fake.DeleteObjectCallCount()).To(Equal(1))
		Expect(fake.GetCallCount()).To(Equal(1))
	})

})
