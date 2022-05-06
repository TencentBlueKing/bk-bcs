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

package gamedeployment

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	common "github.com/Tencent/bk-bcs/bcs-common/common/version"
	gdv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	gdclientset "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/client/clientset/versioned"
	gdscheme "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/client/clientset/versioned/scheme"
	gdinformers "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/client/informers/externalversions/tkex/v1alpha1"
	gadlister "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/client/listers/tkex/v1alpha1"
	gdcore "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/core"
	gdmetrics "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/metrics"
	revisioncontrol "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/revision"
	scalecontrol "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/scale"
	updatecontrol "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/update"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/util"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/util/constants"
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	hookclientset "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/clientset/versioned"
	hookinformers "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/informers/externalversions/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/postinplace"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/predelete"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/preinplace"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/expectations"

	v1 "k8s.io/api/core/v1"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/controller/history"
)

var (
	scaleExpectations  = expectations.NewScaleExpectations()
	updateExpectations = expectations.NewUpdateExpectations(util.GetPodRevision)
)

// GameDeploymentController controls gamedeployments, is responsible for synchronizing Gamedeployment objects stored
// in the system with actual running pods
type GameDeploymentController struct {
	// client interface
	kubeClient clientset.Interface
	// apiextension client interface
	apiextensionClient apiextension.Interface
	// GroupVersionKind indicates the controller type.
	// Different instances of this struct may handle different GVKs.
	// For example, this struct can be used (with adapters) to handle GameDeploymentController.
	schema.GroupVersionKind
	// gdClient is a clientset for our own API group.
	gdClient gdclientset.Interface
	// podControl is used for patching pods.
	podControl controller.PodControlInterface
	// podLister is able to list/get pods from a shared informer's store
	podLister corelisters.PodLister
	// podListerSynced returns true if the pod shared informer has synced at least once
	podListerSynced cache.InformerSynced
	// nodeLister is able to list/get nodes from a shared informer's store
	nodeLister corelisters.NodeLister
	// nodeListerSynced returns true if the node shared informer has synced at least once
	nodeListerSynced cache.InformerSynced
	// gdLister is able to list/get  gamedeployments from a shared informer's store
	gdLister gadlister.GameDeploymentLister
	// gdListerSynced returns true if the gamedeployments store has been synced at least once.
	gdListerSynced cache.InformerSynced
	// revListerSynced returns true if the rev shared informer has synced at least once
	revListerSynced cache.InformerSynced

	hookRunListerSynced      cache.InformerSynced
	hookTemplateListerSynced cache.InformerSynced

	control GameDeploymentControlInterface

	// Controllers that need to be synced
	queue workqueue.RateLimitingInterface
	// metrics used to collect prom metrics
	metrics *gdmetrics.Metrics
}

