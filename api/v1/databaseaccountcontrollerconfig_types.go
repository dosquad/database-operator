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
	cfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

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
	DatabaseDSN string `json:"dsn,omitempty"`
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
