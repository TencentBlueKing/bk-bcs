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

package core

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"

	mockprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/mocks"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes/fake"
	v1appslister "k8s.io/client-go/listers/apps/v1"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/klog"
)

type podListerMock struct {
	mock.Mock
}

func (l *podListerMock) List() ([]*apiv1.Pod, error) {
	args := l.Called()
	return args.Get(0).([]*apiv1.Pod), args.Error(1)
}

type podDisruptionBudgetListerMock struct {
	mock.Mock
}

func (l *podDisruptionBudgetListerMock) List() ([]*policyv1.PodDisruptionBudget, error) {
	args := l.Called()
	return args.Get(0).([]*policyv1.PodDisruptionBudget), args.Error(1)
}

type daemonSetListerMock struct {
	mock.Mock
}

func (l *daemonSetListerMock) List(selector labels.Selector) ([]*appsv1.DaemonSet, error) {
	args := l.Called(selector)
	return args.Get(0).([]*appsv1.DaemonSet), args.Error(1)
}

func (l *daemonSetListerMock) DaemonSets(namespace string) v1appslister.DaemonSetNamespaceLister {
	args := l.Called(namespace)
	return args.Get(0).(v1appslister.DaemonSetNamespaceLister)
}

func (l *daemonSetListerMock) GetPodDaemonSets(pod *apiv1.Pod) ([]*appsv1.DaemonSet, error) {
	args := l.Called()
	return args.Get(0).([]*appsv1.DaemonSet), args.Error(1)
}

func (l *daemonSetListerMock) GetHistoryDaemonSets(history *appsv1.ControllerRevision) ([]*appsv1.DaemonSet, error) {
	args := l.Called()
	return args.Get(0).([]*appsv1.DaemonSet), args.Error(1)
}

type onScaleUpMock struct {
	mock.Mock
}

func (m *onScaleUpMock) ScaleUp(id string, delta int) error {
	klog.Infof("Scale up: %v %v", id, delta)
	args := m.Called(id, delta)
	return args.Error(0)
}

type onScaleDownMock struct {
	mock.Mock
}

func (m *onScaleDownMock) ScaleDown(id string, name string) error {
	klog.Infof("Scale down: %v %v", id, name)
	args := m.Called(id, name)
	return args.Error(0)
}

type onNodeGroupCreateMock struct {
	mock.Mock
}

func (m *onNodeGroupCreateMock) Create(id string) error {
	klog.Infof("Create group: %v", id)
	args := m.Called(id)
	return args.Error(0)
}

type onNodeGroupDeleteMock struct {
	mock.Mock
}

func (m *onNodeGroupDeleteMock) Delete(id string) error {
	klog.Infof("Delete group: %v", id)
	args := m.Called(id)
	return args.Error(0)
}

func TestBufferedAutoscalerRunOnce(t *testing.T) {
	readyNodeLister := kubernetes.NewTestNodeLister(nil)
	allNodeLister := kubernetes.NewTestNodeLister(nil)
	scheduledPodMock := &podListerMock{}
	unschedulablePodMock := &podListerMock{}
	podDisruptionBudgetListerMock := &podDisruptionBudgetListerMock{}
	daemonSetListerMock := &daemonSetListerMock{}
	onScaleUpMock := &onScaleUpMock{}
	onScaleDownMock := &onScaleDownMock{}

	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Now())
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Now())
	n3 := BuildTestNode("n3", 1000, 1000)

	p1 := BuildTestPod("p1", 600, 100)
	p1.Spec.NodeName = "n1"
	p2 := BuildTestPod("p2", 600, 100)

	tn := BuildTestNode("tn", 1000, 1000)
	tni := schedulernodeinfo.NewNodeInfo()
	tni.SetNode(tn)

	provider := testprovider.NewTestAutoprovisioningCloudProvider(
		func(id string, delta int) error {
			return onScaleUpMock.ScaleUp(id, delta)
		}, func(id string, name string) error {
			return onScaleDownMock.ScaleDown(id, name)
		},
		nil, nil,
		nil, map[string]*schedulernodeinfo.NodeInfo{"ng1": tni, "ng2": tni})
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNode("ng1", n1)
	ng1, ok := reflect.ValueOf(provider.GetNodeGroup("ng1")).Interface().(*testprovider.TestNodeGroup)
	if !ok {
		t.Logf("GetNodeGroup returns bad values")
	}
	assert.NotNil(t, ng1)
	assert.NotNil(t, provider)

	// Create context with mocked lister registry.
	options := config.AutoscalingOptions{
		EstimatorName:                       estimator.BinpackingEstimatorName,
		ScaleDownEnabled:                    true,
		ScaleDownUtilizationThreshold:       0.5,
		MaxNodesTotal:                       1,
		MaxCoresTotal:                       10,
		MaxMemoryTotal:                      100000,
		ScaleDownUnreadyTime:                time.Minute,
		ScaleDownUnneededTime:               time.Minute,
		FilterOutSchedulablePodsUsesPacking: true,
	}
	processorCallbacks := newBufferedAutoscalerProcessorCallbacks()
	context := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, processorCallbacks)
	listerRegistry := kube_util.NewListerRegistry(allNodeLister, readyNodeLister, scheduledPodMock,
		unschedulablePodMock, podDisruptionBudgetListerMock, daemonSetListerMock,
		nil, nil, nil, nil)
	context.ListerRegistry = listerRegistry

	clusterStateConfig := clusterstate.ClusterStateRegistryConfig{
		OkTotalUnreadyCount:  1,
		MaxNodeProvisionTime: 10 * time.Second,
	}

	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterStateConfig, context.LogRecorder, newBackoff())
	sd := NewScaleDown(&context, clusterState, 0)

	autoscaler := &BufferedAutoscaler{
		Context:               &context,
		clusterStateRegistry:  clusterState,
		lastScaleUpTime:       time.Now(),
		lastScaleDownFailTime: time.Now(),
		scaleDown:             sd,
		processors:            NewTestProcessors(),
		processorCallbacks:    processorCallbacks,
		initialized:           true,
	}

	// MaxNodesTotal reached.
	readyNodeLister.SetNodes([]*apiv1.Node{n1})
	allNodeLister.SetNodes([]*apiv1.Node{n1})
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1}, nil).Times(4)
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{p2}, nil).Once()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Twice()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()

	err := autoscaler.RunOnce(time.Now())
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Scale up.
	readyNodeLister.SetNodes([]*apiv1.Node{n1})
	allNodeLister.SetNodes([]*apiv1.Node{n1})
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1}, nil).Times(3)
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{p2}, nil).Once()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Twice()
	onScaleUpMock.On("ScaleUp", "ng1", 1).Return(nil).Once()

	context.MaxNodesTotal = 10
	err = autoscaler.RunOnce(time.Now().Add(time.Hour))
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Mark unneeded nodes.
	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1}, nil).Times(4)
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{}, nil).Once()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Twice()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()

	provider.AddNode("ng1", n2)
	ng1.SetTargetSize(2)

	err = autoscaler.RunOnce(time.Now().Add(2 * time.Hour))
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Scale down.
	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1}, nil).Times(4)
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{}, nil).Once()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Twice()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()
	onScaleDownMock.On("ScaleDown", "ng1", "n2").Return(nil).Once()

	err = autoscaler.RunOnce(time.Now().Add(3 * time.Hour))
	waitForDeleteToFinish(t, autoscaler.scaleDown)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Mark unregistered nodes.
	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1}, nil).Times(4)
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{p2}, nil).Once()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Twice()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()

	provider.AddNodeGroup("ng2", 0, 10, 1)
	provider.AddNode("ng2", n3)

	err = autoscaler.RunOnce(time.Now().Add(4 * time.Hour))
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Remove unregistered nodes.
	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1}, nil).Times(4)
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{p2}, nil).Once()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Twice()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()
	onScaleDownMock.On("ScaleDown", "ng2", "n3").Return(nil).Once()

	err = autoscaler.RunOnce(time.Now().Add(5 * time.Hour))
	waitForDeleteToFinish(t, autoscaler.scaleDown)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)
}

