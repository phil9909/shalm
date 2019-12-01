package impl

import (
	"encoding/base64"
	"errors"
	"io"

	"github.com/kramerul/shalm/pkg/chart"
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
				GetStub: func(kind string, name string, writer io.Writer, k8s *chart.K8sOptions) error {
					writer.Write([]byte("data:\n" +
						"  username: " + base64.StdEncoding.EncodeToString([]byte(username)) + "\n" +
						"  password: " + base64.StdEncoding.EncodeToString([]byte(password)) + "\n"))
					return nil
				},
			}
			user_credential := &userCredential{}
			err := user_credential.GetOrCreate(&k8s)
			Expect(err).NotTo(HaveOccurred())
			Expect(user_credential.username).To(Equal(username))
			Expect(user_credential.password).To(Equal(password))

			value, err := user_credential.Attr("username")
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal(starlark.String(username)))
		})

		It("creates new random username and password if user_credential doesn't exist", func() {
			k8s := fakes.FakeK8s{
				GetStub: func(kind string, name string, writer io.Writer, k8s *chart.K8sOptions) error {
					return errors.New("NotFound")
				},
				IsNotExistStub: func(err error) bool {
					return true
				},
			}
			user_credential := &userCredential{}
			err := user_credential.GetOrCreate(&k8s)
			Expect(err).NotTo(HaveOccurred())
			Expect(user_credential.username).To(HaveLen(16))
			Expect(user_credential.password).To(HaveLen(16))
			_, err = user_credential.Attr("username")
			Expect(err).NotTo(HaveOccurred())
		})

		It("fails on other errors", func() {
			k8s := fakes.FakeK8s{
				GetStub: func(kind string, name string, writer io.Writer, k8s *chart.K8sOptions) error {
					return errors.New("Other")
				},
				IsNotExistStub: func(err error) bool {
					return false
				},
			}
			user_credential := &userCredential{}
			err := user_credential.GetOrCreate(&k8s)
			Expect(err).To(HaveOccurred())
		})

	})

	It("behaves like starlark value", func() {
		user_credential := &userCredential{name: "name", username: "username"}
		Expect(user_credential.String()).To(ContainSubstring("name = name"))
		Expect(user_credential.String()).To(ContainSubstring("username = username"))
		Expect(func() { user_credential.Hash() }).Should(Panic())
		Expect(user_credential.Truth()).To(BeEquivalentTo(false))
		value, err := user_credential.Attr("name")
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(Equal(starlark.String("name")))
		Expect(user_credential.AttrNames()).To(ContainElement("name"))

		value, err = user_credential.Attr("username")
		Expect(err).To(HaveOccurred())
	})

	It("username and password can only be read after GetOrCreate", func() {
		user_credential := &userCredential{name: "name", username: "username"}

		_, err := user_credential.Attr("username")
		Expect(err).To(HaveOccurred())
		_, err = user_credential.Attr("password")
		Expect(err).To(HaveOccurred())
	})

})
