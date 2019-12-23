package shalm

import (
	"bytes"
	"io"

	"go.starlark.net/starlark"

	. "github.com/kramerul/shalm/pkg/shalm/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Chart Proxy", func() {

	Context("proxies apply and delete calls", func() {
		thread := &starlark.Thread{Name: "test"}
		repo := NewRepo()
		var dir TestDir
		var chart ChartValue
		BeforeEach(func() {
			dir = NewTestDir()
			dir.WriteFile("Chart.yaml", []byte("name: mariadb\nversion: 6.12.2\n"), 0644)
			dir.WriteFile("values.yaml", []byte("replicas: \"1\"\ntimeout: \"30s\"\n"), 0644)
			args := starlark.Tuple{starlark.String("hello")}
			kwargs := []starlark.Tuple{starlark.Tuple{starlark.String("key"), starlark.String("value")}}
			impl, err := newChart(thread, repo, dir.Root(), "namespace", args, kwargs)
			Expect(err).NotTo(HaveOccurred())
			chart, err = newChartProxy(impl, "http://test.com", args, kwargs)
			Expect(err).NotTo(HaveOccurred())

		})
		AfterEach(func() {
			dir.Remove()

		})

		It("applies a ShalmChart to k8s", func() {
			buffer := &bytes.Buffer{}
			k := &FakeK8s{
				ApplyStub: func(cb func(io.Writer) error, options *K8sOptions) error {
					return cb(buffer)
				},
			}
			k.ForNamespaceStub = func(s string) K8s {
				return k
			}
			err := chart.Apply(thread, k)
			Expect(err).NotTo(HaveOccurred())
			Expect(k.ApplyCallCount()).To(Equal(1))
			Expect(buffer.String()).To(ContainSubstring(`{"kind":"Namespace","apiVersion":"v1","metadata":{"name":"namespace","creationTimestamp":null},"spec":{},"status":{}}`))
			Expect(buffer.String()).To(ContainSubstring(`"spec":{"values":{"replicas":"1","timeout":"30s"},"args":["hello"],"kwargs":{"key":"value"},"namespace":"namespace","url":"http://test.com"}`))
			Expect(buffer.String()).To(ContainSubstring(`"name":"mariadb","namespace":"namespace"`))
		})
		It("deletes a ShalmChart from k8s", func() {
			k := &FakeK8s{}
			k.ForNamespaceStub = func(s string) K8s {
				return k
			}
			err := chart.Delete(thread, k)
			Expect(err).NotTo(HaveOccurred())
			Expect(k.DeleteObjectCallCount()).To(Equal(1))
			kind, name, _ := k.DeleteObjectArgsForCall(0)
			Expect(kind).To(Equal("ShalmChart"))
			Expect(name).To(Equal("mariadb"))
		})
	})

})
