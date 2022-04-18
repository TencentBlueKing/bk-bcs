/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package scaler

import (
	"fmt"
	"math"
	"strconv"
	"sync"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	autoscalinginternal "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta/testrestmapper"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"

	scalefake "k8s.io/client-go/scale/fake"
	cmapi "k8s.io/metrics/pkg/apis/custom_metrics/v1beta2"
	emapi "k8s.io/metrics/pkg/apis/external_metrics/v1beta1"
	metricsapi "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsfake "k8s.io/metrics/pkg/client/clientset/versioned/fake"
	cmfake "k8s.io/metrics/pkg/client/custom_metrics/fake"
	emfake "k8s.io/metrics/pkg/client/external_metrics/fake"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/apis/autoscaling"
	autoscalingv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/apis/autoscaling/v1alpha1"
	autoscalingfake "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/client/clientset/versioned/fake"
	autoscalinginformer "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/client/informers/externalversions"
	metricsclient "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/metrics"
)

var statusOk = []autoscalingv1alpha1.GeneralPodAutoscalerCondition{
	{Type: autoscalingv1alpha1.AbleToScale, Status: v1.ConditionTrue, Reason: "SucceededRescale"},
	{Type: autoscalingv1alpha1.ScalingActive, Status: v1.ConditionTrue, Reason: "ValidMetricFound"},
	{Type: autoscalingv1alpha1.ScalingLimited, Status: v1.ConditionFalse, Reason: "DesiredWithinRange"},
}

// statusOkWithOverrides returns the "ok" status with the given conditions as overridden
func statusOkWithOverrides(overrides ...autoscalingv1alpha1.GeneralPodAutoscalerCondition) []autoscalingv1alpha1.GeneralPodAutoscalerCondition {
	resv2 := make([]autoscalingv1alpha1.GeneralPodAutoscalerCondition, len(statusOk))
	copy(resv2, statusOk)
	for _, override := range overrides {
		resv2 = setConditionInList(resv2, override.Type, override.Status, override.Reason, override.Message)
	}

	// copy to a v1 slice
	resv1 := make([]autoscalingv1alpha1.GeneralPodAutoscalerCondition, len(resv2))
	for i, cond := range resv2 {
		resv1[i] = autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.GeneralPodAutoscalerConditionType(cond.Type),
			Status: cond.Status,
			Reason: cond.Reason,
		}
	}

	return resv1
}

func alwaysReady() bool { return true }

type fakeResource struct {
	name       string
	apiVersion string
	kind       string
}

type testCase struct {
	sync.Mutex
	minReplicas    int32
	maxReplicas    int32
	specReplicas   int32
	statusReplicas int32

	// CPU target utilization as a percentage of the requested resources.
	CPUTarget                    int32
	CPUCurrent                   int32
	verifyCPUCurrent             bool
	reportedLevels               []uint64
	reportedCPURequests          []resource.Quantity
	reportedCPULimits            []resource.Quantity
	reportedPodReadiness         []v1.ConditionStatus
	reportedPodStartTime         []metav1.Time
	reportedPodPhase             []v1.PodPhase
	reportedPodDeletionTimestamp []bool
	scaleUpdated                 bool
	statusUpdated                bool
	eventCreated                 bool
	verifyEvents                 bool
	useMetricsAPI                bool
	computeByLimits              bool
	metricsTarget                []autoscalingv1alpha1.MetricSpec
	expectedDesiredReplicas      int32
	expectedConditions           []autoscalingv1alpha1.GeneralPodAutoscalerCondition
	// Channel with names of GPA objects which we have reconciled.
	processed chan string

	// Target resource information.
	resource *fakeResource

	// Last scale time
	lastScaleTime *metav1.Time

	// override the test clients
	testClient        *fake.Clientset
	testGpaClient     *autoscalingfake.Clientset
	testMetricsClient *metricsfake.Clientset
	testCMClient      *cmfake.FakeCustomMetricsClient
	testEMClient      *emfake.FakeExternalMetricsClient
	testScaleClient   *scalefake.FakeScaleClient

	recommendations []timestampedRecommendation
}

// Needs to be called under a lock.
func (tc *testCase) computeCPUCurrent() {
	if len(tc.reportedLevels) != len(tc.reportedCPURequests) || len(tc.reportedLevels) == 0 {
		return
	}
	reported := 0
	for _, r := range tc.reportedLevels {
		reported += int(r)
	}
	requested := 0
	if tc.computeByLimits {
		for _, lim := range tc.reportedCPULimits {
			requested += int(lim.MilliValue())
		}
	} else {
		for _, req := range tc.reportedCPURequests {
			requested += int(req.MilliValue())
		}
	}

	tc.CPUCurrent = int32(100 * reported / requested)
}

func init() {
	// set this high so we don't accidentally run into it when testing
	scaleUpLimitFactor = 8
}

