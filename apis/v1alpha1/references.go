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
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

// OrganizationID extracts ID of organization from Organization resource.
func OrganizationID() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		if cr, ok := mg.(*Organization); ok {
			return cr.Status.AtProvider.ID
		}
		return ""
	}
}

// BucketID extracts ID of organization from Bucket resource.
func BucketID() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		if cr, ok := mg.(*Bucket); ok {
			return cr.Status.AtProvider.ID
		}
		return ""
	}
}
