package impl

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/kramerul/shalm/pkg/chart"
	"go.starlark.net/starlark"
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
	for _, a := range authOpts {
		a(r)
	}

	return r
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
	return nil, fmt.Errorf("Chart not found for url %s", ref)
}

func newChartFromFile(thread *starlark.Thread, repo chart.Repo, dir string, tarFile string, parent chart.Chart, args starlark.Tuple, kwargs []starlark.Tuple) (chart.ChartValue, error) {
	in, err := os.Open(tarFile)
	if err != nil {
		return nil, err
	}
	defer in.Close()
	return NewChartFromPackage(thread, repo, dir, in, parent, args, kwargs)
}
