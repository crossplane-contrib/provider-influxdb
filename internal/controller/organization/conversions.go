/*
Copyright 2020 The Crossplane Authors.

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

package organization

import (
	"github.com/influxdata/influxdb-client-go/v2/domain"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/crossplane-contrib/provider-influxdb/apis/orgs/v1alpha1"
)

// GetOrganizationObservation converts an Organization response to an observation.
func GetOrganizationObservation(org *domain.Organization) v1alpha1.OrganizationObservation { // nolint:gocyclo
	o := v1alpha1.OrganizationObservation{
		ID: pointer.StringDeref(org.Id, ""),
	}
	if org.Status != nil {
		o.Status = string(*org.Status)
	}
	if org.CreatedAt != nil {
		o.CreatedAt = metav1.NewTime(*org.CreatedAt)
	}
	if org.UpdatedAt != nil {
		o.UpdatedAt = metav1.NewTime(*org.UpdatedAt)
	}
	if org.Links != nil {
		if org.Links.Labels != nil {
			o.Links.Labels = string(*org.Links.Labels)
		}
		if org.Links.Dashboards != nil {
			o.Links.Dashboards = string(*org.Links.Dashboards)
		}
		if org.Links.Members != nil {
			o.Links.Members = string(*org.Links.Members)
		}
		if org.Links.Buckets != nil {
			o.Links.Buckets = string(*org.Links.Buckets)
		}
		if org.Links.Owners != nil {
			o.Links.Owners = string(*org.Links.Owners)
		}
		if org.Links.Secrets != nil {
			o.Links.Secrets = string(*org.Links.Secrets)
		}
		if org.Links.Self != nil {
			o.Links.Self = string(*org.Links.Self)
		}
		if org.Links.Tasks != nil {
			o.Links.Tasks = string(*org.Links.Tasks)
		}
	}
	return o
}
