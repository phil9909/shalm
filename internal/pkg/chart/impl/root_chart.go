package impl

import (
	"io"
	"os"

	"github.com/kramerul/shalm/internal/pkg/chart/api"
	"github.com/spf13/afero"
	"go.starlark.net/starlark"
)

type rootChart struct {
	dir       string
	namespace string
	fs        afero.Fs
}

var (
	_ api.Chart = (*rootChart)(nil)
)

// NewRootChart -
func NewRootChart(namespace string) (api.Chart, error) {
	var err error
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return NewRootChartForDir(namespace, cwd, afero.NewOsFs()), nil
}

// NewRootChartForDir -
func NewRootChartForDir(namespace string, dir string, fs afero.Fs) api.Chart {
	return &rootChart{dir: dir, namespace: namespace, fs: fs}
}

func (c *rootChart) GetFs() afero.Fs {
	return c.fs
}

func (c *rootChart) GetName() string {
	return "root"
}

func (c *rootChart) GetNamespace() string {
	return c.namespace
}

func (c *rootChart) GetDir() string {
	return c.dir
}

func (c *rootChart) Walk(cb func(name string, size int64, body io.Reader, err error) error) error {
	return nil
}

func (c *rootChart) Apply(thread *starlark.Thread, k api.K8s) error {
	return nil
}
func (c *rootChart) Delete(thread *starlark.Thread, k api.K8s) error {
	return nil
}
func (c *rootChart) Template(thread *starlark.Thread) (string, error) {
	return "", nil
}