// NewGameDeploymentController creates a new gamedeployment controller.
func NewGameDeploymentController(
	podInformer coreinformers.PodInformer,
	nodeInformer coreinformers.NodeInformer,
	deployInformer gdinformers.GameDeploymentInformer,
	hookRunInformer hookinformers.HookRunInformer,
	hookTemplateInformer hookinformers.HookTemplateInformer,
	revInformer appsinformers.ControllerRevisionInformer,
	kubeClient clientset.Interface,
	apiextensionClient apiextension.Interface,
	gdClient gdclientset.Interface,
	recorder record.EventRecorder,
	hookClient hookclientset.Interface,
	historyClient history.Interface) *GameDeploymentController {

	gdscheme.AddToScheme(scheme.Scheme)

	preDeleteControl := predelete.New(kubeClient, hookClient, recorder, hookRunInformer.Lister(), hookTemplateInformer.Lister())
	preInplaceControl := preinplace.New(kubeClient, hookClient, recorder, hookRunInformer.Lister(), hookTemplateInformer.Lister())
	postInpalceControl := postinplace.New(kubeClient, hookClient, recorder,
		hookRunInformer.Lister(), hookTemplateInformer.Lister())
	metrics := gdmetrics.NewMetrics()
	gdc := &GameDeploymentController{
		kubeClient:         kubeClient,
		apiextensionClient: apiextensionClient,
		GroupVersionKind:   util.ControllerKind,
		gdClient:           gdClient,
		control: NewDefaultGameDeploymentControl(
			kubeClient,
			gdClient,
			hookClient,
			podInformer.Lister(),
			hookRunInformer.Lister(),
			hookTemplateInformer.Lister(),
			scalecontrol.New(kubeClient, gdClient, recorder, scaleExpectations, hookRunInformer.Lister(),
				hookTemplateInformer.Lister(), nodeInformer.Lister(), preDeleteControl, metrics),
			updatecontrol.New(kubeClient, recorder, scaleExpectations, updateExpectations, hookRunInformer.Lister(),
				hookTemplateInformer.Lister(), preDeleteControl, preInplaceControl, postInpalceControl, metrics),
			NewRealGameDeploymentStatusUpdater(gdClient, deployInformer.Lister(), recorder, metrics),
			historyClient,
			revisioncontrol.NewRevisionControl(),
			recorder,
			preDeleteControl,
			metrics,
		),
		queue:                    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), constants.GameDeploymentController),
		metrics:                  metrics,
		revListerSynced:          revInformer.Informer().HasSynced,
		podControl:               controller.RealPodControl{KubeClient: kubeClient, Recorder: recorder},
		hookRunListerSynced:      hookRunInformer.Informer().HasSynced,
		hookTemplateListerSynced: hookTemplateInformer.Informer().HasSynced,
	}

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		// lookup the gamedeployment and enqueue
		AddFunc: gdc.addPod,
		// lookup current and old gamedeployment if labels changed
		UpdateFunc: gdc.updatePod,
		// lookup gamedeployment accounting for deletion tombstones
		DeleteFunc: gdc.deletePod,
	})
	gdc.podLister = podInformer.Lister()
	gdc.podListerSynced = podInformer.Informer().HasSynced
	gdc.nodeLister = nodeInformer.Lister()
	gdc.nodeListerSynced = nodeInformer.Informer().HasSynced

	deployInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: gdc.enqueueGameDeployment,
			UpdateFunc: func(old, cur interface{}) {
				oldPS := old.(*gdv1alpha1.GameDeployment)
				curPS := cur.(*gdv1alpha1.GameDeployment)
				if oldPS.Status.Replicas != curPS.Status.Replicas {
					klog.Infof("Observed updated replica count for GameDeployment: %v, %d->%d", curPS.Name, oldPS.Status.Replicas, curPS.Status.Replicas)
				}
				gdc.enqueueGameDeployment(cur)
			},
			DeleteFunc: gdc.enqueueGameDeployment,
		},
	)
	gdc.gdLister = deployInformer.Lister()
	gdc.gdListerSynced = deployInformer.Informer().HasSynced

	hookRunInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			gdc.enqueueGameDeploymentForHook(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newHookRun := newObj.(*hookv1alpha1.HookRun)
			oldHookRun := oldObj.(*hookv1alpha1.HookRun)
			if newHookRun.Status.Phase == oldHookRun.Status.Phase {
				return
			}
			gdc.enqueueGameDeploymentForHook(newObj)
		},
		DeleteFunc: func(obj interface{}) {
			gdc.enqueueGameDeploymentForHook(obj)
		},
	})

	return gdc
}

// enqueueGameDeploymentForHook enqueue a GameDeployment caused by HookRun
func (gdc *GameDeploymentController) enqueueGameDeploymentForHook(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		klog.Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}

	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		refGV, err := schema.ParseGroupVersion(ownerRef.APIVersion)
		if err != nil {
			klog.Errorf("Could not parse OwnerReference %v APIVersion: %v", ownerRef, err)
			return
		}
		// If this object is not owned by GameDeployment, we should not do anything more with it.
		if ownerRef.Kind != util.ControllerKind.Kind || refGV.Group != util.ControllerKind.Group {
			return
		}
		namespace := object.GetNamespace()
		deploy := cache.ExplicitKey(namespace + "/" + ownerRef.Name)
		klog.Infof("Enqueuing GameDeployment %s for HookRun %s/%s", deploy, namespace, object.GetName())
		gdc.enqueueGameDeployment(deploy)
	}
}

