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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	configv1alpha1 "k8s.io/component-base/config/v1alpha1"
	cfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

const ()

// DatabaseAccountControllerConfig is the Schema for the databaseaccountcontrollerconfigs API.
//
// +kubebuilder:object:root=true
type DatabaseAccountControllerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// ControllerManagerConfigurationSpec returns the contfigurations for controllers
	cfg.ControllerManagerConfigurationSpec `json:",inline"`

	Debug DatabaseAccountControllerConfigDebug `json:"debug,omitempty"`

	// DatabaseDSN is the DSN for the database that will be used for creating accounts and databases on.
	DatabaseDSN PostgreSQLDSN `json:"dsn,omitempty"`

	// RelayImage is the image used for the relay pod.
	//+optional
	// +kubebuilder:default:="edoburu/pgbouncer:1.20.1-p0"
	RelayImage string `json:"relayImage,omitempty"`

	// LeaderElection config
	//+optional
	LeaderElection *configv1alpha1.LeaderElectionConfiguration `json:"leaderElection,omitempty"`
}

// DatabaseAccountControllerConfigList contains a list of DatabaseAccountControllerConfig.
//
// +kubebuilder:object:root=true
type DatabaseAccountControllerConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatabaseAccountControllerConfig `json:"items"`
}

type DatabaseAccountControllerConfigDebug struct {
	ReconcileSleep int `json:"reconcileSleep,omitempty"`
}

func init() {
	SchemeBuilder.Register(&DatabaseAccountControllerConfig{}, &DatabaseAccountControllerConfigList{})
}

func (d *DatabaseAccountControllerConfig) GetRelayImage() string {
	if d.RelayImage == "" {
		return DefaultRelayImage
	}

	return d.RelayImage
}

func (d *DatabaseAccountControllerConfig) GetDSNHost() string {
	return d.DatabaseDSN.Host()
}
