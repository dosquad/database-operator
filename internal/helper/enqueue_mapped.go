/*
Copyright 2018 The Kubernetes Authors.

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

package helper

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/oklog/ulid/v2"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type empty struct{}

// MapFunc is the signature required for enqueueing requests from a generic function.
// This type is usually used with EnqueueRequestsFromMapFunc when registering an event handler.
type MapFunc func(context.Context, client.Object) []reconcile.Request

// EnqueueRequestsFromMapFunc enqueues Requests by running a transformation function that outputs a collection
// of reconcile.Requests on each Event.  The reconcile.Requests may be for an arbitrary set of objects
// defined by some user specified transformation of the source Event.  (e.g. trigger Reconciler for a set of objects
// in response to a cluster resize event caused by adding or deleting a Node)
//
// EnqueueRequestsFromMapFunc is frequently used to fan-out updates from one object to one or more other
// objects of a differing type.
//
// For UpdateEvents which contain both a new and old object, the transformation function is run on both
// objects and both sets of Requests are enqueue.
func EnqueueRequestsFromMapFunc(fn MapFunc) handler.EventHandler {
	return &enqueueRequestsFromMapFunc{
		toRequests: fn,
	}
}

var _ handler.EventHandler = &enqueueRequestsFromMapFunc{}

type enqueueRequestsFromMapFunc struct {
	// Mapper transforms the argument into a slice of keys to be reconciled
	toRequests MapFunc
}

// Create implements EventHandler.
func (e *enqueueRequestsFromMapFunc) Create(
	ctx context.Context,
	evt event.TypedCreateEvent[client.Object],
	q workqueue.TypedRateLimitingInterface[reconcile.Request],
) {
	ctx, logger := e.loggerDetail(ctx, evt.Object)

	reqs := map[reconcile.Request]empty{}
	logger.V(1).Info("EventHandler():Create()")
	e.mapAndEnqueue(ctx, q, evt.Object, reqs)
}

// Update implements EventHandler.
func (e *enqueueRequestsFromMapFunc) Update(
	ctx context.Context,
	evt event.TypedUpdateEvent[client.Object],
	q workqueue.TypedRateLimitingInterface[reconcile.Request],
) {
	ctx, logger := e.loggerDetail(ctx, evt.ObjectNew)

	reqs := map[reconcile.Request]empty{}
	// logger.V(1).Info("EventHandler():Update():Old Object")
	// e.mapAndEnqueue(ctx, q, evt.ObjectOld, reqs)
	logger.V(1).Info("EventHandler():Update():New Object")
	e.mapAndEnqueue(ctx, q, evt.ObjectNew, reqs)
}

// Delete implements EventHandler.
func (e *enqueueRequestsFromMapFunc) Delete(
	ctx context.Context,
	evt event.TypedDeleteEvent[client.Object],
	q workqueue.TypedRateLimitingInterface[reconcile.Request],
) {
	ctx, logger := e.loggerDetail(ctx, evt.Object)

	reqs := map[reconcile.Request]empty{}
	logger.V(1).Info("EventHandler():Delete()")
	e.mapAndEnqueue(ctx, q, evt.Object, reqs)
}

// Generic implements EventHandler.
func (e *enqueueRequestsFromMapFunc) Generic(
	ctx context.Context,
	evt event.TypedGenericEvent[client.Object],
	q workqueue.TypedRateLimitingInterface[reconcile.Request],
) {
	ctx, logger := e.loggerDetail(ctx, evt.Object)

	reqs := map[reconcile.Request]empty{}
	logger.V(1).Info("EventHandler():Generic()")
	e.mapAndEnqueue(ctx, q, evt.Object, reqs)
}

func (e *enqueueRequestsFromMapFunc) loggerDetail(
	ctx context.Context,
	obj client.Object,
) (context.Context, logr.Logger) {
	logger := log.FromContext(ctx)
	callID := ulid.Make().String()
	logger = logger.WithValues(
		"enqueueID", callID,
		"enqueueName", obj.GetName(),
		"enqueueNamespace", obj.GetNamespace(),
	)
	ctx = log.IntoContext(ctx, logger)
	return ctx, logger
}

func (e *enqueueRequestsFromMapFunc) mapAndEnqueue(
	ctx context.Context,
	q workqueue.TypedRateLimitingInterface[reconcile.Request],
	object client.Object,
	reqs map[reconcile.Request]empty,
) {
	for _, req := range e.toRequests(ctx, object) {
		_, ok := reqs[req]
		if !ok {
			q.Add(req)
			reqs[req] = empty{}
		}
	}
}
