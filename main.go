/*
Copyright 2022.

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
	"strings"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	keystonev1 "github.com/openstack-k8s-operators/keystone-operator/api/v1beta1"

	rabbitmqv1 "github.com/openstack-k8s-operators/infra-operator/apis/rabbitmq/v1beta1"
	ansibleeev1 "github.com/openstack-k8s-operators/openstack-ansibleee-operator/api/v1alpha1"

	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	telemetryv1beta1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"
	"github.com/openstack-k8s-operators/telemetry-operator/controllers"
	"github.com/openstack-k8s-operators/telemetry-operator/pkg/common"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(telemetryv1.AddToScheme(scheme))
	utilruntime.Must(keystonev1.AddToScheme(scheme))
	utilruntime.Must(rabbitmqv1.AddToScheme(scheme))
	utilruntime.Must(ansibleeev1.AddToScheme(scheme))
	utilruntime.Must(telemetryv1beta1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8083", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8084", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "fa1814a2.openstack.org",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	cfg, err := config.GetConfig()
	if err != nil {
		setupLog.Error(err, "")
		os.Exit(1)
	}
	kclient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		setupLog.Error(err, "")
		os.Exit(1)
	}

	if err = (&controllers.TelemetryReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		Kclient: kclient,
		Log:     ctrl.Log.WithName("controllers").WithName("Telemetry"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create Telemetry controller")
		os.Exit(1)
	}

	if err = (&controllers.CeilometerCentralReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		Kclient: kclient,
		Log:     ctrl.Log.WithName("controllers").WithName("CeilometerCentral"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create CeilometerCentral controller")
		os.Exit(1)
	}

	if err = (&controllers.CeilometerComputeReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		Kclient: kclient,
		Log:     ctrl.Log.WithName("controllers").WithName("CeilometerCompute"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create CeilometerCompute controller")
		os.Exit(1)
	}

	if err = (&controllers.InfraComputeReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		Kclient: kclient,
		Log:     ctrl.Log.WithName("controllers").WithName("InfraCompute"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create InfraCompute controller")
		os.Exit(1)
	}

	// Acquire environmental defaults and initialize Telemetry defaults with them
	telemetryDefaults := telemetryv1.TelemetryDefaults{
		CentralContainerImageURL:      common.GetEnvDefault("CEILOMETER_CENTRAL_IMAGE_URL_DEFAULT", telemetryv1.CeilometerCentralContainerImage),
		CentralInitContainerImageURL:  common.GetEnvDefault("CEILOMETER_CENTRAL_INIT_IMAGE_URL_DEFAULT", telemetryv1.CeilometerCentralInitContainerImage),
		ComputeContainerImageURL:      common.GetEnvDefault("CEILOMETER_COMPUTE_IMAGE_URL_DEFAULT", telemetryv1.CeilometerComputeContainerImage),
		ComputeInitContainerImageURL:  common.GetEnvDefault("CEILOMETER_COMPUTE_INIT_IMAGE_URL_DEFAULT", telemetryv1.CeilometerComputeInitContainerImage),
		NotificationContainerImageURL: common.GetEnvDefault("CEILOMETER_NOTIFICATION_IMAGE_URL_DEFAULT", telemetryv1.CeilometerNotificationContainerImage),
		NodeExporterContainerImageURL: common.GetEnvDefault("NODE_EXPORTER_IMAGE_URL_DEFAULT", telemetryv1.NodeExporterContainerImage),
		SgCoreContainerImageURL:       common.GetEnvDefault("CEILOMETER_SGCORE_IMAGE_URL_DEFAULT", telemetryv1.CeilometerSgCoreContainerImage),
	}

	telemetryv1.SetupTelemetryDefaults(telemetryDefaults)

	// Setup webhooks if requested
	if strings.ToLower(os.Getenv("ENABLE_WEBHOOKS")) != "false" {
		if err = (&telemetryv1.Telemetry{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Telemetry")
			os.Exit(1)
		}
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
