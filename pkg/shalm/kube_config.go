package shalm

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path"
)

func kubeConfigFromEnv() string {
	kubeconfig, ok := os.LookupEnv("KUBECONFIG")
	if ok {
		return kubeconfig
	}
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	kubeconfig = path.Join(home, ".kube", "config")
	return kubeconfig
}

func kubeConfigFromContent(content string) (string, error) {
	c, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		c = []byte(content)
	}
	md5Sum := md5.Sum(c)
	filename := path.Join(os.TempDir(), hex.EncodeToString(md5Sum[:])+".kubeconfig")
	err = ioutil.WriteFile(filename, c, 0644)
	if err != nil {
		return "", err
	}
	return filename, nil
}
