package shalm

import (
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"runtime"

	shalmv1a1 "github.com/kramerul/shalm/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.starlark.net/starlark"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
	root       = path.Join(filepath.Dir(b), "..", "..")
	example    = path.Join(root, "charts", "example", "simple")
)

var _ = Describe("OCIRepo", func() {

	Context("push chart", func() {
		var repo Repo
		var thread *starlark.Thread

		BeforeEach(func() {
			thread = &starlark.Thread{Name: "main"}
			repo = NewRepo()
		})
		It("reads chart from directory", func() {
			chart, err := repo.Get(thread, path.Join(example, "mariadb"), WithNamespace("namespace"))
			Expect(err).ToNot(HaveOccurred())
			Expect(chart.GetName()).To(Equal("mariadb"))
		})
		It("reads chart from tar file", func() {
			chart, err := repo.Get(thread, path.Join(example, "mariadb-6.12.2.tgz"), WithNamespace("namespace"))
			Expect(err).ToNot(HaveOccurred())
			Expect(chart.GetName()).To(Equal("mariadb"))
		})
		It("reads chart from http", func() {

			http.HandleFunc("/mariadb.tgz", func(w http.ResponseWriter, r *http.Request) {
				content, _ := ioutil.ReadFile(path.Join(example, "mariadb-6.12.2.tgz"))
				w.Write(content)
			})

			go http.ListenAndServe("127.0.0.1:8675", nil)
			chart, err := repo.Get(thread, "http://localhost:8675/mariadb.tgz", WithNamespace("namespace"))
			Expect(err).ToNot(HaveOccurred())
			Expect(chart.GetName()).To(Equal("mariadb"))
		})
		It("creates a proxy", func() {
			chart, err := repo.Get(thread, path.Join(example, "mariadb-6.12.2.tgz"), WithNamespace("namespace"), WithProxy(true))
			Expect(err).ToNot(HaveOccurred())
			Expect(chart.GetName()).To(Equal("mariadb"))
			Expect(chart).To(BeAssignableToTypeOf(&chartProxy{}))
		})
		It("creates chart from spec", func() {
			tgz, err := ioutil.ReadFile(path.Join(example, "mariadb-6.12.2.tgz"))
			Expect(err).ToNot(HaveOccurred())
			chart, err := repo.GetFromSpec(thread, &shalmv1a1.ChartSpec{
				Namespace: "namespace",
				ChartTgz:  tgz,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(chart.GetName()).To(Equal("mariadb"))
		})

	})
})
