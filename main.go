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

package main

import (
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/bharathi-tenneti/memcached-operator-metrics/api/metrics"
	cachev1alpha1 "github.com/bharathi-tenneti/memcached-operator-metrics/api/v1alpha1"
	"github.com/bharathi-tenneti/memcached-operator-metrics/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme          = runtime.NewScheme()
	setupLog        = ctrl.Log.WithName("setup")
	metricsRegistry metrics.RegistererGathererPredicater
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = cachev1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme

	metricsRegistry = metrics.NewDefaultRegistry()

}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "f1c5ece8.example.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err := mgr.Add(&metrics.Server{
		Gatherer:      metricsRegistry,
		ListenAddress: "0.0.0.0:8686",
	}); err != nil {
		os.Exit(1)
	}
	crInfo := metrics.NewCRInfoGauge()
	timeInfo := metrics.NewTimeInfo()
	summaryInfo := metrics.NewSummaryInfo()

	metricsRegistry.MustRegister(crInfo)
	metricsRegistry.MustRegister(timeInfo)
	metricsRegistry.MustRegister(summaryInfo)

	var predicates []predicate.Predicate
	predicates = append(predicates, metricsRegistry.Predicate())

	if err = (&controllers.MemcachedReconciler{
		Client:  mgr.GetClient(),
		Log:     ctrl.Log.WithName("controllers").WithName("Memcached"),
		Scheme:  mgr.GetScheme(),
		TimeVec: timeInfo,
	}).SetupWithManager(mgr, predicates...); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Memcached")
		os.Exit(1)
	}

	// Reconciler for watching only Metrics.
	if err = (&controllers.MemcachedMetricsReconciler{
		Client:     mgr.GetClient(),
		Log:        ctrl.Log.WithName("controllers").WithName("Memcached_metrics"),
		Scheme:     mgr.GetScheme(),
		SummaryVec: summaryInfo,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Memcached")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
