package impl

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/kramerul/shalm/pkg/chart/api"
	"go.starlark.net/starlark"
	"gopkg.in/yaml.v2"
)

// SecretData -
type SecretData struct {
	UsernameBase64 string `yaml:"username"`
	PasswordBase64 string `yaml:"password"`
}

// Secret -
type Secret struct {
	APIVersion string     `yaml:"apiVersion"`
	Kind       string     `yaml:"kind"`
	Type       string     `yaml:"type"`
	MetaData   MetaData   `yaml:"metadata"`
	Data       SecretData `yaml:"data,omitempty"`
}

type credential struct {
	username   string
	password   string
	name       string
	hostname   string
	port       string
	uri        string
	additional map[string]string
	applied    bool
}

var (
	_ api.CredentialValue = (*credential)(nil)
)

// String -
func (c *credential) String() string {
	buf := new(strings.Builder)
	buf.WriteString("credential")
	buf.WriteByte('(')
	buf.WriteString("name = ")
	buf.WriteString(c.name)
	buf.WriteString("username = ")
	buf.WriteString(c.username)
	buf.WriteString("password = ")
	buf.WriteString(c.password)
	buf.WriteString("hostname = ")
	buf.WriteString(c.hostname)
	buf.WriteString("port = ")
	buf.WriteString(c.port)
	buf.WriteString("uri = ")
	buf.WriteString(c.uri)

	s := 0
	for i, e := range c.additional {
		if s > 0 {
			buf.WriteString(", ")
		}
		s++
		buf.WriteString(i)
		buf.WriteString(" = ")
		buf.WriteString(e)
	}
	buf.WriteByte(')')
	return buf.String()
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

func (c *credential) GetOrCreate(k8s api.K8s) error {
	var buffer bytes.Buffer
	err := k8s.Get("secret", c.name, &buffer, &api.K8sOptions{Namespaced: true})
	if err != nil {
		if !k8s.IsNotExist(err) {
			return err
		}
		c.username = createRandomString(16)
		c.password = createRandomString(16)
	} else {
		var secret Secret
		dec := yaml.NewDecoder(&buffer)
		err = dec.Decode(&secret)
		if err != nil {
			return err
		}
		content, err := base64.StdEncoding.DecodeString(secret.Data.UsernameBase64)
		if err != nil {
			return err
		}
		c.username = string(content)
		content, err = base64.StdEncoding.DecodeString(secret.Data.PasswordBase64)
		if err != nil {
			return err
		}
		c.password = string(content)

	}
	c.applied = true
	return nil
}

func (c *credential) secret(namespace string) *Secret {
	username := c.username
	password := c.password
	if !c.applied {
		username = "????????"
		password = "????????"
	}
	return &Secret{
		APIVersion: "v1",
		Kind:       "Secret",
		Type:       "Opaque",
		MetaData: MetaData{
			Name:      c.name,
			Namespace: namespace,
		},
		Data: SecretData{
			UsernameBase64: base64.StdEncoding.EncodeToString([]byte(username)),
			PasswordBase64: base64.StdEncoding.EncodeToString([]byte(password)),
		},
	}
}

// Type -
func (c *credential) Type() string { return "credential" }

// Freeze -
func (c *credential) Freeze() {}

// Truth -
func (c *credential) Truth() starlark.Bool { return false }

// Hash -
func (c *credential) Hash() (uint32, error) { panic("implement me") }

// Attr -
func (c *credential) Attr(name string) (starlark.Value, error) {
	switch name {
	case "username":
		if !c.applied {
			return nil, errors.New("password is only available after credential is applied")
		}
		return starlark.String(c.username), nil
	case "password":
		if !c.applied {
			return nil, errors.New("password is only available after credential is applied")
		}
		return starlark.String(c.password), nil
	case "hostname":
		return starlark.String(c.hostname), nil
	case "port":
		return starlark.String(c.port), nil
	case "name":
		return starlark.String(c.name), nil
	case "uri":
		if !c.applied {
			return nil, errors.New("uri is only available after credential is applied")
		}
		return starlark.String(c.uri), nil
	default:
		val, ok := c.additional[name]
		if !ok {
			return starlark.None, starlark.NoSuchAttrError(fmt.Sprintf("credential has no .%s attribute", name))
		}
		return starlark.String(val), nil
	}
}

// AttrNames -
func (c *credential) AttrNames() []string {
	keys := []string{"username", "password", "hostname", "port", "name", "uri"}
	for k := range c.additional {
		keys = append(keys, k)
	}
	return keys
}
