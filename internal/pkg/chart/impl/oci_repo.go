package impl

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/containerd/containerd/remotes"

	"github.com/containerd/containerd/remotes/docker"
	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// OciRepo -
type OciRepo struct {
	BaseDir  string
	resolver remotes.Resolver
}

const (
	customMediaType = "application/x-tar"
)

// NewOciRepo -
func NewOciRepo(basedir string, auth func(repository string) (username string, password string, err error)) *OciRepo {
	return &OciRepo{
		BaseDir: basedir,
		resolver: docker.NewResolver(docker.ResolverOptions{
			Hosts: docker.ConfigureDefaultRegistries(
				docker.WithPlainHTTP(func(repository string) (bool, error) {
					fmt.Println(repository)
					if repository == "localhost" || strings.HasPrefix(repository, "localhost:") {
						return true, nil
					}
					return false, nil
				}),
				docker.WithAuthorizer(docker.NewDockerAuthorizer(docker.WithAuthCreds(auth))),
			)}),
	}
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
func (r *OciRepo) Push(name string, ref string) error {
	dir, err := r.Directory(name)
	if err != nil {
		return err
	}
	buffer := bytes.Buffer{}
	os.MkdirAll(r.tarDir(), 0755)
	err = tarCreate(dir, &buffer)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Push file(s) w custom mediatype to registry
	memoryStore := content.NewMemoryStore()
	desc := memoryStore.Add(name, customMediaType, buffer.Bytes())
	pushContents := []ocispec.Descriptor{desc}
	desc, err = oras.Push(ctx, r.resolver, ref, memoryStore, pushContents)
	if err != nil {
		return err
	}

	return nil
}

// Pull -
func (r *OciRepo) Pull(name string) error {
	// Pull file(s) from registry and save to disk
	fileStore := content.NewFileStore(r.cacheDir())
	defer fileStore.Close()
	allowedMediaTypes := []string{customMediaType}
	ctx := context.Background()
	_, _, err := oras.Pull(ctx, r.resolver, name, fileStore, oras.WithAllowedMediaTypes(allowedMediaTypes))
	if err != nil {
		return err
	}
	return nil
}

func tarCreate(dir string, writer io.Writer) error {
	tw := tar.NewWriter(writer)
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