func TestBufferedAutoscalerRunOnceWithAutoprovisionedEnabled(t *testing.T) {
	readyNodeLister := kubernetes.NewTestNodeLister(nil)
	allNodeLister := kubernetes.NewTestNodeLister(nil)
	scheduledPodMock := &podListerMock{}
	unschedulablePodMock := &podListerMock{}
	podDisruptionBudgetListerMock := &podDisruptionBudgetListerMock{}
	daemonSetListerMock := &daemonSetListerMock{}
	onScaleUpMock := &onScaleUpMock{}
	onScaleDownMock := &onScaleDownMock{}
	onNodeGroupCreateMock := &onNodeGroupCreateMock{}
	onNodeGroupDeleteMock := &onNodeGroupDeleteMock{}
	nodeGroupManager := &mockAutoprovisioningNodeGroupManager{t}
	nodeGroupListProcessor := &mockAutoprovisioningNodeGroupListProcessor{t}

	n1 := BuildTestNode("n1", 100, 1000)
	SetNodeReadyState(n1, true, time.Now())
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Now())

	p1 := BuildTestPod("p1", 100, 100)
	p1.Spec.NodeName = "n1"
	p2 := BuildTestPod("p2", 600, 100)

	tn1 := BuildTestNode("tn1", 100, 1000)
	SetNodeReadyState(tn1, true, time.Now())
	tni1 := schedulernodeinfo.NewNodeInfo()
	err := tni1.SetNode(tn1)
	if err != nil {
		t.Logf("SetNode failed")
	}
	tn2 := BuildTestNode("tn2", 1000, 1000)
	SetNodeReadyState(tn2, true, time.Now())
	tni2 := schedulernodeinfo.NewNodeInfo()
	err = tni2.SetNode(tn2)
	if err != nil {
		t.Logf("SetNode failed")
	}
	tn3 := BuildTestNode("tn3", 100, 1000)
	SetNodeReadyState(tn2, true, time.Now())
	tni3 := schedulernodeinfo.NewNodeInfo()
	err = tni3.SetNode(tn3)
	if err != nil {
		t.Logf("SetNode failed")
	}

	provider := testprovider.NewTestAutoprovisioningCloudProvider(
		func(id string, delta int) error {
			return onScaleUpMock.ScaleUp(id, delta)
		}, func(id string, name string) error {
			return onScaleDownMock.ScaleDown(id, name)
		}, func(id string) error {
			return onNodeGroupCreateMock.Create(id)
		}, func(id string) error {
			return onNodeGroupDeleteMock.Delete(id)
		},
		[]string{"TN1", "TN2"}, map[string]*schedulernodeinfo.NodeInfo{"TN1": tni1, "TN2": tni2, "ng1": tni3})
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddAutoprovisionedNodeGroup("autoprovisioned-TN1", 0, 10, 0, "TN1")
	autoprovisionedTN1, ok := reflect.ValueOf(
		provider.GetNodeGroup("autoprovisioned-TN1")).Interface().(*testprovider.TestNodeGroup)
	if !ok {
		t.Logf("GetNodeGroup returns bad values")
	}
	assert.NotNil(t, autoprovisionedTN1)
	provider.AddNode("ng1,", n1)
	assert.NotNil(t, provider)

	processors := NewTestProcessors()
	processors.NodeGroupManager = nodeGroupManager
	processors.NodeGroupListProcessor = nodeGroupListProcessor

	// Create context with mocked lister registry.
	options := config.AutoscalingOptions{
		EstimatorName:                       estimator.BinpackingEstimatorName,
		ScaleDownEnabled:                    true,
		ScaleDownUtilizationThreshold:       0.5,
		MaxNodesTotal:                       100,
		MaxCoresTotal:                       100,
		MaxMemoryTotal:                      100000,
		ScaleDownUnreadyTime:                time.Minute,
		ScaleDownUnneededTime:               time.Minute,
		NodeAutoprovisioningEnabled:         true,
		MaxAutoprovisionedNodeGroupCount:    10, // Pods with null priority are always non expendable. Test if it works.
		FilterOutSchedulablePodsUsesPacking: true,
	}
	processorCallbacks := newBufferedAutoscalerProcessorCallbacks()
	context := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, processorCallbacks)
	listerRegistry := kube_util.NewListerRegistry(allNodeLister, readyNodeLister, scheduledPodMock,
		unschedulablePodMock, podDisruptionBudgetListerMock, daemonSetListerMock,
		nil, nil, nil, nil)
	context.ListerRegistry = listerRegistry

	clusterStateConfig := clusterstate.ClusterStateRegistryConfig{
		OkTotalUnreadyCount:  0,
		MaxNodeProvisionTime: 10 * time.Second,
	}
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterStateConfig, context.LogRecorder, newBackoff())

	sd := NewScaleDown(&context, clusterState, 0)

	autoscaler := &BufferedAutoscaler{
		Context:               &context,
		clusterStateRegistry:  clusterState,
		lastScaleUpTime:       time.Now(),
		lastScaleDownFailTime: time.Now(),
		scaleDown:             sd,
		processors:            processors,
		processorCallbacks:    processorCallbacks,
		initialized:           true,
	}

	// Scale up.
	readyNodeLister.SetNodes([]*apiv1.Node{n1})
	allNodeLister.SetNodes([]*apiv1.Node{n1})
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1}, nil).Times(3)
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{p2}, nil).Once()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Twice()
	onNodeGroupCreateMock.On("Create", "autoprovisioned-TN2").Return(nil).Once()
	onScaleUpMock.On("ScaleUp", "autoprovisioned-TN2", 1).Return(nil).Once()

	err = autoscaler.RunOnce(time.Now().Add(time.Hour))
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Fix target size.
	autoprovisionedTN1.SetTargetSize(0)

	// Remove autoprovisioned node group and mark unneeded nodes.
	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1}, nil).Times(4)
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{}, nil).Once()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Twice()
	onNodeGroupDeleteMock.On("Delete", "autoprovisioned-TN1").Return(nil).Once()
	onScaleUpMock.On("ScaleUp", "ng1", -1).Return(nil).Once()

	provider.AddAutoprovisionedNodeGroup("autoprovisioned-TN2", 0, 10, 1, "TN1")
	provider.AddNode("autoprovisioned-TN2", n2)

	err = autoscaler.RunOnce(time.Now().Add(1 * time.Hour))
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Scale down.
	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1}, nil).Times(4)
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{}, nil).Once()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Twice()
	onNodeGroupDeleteMock.On("Delete", "autoprovisioned-"+
		"TN1").Return(nil).Once()
	onScaleDownMock.On("ScaleDown", "autoprovisioned-TN2", "n2").Return(nil).Once()

	err = autoscaler.RunOnce(time.Now().Add(2 * time.Hour))
	waitForDeleteToFinish(t, autoscaler.scaleDown)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)
}

