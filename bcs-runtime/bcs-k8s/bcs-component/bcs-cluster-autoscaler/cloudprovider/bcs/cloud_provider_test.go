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
	"reflect"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/cloudprovider/bcs/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/cloudprovider/bcs/clustermanager/mocks"
	"github.com/golang/mock/gomock"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

func TestBuildBcsCloudProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := mocks.NewMockNodePoolClientInterface(ctrl)

	type args struct {
		cache           *NodeGroupCache
		client          clustermanager.NodePoolClientInterface
		discoveryOpts   cloudprovider.NodeGroupDiscoveryOptions
		resourceLimiter *cloudprovider.ResourceLimiter
	}
	tests := []struct {
		name    string
		args    args
		want    cloudprovider.CloudProvider
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "len(specs) <= 0",
			args: args{
				discoveryOpts: cloudprovider.NodeGroupDiscoveryOptions{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wrong specs",
			args: args{
				discoveryOpts: cloudprovider.NodeGroupDiscoveryOptions{
					NodeGroupSpecs: []string{"xx:xx:xx"},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "build cloud provider normal",
			args: args{
				cache:  NewNodeGroupCache(nil),
				client: m,
				discoveryOpts: cloudprovider.NodeGroupDiscoveryOptions{
					NodeGroupSpecs: []string{"0:5:test"},
				},
			},
			want: &Provider{
				NodeGroupCache: &NodeGroupCache{
					registeredGroups: []*NodeGroup{
						{
							InstanceRef: InstanceRef{
								Name: "test",
							},
							nodeGroupID: "test",
							scalingType: ScalingTypeClassic,
							maxSize:     5,
							minSize:     0,
							client:      m,
						},
					},
					instanceToGroup:        make(map[InstanceRef]*NodeGroup),
					instanceToCreationType: make(map[InstanceRef]CreationType),
					getNodes:               nil,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildBcsCloudProvider(tt.args.cache, tt.args.client, tt.args.discoveryOpts, tt.args.resourceLimiter)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildBcsCloudProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildBcsCloudProvider() = %+v, want %+v", got, tt.want)

			}
		})
	}
}

func TestProvider_NodeGroups(t *testing.T) {
	ng1 := NodeGroup{
		nodeGroupID: "test",
		maxSize:     5,
		minSize:     0,
	}
	ng2 := NodeGroup{
		nodeGroupID: "test2",
		maxSize:     10,
		minSize:     5,
	}
	type fields struct {
		NodeGroupCache  *NodeGroupCache
		resourceLimiter *cloudprovider.ResourceLimiter
	}
	tests := []struct {
		name   string
		fields fields
		want   []cloudprovider.NodeGroup
	}{
		// TODO: Add test cases.
		{
			name: "test NodeGroups",
			fields: fields{
				NodeGroupCache: &NodeGroupCache{
					registeredGroups: []*NodeGroup{&ng1, &ng2},
				},
			},
			want: []cloudprovider.NodeGroup{&ng1, &ng2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cloud := &Provider{
				NodeGroupCache:  tt.fields.NodeGroupCache,
				resourceLimiter: tt.fields.resourceLimiter,
			}
			if got := cloud.NodeGroups(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Provider.NodeGroups() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_NodeGroupForNode(t *testing.T) {
	ng1 := NodeGroup{
		nodeGroupID: "test",
		maxSize:     5,
		minSize:     0,
	}
	ins1 := InstanceRef{
		Name: "ins1",
	}
	type fields struct {
		NodeGroupCache  *NodeGroupCache
		resourceLimiter *cloudprovider.ResourceLimiter
	}
	type args struct {
		node *apiv1.Node
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    cloudprovider.NodeGroup
		wantErr bool
	}{
		{
			name: "wrong provider id",
			args: args{
				node: &apiv1.Node{
					Spec: apiv1.NodeSpec{
						ProviderID: "xx",
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "instance found in node group",
			fields: fields{
				NodeGroupCache: &NodeGroupCache{
					instanceToGroup: map[InstanceRef]*NodeGroup{
						ins1: &ng1,
					},
				},
			},
			args: args{
				node: &apiv1.Node{
					Spec: apiv1.NodeSpec{
						ProviderID: "qcloud:///100003/ins1",
					},
				},
			},
			want:    &ng1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cloud := &Provider{
				NodeGroupCache:  tt.fields.NodeGroupCache,
				resourceLimiter: tt.fields.resourceLimiter,
			}
			got, err := cloud.NodeGroupForNode(tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.NodeGroupForNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Provider.NodeGroupForNode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_Refresh(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := mocks.NewMockNodePoolClientInterface(ctrl)
	m.EXPECT().GetNodes(gomock.Eq("test")).Return(
		[]*clustermanager.Node{
			{
				NodeGroupID: "test",
				NodeID:      "n1",
			},
		}, nil,
	)
	m.EXPECT().GetPool(gomock.Eq("test")).Return(
		&clustermanager.NodeGroup{
			NodeGroupID: "test",
			AutoScaling: &clustermanager.AutoScalingGroup{
				MaxSize: 5,
				MinSize: 0,
			},
		}, nil,
	)
	type fields struct {
		NodeGroupCache  *NodeGroupCache
		resourceLimiter *cloudprovider.ResourceLimiter
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "node group cache is nil",
			fields:  fields{},
			wantErr: true,
		},
		{
			name: "fresh normal",
			fields: fields{
				NodeGroupCache: &NodeGroupCache{
					registeredGroups: []*NodeGroup{
						{
							nodeGroupID: "test",
							client:      m,
						},
					},
					lastUpdateTime: time.Now().Add(-5 * time.Minute),
					getNodes:       m.GetNodes,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cloud := &Provider{
				NodeGroupCache:  tt.fields.NodeGroupCache,
				resourceLimiter: tt.fields.resourceLimiter,
			}
			if err := cloud.Refresh(); (err != nil) != tt.wantErr {
				t.Errorf("Provider.Refresh() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
