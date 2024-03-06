/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package manifest xxx
package manifest

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	appsapi "github.com/clusternet/clusternet/pkg/apis/apps/v1alpha1"
	clusternetclientset "github.com/clusternet/clusternet/pkg/generated/clientset/versioned"
	appinformers "github.com/clusternet/clusternet/pkg/generated/informers/externalversions/apps/v1alpha1"
	applisters "github.com/clusternet/clusternet/pkg/generated/listers/apps/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-clusternet-controller/pkg/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-clusternet-controller/pkg/nspolicy"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-clusternet-controller/pkg/util"
)

// Controller manifest controller
type Controller struct {
	clusternetClient clusternetclientset.Interface

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface

	manifestLister applisters.ManifestLister
	manifestSynced cache.InformerSynced
	nsLister       corelisters.NamespaceLister
	nsSynced       cache.InformerSynced
}

// NewController new controller
func NewController(clusternetClient clusternetclientset.Interface,
	manifestInformer appinformers.ManifestInformer,
	nsInformer coreinformers.NamespaceInformer) (*Controller, error) {
	c := &Controller{
		clusternetClient: clusternetClient,
		workqueue:        workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "manifest"),
		manifestLister:   manifestInformer.Lister(),
		manifestSynced:   manifestInformer.Informer().HasSynced,
		nsLister:         nsInformer.Lister(),
		nsSynced:         nsInformer.Informer().HasSynced,
	}

	// Manage the addition/update of Manifest
	manifestInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addManifest,
		UpdateFunc: c.updateManifest,
		DeleteFunc: c.deleteManifest,
	})

	return c, nil
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Info("starting manifest controller...")
	defer klog.Info("shutting down manifest controller")

	// Wait for the caches to be synced before starting workers
	if !cache.WaitForNamedCacheSync("manifest-controller", stopCh, c.manifestSynced, c.nsSynced) {
		return
	}

	klog.V(5).Infof("starting %d worker threads", workers)
	// Launch workers to process Manifest resources
	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
}

func (c *Controller) addManifest(obj interface{}) {
	manifest := obj.(*appsapi.Manifest)
	klog.V(4).Infof("adding Manifest %q", klog.KObj(manifest))
	c.enqueue(manifest)
}

func (c *Controller) updateManifest(old, cur interface{}) {
	newManifest := cur.(*appsapi.Manifest)
	c.enqueue(newManifest)
}

