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

// DatabaseRetentionPolicyMappingParameters are the configurable fields of a DatabaseRetentionPolicyMapping.
type DatabaseRetentionPolicyMappingParameters struct {
	// BucketID is the ID of the Bucket this DatabaseRetentionPolicyMapping will
	// be applied.
	// Either BucketID or BucketIDRef or BucketIDSelector has to be given during
	// creation.
	// +crossplane:generate:reference:type=Organization
	// +crossplane:generate:reference:extractor=OrganizationID()
	BucketID string `json:"bucketID,omitempty"`

	// BucketIDRef references a Bucket to retrieve its ID to populate BucketID.
	// +optional
	// +immutable
	BucketIDRef *xpv1.Reference `json:"bucketIDRef,omitempty"`

	// BucketIDSelector selects a reference to a Bucket to populate BucketIDRef.
	// +optional
	BucketIDSelector *xpv1.Selector `json:"bucketIDSelector,omitempty"`

	// InfluxDB v1 database
	Database string `json:"database"`

	// Specify if this mapping represents the default retention policy for the database specificed.
	Default *bool `json:"default,omitempty"`

	// The organization that owns this mapping.
	// Either Org or OrgRef or OrgSelector has to be given during
	// creation.
	// +crossplane:generate:reference:type=Organization
	Org string `json:"org,omitempty"`

	// OrgRef references an Organization to retrieve its name to populate Org.
	// +optional
	// +immutable
	OrgRef *xpv1.Reference `json:"orgRef,omitempty"`

	// OrgSelector selects a reference to an Organization to populate OrgRef.
	// +optional
	OrgSelector *xpv1.Selector `json:"orgSelector,omitempty"`

	// InfluxDB v1 retention policy
	RetentionPolicy string `json:"retentionPolicy"`
}

// DatabaseRetentionPolicyMappingObservation are the observable fields of a DatabaseRetentionPolicyMapping.
type DatabaseRetentionPolicyMappingObservation struct {
	Links DBRPLinks `json:"links,omitempty"`
}

// DBRPLinks defines model for Links.
type DBRPLinks struct {
	// URI of resource.
	Self string `json:"self"`
}

// A DatabaseRetentionPolicyMappingSpec defines the desired state of a DatabaseRetentionPolicyMapping.
type DatabaseRetentionPolicyMappingSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       DatabaseRetentionPolicyMappingParameters `json:"forProvider"`
}

// A DatabaseRetentionPolicyMappingStatus represents the observed state of a DatabaseRetentionPolicyMapping.
type DatabaseRetentionPolicyMappingStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          DatabaseRetentionPolicyMappingObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// An DatabaseRetentionPolicyMapping represents an organization in InfluxDB.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,influxdb},shortName=dbrp
type DatabaseRetentionPolicyMapping struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseRetentionPolicyMappingSpec   `json:"spec"`
	Status DatabaseRetentionPolicyMappingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DatabaseRetentionPolicyMappingList contains a list of DatabaseRetentionPolicyMapping.
type DatabaseRetentionPolicyMappingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatabaseRetentionPolicyMapping `json:"items"`
}

// DatabaseRetentionPolicyMapping type metadata.
var (
	DatabaseRetentionPolicyMappingKind             = reflect.TypeOf(DatabaseRetentionPolicyMapping{}).Name()
	DatabaseRetentionPolicyMappingGroupKind        = schema.GroupKind{Group: Group, Kind: DatabaseRetentionPolicyMappingKind}.String()
	DatabaseRetentionPolicyMappingKindAPIVersion   = DatabaseRetentionPolicyMappingKind + "." + SchemeGroupVersion.String()
	DatabaseRetentionPolicyMappingGroupVersionKind = SchemeGroupVersion.WithKind(DatabaseRetentionPolicyMappingKind)
)

func init() {
	SchemeBuilder.Register(&DatabaseRetentionPolicyMapping{}, &DatabaseRetentionPolicyMappingList{})
}
