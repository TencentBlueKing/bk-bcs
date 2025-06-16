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

// Package scaler xxx
package scaler

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	autoscalinginternal "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	scaleclient "k8s.io/client-go/scale"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/controller"

	autoscaling "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/apis/autoscaling/v1alpha1"
	autoscalingscheme "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/client/clientset/versioned/scheme"
	autoscalingclient "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/client/clientset/versioned/typed/autoscaling/v1alpha1"
	autoscalinginformers "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/client/informers/externalversions/autoscaling/v1alpha1"
	autoscalinglisters "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/client/listers/autoscaling/v1alpha1"
	metricsclient "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/monitor"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/scalercore"
)

var (
	scaleUpLimitFactor  = 2.0
	scaleUpLimitMinimum = 4.0
	computeByLimitsKey  = "compute-by-limits"
	metricsServer       monitor.PrometheusMetricServer
)

type timestampedRecommendation struct {
	recommendation int32
	timestamp      time.Time
}

type timestampedScaleEvent struct {
	replicaChange int32 // positive for scaleUp, negative for scaleDown
	timestamp     time.Time
	outdated      bool
}

// GeneralController is responsible for the synchronizing GPA objects stored
// in the system with the actual deployments/replication controllers they
// control.
type GeneralController struct {
	scaleNamespacer scaleclient.ScalesGetter
	gpaNamespacer   autoscalingclient.GeneralPodAutoscalersGetter
	mapper          apimeta.RESTMapper

	replicaCalc   *ReplicaCalculator
	eventRecorder record.EventRecorder

	downscaleStabilisationWindow time.Duration

	// gpaLister is able to list/get GPAs from the shared cache from the informer passed in to
	// NewGeneralController.
	gpaLister       autoscalinglisters.GeneralPodAutoscalerLister
	gpaListerSynced cache.InformerSynced

	// podLister is able to list/get Pods from the shared cache from the informer passed in to
	// NewGeneralController.
	podLister       corelisters.PodLister
	podListerSynced cache.InformerSynced

	// Controllers that need to be synced
	queue workqueue.RateLimitingInterface

	// Latest unstabilized recommendations for each autoscaler.
	recommendations map[string][]timestampedRecommendation

	// Multi goroutine read and write recommendations may unsafe.
	recommendationsLock sync.Mutex

	// Latest autoscaler events
	scaleUpEvents   map[string][]timestampedScaleEvent
	scaleDownEvents map[string][]timestampedScaleEvent

	// Multi goroutine read and write scaleUp/scaleDown events may unsafe.
	scaleUpEventsLock   sync.RWMutex
	scaleDownEventsLock sync.RWMutex

	doingCron sync.Map // nolint

	// Multi goroutines for autoscaler
	workers int
}

// NewGeneralController creates a new GeneralController.
func NewGeneralController(
	evtNamespacer v1core.EventsGetter,
	scaleNamespacer scaleclient.ScalesGetter,
	gpaNamespacer autoscalingclient.GeneralPodAutoscalersGetter,
	mapper apimeta.RESTMapper,
	metricsClient metricsclient.MetricsClient,
	gpaInformer autoscalinginformers.GeneralPodAutoscalerInformer,
	podInformer coreinformers.PodInformer,
	resyncPeriod time.Duration,
	downscaleStabilisationWindow time.Duration,
	tolerance float64,
	cpuInitializationPeriod,
	delayOfInitialReadinessStatus time.Duration,
	workers int,

) *GeneralController {
	_ = autoscalingscheme.AddToScheme(scheme.Scheme)
	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(klog.Infof)
	broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: evtNamespacer.Events(v1.NamespaceAll)})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "pod-autoscaler"})

	gpaController := &GeneralController{
		eventRecorder:                recorder,
		scaleNamespacer:              scaleNamespacer,
		gpaNamespacer:                gpaNamespacer,
		downscaleStabilisationWindow: downscaleStabilisationWindow,
		queue: workqueue.NewNamedRateLimitingQueue(
			NewDefaultGPARateLimiter(resyncPeriod), "podautoscaler"),
		mapper:          mapper,
		recommendations: map[string][]timestampedRecommendation{},
		scaleUpEvents:   map[string][]timestampedScaleEvent{},
		scaleDownEvents: map[string][]timestampedScaleEvent{},
		workers:         workers,
	}

	gpaInformer.Informer().AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    gpaController.enqueueGPA,
			UpdateFunc: gpaController.updateGPA,
			DeleteFunc: gpaController.deleteGPA,
		},
		resyncPeriod,
	)
	gpaController.gpaLister = gpaInformer.Lister()
	gpaController.gpaListerSynced = gpaInformer.Informer().HasSynced

	gpaController.podLister = podInformer.Lister()
	gpaController.podListerSynced = podInformer.Informer().HasSynced

	replicaCalc := NewReplicaCalculator(
		metricsClient,
		gpaController.podLister,
		tolerance,
		cpuInitializationPeriod,
		delayOfInitialReadinessStatus,
	)
	gpaController.replicaCalc = replicaCalc

	return gpaController
}

// Run begins watching and syncing.
func (a *GeneralController) Run(ctx context.Context) {
	defer utilruntime.HandleCrash()
	defer a.queue.ShutDown()

	klog.Infof("Starting GPA controller")
	defer klog.Infof("Shutting down GPA controller")

	if !cache.WaitForNamedCacheSync("GPA", ctx.Done(), a.gpaListerSynced, a.podListerSynced) {
		return
	}

	// start a single worker (we may wish to start more in the future)
	// go wait.Until(a.worker, time.Second, stopCh)

	// Launch workers to process gpa
	for i := 0; i < a.workers; i++ {
		go wait.UntilWithContext(ctx, a.worker, time.Second)
	}

	<-ctx.Done()
}

// updateGPA obj could be an *v1.GeneralPodAutoscaler, or a DeletionFinalStateUnknown marker item.
func (a *GeneralController) updateGPA(old, cur interface{}) {
	a.enqueueGPA(cur)
}

// enqueueGPA obj could be an *v1.GeneralPodAutoscaler, or a DeletionFinalStateUnknown marker item.
func (a *GeneralController) enqueueGPA(obj interface{}) {
	key, err := controller.KeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %+v: %v", obj, err))
		return
	}

	// Requests are always added to queue with resyncPeriod delay.  If there's already
	// request for the GPA in the queue then a new request is always dropped. Requests spend resync
	// interval in queue so GPAs are processed every resync interval.
	a.queue.AddRateLimited(key)
}

