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

package statefulsetplus

import (
	"fmt"
	"reflect"
	"time"

	stsplus "bk-bcs/bcs-k8s/tkex-statefulsetplus-operator/pkg/apis/tkex/v1alpha1"
	tkexclientset "bk-bcs/bcs-k8s/tkex-statefulsetplus-operator/pkg/clientset/internalclientset"
	tkexscheme "bk-bcs/bcs-k8s/tkex-statefulsetplus-operator/pkg/clientset/internalclientset/scheme"
	stsplusinformers "bk-bcs/bcs-k8s/tkex-statefulsetplus-operator/pkg/informers/tkex/v1alpha1"
	stspluslisters "bk-bcs/bcs-k8s/tkex-statefulsetplus-operator/pkg/listers/tkex/v1alpha1"
	"bk-bcs/bcs-k8s/tkex-statefulsetplus-operator/pkg/util/constants"

	"github.com/golang/glog"
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
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/controller/history"
)

const (
	// period to relist statefulsets and verify pets
	statefulSetResyncPeriod = 30 * time.Second

	//AnnotationUpgradeContainerInPlaceKey = "tkex/upgrade-in-place"
)

// controllerKind contains the schema.GroupVersionKind for this controller type.
var controllerKind = apps.SchemeGroupVersion.WithKind("StatefulSetPlus")

// StatefulSetPlusController controls statefulsets.
type StatefulSetPlusController struct {
	// client interface
	kubeClient clientset.Interface
	// tkexClient is a clientset for our own API group.
	tkexClient tkexclientset.Interface
	// control returns an interface capable of syncing a stateful set.
	// Abstracted out for testing.
	control StatefulSetPlusControlInterface
	// podControl is used for patching pods.
	podControl controller.PodControlInterface
	// podLister is able to list/get pods from a shared informer's store
	podLister corelisters.PodLister
	// podListerSynced returns true if the pod shared informer has synced at least once
	podListerSynced cache.InformerSynced
	// setLister is able to list/get stateful sets from a shared informer's store
	setLister stspluslisters.StatefulSetPlusLister
	// setListerSynced returns true if the stateful set shared informer has synced at least once
	setListerSynced cache.InformerSynced
	// pvcListerSynced returns true if the pvc shared informer has synced at least once
	pvcListerSynced cache.InformerSynced
	// revListerSynced returns true if the rev shared informer has synced at least once
	revListerSynced cache.InformerSynced
	// StatefulSetPluses that need to be synced.
	queue workqueue.RateLimitingInterface
}