// Run runs the gamedeployment controller.
func (gdc *GameDeploymentController) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer gdc.queue.ShutDown()

	klog.Infof("Starting gamedeployment controller")
	defer klog.Infof("Shutting down gamedeployment controller")

	if !cache.WaitForNamedCacheSync(constants.GameDeploymentController, stopCh, gdc.podListerSynced, gdc.gdListerSynced,
		gdc.revListerSynced, gdc.hookRunListerSynced, gdc.hookTemplateListerSynced, gdc.nodeListerSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	imageVersion, CRDVersion := gdc.getVersion()
	gdc.metrics.CollectOperatorVersion(imageVersion, CRDVersion,
		common.BcsVersion, common.BcsGitHash, common.BcsBuildTime)

	for i := 0; i < workers; i++ {
		go wait.Until(gdc.worker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// addPod adds the gamedeployment for the pod to the sync queue
func (gdc *GameDeploymentController) addPod(obj interface{}) {
	pod := obj.(*v1.Pod)
	if pod.DeletionTimestamp != nil {
		// on a restart of the controller manager, it's possible a new pod shows up in a state that
		// is already pending deletion. Prevent the pod from being a creation observation.
		gdc.deletePod(pod)
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(pod); controllerRef != nil {
		deploy := gdc.resolveControllerRef(pod.Namespace, controllerRef)
		if deploy == nil {
			klog.V(4).Infof("Pod %s/%s not controlled by GameDeployment-Operator", pod.Namespace, pod.Name)
			return
		}
		key := fmt.Sprintf("%s/%s", deploy.Namespace, deploy.Name)
		klog.Infof("Pod %s/%s created, labels: %+v, owner: %s", pod.Namespace, pod.Name, pod.Labels, key)
		scaleExpectations.ObserveScale(key, expectations.Create, pod.Name)
		gdc.enqueueGameDeployment(deploy)
		return
	}

	// Otherwise, it's an orphan. Get a list of all matching controllers and sync
	// them to see if anyone wants to adopt it.
	deploys := gdc.getDeploymentsForPod(pod)
	if len(deploys) == 0 {
		klog.Infof("Pod %s/%s is orphan, but not controlled by GameDeployment-Operator", pod.Namespace, pod.Name)
		return
	}
	klog.Infof("Orphan Pod %s/%s created, labels: %+v", pod.Namespace, pod.Name, pod.Labels)
	for _, deploy := range deploys {
		gdc.enqueueGameDeployment(deploy)
	}
}

// updatePod adds the gamedeployment for the current and old pods to the sync queue.
func (gdc *GameDeploymentController) updatePod(old, cur interface{}) {
	curPod := cur.(*v1.Pod)
	oldPod := old.(*v1.Pod)
	if curPod.ResourceVersion == oldPod.ResourceVersion {
		// Periodic resync will send update events for all known pods.
		// Two different versions of the same pod will always have different RVs.
		klog.V(4).Infof("Pod %s/%s update event trigger, but nothing changed, ResourceVersion: %s", curPod.Namespace, curPod.Name, curPod.ResourceVersion)
		return
	}

	labelChanged := !reflect.DeepEqual(curPod.Labels, oldPod.Labels)
	if curPod.DeletionTimestamp != nil {
		// when a pod is deleted gracefully it's deletion timestamp is first modified to reflect a grace period,
		// and after such time has passed, the kubelet actually deletes it from the store. We receive an update
		// for modification of the deletion timestamp and expect an rs to create more replicas asap, not wait
		// until the kubelet actually deletes the pod. This is different from the Phase of a pod changing, because
		// an rs never initiates a phase change, and so is never asleep waiting for the same.
		gdc.deletePod(curPod)
		if labelChanged {
			// we don't need to check the oldPod.DeletionTimestamp because DeletionTimestamp cannot be unset.
			gdc.deletePod(oldPod)
		}
		return
	}

	curControllerRef := metav1.GetControllerOf(curPod)
	oldControllerRef := metav1.GetControllerOf(oldPod)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	if controllerRefChanged && oldControllerRef != nil {
		// The ControllerRef was changed. Sync the old controller, if any.
		if deploy := gdc.resolveControllerRef(oldPod.Namespace, oldControllerRef); deploy != nil {
			gdc.enqueueGameDeployment(deploy)
		}
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		deploy := gdc.resolveControllerRef(curPod.Namespace, curControllerRef)
		if deploy == nil {
			return
		}
		key := fmt.Sprintf("%s/%s", deploy.Namespace, deploy.Name)
		klog.V(4).Infof("Pod %s updated, objectMeta %+v -> %+v, owner: %s.", curPod.Name, oldPod.ObjectMeta, curPod.ObjectMeta, key)
		gdc.enqueueGameDeployment(deploy)
		return
	}

	// Otherwise, it's an orphan. If anything changed, sync matching controllers
	// to see if anyone wants to adopt it now.
	if labelChanged || controllerRefChanged {
		deploys := gdc.getDeploymentsForPod(curPod)
		if len(deploys) == 0 {
			klog.V(4).Infof("Pod %s/%s is orphan in updated, but not controlled by GameDeployment-Operator", curPod.Namespace, curPod.Name)
			return
		}
		klog.Infof("Orphan Pod %s/%s updated, objectMeta %+v -> %+v.", curPod.Namespace, curPod.Name, oldPod.ObjectMeta, curPod.ObjectMeta)
		for _, deploy := range deploys {
			gdc.enqueueGameDeployment(deploy)
		}
	}
}

// deletePod enqueues the gamedeployment for the pod accounting for deletion tombstones.
func (gdc *GameDeploymentController) deletePod(obj interface{}) {
	pod, ok := obj.(*v1.Pod)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value. Note that this value might be stale. If the pod
	// changed labels the new GameDeployment will not be woken up till the periodic resync.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("couldn't get object from tombstone %+v", obj))
			return
		}
		pod, ok = tombstone.Obj.(*v1.Pod)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("tombstone contained object that is not a pod %+v", obj))
			return
		}
	}

	controllerRef := metav1.GetControllerOf(pod)
	if controllerRef == nil {
		// No controller should care about orphans being deleted.
		return
	}
	deploy := gdc.resolveControllerRef(pod.Namespace, controllerRef)
	if deploy == nil {
		return
	}
	key := fmt.Sprintf("%s/%s", deploy.Namespace, deploy.Name)
	klog.V(3).Infof("Pod %s/%s deleted through %v, owner: %s.", pod.Namespace, pod.Name, utilruntime.GetCaller(), key)
	scaleExpectations.ObserveScale(key, expectations.Delete, pod.Name)
	gdc.enqueueGameDeployment(deploy)
}

