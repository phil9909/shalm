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

	"github.com/kramerul/shalm/pkg/shalm"
	"go.starlark.net/starlark"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	shalmv1a1 "github.com/kramerul/shalm/api/v1alpha1"
)

// ShalmChartReconciler reconciles a ShalmChart object
type ShalmChartReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Repo   shalm.Repo
	K8s    func(kubeconfig string) shalm.K8s
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
	thread := &starlark.Thread{Name: "main"}
	spec := &shalmChart.Spec
	chart, err := r.Repo.GetFromGo(thread, spec.URL, spec.Namespace, false, spec.Args, spec.KwArgs)
	if err != nil {
		return result, err
	}
	err = chart.Apply(thread, r.K8s(spec.KubeConfig))

	return result, nil
}

// SetupWithManager -
func (r *ShalmChartReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&shalmv1a1.ShalmChart{}).
		Complete(r)
}
