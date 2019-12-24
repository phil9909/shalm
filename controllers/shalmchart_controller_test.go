package controllers

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o client_test.go sigs.k8s.io/controller-runtime/pkg/client.Client
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o ./fake_k8s_test.go ../pkg/shalm K8s

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	goruntime "runtime"
	"time"

	"github.com/kramerul/shalm/pkg/shalm"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	shalmv1a1 "github.com/kramerul/shalm/api/v1alpha1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	_, b, _, _ = goruntime.Caller(0)
	basepath   = filepath.Dir(b)
	root       = path.Join(filepath.Dir(b), "..")
	example    = path.Join(root, "example", "simple")
)

var _ = Describe("ShalmChartReconciler", func() {

	var server http.Server
	http.HandleFunc("/mariadb.tgz", func(w http.ResponseWriter, r *http.Request) {
		content, _ := ioutil.ReadFile(path.Join(example, "mariadb-6.12.2.tgz"))
		w.Write(content)
	})

	BeforeEach(func() {
		server = http.Server{Addr: "127.0.0.1:8675", Handler: nil}
		go server.ListenAndServe()
	})

	AfterEach(func() {
		server.Shutdown(context.Background())
	})

	It("applies shalm chart correct", func() {

		buffer := &bytes.Buffer{}
		k8s := &FakeK8s{
			ApplyStub: func(cb func(io.Writer) error, options *shalm.K8sOptions) error {
				return cb(buffer)
			},
		}
		k8s.ForNamespaceStub = func(s string) shalm.K8s {
			return k8s
		}
		chart := shalmv1a1.ShalmChart{
			Spec: shalmv1a1.ShalmChartSpec{
				Values:     nil,
				Args:       nil,
				KwArgs:     nil,
				KubeConfig: "",
				Namespace:  "",
				URL:        "http://localhost:8675/mariadb.tgz",
			},
		}

		client := &FakeClient{
			GetStub: func(ctx context.Context, name types.NamespacedName, object runtime.Object) error {
				switch object := object.(type) {
				case *shalmv1a1.ShalmChart:
					chart.DeepCopyInto(object)
					return nil
				}
				return apierrors.NewNotFound(schema.GroupResource{}, name.String())
			},
			UpdateStub: func(ctx context.Context, object runtime.Object, options ...client.UpdateOption) error {
				switch object := object.(type) {
				case *shalmv1a1.ShalmChart:
					object.DeepCopyInto(&chart)
					return nil
				}
				return apierrors.NewNotFound(schema.GroupResource{}, "???")
			},
		}
		reconciler := ShalmChartReconciler{
			Client: client,
			Log:    ctrl.Log.WithName("reconciler"),
			Scheme: nil,
			Repo:   shalm.NewRepo(),
			K8s: func(kubeconfig string) shalm.K8s {
				return k8s
			},
		}
		_, err := reconciler.Reconcile(ctrl.Request{})
		Expect(err).NotTo(HaveOccurred())
		Expect(buffer.String()).To(ContainSubstring("serviceName: mariadb-master"))
		Expect(chart.ObjectMeta.Finalizers).To(ContainElement("controller.shalm.kramerul.github.com"))
		Expect(k8s.ApplyCallCount()).To(Equal(1))
	})
	It("deletes shalm chart correct", func() {

		buffer := &bytes.Buffer{}
		k8s := &FakeK8s{
			DeleteStub: func(cb func(io.Writer) error, options *shalm.K8sOptions) error {
				return cb(buffer)
			},
		}
		k8s.ForNamespaceStub = func(s string) shalm.K8s {
			return k8s
		}
		chart := shalmv1a1.ShalmChart{
			ObjectMeta: v1.ObjectMeta{
				Finalizers: []string{"controller.shalm.kramerul.github.com"},
				DeletionTimestamp: &v1.
					Time{Time: time.Now()},
			},
			Spec: shalmv1a1.ShalmChartSpec{
				Values:     nil,
				Args:       nil,
				KwArgs:     nil,
				KubeConfig: "",
				Namespace:  "",
				URL:        "http://localhost:8675/mariadb.tgz",
			},
		}

		client := &FakeClient{
			GetStub: func(ctx context.Context, name types.NamespacedName, object runtime.Object) error {
				switch object := object.(type) {
				case *shalmv1a1.ShalmChart:
					chart.DeepCopyInto(object)
					return nil
				}
				return apierrors.NewNotFound(schema.GroupResource{}, name.String())
			},
			UpdateStub: func(ctx context.Context, object runtime.Object, options ...client.UpdateOption) error {
				switch object := object.(type) {
				case *shalmv1a1.ShalmChart:
					object.DeepCopyInto(&chart)
					return nil
				}
				return apierrors.NewNotFound(schema.GroupResource{}, "???")
			},
		}
		reconciler := ShalmChartReconciler{
			Client: client,
			Log:    ctrl.Log.WithName("reconciler"),
			Scheme: nil,
			Repo:   shalm.NewRepo(),
			K8s: func(kubeconfig string) shalm.K8s {
				return k8s
			},
		}
		_, err := reconciler.Reconcile(ctrl.Request{})
		Expect(err).NotTo(HaveOccurred())
		Expect(chart.ObjectMeta.Finalizers).NotTo(ContainElement("controller.shalm.kramerul.github.com"))
		Expect(k8s.DeleteCallCount()).To(Equal(1))
		Expect(buffer.String()).To(ContainSubstring("serviceName: mariadb-master"))
	})
})
