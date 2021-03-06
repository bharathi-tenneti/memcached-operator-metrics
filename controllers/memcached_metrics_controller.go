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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/bharathi-tenneti/memcached-operator-metrics/api/metrics"
	cachev1alpha1 "github.com/bharathi-tenneti/memcached-operator-metrics/api/v1alpha1"
)

// MemcachedMetricsReconciler reconciles a Memcached object`
type MemcachedMetricsReconciler struct {
	client.Client
	Log                     logr.Logger
	Scheme                  *runtime.Scheme
	maxConcurrentReconciles int
	SummaryVec              *metrics.SummaryInfo
}

// +kubebuilder:rbac:groups=cache.example.com,resources=memcacheds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cache.example.com,resources=memcacheds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

func (r *MemcachedMetricsReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("memcached", req.NamespacedName)

	// Fetch the Memcached instance
	memcached := &cachev1alpha1.Memcached{}

	err := r.Get(ctx, req.NamespacedName, memcached)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Memcached resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Memcached")
		return ctrl.Result{}, err
	}

	labels := map[string]string{
		"name":       memcached.Name,
		"namespace":  memcached.Namespace,
		"apiversion": memcached.APIVersion,
		"kind":       memcached.Kind,
	}
	m, metricErr := r.SummaryVec.GetMetricWith(labels)
	if metricErr != nil {
		panic(metricErr)
	}
	// Delete metrics if no memcached resources are found.
	if memcached.GetFinalizers() != nil && memcached.GetDeletionTimestamp() != nil {
		for _, f := range memcached.GetFinalizers() {
			if f == "cleanup-summary-metrics" {
				r.SummaryVec.Delete(labels)
				controllerutil.RemoveFinalizer(memcached, "cleanup-summary-metrics")
				r.Update(ctx, memcached)
				return ctrl.Result{}, nil
			}
		}
	}
	// set the Finalizer and metrics for memcached
	controllerutil.AddFinalizer(memcached, "cleanup-summary-metrics")
	r.Update(ctx, memcached)
	m.SetToCurrentTime()
	return ctrl.Result{}, nil
}

// SetupWithManager ...
func (r *MemcachedMetricsReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.Memcached{}).
		Complete(r)
}