func TestBufferedAutoscalerRunOnceWithALongUnregisteredNode(t *testing.T) {
	readyNodeLister := kubernetes.NewTestNodeLister(nil)
	allNodeLister := kubernetes.NewTestNodeLister(nil)
	scheduledPodMock := &podListerMock{}
	unschedulablePodMock := &podListerMock{}
	podDisruptionBudgetListerMock := &podDisruptionBudgetListerMock{}
	daemonSetListerMock := &daemonSetListerMock{}
	onScaleUpMock := &onScaleUpMock{}
	onScaleDownMock := &onScaleDownMock{}

	now := time.Now()
	later := now.Add(1 * time.Minute)

	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Now())
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Now())

	p1 := BuildTestPod("p1", 600, 100)
	p1.Spec.NodeName = "n1"
	p2 := BuildTestPod("p2", 600, 100)

	provider := testprovider.NewTestCloudProvider(
		func(id string, delta int) error {
			return onScaleUpMock.ScaleUp(id, delta)
		}, func(id string, name string) error {
			return onScaleDownMock.ScaleDown(id, name)
		})
	provider.AddNodeGroup("ng1", 2, 10, 2)
	provider.AddNode("ng1", n1)

	// broken node, that will be just hanging out there during
	// the test (it can't be removed since that would validate group min size)
	brokenNode := BuildTestNode("broken", 1000, 1000)
	provider.AddNode("ng1", brokenNode)

	ng1, ok := reflect.ValueOf(provider.GetNodeGroup("ng1")).Interface().(*testprovider.TestNodeGroup)
	if !ok {
		t.Logf("GetNodeGroup returns bad values")
	}
	assert.NotNil(t, ng1)
	assert.NotNil(t, provider)

	// Create context with mocked lister registry.
	options := config.AutoscalingOptions{
		EstimatorName:                       estimator.BinpackingEstimatorName,
		ScaleDownEnabled:                    true,
		ScaleDownUtilizationThreshold:       0.5,
		MaxNodesTotal:                       10,
		MaxCoresTotal:                       10,
		MaxMemoryTotal:                      100000,
		ScaleDownUnreadyTime:                time.Minute,
		ScaleDownUnneededTime:               time.Minute,
		MaxNodeProvisionTime:                10 * time.Second,
		FilterOutSchedulablePodsUsesPacking: true,
	}
	processorCallbacks := newBufferedAutoscalerProcessorCallbacks()
	context := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, processorCallbacks)
	listerRegistry := kube_util.NewListerRegistry(allNodeLister, readyNodeLister, scheduledPodMock,
		unschedulablePodMock, podDisruptionBudgetListerMock, daemonSetListerMock,
		nil, nil, nil, nil)
	context.ListerRegistry = listerRegistry

	clusterStateConfig := clusterstate.ClusterStateRegistryConfig{
		OkTotalUnreadyCount:  1,
		MaxNodeProvisionTime: 10 * time.Second,
	}
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterStateConfig, context.LogRecorder, newBackoff())
	// broken node detected as unregistered

	nodes := []*apiv1.Node{n1}
	err := clusterState.UpdateNodes(nodes, nil, now)
	if err != nil {
		t.Logf("UpdateNodes failed")
	}

	// broken node failed to register in time
	err = clusterState.UpdateNodes(nodes, nil, later)
	if err != nil {
		t.Logf("UpdateNodes failed")
	}

	sd := NewScaleDown(&context, clusterState, 0)

	autoscaler := &BufferedAutoscaler{
		Context:               &context,
		clusterStateRegistry:  clusterState,
		lastScaleUpTime:       time.Now(),
		lastScaleDownFailTime: time.Now(),
		scaleDown:             sd,
		processors:            NewTestProcessors(),
		processorCallbacks:    processorCallbacks,
	}

	// Scale up.
	readyNodeLister.SetNodes([]*apiv1.Node{n1})
	allNodeLister.SetNodes([]*apiv1.Node{n1})
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1}, nil).Times(3)
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{p2}, nil).Once()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Twice()
	onScaleUpMock.On("ScaleUp", "ng1", 1).Return(nil).Once()

	err = autoscaler.RunOnce(later.Add(time.Hour))
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Remove broken node after going over min size
	provider.AddNode("ng1", n2)
	ng1.SetTargetSize(3)

	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1}, nil).Times(4)
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{p2}, nil).Once()
	onScaleDownMock.On("ScaleDown", "ng1", "broken").Return(nil).Once()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Twice()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()

	err = autoscaler.RunOnce(later.Add(2 * time.Hour))
	waitForDeleteToFinish(t, autoscaler.scaleDown)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)
}