// getGameDeploymentForPod returns a list of GameDeployments that potentially match
// a given pod.
func (gdc *GameDeploymentController) getDeploymentsForPod(pod *v1.Pod) []*gdv1alpha1.GameDeployment {
	deploys, err := util.GetPodGameDeployments(pod, gdc.gdLister)
	if err != nil {
		return nil
	}
	// More than one set is selecting the same Pod
	if len(deploys) > 1 {
		// ControllerRef will ensure we don't do anything crazy, but more than one
		// item in this list nevertheless constitutes user error.
		utilruntime.HandleError(
			fmt.Errorf(
				"user error: more than one GameDeployment is selecting pods with labels: %+v",
				pod.Labels))
	}
	return deploys
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the correct Kind.
func (gdc *GameDeploymentController) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *gdv1alpha1.GameDeployment {
	// Parse the Group out of the OwnerReference to compare it to what was parsed out of the requested OwnerType
	refGV, err := schema.ParseGroupVersion(controllerRef.APIVersion)
	if err != nil {
		klog.Errorf("Could not parse OwnerReference %v APIVersion: %v", controllerRef, err)
		return nil
	}

	// We can't look up by UID, so look up by Name and then verify UID.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != util.ControllerKind.Kind || refGV.Group != util.ControllerKind.Group {
		return nil
	}
	deploy, err := gdc.gdLister.GameDeployments(namespace).Get(controllerRef.Name)
	if err != nil {
		return nil
	}

	if deploy.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return deploy
}

