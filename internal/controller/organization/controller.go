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
	"context"
	"fmt"
	"strings"

	"k8s.io/utils/pointer"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	"github.com/crossplane/crossplane-runtime/pkg/meta"

	"github.com/influxdata/influxdb-client-go/v2/domain"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"

	"github.com/influxdata/influxdb-client-go/v2/api"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-influxdb/apis/orgs/v1alpha1"
	apisv1alpha1 "github.com/crossplane-contrib/provider-influxdb/apis/v1alpha1"
)

const (
	errNotOrganization = "managed resource is not an Organization custom resource"
	errTrackPCUsage    = "cannot track ProviderConfig usage"
	errGetPC           = "cannot get ProviderConfig"
	errGetCreds        = "cannot get credentials"
)

// Setup adds a controller that reconciles Organization managed resources.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.OrganizationGroupKind)

	o := controller.Options{
		RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.OrganizationGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:  mgr.GetClient(),
			usage: resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
		}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1alpha1.Organization{}).
		Complete(r)
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube  client.Client
	usage resource.Tracker
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Organization)
	if !ok {
		return nil, errors.New(errNotOrganization)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	pc := &apisv1alpha1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPC)
	}

	cd := pc.Spec.Credentials
	token, err := resource.CommonCredentialExtractor(ctx, cd.Source, c.kube, cd.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	return &external{api: influxdb2.NewClient(pc.Spec.Endpoint, string(token)).OrganizationsAPI()}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	// A 'client' used to connect to the external resource API. In practice this
	// would be something like an AWS SDK client.
	api api.OrganizationsAPI
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Organization)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotOrganization)
	}

	org, err := c.api.FindOrganizationByName(ctx, meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(IsNotFound(meta.GetExternalName(cr)), err), "cannot find organization")
	}

	cr.Status.AtProvider = GetOrganizationObservation(org)
	switch cr.Status.AtProvider.Status {
	case "active":
		cr.SetConditions(v1.Available())
	case "inactive":
		cr.SetConditions(v1.Unavailable())
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: pointer.StringDeref(org.Description, "") != pointer.StringDeref(cr.Spec.ForProvider.Description, ""),
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Organization)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotOrganization)
	}

	_, err := c.api.CreateOrganization(ctx, &domain.Organization{
		Description: cr.Spec.ForProvider.Description,
		Name:        meta.GetExternalName(cr),
	})

	return managed.ExternalCreation{}, errors.Wrap(err, "cannot create organization")
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Organization)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotOrganization)
	}

	_, err := c.api.UpdateOrganization(ctx, &domain.Organization{
		Name:        meta.GetExternalName(cr),
		Description: cr.Spec.ForProvider.Description,
		Id:          &cr.Status.AtProvider.ID,
	})

	return managed.ExternalUpdate{}, errors.Wrap(err, "cannot update organization")
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Organization)
	if !ok {
		return errors.New(errNotOrganization)
	}
	// NOTE(muvaf): The call returns nil error if the org does not exist.
	return errors.Wrap(c.api.DeleteOrganization(ctx, &domain.Organization{Id: &cr.Status.AtProvider.ID}), "cannot delete organization")
}

// IsNotFound returns an ErrorIs function specific to the given org name.
func IsNotFound(name string) resource.ErrorIs {
	return func(err error) bool {
		// NOTE(muvaf): There is no other way currently to know whether the error
		// is 404 since the client returns bare fmt.Errorf.
		// See source code of FindOrganizationByName function for more details.
		return strings.Contains(err.Error(), fmt.Sprintf("organization '%s' not found", name))
	}
}
