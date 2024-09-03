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

package tasks

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	k8scorev1 "k8s.io/api/core/v1"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/azure/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

func importClusterNodesToCM(ctx context.Context, nodes []k8scorev1.Node, clusterID string) error {
	for _, n := range nodes {
		innerIP := ""
		for _, v := range n.Status.Addresses {
			if v.Type == k8scorev1.NodeInternalIP {
				innerIP = v.Address
				break
			}
		}
		if innerIP == "" {
			continue
		}
		node, err := cloudprovider.GetStorageModel().GetNodeByIP(ctx, innerIP)
		if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
			blog.Errorf("importClusterNodes GetNodeByIP[%s] failed: %v", innerIP, err)
			// no import node when found err
			continue
		}

		if node == nil {
			node = &proto.Node{
				InnerIP:   innerIP,
				Status:    common.StatusRunning,
				ClusterID: clusterID,
			}
			err = cloudprovider.GetStorageModel().CreateNode(ctx, node)
			if err != nil {
				blog.Errorf("importClusterNodes CreateNode[%s] failed: %v", innerIP, err)
			}
			continue
		}
	}

	return nil
}

func setModuleInfo(group *proto.NodeGroup, bkBizIDString string) {
	if group.NodeTemplate != nil && group.NodeTemplate.Module != nil &&
		len(group.NodeTemplate.Module.ScaleOutModuleID) != 0 {
		bkBizID, _ := strconv.Atoi(bkBizIDString)
		bkModuleID, _ := strconv.Atoi(group.NodeTemplate.Module.ScaleOutModuleID)
		group.NodeTemplate.Module.ScaleOutModuleName = cloudprovider.GetModuleName(bkBizID, bkModuleID)
	}
}

func getClusterResourceGroup(cls *proto.Cluster) string {
	if cls.GetExtraInfo() == nil {
		return ""
	}

	rg, ok := cls.GetExtraInfo()[common.ClusterResourceGroup]
	if ok {
		return rg
	}

	return ""
}

func setCloudNodeGroupID(group *proto.NodeGroup) {
	if group == nil || len(group.NodeGroupID) == 0 {
		return
	}
	idx := strings.LastIndexByte(group.NodeGroupID, '-')
	group.CloudNodeGroupID = strings.ToLower(group.NodeGroupID[idx+1:])
}

func getCloudNodeGroupID(group *proto.NodeGroup) string {
	if group == nil || len(group.NodeGroupID) == 0 {
		return ""
	}
	idx := strings.LastIndexByte(group.NodeGroupID, '-')
	return strings.ToLower(group.NodeGroupID[idx+1:])
}

func updateVmss(rootCtx context.Context, cli api.AksService, group *proto.NodeGroup,
	set *armcompute.VirtualMachineScaleSet, rg, nrg string) (*armcompute.VirtualMachineScaleSet, error) {
	taskID := cloudprovider.GetTaskIDFromContext(rootCtx)

	group.AutoScaling.AutoScalingName = api.RegexpSetNodeGroupResourcesName(set)
	// 设置虚拟规模集网络
	err := api.SetVmSetNetWork(rootCtx, cli, group, rg, nrg, set)
	if err != nil {
		return nil, fmt.Errorf("updateVmss[%s]: call SetVmSetNetWork[%s] falied, %v", taskID,
			group.CloudNodeGroupID, err)
	}
	// 设置虚拟规模密码
	api.SetVmSetPasswd(group, set)
	// 设置虚拟规模SSH免密登录公钥
	api.SetVmSetSSHPublicKey(group, set)
	// 买数据盘
	api.BuyDataDisk(group, set)
	// 用户数据
	api.SetUserData(group, set)
	api.SetImageReferenceNull(set)

	ctx, cancel := context.WithTimeout(rootCtx, 5*time.Minute)
	defer cancel()
	vmSet, err := cli.UpdateSetWithName(ctx, set, rg, *set.Name)
	if err != nil {
		blog.Errorf("updateVmss[%s]: call UpdateSetWithName[%s][%s] failed, %v",
			taskID, rg, *set.Name, err)
		return nil, fmt.Errorf("updateVmss call UpdateSetWithName[%s][%s] failed, %v", rg, *set.Name, err)
	}
	blog.Infof("updateVmss[%s]: UpdateSetWithName[%s] set successful", taskID, *vmSet.Name)

	return vmSet, nil
}
