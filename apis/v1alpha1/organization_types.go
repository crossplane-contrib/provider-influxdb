/*
Copyright 2021 The Crossplane Authors.

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

package v1alpha1

import (
	"reflect"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// OrganizationParameters are the configurable fields of a Organization.
type OrganizationParameters struct {
	Description *string `json:"description,omitempty"`
}

// OrganizationObservation are the observable fields of a Organization.
type OrganizationObservation struct {
	ID        string            `json:"id,omitempty"`
	Status    string            `json:"status,omitempty"`
	CreatedAt metav1.Time       `json:"createdAt,omitempty"`
	UpdatedAt metav1.Time       `json:"updatedAt,omitempty"`
	Links     OrganizationLinks `json:"links,omitempty"`
}

// OrganizationLinks is the URIs of all links.
type OrganizationLinks struct {
	// URI of resource.
	Buckets string `json:"buckets,omitempty"`

	// URI of resource.
	Dashboards string `json:"dashboards,omitempty"`

	// URI of resource.
	Labels string `json:"labels,omitempty"`

	// URI of resource.
	Members string `json:"members,omitempty"`

	// URI of resource.
	Owners string `json:"owners,omitempty"`

	// URI of resource.
	Secrets string `json:"secrets,omitempty"`

	// URI of resource.
	Self string `json:"self,omitempty"`

	// URI of resource.
	Tasks string `json:"tasks,omitempty"`
}

// A OrganizationSpec defines the desired state of a Organization.
type OrganizationSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       OrganizationParameters `json:"forProvider"`
}

// A OrganizationStatus represents the observed state of a Organization.
type OrganizationStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          OrganizationObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// An Organization represents an organization in InfluxDB.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,influxdb}
type Organization struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OrganizationSpec   `json:"spec"`
	Status OrganizationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OrganizationList contains a list of Organization.
type OrganizationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Organization `json:"items"`
}

// Organization type metadata.
var (
	OrganizationKind             = reflect.TypeOf(Organization{}).Name()
	OrganizationGroupKind        = schema.GroupKind{Group: Group, Kind: OrganizationKind}.String()
	OrganizationKindAPIVersion   = OrganizationKind + "." + SchemeGroupVersion.String()
	OrganizationGroupVersionKind = SchemeGroupVersion.WithKind(OrganizationKind)
)

func init() {
	SchemeBuilder.Register(&Organization{}, &OrganizationList{})
}
