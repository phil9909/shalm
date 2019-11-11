package repo

import (
	"archive/tar"
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
	return path.Join(r.BaseDir)
}

func (r *OciRepo) tarDir() string {
	return path.Join(r.BaseDir, "tgz")
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
	os.MkdirAll(r.tarDir(), 0755)
	tarFile := path.Join(r.tarDir(), name+".tar")
	err = tarCreate(dir, tarFile)
	if err != nil {
		return err
	}
	_, err = crane.Load(tarFile)
	return err
}

func tarCreate(dir string, dest string) error {
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	// gzip := gzip.NewWriter(out)
	// defer gzip.Close()
	tw := tar.NewWriter(out)
	defer tw.Close()
	return filepath.Walk(dir, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, file)
		if err != nil {
			return err
		}
		hdr := &tar.Header{
			Name: rel,
			Mode: int64(info.Mode()),
			Size: info.Size(),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		body, err := os.Open(file)
		if err != nil {
			return err
		}
		defer body.Close()
		if _, err := io.Copy(tw, body); err != nil {
			return err
		}
		return nil
	})
}
