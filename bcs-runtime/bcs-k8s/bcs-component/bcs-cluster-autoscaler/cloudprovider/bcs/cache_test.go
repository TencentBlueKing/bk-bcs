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
	"reflect"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/cloudprovider/bcs/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/cloudprovider/bcs/clustermanager/mocks"
	"github.com/golang/mock/gomock"
)

func TestNodeGroupCache_GetRegisteredNodeGroups(t *testing.T) {
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
		registeredGroups       []*NodeGroup
		instanceToGroup        map[InstanceRef]*NodeGroup
		instanceToCreationType map[InstanceRef]CreationType
		lastUpdateTime         time.Time
		getNodes               GetNodes
	}
	tests := []struct {
		name   string
		fields fields
		want   []*NodeGroup
	}{
		{
			name: "test GetRegisteredNodeGroups",
			fields: fields{
				registeredGroups: []*NodeGroup{&ng1, &ng2},
			},
			want: []*NodeGroup{&ng1, &ng2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &NodeGroupCache{
				registeredGroups:       tt.fields.registeredGroups,
				instanceToGroup:        tt.fields.instanceToGroup,
				instanceToCreationType: tt.fields.instanceToCreationType,
				getNodes:               tt.fields.getNodes,
			}
			if got := m.GetRegisteredNodeGroups(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeGroupCache.GetRegisteredNodeGroups() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestNodeGroupCache_FindForInstance(t *testing.T) {
	ng1 := NodeGroup{
		nodeGroupID: "test",
		maxSize:     5,
		minSize:     0,
	}
	ins1 := InstanceRef{
		Name: "ins1",
		IP:   "127.0.0.1",
	}
	ins2 := InstanceRef{
		Name: "ins2",
		IP:   "127.0.0.2",
	}
	testFunc := func(ng string) ([]*clustermanager.Node, error) {
		return nil, fmt.Errorf("failed GetNodes")
	}
	type args struct {
		instance *InstanceRef
	}
	type fields struct {
		registeredGroups       []*NodeGroup
		instanceToGroup        map[InstanceRef]*NodeGroup
		instanceToCreationType map[InstanceRef]CreationType
		lastUpdateTime         time.Time
		getNodes               GetNodes
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *NodeGroup
		wantErr bool
	}{
		{
			name: "test FindForInstance",
			fields: fields{
				instanceToGroup: map[InstanceRef]*NodeGroup{
					ins1: &ng1,
				},
			},
			args: args{
				instance: &ins1,
			},
			want:    &ng1,
			wantErr: false,
		},
		{
			name: "test FindForInstance with regenerate, abnormals",
			fields: fields{
				getNodes:         testFunc,
				lastUpdateTime:   time.Now().Add(-5 * time.Minute),
				registeredGroups: []*NodeGroup{&ng1},
			},
			args: args{
				instance: &ins2,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &NodeGroupCache{
				registeredGroups:       tt.fields.registeredGroups,
				instanceToGroup:        tt.fields.instanceToGroup,
				instanceToCreationType: tt.fields.instanceToCreationType,
				lastUpdateTime:         tt.fields.lastUpdateTime,
				getNodes:               tt.fields.getNodes,
			}
			got, err := m.FindForInstance(tt.args.instance)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeGroupCache.FindForInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeGroupCache.FindForInstance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeGroupCache_CheckInstancesTerminateByAs(t *testing.T) {
	ins1 := InstanceRef{
		Name: "ins1",
		IP:   "127.0.0.1",
	}
	ins2 := InstanceRef{
		Name: "ins2",
		IP:   "127.0.0.2",
	}
	type fields struct {
		registeredGroups       []*NodeGroup
		instanceToGroup        map[InstanceRef]*NodeGroup
		instanceToCreationType map[InstanceRef]CreationType
		lastUpdateTime         time.Time
		getNodes               GetNodes
	}
	type args struct {
		instances []*InstanceRef
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "instance terminated not by as",
			fields: fields{
				instanceToCreationType: map[InstanceRef]CreationType{
					ins1: CreationTypeAuto,
				},
			},
			args: args{
				instances: []*InstanceRef{&ins1},
			},
			want: false,
		},
		{
			name: "instance terminated by as",
			fields: fields{
				instanceToCreationType: map[InstanceRef]CreationType{
					ins2: CreationTypeManual,
				},
			},
			args: args{
				instances: []*InstanceRef{&ins2},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &NodeGroupCache{
				registeredGroups:       tt.fields.registeredGroups,
				instanceToGroup:        tt.fields.instanceToGroup,
				instanceToCreationType: tt.fields.instanceToCreationType,
				lastUpdateTime:         tt.fields.lastUpdateTime,
				getNodes:               tt.fields.getNodes,
			}
			if got := m.CheckInstancesTerminateByAs(tt.args.instances); got != tt.want {
				t.Errorf("NodeGroupCache.CheckInstancesTerminateByAs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeGroupCache_regenerateCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := mocks.NewMockNodePoolClientInterface(ctrl)
	m.EXPECT().GetNodes(gomock.Eq("test")).Return(
		[]*clustermanager.Node{}, fmt.Errorf("get nodes failed"),
	)
	m.EXPECT().GetNodes(gomock.Eq("test2")).Return(
		[]*clustermanager.Node{
			{
				NodeGroupID: "test2",
				NodeID:      "n2",
			},
		}, nil,
	)
	m.EXPECT().GetPool(gomock.Eq("test2")).Return(
		&clustermanager.NodeGroup{
			NodeGroupID: "test2",
		}, nil,
	)
	m.EXPECT().GetNodes(gomock.Eq("test3")).Return(
		[]*clustermanager.Node{
			{
				NodeGroupID: "test3",
				NodeID:      "n3",
			},
		}, nil,
	)
	m.EXPECT().GetPool(gomock.Eq("test3")).Return(
		&clustermanager.NodeGroup{
			NodeGroupID: "test3",
			AutoScaling: &clustermanager.AutoScalingGroup{
				MaxSize: 0,
			},
		}, nil,
	)
	m.EXPECT().GetNodes(gomock.Eq("test4")).Return(
		[]*clustermanager.Node{
			{
				NodeGroupID: "test4",
				NodeID:      "n4",
			},
		}, nil,
	)
	m.EXPECT().GetPool(gomock.Eq("test4")).Return(
		&clustermanager.NodeGroup{
			NodeGroupID: "test4",
			AutoScaling: &clustermanager.AutoScalingGroup{
				MaxSize: 5,
				MinSize: 0,
			},
		}, nil,
	)
	type fields struct {
		registeredGroups       []*NodeGroup
		instanceToGroup        map[InstanceRef]*NodeGroup
		instanceToCreationType map[InstanceRef]CreationType
		lastUpdateTime         time.Time
		getNodes               GetNodes
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "last update time too close",
			fields: fields{
				lastUpdateTime: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "failed to getNodes",
			fields: fields{
				registeredGroups: []*NodeGroup{
					{
						nodeGroupID: "test",
						client:      m,
					},
				},
				lastUpdateTime: time.Now().Add(-5 * time.Minute),
				getNodes:       m.GetNodes,
			},
			wantErr: true,
		},
		{
			name: "failed to get node group",
			fields: fields{
				registeredGroups: []*NodeGroup{
					{
						nodeGroupID: "test2",
						client:      m,
					},
				},
				lastUpdateTime: time.Now().Add(-5 * time.Minute),
				getNodes:       m.GetNodes,
			},
			wantErr: true,
		},
		{
			name: "maxsize equals to 0",
			fields: fields{
				registeredGroups: []*NodeGroup{
					{
						nodeGroupID: "test3",
						client:      m,
					},
				},
				lastUpdateTime: time.Now().Add(-5 * time.Minute),
				getNodes:       m.GetNodes,
			},
			wantErr: true,
		},
		{
			name: "regenerate normal",
			fields: fields{
				registeredGroups: []*NodeGroup{
					{
						nodeGroupID: "test4",
						client:      m,
					},
				},
				lastUpdateTime: time.Now().Add(-5 * time.Minute),
				getNodes:       m.GetNodes,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &NodeGroupCache{
				registeredGroups:       tt.fields.registeredGroups,
				instanceToGroup:        tt.fields.instanceToGroup,
				instanceToCreationType: tt.fields.instanceToCreationType,
				lastUpdateTime:         tt.fields.lastUpdateTime,
				getNodes:               tt.fields.getNodes,
			}
			if err := m.regenerateCache(); (err != nil) != tt.wantErr {
				t.Errorf("NodeGroupCache.regenerateCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeGroupCache_SetNodeGroupMinSize(t *testing.T) {
	ng1 := NodeGroup{
		nodeGroupID: "test",
		minSize:     0,
	}
	type fields struct {
		registeredGroups       []*NodeGroup
		instanceToGroup        map[InstanceRef]*NodeGroup
		instanceToCreationType map[InstanceRef]CreationType
		lastUpdateTime         time.Time
		getNodes               GetNodes
	}
	type args struct {
		groupID string
		num     int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "set minsize normal",
			fields: fields{
				registeredGroups: []*NodeGroup{&ng1},
			},
			args: args{
				groupID: "test",
				num:     3,
			},
			wantErr: false,
		},
		{
			name: "set minsize abnormal",
			fields: fields{
				registeredGroups: []*NodeGroup{},
			},
			args: args{
				groupID: "test",
				num:     3,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &NodeGroupCache{
				registeredGroups:       tt.fields.registeredGroups,
				instanceToGroup:        tt.fields.instanceToGroup,
				instanceToCreationType: tt.fields.instanceToCreationType,
				lastUpdateTime:         tt.fields.lastUpdateTime,
				getNodes:               tt.fields.getNodes,
			}
			if err := m.SetNodeGroupMinSize(tt.args.groupID, tt.args.num); (err != nil) != tt.wantErr {
				t.Errorf("NodeGroupCache.SetNodeGroupMinSize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