func (a *GeneralController) deleteGPA(obj interface{}) {
	key, err := controller.KeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %+v: %v", obj, err))
		return
	}

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get namespace/name for key %s: %v", key, err))
		return
	}
	metricsServer.ResetScalerMetrics(namespace, name)

	a.queue.Forget(key)
}

func (a *GeneralController) worker(ctx context.Context) {
	for a.processNextWorkItem(ctx) {
	}
	klog.Infof("general pod autoscaler controller worker shutting down")
}

func (a *GeneralController) processNextWorkItem(ctx context.Context) bool {
	key, quit := a.queue.Get()
	if quit {
		return false
	}
	defer a.queue.Done(key)

	deleted, err := a.reconcileKey(ctx, key.(string))
	if err != nil {
		utilruntime.HandleError(err)
	}
	// Add request processing GPA to queue with resyncPeriod delay.
	// Requests are always added to queue with resyncPeriod delay. If there's already request
	// for the GPA in the queue then a new request is always dropped. Requests spend resyncPeriod
	// in queue so GPAs are processed every resyncPeriod.
	// Request is added here just in case last resync didn't insert request into the queue. This
	// happens quite often because there is race condition between adding request after resyncPeriod
	// and removing them from queue. Request can be added by resync before previous request is
	// removed from queue. If we didn't add request here then in this case one request would be dropped
	// and GPA would processed after 2 x resyncPeriod.
	if !deleted {
		a.queue.AddRateLimited(key)
	}
	return true
}

func (a *GeneralController) reconcileKey(ctx context.Context, key string) (deleted bool, err error) {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return true, err
	}

	gpa, err := a.gpaLister.GeneralPodAutoscalers(namespace).Get(name)
	if kerrors.IsNotFound(err) {
		klog.Infof("General Pod Autoscaler %s has been deleted", klog.KRef(namespace, name))

		a.recommendationsLock.Lock()
		delete(a.recommendations, key)
		a.recommendationsLock.Unlock()

		a.scaleUpEventsLock.Lock()
		delete(a.scaleUpEvents, key)
		a.scaleUpEventsLock.Unlock()

		a.scaleDownEventsLock.Lock()
		delete(a.scaleDownEvents, key)
		a.scaleDownEventsLock.Unlock()

		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, a.reconcileAutoscaler(ctx, gpa, key)
}

func (a *GeneralController) recordInitialRecommendation(currentReplicas int32, key string) {
	// add lock
	a.recommendationsLock.Lock()
	defer a.recommendationsLock.Unlock()

	if a.recommendations[key] == nil {
		a.recommendations[key] = []timestampedRecommendation{{currentReplicas, time.Now()}}
	}
}

