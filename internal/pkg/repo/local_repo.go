package repo

import (
	"fmt"
	"os"
	"path"
)

// LocalRepo -
type LocalRepo struct {
	BaseDir string
}

// Directory -
func (r *LocalRepo) Directory(name string) (string, error) {
	dir := path.Join(r.BaseDir, name)
	if _, err := os.Stat(dir); err != nil {
		return "", fmt.Errorf("directory %s: %s", dir, err.Error())
	}
	return dir, nil
}
