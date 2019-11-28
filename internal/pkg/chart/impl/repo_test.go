package impl

import (
	"io/ioutil"
	"net/http"
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
		var rootChart api.Chart

		BeforeEach(func() {
			thread = &starlark.Thread{Name: "my thread"}
			repo = NewRepo(WithAuthCreds(func(repo string) (string, string, error) {
				// return "_json_key", os.Getenv("GCR_ADMIN_CREDENTIALS"), nil
				return "", "", nil
			}))
			rootChart = NewRootChartForDir("default", example)

		})
		It("reads chart from directory", func() {
			chart, err := repo.Get(thread, rootChart, "mariadb", nil, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(chart.GetName()).To(Equal("mariadb"))
		})
		It("reads chart from tar file", func() {
			chart, err := repo.Get(thread, rootChart, "mariadb.tgz", nil, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(chart.GetName()).To(Equal("mariadb"))
		})
		It("reads chart from http", func() {

			http.HandleFunc("/mariadb.tgz", func(w http.ResponseWriter, r *http.Request) {
				content, _ := ioutil.ReadFile(path.Join(example, "mariadb.tgz"))
				w.Write(content)
			})

			go http.ListenAndServe("127.0.0.1:8675", nil)
			chart, err := repo.Get(thread, rootChart, "http://localhost:8675/mariadb.tgz", nil, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(chart.GetName()).To(Equal("mariadb"))
		})

	})
})
