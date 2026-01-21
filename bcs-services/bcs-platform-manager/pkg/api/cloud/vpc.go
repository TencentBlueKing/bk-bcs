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

// Package cloud cloud operate
package cloud

import (
	"context"

	cluproto "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/bcs/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/types"
)

// ListCloudVpcsPage 获取云VPC分页列表
// @Summary 获取云VPC分页列表
// @Tags    Cloud
// @Produce json
// @Success 200 {array} types.ListCloudVpcsPageResponse
// @Router  /clouds/{cloudID}/vpcs/page [get]
func ListCloudVpcsPage(
	c context.Context, req *types.ListCloudVpcsPageRequest) (*types.ListCloudVpcsPageResponse, error) {
	sr, err := clustermanager.ListCloudVpcsPage(c, &cluproto.ListCloudVpcsPageRequest{
		CloudID:           req.CloudID,
		Region:            req.Region,
		AccountID:         req.AccountID,
		VpcID:             req.VpcID,
		ResourceGroupName: req.ResourceGroupName,
		VpcName:           req.VpcName,
		Offset:            req.Offset,
		Limit:             req.Limit,
	})
	if err != nil {
		return nil, err
	}
	cloudVpcs := make([]types.CloudVpcs, 0)
	for _, vpc := range sr.Data {
		cloudVpcs = append(cloudVpcs, types.CloudVpcs{
			VpcName:                vpc.VpcName,
			VpcID:                  vpc.VpcID,
			Region:                 vpc.Region,
			OverlayCidr:            vpc.OverlayCidr,
			AvailableOverlayIpNum:  vpc.AvailableOverlayIpNum,
			AvailableOverlayCidr:   vpc.AvailableOverlayCidr,
			TotalOverlayIpNum:      vpc.TotalOverlayIpNum,
			UnderlayCidr:           vpc.UnderlayCidr,
			AvailableUnderlayIpNum: vpc.AvailableUnderlayIpNum,
			AvailableUnderlayCidr:  vpc.AvailableUnderlayCidr,
			TotalUnderlayIpNum:     vpc.TotalUnderlayIpNum,
			OverlayIpUsageRate:     calculateUsageRate(vpc.AvailableOverlayIpNum, vpc.TotalOverlayIpNum),
			UnderlayIpUsageRate:    calculateUsageRate(vpc.AvailableUnderlayIpNum, vpc.TotalUnderlayIpNum),
			CreateTime:             vpc.CreateTime,
			OverlayIPCidr:          convertOverlayIPCidr(vpc.OverlayIPCidr),
		})
	}
	return &types.ListCloudVpcsPageResponse{
		Total:     sr.Total,
		CloudVpcs: cloudVpcs,
	}, nil
}

// calculateUsageRate calculate usage rate
func calculateUsageRate(available, total uint32) float64 {
	if total == 0 {
		return 0
	}
	return 1 - (float64(available) / float64(total))
}

// convertOverlayIPCidr convert overlay ip cidr from proto to types
func convertOverlayIPCidr(proto []*cluproto.OverlayIPCidr) []types.OverlayIPCidr {
	overlayIPCidr := make([]types.OverlayIPCidr, 0)
	for _, ipCidr := range proto {
		overlayIPCidr = append(overlayIPCidr, types.OverlayIPCidr{
			Cidr:  ipCidr.Cidr,
			IpNum: ipCidr.IpNum,
		})
	}
	return overlayIPCidr
}

// ListCloudVpcCluster 获取云VPC关联的集群列表
// @Summary 获取云VPC关联的集群列表
// @Tags    Cloud
// @Produce json
// @Success 200 {array} types.ListCloudVpcClusterResponse
// @Router  /clouds/{cloudID}/vpc/{vpcID}/cluster [get]
func ListCloudVpcCluster(
	c context.Context, req *types.ListCloudVpcClusterRequest) (*types.ListCloudVpcClusterResponse, error) {
	sr, err := clustermanager.ListCloudVpcCluster(c, &cluproto.ListCloudVpcClusterRequest{
		CloudID:   req.CloudID,
		Region:    req.Region,
		AccountID: req.AccountID,
		VpcID:     req.VpcID,
		Offset:    req.Offset,
		Limit:     req.Limit,
	})
	if err != nil {
		return nil, err
	}
	cloudCluster := make([]types.CloudCluster, 0)
	for _, vpc := range sr.Data {
		cloudCluster = append(cloudCluster, types.CloudCluster{
			ClusterID:     vpc.ClusterID,
			OverlayIPCidr: convertOverlayIPCidr(vpc.OverlayIPCidr),
		})
	}
	return &types.ListCloudVpcClusterResponse{
		Total:        sr.Total,
		CloudCluster: cloudCluster,
	}, nil
}

// UpdateCloudVpcs 更新云VPC
// @Summary 更新云VPC
// @Tags    Cloud
// @Produce json
// @Success 200 {array} types.UpdateCloudVpcsResponse
// @Router  /clouds/{cloudID}/vpc/{vpcID} [put]
func UpdateCloudVpcs(
	c context.Context, req *types.UpdateCloudVpcsRequest) (*types.UpdateCloudVpcsResponse, error) {
	_, err := clustermanager.UpdateCloudVpcs(c, &cluproto.UpdateCloudVpcsRequest{
		CloudID:           req.CloudID,
		Region:            req.Region,
		AccountID:         req.AccountID,
		VpcID:             req.VpcID,
		ResourceGroupName: req.ResourceGroupName,
		VpcName:           req.VpcName,
	})
	if err != nil {
		return nil, err
	}
	return &types.UpdateCloudVpcsResponse{}, nil
}

