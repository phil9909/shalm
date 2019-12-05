package impl

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/kramerul/shalm/pkg/chart"
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
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Type       string            `yaml:"type"`
	MetaData   MetaData          `yaml:"metadata"`
	Data       map[string]string `yaml:"data,omitempty"`
}

type userCredential struct {
	username    string
	password    string
	usernameKey string
	passwordKey string
	name        string
}

var (
	_ chart.CredentialValue = (*userCredential)(nil)
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

func (c *userCredential) GetOrCreate(k8s chart.K8s) error {
	c.setDefaultKeys()
	var buffer bytes.Buffer
	err := k8s.Get("user_credential", c.name, &buffer, &chart.K8sOptions{Namespaced: true})
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
		var secret Secret
		dec := yaml.NewDecoder(&buffer)
		err = dec.Decode(&secret)
		if err != nil {
			return err
		}
		if c.username == "" {
			content, err := base64.StdEncoding.DecodeString(secret.Data[c.usernameKey])
			if err != nil {
				return err
			}
			c.username = string(content)
		}
		if c.password == "" {
			content, err := base64.StdEncoding.DecodeString(secret.Data[c.passwordKey])
			if err != nil {
				return err
			}
			c.password = string(content)
		}

	}
	return nil
}

func (c *userCredential) secret(namespace string) *Secret {
	c.setDefaultKeys()
	return &Secret{
		APIVersion: "v1",
		Kind:       "Secret",
		Type:       "Opaque",
		MetaData: MetaData{
			Name:      c.name,
			Namespace: namespace,
		},
		Data: map[string]string{
			c.usernameKey: base64.StdEncoding.EncodeToString([]byte(c.username)),
			c.passwordKey: base64.StdEncoding.EncodeToString([]byte(c.password)),
		},
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