// reconcileAutoscaler TODO
// nolint funlen
func (a *GeneralController) reconcileAutoscaler(ctx context.Context, gpa *autoscaling.GeneralPodAutoscaler,
	key string) (retErr error) {
	// set default value, call Default() function will invoke scheme's defaulterFuncs
	scheme.Scheme.Default(gpa)
	klog.Infof("Reconciling %s", klog.KRef(gpa.Namespace, gpa.Name))

	// make a copy so that we never mutate the shared informer cache (conversion can mutate the object)
	gpaStatusOriginal := gpa.Status.DeepCopy()

	reference := fmt.Sprintf("%s/%s/%s", gpa.Spec.ScaleTargetRef.Kind, gpa.Namespace, gpa.Spec.ScaleTargetRef.Name)

	targetGV, err := schema.ParseGroupVersion(gpa.Spec.ScaleTargetRef.APIVersion)
	if err != nil {
		a.eventRecorder.Event(gpa, v1.EventTypeWarning, "FailedGetScale", err.Error())
		setCondition(gpa, autoscaling.AbleToScale, v1.ConditionFalse, "FailedGetScale",
			"the GPA controller was unable to get the target's current scale: %v", err)
		if updateErr := a.updateStatusIfNeeded(ctx, gpaStatusOriginal, gpa); updateErr != nil {
			utilruntime.HandleError(updateErr)
		}
		return fmt.Errorf("invalid API version in scale target reference: %v", err)
	}

	targetGK := schema.GroupKind{
		Group: targetGV.Group,
		Kind:  gpa.Spec.ScaleTargetRef.Kind,
	}

	mappings, err := a.mapper.RESTMappings(targetGK)
	if err != nil {
		a.eventRecorder.Event(gpa, v1.EventTypeWarning, "FailedGetScale", err.Error())
		setCondition(gpa, autoscaling.AbleToScale, v1.ConditionFalse, "FailedGetScale",
			"the GPA controller was unable to get the target's current scale: %v", err)
		if updateErr := a.updateStatusIfNeeded(ctx, gpaStatusOriginal, gpa); updateErr != nil {
			utilruntime.HandleError(updateErr)
		}
		return fmt.Errorf("unable to determine resource for scale target reference: %v", err)
	}

	scale, targetGR, err := a.scaleForResourceMappings(ctx, gpa.Namespace, gpa.Spec.ScaleTargetRef.Name, mappings)
	if err != nil {
		a.eventRecorder.Event(gpa, v1.EventTypeWarning, "FailedGetScale", err.Error())
		setCondition(gpa, autoscaling.AbleToScale, v1.ConditionFalse, "FailedGetScale",
			"the GPA controller was unable to get the target's current scale: %v", err)
		if updateErr := a.updateStatusIfNeeded(ctx, gpaStatusOriginal, gpa); updateErr != nil {
			klog.Error(updateErr)
		}
		return fmt.Errorf("failed to query scale subresource for %s: %v", reference, err)
	}
	setCondition(gpa, autoscaling.AbleToScale, v1.ConditionTrue, "SucceededGetScale",
		"the GPA controller was able to get the target's current scale")
	currentReplicas := scale.Spec.Replicas
	a.recordInitialRecommendation(currentReplicas, key)

	var (
		metricStatuses        []autoscaling.MetricStatus
		metricDesiredReplicas int32
		metricName            string
	)

	desiredReplicas := int32(0)
	rescaleReason := ""

	var minReplicas int32

	if gpa.Spec.MinReplicas != nil {
		minReplicas = *gpa.Spec.MinReplicas
	} else {
		// Default value
		minReplicas = 1
	}

	rescale := true
	// nolint
	if disableRescale(currentReplicas, minReplicas) {
		// Autoscaling is disabled for this resource
		desiredReplicas = 0
		rescale = false
		setCondition(gpa, autoscaling.ScalingActive, v1.ConditionFalse, "ScalingDisabled",
			"scaling is disabled since the replica count of the target is zero")
	} else if currentReplicas > gpa.Spec.MaxReplicas {
		rescaleReason = "Current number of replicas above Spec.MaxReplicas"
		desiredReplicas = gpa.Spec.MaxReplicas
	} else if currentReplicas < minReplicas {
		rescaleReason = "Current number of replicas below Spec.MinReplicas"
		desiredReplicas = minReplicas
	} else {
		var metricTimestamp time.Time
		if isEmpty(gpa.Spec.AutoScalingDrivenMode) {
			return nil
		}
		metricDesiredReplicas = -1
		metricDesiredReplicas, metricName, metricStatuses, metricTimestamp, err = a.computeReplicas(ctx, gpa,
			scale, currentReplicas, gpaStatusOriginal)
		if err != nil && metricDesiredReplicas == -1 {
			// computeReplicasForMetrics may return both non-zero metricDesiredReplicas and an error.
			// That means some metrics still work and GPA should perform scaling based on them.
			klog.Errorf("computeReplicas failed: %v", err)
			return err
		}
		if err != nil {
			// We proceed to scaling, but return this error from reconcileAutoscaler() finally.
			retErr = err
		}
		// if all mode can not give a valid replicas, use scale spec replicas
		if metricDesiredReplicas == -1 {
			metricDesiredReplicas = scale.Spec.Replicas
		}

		klog.V(4).Infof("All-Mode: proposing %v desired replicas (based on %s from %s) for %s",
			metricDesiredReplicas, metricName, metricTimestamp, reference)
		desiredReplicas, rescaleReason = a.getDesiredReplicas(gpa, key, metricDesiredReplicas, desiredReplicas,
			currentReplicas, minReplicas, metricName)
		klog.V(4).Infof("After normalizing, the replicas is %d", desiredReplicas)
		rescale = desiredReplicas != currentReplicas
	}
	metricsServer.RecordGPAReplicas(gpa, minReplicas, desiredReplicas)

	if rescale {
		scale.Spec.Replicas = desiredReplicas
		startTime := time.Now()
		_, err = a.scaleNamespacer.Scales(gpa.Namespace).Update(ctx, targetGR, scale, metav1.UpdateOptions{})
		if err != nil {
			metricsServer.RecordScalerUpdateDuration(gpa, "failure", time.Since(startTime))
			metricsServer.RecordScalerReplicasUpdateDuration(gpa, "failure", time.Since(startTime))
			a.eventRecorder.Eventf(gpa, v1.EventTypeWarning, "FailedRescale",
				"New size: %d; reason: %s; error: %v", desiredReplicas, rescaleReason, err.Error())
			setCondition(gpa, autoscaling.AbleToScale, v1.ConditionFalse, "FailedUpdateScale",
				"the GPA controller was unable to update the target scale: %v", err)
			a.setCurrentReplicasInStatus(gpa, currentReplicas)
			if updateErr := a.updateStatusIfNeeded(ctx, gpaStatusOriginal, gpa); updateErr != nil {
				utilruntime.HandleError(updateErr)
			}
			return fmt.Errorf("failed to rescale %s: %v", reference, err)
		}
		metricsServer.RecordScalerUpdateDuration(gpa, "success", time.Since(startTime))
		metricsServer.RecordScalerReplicasUpdateDuration(gpa, "success", time.Since(startTime))
		setCondition(gpa, autoscaling.AbleToScale, v1.ConditionTrue, "SucceededRescale",
			"the GPA controller was able to update the target scale to %d", desiredReplicas)
		a.eventRecorder.Eventf(gpa, v1.EventTypeNormal, "SuccessfulRescale", "New size: %d; reason: %s",
			desiredReplicas, rescaleReason)
		a.storeScaleEvent(gpa.Spec.Behavior, key, currentReplicas, desiredReplicas)
		klog.Infof("Successful rescale of %s, old size: %d, new size: %d, reason: %s",
			gpa.Name, currentReplicas, desiredReplicas, rescaleReason)
	} else {
		klog.V(4).Infof("decided not to scale %s to %v (last scale time was %s)",
			reference, desiredReplicas, gpa.Status.LastScaleTime)
		desiredReplicas = currentReplicas
	}
	a.setStatus(gpa, currentReplicas, desiredReplicas, metricStatuses, rescale)
	err = a.updateStatusIfNeeded(ctx, gpaStatusOriginal, gpa)
	if err != nil {
		return err
	}
	return retErr
}

