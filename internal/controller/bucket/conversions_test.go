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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"k8s.io/utils/pointer"

	"github.com/crossplane-contrib/provider-influxdb/apis/v1alpha1"
)

func TestLateInitialize(t *testing.T) {
	sType := domain.SchemaTypeExplicit
	type args struct {
		params *v1alpha1.BucketParameters
		obs    *domain.Bucket
	}
	type want struct {
		params *v1alpha1.BucketParameters
		res    bool
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"UpToDate": {
			args: args{
				params: &v1alpha1.BucketParameters{
					Description: pointer.String("name"),
				},
				obs: &domain.Bucket{
					Description: pointer.String("name"),
				},
			},
			want: want{
				params: &v1alpha1.BucketParameters{
					Description: pointer.String("name"),
				},
				res: false,
			},
		},
		"LateInitDescription": {
			args: args{
				params: &v1alpha1.BucketParameters{},
				obs: &domain.Bucket{
					Description: pointer.String("LIname"),
				},
			},
			want: want{
				params: &v1alpha1.BucketParameters{
					Description: pointer.String("LIname"),
				},
				res: true,
			},
		},
		"LateInitSchemaType": {
			args: args{
				params: &v1alpha1.BucketParameters{},
				obs: &domain.Bucket{
					Description: pointer.String("name"),
					SchemaType:  &sType,
				},
			},
			want: want{
				params: &v1alpha1.BucketParameters{
					Description: pointer.String("name"),
					SchemaType:  "explicit",
				},
				res: true,
			},
		},
		"LateInitRetentionRules": {
			args: args{
				params: &v1alpha1.BucketParameters{
					Description: pointer.String("name"),
				},
				obs: &domain.Bucket{
					Description:    pointer.String("LIname"),
					RetentionRules: []domain.RetentionRule{{Type: "expire"}},
				},
			},
			want: want{
				params: &v1alpha1.BucketParameters{
					Description:    pointer.String("name"),
					RetentionRules: []v1alpha1.RetentionRule{{Type: "expire"}},
				},
				res: true,
			},
		},
		"LateInitIgnoreRetentionRules": {
			args: args{
				params: &v1alpha1.BucketParameters{
					Description:    pointer.String("name"),
					RetentionRules: []v1alpha1.RetentionRule{{Type: "expire", EverySeconds: 3600}},
				},
				obs: &domain.Bucket{
					Description:    pointer.String("LIname"),
					RetentionRules: []domain.RetentionRule{{Type: "expire"}},
				},
			},
			want: want{
				params: &v1alpha1.BucketParameters{
					Description:    pointer.String("name"),
					RetentionRules: []v1alpha1.RetentionRule{{Type: "expire", EverySeconds: 3600}},
				},
				res: false,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			res := LateInitialize(tc.args.params, tc.args.obs)

			if diff := cmp.Diff(tc.want.res, res); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.params, tc.args.params); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}
		})
	}
}
func TestIsUpToDate(t *testing.T) {
	type args struct {
		params v1alpha1.BucketParameters
		obs    *domain.Bucket
	}
	cases := map[string]struct {
		args args
		want bool
	}{
		"UpToDate": {
			args: args{
				params: v1alpha1.BucketParameters{
					Description: pointer.String("name"),
				},
				obs: &domain.Bucket{
					Description: pointer.String("name"),
				},
			},
			want: true,
		},
		"DescriptionDifferent": {
			args: args{
				params: v1alpha1.BucketParameters{
					Description: pointer.String("name"),
				},
				obs: &domain.Bucket{
					Description: pointer.String("differentName"),
				},
			},
			want: false,
		},
		"DefaultExpire": {
			args: args{
				params: v1alpha1.BucketParameters{
					Description: pointer.String("name"),
				},
				obs: &domain.Bucket{
					Description:    pointer.String("name"),
					RetentionRules: []domain.RetentionRule{{Type: "expire"}},
				},
			},
			want: true,
		},
		"ParamSetToUnexpire": {
			args: args{
				params: v1alpha1.BucketParameters{
					Description:    pointer.String("name"),
					RetentionRules: []v1alpha1.RetentionRule{{Type: "expire"}},
				},
				obs: &domain.Bucket{
					Description: pointer.String("name"),
				},
			},
			want: true,
		},
		"ParamNonDefault": {
			args: args{
				params: v1alpha1.BucketParameters{
					Description:    pointer.String("name"),
					RetentionRules: []v1alpha1.RetentionRule{{Type: "expire", EverySeconds: 3600}},
				},
				obs: &domain.Bucket{
					Description: pointer.String("name"),
				},
			},
			want: false,
		},
		"ShardGroupMatch": {
			args: args{
				params: v1alpha1.BucketParameters{
					Description:    pointer.String("name"),
					RetentionRules: []v1alpha1.RetentionRule{{Type: "expire", ShardGroupDurationSeconds: pointer.Int64(3600)}},
				},
				obs: &domain.Bucket{
					Description:    pointer.String("name"),
					RetentionRules: []domain.RetentionRule{{Type: "expire", ShardGroupDurationSeconds: pointer.Int64(3600)}},
				},
			},
			want: true,
		},
		"ShardGroupUnequal": {
			args: args{
				params: v1alpha1.BucketParameters{
					Description:    pointer.String("name"),
					RetentionRules: []v1alpha1.RetentionRule{{Type: "expire", ShardGroupDurationSeconds: pointer.Int64(3600)}},
				},
				obs: &domain.Bucket{
					Description:    pointer.String("name"),
					RetentionRules: []domain.RetentionRule{{Type: "expire", ShardGroupDurationSeconds: pointer.Int64(7200)}},
				},
			},
			want: false,
		},
		"RuleLengthCheck": {
			args: args{
				params: v1alpha1.BucketParameters{
					Description:    pointer.String("name"),
					RetentionRules: []v1alpha1.RetentionRule{{Type: "expire"}},
				},
				obs: &domain.Bucket{
					Description: pointer.String("name"),
					RetentionRules: []domain.RetentionRule{
						{Type: "expire"},
						{Type: "expire", EverySeconds: 3600},
					},
				},
			},
			want: false,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			res := IsUpToDate(tc.args.params, tc.args.obs)

			if diff := cmp.Diff(tc.want, res); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}
		})
	}
}
