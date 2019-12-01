package impl

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/kramerul/shalm/pkg/chart"
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

var _ chart.Repo = &OciRepo{}

const (
	customMediaType = "application/tar"
)

// NewRepo -
func NewRepo(authOpts ...RepoOpts) chart.Repo {
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
				}
				return "", "", nil
			}))),
		)})
	for _, a := range authOpts {
		a(r)
	}

	return r
}

// Push -
func (r *OciRepo) Push(chart chart.Chart, ref string) error {
	// return r.testOras(ref)
	buffer := bytes.Buffer{}
	if err := chart.Package(&buffer); err != nil {
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
func (r *OciRepo) Get(thread *starlark.Thread, parent chart.Chart, ref string, args starlark.Tuple, kwargs []starlark.Tuple) (chart.ChartValue, error) {
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
		return newChartFromFile(thread, r, cacheDir, dir, parent, args, kwargs)
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
		return NewChartFromPackage(thread, r, cacheDir, res.Body, parent, args, kwargs)
	}
	// Pull file(s) from registry and save to disk
	fileStore := content.NewFileStore(cacheDir)
	defer fileStore.Close()
	allowedMediaTypes := []string{customMediaType}
	ctx := context.Background()
	_, _, err := oras.Pull(ctx, r.resolver, ref, fileStore, oras.WithAllowedMediaTypes(allowedMediaTypes))
	if err != nil {
		return nil, err
	}
	return newChartFromFile(thread, r, cacheDir, path.Join(cacheDir, "chart.tar"), parent, args, kwargs)
}

func newChartFromFile(thread *starlark.Thread, repo chart.Repo, dir string, tarFile string, parent chart.Chart, args starlark.Tuple, kwargs []starlark.Tuple) (chart.ChartValue, error) {
	in, err := os.Open(tarFile)
	if err != nil {
		return nil, err
	}
	defer in.Close()
	return NewChartFromPackage(thread, repo, dir, in, parent, args, kwargs)
}
