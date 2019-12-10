package shalm

import (
	"encoding/base64"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("kube config", func() {

	It("read it from env", func() {
		kubeconfig := kubeConfigFromEnv()
		Expect(kubeconfig).NotTo(BeEmpty())
	})

	It("read it from env", func() {
		os.Unsetenv("KUBECONFIG")
		kubeconfig := kubeConfigFromEnv()
		Expect(kubeconfig).NotTo(BeEmpty())
	})

	It("read it from base64 encoded value", func() {
		os.Unsetenv("KUBECONFIG")
		kubeconfig, err := kubeConfigFromContent(base64.StdEncoding.EncodeToString([]byte("Hello world")))
		Expect(err).NotTo(HaveOccurred())
		Expect(kubeconfig).NotTo(BeEmpty())
		content, err := ioutil.ReadFile(kubeconfig)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal("Hello world"))
	})

	It("read it from plain value", func() {
		os.Unsetenv("KUBECONFIG")
		kubeconfig, err := kubeConfigFromContent("Hello world")
		Expect(err).NotTo(HaveOccurred())
		Expect(kubeconfig).NotTo(BeEmpty())
		content, err := ioutil.ReadFile(kubeconfig)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal("Hello world"))
	})

})