func (tc *testCase) prepareTestClient(t *testing.T) (*fake.Clientset, *metricsfake.Clientset, *cmfake.FakeCustomMetricsClient,
	*emfake.FakeExternalMetricsClient, *scalefake.FakeScaleClient, *autoscalingfake.Clientset) {
	namespace := "test-namespace"
	gpaName := "test-gpa"
	podNamePrefix := "test-pod"
	labelSet := map[string]string{"name": podNamePrefix}
	selector := labels.SelectorFromSet(labelSet).String()

	tc.Lock()

	tc.scaleUpdated = false
	tc.statusUpdated = false
	tc.eventCreated = false
	tc.processed = make(chan string, 100)
	if tc.CPUCurrent == 0 {
		tc.computeCPUCurrent()
	}

	if tc.resource == nil {
		tc.resource = &fakeResource{
			name:       "test-rc",
			apiVersion: "v1",
			kind:       "ReplicationController",
		}
	}
	tc.Unlock()

	fakeClient := &fake.Clientset{}
	fakeGPAClient := &autoscalingfake.Clientset{}

	fakeGPAClient.AddReactor("list", "generalpodautoscalers", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		tc.Lock()
		defer tc.Unlock()

		obj := &autoscalingv1alpha1.GeneralPodAutoscalerList{
			Items: []autoscalingv1alpha1.GeneralPodAutoscaler{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      gpaName,
						Namespace: namespace,
					},
					Spec: autoscalingv1alpha1.GeneralPodAutoscalerSpec{
						ScaleTargetRef: autoscalingv1alpha1.CrossVersionObjectReference{
							Kind:       tc.resource.kind,
							Name:       tc.resource.name,
							APIVersion: tc.resource.apiVersion,
						},
						MinReplicas: &tc.minReplicas,
						MaxReplicas: tc.maxReplicas,
					},
					Status: autoscalingv1alpha1.GeneralPodAutoscalerStatus{
						CurrentReplicas: tc.specReplicas,
						DesiredReplicas: tc.specReplicas,
						LastScaleTime:   tc.lastScaleTime,
					},
				},
			},
		}

		annotations := obj.Items[0].Annotations
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations[computeByLimitsKey] = strconv.FormatBool(tc.computeByLimits)
		obj.Items[0].Annotations = annotations
		obj.Items[0].Spec.AutoScalingDrivenMode = autoscalingv1alpha1.AutoScalingDrivenMode{
			MetricMode: &autoscalingv1alpha1.MetricMode{},
		}

		if tc.CPUTarget > 0 {
			obj.Items[0].Spec.MetricMode.Metrics = []autoscalingv1alpha1.MetricSpec{
				{
					Type: autoscalingv1alpha1.ResourceMetricSourceType,
					Resource: &autoscalingv1alpha1.ResourceMetricSource{
						Name: v1.ResourceCPU,
						Target: autoscalingv1alpha1.MetricTarget{
							AverageUtilization: &tc.CPUTarget,
						},
					},
				},
			}
		}

		if len(tc.metricsTarget) > 0 {
			obj.Items[0].Spec.MetricMode.Metrics = append(obj.Items[0].Spec.MetricMode.Metrics, tc.metricsTarget...)
		}

		if len(obj.Items[0].Spec.MetricMode.Metrics) == 0 {
			// manually add in the defaulting logic
			obj.Items[0].Spec.MetricMode.Metrics = []autoscalingv1alpha1.MetricSpec{
				{
					Type: autoscalingv1alpha1.ResourceMetricSourceType,
					Resource: &autoscalingv1alpha1.ResourceMetricSource{
						Name: v1.ResourceCPU,
					},
				},
			}
		}
		return true, obj, nil
	})

	fakeClient.AddReactor("list", "pods", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		tc.Lock()
		defer tc.Unlock()

		obj := &v1.PodList{}

		specifiedCPURequests := tc.reportedCPURequests != nil
		specifiedCPULimits := tc.reportedCPULimits != nil

		numPodsToCreate := int(tc.statusReplicas)
		if specifiedCPURequests {
			numPodsToCreate = len(tc.reportedCPURequests)
		}

		for i := 0; i < numPodsToCreate; i++ {
			podReadiness := v1.ConditionTrue
			if tc.reportedPodReadiness != nil {
				podReadiness = tc.reportedPodReadiness[i]
			}
			var podStartTime metav1.Time
			if tc.reportedPodStartTime != nil {
				podStartTime = tc.reportedPodStartTime[i]
			}

			podPhase := v1.PodRunning
			if tc.reportedPodPhase != nil {
				podPhase = tc.reportedPodPhase[i]
			}

			podDeletionTimestamp := false
			if tc.reportedPodDeletionTimestamp != nil {
				podDeletionTimestamp = tc.reportedPodDeletionTimestamp[i]
			}

			podName := fmt.Sprintf("%s-%d", podNamePrefix, i)

			reportedCPURequest := resource.MustParse("1.0")
			if specifiedCPURequests {
				reportedCPURequest = tc.reportedCPURequests[i]
			}

			reportedCPULimit := resource.MustParse("1.0")
			if specifiedCPULimits {
				reportedCPULimit = tc.reportedCPULimits[i]
			}

			pod := v1.Pod{
				Status: v1.PodStatus{
					Phase: podPhase,
					Conditions: []v1.PodCondition{
						{
							Type:               v1.PodReady,
							Status:             podReadiness,
							LastTransitionTime: podStartTime,
						},
					},
					StartTime: &podStartTime,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: namespace,
					Labels: map[string]string{
						"name": podNamePrefix,
					},
				},

				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "container",
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU: reportedCPURequest,
								},
								Limits: v1.ResourceList{
									v1.ResourceCPU: reportedCPULimit,
								},
							},
						},
					},
				},
			}
			if podDeletionTimestamp {
				pod.DeletionTimestamp = &metav1.Time{Time: time.Now()}
			}
			obj.Items = append(obj.Items, pod)
		}
		return true, obj, nil
	})

	fakeGPAClient.AddReactor("update", "generalpodautoscalers", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		handled, obj, err := func() (handled bool, ret *autoscalingv1alpha1.GeneralPodAutoscaler, err error) {
			tc.Lock()
			defer tc.Unlock()
			obj := action.(core.UpdateAction).GetObject().(*autoscalingv1alpha1.GeneralPodAutoscaler)
			assert.Equal(t, namespace, obj.Namespace, "the GPA namespace should be as expected")
			assert.Equal(t, gpaName, obj.Name, "the GPA name should be as expected")
			assert.Equal(t, tc.expectedDesiredReplicas, obj.Status.DesiredReplicas, "the desired replica count reported in the object status should be as expected")
			// Every time we reconcile GPA object we are updating status.
			tc.statusUpdated = true
			return true, obj, nil
		}()
		if obj != nil {
			tc.processed <- obj.Name
		}
		return handled, obj, err
	})

	fakeScaleClient := &scalefake.FakeScaleClient{}
	fakeScaleClient.AddReactor("get", "replicationcontrollers", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		tc.Lock()
		defer tc.Unlock()

		obj := &autoscalinginternal.Scale{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tc.resource.name,
				Namespace: namespace,
			},
			Spec: autoscalinginternal.ScaleSpec{
				Replicas: tc.specReplicas,
			},
			Status: autoscalinginternal.ScaleStatus{
				Replicas: tc.statusReplicas,
				Selector: selector,
			},
		}
		return true, obj, nil
	})

	fakeScaleClient.AddReactor("get", "deployments", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		tc.Lock()
		defer tc.Unlock()

		obj := &autoscalinginternal.Scale{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tc.resource.name,
				Namespace: namespace,
			},
			Spec: autoscalinginternal.ScaleSpec{
				Replicas: tc.specReplicas,
			},
			Status: autoscalinginternal.ScaleStatus{
				Replicas: tc.statusReplicas,
				Selector: selector,
			},
		}
		return true, obj, nil
	})

	fakeScaleClient.AddReactor("get", "replicasets", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		tc.Lock()
		defer tc.Unlock()

		obj := &autoscalinginternal.Scale{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tc.resource.name,
				Namespace: namespace,
			},
			Spec: autoscalinginternal.ScaleSpec{
				Replicas: tc.specReplicas,
			},
			Status: autoscalinginternal.ScaleStatus{
				Replicas: tc.statusReplicas,
				Selector: selector,
			},
		}
		return true, obj, nil
	})

	fakeScaleClient.AddReactor("update", "replicationcontrollers", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		tc.Lock()
		defer tc.Unlock()

		obj := action.(core.UpdateAction).GetObject().(*autoscalinginternal.Scale)
		replicas := action.(core.UpdateAction).GetObject().(*autoscalinginternal.Scale).Spec.Replicas
		assert.Equal(t, tc.expectedDesiredReplicas, replicas, "the replica count of the RC should be as expected")
		tc.scaleUpdated = true
		return true, obj, nil
	})

	fakeScaleClient.AddReactor("update", "deployments", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		tc.Lock()
		defer tc.Unlock()

		obj := action.(core.UpdateAction).GetObject().(*autoscalinginternal.Scale)
		replicas := action.(core.UpdateAction).GetObject().(*autoscalinginternal.Scale).Spec.Replicas
		assert.Equal(t, tc.expectedDesiredReplicas, replicas, "the replica count of the deployment should be as expected")
		tc.scaleUpdated = true
		return true, obj, nil
	})

	fakeScaleClient.AddReactor("update", "replicasets", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		tc.Lock()
		defer tc.Unlock()

		obj := action.(core.UpdateAction).GetObject().(*autoscalinginternal.Scale)
		replicas := action.(core.UpdateAction).GetObject().(*autoscalinginternal.Scale).Spec.Replicas
		assert.Equal(t, tc.expectedDesiredReplicas, replicas, "the replica count of the replicaset should be as expected")
		tc.scaleUpdated = true
		return true, obj, nil
	})

	fakeWatch := watch.NewFake()
	fakeClient.AddWatchReactor("*", core.DefaultWatchReactor(fakeWatch, nil))
	fakeGPAClient.AddWatchReactor("*", core.DefaultWatchReactor(fakeWatch, nil))

	fakeMetricsClient := &metricsfake.Clientset{}
	fakeMetricsClient.AddReactor("list", "pods", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		tc.Lock()
		defer tc.Unlock()

		metrics := &metricsapi.PodMetricsList{}
		for i, cpu := range tc.reportedLevels {
			// NB: the list reactor actually does label selector filtering for us,
			// so we have to make sure our results match the label selector
			podMetric := metricsapi.PodMetrics{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%d", podNamePrefix, i),
					Namespace: namespace,
					Labels:    labelSet,
				},
				Timestamp: metav1.Time{Time: time.Now()},
				Window:    metav1.Duration{Duration: time.Minute},
				Containers: []metricsapi.ContainerMetrics{
					{
						Name: "container",
						Usage: v1.ResourceList{
							v1.ResourceCPU: *resource.NewMilliQuantity(
								int64(cpu),
								resource.DecimalSI),
							v1.ResourceMemory: *resource.NewQuantity(
								int64(1024*1024),
								resource.BinarySI),
						},
					},
				},
			}
			metrics.Items = append(metrics.Items, podMetric)
		}

		return true, metrics, nil
	})

	fakeCMClient := &cmfake.FakeCustomMetricsClient{}
	fakeCMClient.AddReactor("get", "*", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		tc.Lock()
		defer tc.Unlock()

		getForAction, wasGetFor := action.(cmfake.GetForAction)
		if !wasGetFor {
			return true, nil, fmt.Errorf("expected a get-for action, got %v instead", action)
		}

		if getForAction.GetName() == "*" {
			metrics := &cmapi.MetricValueList{}

			// multiple objects
			assert.Equal(t, "pods", getForAction.GetResource().Resource, "the type of object that we requested multiple metrics for should have been pods")
			assert.Equal(t, "qps", getForAction.GetMetricName(), "the metric name requested should have been qps, as specified in the metric spec")

			for i, level := range tc.reportedLevels {
				podMetric := cmapi.MetricValue{
					DescribedObject: v1.ObjectReference{
						Kind:      "Pod",
						Name:      fmt.Sprintf("%s-%d", podNamePrefix, i),
						Namespace: namespace,
					},
					Timestamp: metav1.Time{Time: time.Now()},
					Metric: cmapi.MetricIdentifier{
						Name: "qps",
					},
					Value: *resource.NewMilliQuantity(int64(level), resource.DecimalSI),
				}
				metrics.Items = append(metrics.Items, podMetric)
			}

			return true, metrics, nil
		}

		name := getForAction.GetName()
		mapper := testrestmapper.TestOnlyStaticRESTMapper(testScheme())
		metrics := &cmapi.MetricValueList{}
		var matchedTarget *autoscalingv1alpha1.MetricSpec
		for i, target := range tc.metricsTarget {
			if target.Type == autoscalingv1alpha1.ObjectMetricSourceType && name == target.Object.DescribedObject.Name {
				gk := schema.FromAPIVersionAndKind(target.Object.DescribedObject.APIVersion, target.Object.DescribedObject.Kind).GroupKind()
				mapping, err := mapper.RESTMapping(gk)
				if err != nil {
					t.Logf("unable to get mapping for %s: %v", gk.String(), err)
					continue
				}
				groupResource := mapping.Resource.GroupResource()

				if getForAction.GetResource().Resource == groupResource.String() {
					matchedTarget = &tc.metricsTarget[i]
				}
			}
		}
		assert.NotNil(t, matchedTarget, "this request should have matched one of the metric specs")
		assert.Equal(t, "qps", getForAction.GetMetricName(), "the metric name requested should have been qps, as specified in the metric spec")

		metrics.Items = []cmapi.MetricValue{
			{
				DescribedObject: v1.ObjectReference{
					Kind:       matchedTarget.Object.DescribedObject.Kind,
					APIVersion: matchedTarget.Object.DescribedObject.APIVersion,
					Name:       name,
				},
				Timestamp: metav1.Time{Time: time.Now()},
				Metric: cmapi.MetricIdentifier{
					Name: "qps",
				},
				Value: *resource.NewMilliQuantity(int64(tc.reportedLevels[0]), resource.DecimalSI),
			},
		}

		return true, metrics, nil
	})

	fakeEMClient := &emfake.FakeExternalMetricsClient{}

	fakeEMClient.AddReactor("list", "*", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		tc.Lock()
		defer tc.Unlock()

		listAction, wasList := action.(core.ListAction)
		if !wasList {
			return true, nil, fmt.Errorf("expected a list action, got %v instead", action)
		}

		metrics := &emapi.ExternalMetricValueList{}

		assert.Equal(t, "qps", listAction.GetResource().Resource, "the metric name requested should have been qps, as specified in the metric spec")

		for _, level := range tc.reportedLevels {
			metric := emapi.ExternalMetricValue{
				Timestamp:  metav1.Time{Time: time.Now()},
				MetricName: "qps",
				Value:      *resource.NewMilliQuantity(int64(level), resource.DecimalSI),
			}
			metrics.Items = append(metrics.Items, metric)
		}

		return true, metrics, nil
	})

	return fakeClient, fakeMetricsClient, fakeCMClient, fakeEMClient, fakeScaleClient, fakeGPAClient
}

func (tc *testCase) verifyResults(t *testing.T) {
	tc.Lock()
	defer tc.Unlock()
	assert.Equal(t, tc.specReplicas != tc.expectedDesiredReplicas, tc.scaleUpdated, "the scale should only be updated if we expected a change in replicas")
	assert.True(t, tc.statusUpdated, "the status should have been updated")
	if tc.verifyEvents {
		assert.Equal(t, tc.specReplicas != tc.expectedDesiredReplicas, tc.eventCreated, "an event should have been created only if we expected a change in replicas")
	}
}

func (tc *testCase) setupController(t *testing.T) (*GeneralController, informers.SharedInformerFactory, autoscalinginformer.SharedInformerFactory) {
	testClient, testMetricsClient, testCMClient, testEMClient, testScaleClient, testGPAClient := tc.prepareTestClient(t)

	if tc.testClient != nil {
		testClient = tc.testClient
	}
	if tc.testMetricsClient != nil {
		testMetricsClient = tc.testMetricsClient
	}
	if tc.testCMClient != nil {
		testCMClient = tc.testCMClient
	}
	if tc.testEMClient != nil {
		testEMClient = tc.testEMClient
	}
	if tc.testScaleClient != nil {
		testScaleClient = tc.testScaleClient
	}
	if tc.testGpaClient != nil {
		testGPAClient = tc.testGpaClient
	}
	metricsClient := metricsclient.NewRESTMetricsClient(
		testMetricsClient.MetricsV1beta1(),
		testCMClient,
		testEMClient,
	)

	eventClient := &fake.Clientset{}
	eventClient.AddReactor("create", "events", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		tc.Lock()
		defer tc.Unlock()

		obj := action.(core.CreateAction).GetObject().(*v1.Event)
		if tc.verifyEvents {
			switch obj.Reason {
			case "SuccessfulRescale":
				computeResourceUtilizationRatioBy := "request"
				if tc.computeByLimits {
					computeResourceUtilizationRatioBy = "limit"
				}
				assert.Equal(t, fmt.Sprintf("New size: %d; reason: cpu resource utilization (percentage of %s) above target", tc.expectedDesiredReplicas, computeResourceUtilizationRatioBy), obj.Message)
			case "DesiredReplicasComputed":
				assert.Equal(t, fmt.Sprintf(
					"Computed the desired num of replicas: %d (avgCPUutil: %d, current replicas: %d)",
					tc.expectedDesiredReplicas,
					(int64(tc.reportedLevels[0])*100)/tc.reportedCPURequests[0].MilliValue(), tc.specReplicas), obj.Message)
			default:
				assert.False(t, true, fmt.Sprintf("Unexpected event: %s / %s", obj.Reason, obj.Message))
			}
		}
		tc.eventCreated = true
		return true, obj, nil
	})

	scalerFactory := autoscalinginformer.NewSharedInformerFactory(testGPAClient, 0)
	informerFactory := informers.NewSharedInformerFactory(testClient, 0)

	defaultDownscalestabilizationWindow := 5 * time.Minute
	gpaController := NewGeneralController(
		eventClient.CoreV1(),
		testScaleClient,
		testGPAClient.AutoscalingV1alpha1(),
		testrestmapper.TestOnlyStaticRESTMapper(testScheme()),
		metricsClient,
		scalerFactory.Autoscaling().V1alpha1().GeneralPodAutoscalers(),
		informerFactory.Core().V1().Pods(),
		0,
		defaultDownscalestabilizationWindow,
		defaultTestingTolerance,
		defaultTestingCPUInitializationPeriod,
		defaultTestingDelayOfInitialReadinessStatus,
	)
	gpaController.gpaListerSynced = alwaysReady
	if tc.recommendations != nil {
		gpaController.recommendations["test-namespace/test-gpa"] = tc.recommendations
	}

	return gpaController, informerFactory, scalerFactory
}

