package impl

import (
	"path"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.starlark.net/starlark"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
	root       = path.Join(filepath.Dir(b), "..", "..", "..", "..")
)

var _ = Describe("OCIRepo", func() {

	Context("push chart", func() {
		var repo *OciRepo
		var localRepo *LocalRepo
		var thread *starlark.Thread

		BeforeSuite(func() {
			thread = &starlark.Thread{Name: "my thread"}
			localRepo = &LocalRepo{BaseDir: path.Join(root, "example")}
			repo = NewOciRepo(func(repo string) (string, string, error) {
				return "", "", nil
			})

		})
		It("pushes chart correct", func() {
			chart, err := localRepo.Get(thread, "mariadb", nil, nil)
			Expect(err).ToNot(HaveOccurred())
			err = repo.Push(chart, "localhost:5000/mariadb:current")
			Expect(err).ToNot(HaveOccurred())
		})
		It("pulls chart correct", func() {
			chart, err := repo.Get(thread, "localhost:5000/mariadb:current", nil, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(chart.(*chartImpl).Version.String()).To(Equal("6.12.2"))
		})

	})
})