func (a *GeneralController) computeReplicas(ctx context.Context, gpa *autoscaling.GeneralPodAutoscaler,
	scale *autoscalinginternal.Scale, currentReplicas int32,
	gpaStatusOriginal *autoscaling.GeneralPodAutoscalerStatus) (int32, string, []autoscaling.MetricStatus, time.Time,
	error) {
	reference := fmt.Sprintf("%s/%s/%s", gpa.Spec.ScaleTargetRef.Kind, gpa.Namespace, gpa.Spec.ScaleTargetRef.Name)
	results := make([]result, 0)
	errs := make([]error, 0)
	if gpa.Spec.MetricMode != nil {
		// get replicas from metric mode
		metricDesiredReplicas, metricName, metricStatuses, metricTimestamp, metricErr := a.computeReplicasForMetrics(ctx, gpa,
			scale, gpa.Spec.MetricMode.Metrics)
		if metricErr != nil {
			errs = append(errs, metricErr)
		}
		if metricDesiredReplicas != -1 {
			results = append(results, result{
				metricDesiredReplicas, metricName, metricStatuses, metricTimestamp, gpa.Spec.MetricMode.Proirity,
			})
		}
		klog.V(4).Infof("Metric-Mode: proposing %v desired replicas (based on %s from %s) for %s (priority: %d)",
			metricDesiredReplicas, metricName, metricTimestamp, reference, gpa.Spec.MetricMode.Proirity)
	}
	if gpa.Spec.WebhookMode != nil {
		wbDesiredReplicas, wbName, wbStatuses, wbTimestamp, wbErr := a.computeReplicasForSimple(gpa,
			scale, scalercore.NewWebhookScaler(gpa.Spec.WebhookMode))
		if wbErr != nil {
			errs = append(errs, wbErr)
		}
		if wbDesiredReplicas != -1 {
			results = append(results, result{
				wbDesiredReplicas, wbName, wbStatuses, wbTimestamp, gpa.Spec.WebhookMode.Proirity,
			})
		}
		klog.V(4).Infof("Webhook-Mode: proposing %v desired replicas (based on %s from %s) for %s (priority: %d)",
			wbDesiredReplicas, wbName, wbTimestamp, reference, gpa.Spec.WebhookMode.Proirity)
	}
	if gpa.Spec.TimeMode != nil {
		tmDesiredReplicas, tmName, tmStatuses, tmTimestamp, tmErr := a.computeReplicasForSimple(gpa,
			scale, scalercore.NewCronScaler(gpa.Spec.TimeMode.TimeRanges))
		if tmErr != nil {
			errs = append(errs, tmErr)
		}
		if tmDesiredReplicas != -1 {
			results = append(results, result{
				tmDesiredReplicas, tmName, tmStatuses, tmTimestamp, gpa.Spec.TimeMode.Proirity,
			})
		}
		klog.V(4).Infof("Time-Mode: proposing %v desired replicas (based on %s from %s) for %s (priority: %d)",
			tmDesiredReplicas, tmName, tmTimestamp, reference, gpa.Spec.TimeMode.Proirity)
	}
	var err error
	if len(errs) > 0 {
		err = errors.Join(errs...)
	}
	// computeReplicas may return both non-zero desiredReplicas and errors.
	// That means some metrics still work and HPA should perform scaling based on them.
	if err != nil && len(results) == 0 {
		a.setCurrentReplicasInStatus(gpa, currentReplicas)
		if updateErr := a.updateStatusIfNeeded(ctx, gpaStatusOriginal, gpa); updateErr != nil {
			utilruntime.HandleError(updateErr)
		}
		a.eventRecorder.Event(gpa, v1.EventTypeWarning, "FailedComputeMetricsReplicas", err.Error())
		return -1, "", nil, time.Now(), err
	}
	if err != nil {
		// We proceed to scaling, but return this error from reconcileAutoscaler() finally.
		klog.Warning(err)
	}
	if len(results) == 0 {
		// return if no mode work
		return -1, "", nil, time.Now(), err
	}
	// sort results by priority
	sortResults(results)
	for _, result := range results {
		// if the result with high proirity, but replicas is -1, should not return it
		if result.replicas == -1 {
			continue
		}
		return result.replicas, result.metric, result.statuses, result.timestamp, err
	}
	// should not arrive here
	return -1, "", nil, time.Now(), err
}

// stabilizeRecommendation :
// - replaces old recommendation with the newest recommendation,
// - returns max of recommendations that are not older than downscaleStabilisationWindow.
func (a *GeneralController) stabilizeRecommendation(key string, prenormalizedDesiredReplicas int32) int32 {
	// add lock
	a.recommendationsLock.Lock()
	defer a.recommendationsLock.Unlock()

	maxRecommendation := prenormalizedDesiredReplicas
	foundOldSample := false
	oldSampleIndex := 0
	cutoff := time.Now().Add(-a.downscaleStabilisationWindow)
	for i, rec := range a.recommendations[key] {
		if rec.timestamp.Before(cutoff) {
			foundOldSample = true
			oldSampleIndex = i
		} else if rec.recommendation > maxRecommendation {
			maxRecommendation = rec.recommendation
		}
	}
	if foundOldSample {
		a.recommendations[key][oldSampleIndex] = timestampedRecommendation{
			prenormalizedDesiredReplicas, time.Now()}
	} else {
		a.recommendations[key] = append(a.recommendations[key], timestampedRecommendation{
			prenormalizedDesiredReplicas, time.Now()})
	}
	return maxRecommendation
}

// normalizeDesiredReplicas takes the metrics desired replicas value and normalizes it based on
// the appropriate conditions (i.e. < maxReplicas, > minReplicas, etc...)
func (a *GeneralController) normalizeDesiredReplicas(gpa *autoscaling.GeneralPodAutoscaler,
	key string, currentReplicas int32, prenormalizedDesiredReplicas int32, minReplicas int32) int32 {
	stabilizedRecommendation := a.stabilizeRecommendation(key, prenormalizedDesiredReplicas)
	if stabilizedRecommendation != prenormalizedDesiredReplicas {
		setCondition(gpa, autoscaling.AbleToScale, v1.ConditionTrue, "ScaleDownStabilized",
			"recent recommendations were higher than current one, applying the highest recent recommendation")
	} else {
		setCondition(gpa, autoscaling.AbleToScale, v1.ConditionTrue, "ReadyForNewScale",
			"recommended size matches current size")
	}

	desiredReplicas, condition, reason := convertDesiredReplicasWithRules(currentReplicas,
		stabilizedRecommendation, minReplicas, gpa.Spec.MaxReplicas)

	if desiredReplicas == stabilizedRecommendation {
		setCondition(gpa, autoscaling.ScalingLimited, v1.ConditionFalse, condition, reason)
	} else {
		setCondition(gpa, autoscaling.ScalingLimited, v1.ConditionTrue, condition, reason)
	}

	return desiredReplicas
}

// NormalizationArg is used to pass all needed information between functions as one structure
type NormalizationArg struct {
	Key               string
	ScaleUpBehavior   *autoscaling.GPAScalingRules
	ScaleDownBehavior *autoscaling.GPAScalingRules
	MinReplicas       int32
	MaxReplicas       int32
	CurrentReplicas   int32
	DesiredReplicas   int32
}