func (tc *testCase) runTestWithController(t *testing.T, gpaController *GeneralController, informerFactory informers.SharedInformerFactory,
	scalerFactory autoscalinginformer.SharedInformerFactory) {
	stop := make(chan struct{})
	defer close(stop)
	scalerFactory.Start(stop)
	informerFactory.Start(stop)
	go gpaController.Run(stop)
	tc.Lock()
	shouldWait := tc.verifyEvents
	tc.Unlock()
	if shouldWait {
		// We need to wait for events to be broadcasted (sleep for longer than record.sleepDuration).
		timeoutTime := time.Now().Add(2 * time.Second)
		for now := time.Now(); timeoutTime.After(now); now = time.Now() {
			sleepUntil := timeoutTime.Sub(now)
			select {
			case <-tc.processed:
				// drain the chan of any sent events to keep it from filling before the timeout
			case <-time.After(sleepUntil):
				// timeout reached, ready to verifyResults
			}
		}
	} else {
		// Wait for GPA to be processed.
		<-tc.processed
	}
	tc.verifyResults(t)
}

func (tc *testCase) runTest(t *testing.T) {
	gpaController, informerFactory, scalerFactory := tc.setupController(t)
	tc.runTestWithController(t, gpaController, informerFactory, scalerFactory)
}

func TestScaleUp(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 5,
		CPUTarget:               30,
		verifyCPUCurrent:        true,
		reportedLevels:          []uint64{300, 500, 700},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		useMetricsAPI:           true,
	}
	tc.runTest(t)
}

func TestScaleUpUnreadyLessScale(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 4,
		CPUTarget:               30,
		CPUCurrent:              60,
		verifyCPUCurrent:        true,
		reportedLevels:          []uint64{300, 500, 700},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		reportedPodReadiness:    []v1.ConditionStatus{v1.ConditionFalse, v1.ConditionTrue, v1.ConditionTrue},
		useMetricsAPI:           true,
	}
	tc.runTest(t)
}

func TestScaleUpHotCpuLessScale(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 4,
		CPUTarget:               30,
		CPUCurrent:              60,
		verifyCPUCurrent:        true,
		reportedLevels:          []uint64{300, 500, 700},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		reportedPodStartTime:    []metav1.Time{hotCpuCreationTime(), coolCpuCreationTime(), coolCpuCreationTime()},
		useMetricsAPI:           true,
	}
	tc.runTest(t)
}

func TestScaleUpUnreadyNoScale(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 3,
		CPUTarget:               30,
		CPUCurrent:              40,
		verifyCPUCurrent:        true,
		reportedLevels:          []uint64{400, 500, 700},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		reportedPodReadiness:    []v1.ConditionStatus{v1.ConditionTrue, v1.ConditionFalse, v1.ConditionFalse},
		useMetricsAPI:           true,
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionTrue,
			Reason: "ReadyForNewScale",
		}),
	}
	tc.runTest(t)
}

func TestScaleUpHotCpuNoScale(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 3,
		CPUTarget:               30,
		CPUCurrent:              40,
		verifyCPUCurrent:        true,
		reportedLevels:          []uint64{400, 500, 700},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		reportedPodReadiness:    []v1.ConditionStatus{v1.ConditionTrue, v1.ConditionFalse, v1.ConditionFalse},
		reportedPodStartTime:    []metav1.Time{coolCpuCreationTime(), hotCpuCreationTime(), hotCpuCreationTime()},
		useMetricsAPI:           true,
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionTrue,
			Reason: "ReadyForNewScale",
		}),
	}
	tc.runTest(t)
}

func TestScaleUpIgnoresFailedPods(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            2,
		statusReplicas:          2,
		expectedDesiredReplicas: 4,
		CPUTarget:               30,
		CPUCurrent:              60,
		verifyCPUCurrent:        true,
		reportedLevels:          []uint64{500, 700},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		reportedPodReadiness:    []v1.ConditionStatus{v1.ConditionTrue, v1.ConditionTrue, v1.ConditionFalse, v1.ConditionFalse},
		reportedPodPhase:        []v1.PodPhase{v1.PodRunning, v1.PodRunning, v1.PodFailed, v1.PodFailed},
		useMetricsAPI:           true,
	}
	tc.runTest(t)
}

func TestScaleUpIgnoresDeletionPods(t *testing.T) {
	tc := testCase{
		minReplicas:                  2,
		maxReplicas:                  6,
		specReplicas:                 2,
		statusReplicas:               2,
		expectedDesiredReplicas:      4,
		CPUTarget:                    30,
		CPUCurrent:                   60,
		verifyCPUCurrent:             true,
		reportedLevels:               []uint64{500, 700},
		reportedCPURequests:          []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		reportedPodReadiness:         []v1.ConditionStatus{v1.ConditionTrue, v1.ConditionTrue, v1.ConditionFalse, v1.ConditionFalse},
		reportedPodPhase:             []v1.PodPhase{v1.PodRunning, v1.PodRunning, v1.PodRunning, v1.PodRunning},
		reportedPodDeletionTimestamp: []bool{false, false, true, true},
		useMetricsAPI:                true,
	}
	tc.runTest(t)
}

func TestScaleUpDeployment(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 5,
		CPUTarget:               30,
		verifyCPUCurrent:        true,
		reportedLevels:          []uint64{300, 500, 700},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		useMetricsAPI:           true,
		resource: &fakeResource{
			name:       "test-dep",
			apiVersion: "apps/v1",
			kind:       "Deployment",
		},
	}
	tc.runTest(t)
}

func TestScaleUpReplicaSet(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 5,
		CPUTarget:               30,
		verifyCPUCurrent:        true,
		reportedLevels:          []uint64{300, 500, 700},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		useMetricsAPI:           true,
		resource: &fakeResource{
			name:       "test-replicaset",
			apiVersion: "apps/v1",
			kind:       "ReplicaSet",
		},
	}
	tc.runTest(t)
}

func TestScaleUpCM(t *testing.T) {
	averageValue := resource.MustParse("15.0")
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 4,
		CPUTarget:               0,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.PodsMetricSourceType,
				Pods: &autoscalingv1alpha1.PodsMetricSource{
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name: "qps",
					},
					Target: autoscalingv1alpha1.MetricTarget{
						AverageValue: &averageValue,
					},
				},
			},
		},
		reportedLevels:      []uint64{20000, 10000, 30000},
		reportedCPURequests: []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
	}
	tc.runTest(t)
}

// TestScaleUpCMUnreadyAndHotCpuNLS Test Scale Up CM Unready And Hot Cpu No Less Scale
func TestScaleUpCMUnreadyAndHotCpuNLS(t *testing.T) {
	averageValue := resource.MustParse("15.0")
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 6,
		CPUTarget:               0,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.PodsMetricSourceType,
				Pods: &autoscalingv1alpha1.PodsMetricSource{
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name: "qps",
					},
					Target: autoscalingv1alpha1.MetricTarget{
						AverageValue: &averageValue,
					},
				},
			},
		},
		reportedLevels:       []uint64{50000, 10000, 30000},
		reportedPodReadiness: []v1.ConditionStatus{v1.ConditionTrue, v1.ConditionTrue, v1.ConditionFalse},
		reportedPodStartTime: []metav1.Time{coolCpuCreationTime(), coolCpuCreationTime(), hotCpuCreationTime()},
		reportedCPURequests:  []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
	}
	tc.runTest(t)
}

func TestScaleUpCMUnreadyandCpuHot(t *testing.T) {
	averageValue := resource.MustParse("15.0")
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 6,
		CPUTarget:               0,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.PodsMetricSourceType,
				Pods: &autoscalingv1alpha1.PodsMetricSource{
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name: "qps",
					},
					Target: autoscalingv1alpha1.MetricTarget{
						AverageValue: &averageValue,
					},
				},
			},
		},
		reportedLevels:       []uint64{50000, 15000, 30000},
		reportedPodReadiness: []v1.ConditionStatus{v1.ConditionFalse, v1.ConditionTrue, v1.ConditionFalse},
		reportedPodStartTime: []metav1.Time{hotCpuCreationTime(), coolCpuCreationTime(), hotCpuCreationTime()},
		reportedCPURequests:  []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionTrue,
			Reason: "SucceededRescale",
		}, autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.ScalingLimited,
			Status: v1.ConditionTrue,
			Reason: "TooManyReplicas",
		}),
	}
	tc.runTest(t)
}

// TestScaleUpHotCpuNoScaleWouldSD 原方法名TestScaleUpHotCpuNoScaleWouldScaleDown
//
//TestScaleUpHotCpuNoScaleWouldSD Test Scale Up Hot Cpu No Scale Would Scale Down
func TestScaleUpHotCpuNoScaleWouldSD(t *testing.T) {
	averageValue := resource.MustParse("15.0")
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 6,
		CPUTarget:               0,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.PodsMetricSourceType,
				Pods: &autoscalingv1alpha1.PodsMetricSource{
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name: "qps",
					},
					Target: autoscalingv1alpha1.MetricTarget{
						AverageValue: &averageValue,
					},
				},
			},
		},
		reportedLevels:       []uint64{50000, 15000, 30000},
		reportedCPURequests:  []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		reportedPodStartTime: []metav1.Time{hotCpuCreationTime(), coolCpuCreationTime(), hotCpuCreationTime()},
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionTrue,
			Reason: "SucceededRescale",
		}, autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.ScalingLimited,
			Status: v1.ConditionTrue,
			Reason: "TooManyReplicas",
		}),
	}
	tc.runTest(t)
}

func TestScaleUpCMObject(t *testing.T) {
	targetValue := resource.MustParse("15.0")
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 4,
		CPUTarget:               0,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.ObjectMetricSourceType,
				Object: &autoscalingv1alpha1.ObjectMetricSource{
					DescribedObject: autoscalingv1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "some-deployment",
					},
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name: "qps",
					},
					Target: autoscalingv1alpha1.MetricTarget{
						Value: &targetValue,
						Type:  autoscalingv1alpha1.ValueMetricType,
					},
				},
			},
		},
		reportedLevels: []uint64{20000},
	}
	tc.runTest(t)
}

func TestScaleUpFromZeroCMObject(t *testing.T) {
	targetValue := resource.MustParse("15.0")
	tc := testCase{
		minReplicas:             0,
		maxReplicas:             6,
		specReplicas:            0,
		statusReplicas:          0,
		expectedDesiredReplicas: 2,
		CPUTarget:               0,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.ObjectMetricSourceType,
				Object: &autoscalingv1alpha1.ObjectMetricSource{
					DescribedObject: autoscalingv1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "some-deployment",
					},
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name: "qps",
					},
					Target: autoscalingv1alpha1.MetricTarget{
						Value: &targetValue,
						Type:  autoscalingv1alpha1.ValueMetricType,
					},
				},
			},
		},
		reportedLevels: []uint64{20000},
	}
	tc.runTest(t)
}

// TestScaleUpFromZeroIgnoresTCMO 原方法名TestScaleUpFromZeroIgnoresToleranceCMObject
//
//TestScaleUpFromZeroIgnoresTCMO Test Scale Up From Zero Ignores Tolerance CM Object
func TestScaleUpFromZeroIgnoresTCMO(t *testing.T) {
	targetValue := resource.MustParse("1.0")
	tc := testCase{
		minReplicas:             0,
		maxReplicas:             6,
		specReplicas:            0,
		statusReplicas:          0,
		expectedDesiredReplicas: 1,
		CPUTarget:               0,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.ObjectMetricSourceType,
				Object: &autoscalingv1alpha1.ObjectMetricSource{
					DescribedObject: autoscalingv1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "some-deployment",
					},
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name: "qps",
					},
					Target: autoscalingv1alpha1.MetricTarget{
						Value: &targetValue,
						Type:  autoscalingv1alpha1.ValueMetricType,
					},
				},
			},
		},
		reportedLevels: []uint64{1000},
	}
	tc.runTest(t)
}

func TestScaleUpPerPodCMObject(t *testing.T) {
	targetAverageValue := resource.MustParse("10.0")
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 4,
		CPUTarget:               0,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.ObjectMetricSourceType,
				Object: &autoscalingv1alpha1.ObjectMetricSource{
					DescribedObject: autoscalingv1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "some-deployment",
					},
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name: "qps",
					},
					Target: autoscalingv1alpha1.MetricTarget{
						AverageValue: &targetAverageValue,
						Type:         autoscalingv1alpha1.AverageValueMetricType,
					},
				},
			},
		},
		reportedLevels: []uint64{40000},
	}
	tc.runTest(t)
}