// NewStatefulSetPlusController creates a new statefulset controller.
func NewStatefulSetPlusController(
	podInformer coreinformers.PodInformer,
	setInformer stsplusinformers.StatefulSetPlusInformer,
	pvcInformer coreinformers.PersistentVolumeClaimInformer,
	revInformer appsinformers.ControllerRevisionInformer,
	kubeClient clientset.Interface,
	stsplusClient tkexclientset.Interface,
) *StatefulSetPlusController {

	tkexscheme.AddToScheme(scheme.Scheme)
	glog.V(3).Info("StatefulSetPlus Controller is creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: constants.OperatorName})

	ssc := &StatefulSetPlusController{
		kubeClient: kubeClient,
		tkexClient: stsplusClient,
		control: NewDefaultStatefulSetPlusControl(
			NewRealStatefulPlusPodControl(
				kubeClient,
				podInformer.Lister(),
				pvcInformer.Lister(),
				recorder),
			NewRealStatefulSetPlusStatusUpdater(stsplusClient, setInformer.Lister()),
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

	setInformer.Informer().AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc: ssc.enqueueStatefulSetPlus,
			UpdateFunc: func(old, cur interface{}) {
				oldPS := old.(*stsplus.StatefulSetPlus)
				curPS := cur.(*stsplus.StatefulSetPlus)
				if oldPS.Status.Replicas != curPS.Status.Replicas {
					glog.Infof("Observed updated replica count for StatefulSetPlus: %v, %d->%d", curPS.Name, oldPS.Status.Replicas, curPS.Status.Replicas)
				}
				ssc.enqueueStatefulSetPlus(cur)
			},
			DeleteFunc: ssc.enqueueStatefulSetPlus,
		},
		statefulSetResyncPeriod,
	)
	ssc.setLister = setInformer.Lister()
	ssc.setListerSynced = setInformer.Informer().HasSynced

	// TODO: Watch volumes
	return ssc
}

// Run runs the statefulset controller.
func (ssc *StatefulSetPlusController) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer ssc.queue.ShutDown()

	glog.Infof("Starting stateful set controller")
	defer glog.Infof("Shutting down statefulset controller")

	if !controller.WaitForCacheSync(constants.OperatorName, stopCh, ssc.podListerSynced, ssc.setListerSynced, ssc.pvcListerSynced, ssc.revListerSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < workers; i++ {
		go wait.Until(ssc.worker, time.Second, stopCh)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

// addPod adds the statefulset for the pod to the sync queue
func (ssc *StatefulSetPlusController) addPod(obj interface{}) {

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
			glog.V(4).Infof("Pod %s/%s not controlled by StatefulSetPlus-Operator", pod.Namespace, pod.Name)
			return
		}
		glog.Infof("Pod %s/%s created, labels: %+v", pod.Namespace, pod.Name, pod.Labels)
		ssc.enqueueStatefulSetPlus(set)
		return
	}

	// Otherwise, it's an orphan. Get a list of all matching controllers and sync
	// them to see if anyone wants to adopt it.
	sets := ssc.getStatefulSetPlusesForPod(pod)
	if len(sets) == 0 {
		glog.Infof("Pod %s/%s is orphan, but not controlled by StatefulSetPlus-Operator", pod.Namespace, pod.Name)
		return
	}
	glog.Infof("Orphan Pod %s/%s created, labels: %+v", pod.Namespace, pod.Name, pod.Labels)
	for _, set := range sets {
		ssc.enqueueStatefulSetPlus(set)
	}
}

// updatePod adds the statefulset for the current and old pods to the sync queue.
func (ssc *StatefulSetPlusController) updatePod(old, cur interface{}) {
	curPod := cur.(*v1.Pod)
	oldPod := old.(*v1.Pod)
	if curPod.ResourceVersion == oldPod.ResourceVersion {
		// Periodic resync will send update events for all known pods.
		// Two different versions of the same pod will always have different RVs.
		glog.V(4).Infof("Pod %s/%s update event trigger, but nohting changed, ResourceVersion: %s", curPod.Namespace, curPod.Name, curPod.ResourceVersion)
		return
	}

	labelChanged := !reflect.DeepEqual(curPod.Labels, oldPod.Labels)

	curControllerRef := metav1.GetControllerOf(curPod)
	oldControllerRef := metav1.GetControllerOf(oldPod)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	if controllerRefChanged && oldControllerRef != nil {
		// The ControllerRef was changed. Sync the old controller, if any.
		if set := ssc.resolveControllerRef(oldPod.Namespace, oldControllerRef); set != nil {
			ssc.enqueueStatefulSetPlus(set)
		}
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		set := ssc.resolveControllerRef(curPod.Namespace, curControllerRef)
		if set == nil {
			return
		}
		glog.V(4).Infof("Pod %s updated, objectMeta %+v -> %+v.", curPod.Name, oldPod.ObjectMeta, curPod.ObjectMeta)
		ssc.enqueueStatefulSetPlus(set)
		return
	}

	// Otherwise, it's an orphan. If anything changed, sync matching controllers
	// to see if anyone wants to adopt it now.
	if labelChanged || controllerRefChanged {
		sets := ssc.getStatefulSetPlusesForPod(curPod)
		if len(sets) == 0 {
			glog.V(4).Infof("Pod %s/%s is orphan in updated, but not controlled by StatefulSetPlus-Operator", curPod.Namespace, curPod.Name)
			return
		}
		glog.Infof("Orphan Pod %s/%s updated, objectMeta %+v -> %+v.", curPod.Namespace, curPod.Name, oldPod.ObjectMeta, curPod.ObjectMeta)
		for _, set := range sets {
			ssc.enqueueStatefulSetPlus(set)
		}
	}
}

// deletePod enqueues the statefulset for the pod accounting for deletion tombstones.
func (ssc *StatefulSetPlusController) deletePod(obj interface{}) {
	pod, ok := obj.(*v1.Pod)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value. Note that this value might be stale. If the pod
	// changed labels the new StatefulSetPlus will not be woken up till the periodic resync.
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
	glog.V(3).Infof("Pod %s/%s deleted through %v.", pod.Namespace, pod.Name, utilruntime.GetCaller())
	ssc.enqueueStatefulSetPlus(set)
}

// getPodsForStatefulSetPlus returns the Pods that a given StatefulSetPlus should manage.
// It also reconciles ControllerRef by adopting/orphaning.
//
// NOTE: Returned Pods are pointers to objects from the cache.
//       If you need to modify one, you need to copy it first.
func (ssc *StatefulSetPlusController) getPodsForStatefulSetPlus(set *stsplus.StatefulSetPlus, selector labels.Selector) ([]*v1.Pod, error) {
	// List all pods to include the pods that don't match the selector anymore but
	// has a ControllerRef pointing to this StatefulSetPlus.
	pods, err := ssc.podLister.Pods(set.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	filter := func(pod *v1.Pod) bool {
		// Only claim if it matches our StatefulSetPlus name. Otherwise release/ignore.
		return isMemberOf(set, pod)
	}

	// If any adoptions are attempted, we should first recheck for deletion with
	// an uncached quorum read sometime after listing Pods (see #42639).
	canAdoptFunc := controller.RecheckDeletionTimestamp(func() (metav1.Object, error) {
		fresh, err := ssc.tkexClient.TkexV1alpha1().StatefulSetPluses(set.Namespace).Get(set.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		if fresh.UID != set.UID {
			return nil, fmt.Errorf("original StatefulSetPlus %v/%v is gone: got uid %v, wanted %v", set.Namespace, set.Name, fresh.UID, set.UID)
		}
		return fresh, nil
	})

	cm := controller.NewPodControllerRefManager(ssc.podControl, set, selector, controllerKind, canAdoptFunc)
	return cm.ClaimPods(pods, filter)
}

// adoptOrphanRevisions adopts any orphaned ControllerRevisions matched by set's Selector.
func (ssc *StatefulSetPlusController) adoptOrphanRevisions(set *stsplus.StatefulSetPlus) error {
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
		fresh, err := ssc.tkexClient.TkexV1alpha1().StatefulSetPluses(set.Namespace).Get(set.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if fresh.UID != set.UID {
			return fmt.Errorf("original StatefulSetPlus %v/%v is gone: got uid %v, wanted %v", set.Namespace, set.Name, fresh.UID, set.UID)
		}
		return ssc.control.AdoptOrphanRevisions(set, revisions)
	}
	return nil
}

// getStatefulSetPlusesForPod returns a list of StatefulSetPluses that potentially match
// a given pod.
func (ssc *StatefulSetPlusController) getStatefulSetPlusesForPod(pod *v1.Pod) []*stsplus.StatefulSetPlus {
	sets, err := GetPodStatefulSetPluses(pod, ssc.setLister)
	if err != nil {
		return nil
	}
	// More than one set is selecting the same Pod
	if len(sets) > 1 {
		// ControllerRef will ensure we don't do anything crazy, but more than one
		// item in this list nevertheless constitutes user error.
		utilruntime.HandleError(
			fmt.Errorf(
				"user error: more than one StatefulSetPlus is selecting pods with labels: %+v",
				pod.Labels))
	}
	return sets
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the correct Kind.
func (ssc *StatefulSetPlusController) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *stsplus.StatefulSetPlus {
	// We can't look up by UID, so look up by Name and then verify UID.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		return nil
	}
	set, err := ssc.setLister.StatefulSetPluses(namespace).Get(controllerRef.Name)
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

// enqueueStatefulSetPlus enqueues the given statefulset in the work queue.
func (ssc *StatefulSetPlusController) enqueueStatefulSetPlus(obj interface{}) {
	key, err := controller.KeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Cound't get key for object %+v: %v", obj, err))
		return
	}
	glog.V(4).Infof("enqueueStatefulSetPlus enqueue item: %s", key)
	ssc.queue.Add(key)
}

// processNextWorkItem dequeues items, processes them, and marks them done. It enforces that the syncHandler is never
// invoked concurrently with the same key.
func (ssc *StatefulSetPlusController) processNextWorkItem() bool {
	key, quit := ssc.queue.Get()
	if quit {
		return false
	}
	defer ssc.queue.Done(key)
	glog.Infof("processNextWorkItem get item: %#v", key)
	if err := ssc.sync(key.(string)); err != nil {
		utilruntime.HandleError(fmt.Errorf("Error syncing StatefulSetPlus %v, requeuing: %v", key.(string), err))
		ssc.queue.AddRateLimited(key)
	} else {
		ssc.queue.Forget(key)
	}
	return true
}

// worker runs a worker goroutine that invokes processNextWorkItem until the controller's queue is closed
func (ssc *StatefulSetPlusController) worker() {
	for ssc.processNextWorkItem() {
	}
}

// sync syncs the given statefulset.
func (ssc *StatefulSetPlusController) sync(key string) error {
	startTime := time.Now()
	defer func() {
		glog.V(3).Infof("Finished syncing statefulsetplus %q (%v)", key, time.Since(startTime))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	set, err := ssc.setLister.StatefulSetPluses(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.Infof("StatefulSetPlus %s has been deleted", key)
		return nil
	}
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to retrieve StatefulSetPlus %v from store: %v", key, err))
		return err
	}

	selector, err := metav1.LabelSelectorAsSelector(set.Spec.Selector)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("error converting StatefulSetPlus %v selector: %v", key, err))
		// This is a non-transient error, so don't retry.
		return nil
	}

	if err := ssc.adoptOrphanRevisions(set); err != nil {
		return err
	}

	pods, err := ssc.getPodsForStatefulSetPlus(set, selector)
	if err != nil {
		return err
	}

	return ssc.syncStatefulSetPlus(set, pods)
}

// syncStatefulSetPlus syncs a tuple of (statefulset, []*v1.Pod).
func (ssc *StatefulSetPlusController) syncStatefulSetPlus(set *stsplus.StatefulSetPlus, pods []*v1.Pod) error {
	//glog.Infof("Syncing StatefulSetPlus %s/%s with %d pods", set.Namespace, set.Name, len(pods))
	// TODO: investigate where we mutate the set during the update as it is not obvious.
	if err := ssc.control.UpdateStatefulSetPlus(set.DeepCopy(), pods); err != nil {
		return err
	}
	glog.Infof("Successfully synced StatefulSetPlus %s/%s successful, pod length: %d", set.Namespace, set.Name, len(pods))
	return nil
}