// enqueueGameDeployment enqueues the given gamedeployment in the work queue.
func (gdc *GameDeploymentController) enqueueGameDeployment(obj interface{}) {
	key, err := controller.KeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("cound't get key for object %+v: %v", obj, err))
		return
	}
	klog.V(4).Infof("enqueueGameDeployment enqueue item: %s", key)
	gdc.queue.Add(key)
}

// obj could be an GameDeployment, or a DeletionFinalStateUnknown marker item.
func (gdc *GameDeploymentController) enqueueGameDeploymentAfter(obj interface{}, after time.Duration) {
	key, err := controller.KeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %+v: %v", obj, err))
		return
	}
	gdc.queue.AddAfter(key, after)
}

// processNextWorkItem dequeues items, processes them, and marks them done. It enforces that the syncHandler is never
// invoked concurrently with the same key.
func (gdc *GameDeploymentController) processNextWorkItem() bool {
	key, quit := gdc.queue.Get()
	if quit {
		return false
	}
	defer gdc.queue.Done(key)
	klog.Infof("processNextWorkItem get item: %#v", key)
	if err := gdc.sync(key.(string)); err != nil {
		utilruntime.HandleError(fmt.Errorf("error syncing GameDeployment %v, requeuing: %v", key.(string), err))
		gdc.queue.AddRateLimited(key)
	} else {
		gdc.queue.Forget(key)
	}
	return true
}

// worker runs a worker goroutine that invokes processNextWorkItem until the controller's queue is closed
func (gdc *GameDeploymentController) worker() {
	for gdc.processNextWorkItem() {
	}
}

// sync syncs the given gamedeployment.
func (gdc *GameDeploymentController) sync(key string) (retErr error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		if retErr == nil {
			klog.Infof("Finished syncing GameDeployment %s, cost time: %v", key, duration)
			gdc.metrics.CollectReconcileDuration(key, gdmetrics.SuccessStatus, duration)

		} else {
			klog.Errorf("Failed syncing GameDeployment %s, err: %v", key, retErr)
			gdc.metrics.CollectReconcileDuration(key, gdmetrics.FailureStatus, duration)
		}
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	// in some case, the GameDeployment get from the informer cache may not be the latest, so get from apiserver directly
	// deploy, err := gdc.gdClient.TkexV1alpha1().GameDeployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	cachedDeploy, err := gdc.gdLister.GameDeployments(namespace).Get(name)
	deploy := cachedDeploy.DeepCopy()

	if errors.IsNotFound(err) {
		// Object not found, return.  Created objects are automatically garbage collected.
		// For additional cleanup logic use finalizers.
		klog.Infof("GameDeployment %s has been deleted", key)
		scaleExpectations.DeleteExpectations(key)
		updateExpectations.DeleteExpectations(key)
		return nil
	}
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to retrieve GameDeployment %v from store: %v", key, err))
		return err
	}

	// It's strange that the GameDeployment's GroupVersionKind is nil, to have to set it here
	deploy.SetGroupVersionKind(util.ControllerKind)

	coreControl := gdcore.New(deploy)
	if coreControl.IsInitializing() {
		klog.V(4).Infof("GameDeployment %s skip sync for initializing", key)
		return nil
	}

	// If scaling expectations have not satisfied yet, just skip this sync.
	// if scaleSatisfied, scaleDirtyPods := scaleExpectations.SatisfiedExpectations(key); !scaleSatisfied {
	// 	klog.V(4).Infof("Not satisfied scale for %v, scaleDirtyPods=%v", key, scaleDirtyPods)
	// 	return nil
	// }

	selector, err := metav1.LabelSelectorAsSelector(deploy.Spec.Selector)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("error converting GameDeployment %v selector: %v", key, err))
		// This is a non-transient error, so don't retry.
		return nil
	}

	// list all pods to include the pods that don't match the deploy`s selector
	// anymore but has the stale controller ref.
	pods, allPods, err := gdc.getPodsForGameDeployment(deploy, selector)
	if err != nil {
		return err
	}

	delayDuration, newStatus, updateErr := gdc.control.UpdateGameDeployment(deploy, pods, allPods)

	if updateErr == nil && deploy.Spec.MinReadySeconds > 0 && newStatus.AvailableReplicas != newStatus.ReadyReplicas {
		minReadyDuration := time.Second * time.Duration(deploy.Spec.MinReadySeconds)
		if delayDuration == 0 || minReadyDuration < delayDuration {
			delayDuration = minReadyDuration
		}
	}
	if delayDuration > 0 {
		gdc.enqueueGameDeploymentAfter(deploy, delayDuration)
	}

	return updateErr
}

