package shalm

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

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
func (r *repoImpl) Get(thread *starlark.Thread, url string, namespace string, args starlark.Tuple, kwargs []starlark.Tuple) (ChartValue, error) {
	md5Sum := md5.Sum([]byte(url))
	cacheDir := path.Join(r.cacheDir, hex.EncodeToString(md5Sum[:]))
	os.RemoveAll(cacheDir)

	if strings.HasPrefix(url, "https:") || strings.HasPrefix(url, "http:") {
		res, err := r.httpClient.Get(url)
		if err != nil {
			return nil, fmt.Errorf("Error fetching %s: %v", url, err)
		}
		if res.StatusCode != 200 {
			return nil, fmt.Errorf("Error fetching %s: status=%d", url, res.StatusCode)
		}
		defer res.Body.Close()
		return newChartFromPackage(thread, r, cacheDir, res.Body, namespace, args, kwargs)
	}
	if stat, err := os.Stat(url); err == nil {
		if stat.IsDir() {
			return newChart(thread, r, url, namespace, args, kwargs)
		}
		return newChartFromFile(thread, r, cacheDir, url, namespace, args, kwargs)
	}
	return nil, fmt.Errorf("Chart not found for url %s", url)
}

func newChartFromFile(thread *starlark.Thread, repo Repo, dir string, tarFile string, namespace string, args starlark.Tuple, kwargs []starlark.Tuple) (ChartValue, error) {
	in, err := os.Open(tarFile)
	if err != nil {
		return nil, err
	}
	defer in.Close()
	return newChartFromPackage(thread, repo, dir, in, namespace, args, kwargs)
}
