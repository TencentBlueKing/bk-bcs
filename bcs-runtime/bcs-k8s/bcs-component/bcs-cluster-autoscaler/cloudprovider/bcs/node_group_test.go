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

package bcs

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/cloudprovider/bcs/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/cloudprovider/bcs/clustermanager/mocks"
	"github.com/golang/mock/gomock"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

func TestNodeGroup_MaxSize(t *testing.T) {
	type fields struct {
		InstanceRef  InstanceRef
		scalingType  string
		instanceType string
		nodeGroupID  string
		minSize      int
		maxSize      int
		closedSize   int
		soldout      bool
		nodeCache    map[string]string
		client       clustermanager.NodePoolClientInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		// TODO: Add test cases.
		{
			name: "get max size 10",
			fields: fields{
				maxSize: 10,
			},
			want: 10,
		},
		{
			name: "get max size 999999",
			fields: fields{
				maxSize: 999999,
			},
			want: 999999,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := &NodeGroup{
				InstanceRef:  tt.fields.InstanceRef,
				scalingType:  tt.fields.scalingType,
				instanceType: tt.fields.instanceType,
				nodeGroupID:  tt.fields.nodeGroupID,
				minSize:      tt.fields.minSize,
				maxSize:      tt.fields.maxSize,
				closedSize:   tt.fields.closedSize,
				soldout:      tt.fields.soldout,
				nodeCache:    tt.fields.nodeCache,
				client:       tt.fields.client,
			}
			if got := group.MaxSize(); got != tt.want {
				t.Errorf("NodeGroup.MaxSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeGroup_MinSize(t *testing.T) {
	type fields struct {
		InstanceRef  InstanceRef
		scalingType  string
		instanceType string
		nodeGroupID  string
		minSize      int
		maxSize      int
		closedSize   int
		soldout      bool
		nodeCache    map[string]string
		client       clustermanager.NodePoolClientInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "get min size 0",
			fields: fields{
				minSize: 0,
			},
			want: 0,
		},
		{
			name: "get min size 10",
			fields: fields{
				minSize: 10,
			},
			want: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := &NodeGroup{
				InstanceRef:  tt.fields.InstanceRef,
				scalingType:  tt.fields.scalingType,
				instanceType: tt.fields.instanceType,
				nodeGroupID:  tt.fields.nodeGroupID,
				minSize:      tt.fields.minSize,
				maxSize:      tt.fields.maxSize,
				closedSize:   tt.fields.closedSize,
				soldout:      tt.fields.soldout,
				client:       tt.fields.client,
			}
			if got := group.MinSize(); got != tt.want {
				t.Errorf("NodeGroup.MinSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeGroup_TargetSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := mocks.NewMockNodePoolClientInterface(ctrl)
	m.EXPECT().GetPoolConfig(gomock.Eq("test1")).Return(
		func() *clustermanager.AutoScalingGroup {
			return &clustermanager.AutoScalingGroup{
				DesiredSize: 10,
			}
		}(), nil,
	)
	m.EXPECT().GetPoolConfig(gomock.Eq("test2")).Return(
		nil, fmt.Errorf("Internal Error"),
	)

	type fields struct {
		InstanceRef  InstanceRef
		scalingType  string
		instanceType string
		nodeGroupID  string
		minSize      int
		maxSize      int
		closedSize   int
		soldout      bool
		nodeCache    map[string]string
		client       clustermanager.NodePoolClientInterface
	}
	tests := []struct {
		name    string
		fields  fields
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "get target normal",
			fields: fields{
				nodeGroupID: "test1",
				client:      m,
			},
			want:    10,
			wantErr: false,
		},
		{
			name: "get target abnormal",
			fields: fields{
				nodeGroupID: "test2",
				client:      m,
			},
			want:    -1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := &NodeGroup{
				InstanceRef:  tt.fields.InstanceRef,
				scalingType:  tt.fields.scalingType,
				instanceType: tt.fields.instanceType,
				nodeGroupID:  tt.fields.nodeGroupID,
				minSize:      tt.fields.minSize,
				maxSize:      tt.fields.maxSize,
				closedSize:   tt.fields.closedSize,
				soldout:      tt.fields.soldout,
				nodeCache:    tt.fields.nodeCache,
				client:       tt.fields.client,
			}
			got, err := group.TargetSize()
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeGroup.TargetSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NodeGroup.TargetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeGroup_IncreaseSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := mocks.NewMockNodePoolClientInterface(ctrl)
	m.EXPECT().GetPoolConfig(gomock.Eq("test1")).Return(
		func() *clustermanager.AutoScalingGroup {
			return &clustermanager.AutoScalingGroup{
				DesiredSize: 1,
			}
		}(), nil,
	).Times(4)
	m.EXPECT().UpdateDesiredNode(gomock.Eq("test1"), gomock.Eq(4)).Return(nil).Times(2)

	m.EXPECT().GetPoolConfig(gomock.Eq("test2")).Return(
		func() *clustermanager.AutoScalingGroup {
			return &clustermanager.AutoScalingGroup{
				DesiredSize: 1,
			}
		}(), nil,
	)
	m.EXPECT().UpdateDesiredNode(gomock.Eq("test2"), gomock.Eq(1)).Return(nil)

	m.EXPECT().GetPoolConfig(gomock.Eq("test3")).Return(
		func() *clustermanager.AutoScalingGroup {
			return &clustermanager.AutoScalingGroup{
				DesiredSize: 1,
			}
		}(), nil,
	)
	m.EXPECT().UpdateDesiredNode(gomock.Eq("test3"), gomock.Eq(1)).Return(fmt.Errorf("Internal Error"))

	type fields struct {
		InstanceRef  InstanceRef
		scalingType  string
		instanceType string
		nodeGroupID  string
		minSize      int
		maxSize      int
		closedSize   int
		soldout      bool
		nodeCache    map[string]string
		client       clustermanager.NodePoolClientInterface
	}
	type args struct {
		delta int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "increase size normal",
			fields: fields{
				nodeGroupID: "test1",
				maxSize:     10,
				client:      m,
			},
			args: args{
				delta: 3,
			},
			wantErr: false,
		},
		{
			name: "increase size negative",
			fields: fields{
				nodeGroupID: "test1",
				maxSize:     10,
				client:      m,
			},
			args: args{
				delta: -1,
			},
			wantErr: true,
		},
		{
			name: "increase size too large",
			fields: fields{
				nodeGroupID: "test1",
				maxSize:     10,
				client:      m,
			},
			args: args{
				delta: 20,
			},
			wantErr: true,
		},
		{
			name: "soldout and not wake up stop",
			fields: fields{
				nodeGroupID: "test1",
				maxSize:     10,
				scalingType: ScalingTypeClassic,
				soldout:     true,
				client:      m,
			},
			args: args{
				delta: 3,
			},
			wantErr: true,
		},
		{
			name: "soldout, wake up stop, update normal",
			fields: fields{
				nodeGroupID: "test2",
				maxSize:     10,
				scalingType: ScalingTypeWakeUpStopped,
				soldout:     true,
				client:      m,
			},
			args: args{
				delta: 3,
			},
			wantErr: true,
		},
		{
			name: "soldout, wake up stop, update abnormal",
			fields: fields{
				nodeGroupID: "test3",
				maxSize:     10,
				scalingType: ScalingTypeWakeUpStopped,
				soldout:     true,
				client:      m,
			},
			args: args{
				delta: 4,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := &NodeGroup{
				InstanceRef:  tt.fields.InstanceRef,
				scalingType:  tt.fields.scalingType,
				instanceType: tt.fields.instanceType,
				nodeGroupID:  tt.fields.nodeGroupID,
				minSize:      tt.fields.minSize,
				maxSize:      tt.fields.maxSize,
				closedSize:   tt.fields.closedSize,
				soldout:      tt.fields.soldout,
				nodeCache:    tt.fields.nodeCache,
				client:       tt.fields.client,
			}
			if err := group.IncreaseSize(tt.args.delta); (err != nil) != tt.wantErr {
				t.Errorf("NodeGroup.IncreaseSize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeGroup_buildNodeFromTemplate(t *testing.T) {
	testTime := time.Now()
	monkey.Patch(time.Now, func() time.Time {
		return testTime
	})
	testRand := int64(123)
	monkey.Patch(rand.Int63, func() int64 {
		return testRand
	})
	type fields struct {
		InstanceRef  InstanceRef
		scalingType  string
		instanceType string
		nodeGroupID  string
		minSize      int
		maxSize      int
		closedSize   int
		soldout      bool
		nodeCache    map[string]string
		client       clustermanager.NodePoolClientInterface
	}
	type args struct {
		template *nodeTemplate
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *apiv1.Node
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "build node from template",
			fields: fields{
				nodeGroupID: "test",
			},
			args: args{
				template: func() *nodeTemplate {
					return &nodeTemplate{
						InstanceType: "xx",
						Region:       "SZ",
						Resources: map[apiv1.ResourceName]resource.Quantity{
							apiv1.ResourceCPU:     *resource.NewQuantity(10, resource.DecimalSI),
							gpu.ResourceNvidiaGPU: *resource.NewQuantity(0, resource.DecimalSI),
						},
					}
				}(),
			},
			want: &apiv1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:     fmt.Sprintf("%s-%d", "test", rand.Int63()),
					SelfLink: "/api/v1/nodes/" + fmt.Sprintf("%s-%d", "test", rand.Int63()),
					Labels: map[string]string{
						"beta.kubernetes.io/arch":                  "amd64",
						"beta.kubernetes.io/instance-type":         "xx",
						"beta.kubernetes.io/os":                    "linux",
						"failure-domain.beta.kubernetes.io/region": "SZ",
						"kubernetes.io/hostname":                   fmt.Sprintf("%s-%d", "test", rand.Int63()),
					},
				},
				Status: apiv1.NodeStatus{
					Capacity: apiv1.ResourceList{
						apiv1.ResourceCPU:     *resource.NewQuantity(10, resource.DecimalSI),
						apiv1.ResourcePods:    *resource.NewQuantity(110, resource.DecimalSI),
						gpu.ResourceNvidiaGPU: *resource.NewQuantity(0, resource.DecimalSI),
					},
					Allocatable: apiv1.ResourceList{
						apiv1.ResourceCPU:     *resource.NewQuantity(10, resource.DecimalSI),
						apiv1.ResourcePods:    *resource.NewQuantity(110, resource.DecimalSI),
						gpu.ResourceNvidiaGPU: *resource.NewQuantity(0, resource.DecimalSI),
					},
					Conditions: cloudprovider.BuildReadyConditions(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := &NodeGroup{
				InstanceRef:  tt.fields.InstanceRef,
				scalingType:  tt.fields.scalingType,
				instanceType: tt.fields.instanceType,
				nodeGroupID:  tt.fields.nodeGroupID,
				minSize:      tt.fields.minSize,
				maxSize:      tt.fields.maxSize,
				closedSize:   tt.fields.closedSize,
				soldout:      tt.fields.soldout,
				nodeCache:    tt.fields.nodeCache,
				client:       tt.fields.client,
			}
			got, err := group.buildNodeFromTemplate(tt.args.template)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeGroup.buildNodeFromTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeGroup.buildNodeFromTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getIP(t *testing.T) {
	type args struct {
		node *apiv1.Node
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "node with InterlIP",
			args: args{
				node: func() *apiv1.Node {
					tmp := apiv1.Node{
						Status: apiv1.NodeStatus{
							Addresses: []apiv1.NodeAddress{
								{
									Type:    apiv1.NodeInternalIP,
									Address: "127.0.0.1",
								},
							},
						},
					}
					return &tmp
				}(),
			},
			want: "127.0.0.1",
		},
		{
			name: "node with ExternalIP",
			args: args{
				node: func() *apiv1.Node {
					tmp := apiv1.Node{
						Status: apiv1.NodeStatus{
							Addresses: []apiv1.NodeAddress{
								{
									Type:    apiv1.NodeExternalIP,
									Address: "127.0.0.1",
								},
							},
						},
					}
					return &tmp
				}(),
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getIP(tt.args.node); got != tt.want {
				t.Errorf("getIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeGroup_DecreaseTargetSize(t *testing.T) {
	n1 := clustermanager.Node{
		NodeID:      "n1",
		NodeGroupID: "test",
		Status:      "running",
	}
	n2 := clustermanager.Node{
		NodeID:      "n2",
		NodeGroupID: "test",
		Status:      "failing",
	}
	n3 := clustermanager.Node{
		NodeID:      "n3",
		NodeGroupID: "test2",
		Status:      "running",
	}
	n4 := clustermanager.Node{
		NodeID:      "n4",
		NodeGroupID: "test",
		Status:      "creating",
	}
	ctrl := gomock.NewController(t)
	m := mocks.NewMockNodePoolClientInterface(ctrl)
	m.EXPECT().GetPoolConfig(gomock.Eq("test")).Return(
		func() *clustermanager.AutoScalingGroup {
			return &clustermanager.AutoScalingGroup{
				DesiredSize: 5,
			}
		}(), nil,
	).Times(2)
	m.EXPECT().GetNodes(gomock.Eq("test")).Return(
		[]*clustermanager.Node{&n1, &n2, &n3, &n4}, nil,
	).Times(2)
	m.EXPECT().UpdateDesiredSize(gomock.Eq("test"), gomock.Eq(4)).Return(nil)
	type fields struct {
		InstanceRef  InstanceRef
		scalingType  string
		instanceType string
		nodeGroupID  string
		minSize      int
		maxSize      int
		closedSize   int
		soldout      bool
		nodeCache    map[string]string
		client       clustermanager.NodePoolClientInterface
	}
	type args struct {
		delta int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "negative delta, normal",
			fields: fields{
				nodeGroupID: "test",
				client:      m,
			},
			args: args{
				delta: -1,
			},
			wantErr: false,
		},
		{
			name: "negative delta, abnormal",
			fields: fields{
				nodeGroupID: "test",
				client:      m,
			},
			args: args{
				delta: -4,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := &NodeGroup{
				InstanceRef:  tt.fields.InstanceRef,
				scalingType:  tt.fields.scalingType,
				instanceType: tt.fields.instanceType,
				nodeGroupID:  tt.fields.nodeGroupID,
				minSize:      tt.fields.minSize,
				maxSize:      tt.fields.maxSize,
				closedSize:   tt.fields.closedSize,
				soldout:      tt.fields.soldout,
				nodeCache:    tt.fields.nodeCache,
				client:       tt.fields.client,
			}
			if err := group.DecreaseTargetSize(tt.args.delta); (err != nil) != tt.wantErr {
				t.Errorf("NodeGroup.DecreaseTargetSize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeGroup_Belongs(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := mocks.NewMockNodePoolClientInterface(ctrl)
	m.EXPECT().GetNode(gomock.Eq("127.0.0.1")).Return(
		&clustermanager.Node{
			NodeGroupID: "test",
			InnerIP:     "127.0.0.1",
		}, nil,
	).Times(2)
	type fields struct {
		InstanceRef  InstanceRef
		scalingType  string
		instanceType string
		nodeGroupID  string
		minSize      int
		maxSize      int
		closedSize   int
		soldout      bool
		nodeCache    map[string]string
		client       clustermanager.NodePoolClientInterface
	}
	type args struct {
		node *apiv1.Node
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Belongs normal",
			fields: fields{
				nodeGroupID: "test",
				client:      m,
			},
			args: args{
				node: &apiv1.Node{
					Status: apiv1.NodeStatus{
						Addresses: []apiv1.NodeAddress{
							{
								Type:    apiv1.NodeInternalIP,
								Address: "127.0.0.1",
							},
						},
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Belongs abnormal",
			fields: fields{
				nodeGroupID: "test2",
				client:      m,
			},
			args: args{
				node: &apiv1.Node{
					Status: apiv1.NodeStatus{
						Addresses: []apiv1.NodeAddress{
							{
								Type:    apiv1.NodeInternalIP,
								Address: "127.0.0.1",
							},
						},
					},
				},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Belongs empty ip, wrong provider id",
			fields: fields{
				nodeGroupID: "test2",
				client:      m,
			},
			args: args{
				node: &apiv1.Node{
					Spec: apiv1.NodeSpec{
						ProviderID: "abcd",
					},
				},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Belongs empty ip, empty node cache",
			fields: fields{
				nodeGroupID: "test2",
				client:      m,
			},
			args: args{
				node: &apiv1.Node{
					Spec: apiv1.NodeSpec{
						ProviderID: "qcloud:///100003/ins-3ven36lk",
					},
				},
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := &NodeGroup{
				InstanceRef:  tt.fields.InstanceRef,
				scalingType:  tt.fields.scalingType,
				instanceType: tt.fields.instanceType,
				nodeGroupID:  tt.fields.nodeGroupID,
				minSize:      tt.fields.minSize,
				maxSize:      tt.fields.maxSize,
				closedSize:   tt.fields.closedSize,
				soldout:      tt.fields.soldout,
				nodeCache:    tt.fields.nodeCache,
				client:       tt.fields.client,
			}
			got, err := group.Belongs(tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeGroup.Belongs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NodeGroup.Belongs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeGroup_DeleteNodes(t *testing.T) {
	n1 := apiv1.Node{
		Status: apiv1.NodeStatus{
			Addresses: []apiv1.NodeAddress{
				{
					Type:    apiv1.NodeInternalIP,
					Address: "127.0.0.1",
				},
			},
		},
	}

	n2 := apiv1.Node{
		Status: apiv1.NodeStatus{
			Addresses: []apiv1.NodeAddress{
				{
					Type:    apiv1.NodeInternalIP,
					Address: "127.0.0.2",
				},
			},
		},
	}
	n3 := apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "qcloud:///100003/ins-3ven36lk",
		},
	}
	ctrl := gomock.NewController(t)
	m := mocks.NewMockNodePoolClientInterface(ctrl)
	m.EXPECT().GetPoolConfig(gomock.Eq("test")).Return(
		func() *clustermanager.AutoScalingGroup {
			return &clustermanager.AutoScalingGroup{
				DesiredSize: 5,
				MinSize:     0,
			}
		}(), nil,
	).Times(4)
	m.EXPECT().GetNode(gomock.Eq("127.0.0.1")).Return(
		&clustermanager.Node{
			NodeGroupID: "test",
			InnerIP:     "127.0.0.1",
		}, nil,
	)
	m.EXPECT().GetNode(gomock.Eq("127.0.0.2")).Return(
		&clustermanager.Node{
			NodeGroupID: "test2",
			InnerIP:     "127.0.0.2",
		}, nil,
	)
	m.EXPECT().GetNode(gomock.Eq("127.0.0.3")).Return(
		&clustermanager.Node{
			NodeGroupID: "test",
			InnerIP:     "127.0.0.3",
		}, nil,
	)
	m.EXPECT().RemoveNodes(gomock.Eq("test"), gomock.Eq([]string{"127.0.0.1"})).Return(nil)
	m.EXPECT().RemoveNodes(gomock.Eq("test"), gomock.Eq([]string{"127.0.0.3"})).Return(
		fmt.Errorf("remove node failed"),
	)
	type fields struct {
		InstanceRef  InstanceRef
		scalingType  string
		instanceType string
		nodeGroupID  string
		minSize      int
		maxSize      int
		closedSize   int
		soldout      bool
		nodeCache    map[string]string
		client       clustermanager.NodePoolClientInterface
	}
	type args struct {
		nodes []*apiv1.Node
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "delete nodes normal",
			fields: fields{
				nodeGroupID: "test",
				minSize:     0,
				client:      m,
			},
			args: args{
				nodes: []*apiv1.Node{&n1},
			},
			wantErr: false,
		},
		{
			name: "delete nodes with wrong nodegroup id",
			fields: fields{
				nodeGroupID: "test",
				minSize:     0,
				client:      m,
			},
			args: args{
				nodes: []*apiv1.Node{&n2},
			},
			wantErr: true,
		},
		{
			name: "delete nodes with empty ip",
			fields: fields{
				nodeGroupID: "test",
				minSize:     0,
				nodeCache: map[string]string{
					"ins-3ven36lk": "127.0.0.3",
				},
				client: m,
			},
			args: args{
				nodes: []*apiv1.Node{&n3},
			},
			wantErr: true,
		},
		{
			name: "minsize reach",
			fields: fields{
				nodeGroupID: "test",
				minSize:     5,
				nodeCache:   map[string]string{},
				client:      m,
			},
			args: args{
				nodes: []*apiv1.Node{&n3},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := &NodeGroup{
				InstanceRef:  tt.fields.InstanceRef,
				scalingType:  tt.fields.scalingType,
				instanceType: tt.fields.instanceType,
				nodeGroupID:  tt.fields.nodeGroupID,
				minSize:      tt.fields.minSize,
				maxSize:      tt.fields.maxSize,
				closedSize:   tt.fields.closedSize,
				soldout:      tt.fields.soldout,
				nodeCache:    tt.fields.nodeCache,
				client:       tt.fields.client,
			}
			if err := group.DeleteNodes(tt.args.nodes); (err != nil) != tt.wantErr {
				t.Errorf("NodeGroup.DeleteNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeGroup_Nodes(t *testing.T) {
	n1 := clustermanager.Node{
		NodeID:      "n1",
		Zone:        123,
		NodeGroupID: "test",
		InnerIP:     "127.0.0.1",
		Status:      "running",
	}
	n2 := clustermanager.Node{
		NodeID:      "n2",
		Zone:        123,
		NodeGroupID: "test",
		InnerIP:     "127.0.0.2",
		Status:      "creating",
	}
	n3 := clustermanager.Node{
		NodeID:      "n3",
		Zone:        123,
		NodeGroupID: "test",
		InnerIP:     "127.0.0.3",
		Status:      "deleting",
	}
	n4 := clustermanager.Node{
		NodeID:      "n4",
		Zone:        123,
		NodeGroupID: "test2",
		InnerIP:     "127.0.0.4",
		Status:      "running",
	}
	n5 := clustermanager.Node{
		NodeID:      "n5",
		Zone:        123,
		NodeGroupID: "test",
		InnerIP:     "127.0.0.5",
		Status:      "DELETING",
	}
	i1 := cloudprovider.Instance{
		Id: "qcloud:///123/n1",
		Status: &cloudprovider.InstanceStatus{
			State: cloudprovider.InstanceRunning,
		},
	}
	i2 := cloudprovider.Instance{
		Id: "qcloud:///123/n2",
		Status: &cloudprovider.InstanceStatus{
			State: cloudprovider.InstanceCreating,
		},
	}
	i3 := cloudprovider.Instance{
		Id: "qcloud:///123/n3",
		Status: &cloudprovider.InstanceStatus{
			State: cloudprovider.InstanceDeleting,
		},
	}
	ctrl := gomock.NewController(t)
	m := mocks.NewMockNodePoolClientInterface(ctrl)
	m.EXPECT().GetNodes(gomock.Eq("test")).Return(
		[]*clustermanager.Node{&n1, &n2, &n3, &n4, &n5}, nil,
	)
	m.EXPECT().GetNodes(gomock.Eq("test2")).Return(
		[]*clustermanager.Node{}, fmt.Errorf("get nodes falied"),
	)
	type fields struct {
		InstanceRef  InstanceRef
		scalingType  string
		instanceType string
		nodeGroupID  string
		minSize      int
		maxSize      int
		closedSize   int
		soldout      bool
		nodeCache    map[string]string
		client       clustermanager.NodePoolClientInterface
	}
	tests := []struct {
		name    string
		fields  fields
		want    []cloudprovider.Instance
		wantErr bool
	}{
		{
			name: "get nodes with nodegroup test",
			fields: fields{
				nodeGroupID: "test",
				client:      m,
			},
			want:    []cloudprovider.Instance{i1, i2, i3},
			wantErr: false,
		},
		{
			name: "get nodes with nodegroup test2",
			fields: fields{
				nodeGroupID: "test2",
				client:      m,
			},
			want:    []cloudprovider.Instance{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := &NodeGroup{
				InstanceRef:  tt.fields.InstanceRef,
				scalingType:  tt.fields.scalingType,
				instanceType: tt.fields.instanceType,
				nodeGroupID:  tt.fields.nodeGroupID,
				minSize:      tt.fields.minSize,
				maxSize:      tt.fields.maxSize,
				closedSize:   tt.fields.closedSize,
				soldout:      tt.fields.soldout,
				nodeCache:    tt.fields.nodeCache,
				client:       tt.fields.client,
			}
			got, err := group.Nodes()
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeGroup.Nodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeGroup.Nodes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeGroup_TemplateNodeInfo(t *testing.T) {
	nodeGroup := clustermanager.NodeGroup{
		NodeGroupID: "test",
		Name:        "test",
		Region:      "nj",
		AutoScaling: &clustermanager.AutoScalingGroup{
			MaxSize:     5,
			MinSize:     0,
			DesiredSize: 2,
		},
		LaunchTemplate: &clustermanager.LaunchConfiguration{
			CPU:          10,
			InstanceType: "xx",
		},
	}
	node := apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:     fmt.Sprintf("%s-%d", "test", rand.Int63()),
			SelfLink: "/api/v1/nodes/" + fmt.Sprintf("%s-%d", "test", rand.Int63()),
			Labels: map[string]string{
				"beta.kubernetes.io/arch":                  "amd64",
				"beta.kubernetes.io/instance-type":         "xx",
				"beta.kubernetes.io/os":                    "linux",
				"failure-domain.beta.kubernetes.io/region": "nj",
				"kubernetes.io/hostname":                   fmt.Sprintf("%s-%d", "test", rand.Int63()),
			},
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:     *resource.NewQuantity(10, resource.DecimalSI),
				apiv1.ResourcePods:    *resource.NewQuantity(110, resource.DecimalSI),
				gpu.ResourceNvidiaGPU: *resource.NewQuantity(0, resource.DecimalSI),
			},
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:     *resource.NewQuantity(10, resource.DecimalSI),
				apiv1.ResourcePods:    *resource.NewQuantity(110, resource.DecimalSI),
				gpu.ResourceNvidiaGPU: *resource.NewQuantity(0, resource.DecimalSI),
			},
			Conditions: cloudprovider.BuildReadyConditions(),
		},
	}
	info := nodeinfo.NewNodeInfo()
	info.SetNode(&node)

	ctrl := gomock.NewController(t)
	m := mocks.NewMockNodePoolClientInterface(ctrl)
	m.EXPECT().GetPool(gomock.Eq("test")).Return(&nodeGroup, nil)

	type fields struct {
		InstanceRef  InstanceRef
		scalingType  string
		instanceType string
		nodeGroupID  string
		minSize      int
		maxSize      int
		closedSize   int
		soldout      bool
		nodeCache    map[string]string
		client       clustermanager.NodePoolClientInterface
	}
	tests := []struct {
		name    string
		fields  fields
		want    *nodeinfo.NodeInfo
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "TemplateNodeInfo normal",
			fields: fields{
				nodeGroupID: "test",
				client:      m,
			},
			want:    info,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := &NodeGroup{
				InstanceRef:  tt.fields.InstanceRef,
				scalingType:  tt.fields.scalingType,
				instanceType: tt.fields.instanceType,
				nodeGroupID:  tt.fields.nodeGroupID,
				minSize:      tt.fields.minSize,
				maxSize:      tt.fields.maxSize,
				closedSize:   tt.fields.closedSize,
				soldout:      tt.fields.soldout,
				nodeCache:    tt.fields.nodeCache,
				client:       tt.fields.client,
			}
			got, err := group.TemplateNodeInfo()
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeGroup.TemplateNodeInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.AllocatableResource(), tt.want.AllocatableResource()) {
				t.Errorf("NodeGroup.TemplateNodeInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeGroup_Id(t *testing.T) {
	type fields struct {
		InstanceRef  InstanceRef
		scalingType  string
		instanceType string
		nodeGroupID  string
		minSize      int
		maxSize      int
		closedSize   int
		soldout      bool
		nodeCache    map[string]string
		client       clustermanager.NodePoolClientInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
		{
			name: "test Id",
			fields: fields{
				nodeGroupID: "test",
			},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := &NodeGroup{
				InstanceRef:  tt.fields.InstanceRef,
				scalingType:  tt.fields.scalingType,
				instanceType: tt.fields.instanceType,
				nodeGroupID:  tt.fields.nodeGroupID,
				minSize:      tt.fields.minSize,
				maxSize:      tt.fields.maxSize,
				closedSize:   tt.fields.closedSize,
				soldout:      tt.fields.soldout,
				nodeCache:    tt.fields.nodeCache,
				client:       tt.fields.client,
			}
			if got := group.Id(); got != tt.want {
				t.Errorf("NodeGroup.Id() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeGroup_Debug(t *testing.T) {
	type fields struct {
		InstanceRef  InstanceRef
		scalingType  string
		instanceType string
		nodeGroupID  string
		minSize      int
		maxSize      int
		closedSize   int
		soldout      bool
		nodeCache    map[string]string
		client       clustermanager.NodePoolClientInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test Debug",
			fields: fields{
				nodeGroupID: "test",
				minSize:     0,
				maxSize:     5,
			},
			want: "test (0:5)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := &NodeGroup{
				InstanceRef:  tt.fields.InstanceRef,
				scalingType:  tt.fields.scalingType,
				instanceType: tt.fields.instanceType,
				nodeGroupID:  tt.fields.nodeGroupID,
				minSize:      tt.fields.minSize,
				maxSize:      tt.fields.maxSize,
				closedSize:   tt.fields.closedSize,
				soldout:      tt.fields.soldout,
				nodeCache:    tt.fields.nodeCache,
				client:       tt.fields.client,
			}
			if got := group.Debug(); got != tt.want {
				t.Errorf("NodeGroup.Debug() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeGroup_GetNodeGroup(t *testing.T) {
	ng := clustermanager.NodeGroup{
		NodeGroupID:     "test",
		EnableAutoscale: true,
	}
	ctrl := gomock.NewController(t)
	m := mocks.NewMockNodePoolClientInterface(ctrl)
	m.EXPECT().GetPool(gomock.Eq("test")).Return(&ng, nil)
	type fields struct {
		InstanceRef  InstanceRef
		scalingType  string
		instanceType string
		nodeGroupID  string
		minSize      int
		maxSize      int
		closedSize   int
		soldout      bool
		nodeCache    map[string]string
		client       clustermanager.NodePoolClientInterface
	}
	tests := []struct {
		name    string
		fields  fields
		want    *clustermanager.NodeGroup
		wantErr bool
	}{
		{
			name: "test GetNodeGroup",
			fields: fields{
				nodeGroupID: "test",
				client:      m,
			},
			want:    &ng,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := &NodeGroup{
				InstanceRef:  tt.fields.InstanceRef,
				scalingType:  tt.fields.scalingType,
				instanceType: tt.fields.instanceType,
				nodeGroupID:  tt.fields.nodeGroupID,
				minSize:      tt.fields.minSize,
				maxSize:      tt.fields.maxSize,
				closedSize:   tt.fields.closedSize,
				soldout:      tt.fields.soldout,
				nodeCache:    tt.fields.nodeCache,
				client:       tt.fields.client,
			}
			got, err := group.GetNodeGroup()
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeGroup.GetNodeGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeGroup.GetNodeGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}
