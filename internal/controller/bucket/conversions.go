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

package bucket

import (
	"fmt"
	"sort"
	"strings"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/crossplane-contrib/provider-influxdb/apis/v1alpha1"
)

// GenerateBucketObservation converts an Bucket response to an observation.
func GenerateBucketObservation(b *domain.Bucket) v1alpha1.BucketObservation { // nolint:gocyclo
	o := v1alpha1.BucketObservation{
		ID: pointer.StringDeref(b.Id, ""),
	}
	if b.Type != nil {
		o.Type = string(*b.Type)
	}
	if b.CreatedAt != nil {
		o.CreatedAt = metav1.NewTime(*b.CreatedAt)
	}
	if b.UpdatedAt != nil {
		o.UpdatedAt = metav1.NewTime(*b.UpdatedAt)
	}
	if b.Links != nil {
		if b.Links.Labels != nil {
			o.Links.Labels = string(*b.Links.Labels)
		}
		if b.Links.Members != nil {
			o.Links.Members = string(*b.Links.Members)
		}
		if b.Links.Org != nil {
			o.Links.Org = string(*b.Links.Org)
		}
		if b.Links.Owners != nil {
			o.Links.Owners = string(*b.Links.Owners)
		}
		if b.Links.Self != nil {
			o.Links.Self = string(*b.Links.Self)
		}
		if b.Links.Write != nil {
			o.Links.Write = string(*b.Links.Write)
		}
	}
	if b.Labels != nil && len(*b.Labels) != 0 {
		o.Labels = make([]v1alpha1.Label, len(*b.Labels))
		for i, l := range *b.Labels {
			o.Labels[i] = v1alpha1.Label{
				ID:    pointer.StringDeref(l.Id, ""),
				Name:  pointer.StringDeref(l.Name, ""),
				OrgID: pointer.StringDeref(l.OrgID, ""),
			}
			if l.Properties != nil && len(l.Properties.AdditionalProperties) != 0 {
				o.Labels[i].Properties.AdditionalProperties = make(map[string]string, len(l.Properties.AdditionalProperties))
				for k, v := range l.Properties.AdditionalProperties {
					o.Labels[i].Properties.AdditionalProperties[k] = v
				}
			}
		}
		sort.SliceStable(o.Labels, func(i, j int) bool {
			return o.Labels[i].ID < o.Labels[j].ID
		})
	}
	return o
}

// GenerateBucket returns a Bucket model that the InfluxDB API accepts for creation
// and update.
func GenerateBucket(name string, params v1alpha1.BucketParameters) *domain.Bucket {
	sType := domain.SchemaType(params.SchemaType)
	out := &domain.Bucket{
		Name:        name,
		Description: params.Description,
		OrgID:       params.OrgID,
		Rp:          params.RP,
		SchemaType:  &sType,
	}
	if len(params.RetentionRules) != 0 {
		out.RetentionRules = make([]domain.RetentionRule, len(params.RetentionRules))
		for i, rr := range params.RetentionRules {
			out.RetentionRules[i] = domain.RetentionRule{
				EverySeconds:              rr.EverySeconds,
				ShardGroupDurationSeconds: rr.ShardGroupDurationSeconds,
				Type:                      domain.RetentionRuleType(rr.Type),
			}
		}
	}
	return out
}

// LateInitialize sets the defaults from the API if user didn't set a value for
// such fields.
func LateInitialize(params *v1alpha1.BucketParameters, obs *domain.Bucket) bool {
	li := resource.NewLateInitializer()
	params.Description = li.LateInitializeStringPtr(params.Description, obs.Description)
	params.RP = li.LateInitializeStringPtr(params.RP, obs.Rp)
	if params.SchemaType == "" && obs.SchemaType != nil && string(*obs.SchemaType) != "" {
		params.SchemaType = string(*obs.SchemaType)
		return true
	}
	sort.SliceStable(params.RetentionRules, func(i, j int) bool {
		return params.RetentionRules[i].Type < params.RetentionRules[j].Type
	})
	sort.SliceStable(obs.RetentionRules, func(i, j int) bool {
		return string(obs.RetentionRules[i].Type) < string(obs.RetentionRules[j].Type)
	})
	for i := range params.RetentionRules {
		if params.RetentionRules[i].Type == string(obs.RetentionRules[i].Type) {
			params.RetentionRules[i].ShardGroupDurationSeconds = li.LateInitializeInt64Ptr(params.RetentionRules[i].ShardGroupDurationSeconds, obs.RetentionRules[i].ShardGroupDurationSeconds)
		}
	}
	return li.IsChanged()
}

// IsUpToDate returns whether an update call is necessary.
func IsUpToDate(params v1alpha1.BucketParameters, obs *domain.Bucket) bool {
	if len(params.RetentionRules) != len(obs.RetentionRules) {
		return false
	}
	sort.SliceStable(params.RetentionRules, func(i, j int) bool {
		return params.RetentionRules[i].Type < params.RetentionRules[j].Type
	})
	sort.SliceStable(obs.RetentionRules, func(i, j int) bool {
		return string(obs.RetentionRules[i].Type) < string(obs.RetentionRules[j].Type)
	})
	for i := range params.RetentionRules {
		if params.RetentionRules[i].Type != string(obs.RetentionRules[i].Type) ||
			params.RetentionRules[i].EverySeconds != obs.RetentionRules[i].EverySeconds ||
			pointer.Int64Deref(params.RetentionRules[i].ShardGroupDurationSeconds, 0) != pointer.Int64Deref(obs.RetentionRules[i].ShardGroupDurationSeconds, 0) {
			return false
		}
	}
	return pointer.StringDeref(obs.Description, "") == pointer.StringDeref(params.Description, "")
}

// IsNotFoundFn returns an ErrorIs function that can tell whether the error is
// of kind NotFound.
func IsNotFoundFn(name string) resource.ErrorIs {
	return func(err error) bool {
		return strings.Contains(err.Error(), fmt.Sprintf("bucket '%s' not found", name))
	}
}
