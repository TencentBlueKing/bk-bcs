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

package core

import (
	"reflect"
	"strings"
	"testing"
	"time"

	contextinternal "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/kubernetes/fake"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

func TestGenerateAutoscalerRequest(t *testing.T) {
	onScaleUpMock := &onScaleUpMock{}
	onScaleDownMock := &onScaleDownMock{}

	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Now())
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Now())

	p1 := BuildTestPod("p1", 600, 100)
	p1.Spec.NodeName = "n1"

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
	provider.AddNodeGroup("ng2", 0, 10, 2)
	provider.AddNode("ng2", n2)
	ng1, ok := reflect.ValueOf(provider.GetNodeGroup("ng1")).Interface().(*testprovider.TestNodeGroup)
	if !ok {
		t.Logf("GetNodeGroup returns bad values")
	}
	assert.NotNil(t, ng1)
	assert.NotNil(t, provider)

	reader := strings.NewReader("11111111-1111-1111-1111-111111111111")
	uuid.SetRand(reader)

	type args struct {
		nodeGroups    []cloudprovider.NodeGroup
		upcomingNodes map[string]int
	}
	tests := []struct {
		name    string
		args    args
		want    *AutoscalerRequest
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "normal case",
			args: args{
				nodeGroups: provider.NodeGroups(),
				upcomingNodes: map[string]int{
					"ng2": 1,
				},
			},
			want: &AutoscalerRequest{
				UID: apitypes.UID("31313131-3131-4131-ad31-3131312d3131"),
				NodeGroups: map[string]*NodeGroup{
					"ng1": {
						NodeGroupID:  "ng1",
						MaxSize:      10,
						MinSize:      1,
						DesiredSize:  1,
						UpcomingSize: 0,
						NodeTemplate: Template{
							CPU:    1,
							Mem:    1000,
							GPU:    0,
							Labels: map[string]string{},
						},
						NodeIPs: []string{"n1"},
					},
					"ng2": {
						NodeGroupID:  "ng2",
						MaxSize:      10,
						MinSize:      0,
						DesiredSize:  2,
						UpcomingSize: 1,
						NodeTemplate: Template{
							CPU:    1,
							Mem:    1000,
							GPU:    0,
							Labels: map[string]string{},
						},
						NodeIPs: []string{"n2"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "normal case with empty ngs",
			args: args{
				nodeGroups:    []cloudprovider.NodeGroup{},
				upcomingNodes: map[string]int{},
			},
			want: &AutoscalerRequest{
				UID:        apitypes.UID("31312d31-3131-412d-b131-313131313131"),
				NodeGroups: map[string]*NodeGroup{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateAutoscalerRequest(tt.args.nodeGroups, tt.args.upcomingNodes)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateAutoscalerRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("%v", got.NodeGroups["ng1"])
			t.Logf("%v", tt.want.NodeGroups["ng1"])
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateAutoscalerRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleResponse(t *testing.T) {
	readyNodeLister := kubernetes.NewTestNodeLister(nil)
	allNodeLister := kubernetes.NewTestNodeLister(nil)
	scheduledPodMock := &podListerMock{}
	unschedulablePodMock := &podListerMock{}
	podDisruptionBudgetListerMock := &podDisruptionBudgetListerMock{}
	daemonSetListerMock := &daemonSetListerMock{}
	onScaleUpMock := &onScaleUpMock{}
	onScaleDownMock := &onScaleDownMock{}

	p1 := BuildTestPod("p1", 600, 100)
	p1.Spec.NodeName = "n1"
	p2 := BuildTestPod("p2", 600, 100)
	p2.Spec.NodeName = "n2"

	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Now())
	injectNodeIP(n1, "n1")
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Now())
	injectNodeIP(n2, "n2")
	n2.ObjectMeta.Annotations = map[string]string{
		"io.tencent.bcs.dev/node-deletion-cost": "200",
	}
	tn2 := schedulernodeinfo.NewNodeInfo(p2)
	tn2.SetNode(n2)
	n3 := BuildTestNode("n3", 1000, 1000)
	injectNodeIP(n3, "n3")
	n3.ObjectMeta.Annotations = map[string]string{
		"io.tencent.bcs.dev/node-deletion-cost": "30",
	}
	SetNodeReadyState(n3, true, time.Now())
	tn3 := schedulernodeinfo.NewNodeInfo()
	tn3.SetNode(n3)
	n4 := BuildTestNode("n4", 1000, 1000)
	injectNodeIP(n4, "n4")
	n4.ObjectMeta.Annotations = map[string]string{
		"io.tencent.bcs.dev/node-deletion-cost": "200",
	}
	SetNodeReadyState(n4, true, time.Now())
	tn4 := schedulernodeinfo.NewNodeInfo()
	tn4.SetNode(n4)

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
	request := &AutoscalerRequest{
		UID: apitypes.UID("31313131-3131-4131-ad31-3131312d3131"),
		NodeGroups: map[string]*NodeGroup{
			"ng1": {
				NodeGroupID:  "ng1",
				MaxSize:      10,
				MinSize:      1,
				DesiredSize:  1,
				UpcomingSize: 0,
				NodeTemplate: Template{
					CPU:    1,
					Mem:    1000,
					GPU:    0,
					Labels: map[string]string{},
				},
				NodeIPs: []string{"n1"},
			},
			"ng2": {
				NodeGroupID:  "ng2",
				MaxSize:      10,
				MinSize:      0,
				DesiredSize:  4,
				UpcomingSize: 1,
				NodeTemplate: Template{
					CPU:    1,
					Mem:    1000,
					GPU:    0,
					Labels: map[string]string{},
				},
				NodeIPs: []string{"n2", "n3", "n4"},
			},
		},
	}

	type args struct {
		review             ClusterAutoscalerReview
		nodes              []*corev1.Node
		nodeNameToNodeInfo map[string]*schedulernodeinfo.NodeInfo
		sd                 *ScaleDown
	}
	tests := []struct {
		name    string
		args    args
		want    ScaleUpOptions
		want1   ScaleDownCandidates
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "normal case with scale up",
			args: args{
				review: ClusterAutoscalerReview{
					Request: request,
					Response: &AutoscalerResponse{
						UID: apitypes.UID("31313131-3131-4131-ad31-3131312d3131"),
						ScaleUps: []*ScaleUpPolicy{
							{NodeGroupID: "ng1", DesiredSize: 5},
						},
					},
				},
				nodes:              []*corev1.Node{n1, n2},
				nodeNameToNodeInfo: map[string]*schedulernodeinfo.NodeInfo{},
				sd:                 sd,
			},
			want:    ScaleUpOptions{"ng1": 5},
			want1:   nil,
			wantErr: false,
		},
		{
			name: "normal case with scale down ips",
			args: args{
				review: ClusterAutoscalerReview{
					Request: request,
					Response: &AutoscalerResponse{
						UID: apitypes.UID("31313131-3131-4131-ad31-3131312d3131"),
						ScaleDowns: []*ScaleDownPolicy{
							{
								NodeGroupID: "ng2",
								Type:        NodeIPsScaleDownType,
								NodeIPs:     []string{"n2"},
							},
						},
					},
				},
				nodes:              []*corev1.Node{n1, n2},
				nodeNameToNodeInfo: map[string]*schedulernodeinfo.NodeInfo{},
				sd:                 sd,
			},
			want:    nil,
			want1:   ScaleDownCandidates{"n2"},
			wantErr: false,
		},
		{
			name: "scale down nonexistent ips",
			args: args{
				review: ClusterAutoscalerReview{
					Request: request,
					Response: &AutoscalerResponse{
						UID: apitypes.UID("31313131-3131-4131-ad31-3131312d3131"),
						ScaleDowns: []*ScaleDownPolicy{
							{
								NodeGroupID: "ng2",
								Type:        NodeIPsScaleDownType,
								NodeIPs:     []string{"n100", "n200"},
							},
						},
					},
				},
				nodes:              []*corev1.Node{n1, n2},
				nodeNameToNodeInfo: map[string]*schedulernodeinfo.NodeInfo{},
				sd:                 sd,
			},
			want:    nil,
			want1:   ScaleDownCandidates{},
			wantErr: false,
		},
		{
			name: "normal case with scale down num",
			args: args{
				review: ClusterAutoscalerReview{
					Request: request,
					Response: &AutoscalerResponse{
						UID: apitypes.UID("31313131-3131-4131-ad31-3131312d3131"),
						ScaleDowns: []*ScaleDownPolicy{
							{
								NodeGroupID: "ng2",
								Type:        NodeNumScaleDownType,
								NodeNum:     1,
							},
						},
					},
				},
				nodes: []*corev1.Node{n1, n2, n3, n4},
				nodeNameToNodeInfo: map[string]*schedulernodeinfo.NodeInfo{
					"n2": tn2,
					"n3": tn3,
					"n4": tn4,
				},
				sd: sd,
			},
			want:    nil,
			want1:   ScaleDownCandidates{"n3", "n4"},
			wantErr: false,
		},
		{
			name: "normal case when scale down num equals to desired size",
			args: args{
				review: ClusterAutoscalerReview{
					Request: request,
					Response: &AutoscalerResponse{
						UID: apitypes.UID("31313131-3131-4131-ad31-3131312d3131"),
						ScaleDowns: []*ScaleDownPolicy{
							{
								NodeGroupID: "ng2",
								Type:        NodeNumScaleDownType,
								NodeNum:     4,
							},
						},
					},
				},
				nodes: []*corev1.Node{n1, n2, n3, n4},
				nodeNameToNodeInfo: map[string]*schedulernodeinfo.NodeInfo{
					"n2": tn2,
					"n3": tn3,
					"n4": tn4,
				},
				sd: sd,
			},
			want:    nil,
			want1:   ScaleDownCandidates{},
			wantErr: false,
		},
		{
			name: "normal case when scale down num equals to length of Node IPs",
			args: args{
				review: ClusterAutoscalerReview{
					Request: request,
					Response: &AutoscalerResponse{
						UID: apitypes.UID("31313131-3131-4131-ad31-3131312d3131"),
						ScaleDowns: []*ScaleDownPolicy{
							{
								NodeGroupID: "ng2",
								Type:        NodeNumScaleDownType,
								NodeNum:     3,
							},
						},
					},
				},
				nodes: []*corev1.Node{n1, n2, n3, n4},
				nodeNameToNodeInfo: map[string]*schedulernodeinfo.NodeInfo{
					"n2": tn2,
					"n3": tn3,
					"n4": tn4,
				},
				sd: sd,
			},
			want:    nil,
			want1:   ScaleDownCandidates{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := HandleResponse(tt.args.review, tt.args.nodes, tt.args.nodeNameToNodeInfo, tt.args.sd)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HandleResponse() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("HandleResponse() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func injectNodeIP(node *corev1.Node, IP string) {
	node.Status.Addresses = []corev1.NodeAddress{
		{
			Type:    corev1.NodeInternalIP,
			Address: IP,
		},
	}
}

func TestExecuteScaleUp(t *testing.T) {
	readyNodeLister := kubernetes.NewTestNodeLister(nil)
	allNodeLister := kubernetes.NewTestNodeLister(nil)
	scheduledPodMock := &podListerMock{}
	unschedulablePodMock := &podListerMock{}
	podDisruptionBudgetListerMock := &podDisruptionBudgetListerMock{}
	daemonSetListerMock := &daemonSetListerMock{}
	onScaleUpMock := &onScaleUpMock{}
	onScaleDownMock := &onScaleDownMock{}

	onScaleUpMock.On("ScaleUp", "ng1", 4).Return(nil).Once()
	onScaleUpMock.On("ScaleUp", "ng2", 8).Return(nil).Once()

	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Now())

	provider := testprovider.NewTestAutoprovisioningCloudProvider(
		func(id string, delta int) error {
			return onScaleUpMock.ScaleUp(id, delta)
		}, func(id string, name string) error {
			return onScaleDownMock.ScaleDown(id, name)
		},
		nil, nil,
		nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNode("ng1", n1)
	provider.AddNodeGroup("ng2", 0, 10, 0)

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
	type args struct {
		context              *contextinternal.Context
		clusterStateRegistry *clusterstate.ClusterStateRegistry
		options              ScaleUpOptions
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "scale up ng1 and ng2 normally",
			args: args{
				context:              &context,
				clusterStateRegistry: clusterState,
				options: ScaleUpOptions{
					"ng1": 5,
					"ng2": 8,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ExecuteScaleUp(tt.args.context, tt.args.clusterStateRegistry,
				tt.args.options); (err != nil) != tt.wantErr {
				t.Errorf("ExecuteScaleUp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecuteScaleDown(t *testing.T) {
	readyNodeLister := kubernetes.NewTestNodeLister(nil)
	allNodeLister := kubernetes.NewTestNodeLister(nil)
	scheduledPodMock := &podListerMock{}
	unschedulablePodMock := &podListerMock{}
	podDisruptionBudgetListerMock := &podDisruptionBudgetListerMock{}
	daemonSetListerMock := &daemonSetListerMock{}
	onScaleUpMock := &onScaleUpMock{}
	onScaleDownMock := &onScaleDownMock{}

	onScaleDownMock.On("ScaleDown", "ng1", "n1").Return(nil).Once()
	onScaleDownMock.On("ScaleDown", "ng2", "n2").Return(nil).Once()

	p1 := BuildTestPod("p1", 600, 100)
	p1.Spec.NodeName = "n1"
	p2 := BuildTestPod("p2", 600, 100)
	p2.Spec.NodeName = "n2"

	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Now())
	injectNodeIP(n1, "n1")
	tn1 := schedulernodeinfo.NewNodeInfo(p1)
	tn1.SetNode(n1)
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Now())
	injectNodeIP(n2, "n2")
	n2.ObjectMeta.Annotations = map[string]string{
		"io.tencent.bcs.dev/node-deletion-cost": "200",
	}
	tn2 := schedulernodeinfo.NewNodeInfo(p2)
	tn2.SetNode(n2)
	n3 := BuildTestNode("n3", 1000, 1000)
	injectNodeIP(n3, "n3")
	n3.ObjectMeta.Annotations = map[string]string{
		"io.tencent.bcs.dev/node-deletion-cost": "30",
	}
	SetNodeReadyState(n3, true, time.Now())
	tn3 := schedulernodeinfo.NewNodeInfo()
	tn3.SetNode(n3)
	n4 := BuildTestNode("n4", 1000, 1000)
	injectNodeIP(n4, "n4")
	n4.ObjectMeta.Annotations = map[string]string{
		"io.tencent.bcs.dev/node-deletion-cost": "200",
	}
	SetNodeReadyState(n4, true, time.Now())
	tn4 := schedulernodeinfo.NewNodeInfo()
	tn4.SetNode(n4)

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

	type args struct {
		context            *contextinternal.Context
		sd                 *ScaleDown
		nodes              []*corev1.Node
		candidates         ScaleDownCandidates
		nodeNameToNodeInfo map[string]*schedulernodeinfo.NodeInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "scale down ng1 normally",
			args: args{
				context:    &context,
				sd:         sd,
				nodes:      []*corev1.Node{n1, n2, n3, n4},
				candidates: ScaleDownCandidates{"n1", "n2"},
				nodeNameToNodeInfo: map[string]*schedulernodeinfo.NodeInfo{
					"n1": tn1,
					"n2": tn2,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ExecuteScaleDown(tt.args.context, tt.args.sd, tt.args.nodes, tt.args.candidates, tt.args.nodeNameToNodeInfo); (err != nil) != tt.wantErr {
				t.Errorf("ExecuteScaleDown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