// ListCloudSubnets 获取云子网列表
// @Summary 获取云子网列表
// @Tags    Cloud
// @Produce json
// @Success 200 {array} types.ListCloudSubnetsResponse
// @Router  /clouds/{cloudID}/vpc/{vpcID}/subnets [get]
func ListCloudSubnets(
	c context.Context, req *types.ListCloudSubnetsRequest) (*types.ListCloudSubnetsResponse, error) {
	cloudSubnets, err := clustermanager.ListCloudSubnets(c, &cluproto.ListCloudSubnetsRequest{
		CloudID:   req.CloudID,
		Region:    req.Region,
		AccountID: req.AccountID,
		VpcID:     req.VpcID,
	})
	if err != nil {
		return nil, err
	}

	CloudSubnets := make([]types.ListCloudSubnets, 0)
	for _, vpc := range cloudSubnets.Data {
		CloudSubnets = append(CloudSubnets, types.ListCloudSubnets{
			SubnetName:              vpc.SubnetName,
			SubnetID:                vpc.SubnetID,
			VpcID:                   vpc.VpcID,
			CidrRange:               vpc.CidrRange,
			Ipv6CidrRange:           vpc.Ipv6CidrRange,
			Zone:                    vpc.Zone,
			AvailableIPAddressCount: vpc.AvailableIPAddressCount,
			ZoneName:                vpc.ZoneName,
			Cluster:                 convertCluster(vpc.Cluster),
			HwNeutronSubnetID:       vpc.HwNeutronSubnetID,
			TotalIpAddressCount:     vpc.TotalIpAddressCount,
		})
	}

	return &types.ListCloudSubnetsResponse{
		Total:   uint32(len(CloudSubnets)),
		Subnets: CloudSubnets,
	}, nil
}

// CreateCloudSubnets 创建云子网
// @Summary 创建云子网
// @Tags    Cloud
// @Produce json
// @Success 200 {array} types.CreateCloudSubnetsResponse
// @Router  /clouds/{cloudID}/vpc/{vpcID}/subnets [post]
func CreateCloudSubnets(
	c context.Context, req *types.CreateCloudSubnetsRequest) (*types.CreateCloudSubnetsResponse, error) {
	cloudSubnets, err := clustermanager.CreateCloudSubnets(c, &cluproto.CreateCloudSubnetsRequest{
		CloudID:    req.CloudID,
		Region:     req.Region,
		AccountID:  req.AccountID,
		VpcID:      req.VpcID,
		SubnetName: req.SubnetName,
		CidrBlock:  req.CidrBlock,
		Zone:       req.Zone,
	})
	if err != nil {
		return nil, err
	}

	return &types.CreateCloudSubnetsResponse{
		Subnet: types.CloudSubnets{
			SubnetName:              cloudSubnets.Data.SubnetName,
			SubnetID:                cloudSubnets.Data.SubnetID,
			VpcID:                   cloudSubnets.Data.VpcID,
			CidrRange:               cloudSubnets.Data.CidrRange,
			Ipv6CidrRange:           cloudSubnets.Data.Ipv6CidrRange,
			Zone:                    cloudSubnets.Data.Zone,
			AvailableIPAddressCount: cloudSubnets.Data.AvailableIPAddressCount,
			ZoneName:                cloudSubnets.Data.ZoneName,
			Cluster:                 convertCluster(cloudSubnets.Data.Cluster),
			HwNeutronSubnetID:       cloudSubnets.Data.HwNeutronSubnetID,
			TotalIpAddressCount:     cloudSubnets.Data.TotalIpAddressCount,
		},
	}, nil
}

func convertCluster(proto *cluproto.ClusterInfo) types.ClusterInfo {
	if proto != nil {
		return types.ClusterInfo{
			ClusterName: proto.ClusterName,
			ClusterID:   proto.ClusterID,
		}
	}
	return types.ClusterInfo{
		ClusterName: "",
		ClusterID:   "",
	}
}

// UpdateCloudSubnets 更新云子网
// @Summary 更新云子网
// @Tags    Cloud
// @Produce json
// @Success 200 {array} types.UpdateCloudSubnetsResponse
// @Router  /clouds/{cloudID}/vpc/{vpcID}/subnets [put]
func UpdateCloudSubnets(
	c context.Context, req *types.UpdateCloudSubnetsRequest) (*types.UpdateCloudSubnetsResponse, error) {
	_, err := clustermanager.UpdateCloudSubnets(c, &cluproto.UpdateCloudSubnetsRequest{
		CloudID:    req.CloudID,
		Region:     req.Region,
		AccountID:  req.AccountID,
		SubnetName: req.SubnetName,
		SubnetID:   req.SubnetID,
	})
	if err != nil {
		return nil, err
	}

	return &types.UpdateCloudSubnetsResponse{}, nil
}

// DeleteCloudSubnets 删除云子网
// @Summary 删除云子网
// @Tags    Cloud
// @Produce json
// @Success 200 {array} types.DeleteCloudSubnetsResponse
// @Router  /clouds/{cloudID}/vpc/{vpcID}/subnets [delete]
func DeleteCloudSubnets(
	c context.Context, req *types.DeleteCloudSubnetsRequest) (*types.DeleteCloudSubnetsResponse, error) {
	_, err := clustermanager.DeleteCloudSubnets(c, &cluproto.DeleteCloudSubnetsRequest{
		CloudID:   req.CloudID,
		Region:    req.Region,
		AccountID: req.AccountID,
		SubnetID:  req.SubnetID,
	})
	if err != nil {
		return nil, err
	}

	return &types.DeleteCloudSubnetsResponse{}, nil
}
