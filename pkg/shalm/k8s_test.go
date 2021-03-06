package shalm

import (
	"bytes"
	"io"

	. "github.com/kramerul/shalm/pkg/shalm/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("k8s", func() {
	k8s := k8sImpl{cmd: "echo"}

	It("apply works", func() {
		err := k8s.Apply(func(writer io.Writer) error { return nil }, &K8sOptions{})
		Expect(err).NotTo(HaveOccurred())
	})
	It("delete works", func() {
		err := k8s.Delete(func(writer io.Writer) error { return nil }, &K8sOptions{})
		Expect(err).NotTo(HaveOccurred())
	})
	It("delete object works", func() {
		err := k8s.DeleteObject("kind", "name", &K8sOptions{})
		Expect(err).NotTo(HaveOccurred())
	})
	It("rollout status works", func() {
		err := k8s.RolloutStatus("kind", "name", &K8sOptions{})
		Expect(err).NotTo(HaveOccurred())
	})
	It("for namespace works", func() {
		k2 := k8s.ForNamespace("ns")
		Expect(k2.(*k8sImpl).namespace).To(Equal("ns"))
	})
	It("get works", func() {
		writer := &bytes.Buffer{}
		err := k8s.Get("kind", "name", writer, &K8sOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(writer.String()).To(Equal("get kind name -o json\n"))
	})
	It("KubeConfigContent works", func() {
		dir := NewTestDir()
		defer dir.Remove()
		dir.MkdirAll("chart2/templates", 0755)
		dir.WriteFile("kubeconfig", []byte("hello"), 0644)
		kubeconfig := dir.Join("kubeconfig")
		k8s := k8sImpl{kubeconfig: &kubeconfig}
		content := k8s.KubeConfigContent()
		Expect(content).NotTo(BeNil())
		Expect(*content).To(Equal("hello"))
	})
	// It("watch works", func() {
	// 	reader, err := k8s.Watch("kind", "name", &K8sOptions{})
	// 	Expect(err).NotTo(HaveOccurred())
	// 	defer reader.Close()
	// 	data, err := ioutil.ReadAll(reader)
	// 	Expect(string(data)).To(Equal(""))
	// })

})
