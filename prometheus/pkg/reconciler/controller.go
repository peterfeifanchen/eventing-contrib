/*
Copyright 2019 The Knative Authors

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

package reconciler

import (
	"context"

	"k8s.io/client-go/tools/cache"
	"knative.dev/eventing-contrib/pkg/apis/sources/v1alpha1"
	eventtypeinformer "knative.dev/eventing/pkg/client/injection/informers/eventing/v1alpha1/eventtype"
	"knative.dev/eventing/pkg/reconciler"
	deploymentinformer "knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/resolver"

	"knative.dev/eventing-contrib/prometheus/pkg/client/injection/client"
	prometheusinformer "knative.dev/eventing-contrib/prometheus/pkg/client/injection/informers/sources/v1alpha1/prometheussource"
)

const (
	// ReconcilerName is the name of the reconciler
	ReconcilerName = "PrometheusSource"

	// controllerAgentName is the string used by this controller to identify
	// itself when creating events.
	controllerAgentName = "prometheus-source-controller"
)

// NewController initializes the controller and is called by the generated code
// Registers event handlers to enqueue events
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	deploymentInformer := deploymentinformer.Get(ctx)
	prometheusSourceInformer := prometheusinformer.Get(ctx)
	eventTypeInformer := eventtypeinformer.Get(ctx)

	r := &Reconciler{
		Base:                   reconciler.NewBase(ctx, controllerAgentName, cmw),
		prometheussourceLister: prometheusSourceInformer.Lister(),
		deploymentLister:       deploymentInformer.Lister(),
		prometheusClientSet:    client.Get(ctx),
	}
	impl := controller.NewImpl(r, r.Logger, ReconcilerName)
	r.sinkResolver = resolver.NewURIResolver(ctx, impl.EnqueueKey)

	r.Logger.Info("Setting up event handlers")
	prometheusSourceInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	deploymentInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("PrometheusSource")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	eventTypeInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("PrometheusSource")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}
