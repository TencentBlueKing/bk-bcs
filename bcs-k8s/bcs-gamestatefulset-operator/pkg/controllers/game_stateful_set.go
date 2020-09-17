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

package gamestatefulset

import (
	"fmt"
	"reflect"
	"time"

	stsplus "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamestatefulset-operator/pkg/apis/tkex/v1alpha1"
	tkexclientset "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamestatefulset-operator/pkg/clientset/internalclientset"
	tkexscheme "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamestatefulset-operator/pkg/clientset/internalclientset/scheme"
	gamestateinformers "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamestatefulset-operator/pkg/informers/tkex/v1alpha1"
	gamestatelister "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamestatefulset-operator/pkg/listers/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamestatefulset-operator/pkg/util/constants"
	"github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/update/hotpatchupdate"
	"github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/update/inplaceupdate"
	"github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/util/requeueduration"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/controller/history"
)

var (
	// controllerKind contains the schema.GroupVersionKind for this controller type.
	controllerKind = stsplus.SchemeGroupVersion.WithKind("GameStatefulSet")

	durationStore = requeueduration.DurationStore{}
)

// GameStatefulSetController controls statefulsets.
type GameStatefulSetController struct {
	// client interface
	kubeClient clientset.Interface
	// tkexClient is a clientset for our own API group.
	tkexClient tkexclientset.Interface
	// control returns an interface capable of syncing a stateful set.
	// Abstracted out for testing.
	control GameStatefulSetControlInterface
	// podControl is used for patching pods.
	podControl controller.PodControlInterface
	// podLister is able to list/get pods from a shared informer's store
	podLister corelisters.PodLister
	// podListerSynced returns true if the pod shared informer has synced at least once
	podListerSynced cache.InformerSynced
	// setLister is able to list/get stateful sets from a shared informer's store
	setLister gamestatelister.GameStatefulSetLister
	// setListerSynced returns true if the stateful set shared informer has synced at least once
	setListerSynced cache.InformerSynced
	// pvcListerSynced returns true if the pvc shared informer has synced at least once
	pvcListerSynced cache.InformerSynced
	// revListerSynced returns true if the rev shared informer has synced at least once
	revListerSynced cache.InformerSynced
	// GameStatefulSetes that need to be synced.
	queue workqueue.RateLimitingInterface
}

// NewGameStatefulSetController creates a new statefulset controller.
func NewGameStatefulSetController(
	podInformer coreinformers.PodInformer,
	setInformer gamestateinformers.GameStatefulSetInformer,
	pvcInformer coreinformers.PersistentVolumeClaimInformer,
	revInformer appsinformers.ControllerRevisionInformer,
	kubeClient clientset.Interface,
	stsplusClient tkexclientset.Interface,
) *GameStatefulSetController {

	tkexscheme.AddToScheme(scheme.Scheme)
	klog.V(3).Info("GameStatefulSet Controller is creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: constants.OperatorName})

	ssc := &GameStatefulSetController{
		kubeClient: kubeClient,
		tkexClient: stsplusClient,
		control: NewDefaultGameStatefulSetControl(
			NewRealGameStatefulSetPodControl(
				kubeClient,
				podInformer.Lister(),
				pvcInformer.Lister(),
				recorder),
			inplaceupdate.NewForTypedClient(kubeClient, apps.ControllerRevisionHashLabelKey),
			hotpatchupdate.NewForTypedClient(kubeClient, apps.ControllerRevisionHashLabelKey),
			NewRealGameStatefulSetStatusUpdater(stsplusClient, setInformer.Lister()),
			history.NewHistory(kubeClient, revInformer.Lister()),
			recorder,
		),
		pvcListerSynced: pvcInformer.Informer().HasSynced,
		queue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), constants.OperatorName),
		podControl:      controller.RealPodControl{KubeClient: kubeClient, Recorder: recorder},
		revListerSynced: revInformer.Informer().HasSynced,
	}

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		// lookup the statefulset and enqueue
		AddFunc: ssc.addPod,
		// lookup current and old statefulset if labels changed
		UpdateFunc: ssc.updatePod,
		// lookup statefulset accounting for deletion tombstones
		DeleteFunc: ssc.deletePod,
	})
	ssc.podLister = podInformer.Lister()
	ssc.podListerSynced = podInformer.Informer().HasSynced

	setInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: ssc.enqueueGameStatefulSet,
			UpdateFunc: func(old, cur interface{}) {
				oldPS := old.(*stsplus.GameStatefulSet)
				curPS := cur.(*stsplus.GameStatefulSet)
				if oldPS.Status.Replicas != curPS.Status.Replicas {
					klog.Infof("Observed updated replica count for GameStatefulSet: %v, %d->%d", curPS.Name, oldPS.Status.Replicas, curPS.Status.Replicas)
				}
				ssc.enqueueGameStatefulSet(cur)
			},
			DeleteFunc: ssc.enqueueGameStatefulSet,
		},
	)
	ssc.setLister = setInformer.Lister()
	ssc.setListerSynced = setInformer.Informer().HasSynced

	// TODO: Watch volumes
	return ssc
}

