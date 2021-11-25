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
	"net/http"
	"strings"

	"github.com/influxdata/influxdb-client-go/v2/domain"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
	influxdbv2 "github.com/influxdata/influxdb-client-go/v2"
	apihttp "github.com/influxdata/influxdb-client-go/v2/api/http"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-influxdb/apis/v1alpha1"
)

const (
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get referenced ProviderConfig"
	errGetCreds     = "cannot get credentials"
)

// NewClient returns the base InfluxDB client.
func NewClient(ctx context.Context, kube client.Client, mg resource.Managed) (influxdbv2.Client, error) {
	pc := &v1alpha1.ProviderConfig{}
	if err := kube.Get(ctx, types.NamespacedName{Name: mg.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPC)
	}

	if err := resource.NewProviderConfigUsageTracker(kube, &v1alpha1.ProviderConfigUsage{}).Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	cd := pc.Spec.Credentials
	token, err := resource.CommonCredentialExtractor(ctx, cd.Source, kube, cd.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}
	return influxdbv2.NewClient(pc.Spec.Endpoint, string(token)), nil
}

// NewClientWithResponses returns the bare client. Use this only if NewClient
// does not meet your needs.
func NewClientWithResponses(ctx context.Context, kube client.Client, mg resource.Managed) (*domain.ClientWithResponses, error) {
	pc := &v1alpha1.ProviderConfig{}
	if err := kube.Get(ctx, types.NamespacedName{Name: mg.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPC)
	}

	if err := resource.NewProviderConfigUsageTracker(kube, &v1alpha1.ProviderConfigUsage{}).Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	cd := pc.Spec.Credentials
	token, err := resource.CommonCredentialExtractor(ctx, cd.Source, kube, cd.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}
	normServerURL := pc.Spec.Endpoint
	if !strings.HasSuffix(normServerURL, "/") {
		// For subsequent path parts concatenation, url has to end with '/'
		normServerURL = pc.Spec.Endpoint + "/"
	}
	authorization := ""
	if len(string(token)) > 0 {
		authorization = "Token " + string(token)
	}
	service := apihttp.NewService(normServerURL, authorization, apihttp.DefaultOptions())
	return domain.NewClientWithResponses(service), nil
}

// IsNotFound returns whether the error is of type NotFound.
func IsNotFound(err error) bool {
	hErr, ok := err.(*apihttp.Error)
	return ok && hErr.StatusCode == http.StatusNotFound
}
