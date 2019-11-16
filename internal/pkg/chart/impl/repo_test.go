package impl

import (
	"path"
	"path/filepath"
	"runtime"

	"github.com/kramerul/shalm/internal/pkg/chart/api"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.starlark.net/starlark"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
	root       = path.Join(filepath.Dir(b), "..", "..", "..", "..")
	example    = path.Join(root, "example")
)

var _ = Describe("OCIRepo", func() {

	Context("push chart", func() {
		var repo api.Repo
		var thread *starlark.Thread

		BeforeEach(func() {
			thread = &starlark.Thread{Name: "my thread"}
			repo = NewRepo()

		})
		It("pushes chart correct", func() {
			rc, err := NewRootChart(example)
			Expect(err).ToNot(HaveOccurred())
			_, err = repo.Get(thread, rc, "mariadb", nil, nil)
			Expect(err).ToNot(HaveOccurred())
			// err = repo.Push(chart, "localhost:5000/mariadb:current")
			// Expect(err).ToNot(HaveOccurred())
		})
		// It("pulls chart correct", func() {
		// 	chart, err := repo.Get(thread, nil, "localhost:5000/mariadb:current", nil, nil)
		// 	Expect(err).ToNot(HaveOccurred())
		// 	Expect(chart.(*chartImpl).Version.String()).To(Equal("6.12.2"))
		// })

	})
})
