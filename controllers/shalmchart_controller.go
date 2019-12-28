/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"reflect"

	"github.com/kramerul/shalm/pkg/shalm"
	"go.starlark.net/starlark"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"

	shalmv1a1 "github.com/kramerul/shalm/api/v1alpha1"
)

var myFinalizerName = "controller.shalm.kramerul.github.com"

// ShalmChartReconciler reconciles a ShalmChart object
type ShalmChartReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Repo   shalm.Repo
	K8s    func(kubeconfig string) shalm.K8s
}

type shalmChartPredicate struct {
}

// +kubebuilder:rbac:groups=shalm.kramerul.github.com,resources=shalmcharts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=shalm.kramerul.github.com,resources=shalmcharts/status,verbs=get;update;patch

// Reconcile -
func (r *ShalmChartReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	result := ctrl.Result{}
	ctx := context.Background()
	_ = r.Log.WithValues("shalmchart", req.NamespacedName)

	var shalmChart shalmv1a1.ShalmChart
	err := r.Client.Get(ctx, client.ObjectKey{Name: req.Name, Namespace: req.Namespace}, &shalmChart)
	if err != nil {
		return result, err
	}
	if shalmChart.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(shalmChart.ObjectMeta.Finalizers, myFinalizerName) {
			shalmChart.ObjectMeta.Finalizers = append(shalmChart.ObjectMeta.Finalizers, myFinalizerName)
			if err := r.Update(context.Background(), &shalmChart); err != nil {
				return result, err
			}
		}
		return result, r.apply(&shalmChart.Spec)
	}
	if containsString(shalmChart.ObjectMeta.Finalizers, myFinalizerName) {
		if err := r.delete(&shalmChart.Spec); err != nil {
			return result, err
		}

		shalmChart.ObjectMeta.Finalizers = removeString(shalmChart.ObjectMeta.Finalizers, myFinalizerName)
		if err := r.Update(context.Background(), &shalmChart); err != nil {
			return result, err
		}
	}

	return result, err

}

func (r *ShalmChartReconciler) apply(spec *shalmv1a1.ShalmChartSpec) error {
	thread := &starlark.Thread{Name: "main"}
	chart, err := r.Repo.GetFromSpec(thread, spec)
	if err != nil {
		return err
	}
	return chart.Apply(thread, r.K8s(spec.KubeConfig))
}

func (r *ShalmChartReconciler) delete(spec *shalmv1a1.ShalmChartSpec) error {
	thread := &starlark.Thread{Name: "main"}
	chart, err := r.Repo.GetFromSpec(thread, spec)
	if err != nil {
		return err
	}
	return chart.Delete(thread, r.K8s(spec.KubeConfig))
}

// SetupWithManager -
func (r *ShalmChartReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&shalmv1a1.ShalmChart{}).
		WithEventFilter(&shalmChartPredicate{}).
		Complete(r)
}

// Create -
func (r *shalmChartPredicate) Create(event.CreateEvent) bool {
	return true
}

// Delete -
func (r *shalmChartPredicate) Delete(event.DeleteEvent) bool {
	return false
}

// Update -
func (r *shalmChartPredicate) Update(ev event.UpdateEvent) bool {
	old := ev.ObjectOld.(*shalmv1a1.ShalmChart)
	new := ev.ObjectNew.(*shalmv1a1.ShalmChart)
	if !reflect.DeepEqual(old.Spec, new.Spec) {
		return true
	}
	if !containsString(new.ObjectMeta.Finalizers, myFinalizerName) {
		return true
	}
	if !new.ObjectMeta.DeletionTimestamp.IsZero() {
		return true
	}
	return false
}

// Generic -
func (r *shalmChartPredicate) Generic(event.GenericEvent) bool {
	return false
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
