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
 *
 */

package hook

import (
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-hook-operator/pkg/providers"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-hook-operator/pkg/util/constants"
	"github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	tkexclientset "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/bcs-hook/client/clientset/versioned"
	tkexscheme "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/bcs-hook/client/clientset/versioned/scheme"
	tkexinformers "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/bcs-hook/client/informers/externalversions/tkex/v1alpha1"
	hooklister "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/bcs-hook/client/listers/tkex/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/controller"
)

// HookController controls HookRuns, is responsible for synchronizing HookRun objects stored in the system
type HookController struct {
	kubeClient    kubernetes.Interface
	tkexClient    tkexclientset.Interface
	hookRunLister hooklister.HookRunLister
	hookRunSynced cache.InformerSynced

	newProvider func(metric v1alpha1.Metric) (providers.Provider, error)
	queue       workqueue.RateLimitingInterface
	recorder    record.EventRecorder
}

// NewHookController create a new HookController
func NewHookController(
	kubeClient kubernetes.Interface,
	tkexClient tkexclientset.Interface,
	hookRunInformer tkexinformers.HookRunInformer,
	recorder record.EventRecorder) *HookController {

	tkexscheme.AddToScheme(scheme.Scheme)
	controller := &HookController{
		kubeClient:    kubeClient,
		tkexClient:    tkexClient,
		hookRunLister: hookRunInformer.Lister(),
		hookRunSynced: hookRunInformer.Informer().HasSynced,
		recorder:      recorder,
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), constants.HookRunController),
	}

	providerFactory := providers.ProviderFactory{
		KubeClient: kubeClient,
	}
	controller.newProvider = providerFactory.NewProvider

	hookRunInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueHookRun,
		UpdateFunc: func(oldObj, newObj interface{}) {
			controller.enqueueHookRun(newObj)
		},
		DeleteFunc: controller.enqueueHookRun,
	})
	return controller
}

func (hc *HookController) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer hc.queue.ShutDown()

	klog.Infof("Starting HookRun controller")
	defer klog.Infof("Shutting down HookRun controller")

	if !controller.WaitForCacheSync(constants.HookRunController, stopCh, hc.hookRunSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < workers; i++ {
		go wait.Until(hc.worker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// processNextWorkItem dequeues items, processes them, and marks them done. It enforces that the syncHandler is never
// invoked concurrently with the same key.
func (hc *HookController) processNextWorkItem() bool {
	key, quit := hc.queue.Get()
	if quit {
		return false
	}
	defer hc.queue.Done(key)
	klog.Infof("processNextWorkItem get item: %#v", key)
	if err := hc.sync(key.(string)); err != nil {
		utilruntime.HandleError(fmt.Errorf("error syncing HookRun %v, requeuing: %v", key.(string), err))
		hc.queue.AddRateLimited(key)
	} else {
		hc.queue.Forget(key)
	}
	return true
}

// worker runs a worker goroutine that invokes processNextWorkItem until the controller's queue is closed
func (hc *HookController) worker() {
	for hc.processNextWorkItem() {
	}
}

// enqueueHookRun enqueues the given hookrun in the work queue.
func (hc *HookController) enqueueHookRun(obj interface{}) {
	key, err := controller.KeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("cound't get key for object %+v: %v", obj, err))
		return
	}
	klog.V(4).Infof("enqueueHookRun enqueue item: %s", key)
	hc.queue.Add(key)
}

// obj could be an HookRun, or a DeletionFinalStateUnknown marker item.
func (hc *HookController) enqueueHookRunAfter(obj interface{}, after time.Duration) {
	key, err := controller.KeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %+v: %v", obj, err))
		return
	}
	hc.queue.AddAfter(key, after)
}

func (hc *HookController) sync(key string) error {
	startTime := time.Now()

	defer func() {
		klog.V(3).Infof("Finished syncing =HookRun %q (%v)", key, time.Since(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	run, err := hc.hookRunLister.HookRuns(namespace).Get(name)
	if k8serrors.IsNotFound(err) {
		klog.Infof("HookRun %s has been deleted", key)
		return nil
	}
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to retrieve HookRun %v from store: %v", key, err))
		return err
	}
	if run.DeletionTimestamp != nil {
		klog.Info("No reconciliation as HookRun marked for deletion")
		return nil
	}

	updatedRun := hc.reconcileHookRun(run)
	return hc.updateHookRunStatus(run, updatedRun.Status)
}
