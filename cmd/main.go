/*
Copyright 2023.

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
	"context"
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/dosquad/database-operator/accountsvr"
	dbov1 "github.com/dosquad/database-operator/api/v1"
	"github.com/dosquad/database-operator/internal/controller"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	cfg "sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//+kubebuilder:scaffold:imports
)

const (
	defaultLeaderElectionID = "ade51e3.dosquad.github.io"
	controllerName          = "databaseaccount-controller"
	controllerPort          = 9443
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(dbov1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	if err := mainCommand(); err != nil {
		os.Exit(1)
	}
}

func mainCommand() error {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var configFile string
	var err error

	setupLog := ctrl.Log.WithName("setup")

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&configFile, "config", "",
		"The controller will load its initial configuration from this file. "+
			"Omit this flag to use the default configuration values. "+
			"Command-line flags override configuration from this file.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		// Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	ctrlConfig := dbov1.DatabaseAccountControllerConfig{}
	options := ctrl.Options{
		Scheme:                        scheme,
		MetricsBindAddress:            metricsAddr,
		Port:                          controllerPort,
		HealthProbeBindAddress:        probeAddr,
		LeaderElection:                enableLeaderElection,
		LeaderElectionID:              defaultLeaderElectionID,
		LeaderElectionReleaseOnCancel: true,
	}
	if configFile != "" {
		options, err = options.AndFrom(cfg.File().AtPath(configFile).OfKind(&ctrlConfig))
		if err != nil {
			setupLog.Error(err, "unable to load the config file")
			return err
		}
	}

	mgr, mgrErr := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if mgrErr != nil {
		setupLog.Error(mgrErr, "unable to start manager")
		return mgrErr
	}

	svr, svrErr := accountsvr.NewServer(context.Background(), ctrlConfig.DatabaseDSN)
	if svrErr != nil {
		setupLog.Error(svrErr, "unable to start database connection")
		return svrErr
	}
	defer svr.Close(context.Background())

	if err = (&controller.DatabaseAccountReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		Recorder:      controller.NewRecorder(mgr.GetEventRecorderFor(controllerName)),
		AccountServer: svr,
		Config:        &ctrlConfig,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DatabaseAccount")
		return err
	}
	//+kubebuilder:scaffold:builder

	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		return err
	}
	if err = mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		return err
	}

	setupLog.Info("starting manager")
	if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		return err
	}

	return nil
}