// normalizeDesiredReplicasWithBehaviors takes the metrics desired replicas value and normalizes it:
//  1. Apply the basic conditions (i.e. < maxReplicas, > minReplicas, etc...)
//  2. Apply the scale up/down limits from the gpaSpec.Behaviors (i.e. add no more than 4 pods)
//  3. Apply the constraints period (i.e. add no more than 4 pods per minute)
//  4. Apply the stabilization (i.e. add no more than 4 pods per minute, and pick the smallest
//     recommendation during last 5 minutes)
//
// NOCC:tosa/fn_length(设计如此)
func (a *GeneralController) normalizeDesiredReplicasWithBehaviors(gpa *autoscaling.GeneralPodAutoscaler,
	key string, currentReplicas, prenormalizedDesiredReplicas, minReplicas int32) int32 {
	a.maybeInitScaleDownStabilizationWindow(gpa)
	normalizationArg := NormalizationArg{
		Key:               key,
		ScaleUpBehavior:   gpa.Spec.Behavior.ScaleUp,
		ScaleDownBehavior: gpa.Spec.Behavior.ScaleDown,
		MinReplicas:       minReplicas,
		MaxReplicas:       gpa.Spec.MaxReplicas,
		CurrentReplicas:   currentReplicas,
		DesiredReplicas:   prenormalizedDesiredReplicas}
	stabilizedRecommendation, reason, message := a.stabilizeRecommendationWithBehaviors(normalizationArg)
	normalizationArg.DesiredReplicas = stabilizedRecommendation
	if stabilizedRecommendation != prenormalizedDesiredReplicas {
		// "ScaleUpStabilized" || "ScaleDownStabilized"
		setCondition(gpa, autoscaling.AbleToScale, v1.ConditionTrue, reason, message)
	} else {
		setCondition(gpa, autoscaling.AbleToScale, v1.ConditionTrue, "ReadyForNewScale",
			"recommended size matches current size")
	}
	desiredReplicas, reason, message := a.convertDesiredReplicasWithBehaviorRate(normalizationArg)
	if desiredReplicas == stabilizedRecommendation {
		setCondition(gpa, autoscaling.ScalingLimited, v1.ConditionFalse, reason, message)
	} else {
		setCondition(gpa, autoscaling.ScalingLimited, v1.ConditionTrue, reason, message)
	}

	return desiredReplicas
}

// maybeInitScaleDownStabilizationWindow TODO
// NOCC:tosa/fn_length(设计如此)
func (a *GeneralController) maybeInitScaleDownStabilizationWindow(gpa *autoscaling.GeneralPodAutoscaler) {
	behavior := gpa.Spec.Behavior
	if behavior != nil && behavior.ScaleDown != nil && behavior.ScaleDown.StabilizationWindowSeconds == nil {
		stabilizationWindowSeconds := (int32)(a.downscaleStabilisationWindow.Seconds())
		gpa.Spec.Behavior.ScaleDown.StabilizationWindowSeconds = &stabilizationWindowSeconds
	}
}

// getReplicasChangePerPeriod function find all the replica changes per period
func getReplicasChangePerPeriod(periodSeconds int32, scaleEvents []timestampedScaleEvent) int32 {
	period := time.Second * time.Duration(periodSeconds)
	cutoff := time.Now().Add(-period)
	var replicas int32
	for _, rec := range scaleEvents {
		if rec.timestamp.After(cutoff) {
			replicas += rec.replicaChange
		}
	}
	return replicas

}

// getUnableComputeReplicaCountCondition TODO
// NOCC:tosa/fn_length(设计如此)
func (a *GeneralController) getUnableComputeReplicaCountCondition(gpa *autoscaling.GeneralPodAutoscaler,
	reason string, err error) (condition autoscaling.GeneralPodAutoscalerCondition) {
	a.eventRecorder.Event(gpa, v1.EventTypeWarning, reason, err.Error())
	return autoscaling.GeneralPodAutoscalerCondition{
		Type:    autoscaling.ScalingActive,
		Status:  v1.ConditionFalse,
		Reason:  reason,
		Message: fmt.Sprintf("the GPA was unable to compute the replica count: %v", err),
	}
}

// storeScaleEvent stores (adds or replaces outdated) scale event.
// outdated events to be replaced were marked as outdated in the `markScaleEventsOutdated` function
func (a *GeneralController) storeScaleEvent(behavior *autoscaling.GeneralPodAutoscalerBehavior,
	key string, prevReplicas, newReplicas int32) {
	if behavior == nil {
		return // we should not store any event as they will not be used
	}

	var oldSampleIndex int
	var longestPolicyPeriod int32
	foundOldSample := false
	if newReplicas > prevReplicas {
		longestPolicyPeriod = getLongestPolicyPeriod(behavior.ScaleUp)

		a.scaleUpEventsLock.Lock()
		defer a.scaleUpEventsLock.Unlock()
		markScaleEventsOutdated(a.scaleUpEvents[key], longestPolicyPeriod)
		replicaChange := newReplicas - prevReplicas
		for i, event := range a.scaleUpEvents[key] {
			if event.outdated {
				foundOldSample = true
				oldSampleIndex = i
			}
		}
		newEvent := timestampedScaleEvent{replicaChange, time.Now(), false}
		if foundOldSample {
			a.scaleUpEvents[key][oldSampleIndex] = newEvent
		} else {
			a.scaleUpEvents[key] = append(a.scaleUpEvents[key], newEvent)
		}
	} else {
		longestPolicyPeriod = getLongestPolicyPeriod(behavior.ScaleDown)

		a.scaleDownEventsLock.Lock()
		defer a.scaleDownEventsLock.Unlock()
		markScaleEventsOutdated(a.scaleDownEvents[key], longestPolicyPeriod)
		replicaChange := prevReplicas - newReplicas
		for i, event := range a.scaleDownEvents[key] {
			if event.outdated {
				foundOldSample = true
				oldSampleIndex = i
			}
		}
		newEvent := timestampedScaleEvent{replicaChange, time.Now(), false}
		if foundOldSample {
			a.scaleDownEvents[key][oldSampleIndex] = newEvent
		} else {
			a.scaleDownEvents[key] = append(a.scaleDownEvents[key], newEvent)
		}
	}
}

