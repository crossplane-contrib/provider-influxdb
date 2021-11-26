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
	"context"
	"fmt"
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/resource/fake"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"github.com/pkg/errors"
	"k8s.io/utils/pointer"

	"github.com/crossplane-contrib/provider-influxdb/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-influxdb/internal/clients"
)

var (
	errBoom = errors.New("boom")
)

func TestObserve(t *testing.T) {
	type args struct {
		mg  resource.Managed
		api clients.BucketsAPI
	}
	type want struct {
		err error
		obs managed.ExternalObservation
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"NotBucket": {
			args: args{
				mg: &fake.Managed{},
			},
			want: want{
				err: errors.New(errNotBucket),
			},
		},
		"FindFailed": {
			args: args{
				mg: &v1alpha1.Bucket{},
				api: &clients.MockBucketsAPI{
					FindBucketByNameFn: func(_ context.Context, _ string) (*domain.Bucket, error) {
						return nil, errBoom
					},
				},
			},
			want: want{
				err: errors.Wrap(errBoom, errFindBucket),
			},
		},
		"NotFoundCreationNeeded": {
			args: args{
				mg: &v1alpha1.Bucket{},
				api: &clients.MockBucketsAPI{
					FindBucketByNameFn: func(_ context.Context, name string) (*domain.Bucket, error) {
						return nil, fmt.Errorf("bucket '%s' not found", name)
					},
				},
			},
			want: want{
				obs: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"UpdateNeeded": {
			args: args{
				mg: &v1alpha1.Bucket{
					Spec: v1alpha1.BucketSpec{
						ForProvider: v1alpha1.BucketParameters{
							Description: pointer.String("desired"),
						},
					},
				},
				api: &clients.MockBucketsAPI{
					FindBucketByNameFn: func(_ context.Context, _ string) (*domain.Bucket, error) {
						return &domain.Bucket{
							Description: pointer.String("observed"),
						}, nil
					},
				},
			},
			want: want{
				obs: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
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
		api clients.BucketsAPI
	}
	type want struct {
		err error
		cre managed.ExternalCreation
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"NotBucket": {
			args: args{
				mg: &fake.Managed{},
			},
			want: want{
				err: errors.New(errNotBucket),
			},
		},
		"CreateFailed": {
			args: args{
				mg: &v1alpha1.Bucket{},
				api: &clients.MockBucketsAPI{
					CreateBucketFn: func(_ context.Context, _ *domain.Bucket) (*domain.Bucket, error) {
						return nil, errBoom
					},
				},
			},
			want: want{
				err: errors.Wrap(errBoom, errCreateBucket),
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
		api clients.BucketsAPI
	}
	type want struct {
		err error
		obs managed.ExternalUpdate
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"NotBucket": {
			args: args{
				mg: &fake.Managed{},
			},
			want: want{
				err: errors.New(errNotBucket),
			},
		},
		"UpdateFailed": {
			args: args{
				mg: &v1alpha1.Bucket{},
				api: &clients.MockBucketsAPI{
					UpdateBucketFn: func(_ context.Context, _ *domain.Bucket) (*domain.Bucket, error) {
						return nil, errBoom
					},
				},
			},
			want: want{
				err: errors.Wrap(errBoom, errUpdateBucket),
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
		api clients.BucketsAPI
	}
	type want struct {
		err error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"NotBucket": {
			args: args{
				mg: &fake.Managed{},
			},
			want: want{
				err: errors.New(errNotBucket),
			},
		},
		"DeleteWithCorrectName": {
			args: args{
				mg: &v1alpha1.Bucket{
					Status: v1alpha1.BucketStatus{
						AtProvider: v1alpha1.BucketObservation{
							ID: "testid",
						},
					},
				},
				api: &clients.MockBucketsAPI{
					DeleteBucketFn: func(_ context.Context, org *domain.Bucket) error {
						if pointer.StringDeref(org.Id, "") != "testid" {
							t.Errorf("deletion call has to use the id for deletion")
						}
						return nil
					},
				},
			},
		},
		"DeleteFailed": {
			args: args{
				mg: &v1alpha1.Bucket{},
				api: &clients.MockBucketsAPI{
					DeleteBucketFn: func(_ context.Context, _ *domain.Bucket) error {
						return errBoom
					},
				},
			},
			want: want{
				err: errors.Wrap(errBoom, errDeleteBucket),
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
