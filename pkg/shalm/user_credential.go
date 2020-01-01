package shalm

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go.starlark.net/starlark"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
)

var (
	serializer = json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
)

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

func makeUserCredential(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
	s := &userCredential{}
	s.setDefaultKeys()
	if err := starlark.UnpackArgs("user_credential", args, kwargs, "name", &s.name,
		"username_key?", &s.usernameKey, "password_key?", &s.passwordKey,
		"username?", &s.username, "password?", &s.password); err != nil {
		return starlark.None, err
	}
	return s, nil
}

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
		var secret corev1.Secret

		_, _, err = serializer.Decode(buffer.Bytes(), nil, &secret)
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

func (c *userCredential) secret(namespace string) *corev1.Secret {
	c.setDefaultKeys()
	data := map[string][]byte{}
	if c.username != "" {
		data[c.usernameKey] = []byte(c.username)
	}
	if c.password != "" {
		data[c.passwordKey] = []byte(c.password)
	}
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.name,
			Namespace: namespace,
		},
		Type: "Opaque",
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