// stabilizeRecommendationWithBehaviors :
// - replaces old recommendation with the newest recommendation,
// - returns {max,min} of recommendations that are not older than constraints.Scale{Up,Down}.DelaySeconds
// NOCC:tosa/fn_length(设计如此)
func (a *GeneralController) stabilizeRecommendationWithBehaviors(args NormalizationArg) (int32, string, string) {
	now := time.Now()

	foundOldSample := false
	oldSampleIndex := 0

	upRecommendation := args.DesiredReplicas
	upDelaySeconds := *args.ScaleUpBehavior.StabilizationWindowSeconds
	upCutoff := now.Add(-time.Second * time.Duration(upDelaySeconds))

	downRecommendation := args.DesiredReplicas
	downDelaySeconds := *args.ScaleDownBehavior.StabilizationWindowSeconds
	downCutoff := now.Add(-time.Second * time.Duration(downDelaySeconds))

	// Calculate the upper and lower stabilization limits.
	a.recommendationsLock.Lock()
	defer a.recommendationsLock.Unlock()
	for i, rec := range a.recommendations[args.Key] {
		if rec.timestamp.After(upCutoff) {
			upRecommendation = min(rec.recommendation, upRecommendation)
		}
		if rec.timestamp.After(downCutoff) {
			downRecommendation = max(rec.recommendation, downRecommendation)
		}
		if rec.timestamp.Before(upCutoff) && rec.timestamp.Before(downCutoff) {
			foundOldSample = true
			oldSampleIndex = i
		}
	}

	// Bring the recommendation to within the upper and lower limits (stabilize).
	recommendation := args.CurrentReplicas
	if recommendation < upRecommendation {
		recommendation = upRecommendation
	}
	if recommendation > downRecommendation {
		recommendation = downRecommendation
	}

	// Record the unstabilized recommendation.
	if foundOldSample {
		a.recommendations[args.Key][oldSampleIndex] = timestampedRecommendation{args.DesiredReplicas, time.Now()}
	} else {
		a.recommendations[args.Key] = append(a.recommendations[args.Key],
			timestampedRecommendation{args.DesiredReplicas, time.Now()})
	}

	// Determine a human-friendly message.
	var reason, message string
	if args.DesiredReplicas >= args.CurrentReplicas {
		reason = "ScaleUpStabilized"
		message = "recent recommendations were lower than current one, applying the lowest recent recommendation"
	} else {
		reason = "ScaleDownStabilized"
		message = "recent recommendations were higher than current one, applying the highest recent recommendation"
	}
	return recommendation, reason, message

}

// convertDesiredReplicasWithBehaviorRate performs the actual normalization, given the constraint rate
// It doesn't consider the stabilizationWindow, it is done separately
// NOCC:tosa/fn_length(设计如此)
func (a *GeneralController) convertDesiredReplicasWithBehaviorRate(args NormalizationArg) (int32, string, string) {
	var possibleLimitingReason, possibleLimitingMessage string

	if args.DesiredReplicas > args.CurrentReplicas {
		a.scaleUpEventsLock.RLock()
		defer a.scaleUpEventsLock.RUnlock()
		a.scaleDownEventsLock.RLock()
		defer a.scaleDownEventsLock.RUnlock()
		scaleUpLimit := calculateScaleUpLimitWithScalingRules(args.CurrentReplicas,
			a.scaleUpEvents[args.Key], a.scaleDownEvents[args.Key], args.ScaleUpBehavior)
		if scaleUpLimit < args.CurrentReplicas {
			// We shouldn't scale up further until the scaleUpEvents will be cleaned up
			scaleUpLimit = args.CurrentReplicas
		}
		maximumAllowedReplicas := args.MaxReplicas
		if maximumAllowedReplicas > scaleUpLimit {
			maximumAllowedReplicas = scaleUpLimit
			possibleLimitingReason = "ScaleUpLimit"
			possibleLimitingMessage = "the desired replica count is increasing faster than the maximum scale rate"
		} else {
			possibleLimitingReason = "TooManyReplicas"
			possibleLimitingMessage = "the desired replica count is more than the maximum replica count"
		}
		if args.DesiredReplicas > maximumAllowedReplicas {
			return maximumAllowedReplicas, possibleLimitingReason, possibleLimitingMessage
		}
	} else if args.DesiredReplicas < args.CurrentReplicas {
		a.scaleUpEventsLock.RLock()
		defer a.scaleUpEventsLock.RUnlock()
		a.scaleDownEventsLock.RLock()
		defer a.scaleDownEventsLock.RUnlock()
		scaleDownLimit := calculateScaleDownLimitWithBehaviors(args.CurrentReplicas,
			a.scaleUpEvents[args.Key], a.scaleDownEvents[args.Key], args.ScaleDownBehavior)
		if scaleDownLimit > args.CurrentReplicas {
			// We shouldn't scale down further until the scaleDownEvents will be cleaned up
			scaleDownLimit = args.CurrentReplicas
		}
		minimumAllowedReplicas := args.MinReplicas
		if minimumAllowedReplicas < scaleDownLimit {
			minimumAllowedReplicas = scaleDownLimit
			possibleLimitingReason = "ScaleDownLimit"
			possibleLimitingMessage = "the desired replica count is decreasing faster than the maximum scale rate"
		} else {
			possibleLimitingMessage = "the desired replica count is less than the minimum replica count"
			possibleLimitingReason = "TooFewReplicas"
		}
		if args.DesiredReplicas < minimumAllowedReplicas {
			return minimumAllowedReplicas, possibleLimitingReason, possibleLimitingMessage
		}
	}
	return args.DesiredReplicas, "DesiredWithinRange", "the desired count is within the acceptable range"
}

// convertDesiredReplicasWithRules performs the actual normalization,
// without depending on `GeneralController` or `GeneralPodAutoscaler`
func convertDesiredReplicasWithRules(currentReplicas, desiredReplicas,
	gpaMinReplicas, gpaMaxReplicas int32) (int32, string, string) {
	var minimumAllowedReplicas int32
	var maximumAllowedReplicas int32

	var possibleLimitingCondition string
	var possibleLimitingReason string

	minimumAllowedReplicas = gpaMinReplicas

	// Do not upscale too much to prevent incorrect rapid increase of the number of master replicas caused by
	// bogus CPU usage report from heapster/kubelet (like in issue #32304).
	scaleUpLimit := calculateScaleUpLimit(currentReplicas)

	if gpaMaxReplicas > scaleUpLimit {
		maximumAllowedReplicas = scaleUpLimit
		possibleLimitingCondition = "ScaleUpLimit"
		possibleLimitingReason = "the desired replica count is increasing faster than the maximum scale rate"
	} else {
		maximumAllowedReplicas = gpaMaxReplicas
		possibleLimitingCondition = "TooManyReplicas"
		possibleLimitingReason = "the desired replica count is more than the maximum replica count"
	}

	if desiredReplicas < minimumAllowedReplicas {
		possibleLimitingCondition = "TooFewReplicas"
		possibleLimitingReason = "the desired replica count is less than the minimum replica count"

		return minimumAllowedReplicas, possibleLimitingCondition, possibleLimitingReason
	} else if desiredReplicas > maximumAllowedReplicas {
		return maximumAllowedReplicas, possibleLimitingCondition, possibleLimitingReason
	}

	return desiredReplicas, "DesiredWithinRange", "the desired count is within the acceptable range"
}

