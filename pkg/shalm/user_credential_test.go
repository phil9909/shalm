package shalm

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.starlark.net/starlark"
	"gopkg.in/yaml.v2"
)

var _ = Describe("Credential", func() {

	Context("GetOrCreate", func() {
		It("reads username and password from k8s", func() {
			username := "username1"
			password := "password1"
			k8s := FakeK8s{
				GetStub: func(kind string, name string, writer io.Writer, k8s *K8sOptions) error {
					writer.Write([]byte("{\"data\": {" +
						"\"username\": \"" + base64.StdEncoding.EncodeToString([]byte(username)) + "\"," +
						"\"password\": \"" + base64.StdEncoding.EncodeToString([]byte(password)) + "\" " +
						"} }"))
					return nil
				},
			}
			userCred := &userCredential{}
			err := userCred.GetOrCreate(&k8s)
			Expect(err).NotTo(HaveOccurred())
			Expect(userCred.username).To(Equal(username))
			Expect(userCred.password).To(Equal(password))

			value, err := userCred.Attr("username")
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal(starlark.String(username)))
		})

		It("creates new random username and password if user_credential doesn't exist", func() {
			k8s := FakeK8s{
				GetStub: func(kind string, name string, writer io.Writer, k8s *K8sOptions) error {
					return errors.New("NotFound")
				},
				IsNotExistStub: func(err error) bool {
					return true
				},
			}
			userCred := &userCredential{}
			err := userCred.GetOrCreate(&k8s)
			Expect(err).NotTo(HaveOccurred())
			Expect(userCred.username).To(HaveLen(16))
			Expect(userCred.password).To(HaveLen(16))
			_, err = userCred.Attr("username")
			Expect(err).NotTo(HaveOccurred())
		})

		It("fails on other errors", func() {
			k8s := FakeK8s{
				GetStub: func(kind string, name string, writer io.Writer, k8s *K8sOptions) error {
					return errors.New("Other")
				},
				IsNotExistStub: func(err error) bool {
					return false
				},
			}
			userCred := &userCredential{}
			err := userCred.GetOrCreate(&k8s)
			Expect(err).To(HaveOccurred())
		})

	})

	It("behaves like starlark value", func() {
		userCred := &userCredential{name: "name", username: "username"}
		Expect(userCred.String()).To(ContainSubstring("name = name"))
		Expect(userCred.String()).To(ContainSubstring("username = username"))
		Expect(func() { userCred.Hash() }).Should(Panic())
		Expect(userCred.Type()).To(Equal("user_credential"))
		Expect(userCred.Truth()).To(BeEquivalentTo(false))
		value, err := userCred.Attr("name")
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(Equal(starlark.String("name")))
		Expect(userCred.AttrNames()).To(ContainElement("name"))

		userCred = &userCredential{name: "name", username: ""}
		value, err = userCred.Attr("username")
		Expect(err).To(HaveOccurred())
	})

	It("username and password can only be read after GetOrCreate", func() {
		userCred := &userCredential{name: "name"}

		_, err := userCred.Attr("username")
		Expect(err).To(HaveOccurred())
		_, err = userCred.Attr("password")
		Expect(err).To(HaveOccurred())
	})

	Context("dataMap", func() {
		Context("JSON", func() {
			It("Marshal correct", func() {
				d := dataMap{"key": []byte("value")}
				b, err := json.Marshal(d)
				Expect(err).NotTo(HaveOccurred())
				Expect(b).To(Equal([]byte(`{"key":"dmFsdWU="}`)))
			})
			It("Unmarshal correct", func() {
				d := dataMap{}
				err := json.Unmarshal([]byte(`{"key":"dmFsdWU="}`), &d)
				Expect(err).NotTo(HaveOccurred())
				Expect(d).To(HaveKeyWithValue("key", []byte("value")))
			})

		})
		Context("YAML", func() {
			It("Marshal correct", func() {
				d := dataMap{"key": []byte("value")}
				b, err := yaml.Marshal(d)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(Equal("key: dmFsdWU=\n"))
			})
			It("Unmarshal correct", func() {
				d := dataMap{}
				err := yaml.Unmarshal([]byte("key: dmFsdWU=\n"), &d)
				Expect(err).NotTo(HaveOccurred())
				Expect(d).To(HaveKeyWithValue("key", []byte("value")))
			})

		})
	})
})