// getPodsForGameDeployment returns the Pods that a given GameDeployment should manage.
// It also reconciles ControllerRef by adopting/orphaning.
//
// NOTE: Returned Pods are pointers to objects from the cache.
//       If you need to modify one, you need to copy it first.
func (gdc *GameDeploymentController) getPodsForGameDeployment(deploy *gdv1alpha1.GameDeployment,
	selector labels.Selector) ([]*v1.Pod, []*v1.Pod, error) {
	// List all pods to include the pods that don't match the selector anymore but
	// has a ControllerRef pointing to this GameDeployment.
	pods, err := gdc.podLister.Pods(deploy.Namespace).List(labels.Everything())
	if err != nil {
		return nil, nil, err
	}

	filter := controller.IsPodActive

	// If any adoptions are attempted, we should first recheck for deletion with
	// an uncached quorum read sometime after listing Pods (see #42639).
	canAdoptFunc := controller.RecheckDeletionTimestamp(func() (metav1.Object, error) {
		fresh, freshErr := gdc.gdClient.TkexV1alpha1().GameDeployments(deploy.Namespace).Get(context.TODO(), deploy.Name, metav1.GetOptions{})
		if freshErr != nil {
			return nil, freshErr
		}
		if fresh.UID != deploy.UID {
			return nil, fmt.Errorf("original GameDeployment %v/%v is gone: got uid %v, wanted %v",
				deploy.Namespace, deploy.Name, fresh.UID, deploy.UID)
		}
		return fresh, nil
	})

	cm := controller.NewPodControllerRefManager(gdc.podControl, deploy, selector, gdc.GroupVersionKind, canAdoptFunc)
	filteredPods, err := cm.ClaimPods(pods, filter)
	if err != nil {
		return nil, nil, err
	}

	// in some operation such as generate unique index, we need the information of all pods (including terminating pods)
	// to ensure that every pods in the gamedeployment has unique index.
	// OwnerReferences would disappear when pod is terminating,
	// use selector to filter out all pods belongs to this deployment.
	allPods, err := gdc.podLister.Pods(deploy.Namespace).List(selector)
	if err != nil {
		return nil, nil, err
	}

	return filteredPods, allPods, nil
}

// getVersion returns the image version of operator pods, and the version of CRD
func (gdc *GameDeploymentController) getVersion() (imageVersion, CRDVerion string) {
	imageVersion, CRDVerion = "", ""

	deploy, err := gdc.kubeClient.AppsV1().Deployments("bcs-system").Get(
		context.TODO(), "bcs-gamedeployment-operator", metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Failed to get deployment: bcs-system/bcs-gamedeployment-operator, error: %s", err.Error())
	} else {
		imageVersion = strings.Split(deploy.Spec.Template.Spec.Containers[0].Image, ":")[1]
	}

	v1crd, err := gdc.apiextensionClient.ApiextensionsV1().CustomResourceDefinitions().Get(
		context.TODO(), "gamedeployments.tkex.tencent.com", metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Failed to get v1 CRD: gamedeployments.tkex.tencent.com, error: %s", err.Error())
	} else {
		CRDVerion = "v1-" + v1crd.GetAnnotations()["version"]
		return imageVersion, CRDVerion
	}

	v1beta1crd, err := gdc.apiextensionClient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(
		context.TODO(), "gamedeployments.tkex.tencent.com", metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Failed to get v1beta1 CRD: gamedeployments.tkex.tencent.com, error: %s", err.Error())
	} else if CRDVerion == "" {
		CRDVerion = "v1beta1-" + v1beta1crd.GetAnnotations()["version"]
	}

	return imageVersion, CRDVerion
}