func calculateScaleUpLimit(currentReplicas int32) int32 {
	return int32(math.Max(scaleUpLimitFactor*float64(currentReplicas), scaleUpLimitMinimum))
}

// markScaleEventsOutdated set 'outdated=true' flag for all scale events that are not used by any GPA object
func markScaleEventsOutdated(scaleEvents []timestampedScaleEvent, longestPolicyPeriod int32) {
	period := time.Second * time.Duration(longestPolicyPeriod)
	cutoff := time.Now().Add(-period)
	for i, event := range scaleEvents {
		if event.timestamp.Before(cutoff) {
			// outdated scale event are marked for later reuse
			scaleEvents[i].outdated = true
		}
	}
}

func getLongestPolicyPeriod(scalingRules *autoscaling.GPAScalingRules) int32 {
	var longestPolicyPeriod int32
	for _, policy := range scalingRules.Policies {
		if policy.PeriodSeconds > longestPolicyPeriod {
			longestPolicyPeriod = policy.PeriodSeconds
		}
	}
	return longestPolicyPeriod
}

// calculateScaleUpLimitWithScalingRules returns the maximum number of pods
// that could be added for the given GPAScalingRules
// NOCC:tosa/fn_length(设计如此)
func calculateScaleUpLimitWithScalingRules(currentReplicas int32,
	scaleUpEvents, scaleDownEvents []timestampedScaleEvent,
	scalingRules *autoscaling.GPAScalingRules) int32 {
	var result int32
	var proposed int32
	var selectPolicyFn func(int32, int32) int32
	if *scalingRules.SelectPolicy == autoscaling.DisabledPolicySelect {
		return currentReplicas // Scaling is disabled
	} else if *scalingRules.SelectPolicy == autoscaling.MinPolicySelect {
		selectPolicyFn = min // For scaling up, the lowest change ('min' policy) produces a minimum value
		result = math.MaxInt32
	} else {
		selectPolicyFn = max // Use the default policy otherwise to produce a highest possible change
		result = math.MinInt32
	}
	for _, policy := range scalingRules.Policies {
		replicasAddedInCurrentPeriod := getReplicasChangePerPeriod(policy.PeriodSeconds, scaleUpEvents)
		replicasDeletedInCurrentPeriod := getReplicasChangePerPeriod(policy.PeriodSeconds, scaleDownEvents)
		periodStartReplicas := currentReplicas - replicasAddedInCurrentPeriod + replicasDeletedInCurrentPeriod
		if policy.Type == autoscaling.PodsScalingPolicy {
			proposed = periodStartReplicas + policy.Value
		} else if policy.Type == autoscaling.PercentScalingPolicy {
			// the proposal has to be rounded up because the proposed change might not increase the replica count causing the target to never scale up
			proposed = int32(math.Ceil(float64(periodStartReplicas) * (1 + float64(policy.Value)/100)))
		}
		result = selectPolicyFn(result, proposed)
	}
	return result
}

// calculateScaleDownLimitWithBehaviors returns the maximum number of pods
// that could be deleted for the given GPAScalingRules
// NOCC:tosa/fn_length(设计如此)
func calculateScaleDownLimitWithBehaviors(currentReplicas int32, scaleUpEvents, scaleDownEvents []timestampedScaleEvent,
	scalingRules *autoscaling.GPAScalingRules) int32 {
	var result int32
	var proposed int32
	var selectPolicyFn func(int32, int32) int32
	if *scalingRules.SelectPolicy == autoscaling.DisabledPolicySelect {
		return currentReplicas // Scaling is disabled
	} else if *scalingRules.SelectPolicy == autoscaling.MinPolicySelect {
		result = math.MinInt32
		selectPolicyFn = max // For scaling down, the lowest change ('min' policy) produces a maximum value
	} else {
		selectPolicyFn = min // Use the default policy otherwise to produce a highest possible change
		result = math.MaxInt32
	}
	for _, policy := range scalingRules.Policies {
		replicasAddedInCurrentPeriod := getReplicasChangePerPeriod(policy.PeriodSeconds, scaleUpEvents)
		replicasDeletedInCurrentPeriod := getReplicasChangePerPeriod(policy.PeriodSeconds, scaleDownEvents)
		periodStartReplicas := currentReplicas - replicasAddedInCurrentPeriod + replicasDeletedInCurrentPeriod
		if policy.Type == autoscaling.PodsScalingPolicy {
			proposed = periodStartReplicas - policy.Value
		} else if policy.Type == autoscaling.PercentScalingPolicy {
			proposed = int32(float64(periodStartReplicas) * (1 - float64(policy.Value)/100))
		}
		result = selectPolicyFn(result, proposed)
	}
	return result
}

// scaleForResourceMappings attempts to fetch the scale for the
// resource with the given name and namespace, trying each RESTMapping
// in turn until a working one is found.  If none work, the first error
// is returned.  It returns both the scale, as well as the group-resource from
// the working mapping.
func (a *GeneralController) scaleForResourceMappings(ctx context.Context, namespace, name string,
	mappings []*apimeta.RESTMapping) (*autoscalinginternal.Scale, schema.GroupResource, error) {
	var firstErr error
	for i, mapping := range mappings {
		targetGR := mapping.Resource.GroupResource()
		scale, err := a.scaleNamespacer.Scales(namespace).Get(ctx, targetGR, name, metav1.GetOptions{})
		if err == nil {
			return scale, targetGR, nil
		}

		// if this is the first error, remember it,
		// then go on and try other mappings until we find a good one
		if i == 0 {
			firstErr = err
		}
	}

	// make sure we handle an empty set of mappings
	if firstErr == nil {
		firstErr = fmt.Errorf("unrecognized resource")
	}

	return nil, schema.GroupResource{}, firstErr
}

