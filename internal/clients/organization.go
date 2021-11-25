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

// OrganizationsAPI is the set of calls we make in controllers that use Organizations
// API.
type OrganizationsAPI interface {
	// CreateOrganization creates new organization.
	CreateOrganization(ctx context.Context, org *domain.Organization) (*domain.Organization, error)

	// FindOrganizationByName returns an organization found using orgName.
	FindOrganizationByName(ctx context.Context, orgName string) (*domain.Organization, error)

	// UpdateOrganization updates organization.
	UpdateOrganization(ctx context.Context, org *domain.Organization) (*domain.Organization, error)

	// DeleteOrganization deletes an organization.
	DeleteOrganization(ctx context.Context, org *domain.Organization) error
}

// MockOrganizationsAPI mocks OrganizationsAPI.
type MockOrganizationsAPI struct {
	CreateOrganizationFn     func(ctx context.Context, org *domain.Organization) (*domain.Organization, error)
	FindOrganizationByNameFn func(ctx context.Context, orgName string) (*domain.Organization, error)
	UpdateOrganizationFn     func(ctx context.Context, org *domain.Organization) (*domain.Organization, error)
	DeleteOrganizationFn     func(ctx context.Context, org *domain.Organization) error
}

// CreateOrganization calls CreateOrganizationFn.
func (m *MockOrganizationsAPI) CreateOrganization(ctx context.Context, org *domain.Organization) (*domain.Organization, error) {
	return m.CreateOrganizationFn(ctx, org)
}

// FindOrganizationByName calls FindOrganizationByNameFn.
func (m *MockOrganizationsAPI) FindOrganizationByName(ctx context.Context, orgName string) (*domain.Organization, error) {
	return m.FindOrganizationByNameFn(ctx, orgName)
}

// UpdateOrganization calls UpdateOrganizationFn.
func (m *MockOrganizationsAPI) UpdateOrganization(ctx context.Context, org *domain.Organization) (*domain.Organization, error) {
	return m.UpdateOrganizationFn(ctx, org)
}

// DeleteOrganization calls DeleteOrganizationFn.
func (m *MockOrganizationsAPI) DeleteOrganization(ctx context.Context, org *domain.Organization) error {
	return m.DeleteOrganizationFn(ctx, org)
}
