package impl

import (
	"fmt"
	"os"
	"path"

	"github.com/kramerul/shalm/internal/pkg/chart/api"
	"go.starlark.net/starlark"
)

// LocalRepo -
type LocalRepo struct {
	BaseDir string
}

// Get -
func (r *LocalRepo) Get(thread *starlark.Thread, name string, args starlark.Tuple, kwargs []starlark.Tuple) (api.ChartValue, error) {
	dir := path.Join(r.BaseDir, name)
	if _, err := os.Stat(dir); err != nil {
		return nil, fmt.Errorf("directory %s: %s", dir, err.Error())
	}
	return NewChart(thread, r, dir, name, args, kwargs)
}
