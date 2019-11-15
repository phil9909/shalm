package impl

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/kramerul/shalm/internal/pkg/chart/api"
	"go.starlark.net/starlark"

	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// OciRepo -
type OciRepo struct {
	baseDir  string
	cacheDir string
	resolver remotes.Resolver
}

var _ api.Repo = &OciRepo{}

const (
	customMediaType = "application/tar"
)

// NewRepo -
func NewRepo(authOpts ...docker.AuthorizerOpt) api.Repo {
	homedir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return &OciRepo{
		cacheDir: path.Join(homedir, ".shalm", "cache"),
		resolver: docker.NewResolver(docker.ResolverOptions{
			Hosts: docker.ConfigureDefaultRegistries(
				docker.WithPlainHTTP(func(repository string) (bool, error) {
					if repository == "localhost" || strings.HasPrefix(repository, "localhost:") {
						return true, nil
					}
					return false, nil
				}),
				docker.WithAuthorizer(docker.NewDockerAuthorizer(authOpts...)),
			)}),
	}
}

func (r *OciRepo) testOras(ref string) error {

	ctx := context.Background()
	fileName := "hello.txt"
	fileContent := []byte("Hello World!\n")
	customMediaType := "my.custom.media.type"
	// Push file(s) w custom mediatype to registry
	memoryStore := content.NewMemoryStore()
	desc := memoryStore.Add(fileName, customMediaType, fileContent)
	pushContents := []ocispec.Descriptor{desc}
	fmt.Printf("Pushing %s to %s...\n", fileName, ref)
	desc, err := oras.Push(ctx, r.resolver, ref, memoryStore, pushContents)
	if err != nil {
		return err
	}
	fmt.Printf("Pushed to %s with digest %s\n", ref, desc.Digest)

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
}

// Push -
func (r *OciRepo) Push(chart api.Chart, ref string) error {
	// return r.testOras(ref)
	buffer := bytes.Buffer{}
	if err := tarCreate(chart, &buffer); err != nil {
		return err
	}

	ctx := context.Background()

	memoryStore := content.NewMemoryStore()
	desc := memoryStore.Add("chart.tar", customMediaType, buffer.Bytes())
	pushContents := []ocispec.Descriptor{desc}
	if _, err := oras.Push(ctx, r.resolver, ref, memoryStore, pushContents); err != nil {
		return err
	}

	return nil
}

// Get -
func (r *OciRepo) Get(thread *starlark.Thread, parent api.Chart, ref string, args starlark.Tuple, kwargs []starlark.Tuple) (api.ChartValue, error) {
	var dir string
	if filepath.IsAbs(ref) {
		dir = ref
	} else {
		var cwd string
		if parent == nil {
			var err error
			cwd, err = os.Getwd()
			if err != nil {
				return nil, err
			}
		} else {
			cwd = parent.GetDir()
		}
		dir = path.Join(cwd, ref)
	}

	if _, err := os.Stat(dir); err == nil {
		return NewChart(thread, r, dir, args, kwargs)
	}

	dir = path.Join(r.cacheDir, ref)
	if _, err := os.Stat(dir); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
		{
			// Pull file(s) from registry and save to disk
			fileStore := content.NewFileStore(dir)
			defer fileStore.Close()
			allowedMediaTypes := []string{customMediaType}
			ctx := context.Background()
			_, _, err := oras.Pull(ctx, r.resolver, ref, fileStore, oras.WithAllowedMediaTypes(allowedMediaTypes))
			if err != nil {
				return nil, err
			}
		}
		if err = tarExtract(path.Join(dir, "chart.tar"), dir); err != nil {
			return nil, err
		}
	}
	return NewChart(thread, r, dir, args, kwargs)
}

func tarCreate(chart api.Chart, writer io.Writer) error {
	tw := tar.NewWriter(writer)
	defer tw.Close()
	return chart.Walk(func(file string, size int64, body io.Reader, err error) error {
		hdr := &tar.Header{
			Name: file,
			Mode: 0644,
			Size: size,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := io.Copy(tw, body); err != nil {
			return err
		}
		return nil
	})
}

func tarExtract(tarFile string, dir string) error {
	in, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer in.Close()
	// Open and iterate through the files in the archive.
	tr := tar.NewReader(in)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}
		fn := path.Join(dir, hdr.Name)
		if err := os.MkdirAll(path.Dir(fn), 0755); err != nil {
			return err
		}
		out, err := os.Create(fn)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, tr); err != nil {
			log.Fatal(err)
		}
		out.Close()
	}
	return nil
}