func TestScaleUpCMExternal(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 4,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.ExternalMetricSourceType,
				External: &autoscalingv1alpha1.ExternalMetricSource{
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name:     "qps",
						Selector: &metav1.LabelSelector{},
					},
					Target: autoscalingv1alpha1.MetricTarget{
						Value: resource.NewMilliQuantity(6666, resource.DecimalSI),
					},
				},
			},
		},
		reportedLevels: []uint64{8600},
	}
	tc.runTest(t)
}

func TestScaleUpByContainerResource(t *testing.T) {
	var cpuUtilization int32 = 30
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 6,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{{
			Type: autoscalingv1alpha1.ContainerResourceMetricSourceType,
			ContainerResource: &autoscalingv1alpha1.ContainerResourceMetricSource{
				Name: v1.ResourceCPU,
				Target: autoscalingv1alpha1.MetricTarget{
					Type:               autoscalingv1alpha1.UtilizationMetricType,
					AverageUtilization: &cpuUtilization,
				},
				Container: "container",
			},
		}},
		reportedLevels:      []uint64{800, 900, 900},
		reportedCPURequests: []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		useMetricsAPI:       true,
	}
	tc.runTest(t)
}

func TestScaleUpByLimits(t *testing.T) {
	var cpuUtilization int32 = 40
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            2,
		statusReplicas:          2,
		expectedDesiredReplicas: 2,
		computeByLimits:         true,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{{
			Type: autoscalingv1alpha1.ContainerResourceMetricSourceType,
			ContainerResource: &autoscalingv1alpha1.ContainerResourceMetricSource{
				Name: v1.ResourceCPU,
				Target: autoscalingv1alpha1.MetricTarget{
					Type:               autoscalingv1alpha1.UtilizationMetricType,
					AverageUtilization: &cpuUtilization,
				},
				Container: "container",
			},
		}},
		reportedLevels:      []uint64{300, 300, 200},
		reportedCPURequests: []resource.Quantity{resource.MustParse("0.5"), resource.MustParse("0.5"), resource.MustParse("0.5")},
		reportedCPULimits:   []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		useMetricsAPI:       false,
	}
	tc.runTest(t)
}

func TestScaleUpPerPodCMExternal(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 4,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.ExternalMetricSourceType,
				External: &autoscalingv1alpha1.ExternalMetricSource{
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name:     "qps",
						Selector: &metav1.LabelSelector{},
					},
					Target: autoscalingv1alpha1.MetricTarget{
						AverageValue: resource.NewMilliQuantity(2222, resource.DecimalSI),
					},
				},
			},
		},
		reportedLevels: []uint64{8600},
	}
	tc.runTest(t)
}

func TestScaleDown(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            5,
		statusReplicas:          5,
		expectedDesiredReplicas: 3,
		CPUTarget:               50,
		verifyCPUCurrent:        true,
		reportedLevels:          []uint64{100, 300, 500, 250, 250},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		useMetricsAPI:           true,
		recommendations:         []timestampedRecommendation{},
	}
	tc.runTest(t)
}

func TestScaleUpOneMetricInvalid(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 4,
		CPUTarget:               30,
		verifyCPUCurrent:        true,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: "CheddarCheese",
			},
		},
		reportedLevels:      []uint64{300, 400, 500},
		reportedCPURequests: []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
	}
	tc.runTest(t)
}

func TestScaleUpFromZeroOneMetricInvalid(t *testing.T) {
	tc := testCase{
		minReplicas:             0,
		maxReplicas:             6,
		specReplicas:            0,
		statusReplicas:          0,
		expectedDesiredReplicas: 4,
		CPUTarget:               30,
		verifyCPUCurrent:        true,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: "CheddarCheese",
			},
		},
		reportedLevels:      []uint64{300, 400, 500},
		reportedCPURequests: []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		recommendations:     []timestampedRecommendation{},
	}
	tc.runTest(t)
}

func TestScaleUpBothMetricsEmpty(t *testing.T) { // Switch to missing
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 3,
		CPUTarget:               0,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: "CheddarCheese",
			},
		},
		reportedLevels:      []uint64{},
		reportedCPURequests: []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		expectedConditions: []autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			{Type: autoscalingv1alpha1.AbleToScale, Status: v1.ConditionTrue, Reason: "SucceededGetScale"},
			{Type: autoscalingv1alpha1.ScalingActive, Status: v1.ConditionFalse, Reason: "InvalidMetricSourceType"},
		},
	}
	tc.runTest(t)
}

func TestScaleDownStabilizeInitialSize(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            5,
		statusReplicas:          5,
		expectedDesiredReplicas: 5,
		CPUTarget:               50,
		verifyCPUCurrent:        true,
		reportedLevels:          []uint64{100, 300, 500, 250, 250},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		useMetricsAPI:           true,
		recommendations:         nil,
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionTrue,
			Reason: "ReadyForNewScale",
		}, autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionTrue,
			Reason: "ScaleDownStabilized",
		}),
	}
	tc.runTest(t)
}

func TestScaleDownCM(t *testing.T) {
	averageValue := resource.MustParse("20.0")
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            5,
		statusReplicas:          5,
		expectedDesiredReplicas: 3,
		CPUTarget:               0,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.PodsMetricSourceType,
				Pods: &autoscalingv1alpha1.PodsMetricSource{
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name: "qps",
					},
					Target: autoscalingv1alpha1.MetricTarget{
						AverageValue: &averageValue,
					},
				},
			},
		},
		reportedLevels:      []uint64{12000, 12000, 12000, 12000, 12000},
		reportedCPURequests: []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		recommendations:     []timestampedRecommendation{},
	}
	tc.runTest(t)
}

func TestScaleDownCMObject(t *testing.T) {
	targetValue := resource.MustParse("20.0")
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            5,
		statusReplicas:          5,
		expectedDesiredReplicas: 3,
		CPUTarget:               0,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.ObjectMetricSourceType,
				Object: &autoscalingv1alpha1.ObjectMetricSource{
					DescribedObject: autoscalingv1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "some-deployment",
					},
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name: "qps",
					},
					Target: autoscalingv1alpha1.MetricTarget{
						Value: &targetValue,
						Type:  autoscalingv1alpha1.ValueMetricType,
					},
				},
			},
		},
		reportedLevels:      []uint64{12000},
		reportedCPURequests: []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		recommendations:     []timestampedRecommendation{},
	}
	tc.runTest(t)
}

func TestScaleDownToZeroCMObject(t *testing.T) {
	targetValue := resource.MustParse("20.0")
	tc := testCase{
		minReplicas:             0,
		maxReplicas:             6,
		specReplicas:            5,
		statusReplicas:          5,
		expectedDesiredReplicas: 0,
		CPUTarget:               0,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.ObjectMetricSourceType,
				Object: &autoscalingv1alpha1.ObjectMetricSource{
					DescribedObject: autoscalingv1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "some-deployment",
					},
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name: "qps",
					},
					Target: autoscalingv1alpha1.MetricTarget{
						Value: &targetValue,
						Type:  autoscalingv1alpha1.ValueMetricType,
					},
				},
			},
		},
		reportedLevels:      []uint64{0},
		reportedCPURequests: []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		recommendations:     []timestampedRecommendation{},
	}
	tc.runTest(t)
}

func TestScaleDownPerPodCMObject(t *testing.T) {
	targetAverageValue := resource.MustParse("20.0")
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            5,
		statusReplicas:          5,
		expectedDesiredReplicas: 3,
		CPUTarget:               0,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.ObjectMetricSourceType,
				Object: &autoscalingv1alpha1.ObjectMetricSource{
					DescribedObject: autoscalingv1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "some-deployment",
					},
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name: "qps",
					},
					Target: autoscalingv1alpha1.MetricTarget{
						AverageValue: &targetAverageValue,
						Type:         autoscalingv1alpha1.AverageValueMetricType,
					},
				},
			},
		},
		reportedLevels:      []uint64{60000},
		reportedCPURequests: []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		recommendations:     []timestampedRecommendation{},
	}
	tc.runTest(t)
}

func TestScaleDownCMExternal(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            5,
		statusReplicas:          5,
		expectedDesiredReplicas: 3,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.ExternalMetricSourceType,
				External: &autoscalingv1alpha1.ExternalMetricSource{
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name:     "qps",
						Selector: &metav1.LabelSelector{},
					},
					Target: autoscalingv1alpha1.MetricTarget{
						Value: resource.NewMilliQuantity(14400, resource.DecimalSI),
					},
				},
			},
		},
		reportedLevels:  []uint64{8600},
		recommendations: []timestampedRecommendation{},
	}
	tc.runTest(t)
}

func TestScaleDownToZeroCMExternal(t *testing.T) {
	tc := testCase{
		minReplicas:             0,
		maxReplicas:             6,
		specReplicas:            5,
		statusReplicas:          5,
		expectedDesiredReplicas: 0,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.ExternalMetricSourceType,
				External: &autoscalingv1alpha1.ExternalMetricSource{
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name:     "qps",
						Selector: &metav1.LabelSelector{},
					},
					Target: autoscalingv1alpha1.MetricTarget{
						Value: resource.NewMilliQuantity(14400, resource.DecimalSI),
					},
				},
			},
		},
		reportedLevels:  []uint64{0},
		recommendations: []timestampedRecommendation{},
	}
	tc.runTest(t)
}

func TestScaleDownPerPodCMExternal(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            5,
		statusReplicas:          5,
		expectedDesiredReplicas: 3,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.ExternalMetricSourceType,
				External: &autoscalingv1alpha1.ExternalMetricSource{
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name:     "qps",
						Selector: &metav1.LabelSelector{},
					},
					Target: autoscalingv1alpha1.MetricTarget{
						AverageValue: resource.NewMilliQuantity(3000, resource.DecimalSI),
					},
				},
			},
		},
		reportedLevels:  []uint64{8600},
		recommendations: []timestampedRecommendation{},
	}
	tc.runTest(t)
}

func TestScaleDownIncludeUnreadyPods(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            5,
		statusReplicas:          5,
		expectedDesiredReplicas: 2,
		CPUTarget:               50,
		CPUCurrent:              30,
		verifyCPUCurrent:        true,
		reportedLevels:          []uint64{100, 300, 500, 250, 250},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		useMetricsAPI:           true,
		reportedPodReadiness:    []v1.ConditionStatus{v1.ConditionTrue, v1.ConditionTrue, v1.ConditionTrue, v1.ConditionFalse, v1.ConditionFalse},
		recommendations:         []timestampedRecommendation{},
	}
	tc.runTest(t)
}

func TestScaleDownOneMetricInvalid(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            5,
		statusReplicas:          5,
		expectedDesiredReplicas: 3,
		CPUTarget:               50,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: "CheddarCheese",
			},
		},
		reportedLevels:      []uint64{100, 300, 500, 250, 250},
		reportedCPURequests: []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		useMetricsAPI:       true,
		recommendations:     []timestampedRecommendation{},
	}

	tc.runTest(t)
}

func TestScaleDownOneMetricEmpty(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            5,
		statusReplicas:          5,
		expectedDesiredReplicas: 3,
		CPUTarget:               50,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.ExternalMetricSourceType,
				External: &autoscalingv1alpha1.ExternalMetricSource{
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name:     "qps",
						Selector: &metav1.LabelSelector{},
					},
					Target: autoscalingv1alpha1.MetricTarget{
						Type:         autoscalingv1alpha1.AverageValueMetricType,
						AverageValue: resource.NewMilliQuantity(1000, resource.DecimalSI),
					},
				},
			},
		},
		reportedLevels:      []uint64{100, 300, 500, 250, 250},
		reportedCPURequests: []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		useMetricsAPI:       true,
		recommendations:     []timestampedRecommendation{},
	}
	_, _, _, testEMClient, _, _ := tc.prepareTestClient(t)
	testEMClient.PrependReactor("list", "*", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		return true, &emapi.ExternalMetricValueList{}, fmt.Errorf("something went wrong")
	})
	tc.testEMClient = testEMClient
	tc.runTest(t)
}

