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

// BucketParameters are the configurable fields of a Bucket.
type BucketParameters struct {
	Description *string `json:"description,omitempty"`

	// OrgID is the ID of the org this Bucket will be a member of.
	// Either OrgID or OrgIDRef or OrgIDSelector has to be given during creation.
	// +crossplane:generate:reference:type=Organization
	// +crossplane:generate:reference:extractor=OrganizationID()
	OrgID *string `json:"orgID,omitempty"`

	// OrgIDRef references an Organization to retrieve its ID to populate OrgID.
	// +optional
	// +immutable
	OrgIDRef *xpv1.Reference `json:"orgIDRef,omitempty"`

	// OrgIDSelector selects a reference to an Organization to populate OrgIDRef.
	// +optional
	OrgIDSelector *xpv1.Selector `json:"orgIDSelector,omitempty"`

	RP *string `json:"rp,omitempty"`

	// Rules to expire or retain data. No rules means data never expires.
	RetentionRules []RetentionRule `json:"retentionRules"`

	// +kubebuilder:validation:Enum=implicit;explicit
	SchemaType string `json:"schemaType,omitempty"`
}

// RetentionRule defines model for RetentionRule.
type RetentionRule struct {
	// Duration in seconds for how long data will be kept in the database. 0 means infinite.
	EverySeconds int64 `json:"everySeconds"`

	// Shard duration measured in seconds.
	ShardGroupDurationSeconds *int64 `json:"shardGroupDurationSeconds,omitempty"`

	// +kubebuilder:default=expire
	Type string `json:"type"`
}

// BucketObservation are the observable fields of a Bucket.
type BucketObservation struct {
	ID        string      `json:"id,omitempty"`
	CreatedAt metav1.Time `json:"createdAt,omitempty"`
	UpdatedAt metav1.Time `json:"updatedAt,omitempty"`
	Links     BucketLinks `json:"links,omitempty"`
	Type      string      `json:"type,omitempty"`
	Labels    []Label     `json:"labels,omitempty"`
}

// BucketLinks is the URIs of all links.
type BucketLinks struct {
	// URI of resource.
	Labels string `json:"labels,omitempty"`

	// URI of resource.
	Members string `json:"members,omitempty"`

	// URI of resource.
	Org string `json:"org,omitempty"`

	// URI of resource.
	Owners string `json:"owners,omitempty"`

	// URI of resource.
	Self string `json:"self,omitempty"`

	// URI of resource.
	Write string `json:"write,omitempty"`
}

// Label defines model for Label.
type Label struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	OrgID string `json:"orgID,omitempty"`

	// Key/Value pairs associated with this label. Keys can be removed by sending an update with an empty value.
	Properties LabelProperties `json:"properties,omitempty"`
}

// LabelProperties are Key/Value pairs associated with this label. Keys can be
// removed by sending an update with an empty value.
type LabelProperties struct {
	AdditionalProperties map[string]string `json:"additionalProperties,omitempty"`
}

// A BucketSpec defines the desired state of a Bucket.
type BucketSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       BucketParameters `json:"forProvider"`
}

// A BucketStatus represents the observed state of a Bucket.
type BucketStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          BucketObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// An Bucket represents a bucket in InfluxDB.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,influxdb}
type Bucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BucketSpec   `json:"spec"`
	Status BucketStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BucketList contains a list of Bucket.
type BucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Bucket `json:"items"`
}

// Bucket type metadata.
var (
	BucketKind             = reflect.TypeOf(Bucket{}).Name()
	BucketGroupKind        = schema.GroupKind{Group: Group, Kind: BucketKind}.String()
	BucketKindAPIVersion   = BucketKind + "." + SchemeGroupVersion.String()
	BucketGroupVersionKind = SchemeGroupVersion.WithKind(BucketKind)
)

func init() {
	SchemeBuilder.Register(&Bucket{}, &BucketList{})
}
