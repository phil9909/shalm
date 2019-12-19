package shalm

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/kramerul/shalm/pkg/shalm/renderer"
	"go.starlark.net/starlark"
)

type dataMap map[string][]byte

func (d *dataMap) UnmarshalJSON(b []byte) (err error) {
	var m map[string]string
	if err = json.Unmarshal(b, &m); err != nil {
		return
	}
	result := dataMap{}
	for k, v := range m {
		result[k], err = base64.StdEncoding.DecodeString(v)
		if err != nil {
			return
		}
	}

	*d = result
	return
}

func (d dataMap) MarshalJSON() ([]byte, error) {
	m := make(map[string]string)
	for k, v := range d {
		m[k] = base64.StdEncoding.EncodeToString(v)
	}
	return json.Marshal(m)
}

func (d *dataMap) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	var m map[string]string
	if err = unmarshal(&m); err != nil {
		return
	}
	result := dataMap{}
	for k, v := range m {
		result[k], err = base64.StdEncoding.DecodeString(v)
		if err != nil {
			return
		}
	}
	*d = result
	return
}

func (d dataMap) MarshalYAML() (interface{}, error) {
	m := make(map[string]string)
	for k, v := range d {
		m[k] = base64.StdEncoding.EncodeToString(v)
	}
	return m, nil
}

type secret struct {
	APIVersion string            `json:"apiVersion" yaml:"apiVersion"`
	Kind       string            `json:"kind" yaml:"kind"`
	Type       string            `json:"type" yaml:"type"`
	MetaData   renderer.MetaData `json:"metadata" yaml:"metadata"`
	Data       dataMap           `json:"data,omitempty" yaml:"data,omitempty"`
}

type userCredential struct {
	username    string
	password    string
	usernameKey string
	passwordKey string
	name        string
}

var (
	_ CredentialValue = (*userCredential)(nil)
)

// String -
func (c *userCredential) String() string {
	buf := new(strings.Builder)
	buf.WriteString("user_credential")
	buf.WriteByte('(')
	buf.WriteString("name = ")
	buf.WriteString(c.name)
	buf.WriteString(", username = ")
	buf.WriteString(c.username)
	buf.WriteString(", password = ")
	buf.WriteString(c.password)
	buf.WriteByte(')')
	return buf.String()
}

func (c *userCredential) setDefaultKeys() {
	if c.usernameKey == "" {
		c.usernameKey = "username"
	}
	if c.passwordKey == "" {
		c.passwordKey = "password"
	}
}

func createRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

func (c *userCredential) GetOrCreate(k8s K8s) error {
	c.setDefaultKeys()
	var buffer bytes.Buffer
	err := k8s.Get("secret", c.name, &buffer, &K8sOptions{Namespaced: true})
	if err != nil {
		if !k8s.IsNotExist(err) {
			return err
		}
		if c.username == "" {
			c.username = createRandomString(16)
		}
		if c.password == "" {
			c.password = createRandomString(16)
		}
	} else {
		var secret secret
		dec := json.NewDecoder(&buffer)
		err = dec.Decode(&secret)
		if err != nil {
			return err
		}
		if c.username == "" {
			c.username = string(secret.Data[c.usernameKey])
		}
		if c.password == "" {
			c.password = string(secret.Data[c.passwordKey])
		}

	}
	return nil
}

func (c *userCredential) secret(namespace string) *secret {
	c.setDefaultKeys()
	data := dataMap{}
	if c.username != "" {
		data[c.usernameKey] = []byte(c.username)
	}
	if c.password != "" {
		data[c.passwordKey] = []byte(c.password)
	}
	return &secret{
		APIVersion: "v1",
		Kind:       "Secret",
		Type:       "Opaque",
		MetaData: renderer.MetaData{
			Name:      c.name,
			Namespace: namespace,
		},
		Data: data,
	}
}

// Type -
func (c *userCredential) Type() string { return "user_credential" }

// Freeze -
func (c *userCredential) Freeze() {}

// Truth -
func (c *userCredential) Truth() starlark.Bool { return false }

// Hash -
func (c *userCredential) Hash() (uint32, error) { panic("implement me") }

// Attr -
func (c *userCredential) Attr(name string) (starlark.Value, error) {
	switch name {
	case "name":
		return starlark.String(c.name), nil
	case "username":
		if c.username == "" {
			return nil, errors.New("username is only available after user_credential is applied")
		}
		return starlark.String(c.username), nil
	case "password":
		if c.password == "" {
			return nil, errors.New("password is only available after user_credential is applied")
		}
		return starlark.String(c.password), nil
	default:
		return starlark.None, starlark.NoSuchAttrError(fmt.Sprintf("user_credential has no .%s attribute", name))
	}
}

// AttrNames -
func (c *userCredential) AttrNames() []string {
	return []string{"name", "username", "password"}
}