// setCurrentReplicasInStatus sets the current replica count in the status of the GPA.
func (a *GeneralController) setCurrentReplicasInStatus(gpa *autoscaling.GeneralPodAutoscaler, currentReplicas int32) {
	a.setStatus(gpa, currentReplicas, gpa.Status.DesiredReplicas, gpa.Status.CurrentMetrics, false)
}

// setStatus recreates the status of the given GPA, updating the current and
// desired replicas, as well as the metric statuses
func (a *GeneralController) setStatus(gpa *autoscaling.GeneralPodAutoscaler, currentReplicas,
	desiredReplicas int32, metricStatuses []autoscaling.MetricStatus, rescale bool) {
	gpa.Status = autoscaling.GeneralPodAutoscalerStatus{
		CurrentReplicas: currentReplicas,
		DesiredReplicas: desiredReplicas,
		LastScaleTime:   gpa.Status.LastScaleTime,
		CurrentMetrics:  metricStatuses,
		Conditions:      gpa.Status.Conditions,
	}
	now := metav1.NewTime(time.Now())
	if rescale {
		if gpa.Spec.TimeMode != nil {
			gpa.Status.LastCronScheduleTime = &now
		}
		gpa.Status.LastScaleTime = &now
	}
}

// updateStatusIfNeeded calls updateStatus only if the status of the new GPA is not the same as the old status
func (a *GeneralController) updateStatusIfNeeded(ctx context.Context, oldStatus *autoscaling.GeneralPodAutoscalerStatus,
	newGPA *autoscaling.GeneralPodAutoscaler) error {
	// skip a write if we wouldn't need to update
	if apiequality.Semantic.DeepEqual(oldStatus, &newGPA.Status) {
		return nil
	}
	return a.updateStatus(ctx, newGPA)
}

// updateStatus actually does the update request for the status of the given GPA
func (a *GeneralController) updateStatus(ctx context.Context, gpa *autoscaling.GeneralPodAutoscaler) error {
	_, err := a.gpaNamespacer.GeneralPodAutoscalers(gpa.Namespace).UpdateStatus(ctx, gpa, metav1.UpdateOptions{})
	if err != nil {
		a.eventRecorder.Event(gpa, v1.EventTypeWarning, "FailedUpdateStatus", err.Error())
		return fmt.Errorf("failed to update status for %s: %v", gpa.Name, err)
	}
	klog.V(2).Infof("Successfully updated status for %s", gpa.Name)
	return nil
}

// setCondition sets the specific condition type on the given GPA to the specified value with the given reason
// and message.  The message and args are treated like a format string.  The condition will be added if it is
// not present.
func setCondition(gpa *autoscaling.GeneralPodAutoscaler, conditionType autoscaling.GeneralPodAutoscalerConditionType,
	status v1.ConditionStatus, reason, message string, args ...interface{}) {
	gpa.Status.Conditions = setConditionInList(gpa.Status.Conditions, conditionType, status, reason, message, args...)
}

// setConditionInList sets the specific condition type on the given GPA to the specified value with the given
// reason and message.  The message and args are treated like a format string.  The condition will be added if
// it is not present.  The new list will be returned.
func setConditionInList(inputList []autoscaling.GeneralPodAutoscalerCondition,
	conditionType autoscaling.GeneralPodAutoscalerConditionType, status v1.ConditionStatus, reason, message string,
	args ...interface{}) []autoscaling.GeneralPodAutoscalerCondition {
	resList := inputList
	var existingCond *autoscaling.GeneralPodAutoscalerCondition
	for i, condition := range resList {
		if condition.Type == conditionType {
			// can't take a pointer to an iteration variable
			existingCond = &resList[i]
			break
		}
	}

	if existingCond == nil {
		resList = append(resList, autoscaling.GeneralPodAutoscalerCondition{
			Type: conditionType,
		})
		existingCond = &resList[len(resList)-1]
	}

	if existingCond.Status != status {
		existingCond.LastTransitionTime = metav1.Now()
	}

	existingCond.Status = status
	existingCond.Reason = reason
	existingCond.Message = fmt.Sprintf(message, args...)

	return resList
}

func max(a, b int32) int32 {
	if a >= b {
		return a
	}
	return b
}

func min(a, b int32) int32 {
	if a <= b {
		return a
	}
	return b
}

func isEmpty(a autoscaling.AutoScalingDrivenMode) bool {
	return a.MetricMode == nil && a.EventMode == nil && a.TimeMode == nil && a.WebhookMode == nil
}

func isComputeByLimits(gpa *autoscaling.GeneralPodAutoscaler) bool {
	computeByLimits := false
	if gpa != nil && gpa.Annotations != nil {
		computeByLimits = "true" == gpa.Annotations[computeByLimitsKey]
	}
	return computeByLimits
}

func (a *GeneralController) getDesiredReplicas(gpa *autoscaling.GeneralPodAutoscaler, key string,
	metricDesiredReplicas, desiredReplicas, currentReplicas, minReplicas int32,
	metricName string) (int32, string) {
	rescaleMetric := ""
	rescaleReason := ""
	if metricDesiredReplicas > desiredReplicas {
		desiredReplicas = metricDesiredReplicas
		rescaleMetric = metricName
	}
	if desiredReplicas > currentReplicas {
		rescaleReason = fmt.Sprintf("%s above target", rescaleMetric)
	}
	if desiredReplicas < currentReplicas {
		rescaleReason = "All metrics below target"
	}
	if gpa.Spec.Behavior == nil {
		desiredReplicas = a.normalizeDesiredReplicas(gpa, key, currentReplicas, desiredReplicas, minReplicas)
	} else {
		desiredReplicas = a.normalizeDesiredReplicasWithBehaviors(gpa, key, currentReplicas, desiredReplicas, minReplicas)
	}
	return desiredReplicas, rescaleReason
}

func disableRescale(currentReplicas, minReplicas int32) bool {
	return currentReplicas == 0 && minReplicas != 0
}