func TestBufferedAutoscalerRunOncePodsWithFilterOutSchedulablePodsUsesPackingFalse(t *testing.T) {
	readyNodeLister := kubernetes.NewTestNodeLister(nil)
	allNodeLister := kubernetes.NewTestNodeLister(nil)
	scheduledPodMock := &podListerMock{}
	unschedulablePodMock := &podListerMock{}
	podDisruptionBudgetListerMock := &podDisruptionBudgetListerMock{}
	daemonSetListerMock := &daemonSetListerMock{}
	onScaleUpMock := &onScaleUpMock{}
	onScaleDownMock := &onScaleDownMock{}

	n1 := BuildTestNode("n1", 100, 1000)
	SetNodeReadyState(n1, true, time.Now())
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Now())
	n3 := BuildTestNode("n3", 1000, 1000)
	SetNodeReadyState(n3, true, time.Now())

	// shared owner reference
	ownerRef := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")

	p1 := BuildTestPod("p1", 40, 0)
	p1.OwnerReferences = ownerRef
	p1.Spec.NodeName = "n1"

	p2 := BuildTestPod("p2", 400, 0)
	p2.OwnerReferences = ownerRef
	p2.Spec.NodeName = "n2"

	p3 := BuildTestPod("p3", 400, 0)
	p3.OwnerReferences = ownerRef
	p3.Spec.NodeName = "n2"

	p4 := BuildTestPod("p4", 500, 0)
	p4.OwnerReferences = ownerRef

	p5 := BuildTestPod("p5", 800, 0)
	p5.OwnerReferences = ownerRef
	p5.Status.NominatedNodeName = "n3"

	p6 := BuildTestPod("p6", 1000, 0)
	p6.OwnerReferences = ownerRef

	provider := testprovider.NewTestCloudProvider(
		func(id string, delta int) error {
			return onScaleUpMock.ScaleUp(id, delta)
		}, func(id string, name string) error {
			return onScaleDownMock.ScaleDown(id, name)
		})
	provider.AddNodeGroup("ng1", 0, 10, 1)
	provider.AddNodeGroup("ng2", 0, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng2", n2)
	provider.AddNode("ng2", n3)
	assert.NotNil(t, provider)
	ng2, ok := reflect.ValueOf(provider.GetNodeGroup("ng2")).Interface().(*testprovider.TestNodeGroup)
	if !ok {
		t.Logf("GetNodeGroup returns bad values.")
	}
	assert.NotNil(t, ng2)

	// Create context with mocked lister registry.
	options := config.AutoscalingOptions{
		EstimatorName:                 estimator.BinpackingEstimatorName,
		ScaleDownEnabled:              true,
		ScaleDownUtilizationThreshold: 0.5,
		MaxNodesTotal:                 10,
		MaxCoresTotal:                 10,
		MaxMemoryTotal:                100000,
		ScaleDownUnreadyTime:          time.Minute,
		ScaleDownUnneededTime:         time.Minute,
		ExpendablePodsPriorityCutoff:  10,
		//Turn off filtering schedulables using packing
		FilterOutSchedulablePodsUsesPacking: false,
	}
	processorCallbacks := newBufferedAutoscalerProcessorCallbacks()
	context := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, processorCallbacks)
	listerRegistry := kube_util.NewListerRegistry(allNodeLister, readyNodeLister, scheduledPodMock,
		unschedulablePodMock, podDisruptionBudgetListerMock, daemonSetListerMock,
		nil, nil, nil, nil)
	context.ListerRegistry = listerRegistry

	clusterStateConfig := clusterstate.ClusterStateRegistryConfig{
		OkTotalUnreadyCount:  1,
		MaxNodeProvisionTime: 10 * time.Second,
	}

	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterStateConfig, context.LogRecorder, newBackoff())
	sd := NewScaleDown(&context, clusterState, 0)

	autoscaler := &BufferedAutoscaler{
		Context:               &context,
		clusterStateRegistry:  clusterState,
		lastScaleUpTime:       time.Now(),
		lastScaleDownFailTime: time.Now(),
		scaleDown:             sd,
		processors:            NewTestProcessors(),
		processorCallbacks:    processorCallbacks,
	}

	// Scale up
	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2, n3})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2, n3})
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1, p2, p3}, nil).Times(3)
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{p4, p5, p6}, nil).Once()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Twice()
	onScaleUpMock.On("ScaleUp", "ng2", 2).Return(nil).Once()

	err := autoscaler.RunOnce(time.Now())
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)
}

