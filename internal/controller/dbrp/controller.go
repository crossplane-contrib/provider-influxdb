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
	errNotDatabaseRetentionPolicyMapping    = "managed resource is not an DatabaseRetentionPolicyMapping custom resource"
	errGetDatabaseRetentionPolicyMapping    = "cannot get dbrp"
	errCreateDatabaseRetentionPolicyMapping = "cannot create dbrp"
	errUpdateDatabaseRetentionPolicyMapping = "cannot update dbrp"
	errDeleteDatabaseRetentionPolicyMapping = "cannot delete dbrp"
)

// Setup adds a controller that reconciles DatabaseRetentionPolicyMapping managed resources.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.DatabaseRetentionPolicyMappingGroupKind)

	o := controller.Options{
		RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.DatabaseRetentionPolicyMappingGroupVersionKind),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithInitializers(),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1alpha1.DatabaseRetentionPolicyMapping{}).
		Complete(r)
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube client.Client
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cl, err := clients.NewClientWithResponses(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create a new client")
	}
	return &external{api: cl}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	// A 'client' used to connect to the external resource API. In practice this
	// would be something like an AWS SDK client.
	api clients.DBRPsAPI
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.DatabaseRetentionPolicyMapping)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotDatabaseRetentionPolicyMapping)
	}
	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	dbrps, err := c.api.GetDBRPsWithResponse(ctx, &domain.GetDBRPsParams{
		Org: &cr.Spec.ForProvider.Org,
		Id:  pointer.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetDatabaseRetentionPolicyMapping)
	}
	if dbrps.JSON200.Content != nil && len(*dbrps.JSON200.Content) == 0 {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	dbrp := (*dbrps.JSON200.Content)[0]
	if dbrp.Links != nil {
		cr.Status.AtProvider.Links.Self = string(dbrp.Links.Self)
	}
	cr.SetConditions(v1.Available())
	li := resource.NewLateInitializer()
	cr.Spec.ForProvider.Default = li.LateInitializeBoolPtr(cr.Spec.ForProvider.Default, &dbrp.Default)
	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceLateInitialized: li.IsChanged(),
		ResourceUpToDate: pointer.BoolDeref(cr.Spec.ForProvider.Default, false) == dbrp.Default &&
			cr.Spec.ForProvider.RetentionPolicy == dbrp.RetentionPolicy,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.DatabaseRetentionPolicyMapping)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotDatabaseRetentionPolicyMapping)
	}

	resp, err := c.api.PostDBRPWithResponse(ctx, &domain.PostDBRPParams{}, domain.PostDBRPJSONRequestBody{
		BucketID:        cr.Spec.ForProvider.BucketID,
		Database:        cr.Spec.ForProvider.Database,
		Default:         cr.Spec.ForProvider.Default,
		Org:             &cr.Spec.ForProvider.Org,
		RetentionPolicy: cr.Spec.ForProvider.RetentionPolicy,
	})
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateDatabaseRetentionPolicyMapping)
	}
	meta.SetExternalName(cr, resp.JSON201.Id)
	return managed.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.DatabaseRetentionPolicyMapping)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotDatabaseRetentionPolicyMapping)
	}
	_, err := c.api.PatchDBRPIDWithResponse(ctx,
		meta.GetExternalName(cr),
		&domain.PatchDBRPIDParams{},
		domain.PatchDBRPIDJSONRequestBody{
			Default:         cr.Spec.ForProvider.Default,
			RetentionPolicy: pointer.String(cr.Spec.ForProvider.RetentionPolicy),
		})
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateDatabaseRetentionPolicyMapping)
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.DatabaseRetentionPolicyMapping)
	if !ok {
		return errors.New(errNotDatabaseRetentionPolicyMapping)
	}
	_, err := c.api.DeleteDBRPIDWithResponse(ctx, meta.GetExternalName(cr), &domain.DeleteDBRPIDParams{
		Org: pointer.String(cr.Spec.ForProvider.Org),
	})
	return errors.Wrap(err, errDeleteDatabaseRetentionPolicyMapping)
}
