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

// DBRPsAPI is the set of calls we make in controllers that use DBRPs
// API.
type DBRPsAPI interface {
	PostDBRPWithResponse(ctx context.Context, params *domain.PostDBRPParams, body domain.PostDBRPJSONRequestBody) (*domain.PostDBRPResponse, error)
	GetDBRPsWithResponse(ctx context.Context, params *domain.GetDBRPsParams) (*domain.GetDBRPsResponse, error)
	PatchDBRPIDWithResponse(ctx context.Context, dbrpID string, params *domain.PatchDBRPIDParams, body domain.PatchDBRPIDJSONRequestBody) (*domain.PatchDBRPIDResponse, error)
	DeleteDBRPIDWithResponse(ctx context.Context, dbrpID string, params *domain.DeleteDBRPIDParams) (*domain.DeleteDBRPIDResponse, error)
}

// MockDBRPsAPI mocks DBRPsAPI.
type MockDBRPsAPI struct {
	PostDBRPWithResponseFn     func(ctx context.Context, params *domain.PostDBRPParams, body domain.PostDBRPJSONRequestBody) (*domain.PostDBRPResponse, error)
	GetDBRPsWithResponseFn     func(ctx context.Context, params *domain.GetDBRPsParams) (*domain.GetDBRPsResponse, error)
	PatchDBRPIDWithResponseFn  func(ctx context.Context, dbrpID string, params *domain.PatchDBRPIDParams, body domain.PatchDBRPIDJSONRequestBody) (*domain.PatchDBRPIDResponse, error)
	DeleteDBRPIDWithResponseFn func(ctx context.Context, dbrpID string, params *domain.DeleteDBRPIDParams) (*domain.DeleteDBRPIDResponse, error)
}

// PostDBRPWithResponse calls PostDBRPWithResponseFn.
func (m *MockDBRPsAPI) PostDBRPWithResponse(ctx context.Context, params *domain.PostDBRPParams, body domain.PostDBRPJSONRequestBody) (*domain.PostDBRPResponse, error) {
	return m.PostDBRPWithResponseFn(ctx, params, body)
}

// GetDBRPsWithResponse calls GetDBRPsWithResponseFn.
func (m *MockDBRPsAPI) GetDBRPsWithResponse(ctx context.Context, params *domain.GetDBRPsParams) (*domain.GetDBRPsResponse, error) {
	return m.GetDBRPsWithResponseFn(ctx, params)
}

// PatchDBRPIDWithResponse calls PatchDBRPIDWithResponseFn.
func (m *MockDBRPsAPI) PatchDBRPIDWithResponse(ctx context.Context, dbrpID string, params *domain.PatchDBRPIDParams, body domain.PatchDBRPIDJSONRequestBody) (*domain.PatchDBRPIDResponse, error) {
	return m.PatchDBRPIDWithResponseFn(ctx, dbrpID, params, body)
}

// DeleteDBRPIDWithResponse calls DeleteDBRPIDWithResponseFn..
func (m *MockDBRPsAPI) DeleteDBRPIDWithResponse(ctx context.Context, dbrpID string, params *domain.DeleteDBRPIDParams) (*domain.DeleteDBRPIDResponse, error) {
	return m.DeleteDBRPIDWithResponseFn(ctx, dbrpID, params)
}
