package v1

import (
	"time"

	configv1alpha1 "k8s.io/component-base/config/v1alpha1"
)

// ControllerManagerConfiguration defines the embedded RuntimeConfiguration for controller-runtime clients.
type ControllerManagerConfiguration struct {
	Namespace string `json:"namespace,omitempty"`

	SyncPeriod *time.Duration `json:"syncPeriod,omitempty"`

	LeaderElection configv1alpha1.LeaderElectionConfiguration `json:"leaderElection,omitempty"`

	MetricsBindAddress string `json:"metricsBindAddress,omitempty"`

	Health ControllerManagerConfigurationHealth `json:"health,omitempty"`

	Port *int   `json:"port,omitempty"`
	Host string `json:"host,omitempty"`

	CertDir string `json:"certDir,omitempty"`
}

// ControllerManagerConfigurationHealth defines the health configs.
type ControllerManagerConfigurationHealth struct {
	HealthProbeBindAddress string `json:"healthProbeBindAddress,omitempty"`

	ReadinessEndpointName string `json:"readinessEndpointName,omitempty"`
	LivenessEndpointName  string `json:"livenessEndpointName,omitempty"`
}