func TestBufferedAutoscalerRunOncePodsWithPriorities(t *testing.T) {
	readyNodeLister := kubernetes.NewTestNodeLister(nil)
	allNodeLister := kubernetes.NewTestNodeLister(nil)
	scheduledPodMock := &podListerMock{}
	unschedulablePodMock := &podListerMock{}
	podDisruptionBudgetListerMock := &podDisruptionBudgetListerMock{}
	daemonSetListerMock := &daemonSetListerMock{}
	onScaleUpMock := &onScaleUpMock{}
	onScaleDownMock := &onScaleDownMock{}

	n1 := BuildTestNode("n1", 100, 1000)
	SetNodeReadyState(n1, true, time.Now())
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Now())
	n3 := BuildTestNode("n3", 1000, 1000)
	SetNodeReadyState(n3, true, time.Now())

	// shared owner reference
	ownerRef := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")
	var priority100 int32 = 100
	var priority1 int32 = 1

	p1 := BuildTestPod("p1", 40, 0)
	p1.OwnerReferences = ownerRef
	p1.Spec.NodeName = "n1"
	p1.Spec.Priority = &priority1

	p2 := BuildTestPod("p2", 400, 0)
	p2.OwnerReferences = ownerRef
	p2.Spec.NodeName = "n2"
	p2.Spec.Priority = &priority1

	p3 := BuildTestPod("p3", 400, 0)
	p3.OwnerReferences = ownerRef
	p3.Spec.NodeName = "n2"
	p3.Spec.Priority = &priority100

	p4 := BuildTestPod("p4", 500, 0)
	p4.OwnerReferences = ownerRef
	p4.Spec.Priority = &priority100

	p5 := BuildTestPod("p5", 800, 0)
	p5.OwnerReferences = ownerRef
	p5.Spec.Priority = &priority100
	p5.Status.NominatedNodeName = "n3"

	p6 := BuildTestPod("p6", 1000, 0)
	p6.OwnerReferences = ownerRef
	p6.Spec.Priority = &priority100

	provider := testprovider.NewTestCloudProvider(
		func(id string, delta int) error {
			return onScaleUpMock.ScaleUp(id, delta)
		}, func(id string, name string) error {
			return onScaleDownMock.ScaleDown(id, name)
		})
	provider.AddNodeGroup("ng1", 0, 10, 1)
	provider.AddNodeGroup("ng2", 0, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng2", n2)
	provider.AddNode("ng2", n3)
	assert.NotNil(t, provider)
	ng2, ok := reflect.ValueOf(provider.GetNodeGroup("ng2")).Interface().(*testprovider.TestNodeGroup)
	if !ok {
		t.Logf("GetNodeGroup returns bad values.")
	}
	assert.NotNil(t, ng2)

	// Create context with mocked lister registry.
	options := config.AutoscalingOptions{
		EstimatorName:                       estimator.BinpackingEstimatorName,
		ScaleDownEnabled:                    true,
		ScaleDownUtilizationThreshold:       0.5,
		MaxNodesTotal:                       10,
		MaxCoresTotal:                       10,
		MaxMemoryTotal:                      100000,
		ScaleDownUnreadyTime:                time.Minute,
		ScaleDownUnneededTime:               time.Minute,
		ExpendablePodsPriorityCutoff:        10,
		FilterOutSchedulablePodsUsesPacking: true,
	}
	processorCallbacks := newBufferedAutoscalerProcessorCallbacks()
	context := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, processorCallbacks)
	listerRegistry := kube_util.NewListerRegistry(allNodeLister, readyNodeLister, scheduledPodMock,
		unschedulablePodMock, podDisruptionBudgetListerMock, daemonSetListerMock,
		nil, nil, nil, nil)
	context.ListerRegistry = listerRegistry

	clusterStateConfig := clusterstate.ClusterStateRegistryConfig{
		OkTotalUnreadyCount:  1,
		MaxNodeProvisionTime: 10 * time.Second,
	}

	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterStateConfig, context.LogRecorder, newBackoff())
	sd := NewScaleDown(&context, clusterState, 0)

	autoscaler := &BufferedAutoscaler{
		Context:               &context,
		clusterStateRegistry:  clusterState,
		lastScaleUpTime:       time.Now(),
		lastScaleDownFailTime: time.Now(),
		scaleDown:             sd,
		processors:            NewTestProcessors(),
		processorCallbacks:    processorCallbacks,
	}

	// Scale up
	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2, n3})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2, n3})
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1, p2, p3}, nil).Times(3)
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{p4, p5, p6}, nil).Once()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Twice()
	onScaleUpMock.On("ScaleUp", "ng2", 1).Return(nil).Once()

	err := autoscaler.RunOnce(time.Now())
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Mark unneeded nodes.
	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2, n3})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2, n3})
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1, p2, p3}, nil).Times(4)
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{p4, p5}, nil).Once()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Twice()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()

	ng2.SetTargetSize(2)

	err = autoscaler.RunOnce(time.Now().Add(2 * time.Hour))
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)

	// Scale down.
	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2, n3})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2, n3})
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p1, p2, p3, p4}, nil).Times(4)
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{p5}, nil).Once()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Twice()
	podDisruptionBudgetListerMock.On("List").Return([]*policyv1.PodDisruptionBudget{}, nil).Once()
	onScaleDownMock.On("ScaleDown", "ng1", "n1").Return(nil).Once()

	p4.Spec.NodeName = "n2"

	err = autoscaler.RunOnce(time.Now().Add(3 * time.Hour))
	waitForDeleteToFinish(t, autoscaler.scaleDown)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)
}

