package repo

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

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

func NewOciRepo(basedir string) *OciRepo {
	return &OciRepo{
		BaseDir: basedir,
		resolver: docker.NewResolver(docker.ResolverOptions{
			Hosts: docker.ConfigureDefaultRegistries(
				docker.WithAuthorizer(
					docker.NewDockerAuthorizer(docker.WithAuthCreds(func(repository string) (s string, s2 string, e error) {
						return "_json_key", os.Getenv("GCR_ADMIN_CREDENTIALS"), nil
					})))),
		}),
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
func (r *OciRepo) Push(name string) error {
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
	ref := "gcr.io/peripli/oras:test"

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
	// Pull file(s) from registry and save to disk
	fmt.Printf("Pulling from %s and saving to %s...\n", ref, fileName)
	fileStore := content.NewFileStore("")
	defer fileStore.Close()
	allowedMediaTypes := []string{customMediaType}
	desc, _, err = oras.Pull(ctx, r.resolver, ref, fileStore, oras.WithAllowedMediaTypes(allowedMediaTypes))
	if err != nil {
		return err
	}
	fmt.Printf("Pulled from %s with digest %s\n", ref, desc.Digest)
	fmt.Printf("Try running 'cat %s'\n", fileName)
	return nil
	return err
}

// Pull -
func (r *OciRepo) Pull(name string) error {
	retunr nil
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
