package shalm

import (
	"bytes"
	"io"
	"io/ioutil"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.starlark.net/starlark"
)

var _ = Describe("K8sValue", func() {

	It("behaves like starlark value", func() {
		k8s := &k8sValueImpl{&FakeK8s{
			InspectStub: func() string {
				return "kubeconfig = "
			},
		}}
		Expect(k8s.String()).To(ContainSubstring("kubeconfig = "))
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

	It("methods behave well", func() {
		fake := &FakeK8s{
			GetStub: func(kind string, name string, writer io.Writer, k8s *K8sOptions) error {
				writer.Write([]byte("{}"))
				return nil
			},
		}
		k8s := &k8sValueImpl{fake}
		thread := &starlark.Thread{}
		for _, method := range []string{"rollout_status", "delete", "get"} {
			value, err := k8s.Attr(method)
			_, err = starlark.Call(thread, value, starlark.Tuple{starlark.String("kind"), starlark.String("object")},
				[]starlark.Tuple{{starlark.String("timeout"), starlark.MakeInt(10)},
					{starlark.String("namespaced"), starlark.Bool(true)}})
			Expect(err).NotTo(HaveOccurred())
		}
		{
			value, err := k8s.Attr("wait")
			_, err = starlark.Call(thread, value, starlark.Tuple{starlark.String("kind"), starlark.String("object"), starlark.String("condition")},
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
		Expect(fake.WaitCallCount()).To(Equal(1))
		Expect(fake.DeleteObjectCallCount()).To(Equal(1))
		Expect(fake.GetCallCount()).To(Equal(1))
	})

	It("watches objects", func() {
		fake := &FakeK8s{
			WatchStub: func(kind string, name string, options *K8sOptions) (closer io.ReadCloser, e error) {
				return ioutil.NopCloser(bytes.NewReader([]byte(`{ "key" : "value" }`))), nil
			},
		}
		k8s := &k8sValueImpl{fake}
		thread := &starlark.Thread{}
		watch, err := k8s.Attr("watch")
		value, err := starlark.Call(thread, watch, starlark.Tuple{starlark.String("kind"), starlark.String("object")},
			[]starlark.Tuple{{starlark.String("timeout"), starlark.MakeInt(10)},
				{starlark.String("namespaced"), starlark.Bool(true)}})

		Expect(err).NotTo(HaveOccurred())
		iterable := value.(starlark.Iterable)
		iterator := iterable.Iterate()
		var obj starlark.Value
		found := iterator.Next(&obj)
		Expect(found).To(BeTrue())
		Expect(fake.WatchCallCount()).To(Equal(1))
		dict := unwrapDict(obj).(*starlark.Dict)
		val, found, err := dict.Get(starlark.String("key"))
		Expect(val).To(Equal(starlark.String("value")))
	})

})
