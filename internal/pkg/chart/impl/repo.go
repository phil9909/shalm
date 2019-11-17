package impl

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/kramerul/shalm/internal/pkg/chart/api"
	"go.starlark.net/starlark"

	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// RepoOpts -
type RepoOpts func(*OciRepo)

// WithAuthCreds -
func WithAuthCreds(credentials func(string) (string, string, error)) RepoOpts {
	return func(opt *OciRepo) {
		opt.credentials = credentials
	}
}

// OciRepo -
type OciRepo struct {
	baseDir     string
	cacheDir    string
	resolver    remotes.Resolver
	httpClient  *http.Client
	credentials func(string) (string, string, error)
}

var _ api.Repo = &OciRepo{}

const (
	customMediaType = "application/tar"
)

// NewRepo -
func NewRepo(authOpts ...RepoOpts) api.Repo {
	homedir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	r := &OciRepo{
		cacheDir: path.Join(homedir, ".shalm", "cache"),
		httpClient: &http.Client{
			Timeout: time.Second * 60,
		},
	}
	r.resolver = docker.NewResolver(docker.ResolverOptions{
		Hosts: docker.ConfigureDefaultRegistries(
			docker.WithPlainHTTP(func(repository string) (bool, error) {
				if repository == "localhost" || strings.HasPrefix(repository, "localhost:") {
					return true, nil
				}
				return false, nil
			}),
			docker.WithAuthorizer(docker.NewDockerAuthorizer(docker.WithAuthCreds(func(ref string) (string, string, error) {
				if r.credentials != nil {
					return r.credentials(ref)
				} else {
					return "", "", nil
				}
			}))),
		)})
	for _, a := range authOpts {
		a(r)
	}

	return r
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
		dir = path.Join(parent.GetDir(), ref)
	}
	md5Sum := md5.Sum([]byte(ref))
	cacheDir := path.Join(r.cacheDir, hex.EncodeToString(md5Sum[:]))
	os.RemoveAll(cacheDir)

	if stat, err := os.Stat(dir); err == nil {
		if stat.IsDir() {
			return NewChart(thread, r, dir, parent, args, kwargs)
		}
		if err = tarExtractFromFile(dir, cacheDir); err != nil {
			return nil, err
		}
		return NewChart(thread, r, cacheDir, parent, args, kwargs)

	}

	if strings.HasPrefix(ref, "https:") || strings.HasPrefix(ref, "http:") {
		res, err := r.httpClient.Get(ref)
		if err != nil {
			return nil, fmt.Errorf("Error fetching %s: %v", ref, err)
		}
		if res.StatusCode != 200 {
			return nil, fmt.Errorf("Error fetching %s: status=%d", ref, res.StatusCode)
		}
		defer res.Body.Close()
		if err = tarExtract(res.Body, cacheDir); err != nil {
			return nil, err
		}
	} else {
		// Pull file(s) from registry and save to disk
		fileStore := content.NewFileStore(cacheDir)
		defer fileStore.Close()
		allowedMediaTypes := []string{customMediaType}
		ctx := context.Background()
		_, _, err := oras.Pull(ctx, r.resolver, ref, fileStore, oras.WithAllowedMediaTypes(allowedMediaTypes))
		if err != nil {
			return nil, err
		}
		if err = tarExtractFromFile(path.Join(cacheDir, "chart.tar"), cacheDir); err != nil {
			return nil, err
		}
	}
	return NewChart(thread, r, cacheDir, parent, args, kwargs)
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

func tarExtractFromFile(tarFile string, dir string) error {
	in, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer in.Close()
	return tarExtract(in, dir)
}

func tarExtract(in io.Reader, dir string) error {
	reader := bufio.NewReader(in)
	testBytes, err := reader.Peek(64)
	if err != nil {
		return err
	}
	in = reader
	contentType := http.DetectContentType(testBytes)
	if strings.Contains(contentType, "x-gzip") {
		in, err = gzip.NewReader(in)
		if err != nil {
			return err
		}
	}
	tr := tar.NewReader(in)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}
		if hdr.FileInfo().IsDir() {
			continue
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
