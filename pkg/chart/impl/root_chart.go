package impl

import (
	"io"
	"os"

	"github.com/blang/semver"
	"github.com/kramerul/shalm/pkg/chart"
	"go.starlark.net/starlark"
)

type rootChart struct {
	dir       string
	namespace string
}

var (
	_ chart.Chart = (*rootChart)(nil)
)

// NewRootChart -
func NewRootChart(namespace string) (chart.Chart, error) {
	var err error
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return NewRootChartForDir(namespace, cwd), nil
}

// NewRootChartForDir -
func NewRootChartForDir(namespace string, dir string) chart.Chart {
	return &rootChart{dir: dir, namespace: namespace}
}

func (c *rootChart) GetName() string {
	return "root"
}

func (c *rootChart) GetNamespace() string {
	return c.namespace
}

func (c *rootChart) GetVersion() semver.Version {
	return semver.Version{}
}

func (c *rootChart) GetDir() string {
	return c.dir
}

func (c *rootChart) Package(writer io.Writer) error {
	return nil
}

func (c *rootChart) Apply(thread *starlark.Thread, k chart.K8s) error {
	return nil
}
func (c *rootChart) Delete(thread *starlark.Thread, k chart.K8s) error {
	return nil
}
func (c *rootChart) Template(thread *starlark.Thread) (string, error) {
	return "", nil
}
