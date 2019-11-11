package repo

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/crane"
)

// OciRepo -
type OciRepo struct {
	BaseDir string
}

func (r *OciRepo) cacheDir() string {
	return path.Join(r.BaseDir, "cache")
}

func (r *OciRepo) tarDir() string {
	return path.Join(r.BaseDir, "tar")
}

// Directory -
func (r *OciRepo) Directory(name string) (string, error) {
	dir := path.Join(r.cacheDir(), name)
	if _, err := os.Stat(dir); err != nil {
		return "", fmt.Errorf("directory %s: %s", dir, err.Error())
	}
	return dir, nil
}

// Push -
func (r *OciRepo) Push(name string) error {
	dir, err := r.Directory(name)
	if err != nil {
		return err
	}
	tarFile := path.Join(r.tarDir(), name+"tar.gz")
	err = tarCreate(dir, tarFile)
	if err != nil {
		return err
	}
	_, err = crane.Load(tarFile)
	return err
}

func tarCreate(dir string, dest string) error {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	filepath.Walk(dir, func(root string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		hdr := &tar.Header{
			Name: path.Base(info.Name()),
			Mode: int64(info.Mode()),
			Size: info.Size(),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		body, err := os.Open(path.Join(root, info.Name()))
		if err != nil {
			return err
		}
		defer body.Close()
		if _, err := io.Copy(tw, body); err != nil {
			return err
		}
		return nil
	})
	return nil
}
