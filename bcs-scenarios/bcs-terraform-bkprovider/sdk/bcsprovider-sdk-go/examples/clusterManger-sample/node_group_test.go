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

// Package clusterMangerSample 测试
package clusterMangerSample

import (
	"context"
	"log"
	"testing"

	pb "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/utils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/service/clusterManger"
)

func Test_CreateNodeGroup(t *testing.T) {
	req := &pb.CreateNodeGroupRequest{
		ClusterID: clusterID,
		Name:      ngNameCase,
		AutoScaling: &pb.AutoScalingGroup{
			// 设置上、下限和期望节点数
			MinSize:     0,
			MaxSize:     10,
			DesiredSize: 0,
		},
		LaunchTemplate: &pb.LaunchConfiguration{
			InstanceType:      "SA2.LARGE8",
			InitLoginPassword: ngPasswordCase,
			SecurityGroupIDs: []string{
				securityGroupID,
			},
			// 仅为当前机型
			InstanceChargeType: clusterManger.PostpaidByHour,
		},
		EnableAutoscale: false,
	}

	resp, err := service.CreateNodeGroup(context.TODO(), req)
	if err != nil {
		t.Fatalf("create node group failed, err: %s", err.Error())
	}

	log.Printf("create node group success. resp: %s", utils.ObjToPrettyJson(resp))
}

func Test_DeleteNodeGroup(t *testing.T) {
	req := &pb.DeleteNodeGroupRequest{
		NodeGroupID: nodeGroupID,
	}

	resp, err := service.DeleteNodeGroup(context.TODO(), req)
	if err != nil {
		t.Fatalf("remove node group failed, err: %s", err.Error())
	}

	log.Printf("remove node group success. resp: %s", utils.ObjToPrettyJson(resp))
}

func Test_UpdateNodeGroup(t *testing.T) {
	req := &pb.UpdateNodeGroupRequest{
		NodeGroupID: nodeGroupID,
		ClusterID:   clusterID,
		/*
			支持修改参数, 如下:
		*/
		Name: "test-tfprovider-rename-ng",
		Labels: map[string]string{
			"class": "tfprovider-test",
		},
		Taints:          map[string]string{},
		Tags:            map[string]string{},
		EnableAutoscale: wrapperspb.Bool(false),
		AutoScaling: &pb.AutoScalingGroup{
			MinSize:               0,
			MaxSize:               5,
			DesiredSize:           0,
			MultiZoneSubnetPolicy: clusterManger.Priority,
			RetryPolicy:           clusterManger.ImmediateRetry,
			ScalingMode:           clusterManger.ClassicScaling,
		},
		NodeTemplate:   &pb.NodeTemplate{},
		LaunchTemplate: &pb.LaunchConfiguration{},
	}

	resp, err := service.UpdateNodeGroup(context.TODO(), req)
	if err != nil {
		t.Fatalf("update node group failed, err: %s", err.Error())
	}

	log.Printf("update node group success. resp: %s", utils.ObjToPrettyJson(resp))
}

func Test_UpdateGroupDesiredNode(t *testing.T) {
	req := &pb.UpdateGroupDesiredNodeRequest{
		NodeGroupID: nodeGroupID,
		DesiredNode: 0,
	}

	resp, err := service.UpdateGroupDesiredNode(context.TODO(), req)
	if err != nil {
		t.Fatalf("update desired node number failed, err: %s", err.Error())
	}

	log.Printf("update desired node number success. resp: %s", utils.ObjToPrettyJson(resp))
}

func Test_UpdateGroupMinMaxSize(t *testing.T) {
	req := &pb.UpdateGroupMinMaxSizeRequest{
		NodeGroupID: nodeGroupID,
		MinSize:     0,
		MaxSize:     1,
	}

	resp, err := service.UpdateGroupMinMaxSize(context.TODO(), req)
	if err != nil {
		t.Fatalf("update group max or min size failed, err: %s", err.Error())
	}

	log.Printf("update group max or min size success. resp: %s", utils.ObjToPrettyJson(resp))
}

func Test_GetNodeGroup(t *testing.T) {
	req := &pb.GetNodeGroupRequest{
		NodeGroupID: nodeGroupID,
	}

	resp, err := service.GetNodeGroup(context.TODO(), req)
	if err != nil {
		t.Fatalf("get node group failed, err: %s", err.Error())
	}

	log.Printf("get node group success. resp: %s", utils.ObjToPrettyJson(resp))
}

func Test_ListClusterNodeGroup(t *testing.T) {
	req := &pb.ListClusterNodeGroupRequest{
		ClusterID:    clusterID,
		EnableFilter: true,
	}

	resp, err := service.ListClusterNodeGroup(context.TODO(), req)
	if err != nil {
		t.Fatalf("list node group failed, err: %s", err.Error())
	}

	log.Printf("list node group success. resp: %s", utils.ObjToPrettyJson(resp))
}
