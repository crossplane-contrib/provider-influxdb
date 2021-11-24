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

package clients

import (
	"context"

	"github.com/influxdata/influxdb-client-go/v2/domain"
)

// BucketsAPI is the set of calls we make in controllers that use Buckets
// API.
type BucketsAPI interface {
	// CreateBucket creates a new bucket.
	CreateBucket(ctx context.Context, bucket *domain.Bucket) (*domain.Bucket, error)

	// FindBucketByName returns a bucket found using bucketName.
	FindBucketByName(ctx context.Context, bucketName string) (*domain.Bucket, error)

	// UpdateBucket updates a bucket.
	UpdateBucket(ctx context.Context, bucket *domain.Bucket) (*domain.Bucket, error)

	// DeleteBucket deletes a bucket.
	DeleteBucket(ctx context.Context, bucket *domain.Bucket) error
}

// MockBucketsAPI mocks BucketsAPI.
type MockBucketsAPI struct {
	CreateBucketFn     func(ctx context.Context, org *domain.Bucket) (*domain.Bucket, error)
	FindBucketByNameFn func(ctx context.Context, orgName string) (*domain.Bucket, error)
	UpdateBucketFn     func(ctx context.Context, org *domain.Bucket) (*domain.Bucket, error)
	DeleteBucketFn     func(ctx context.Context, org *domain.Bucket) error
}

// CreateBucket calls CreateBucketFn.
func (m *MockBucketsAPI) CreateBucket(ctx context.Context, org *domain.Bucket) (*domain.Bucket, error) {
	return m.CreateBucketFn(ctx, org)
}

// FindBucketByName calls FindBucketByNameFn.
func (m *MockBucketsAPI) FindBucketByName(ctx context.Context, orgName string) (*domain.Bucket, error) {
	return m.FindBucketByNameFn(ctx, orgName)
}

// UpdateBucket calls UpdateBucketFn.
func (m *MockBucketsAPI) UpdateBucket(ctx context.Context, org *domain.Bucket) (*domain.Bucket, error) {
	return m.UpdateBucketFn(ctx, org)
}

// DeleteBucket calls DeleteBucketFn.
func (m *MockBucketsAPI) DeleteBucket(ctx context.Context, org *domain.Bucket) error {
	return m.DeleteBucketFn(ctx, org)
}