func TestScaleDownIgnoreHotCpuPods(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            5,
		statusReplicas:          5,
		expectedDesiredReplicas: 2,
		CPUTarget:               50,
		CPUCurrent:              30,
		verifyCPUCurrent:        true,
		reportedLevels:          []uint64{100, 300, 500, 250, 250},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		useMetricsAPI:           true,
		reportedPodStartTime:    []metav1.Time{coolCpuCreationTime(), coolCpuCreationTime(), coolCpuCreationTime(), hotCpuCreationTime(), hotCpuCreationTime()},
		recommendations:         []timestampedRecommendation{},
	}
	tc.runTest(t)
}

func TestScaleDownIgnoresFailedPods(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            5,
		statusReplicas:          5,
		expectedDesiredReplicas: 3,
		CPUTarget:               50,
		CPUCurrent:              28,
		verifyCPUCurrent:        true,
		reportedLevels:          []uint64{100, 300, 500, 250, 250},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		useMetricsAPI:           true,
		reportedPodReadiness:    []v1.ConditionStatus{v1.ConditionTrue, v1.ConditionTrue, v1.ConditionTrue, v1.ConditionTrue, v1.ConditionTrue, v1.ConditionFalse, v1.ConditionFalse},
		reportedPodPhase:        []v1.PodPhase{v1.PodRunning, v1.PodRunning, v1.PodRunning, v1.PodRunning, v1.PodRunning, v1.PodFailed, v1.PodFailed},
		recommendations:         []timestampedRecommendation{},
	}
	tc.runTest(t)
}

func TestScaleDownIgnoresDeletionPods(t *testing.T) {
	tc := testCase{
		minReplicas:                  2,
		maxReplicas:                  6,
		specReplicas:                 5,
		statusReplicas:               5,
		expectedDesiredReplicas:      3,
		CPUTarget:                    50,
		CPUCurrent:                   28,
		verifyCPUCurrent:             true,
		reportedLevels:               []uint64{100, 300, 500, 250, 250},
		reportedCPURequests:          []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		useMetricsAPI:                true,
		reportedPodReadiness:         []v1.ConditionStatus{v1.ConditionTrue, v1.ConditionTrue, v1.ConditionTrue, v1.ConditionTrue, v1.ConditionTrue, v1.ConditionFalse, v1.ConditionFalse},
		reportedPodPhase:             []v1.PodPhase{v1.PodRunning, v1.PodRunning, v1.PodRunning, v1.PodRunning, v1.PodRunning, v1.PodRunning, v1.PodRunning},
		reportedPodDeletionTimestamp: []bool{false, false, false, false, false, true, true},
		recommendations:              []timestampedRecommendation{},
	}
	tc.runTest(t)
}

func TestTolerance(t *testing.T) {
	tc := testCase{
		minReplicas:             1,
		maxReplicas:             5,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 3,
		CPUTarget:               100,
		reportedLevels:          []uint64{1010, 1030, 1020},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("0.9"), resource.MustParse("1.0"), resource.MustParse("1.1")},
		useMetricsAPI:           true,
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionTrue,
			Reason: "ReadyForNewScale",
		}),
	}
	tc.runTest(t)
}

func TestToleranceCM(t *testing.T) {
	averageValue := resource.MustParse("20.0")
	tc := testCase{
		minReplicas:             1,
		maxReplicas:             5,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 3,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.PodsMetricSourceType,
				Pods: &autoscalingv1alpha1.PodsMetricSource{
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name: "qps",
					},
					Target: autoscalingv1alpha1.MetricTarget{
						AverageValue: &averageValue,
					},
				},
			},
		},
		reportedLevels:      []uint64{20000, 20001, 21000},
		reportedCPURequests: []resource.Quantity{resource.MustParse("0.9"), resource.MustParse("1.0"), resource.MustParse("1.1")},
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionTrue,
			Reason: "ReadyForNewScale",
		}),
	}
	tc.runTest(t)
}

func TestToleranceCMObject(t *testing.T) {
	targetValue := resource.MustParse("20.0")
	tc := testCase{
		minReplicas:             1,
		maxReplicas:             5,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 3,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.ObjectMetricSourceType,
				Object: &autoscalingv1alpha1.ObjectMetricSource{
					DescribedObject: autoscalingv1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "some-deployment",
					},
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name: "qps",
					},
					Target: autoscalingv1alpha1.MetricTarget{
						Value: &targetValue,
					},
				},
			},
		},
		reportedLevels:      []uint64{20050},
		reportedCPURequests: []resource.Quantity{resource.MustParse("0.9"), resource.MustParse("1.0"), resource.MustParse("1.1")},
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionTrue,
			Reason: "ReadyForNewScale",
		}),
	}
	tc.runTest(t)
}

func TestToleranceCMExternal(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            4,
		statusReplicas:          4,
		expectedDesiredReplicas: 4,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.ExternalMetricSourceType,
				External: &autoscalingv1alpha1.ExternalMetricSource{
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name:     "qps",
						Selector: &metav1.LabelSelector{},
					},
					Target: autoscalingv1alpha1.MetricTarget{
						Value: resource.NewMilliQuantity(8666, resource.DecimalSI),
					},
				},
			},
		},
		reportedLevels: []uint64{8600},
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionTrue,
			Reason: "ReadyForNewScale",
		}),
	}
	tc.runTest(t)
}

func TestTolerancePerPodCMObject(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            4,
		statusReplicas:          4,
		expectedDesiredReplicas: 4,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.ObjectMetricSourceType,
				Object: &autoscalingv1alpha1.ObjectMetricSource{
					DescribedObject: autoscalingv1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "some-deployment",
					},
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name:     "qps",
						Selector: &metav1.LabelSelector{},
					},
					Target: autoscalingv1alpha1.MetricTarget{
						AverageValue: resource.NewMilliQuantity(2200, resource.DecimalSI),
					},
				},
			},
		},
		reportedLevels: []uint64{8600},
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionTrue,
			Reason: "ReadyForNewScale",
		}),
	}
	tc.runTest(t)
}

func TestTolerancePerPodCMExternal(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            4,
		statusReplicas:          4,
		expectedDesiredReplicas: 4,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.ExternalMetricSourceType,
				External: &autoscalingv1alpha1.ExternalMetricSource{
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name:     "qps",
						Selector: &metav1.LabelSelector{},
					},
					Target: autoscalingv1alpha1.MetricTarget{
						AverageValue: resource.NewMilliQuantity(2200, resource.DecimalSI),
					},
				},
			},
		},
		reportedLevels: []uint64{8600},
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionTrue,
			Reason: "ReadyForNewScale",
		}),
	}
	tc.runTest(t)
}

func TestMinReplicas(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             5,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 2,
		CPUTarget:               90,
		reportedLevels:          []uint64{10, 95, 10},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("0.9"), resource.MustParse("1.0"), resource.MustParse("1.1")},
		useMetricsAPI:           true,
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.ScalingLimited,
			Status: v1.ConditionTrue,
			Reason: "TooFewReplicas",
		}),
		recommendations: []timestampedRecommendation{},
	}
	tc.runTest(t)
}

func TestZeroMinReplicasDesiredZero(t *testing.T) {
	tc := testCase{
		minReplicas:             0,
		maxReplicas:             5,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 0,
		CPUTarget:               90,
		reportedLevels:          []uint64{0, 0, 0},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("0.9"), resource.MustParse("1.0"), resource.MustParse("1.1")},
		useMetricsAPI:           true,
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.ScalingLimited,
			Status: v1.ConditionFalse,
			Reason: "DesiredWithinRange",
		}),
		recommendations: []timestampedRecommendation{},
	}
	tc.runTest(t)
}

func TestMinReplicasDesiredZero(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             5,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 2,
		CPUTarget:               90,
		reportedLevels:          []uint64{0, 0, 0},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("0.9"), resource.MustParse("1.0"), resource.MustParse("1.1")},
		useMetricsAPI:           true,
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.ScalingLimited,
			Status: v1.ConditionTrue,
			Reason: "TooFewReplicas",
		}),
		recommendations: []timestampedRecommendation{},
	}
	tc.runTest(t)
}

func TestZeroReplicas(t *testing.T) {
	tc := testCase{
		minReplicas:             3,
		maxReplicas:             5,
		specReplicas:            0,
		statusReplicas:          0,
		expectedDesiredReplicas: 0,
		CPUTarget:               90,
		reportedLevels:          []uint64{},
		reportedCPURequests:     []resource.Quantity{},
		useMetricsAPI:           true,
		expectedConditions: []autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			{Type: autoscalingv1alpha1.AbleToScale, Status: v1.ConditionTrue, Reason: "SucceededGetScale"},
			{Type: autoscalingv1alpha1.ScalingActive, Status: v1.ConditionFalse, Reason: "ScalingDisabled"},
		},
	}
	tc.runTest(t)
}

func TestTooFewReplicas(t *testing.T) {
	tc := testCase{
		minReplicas:             3,
		maxReplicas:             5,
		specReplicas:            2,
		statusReplicas:          2,
		expectedDesiredReplicas: 3,
		CPUTarget:               90,
		reportedLevels:          []uint64{},
		reportedCPURequests:     []resource.Quantity{},
		useMetricsAPI:           true,
		expectedConditions: []autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			{Type: autoscalingv1alpha1.AbleToScale, Status: v1.ConditionTrue, Reason: "SucceededRescale"},
		},
	}
	tc.runTest(t)
}

func TestTooManyReplicas(t *testing.T) {
	tc := testCase{
		minReplicas:             3,
		maxReplicas:             5,
		specReplicas:            10,
		statusReplicas:          10,
		expectedDesiredReplicas: 5,
		CPUTarget:               90,
		reportedLevels:          []uint64{},
		reportedCPURequests:     []resource.Quantity{},
		useMetricsAPI:           true,
		expectedConditions: []autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			{Type: autoscalingv1alpha1.AbleToScale, Status: v1.ConditionTrue, Reason: "SucceededRescale"},
		},
	}
	tc.runTest(t)
}

func TestMaxReplicas(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             5,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 5,
		CPUTarget:               90,
		reportedLevels:          []uint64{8000, 9500, 1000},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("0.9"), resource.MustParse("1.0"), resource.MustParse("1.1")},
		useMetricsAPI:           true,
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.ScalingLimited,
			Status: v1.ConditionTrue,
			Reason: "TooManyReplicas",
		}),
	}
	tc.runTest(t)
}

func TestSuperfluousMetrics(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            4,
		statusReplicas:          4,
		expectedDesiredReplicas: 6,
		CPUTarget:               100,
		reportedLevels:          []uint64{4000, 9500, 3000, 7000, 3200, 2000},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		useMetricsAPI:           true,
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.ScalingLimited,
			Status: v1.ConditionTrue,
			Reason: "TooManyReplicas",
		}),
	}
	tc.runTest(t)
}

func TestMissingMetrics(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            4,
		statusReplicas:          4,
		expectedDesiredReplicas: 3,
		CPUTarget:               100,
		reportedLevels:          []uint64{400, 95},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		useMetricsAPI:           true,
		recommendations:         []timestampedRecommendation{},
	}
	tc.runTest(t)
}

func TestEmptyMetrics(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            4,
		statusReplicas:          4,
		expectedDesiredReplicas: 4,
		CPUTarget:               100,
		reportedLevels:          []uint64{},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		useMetricsAPI:           true,
		expectedConditions: []autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			{Type: autoscalingv1alpha1.AbleToScale, Status: v1.ConditionTrue, Reason: "SucceededGetScale"},
			{Type: autoscalingv1alpha1.ScalingActive, Status: v1.ConditionFalse, Reason: "FailedGetResourceMetric"},
		},
	}
	tc.runTest(t)
}

