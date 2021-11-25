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

package organization

import (
	"context"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/crossplane-contrib/provider-influxdb/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-influxdb/internal/clients"
)

const (
	errNotOrganization    = "managed resource is not an Organization custom resource"
	errFindOrganization   = "cannot find organization"
	errCreateOrganization = "cannot create organization"
	errUpdateOrganization = "cannot update organization"
	errDeleteOrganization = "cannot delete organization"
)

// Setup adds a controller that reconciles Organization managed resources.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.OrganizationGroupKind)

	o := controller.Options{
		RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.OrganizationGroupVersionKind),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1alpha1.Organization{}).
		Complete(r)
}

type connector struct {
	kube client.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cl, err := clients.NewClient(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create a new client")
	}
	return &external{api: cl.OrganizationsAPI()}, nil
}

type external struct {
	api clients.OrganizationsAPI
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Organization)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotOrganization)
	}

	org, err := c.api.FindOrganizationByName(ctx, meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(IsNotFound, err), errFindOrganization)
	}

	cr.Status.AtProvider = GetOrganizationObservation(org)
	switch cr.Status.AtProvider.Status {
	// Empty string also means active.
	case string(domain.OrganizationStatusActive), "":
		cr.SetConditions(v1.Available())
	case string(domain.OrganizationStatusInactive):
		cr.SetConditions(v1.Unavailable())
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: pointer.StringDeref(org.Description, "") == pointer.StringDeref(cr.Spec.ForProvider.Description, ""),
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

	return managed.ExternalCreation{}, errors.Wrap(err, errCreateOrganization)
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

	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateOrganization)
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Organization)
	if !ok {
		return errors.New(errNotOrganization)
	}
	// NOTE(muvaf): The call returns nil error if the org does not exist.
	return errors.Wrap(c.api.DeleteOrganization(ctx, &domain.Organization{Id: &cr.Status.AtProvider.ID}), errDeleteOrganization)
}
