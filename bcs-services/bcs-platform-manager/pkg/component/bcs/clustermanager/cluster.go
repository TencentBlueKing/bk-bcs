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

// Package clustermanager xxx
package clustermanager

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/constants"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/types"
)

// ListCluster list cluster from cluster manager
func ListCluster(ctx context.Context,
	req *clustermanager.ListClusterReq) ([]*types.Cluster, error) {
	cli, close, err := clustermanager.GetClient(constants.ServiceDomain)
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return nil, err
	}
	p, err := cli.ListCluster(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ListCluster error: %s", err)
	}
	if p.Code != 0 {
		return nil, fmt.Errorf("ListCluster error, code: %d, message: %s", p.Code, p.GetMessage())
	}
	var clusterList []*types.Cluster
	for _, cls := range p.Data {
		clusterList = append(clusterList, &types.Cluster{
			ProjectID:       cls.ProjectID,
			ClusterID:       cls.ClusterID,
			ClusterName:     cls.ClusterName,
			BKBizID:         cls.BusinessID,
			Status:          cls.Status,
			IsShared:        cls.IsShared,
			ClusterType:     cls.ClusterType,
			Provider:        cls.Provider,
			Region:          cls.Region,
			VpcID:           cls.VpcID,
			NetworkSettings: convertNetworkSettings(cls),
		})
	}
	return clusterList, nil
}

func convertNetworkSettings(cls *clustermanager.Cluster) *types.NetworkSettings {
	if cls.NetworkSettings == nil {
		return nil
	}
	return &types.NetworkSettings{
		EniSubnetIDs:  cls.NetworkSettings.EniSubnetIDs,
		MaxNodePodNum: int(cls.NetworkSettings.MaxNodePodNum),
		MaxServiceNum: int(cls.NetworkSettings.MaxServiceNum),
		EnableVPCCni:  cls.NetworkSettings.EnableVPCCni,
		SubnetSource:  convertSubnetSource(cls.NetworkSettings.SubnetSource),
	}
}

func convertSubnetSource(cls *clustermanager.SubnetSource) *types.SubnetSource {
	if cls == nil {
		return nil
	}
	return &types.SubnetSource{
		New:     convertNewSubnet(cls.New),
		Existed: convertExistedSubnetIDs(cls.Existed),
	}
}

func convertNewSubnet(cls []*clustermanager.NewSubnet) []*types.NewSubnet {
	if len(cls) == 0 {
		return nil
	}
	var newSubnets []*types.NewSubnet
	for _, subnet := range cls {
		newSubnets = append(newSubnets, &types.NewSubnet{
			Mask:  subnet.Mask,
			Zone:  subnet.Zone,
			IpCnt: subnet.IpCnt,
		})
	}
	return newSubnets
}

func convertExistedSubnetIDs(cls *clustermanager.ExistedSubnetIDs) *types.ExistedSubnetIDs {
	if cls == nil {
		return nil
	}
	return &types.ExistedSubnetIDs{
		IDs: cls.Ids,
	}
}

// GetCluster 获取集群详情
func GetCluster(ctx context.Context, clusterID string) (*types.Cluster, error) {
	cli, close, errr := clustermanager.GetClient(constants.ServiceDomain)
	defer func() {
		if close != nil {
			close()
		}
	}()
	if errr != nil {
		return nil, errr
	}
	p, errr := cli.GetCluster(ctx, &clustermanager.GetClusterReq{
		ClusterID: clusterID,
	})
	if errr != nil {
		return nil, fmt.Errorf("GetCluster error: %s", errr)
	}
	if p.Code != 0 {
		return nil, fmt.Errorf("GetCluster error, code: %d, message: %s", p.Code, p.GetMessage())
	}
	cls := &types.Cluster{
		ProjectID:   p.Data.ProjectID,
		ClusterID:   p.Data.ClusterID,
		ClusterName: p.Data.ClusterName,
		BKBizID:     p.Data.BusinessID,
		Status:      p.Data.Status,
		IsShared:    p.Data.IsShared,
		ClusterType: p.Data.ClusterType,
	}
	return cls, nil
}

// AddSubnetToCluster add cloud subnets cluster to cluster manager
func AddSubnetToCluster(ctx context.Context,
	req *clustermanager.AddSubnetToClusterReq) (*clustermanager.AddSubnetToClusterResp, error) {
	cli, close, err := clustermanager.GetClient(constants.ServiceDomain)
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return nil, err
	}
	p, err := cli.AddSubnetToCluster(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("AddSubnetToCluster error: %s", err)
	}
	if p.Code != 0 {
		return nil, fmt.Errorf("AddSubnetToCluster error, code: %d, message: %s", p.Code, p.GetMessage())
	}
	return p, nil
}