func TestEmptyCPURequest(t *testing.T) {
	tc := testCase{
		minReplicas:             1,
		maxReplicas:             5,
		specReplicas:            1,
		statusReplicas:          1,
		expectedDesiredReplicas: 1,
		CPUTarget:               100,
		reportedLevels:          []uint64{200},
		reportedCPURequests:     []resource.Quantity{},
		useMetricsAPI:           true,
		expectedConditions: []autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			{Type: autoscalingv1alpha1.AbleToScale, Status: v1.ConditionTrue, Reason: "SucceededGetScale"},
			{Type: autoscalingv1alpha1.ScalingActive, Status: v1.ConditionFalse, Reason: "FailedGetResourceMetric"},
		},
	}
	tc.runTest(t)
}

func TestEventCreated(t *testing.T) {
	tc := testCase{
		minReplicas:             1,
		maxReplicas:             5,
		specReplicas:            1,
		statusReplicas:          1,
		expectedDesiredReplicas: 2,
		CPUTarget:               50,
		reportedLevels:          []uint64{200},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("0.2")},
		verifyEvents:            true,
		useMetricsAPI:           true,
	}
	tc.runTest(t)
}

func TestEventNotCreated(t *testing.T) {
	tc := testCase{
		minReplicas:             1,
		maxReplicas:             5,
		specReplicas:            2,
		statusReplicas:          2,
		expectedDesiredReplicas: 2,
		CPUTarget:               50,
		reportedLevels:          []uint64{200, 200},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("0.4"), resource.MustParse("0.4")},
		verifyEvents:            true,
		useMetricsAPI:           true,
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionTrue,
			Reason: "ReadyForNewScale",
		}),
	}
	tc.runTest(t)
}

func TestMissingReports(t *testing.T) {
	tc := testCase{
		minReplicas:             1,
		maxReplicas:             5,
		specReplicas:            4,
		statusReplicas:          4,
		expectedDesiredReplicas: 2,
		CPUTarget:               50,
		reportedLevels:          []uint64{200},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("0.2")},
		useMetricsAPI:           true,
		recommendations:         []timestampedRecommendation{},
	}
	tc.runTest(t)
}

func TestUpscaleCap(t *testing.T) {
	tc := testCase{
		minReplicas:             1,
		maxReplicas:             100,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 24,
		CPUTarget:               10,
		reportedLevels:          []uint64{100, 200, 300},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("0.1"), resource.MustParse("0.1"), resource.MustParse("0.1")},
		useMetricsAPI:           true,
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.ScalingLimited,
			Status: v1.ConditionTrue,
			Reason: "ScaleUpLimit",
		}),
	}
	tc.runTest(t)
}

// TestUpscaleCapGreaterTMR 原方法名TestUpscaleCapGreaterThanMaxReplicas
//
//TestUpscaleCapGreaterTMR Test Up scale Cap Greater Than Max Replicas
func TestUpscaleCapGreaterTMR(t *testing.T) {
	tc := testCase{
		minReplicas:    1,
		maxReplicas:    20,
		specReplicas:   3,
		statusReplicas: 3,
		// expectedDesiredReplicas would be 24 without maxReplicas
		expectedDesiredReplicas: 20,
		CPUTarget:               10,
		reportedLevels:          []uint64{100, 200, 300},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("0.1"), resource.MustParse("0.1"), resource.MustParse("0.1")},
		useMetricsAPI:           true,
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.ScalingLimited,
			Status: v1.ConditionTrue,
			Reason: "TooManyReplicas",
		}),
	}
	tc.runTest(t)
}

func TestMoreReplicasThanSpecNoScale(t *testing.T) {
	tc := testCase{
		minReplicas:             1,
		maxReplicas:             8,
		specReplicas:            4,
		statusReplicas:          5, // Deployment update with 25% surge.
		expectedDesiredReplicas: 4,
		CPUTarget:               50,
		reportedLevels:          []uint64{500, 500, 500, 500, 500},
		reportedCPURequests: []resource.Quantity{
			resource.MustParse("1"),
			resource.MustParse("1"),
			resource.MustParse("1"),
			resource.MustParse("1"),
			resource.MustParse("1"),
		},
		useMetricsAPI: true,
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionTrue,
			Reason: "ReadyForNewScale",
		}),
	}
	tc.runTest(t)
}

func TestConditionInvalidSelectorMissing(t *testing.T) {
	tc := testCase{
		minReplicas:             1,
		maxReplicas:             100,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 3,
		CPUTarget:               10,
		reportedLevels:          []uint64{100, 200, 300},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("0.1"), resource.MustParse("0.1"), resource.MustParse("0.1")},
		useMetricsAPI:           true,
		expectedConditions: []autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			{
				Type:   autoscalingv1alpha1.AbleToScale,
				Status: v1.ConditionTrue,
				Reason: "SucceededGetScale",
			},
			{
				Type:   autoscalingv1alpha1.ScalingActive,
				Status: v1.ConditionFalse,
				Reason: "InvalidSelector",
			},
		},
	}

	_, _, _, _, testScaleClient, _ := tc.prepareTestClient(t)
	tc.testScaleClient = testScaleClient

	testScaleClient.PrependReactor("get", "replicationcontrollers", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		obj := &autoscalinginternal.Scale{
			ObjectMeta: metav1.ObjectMeta{
				Name: tc.resource.name,
			},
			Spec: autoscalinginternal.ScaleSpec{
				Replicas: tc.specReplicas,
			},
			Status: autoscalinginternal.ScaleStatus{
				Replicas: tc.specReplicas,
			},
		}
		return true, obj, nil
	})

	tc.runTest(t)
}

// TestConditionInvalidSU 原方法名TestConditionInvalidSelectorUnparsable
//
//TestConditionInvalidSU Test Condition Invalid Selector Unparsable
func TestConditionInvalidSU(t *testing.T) {
	tc := testCase{
		minReplicas:             1,
		maxReplicas:             100,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 3,
		CPUTarget:               10,
		reportedLevels:          []uint64{100, 200, 300},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("0.1"), resource.MustParse("0.1"), resource.MustParse("0.1")},
		useMetricsAPI:           true,
		expectedConditions: []autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			{
				Type:   autoscalingv1alpha1.AbleToScale,
				Status: v1.ConditionTrue,
				Reason: "SucceededGetScale",
			},
			{
				Type:   autoscalingv1alpha1.ScalingActive,
				Status: v1.ConditionFalse,
				Reason: "InvalidSelector",
			},
		},
	}

	_, _, _, _, testScaleClient, _ := tc.prepareTestClient(t)
	tc.testScaleClient = testScaleClient

	testScaleClient.PrependReactor("get", "replicationcontrollers", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		obj := &autoscalinginternal.Scale{
			ObjectMeta: metav1.ObjectMeta{
				Name: tc.resource.name,
			},
			Spec: autoscalinginternal.ScaleSpec{
				Replicas: tc.specReplicas,
			},
			Status: autoscalinginternal.ScaleStatus{
				Replicas: tc.specReplicas,
				Selector: "cheddar=cheese",
			},
		}
		return true, obj, nil
	})

	tc.runTest(t)
}

func TestConditionFailedGetMetrics(t *testing.T) {
	targetValue := resource.MustParse("15.0")
	averageValue := resource.MustParse("15.0")
	metricsTargets := map[string][]autoscalingv1alpha1.MetricSpec{
		"FailedGetResourceMetric": nil,
		"FailedGetPodsMetric": {
			{
				Type: autoscalingv1alpha1.PodsMetricSourceType,
				Pods: &autoscalingv1alpha1.PodsMetricSource{
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name: "qps",
					},
					Target: autoscalingv1alpha1.MetricTarget{
						AverageValue: &averageValue,
					},
				},
			},
		},
		"FailedGetObjectMetric": {
			{
				Type: autoscalingv1alpha1.ObjectMetricSourceType,
				Object: &autoscalingv1alpha1.ObjectMetricSource{
					DescribedObject: autoscalingv1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "some-deployment",
					},
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name: "qps",
					},
					Target: autoscalingv1alpha1.MetricTarget{
						Value: &targetValue,
					},
				},
			},
		},
		"FailedGetExternalMetric": {
			{
				Type: autoscalingv1alpha1.ExternalMetricSourceType,
				External: &autoscalingv1alpha1.ExternalMetricSource{
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name:     "qps",
						Selector: &metav1.LabelSelector{},
					},
					Target: autoscalingv1alpha1.MetricTarget{
						Value: resource.NewMilliQuantity(300, resource.DecimalSI),
					},
				},
			},
		},
	}

	for reason, specs := range metricsTargets {
		tc := testCase{
			minReplicas:             1,
			maxReplicas:             100,
			specReplicas:            3,
			statusReplicas:          3,
			expectedDesiredReplicas: 3,
			CPUTarget:               10,
			reportedLevels:          []uint64{100, 200, 300},
			reportedCPURequests:     []resource.Quantity{resource.MustParse("0.1"), resource.MustParse("0.1"), resource.MustParse("0.1")},
			useMetricsAPI:           true,
		}
		_, testMetricsClient, testCMClient, testEMClient, _, _ := tc.prepareTestClient(t)
		tc.testMetricsClient = testMetricsClient
		tc.testCMClient = testCMClient
		tc.testEMClient = testEMClient

		testMetricsClient.PrependReactor("list", "pods", func(action core.Action) (handled bool, ret runtime.Object, err error) {
			return true, &metricsapi.PodMetricsList{}, fmt.Errorf("something went wrong")
		})
		testCMClient.PrependReactor("get", "*", func(action core.Action) (handled bool, ret runtime.Object, err error) {
			return true, &cmapi.MetricValueList{}, fmt.Errorf("something went wrong")
		})
		testEMClient.PrependReactor("list", "*", func(action core.Action) (handled bool, ret runtime.Object, err error) {
			return true, &emapi.ExternalMetricValueList{}, fmt.Errorf("something went wrong")
		})

		tc.expectedConditions = []autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			{Type: autoscalingv1alpha1.AbleToScale, Status: v1.ConditionTrue, Reason: "SucceededGetScale"},
			{Type: autoscalingv1alpha1.ScalingActive, Status: v1.ConditionFalse, Reason: reason},
		}
		if specs != nil {
			tc.CPUTarget = 0
		} else {
			tc.CPUTarget = 10
		}
		tc.metricsTarget = specs
		tc.runTest(t)
	}
}

func TestConditionInvalidSourceType(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 3,
		CPUTarget:               0,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: "CheddarCheese",
			},
		},
		reportedLevels: []uint64{20000},
		expectedConditions: []autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			{
				Type:   autoscalingv1alpha1.AbleToScale,
				Status: v1.ConditionTrue,
				Reason: "SucceededGetScale",
			},
			{
				Type:   autoscalingv1alpha1.ScalingActive,
				Status: v1.ConditionFalse,
				Reason: "InvalidMetricSourceType",
			},
		},
	}
	tc.runTest(t)
}

func TestConditionFailedGetScale(t *testing.T) {
	tc := testCase{
		minReplicas:             1,
		maxReplicas:             100,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 3,
		CPUTarget:               10,
		reportedLevels:          []uint64{100, 200, 300},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("0.1"), resource.MustParse("0.1"), resource.MustParse("0.1")},
		useMetricsAPI:           true,
		expectedConditions: []autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			{
				Type:   autoscalingv1alpha1.AbleToScale,
				Status: v1.ConditionFalse,
				Reason: "FailedGetScale",
			},
		},
	}

	_, _, _, _, testScaleClient, _ := tc.prepareTestClient(t)
	tc.testScaleClient = testScaleClient

	testScaleClient.PrependReactor("get", "replicationcontrollers", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		return true, &autoscalinginternal.Scale{}, fmt.Errorf("something went wrong")
	})

	tc.runTest(t)
}

