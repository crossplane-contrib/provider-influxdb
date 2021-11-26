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

package dbrp

import (
	"context"
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/resource/fake"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"github.com/pkg/errors"

	"github.com/crossplane-contrib/provider-influxdb/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-influxdb/internal/clients"
)

var (
	errBoom = errors.New("boom")
)

func TestObserve(t *testing.T) {
	type args struct {
		mg  resource.Managed
		api clients.DBRPsAPI
	}
	type want struct {
		err error
		obs managed.ExternalObservation
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"NotDatabaseRetentionPolicyMapping": {
			args: args{
				mg: &fake.Managed{},
			},
			want: want{
				err: errors.New(errNotDatabaseRetentionPolicyMapping),
			},
		},
		"FindFailed": {
			args: args{
				mg: &v1alpha1.DatabaseRetentionPolicyMapping{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: "test",
						},
					},
				},
				api: &clients.MockDBRPsAPI{
					GetDBRPsWithResponseFn: func(_ context.Context, _ *domain.GetDBRPsParams) (*domain.GetDBRPsResponse, error) {
						return nil, errBoom
					},
				},
			},
			want: want{
				err: errors.Wrap(errBoom, errGetDatabaseRetentionPolicyMapping),
			},
		},
		"NotFoundCreationNeeded": {
			args: args{
				mg: &v1alpha1.DatabaseRetentionPolicyMapping{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: "test",
						},
					},
				},
				api: &clients.MockDBRPsAPI{
					GetDBRPsWithResponseFn: func(_ context.Context, _ *domain.GetDBRPsParams) (*domain.GetDBRPsResponse, error) {
						return &domain.GetDBRPsResponse{JSON200: &domain.DBRPs{Content: &[]domain.DBRP{}}}, nil
					},
				},
			},
			want: want{
				obs: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"NoIDCreationNeeded": {
			args: args{
				mg: &v1alpha1.DatabaseRetentionPolicyMapping{},
			},
			want: want{
				obs: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"UpdateNeeded": {
			args: args{
				mg: &v1alpha1.DatabaseRetentionPolicyMapping{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: "test",
						},
					},
					Spec: v1alpha1.DatabaseRetentionPolicyMappingSpec{
						ForProvider: v1alpha1.DatabaseRetentionPolicyMappingParameters{
							RetentionPolicy: "desired",
						},
					},
				},
				api: &clients.MockDBRPsAPI{
					GetDBRPsWithResponseFn: func(_ context.Context, _ *domain.GetDBRPsParams) (*domain.GetDBRPsResponse, error) {
						return &domain.GetDBRPsResponse{JSON200: &domain.DBRPs{Content: &[]domain.DBRP{
							{
								RetentionPolicy: "observed",
							},
						}}}, nil
					},
				},
			},
			want: want{
				obs: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: true,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			obs, err := (&external{api: tc.args.api}).Observe(context.TODO(), tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.obs, obs); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		mg  resource.Managed
		api clients.DBRPsAPI
	}
	type want struct {
		err error
		cre managed.ExternalCreation
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"NotDatabaseRetentionPolicyMapping": {
			args: args{
				mg: &fake.Managed{},
			},
			want: want{
				err: errors.New(errNotDatabaseRetentionPolicyMapping),
			},
		},
		"CreateFailed": {
			args: args{
				mg: &v1alpha1.DatabaseRetentionPolicyMapping{},
				api: &clients.MockDBRPsAPI{
					PostDBRPWithResponseFn: func(_ context.Context, _ *domain.PostDBRPParams, _ domain.PostDBRPJSONRequestBody) (*domain.PostDBRPResponse, error) {
						return nil, errBoom
					},
				},
			},
			want: want{
				err: errors.Wrap(errBoom, errCreateDatabaseRetentionPolicyMapping),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cre, err := (&external{api: tc.args.api}).Create(context.TODO(), tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Create(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cre, cre); diff != "" {
				t.Errorf("Create(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type args struct {
		mg  resource.Managed
		api clients.DBRPsAPI
	}
	type want struct {
		err error
		obs managed.ExternalUpdate
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"NotDatabaseRetentionPolicyMapping": {
			args: args{
				mg: &fake.Managed{},
			},
			want: want{
				err: errors.New(errNotDatabaseRetentionPolicyMapping),
			},
		},
		"UpdateFailed": {
			args: args{
				mg: &v1alpha1.DatabaseRetentionPolicyMapping{},
				api: &clients.MockDBRPsAPI{
					PatchDBRPIDWithResponseFn: func(_ context.Context, _ string, _ *domain.PatchDBRPIDParams, _ domain.PatchDBRPIDJSONRequestBody) (*domain.PatchDBRPIDResponse, error) {
						return nil, errBoom
					},
				},
			},
			want: want{
				err: errors.Wrap(errBoom, errUpdateDatabaseRetentionPolicyMapping),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			obs, err := (&external{api: tc.args.api}).Update(context.TODO(), tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Update(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.obs, obs); diff != "" {
				t.Errorf("Update(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		mg  resource.Managed
		api clients.DBRPsAPI
	}
	type want struct {
		err error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"NotDatabaseRetentionPolicyMapping": {
			args: args{
				mg: &fake.Managed{},
			},
			want: want{
				err: errors.New(errNotDatabaseRetentionPolicyMapping),
			},
		},
		"DeleteWithCorrectName": {
			args: args{
				mg: &v1alpha1.DatabaseRetentionPolicyMapping{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: "testid",
						},
					},
				},
				api: &clients.MockDBRPsAPI{
					DeleteDBRPIDWithResponseFn: func(_ context.Context, dbrpID string, _ *domain.DeleteDBRPIDParams) (*domain.DeleteDBRPIDResponse, error) {
						if dbrpID != "testid" {
							t.Errorf("deletion call has to use the id in external name for deletion")
						}
						return nil, nil
					},
				},
			},
		},
		"DeleteFailed": {
			args: args{
				mg: &v1alpha1.DatabaseRetentionPolicyMapping{},
				api: &clients.MockDBRPsAPI{
					DeleteDBRPIDWithResponseFn: func(_ context.Context, _ string, _ *domain.DeleteDBRPIDParams) (*domain.DeleteDBRPIDResponse, error) {
						return nil, errBoom
					},
				},
			},
			want: want{
				err: errors.Wrap(errBoom, errDeleteDatabaseRetentionPolicyMapping),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := (&external{api: tc.args.api}).Delete(context.TODO(), tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Delete(...): -want, +got:\n%s", diff)
			}
		})
	}
}
