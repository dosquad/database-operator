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
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DatabaseAccountSpec defines the desired state of DatabaseAccount.
type DatabaseAccountSpec struct {
	// Username is the username for the postgresql Database account.
	Username string `json:"username,omitempty"`

	// OnDelete specifies if the database should be removed when the user is removed.
	//+optional
	//+kubebuilder:default:=delete
	OnDelete DatabaseAccountOnDelete `json:"onDelete,omitempty"`

	// Name is the basename used for the resource, if not specified a UUID will be used.
	//+optional
	Name PostgreSQLResourceName `json:"name,omitempty"`

	// SecretName is the optional name for the secret created with the DSN.
	//+optional
	SecretName string `json:"secretName,omitempty"`

	// SecretTemplate is the optional spec to be added to secrets generated by the DatabaseAccountSpec.
	SecretTemplate DatabaseAccountSpecSecretTemplate `json:"secretTemplate,omitempty"`
}

// DatabaseAccountSpecSecretTemplate defines the desired state of DatabaseAccount.
type DatabaseAccountSpecSecretTemplate struct {
	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May match selectors of replication controllers
	// and services.
	// More info: http://kubernetes.io/docs/user-guide/labels
	//+optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	// More info: http://kubernetes.io/docs/user-guide/annotations
	//+optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

// DatabaseAccountStatus defines the observed state of DatabaseAccount.
type DatabaseAccountStatus struct {
	// State is the progress of creating the account.
	//
	// +optional
	// +kubebuilder:default:=Init
	Stage DatabaseAccountCreateStage `json:"stage,omitempty"`

	// Error is true if the DatabaseAccount is in error.
	//
	// +optional
	Error bool `json:"error,omitempty"`

	// ErrorMessage is the message if the Stage is Error.
	//
	// +optional
	ErrorMessage string `json:"errorMsg,omitempty"`

	// Name is the basename used for the resource.
	Name PostgreSQLResourceName `json:"name,omitempty"`

	// Ready is the boolean for when a resource is ready to use.
	//
	// +kubebuilder:default:=false
	Ready bool `json:"ready,omitempty"`
}

// DatabaseAccount is the Schema for the databaseaccounts API
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Stage",type=string,JSONPath=`.status.stage`,description="deployment stage for database account"
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.ready`,description="ready status of database account"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type DatabaseAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseAccountSpec   `json:"spec,omitempty"`
	Status DatabaseAccountStatus `json:"status,omitempty"`
}

func (d *DatabaseAccount) GetReference() *metav1.OwnerReference {
	return &metav1.OwnerReference{
		APIVersion: d.APIVersion,
		Kind:       d.Kind,
		Name:       d.GetName(),
		UID:        d.GetUID(),
	}
}

func (d *DatabaseAccount) GetSecretName() types.NamespacedName {
	if d.Spec.SecretName != "" {
		return types.NamespacedName{
			Namespace: d.GetNamespace(),
			Name:      d.Spec.SecretName,
		}
	}

	return types.NamespacedName{
		Namespace: d.GetNamespace(),
		Name:      d.GetName(),
	}
}

func (d *DatabaseAccount) GetDatabaseName() (string, error) {
	if len(d.Status.Name) == 0 {
		return "", ErrMissingDatabaseUsername
	}

	return d.Status.Name.String(), nil
}

func (d *DatabaseAccount) UpdateStatus(ctx context.Context, r client.StatusClient) error {
	return r.Status().Update(ctx, d)
}

func (d *DatabaseAccount) Update(ctx context.Context, r client.Client) error {
	return r.Update(ctx, d)
}

// DatabaseAccountList contains a list of DatabaseAccount
//
// +kubebuilder:object:root=true
type DatabaseAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatabaseAccount `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DatabaseAccount{}, &DatabaseAccountList{})
}