func TestConditionFailedUpdateScale(t *testing.T) {
	tc := testCase{
		minReplicas:             1,
		maxReplicas:             5,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 3,
		CPUTarget:               100,
		reportedLevels:          []uint64{150, 150, 150},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("0.1"), resource.MustParse("0.1"), resource.MustParse("0.1")},
		useMetricsAPI:           true,
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionFalse,
			Reason: "FailedUpdateScale",
		}),
	}

	_, _, _, _, testScaleClient, _ := tc.prepareTestClient(t)
	tc.testScaleClient = testScaleClient

	testScaleClient.PrependReactor("update", "replicationcontrollers", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		return true, &autoscalinginternal.Scale{}, fmt.Errorf("something went wrong")
	})

	tc.runTest(t)
}

func NoTestBackoffUpscale(t *testing.T) {
	time2 := metav1.Time{Time: time.Now()}
	tc := testCase{
		minReplicas:             1,
		maxReplicas:             5,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 3,
		CPUTarget:               100,
		reportedLevels:          []uint64{150, 150, 150},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("0.1"), resource.MustParse("0.1"), resource.MustParse("0.1")},
		useMetricsAPI:           true,
		lastScaleTime:           &time2,
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionTrue,
			Reason: "ReadyForNewScale",
		}, autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionTrue,
			Reason: "SucceededRescale",
		}),
	}
	tc.runTest(t)
}

func TestNoBackoffUpscaleCM(t *testing.T) {
	averageValue := resource.MustParse("15.0")
	time2 := metav1.Time{Time: time.Now()}
	tc := testCase{
		minReplicas:             1,
		maxReplicas:             5,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 4,
		CPUTarget:               0,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.PodsMetricSourceType,
				Pods: &autoscalingv1alpha1.PodsMetricSource{
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name: "qps",
					},
					Target: autoscalingv1alpha1.MetricTarget{
						AverageValue: &averageValue,
					},
				},
			},
		},
		reportedLevels:      []uint64{20000, 10000, 30000},
		reportedCPURequests: []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		//useMetricsAPI:       true,
		lastScaleTime: &time2,
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionTrue,
			Reason: "ReadyForNewScale",
		}, autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionTrue,
			Reason: "SucceededRescale",
		}, autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.ScalingLimited,
			Status: v1.ConditionFalse,
			Reason: "DesiredWithinRange",
		}),
	}
	tc.runTest(t)
}

func TestNoBackoffUpscaleCMNoBackoffCpu(t *testing.T) {
	averageValue := resource.MustParse("15.0")
	time2 := metav1.Time{Time: time.Now()}
	tc := testCase{
		minReplicas:             1,
		maxReplicas:             5,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 5,
		CPUTarget:               10,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.PodsMetricSourceType,
				Pods: &autoscalingv1alpha1.PodsMetricSource{
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name: "qps",
					},
					Target: autoscalingv1alpha1.MetricTarget{
						AverageValue: &averageValue,
					},
				},
			},
		},
		reportedLevels:      []uint64{20000, 10000, 30000},
		reportedCPURequests: []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		useMetricsAPI:       true,
		lastScaleTime:       &time2,
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionTrue,
			Reason: "ReadyForNewScale",
		}, autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionTrue,
			Reason: "SucceededRescale",
		}, autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.ScalingLimited,
			Status: v1.ConditionTrue,
			Reason: "TooManyReplicas",
		}),
	}
	tc.runTest(t)
}

func TestStabilizeDownscale(t *testing.T) {
	tc := testCase{
		minReplicas:             1,
		maxReplicas:             5,
		specReplicas:            4,
		statusReplicas:          4,
		expectedDesiredReplicas: 4,
		CPUTarget:               100,
		reportedLevels:          []uint64{50, 50, 50},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("0.1"), resource.MustParse("0.1"), resource.MustParse("0.1")},
		useMetricsAPI:           true,
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionTrue,
			Reason: "ReadyForNewScale",
		}, autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionTrue,
			Reason: "ScaleDownStabilized",
		}),
		recommendations: []timestampedRecommendation{
			{10, time.Now().Add(-10 * time.Minute)},
			{4, time.Now().Add(-1 * time.Minute)},
		},
	}
	tc.runTest(t)
}

// TestComputedToleranceAI 原方法名 TestComputedToleranceAlgImplementation
//
// TestComputedToleranceAI is a regression test which
// back-calculates a minimal percentage for downscaling based on a small percentage
// increase in pod utilization which is calibrated against the tolerance value.
func TestComputedToleranceAI(t *testing.T) {

	startPods := int32(10)
	// 150 mCPU per pod.
	totalUsedCPUOfAllPods := uint64(startPods * 150)
	// Each pod starts out asking for 2X what is really needed.
	// This means we will have a 50% ratio of used/requested
	totalRequestedCPUOfAllPods := int32(2 * totalUsedCPUOfAllPods)
	requestedToUsed := float64(totalRequestedCPUOfAllPods / int32(totalUsedCPUOfAllPods))
	// Spread the amount we ask over 10 pods.  We can add some jitter later in reportedLevels.
	perPodRequested := totalRequestedCPUOfAllPods / startPods

	// Force a minimal scaling event by satisfying  (tolerance < 1 - resourcesUsedRatio).
	target := math.Abs(1/(requestedToUsed*(1-defaultTestingTolerance))) + .01
	finalCPUPercentTarget := int32(target * 100)
	resourcesUsedRatio := float64(totalUsedCPUOfAllPods) / float64(float64(totalRequestedCPUOfAllPods)*target)

	// i.e. .60 * 20 -> scaled down expectation.
	finalPods := int32(math.Ceil(resourcesUsedRatio * float64(startPods)))

	// To breach tolerance we will create a utilization ratio difference of tolerance to usageRatioToleranceValue)
	tc1 := testCase{
		minReplicas:             0,
		maxReplicas:             1000,
		specReplicas:            startPods,
		statusReplicas:          startPods,
		expectedDesiredReplicas: finalPods,
		CPUTarget:               finalCPUPercentTarget,
		reportedLevels: []uint64{
			totalUsedCPUOfAllPods / 10,
			totalUsedCPUOfAllPods / 10,
			totalUsedCPUOfAllPods / 10,
			totalUsedCPUOfAllPods / 10,
			totalUsedCPUOfAllPods / 10,
			totalUsedCPUOfAllPods / 10,
			totalUsedCPUOfAllPods / 10,
			totalUsedCPUOfAllPods / 10,
			totalUsedCPUOfAllPods / 10,
			totalUsedCPUOfAllPods / 10,
		},
		reportedCPURequests: []resource.Quantity{
			resource.MustParse(fmt.Sprint(perPodRequested+100) + "m"),
			resource.MustParse(fmt.Sprint(perPodRequested-100) + "m"),
			resource.MustParse(fmt.Sprint(perPodRequested+10) + "m"),
			resource.MustParse(fmt.Sprint(perPodRequested-10) + "m"),
			resource.MustParse(fmt.Sprint(perPodRequested+2) + "m"),
			resource.MustParse(fmt.Sprint(perPodRequested-2) + "m"),
			resource.MustParse(fmt.Sprint(perPodRequested+1) + "m"),
			resource.MustParse(fmt.Sprint(perPodRequested-1) + "m"),
			resource.MustParse(fmt.Sprint(perPodRequested) + "m"),
			resource.MustParse(fmt.Sprint(perPodRequested) + "m"),
		},
		useMetricsAPI:   true,
		recommendations: []timestampedRecommendation{},
	}
	tc1.runTest(t)

	target = math.Abs(1/(requestedToUsed*(1-defaultTestingTolerance))) + .004
	finalCPUPercentTarget = int32(target * 100)
	tc2 := testCase{
		minReplicas:             0,
		maxReplicas:             1000,
		specReplicas:            startPods,
		statusReplicas:          startPods,
		expectedDesiredReplicas: startPods,
		CPUTarget:               finalCPUPercentTarget,
		reportedLevels: []uint64{
			totalUsedCPUOfAllPods / 10,
			totalUsedCPUOfAllPods / 10,
			totalUsedCPUOfAllPods / 10,
			totalUsedCPUOfAllPods / 10,
			totalUsedCPUOfAllPods / 10,
			totalUsedCPUOfAllPods / 10,
			totalUsedCPUOfAllPods / 10,
			totalUsedCPUOfAllPods / 10,
			totalUsedCPUOfAllPods / 10,
			totalUsedCPUOfAllPods / 10,
		},
		reportedCPURequests: []resource.Quantity{
			resource.MustParse(fmt.Sprint(perPodRequested+100) + "m"),
			resource.MustParse(fmt.Sprint(perPodRequested-100) + "m"),
			resource.MustParse(fmt.Sprint(perPodRequested+10) + "m"),
			resource.MustParse(fmt.Sprint(perPodRequested-10) + "m"),
			resource.MustParse(fmt.Sprint(perPodRequested+2) + "m"),
			resource.MustParse(fmt.Sprint(perPodRequested-2) + "m"),
			resource.MustParse(fmt.Sprint(perPodRequested+1) + "m"),
			resource.MustParse(fmt.Sprint(perPodRequested-1) + "m"),
			resource.MustParse(fmt.Sprint(perPodRequested) + "m"),
			resource.MustParse(fmt.Sprint(perPodRequested) + "m"),
		},
		useMetricsAPI:   true,
		recommendations: []timestampedRecommendation{},
		expectedConditions: statusOkWithOverrides(autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			Type:   autoscalingv1alpha1.AbleToScale,
			Status: v1.ConditionTrue,
			Reason: "ReadyForNewScale",
		}),
	}
	tc2.runTest(t)
}

func TestScaleUpRCImmediately(t *testing.T) {
	time2 := metav1.Time{Time: time.Now()}
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            1,
		statusReplicas:          1,
		expectedDesiredReplicas: 2,
		verifyCPUCurrent:        false,
		reportedLevels:          []uint64{0, 0, 0, 0},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		useMetricsAPI:           true,
		lastScaleTime:           &time2,
		expectedConditions: []autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			{Type: autoscalingv1alpha1.AbleToScale, Status: v1.ConditionTrue, Reason: "SucceededRescale"},
		},
	}
	tc.runTest(t)
}

func TestScaleDownRCImmediately(t *testing.T) {
	time2 := metav1.Time{Time: time.Now()}
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             5,
		specReplicas:            6,
		statusReplicas:          6,
		expectedDesiredReplicas: 5,
		CPUTarget:               50,
		reportedLevels:          []uint64{8000, 9500, 1000},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("0.9"), resource.MustParse("1.0"), resource.MustParse("1.1")},
		useMetricsAPI:           true,
		lastScaleTime:           &time2,
		expectedConditions: []autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			{Type: autoscalingv1alpha1.AbleToScale, Status: v1.ConditionTrue, Reason: "SucceededRescale"},
		},
	}
	tc.runTest(t)
}