func (c *Controller) deleteManifest(obj interface{}) {
	manifest, ok := obj.(*appsapi.Manifest)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		manifest, ok = tombstone.Obj.(*appsapi.Manifest)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("tombstone contained object that is not a Manifest %#v", obj))
			return
		}
	}
	klog.V(4).Infof("deleting Manifest %q", klog.KObj(manifest))
	c.enqueue(manifest)
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Manifest resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("successfully synced Manifest %q", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Manifest resource
// with the current status of the resource.
// nolint funlen
func (c *Controller) syncHandler(key string) error {
	// If an error occurs during handling, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.

	// Convert the namespace/name string into a distinct namespace and name
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	klog.V(4).Infof("start processing Manifest %q", key)
	// Get the Manifest resource with this name
	manifest, err := c.manifestLister.Manifests(ns).Get(name)
	// The Manifest resource may no longer exist, in which case we stop processing.
	if errors.IsNotFound(err) {
		klog.V(2).Infof("Manifest %q has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	if manifest.Template.Raw == nil {
		klog.Warning("manifest.Template.Raw is empty, %q", klog.KObj(manifest))
		return nil
	}
	utd := &unstructured.Unstructured{}
	err = json.Unmarshal(manifest.Template.Raw, &utd.Object)
	if err != nil {
		klog.Errorf("unmarshal error, %q, err=%v", klog.KObj(manifest), err)
		return err
	}

	resourceKind := utd.GroupVersionKind().Kind

	matchAnnotations := util.FindAnnotationsMathKeyPrefix(utd.GetAnnotations())
	deleteSubscription := false
	if manifest.DeletionTimestamp != nil {
		// 删除
		deleteSubscription = true
	}
	matchLabels := map[string]string{
		"bkbcs.tencent.com/resource-kind": resourceKind,
		"bkbcs.tencent.com/resource-ns":   utd.GetNamespace(),
		"bkbcs.tencent.com/resource-name": utd.GetName(),
	}
	subscriptionName := c.genAutoCreateSubscriptionName(utd.GetName())
	subscriptionList, err := c.clusternetClient.AppsV1alpha1().Subscriptions(utd.GetNamespace()).List(
		context.Background(),
		metav1.ListOptions{
			LabelSelector: labels.Set(matchLabels).String(),
		})
	if err != nil {
		return err
	}
	// 只会存在0个或1个
	if len(subscriptionList.Items) > 1 {
		return fmt.Errorf("auto create sub matchLabels match %d", len(subscriptionList.Items))
	}
	if deleteSubscription {
		klog.Infof("start delete subscription %s", subscriptionName)
		// 删除Subscription
		err = c.clusternetClient.AppsV1alpha1().Subscriptions(utd.GetNamespace()).Delete(
			context.Background(), subscriptionList.Items[0].Name, metav1.DeleteOptions{})
		if errors.IsNotFound(err) {
			klog.V(2).Infof("Subscription %s:%s has been deleted", ns, name)
			return nil
		}
		if err != nil {
			return err
		}
		return nil
	}
	nsObj, err := c.nsLister.Get(utd.GetNamespace())
	if err != nil {
		return err
	}
	labelSelector, err := c.genSubscriptionLabel(matchAnnotations, nsObj)
	if err != nil {
		return err
	}
	// 更新或创建Subscription
	if len(subscriptionList.Items) == 0 {

		// create
		subscription := &appsapi.Subscription{
			ObjectMeta: metav1.ObjectMeta{
				Name:      subscriptionName,
				Namespace: utd.GetNamespace(),
				Annotations: map[string]string{
					"bkbcs.tencent.com/created-by": "bcs-clusternet-controller",
				},
				Labels: matchLabels,
			},
			Spec: c.genSubscriptionSpec(labelSelector, utd.GroupVersionKind(), utd.GetNamespace(), utd.GetName()),
		}
		klog.Infof("start create Subscriptions %q", klog.KObj(subscription))
		_, err = c.clusternetClient.AppsV1alpha1().Subscriptions(utd.GetNamespace()).Create(
			context.Background(), subscription, metav1.CreateOptions{})
		if err != nil {
			klog.Errorf("create Subscriptions %q error, err=%+v", klog.KObj(subscription), err)
			return err
		}
		return nil
	}
	// update
	matchSubscription := subscriptionList.Items[0]
	matchSubscription.Spec = c.genSubscriptionSpec(
		labelSelector, utd.GroupVersionKind(), utd.GetNamespace(), utd.GetName())
	klog.Infof("start update Subscriptions %q", klog.KObj(&matchSubscription))
	_, err = c.clusternetClient.AppsV1alpha1().Subscriptions(utd.GetNamespace()).Update(
		context.Background(), &matchSubscription, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("update subscriptions %q error, err=%v", klog.KObj(&matchSubscription), err)
		return err
	}
	return nil
}

func (c *Controller) genSubscriptionLabel(
	matchAnnotation map[string]string, namespace *corev1.Namespace) (*metav1.LabelSelector, error) {
	nsPolicy := nspolicy.NewNamespacePolicy(namespace)
	clusterIDs, err := nsPolicy.GetAvailableClusterIDs()
	if err != nil {
		return nil, err
	}
	requirements := make([]metav1.LabelSelectorRequirement, 0)
	for k, v := range matchAnnotation {
		tmpReq := metav1.LabelSelectorRequirement{
			Key:      k,
			Operator: metav1.LabelSelectorOpIn,
			Values: []string{
				v,
			},
		}
		requirements = append(requirements, tmpReq)
	}
	clusterReq := metav1.LabelSelectorRequirement{
		Key:      constant.AnnotationSubscriptionKeyPrefix + "clusterid",
		Operator: metav1.LabelSelectorOpIn,
		Values:   clusterIDs,
	}
	requirements = append(requirements, clusterReq)
	return &metav1.LabelSelector{
		MatchExpressions: requirements,
	}, nil
}

func (c *Controller) genSubscriptionSpec(
	labelSelector *metav1.LabelSelector,
	groupVersionKind schema.GroupVersionKind,
	ns, name string) appsapi.SubscriptionSpec {
	return appsapi.SubscriptionSpec{
		Subscribers: []appsapi.Subscriber{
			{
				ClusterAffinity: labelSelector,
			},
		},
		Feeds: []appsapi.Feed{
			{
				APIVersion: groupVersionKind.GroupVersion().String(),
				Kind:       groupVersionKind.Kind,
				Namespace:  ns,
				Name:       name,
			},
		},
	}
}

func (c *Controller) genAutoCreateSubscriptionName(name string) string {
	return fmt.Sprintf("%s-%s", name, utilrand.String(12))
}

// enqueue takes a Manifest resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Manifest.
func (c *Controller) enqueue(manifest *appsapi.Manifest) {
	key, err := cache.MetaNamespaceKeyFunc(manifest)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}