func TestBufferedAutoscalerRunOnceWithFilteringOnBinPackingEstimator(t *testing.T) {
	readyNodeLister := kubernetes.NewTestNodeLister(nil)
	allNodeLister := kubernetes.NewTestNodeLister(nil)
	scheduledPodMock := &podListerMock{}
	unschedulablePodMock := &podListerMock{}
	podDisruptionBudgetListerMock := &podDisruptionBudgetListerMock{}
	daemonSetListerMock := &daemonSetListerMock{}
	onScaleUpMock := &onScaleUpMock{}
	onScaleDownMock := &onScaleDownMock{}

	n1 := BuildTestNode("n1", 2000, 1000)
	SetNodeReadyState(n1, true, time.Now())
	n2 := BuildTestNode("n2", 2000, 1000)
	SetNodeReadyState(n2, true, time.Now())

	// shared owner reference
	ownerRef := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")

	p1 := BuildTestPod("p1", 1400, 0)
	p1.OwnerReferences = ownerRef
	p3 := BuildTestPod("p3", 1400, 0)
	p3.Spec.NodeName = "n1"
	p3.OwnerReferences = ownerRef
	p4 := BuildTestPod("p4", 1400, 0)
	p4.Spec.NodeName = "n2"
	p4.OwnerReferences = ownerRef

	provider := testprovider.NewTestCloudProvider(
		func(id string, delta int) error {
			return onScaleUpMock.ScaleUp(id, delta)
		}, func(id string, name string) error {
			return onScaleDownMock.ScaleDown(id, name)
		})
	provider.AddNodeGroup("ng1", 0, 10, 2)
	provider.AddNode("ng1", n1)

	provider.AddNodeGroup("ng2", 0, 10, 1)
	provider.AddNode("ng2", n2)

	assert.NotNil(t, provider)

	// Create context with mocked lister registry.
	options := config.AutoscalingOptions{
		EstimatorName:                 estimator.BinpackingEstimatorName,
		ScaleDownEnabled:              false,
		ScaleDownUtilizationThreshold: 0.5,
		MaxNodesTotal:                 10,
		MaxCoresTotal:                 10,
		MaxMemoryTotal:                100000,
		ExpendablePodsPriorityCutoff:  10,
	}
	processorCallbacks := newBufferedAutoscalerProcessorCallbacks()
	context := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, processorCallbacks)
	listerRegistry := kube_util.NewListerRegistry(allNodeLister, readyNodeLister, scheduledPodMock,
		unschedulablePodMock, podDisruptionBudgetListerMock, daemonSetListerMock,
		nil, nil, nil, nil)
	context.ListerRegistry = listerRegistry

	clusterStateConfig := clusterstate.ClusterStateRegistryConfig{
		OkTotalUnreadyCount:  1,
		MaxNodeProvisionTime: 10 * time.Second,
	}

	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterStateConfig, context.LogRecorder, newBackoff())
	sd := NewScaleDown(&context, clusterState, 0)

	autoscaler := &BufferedAutoscaler{
		Context:               &context,
		clusterStateRegistry:  clusterState,
		lastScaleUpTime:       time.Now(),
		lastScaleDownFailTime: time.Now(),
		scaleDown:             sd,
		processors:            NewTestProcessors(),
		processorCallbacks:    processorCallbacks,
	}

	// Scale up
	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p3, p4}, nil).Times(4)
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{p1}, nil).Once()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Twice()

	err := autoscaler.RunOnce(time.Now())
	assert.NoError(t, err)

	mock.AssertExpectationsForObjects(t, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)
}

func TestBufferedAutoscalerRunOnceWithFilteringOnOldEstimator(t *testing.T) {
	readyNodeLister := kubernetes.NewTestNodeLister(nil)
	allNodeLister := kubernetes.NewTestNodeLister(nil)
	scheduledPodMock := &podListerMock{}
	unschedulablePodMock := &podListerMock{}
	podDisruptionBudgetListerMock := &podDisruptionBudgetListerMock{}
	daemonSetListerMock := &daemonSetListerMock{}
	onScaleUpMock := &onScaleUpMock{}
	onScaleDownMock := &onScaleDownMock{}

	n1 := BuildTestNode("n1", 2000, 1000)
	SetNodeReadyState(n1, true, time.Now())
	n2 := BuildTestNode("n2", 2000, 1000)
	SetNodeReadyState(n2, true, time.Now())

	// shared owner reference
	ownerRef := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")

	p1 := BuildTestPod("p1", 1400, 0)
	p1.OwnerReferences = ownerRef
	p3 := BuildTestPod("p3", 1400, 0)
	p3.Spec.NodeName = "n1"
	p3.OwnerReferences = ownerRef
	p4 := BuildTestPod("p4", 1400, 0)
	p4.Spec.NodeName = "n2"
	p4.OwnerReferences = ownerRef

	provider := testprovider.NewTestCloudProvider(
		func(id string, delta int) error {
			return onScaleUpMock.ScaleUp(id, delta)
		}, func(id string, name string) error {
			return onScaleDownMock.ScaleDown(id, name)
		})
	provider.AddNodeGroup("ng1", 0, 10, 2)
	provider.AddNode("ng1", n1)

	provider.AddNodeGroup("ng2", 0, 10, 1)
	provider.AddNode("ng2", n2)

	assert.NotNil(t, provider)

	// Create context with mocked lister registry.
	options := config.AutoscalingOptions{
		EstimatorName:                 estimator.OldBinpackingEstimatorName,
		ScaleDownEnabled:              false,
		ScaleDownUtilizationThreshold: 0.5,
		MaxNodesTotal:                 10,
		MaxCoresTotal:                 10,
		MaxMemoryTotal:                100000,
		ExpendablePodsPriorityCutoff:  10,
	}
	processorCallbacks := newBufferedAutoscalerProcessorCallbacks()
	context := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, processorCallbacks)
	listerRegistry := kube_util.NewListerRegistry(allNodeLister, readyNodeLister, scheduledPodMock,
		unschedulablePodMock, podDisruptionBudgetListerMock, daemonSetListerMock,
		nil, nil, nil, nil)
	context.ListerRegistry = listerRegistry

	clusterStateConfig := clusterstate.ClusterStateRegistryConfig{
		OkTotalUnreadyCount:  1,
		MaxNodeProvisionTime: 10 * time.Second,
	}

	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterStateConfig, context.LogRecorder, newBackoff())
	sd := NewScaleDown(&context, clusterState, 0)

	autoscaler := &BufferedAutoscaler{
		Context:               &context,
		clusterStateRegistry:  clusterState,
		lastScaleUpTime:       time.Now(),
		lastScaleDownFailTime: time.Now(),
		scaleDown:             sd,
		processors:            NewTestProcessors(),
		processorCallbacks:    processorCallbacks,
	}

	// Scale up
	readyNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	allNodeLister.SetNodes([]*apiv1.Node{n1, n2})
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p3, p4}, nil).Times(4)
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{p1}, nil).Once()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Twice()

	err := autoscaler.RunOnce(time.Now())
	assert.NoError(t, err)

	mock.AssertExpectationsForObjects(t, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)
}

