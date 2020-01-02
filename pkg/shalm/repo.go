package shalm

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	shalmv1a1 "github.com/kramerul/shalm/api/v1alpha1"
	"go.starlark.net/starlark"
)

type repoImpl struct {
	cacheDir   string
	httpClient *http.Client
}

var _ Repo = &repoImpl{}

const (
	customMediaType = "application/tar"
)

// NewRepo -
func NewRepo() Repo {
	homedir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	r := &repoImpl{
		cacheDir: path.Join(homedir, ".shalm", "cache"),
		httpClient: &http.Client{
			Timeout: time.Second * 60,
		},
	}
	return r
}

// Get -
func (r *repoImpl) Get(thread *starlark.Thread, url string, opts ...ChartOption) (ChartValue, error) {

	co := chartOptions(opts)

	proxyFunc := func(chart *chartImpl, err error) (ChartValue, error) {
		return chart, err
	}
	if co.proxy {
		proxyFunc = func(chart *chartImpl, err error) (ChartValue, error) {
			if err != nil {
				return nil, err
			}
			return newChartProxy(chart, url, co.args, co.kwargs)
		}
	}

	if strings.HasPrefix(url, "https:") || strings.HasPrefix(url, "http:") {
		res, err := r.httpClient.Get(url)
		if err != nil {
			return nil, fmt.Errorf("Error fetching %s: %v", url, err)
		}
		if res.StatusCode != 200 {
			return nil, fmt.Errorf("Error fetching %s: status=%d", url, res.StatusCode)
		}
		defer res.Body.Close()
		return proxyFunc(newChartFromReader(thread, r, r.cacheDirForChart([]byte(url)), res.Body, opts...))
	}
	if stat, err := os.Stat(url); err == nil {
		if stat.IsDir() {
			return proxyFunc(newChart(thread, r, url, opts...))
		}
		return proxyFunc(newChartFromFile(thread, r, r.cacheDirForChart([]byte(url)), url, opts...))
	}
	return nil, fmt.Errorf("Chart not found for url %s", url)
}

func (r *repoImpl) cacheDirForChart(data []byte) string {
	md5Sum := md5.Sum(data)
	cacheDir := path.Join(r.cacheDir, hex.EncodeToString(md5Sum[:]))
	os.RemoveAll(cacheDir)
	return cacheDir
}

func (r *repoImpl) GetFromSpec(thread *starlark.Thread, spec *shalmv1a1.ChartSpec) (ChartValue, error) {
	c, err := newChartFromReader(thread, r, r.cacheDirForChart(spec.ChartTgz), bytes.NewReader(spec.ChartTgz),
		WithNamespace(spec.Namespace), WithSuffix(spec.Suffix), WithArgs(toStarlark(spec.Args).(starlark.Tuple)),
		WithKwArgs(kwargsToStarlark(spec.KwArgs)))
	if err != nil {
		return nil, err
	}
	c.mergeValues(spec.Values)
	return c, nil
}

func newChartFromReader(thread *starlark.Thread, repo Repo, dir string, reader io.Reader, opts ...ChartOption) (*chartImpl, error) {
	if err := tarExtract(reader, dir); err != nil {
		return nil, err
	}
	return newChart(thread, repo, dir, opts...)
}

func newChartFromFile(thread *starlark.Thread, repo Repo, dir string, tarFile string, opts ...ChartOption) (*chartImpl, error) {
	in, err := os.Open(tarFile)
	if err != nil {
		return nil, err
	}
	defer in.Close()
	return newChartFromReader(thread, repo, dir, in, opts...)
}