// Run runs the statefulset controller.
func (ssc *GameStatefulSetController) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer ssc.queue.ShutDown()

	klog.Infof("Starting stateful set controller")
	defer klog.Infof("Shutting down statefulset controller")

	if !controller.WaitForCacheSync(constants.OperatorName, stopCh, ssc.podListerSynced, ssc.setListerSynced, ssc.pvcListerSynced, ssc.revListerSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < workers; i++ {
		go wait.Until(ssc.worker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// addPod adds the statefulset for the pod to the sync queue
func (ssc *GameStatefulSetController) addPod(obj interface{}) {

	pod := obj.(*v1.Pod)

	if pod.DeletionTimestamp != nil {
		// on a restart of the controller manager, it's possible a new pod shows up in a state that
		// is already pending deletion. Prevent the pod from being a creation observation.
		ssc.deletePod(pod)
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(pod); controllerRef != nil {
		set := ssc.resolveControllerRef(pod.Namespace, controllerRef)
		if set == nil {
			klog.V(4).Infof("Pod %s/%s not controlled by GameStatefulSet-Operator", pod.Namespace, pod.Name)
			return
		}
		klog.Infof("Pod %s/%s created, labels: %+v", pod.Namespace, pod.Name, pod.Labels)
		ssc.enqueueGameStatefulSet(set)
		return
	}

	// Otherwise, it's an orphan. Get a list of all matching controllers and sync
	// them to see if anyone wants to adopt it.
	sets := ssc.getGameStatefulSetsForPod(pod)
	if len(sets) == 0 {
		klog.Infof("Pod %s/%s is orphan, but not controlled by GameStatefulSet-Operator", pod.Namespace, pod.Name)
		return
	}
	klog.Infof("Orphan Pod %s/%s created, labels: %+v", pod.Namespace, pod.Name, pod.Labels)
	for _, set := range sets {
		ssc.enqueueGameStatefulSet(set)
	}
}

// updatePod adds the statefulset for the current and old pods to the sync queue.
func (ssc *GameStatefulSetController) updatePod(old, cur interface{}) {
	curPod := cur.(*v1.Pod)
	oldPod := old.(*v1.Pod)
	if curPod.ResourceVersion == oldPod.ResourceVersion {
		// Periodic resync will send update events for all known pods.
		// Two different versions of the same pod will always have different RVs.
		klog.V(4).Infof("Pod %s/%s update event trigger, but nohting changed, ResourceVersion: %s", curPod.Namespace, curPod.Name, curPod.ResourceVersion)
		return
	}

	labelChanged := !reflect.DeepEqual(curPod.Labels, oldPod.Labels)

	curControllerRef := metav1.GetControllerOf(curPod)
	oldControllerRef := metav1.GetControllerOf(oldPod)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	if controllerRefChanged && oldControllerRef != nil {
		// The ControllerRef was changed. Sync the old controller, if any.
		if set := ssc.resolveControllerRef(oldPod.Namespace, oldControllerRef); set != nil {
			ssc.enqueueGameStatefulSet(set)
		}
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		set := ssc.resolveControllerRef(curPod.Namespace, curControllerRef)
		if set == nil {
			return
		}
		klog.V(4).Infof("Pod %s updated, objectMeta %+v -> %+v.", curPod.Name, oldPod.ObjectMeta, curPod.ObjectMeta)
		ssc.enqueueGameStatefulSet(set)
		return
	}

	// Otherwise, it's an orphan. If anything changed, sync matching controllers
	// to see if anyone wants to adopt it now.
	if labelChanged || controllerRefChanged {
		sets := ssc.getGameStatefulSetsForPod(curPod)
		if len(sets) == 0 {
			klog.V(4).Infof("Pod %s/%s is orphan in updated, but not controlled by GameStatefulSet-Operator", curPod.Namespace, curPod.Name)
			return
		}
		klog.Infof("Orphan Pod %s/%s updated, objectMeta %+v -> %+v.", curPod.Namespace, curPod.Name, oldPod.ObjectMeta, curPod.ObjectMeta)
		for _, set := range sets {
			ssc.enqueueGameStatefulSet(set)
		}
	}
}

// deletePod enqueues the statefulset for the pod accounting for deletion tombstones.
func (ssc *GameStatefulSetController) deletePod(obj interface{}) {
	pod, ok := obj.(*v1.Pod)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value. Note that this value might be stale. If the pod
	// changed labels the new GameStatefulSet will not be woken up till the periodic resync.
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
	set := ssc.resolveControllerRef(pod.Namespace, controllerRef)
	if set == nil {
		return
	}
	klog.V(3).Infof("Pod %s/%s deleted through %v.", pod.Namespace, pod.Name, utilruntime.GetCaller())
	ssc.enqueueGameStatefulSet(set)
}

// getPodsForGameStatefulSet returns the Pods that a given GameStatefulSet should manage.
// It also reconciles ControllerRef by adopting/orphaning.
//
// NOTE: Returned Pods are pointers to objects from the cache.
//       If you need to modify one, you need to copy it first.
func (ssc *GameStatefulSetController) getPodsForGameStatefulSet(set *stsplus.GameStatefulSet, selector labels.Selector) ([]*v1.Pod, error) {
	// List all pods to include the pods that don't match the selector anymore but
	// has a ControllerRef pointing to this GameStatefulSet.
	pods, err := ssc.podLister.Pods(set.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	filter := func(pod *v1.Pod) bool {
		// Only claim if it matches our GameStatefulSet name. Otherwise release/ignore.
		return isMemberOf(set, pod)
	}

	// If any adoptions are attempted, we should first recheck for deletion with
	// an uncached quorum read sometime after listing Pods (see #42639).
	canAdoptFunc := controller.RecheckDeletionTimestamp(func() (metav1.Object, error) {
		fresh, err := ssc.tkexClient.TkexV1alpha1().GameStatefulSets(set.Namespace).Get(set.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		if fresh.UID != set.UID {
			return nil, fmt.Errorf("original GameStatefulSet %v/%v is gone: got uid %v, wanted %v", set.Namespace, set.Name, fresh.UID, set.UID)
		}
		return fresh, nil
	})

	cm := controller.NewPodControllerRefManager(ssc.podControl, set, selector, controllerKind, canAdoptFunc)
	return cm.ClaimPods(pods, filter)
}

// adoptOrphanRevisions adopts any orphaned ControllerRevisions matched by set's Selector.
func (ssc *GameStatefulSetController) adoptOrphanRevisions(set *stsplus.GameStatefulSet) error {
	revisions, err := ssc.control.ListRevisions(set)
	if err != nil {
		return err
	}
	hasOrphans := false
	for i := range revisions {
		if metav1.GetControllerOf(revisions[i]) == nil {
			hasOrphans = true
			break
		}
	}
	if hasOrphans {
		fresh, err := ssc.tkexClient.TkexV1alpha1().GameStatefulSets(set.Namespace).Get(set.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if fresh.UID != set.UID {
			return fmt.Errorf("original GameStatefulSet %v/%v is gone: got uid %v, wanted %v", set.Namespace, set.Name, fresh.UID, set.UID)
		}
		return ssc.control.AdoptOrphanRevisions(set, revisions)
	}
	return nil
}

// getGameStatefulSetsForPod returns a list of GameStatefulSetes that potentially match
// a given pod.
func (ssc *GameStatefulSetController) getGameStatefulSetsForPod(pod *v1.Pod) []*stsplus.GameStatefulSet {
	sets, err := GetPodGameStatefulSets(pod, ssc.setLister)
	if err != nil {
		return nil
	}
	// More than one set is selecting the same Pod
	if len(sets) > 1 {
		// ControllerRef will ensure we don't do anything crazy, but more than one
		// item in this list nevertheless constitutes user error.
		utilruntime.HandleError(
			fmt.Errorf(
				"user error: more than one GameStatefulSet is selecting pods with labels: %+v",
				pod.Labels))
	}
	return sets
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the correct Kind.
func (ssc *GameStatefulSetController) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *stsplus.GameStatefulSet {
	// We can't look up by UID, so look up by Name and then verify UID.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		return nil
	}
	set, err := ssc.setLister.GameStatefulSets(namespace).Get(controllerRef.Name)
	if err != nil {
		return nil
	}
	if set.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return set
}

// enqueueGameStatefulSet enqueues the given statefulset in the work queue.
func (ssc *GameStatefulSetController) enqueueGameStatefulSet(obj interface{}) {
	key, err := controller.KeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Cound't get key for object %+v: %v", obj, err))
		return
	}
	klog.V(4).Infof("enqueueGameStatefulSet enqueue item: %s", key)
	ssc.queue.Add(key)
}

// processNextWorkItem dequeues items, processes them, and marks them done. It enforces that the syncHandler is never
// invoked concurrently with the same key.
func (ssc *GameStatefulSetController) processNextWorkItem() bool {
	key, quit := ssc.queue.Get()
	if quit {
		return false
	}
	defer ssc.queue.Done(key)
	klog.Infof("processNextWorkItem get item: %#v", key)
	if err := ssc.sync(key.(string)); err != nil {
		utilruntime.HandleError(fmt.Errorf("Error syncing GameStatefulSet %v, requeuing: %v", key.(string), err))
		ssc.queue.AddRateLimited(key)
	} else {
		ssc.queue.Forget(key)
	}
	return true
}

// worker runs a worker goroutine that invokes processNextWorkItem until the controller's queue is closed
func (ssc *GameStatefulSetController) worker() {
	for ssc.processNextWorkItem() {
	}
}

// sync syncs the given statefulset.
func (ssc *GameStatefulSetController) sync(key string) error {
	startTime := time.Now()
	defer func() {
		klog.V(3).Infof("Finished syncing gamestatefulset %q (%v)", key, time.Since(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	set, err := ssc.setLister.GameStatefulSets(namespace).Get(name)
	if errors.IsNotFound(err) {
		klog.Infof("GameStatefulSet %s has been deleted", key)
		return nil
	}
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to retrieve GameStatefulSet %v from store: %v", key, err))
		return err
	}

	selector, err := metav1.LabelSelectorAsSelector(set.Spec.Selector)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("error converting GameStatefulSet %v selector: %v", key, err))
		// This is a non-transient error, so don't retry.
		return nil
	}

	if err := ssc.adoptOrphanRevisions(set); err != nil {
		return err
	}

	pods, err := ssc.getPodsForGameStatefulSet(set, selector)
	if err != nil {
		return err
	}

	return ssc.syncGameStatefulSet(set, pods)
}

// obj could be an GameStatefulSet, or a DeletionFinalStateUnknown marker item.
func (ssc *GameStatefulSetController) enqueueReplicaSetAfter(obj interface{}, after time.Duration) {
	key, err := controller.KeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %+v: %v", obj, err))
		return
	}
	ssc.queue.AddAfter(key, after)
}

// syncGameStatefulSet syncs a tuple of (statefulset, []*v1.Pod).
func (ssc *GameStatefulSetController) syncGameStatefulSet(set *stsplus.GameStatefulSet, pods []*v1.Pod) error {
	//klog.Infof("Syncing GameStatefulSet %s/%s with %d pods", set.Namespace, set.Name, len(pods))
	// TODO: investigate where we mutate the set during the update as it is not obvious.
	err := ssc.control.UpdateGameStatefulSet(set.DeepCopy(), pods)

	delayDuration := durationStore.Pop(getGameStatefulSetKey(set))
	if delayDuration > 0 {
		ssc.enqueueReplicaSetAfter(set, delayDuration)
	}

	if err != nil {
		return err
	}
	klog.Infof("Successfully synced GameStatefulSet %s/%s successful, pod length: %d", set.Namespace, set.Name, len(pods))
	return nil
}
