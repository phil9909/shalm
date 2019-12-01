package impl

import (
	"encoding/base64"
	"errors"
	"io"

	"github.com/kramerul/shalm/pkg/chart/api"
	"github.com/kramerul/shalm/pkg/chart/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.starlark.net/starlark"
)

var _ = Describe("Credential", func() {

	Context("GetOrCreate", func() {
		It("reads username and password from k8s", func() {
			username := "username1"
			password := "password1"
			k8s := fakes.FakeK8s{
				GetStub: func(kind string, name string, writer io.Writer, k8s *api.K8sOptions) error {
					writer.Write([]byte("data:\n" +
						"  username: " + base64.StdEncoding.EncodeToString([]byte(username)) + "\n" +
						"  password: " + base64.StdEncoding.EncodeToString([]byte(password)) + "\n"))
					return nil
				},
			}
			credential := &credential{}
			err := credential.GetOrCreate(&k8s)
			Expect(err).NotTo(HaveOccurred())
			Expect(credential.username).To(Equal(username))
			Expect(credential.password).To(Equal(password))

			value, err := credential.Attr("username")
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal(starlark.String(username)))
		})

		It("creates new random username and password if secret doesn't exist", func() {
			k8s := fakes.FakeK8s{
				GetStub: func(kind string, name string, writer io.Writer, k8s *api.K8sOptions) error {
					return errors.New("NotFound")
				},
				IsNotExistStub: func(err error) bool {
					return true
				},
			}
			credential := &credential{}
			err := credential.GetOrCreate(&k8s)
			Expect(err).NotTo(HaveOccurred())
			Expect(credential.username).To(HaveLen(16))
			Expect(credential.password).To(HaveLen(16))
			_, err = credential.Attr("username")
			Expect(err).NotTo(HaveOccurred())
		})

		It("fails on other errors", func() {
			k8s := fakes.FakeK8s{
				GetStub: func(kind string, name string, writer io.Writer, k8s *api.K8sOptions) error {
					return errors.New("Other")
				},
				IsNotExistStub: func(err error) bool {
					return false
				},
			}
			credential := &credential{}
			err := credential.GetOrCreate(&k8s)
			Expect(err).To(HaveOccurred())
		})

	})

	It("behaves like starlark value", func() {
		credential := &credential{name: "name", username: "username"}
		Expect(credential.String()).To(ContainSubstring("name = name"))
		Expect(credential.String()).To(ContainSubstring("username = username"))
		Expect(func() { credential.Hash() }).Should(Panic())
		Expect(credential.Truth()).To(BeEquivalentTo(false))
		value, err := credential.Attr("name")
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(Equal(starlark.String("name")))
		Expect(credential.AttrNames()).To(ContainElement("name"))

		value, err = credential.Attr("username")
		Expect(err).To(HaveOccurred())
	})

	It("username and password can only be read after GetOrCreate", func() {
		credential := &credential{name: "name", username: "username"}

		_, err := credential.Attr("username")
		Expect(err).To(HaveOccurred())
		_, err = credential.Attr("password")
		Expect(err).To(HaveOccurred())
	})

})
