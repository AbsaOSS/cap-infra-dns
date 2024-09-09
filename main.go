package main

import (
	epclient "github.com/absaoss/cap-infra-dns/pkg/client"
	"os"

	"github.com/absaoss/cap-infra-dns/pkg/controller"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/component-base/version"
	"k8s.io/klog/v2"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

var (
	runtimeScheme = runtime.NewScheme()
	setupLog      = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clusterv1.AddToScheme(runtimeScheme))
	utilruntime.Must(epclient.AddToScheme(runtimeScheme))
}

func main() {
	ctrl.SetLogger(klog.Background())

	ctx := ctrl.SetupSignalHandler()
	opts := ctrl.Options{
		Scheme: runtimeScheme,
	}
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), opts)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controller.ClusterReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(ctx, mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DNSEndpoint")
		os.Exit(1)
	}

	//	if err = index.SetupIndexes(ctx, mgr); err != nil {
	//		setupLog.Error(err, "failed to setup indexes")
	//		os.Exit(1)
	//	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager", "version", version.Get().String())
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
