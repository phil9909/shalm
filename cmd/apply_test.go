package cmd

import (
	"bytes"
	"io"
	"path"
	"path/filepath"
	"runtime"

	"github.com/kramerul/shalm/pkg/shalm"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o ./fake_k8s_test.go ../pkg/shalm K8s

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
	root       = path.Join(filepath.Dir(b), "..")
	example    = path.Join(root, "example", "simple")
)

var _ = Describe("Apply Chart", func() {

	It("produces the correct output", func() {
		writer := bytes.Buffer{}
		k := &FakeK8s{
			ApplyStub: func(i func(io.Writer) error, options *shalm.K8sOptions) error {
				i(&writer)
				return nil
			},
		}
		k.ForNamespaceStub = func(s string) shalm.K8s {
			return k
		}

		err := apply(path.Join(example, "cf"), "mynamespace", shalm.NewK8sValue(k))
		Expect(err).ToNot(HaveOccurred())
		output := writer.String()
		Expect(output).To(ContainSubstring("CREATE OR REPLACE USER 'uaa'"))
		Expect(k.RolloutStatusCallCount()).To(Equal(1))
		Expect(k.ApplyCallCount()).To(Equal(3))
		Expect(k.ForNamespaceCallCount()).To(Equal(3))
		Expect(k.ForNamespaceArgsForCall(0)).To(Equal("mynamespace"))
		Expect(k.ForNamespaceArgsForCall(1)).To(Equal("mynamespace"))
		Expect(k.ForNamespaceArgsForCall(2)).To(Equal("uaa"))
		kind, name, _ := k.RolloutStatusArgsForCall(0)
		Expect(name).To(Equal("mariadb-master"))
		Expect(kind).To(Equal("statefulset"))
	})
})
