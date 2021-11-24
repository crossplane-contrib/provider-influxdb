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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/crossplane-contrib/provider-influxdb/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-influxdb/internal/clients"
)

const (
	errNotBucket    = "managed resource is not an Bucket custom resource"
	errFindBucket   = "cannot find bucket"
	errCreateBucket = "cannot create bucket"
	errUpdateBucket = "cannot update bucket"
	errDeleteBucket = "cannot delete bucket"
)

// Setup adds a controller that reconciles Bucket managed resources.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.BucketGroupKind)

	o := controller.Options{
		RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.BucketGroupVersionKind),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1alpha1.Bucket{}).
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
	cl, err := clients.NewClient(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create a new client")
	}
	return &external{api: cl.BucketsAPI()}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	// A 'client' used to connect to the external resource API. In practice this
	// would be something like an AWS SDK client.
	api clients.BucketsAPI
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotBucket)
	}

	bucket, err := c.api.FindBucketByName(ctx, meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(IsNotFound(meta.GetExternalName(cr)), err), errFindBucket)
	}

	cr.Status.AtProvider = GenerateBucketObservation(bucket)
	cr.SetConditions(v1.Available())
	li := LateInitialize(&cr.Spec.ForProvider, bucket)
	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceLateInitialized: li,
		ResourceUpToDate:        IsUpToDate(cr.Spec.ForProvider, bucket),
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotBucket)
	}

	_, err := c.api.CreateBucket(ctx, GenerateBucket(meta.GetExternalName(cr), cr.Spec.ForProvider))

	return managed.ExternalCreation{}, errors.Wrap(err, errCreateBucket)
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotBucket)
	}
	b := GenerateBucket(meta.GetExternalName(cr), cr.Spec.ForProvider)
	b.Id = &cr.Status.AtProvider.ID
	_, err := c.api.UpdateBucket(ctx, b)

	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateBucket)
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return errors.New(errNotBucket)
	}
	// NOTE(muvaf): The call returns nil error if the org does not exist.
	return errors.Wrap(c.api.DeleteBucket(ctx, &domain.Bucket{Id: &cr.Status.AtProvider.ID}), errDeleteBucket)
}