func TestBufferedAutoscalerRunOnceWithFilteringOnUpcomingNodesEnabledNoScaleUp(t *testing.T) {
	readyNodeLister := kubernetes.NewTestNodeLister(nil)
	allNodeLister := kubernetes.NewTestNodeLister(nil)
	scheduledPodMock := &podListerMock{}
	unschedulablePodMock := &podListerMock{}
	podDisruptionBudgetListerMock := &podDisruptionBudgetListerMock{}
	daemonSetListerMock := &daemonSetListerMock{}
	onScaleUpMock := &onScaleUpMock{}
	onScaleDownMock := &onScaleDownMock{}

	n2 := BuildTestNode("n2", 2000, 1000)
	SetNodeReadyState(n2, true, time.Now())
	n3 := BuildTestNode("n3", 2000, 1000)
	SetNodeReadyState(n3, true, time.Now())

	// shared owner reference
	ownerRef := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")

	p1 := BuildTestPod("p1", 1400, 0)
	p1.OwnerReferences = ownerRef
	p2 := BuildTestPod("p2", 1400, 0)
	p2.Spec.NodeName = "n2"
	p2.OwnerReferences = ownerRef
	p3 := BuildTestPod("p3", 1400, 0)
	p3.Spec.NodeName = "n3"
	p3.OwnerReferences = ownerRef

	provider := testprovider.NewTestCloudProvider(
		func(id string, delta int) error {
			return onScaleUpMock.ScaleUp(id, delta)
		}, func(id string, name string) error {
			return onScaleDownMock.ScaleDown(id, name)
		})
	provider.AddNodeGroup("ng1", 0, 10, 2)
	provider.AddNode("ng1", n2)

	provider.AddNodeGroup("ng2", 0, 10, 1)
	provider.AddNode("ng2", n3)

	assert.NotNil(t, provider)

	// Create context with mocked lister registry.
	options := config.AutoscalingOptions{
		EstimatorName:                 estimator.BinpackingEstimatorName,
		ScaleDownEnabled:              false,
		ScaleDownUtilizationThreshold: 0.5,
		MaxNodesTotal:                 10,
		MaxCoresTotal:                 10,
		MaxMemoryTotal:                100000,
		ExpendablePodsPriorityCutoff:  10,
	}
	processorCallbacks := newBufferedAutoscalerProcessorCallbacks()
	context := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, processorCallbacks)
	listerRegistry := kube_util.NewListerRegistry(allNodeLister, readyNodeLister, scheduledPodMock,
		unschedulablePodMock, podDisruptionBudgetListerMock, daemonSetListerMock,
		nil, nil, nil, nil)
	context.ListerRegistry = listerRegistry

	clusterStateConfig := clusterstate.ClusterStateRegistryConfig{
		OkTotalUnreadyCount:  1,
		MaxNodeProvisionTime: 10 * time.Second,
	}

	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterStateConfig, context.LogRecorder, newBackoff())
	sd := NewScaleDown(&context, clusterState, 0)

	autoscaler := &BufferedAutoscaler{
		Context:               &context,
		clusterStateRegistry:  clusterState,
		lastScaleUpTime:       time.Now(),
		lastScaleDownFailTime: time.Now(),
		scaleDown:             sd,
		processors:            NewTestProcessors(),
		processorCallbacks:    processorCallbacks,
	}

	// Scale up
	readyNodeLister.SetNodes([]*apiv1.Node{n2, n3})
	allNodeLister.SetNodes([]*apiv1.Node{n2, n3})
	scheduledPodMock.On("List").Return([]*apiv1.Pod{p2, p3}, nil).Times(4)
	unschedulablePodMock.On("List").Return([]*apiv1.Pod{p1}, nil).Once()
	daemonSetListerMock.On("List", labels.Everything()).Return([]*appsv1.DaemonSet{}, nil).Twice()

	err := autoscaler.RunOnce(time.Now())
	assert.NoError(t, err)

	mock.AssertExpectationsForObjects(t, scheduledPodMock, unschedulablePodMock,
		podDisruptionBudgetListerMock, daemonSetListerMock, onScaleUpMock, onScaleDownMock)
}

