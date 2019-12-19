package renderer

import (
	"bytes"
	"io"
	"os"

	. "github.com/kramerul/shalm/pkg/shalm/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DirRender", func() {

	fileRenderer := func(filename string, writer io.Writer) error {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(writer, f)
		return err
	}
	It("doesn't set default namespace non namepspaced objects", func() {
		for _, kind := range []string{"Namespace", "ResourceQuota", "CustomResourceDefinition", "ClusterRole",
			"ClusterRoleList", "ClusterRoleBinding", "ClusterRoleBindingList", "APIService"} {
			obj := object{Kind: kind}
			obj.setDefaultNamespace("test")
			Expect(obj.MetaData.Namespace).To(Equal(""))
		}
	})

	It("Sorts in correct order", func() {
		ordinal := 0
		for _, kind := range []string{"Namespace",
			"NetworkPolicy",
			"ResourceQuota",
			"LimitRange",
			"PodSecurityPolicy",
			"PodDisruptionBudget",
			"Secret",
			"ConfigMap",
			"StorageClass",
			"PersistentVolume",
			"PersistentVolumeClaim",
			"ServiceAccount",
			"CustomResourceDefinition",
			"ClusterRole",
			"ClusterRoleList",
			"ClusterRoleBinding",
			"ClusterRoleBindingList",
			"Role",
			"RoleList",
			"RoleBinding",
			"RoleBindingList",
			"Service",
			"DaemonSet",
			"Pod",
			"ReplicationController",
			"ReplicaSet",
			"Deployment",
			"HorizontalPodAutoscaler",
			"StatefulSet",
			"Job",
			"CronJob",
			"Ingress",
			"APIService"} {
			obj := object{Kind: kind}
			ord := obj.kindOrdinal()
			Expect(ord).To(BeNumerically(">", ordinal))
			ordinal = ord
		}
	})

	Context("renders chart", func() {
		It("renders multipe files", func() {
			var err error
			dir := NewTestDir()
			defer dir.Remove()
			dir.WriteFile("test1.yaml", []byte("test: test1"), 0644)
			dir.WriteFile("test2.yml", []byte("test: test2"), 0644)

			writer := &bytes.Buffer{}
			err = DirRender("namespace", writer, &Options{}, DirSpec{dir.Root(), fileRenderer})
			Expect(err).ToNot(HaveOccurred())
			Expect(writer.String()).To(Equal("---\nmetadata:\n  namespace: namespace\ntest: test1\n---\nmetadata:\n  namespace: namespace\ntest: test2\n"))
		})
		It("repects glob patterns", func() {
			var err error
			dir := NewTestDir()
			defer dir.Remove()
			dir.WriteFile("test1.yaml", []byte("test: test1"), 0644)
			dir.WriteFile("test3.yaml", []byte("test: test2"), 0644)
			writer := &bytes.Buffer{}
			err = DirRender("namespace", writer, &Options{Glob: "*[1-2].yaml"}, DirSpec{dir.Root(), fileRenderer})
			Expect(err).ToNot(HaveOccurred())
			Expect(writer.String()).To(Equal("---\nmetadata:\n  namespace: namespace\ntest: test1\n"))
		})
		It("sorts by kind", func() {
			var err error
			dir := NewTestDir()
			defer dir.Remove()
			dir.WriteFile("test1.yaml", []byte("kind: Other"), 0644)
			dir.WriteFile("test2.yaml", []byte("kind: StatefulSet"), 0644)
			dir.WriteFile("test3.yaml", []byte("kind: Service"), 0644)

			By("Sorts in install order")
			writer := &bytes.Buffer{}
			err = DirRender("namespace", writer, &Options{}, DirSpec{dir.Root(), fileRenderer})
			Expect(err).ToNot(HaveOccurred())
			Expect(writer.String()).To(Equal(`---
metadata:
  namespace: namespace
kind: Service
---
metadata:
  namespace: namespace
kind: StatefulSet
---
metadata:
  namespace: namespace
kind: Other
`))

			By("Sorts in uninstall order")
			writer = &bytes.Buffer{}
			err = DirRender("namespace", writer, &Options{UninstallOrder: true}, DirSpec{dir.Root(), fileRenderer})
			Expect(err).ToNot(HaveOccurred())
			Expect(writer.String()).To(Equal(`---
metadata:
  namespace: namespace
kind: Other
---
metadata:
  namespace: namespace
kind: StatefulSet
---
metadata:
  namespace: namespace
kind: Service
`))
		})
	})
})
