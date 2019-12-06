package shalm

import (
	"io/ioutil"
	"path"
	"path/filepath"
)

func (f files) Glob(pattern string) map[string][]byte {
	result := make(map[string][]byte)
	matches, err := filepath.Glob(path.Join(f.dir, pattern))
	if err != nil {
		return result
	}
	for _, match := range matches {
		data, err := ioutil.ReadFile(match)
		if err == nil {
			p, err := filepath.Rel(f.dir, match)
			if err == nil {
				result[p] = data
			} else {
			}
		} else {
		}
	}
	return result
}

func (f files) Get(name string) string {
	data, err := ioutil.ReadFile(path.Join(f.dir, name))
	if err != nil {
		return err.Error()
	}
	return string(data)
}