func TestAvoidUncessaryUpdates(t *testing.T) {
	now := metav1.Time{Time: time.Now().Add(-time.Hour)}
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            2,
		statusReplicas:          2,
		expectedDesiredReplicas: 2,
		CPUTarget:               30,
		CPUCurrent:              40,
		verifyCPUCurrent:        true,
		reportedLevels:          []uint64{400, 500, 700},
		reportedCPURequests:     []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		reportedPodStartTime:    []metav1.Time{coolCpuCreationTime(), hotCpuCreationTime(), hotCpuCreationTime()},
		useMetricsAPI:           true,
		lastScaleTime:           &now,
		recommendations:         []timestampedRecommendation{},
	}
	_, _, _, _, _, gpaClient := tc.prepareTestClient(t)
	tc.testGpaClient = gpaClient
	gpaClient.PrependReactor("list", "generalpodautoscalers", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		tc.Lock()
		defer tc.Unlock()
		// fake out the verification logic and mark that we're done processing
		go func() {
			// wait a tick and then mark that we're finished (otherwise, we have no
			// way to indicate that we're finished, because the function decides not to do anything)
			time.Sleep(1 * time.Second)
			tc.Lock()
			tc.statusUpdated = true
			tc.Unlock()
			tc.processed <- "test-gpa"
		}()

		quantity := resource.MustParse("400m")
		obj := &autoscalingv1alpha1.GeneralPodAutoscalerList{
			Items: []autoscalingv1alpha1.GeneralPodAutoscaler{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-gpa",
						Namespace: "test-namespace",
					},
					Spec: autoscalingv1alpha1.GeneralPodAutoscalerSpec{
						ScaleTargetRef: autoscalingv1alpha1.CrossVersionObjectReference{
							Kind:       "ReplicationController",
							Name:       "test-rc",
							APIVersion: "v1",
						},

						MinReplicas: &tc.minReplicas,
						MaxReplicas: tc.maxReplicas,
					},
					Status: autoscalingv1alpha1.GeneralPodAutoscalerStatus{
						CurrentReplicas: tc.specReplicas,
						DesiredReplicas: tc.specReplicas,
						LastScaleTime:   tc.lastScaleTime,
						CurrentMetrics: []autoscalingv1alpha1.MetricStatus{
							{
								Type: autoscalingv1alpha1.ResourceMetricSourceType,
								Resource: &autoscalingv1alpha1.ResourceMetricStatus{
									Name: v1.ResourceCPU,
									Current: autoscalingv1alpha1.MetricValueStatus{
										AverageValue:       &quantity,
										AverageUtilization: &tc.CPUCurrent,
									},
								},
							},
						},
						Conditions: []autoscalingv1alpha1.GeneralPodAutoscalerCondition{
							{
								Type:               autoscalingv1alpha1.AbleToScale,
								Status:             v1.ConditionTrue,
								LastTransitionTime: *tc.lastScaleTime,
								Reason:             "ReadyForNewScale",
								Message:            "recommended size matches current size",
							},
							{
								Type:               autoscalingv1alpha1.ScalingActive,
								Status:             v1.ConditionTrue,
								LastTransitionTime: *tc.lastScaleTime,
								Reason:             "ValidMetricFound",
								Message:            "the GPA was able to successfully calculate a replica count from cpu resource utilization (percentage of request)",
							},
							{
								Type:               autoscalingv1alpha1.ScalingLimited,
								Status:             v1.ConditionTrue,
								LastTransitionTime: *tc.lastScaleTime,
								Reason:             "TooFewReplicas",
								Message:            "the desired replica count is less than the minimum replica count",
							},
						},
					},
				},
			},
		}
		return true, obj, nil
	})
	gpaClient.PrependReactor("update", "generalpodautoscalers", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		assert.Fail(t, "should not have attempted to update the GPA when nothing changed")
		// mark that we've processed this GPA
		tc.processed <- ""
		return true, nil, fmt.Errorf("unexpected call")
	})

	controller, informerFactory, scalerFactory := tc.setupController(t)
	tc.runTestWithController(t, controller, informerFactory, scalerFactory)
}

func TestConvertDesiredReplicasWithRules(t *testing.T) {
	conversionTestCases := []struct {
		currentReplicas                  int32
		expectedDesiredReplicas          int32
		gpaMinReplicas                   int32
		gpaMaxReplicas                   int32
		expectedConvertedDesiredReplicas int32
		expectedCondition                string
		annotation                       string
	}{
		{
			currentReplicas:                  5,
			expectedDesiredReplicas:          7,
			gpaMinReplicas:                   3,
			gpaMaxReplicas:                   8,
			expectedConvertedDesiredReplicas: 7,
			expectedCondition:                "DesiredWithinRange",
			annotation:                       "prenormalized desired replicas within range",
		},
		{
			currentReplicas:                  3,
			expectedDesiredReplicas:          1,
			gpaMinReplicas:                   2,
			gpaMaxReplicas:                   8,
			expectedConvertedDesiredReplicas: 2,
			expectedCondition:                "TooFewReplicas",
			annotation:                       "prenormalized desired replicas < minReplicas",
		},
		{
			currentReplicas:                  1,
			expectedDesiredReplicas:          0,
			gpaMinReplicas:                   0,
			gpaMaxReplicas:                   10,
			expectedConvertedDesiredReplicas: 0,
			expectedCondition:                "DesiredWithinRange",
			annotation:                       "prenormalized desired replicas within range",
		},
		{
			currentReplicas:                  20,
			expectedDesiredReplicas:          1000,
			gpaMinReplicas:                   1,
			gpaMaxReplicas:                   10,
			expectedConvertedDesiredReplicas: 10,
			expectedCondition:                "TooManyReplicas",
			annotation:                       "maxReplicas is the limit because maxReplicas < scaleUpLimit",
		},
		{
			currentReplicas:                  3,
			expectedDesiredReplicas:          1000,
			gpaMinReplicas:                   1,
			gpaMaxReplicas:                   2000,
			expectedConvertedDesiredReplicas: calculateScaleUpLimit(3),
			expectedCondition:                "ScaleUpLimit",
			annotation:                       "scaleUpLimit is the limit because scaleUpLimit < maxReplicas",
		},
	}

	for _, ctc := range conversionTestCases {
		actualConvertedDesiredReplicas, actualCondition, _ := convertDesiredReplicasWithRules(
			ctc.currentReplicas, ctc.expectedDesiredReplicas, ctc.gpaMinReplicas, ctc.gpaMaxReplicas,
		)

		assert.Equal(t, ctc.expectedConvertedDesiredReplicas, actualConvertedDesiredReplicas, ctc.annotation)
		assert.Equal(t, ctc.expectedCondition, actualCondition, ctc.annotation)
	}
}

func TestNormalizeDesiredReplicas(t *testing.T) {
	tests := []struct {
		name                         string
		key                          string
		recommendations              []timestampedRecommendation
		prenormalizedDesiredReplicas int32
		expectedStabilizedReplicas   int32
		expectedLogLength            int
	}{
		{
			"empty log",
			"",
			[]timestampedRecommendation{},
			5,
			5,
			1,
		},
		{
			"stabilize",
			"",
			[]timestampedRecommendation{
				{4, time.Now().Add(-2 * time.Minute)},
				{5, time.Now().Add(-1 * time.Minute)},
			},
			3,
			5,
			3,
		},
		{
			"no stabilize",
			"",
			[]timestampedRecommendation{
				{1, time.Now().Add(-2 * time.Minute)},
				{2, time.Now().Add(-1 * time.Minute)},
			},
			3,
			3,
			3,
		},
		{
			"no stabilize - old recommendations",
			"",
			[]timestampedRecommendation{
				{10, time.Now().Add(-10 * time.Minute)},
				{9, time.Now().Add(-9 * time.Minute)},
			},
			3,
			3,
			2,
		},
		{
			"stabilize - old recommendations",
			"",
			[]timestampedRecommendation{
				{10, time.Now().Add(-10 * time.Minute)},
				{4, time.Now().Add(-1 * time.Minute)},
				{5, time.Now().Add(-2 * time.Minute)},
				{9, time.Now().Add(-9 * time.Minute)},
			},
			3,
			5,
			4,
		},
	}
	for _, tc := range tests {
		hc := GeneralController{
			downscaleStabilisationWindow: 5 * time.Minute,
			recommendations: map[string][]timestampedRecommendation{
				tc.key: tc.recommendations,
			},
		}
		r := hc.stabilizeRecommendation(tc.key, tc.prenormalizedDesiredReplicas)
		if r != tc.expectedStabilizedReplicas {
			t.Errorf("[%s] got %d stabilized replicas, expected %d", tc.name, r, tc.expectedStabilizedReplicas)
		}
		if len(hc.recommendations[tc.key]) != tc.expectedLogLength {
			t.Errorf("[%s] after  stabilization recommendations log has %d entries, expected %d", tc.name, len(hc.recommendations[tc.key]), tc.expectedLogLength)
		}
	}
}

func TestScaleUpOneMetricEmpty(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            3,
		statusReplicas:          3,
		expectedDesiredReplicas: 4,
		CPUTarget:               30,
		verifyCPUCurrent:        true,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.ExternalMetricSourceType,
				External: &autoscalingv1alpha1.ExternalMetricSource{
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name:     "qps",
						Selector: &metav1.LabelSelector{},
					},
					Target: autoscalingv1alpha1.MetricTarget{
						Value: resource.NewMilliQuantity(100, resource.DecimalSI),
					},
				},
			},
		},
		reportedLevels:      []uint64{300, 400, 500},
		reportedCPURequests: []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
	}
	_, _, _, testEMClient, _, _ := tc.prepareTestClient(t)
	testEMClient.PrependReactor("list", "*", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		return true, &emapi.ExternalMetricValueList{}, fmt.Errorf("something went wrong")
	})
	tc.testEMClient = testEMClient
	tc.runTest(t)
}

func TestNoScaleDownOneMetricInvalid(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            5,
		statusReplicas:          5,
		expectedDesiredReplicas: 5,
		CPUTarget:               50,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: "CheddarCheese",
			},
		},
		reportedLevels:      []uint64{100, 300, 500, 250, 250},
		reportedCPURequests: []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		useMetricsAPI:       true,
		expectedConditions: []autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			{Type: autoscalingv1alpha1.AbleToScale, Status: v1.ConditionTrue, Reason: "ScaleDownStabilized"},
			{Type: autoscalingv1alpha1.ScalingActive, Status: v1.ConditionTrue, Reason: "ValidMetricFound"},
			{Type: autoscalingv1alpha1.ScalingLimited, Status: v1.ConditionFalse, Reason: "DesiredWithinRange"},
		},
	}

	tc.runTest(t)
}

func TestNoScaleDownOneMetricEmpty(t *testing.T) {
	tc := testCase{
		minReplicas:             2,
		maxReplicas:             6,
		specReplicas:            5,
		statusReplicas:          5,
		expectedDesiredReplicas: 5,
		CPUTarget:               50,
		metricsTarget: []autoscalingv1alpha1.MetricSpec{
			{
				Type: autoscalingv1alpha1.ExternalMetricSourceType,
				External: &autoscalingv1alpha1.ExternalMetricSource{
					Metric: autoscalingv1alpha1.MetricIdentifier{
						Name:     "qps",
						Selector: &metav1.LabelSelector{},
					},
					Target: autoscalingv1alpha1.MetricTarget{
						Value: resource.NewMilliQuantity(1000, resource.DecimalSI),
					},
				},
			},
		},
		reportedLevels:      []uint64{100, 300, 500, 250, 250},
		reportedCPURequests: []resource.Quantity{resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0"), resource.MustParse("1.0")},
		useMetricsAPI:       true,
		expectedConditions: []autoscalingv1alpha1.GeneralPodAutoscalerCondition{
			{Type: autoscalingv1alpha1.AbleToScale, Status: v1.ConditionTrue, Reason: "ScaleDownStabilized"},
			{Type: autoscalingv1alpha1.ScalingActive, Status: v1.ConditionTrue, Reason: "ValidMetricFound"},
			{Type: autoscalingv1alpha1.ScalingLimited, Status: v1.ConditionFalse, Reason: "DesiredWithinRange"},
		},
	}
	_, _, _, testEMClient, _, _ := tc.prepareTestClient(t)
	testEMClient.PrependReactor("list", "*", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		return true, &emapi.ExternalMetricValueList{}, fmt.Errorf("something went wrong")
	})
	tc.testEMClient = testEMClient
	tc.runTest(t)
}

func testScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	s.AddKnownTypes(schema.GroupVersion{
		Group:   autoscaling.GroupName,
		Version: "v1alpha1",
	}, &autoscalingv1alpha1.GeneralPodAutoscaler{}, &autoscalingv1alpha1.GeneralPodAutoscalerList{})
	s.AddKnownTypes(schema.GroupVersion{
		Group:   "apps",
		Version: "v1",
	}, &appsv1.Deployment{}, &appsv1.DeploymentList{}, &appsv1.ReplicaSet{}, &appsv1.ReplicaSetList{})
	s.AddKnownTypes(schema.GroupVersion{
		Group:   "",
		Version: "v1",
	}, &v1.Pod{}, &v1.PodList{}, &v1.Event{}, &v1.EventList{}, &v1.ReplicationController{}, &v1.ReplicationControllerList{})
	return s
}