func TestBufferedAutoscalerInstaceCreationErrors(t *testing.T) {

	// setup
	provider := &mockprovider.CloudProvider{}

	// Create context with mocked lister registry.
	options := config.AutoscalingOptions{
		EstimatorName:                       estimator.BinpackingEstimatorName,
		ScaleDownEnabled:                    true,
		ScaleDownUtilizationThreshold:       0.5,
		MaxNodesTotal:                       10,
		MaxCoresTotal:                       10,
		MaxMemoryTotal:                      100000,
		ScaleDownUnreadyTime:                time.Minute,
		ScaleDownUnneededTime:               time.Minute,
		ExpendablePodsPriorityCutoff:        10,
		FilterOutSchedulablePodsUsesPacking: true,
	}
	processorCallbacks := newBufferedAutoscalerProcessorCallbacks()
	context := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, processorCallbacks)

	clusterStateConfig := clusterstate.ClusterStateRegistryConfig{
		OkTotalUnreadyCount:  1,
		MaxNodeProvisionTime: 10 * time.Second,
	}

	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterStateConfig, context.LogRecorder, newBackoff())
	autoscaler := &BufferedAutoscaler{
		Context:               &context,
		clusterStateRegistry:  clusterState,
		lastScaleUpTime:       time.Now(),
		lastScaleDownFailTime: time.Now(),
		processorCallbacks:    processorCallbacks,
	}

	nodeGroupA := &mockprovider.NodeGroup{}
	nodeGroupB := &mockprovider.NodeGroup{}

	// Three nodes with out-of-resources errors
	nodeGroupA.On("Exist").Return(true)
	nodeGroupA.On("Autoprovisioned").Return(false)
	nodeGroupA.On("TargetSize").Return(5, nil)
	nodeGroupA.On("Id").Return("A")
	nodeGroupA.On("DeleteNodes", mock.Anything).Return(nil)
	nodeGroupA.On("Nodes").Return([]cloudprovider.Instance{
		{
			Id: "A1",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceRunning,
			},
		},
		{
			Id: "A2",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
			},
		},
		{
			Id: "A3",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
				ErrorInfo: &cloudprovider.InstanceErrorInfo{
					ErrorClass: cloudprovider.OutOfResourcesErrorClass,
					ErrorCode:  "STOCKOUT",
				},
			},
		},
		{
			Id: "A4",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
				ErrorInfo: &cloudprovider.InstanceErrorInfo{
					ErrorClass: cloudprovider.OutOfResourcesErrorClass,
					ErrorCode:  "STOCKOUT",
				},
			},
		},
		{
			Id: "A5",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
				ErrorInfo: &cloudprovider.InstanceErrorInfo{
					ErrorClass: cloudprovider.OutOfResourcesErrorClass,
					ErrorCode:  "QUOTA",
				},
			},
		},
		{
			Id: "A6",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
				ErrorInfo: &cloudprovider.InstanceErrorInfo{
					ErrorClass: cloudprovider.OtherErrorClass,
					ErrorCode:  "OTHER",
				},
			},
		},
	}, nil).Twice()

	nodeGroupB.On("Exist").Return(true)
	nodeGroupB.On("Autoprovisioned").Return(false)
	nodeGroupB.On("TargetSize").Return(5, nil)
	nodeGroupB.On("Id").Return("B")
	nodeGroupB.On("DeleteNodes", mock.Anything).Return(nil)
	nodeGroupB.On("Nodes").Return([]cloudprovider.Instance{
		{
			Id: "B1",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceRunning,
			},
		},
	}, nil)

	provider.On("NodeGroups").Return([]cloudprovider.NodeGroup{nodeGroupA})
	provider.On("NodeGroupForNode", mock.Anything).Return(
		func(node *apiv1.Node) cloudprovider.NodeGroup {
			if strings.HasPrefix(node.Spec.ProviderID, "A") {
				return nodeGroupA
			}
			if strings.HasPrefix(node.Spec.ProviderID, "B") {
				return nodeGroupB
			}
			return nil
		}, nil)

	now := time.Now()

	clusterState.RefreshCloudProviderNodeInstancesCache()
	// propagate nodes info in cluster state
	err := clusterState.UpdateNodes([]*apiv1.Node{}, nil, now)
	if err != nil {
		t.Logf("UpdateNodes failed")
	}

	// delete nodes with create errors
	autoscaler.deleteCreatedNodesWithErrors()

	// check delete was called on correct nodes
	nodeGroupA.AssertCalled(t, "DeleteNodes", mock.MatchedBy(
		func(nodes []*apiv1.Node) bool {
			if len(nodes) != 4 {
				return false
			}
			names := make(map[string]bool)
			for _, node := range nodes {
				names[node.Spec.ProviderID] = true
			}
			return names["A3"] && names["A4"] && names["A5"] && names["A6"]
		}))

	// TODO assert that scaleup was failed (separately for QUOTA and STOCKOUT)

	clusterState.RefreshCloudProviderNodeInstancesCache()

	// propagate nodes info in cluster state again
	// no changes in what provider returns
	err = clusterState.UpdateNodes([]*apiv1.Node{}, nil, now)
	if err != nil {
		t.Logf("UpdateNodes failed")
	}

	// delete nodes with create errors
	autoscaler.deleteCreatedNodesWithErrors()

	// nodes should be deleted again
	nodeGroupA.AssertCalled(t, "DeleteNodes", mock.MatchedBy(
		func(nodes []*apiv1.Node) bool {
			if len(nodes) != 4 {
				return false
			}
			names := make(map[string]bool)
			for _, node := range nodes {
				names[node.Spec.ProviderID] = true
			}
			return names["A3"] && names["A4"] && names["A5"] && names["A6"]
		}))

	// TODO assert that scaleup is not failed again

	// restub node group A so nodes are no longer reporting errors
	nodeGroupA.On("Nodes").Return([]cloudprovider.Instance{
		{
			Id: "A1",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceRunning,
			},
		},
		{
			Id: "A2",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
			},
		},
		{
			Id: "A3",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceDeleting,
			},
		},
		{
			Id: "A4",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceDeleting,
			},
		},
		{
			Id: "A5",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceDeleting,
			},
		},
		{
			Id: "A6",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceDeleting,
			},
		},
	}, nil)

	clusterState.RefreshCloudProviderNodeInstancesCache()

	// update cluster state
	err = clusterState.UpdateNodes([]*apiv1.Node{}, nil, now)
	if err != nil {
		t.Logf("UpdateNodes failed")
	}

	// delete nodes with create errors
	autoscaler.deleteCreatedNodesWithErrors()

	// we expect no more Delete Nodes
	nodeGroupA.AssertNumberOfCalls(t, "DeleteNodes", 2)
}

func TestBufferedAutoscalerProcessorCallbacks(t *testing.T) {
	processorCallbacks := newBufferedAutoscalerProcessorCallbacks()
	assert.Equal(t, false, processorCallbacks.disableScaleDownForLoop)
	assert.Equal(t, 0, len(processorCallbacks.extraValues))

	processorCallbacks.DisableScaleDownForLoop()
	assert.Equal(t, true, processorCallbacks.disableScaleDownForLoop)
	processorCallbacks.reset()
	assert.Equal(t, false, processorCallbacks.disableScaleDownForLoop)

	_, found := processorCallbacks.GetExtraValue("blah")
	assert.False(t, found)

	processorCallbacks.SetExtraValue("blah", "some value")
	value, found := processorCallbacks.GetExtraValue("blah")
	assert.True(t, found)
	assert.Equal(t, "some value", value)

	processorCallbacks.reset()
	assert.Equal(t, 0, len(processorCallbacks.extraValues))
	_, found = processorCallbacks.GetExtraValue("blah")
	assert.False(t, found)
}
