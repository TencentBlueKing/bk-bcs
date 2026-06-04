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

package cloudvpc

import (
	"context"
	"errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cidrtree"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

// ListCloudVpcClusterAction action for list cloud vpc cluster
type ListCloudVpcClusterAction struct {
	ctx context.Context

	cloud   *cmproto.Cloud
	account *cmproto.CloudAccount

	model store.ClusterManagerModel
	req   *cmproto.ListCloudVpcClusterRequest
	resp  *cmproto.ListCloudVpcClusterResponse

	vpcs []*cmproto.CloudCluster
}

// NewListCloudVpcClusterAction create list action for cloud vpc cluster
func NewListCloudVpcClusterAction(model store.ClusterManagerModel) *ListCloudVpcClusterAction {
	return &ListCloudVpcClusterAction{
		model: model,
	}
}

func (la *ListCloudVpcClusterAction) validate() error {
	if err := la.req.Validate(); err != nil {
		return err
	}

	// get cloud/account info
	err := la.getRelativeData()
	if err != nil {
		return err
	}

	return nil
}

func (la *ListCloudVpcClusterAction) getRelativeData() error {
	cloud, err := actions.GetCloudByCloudID(la.model, la.req.CloudID)
	if err != nil {
		return err
	}
	la.cloud = cloud

	if la.req.AccountID != "" {
		account, err := la.model.GetCloudAccount(la.ctx, la.req.CloudID, la.req.AccountID, false)
		if err != nil {
			return err
		}

		la.account = account
	}

	return nil
}

func (la *ListCloudVpcClusterAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.vpcs
}

// ListCloudVpcCluster list cloud vpc cluster
func (la *ListCloudVpcClusterAction) ListCloudVpcCluster() error {
	// 集群条件筛选
	total, clusterList, err := la.filterClusterList()
	if err != nil {
		return err
	}
	la.vpcs = make([]*cmproto.CloudCluster, 0)
	for _, cluster := range clusterList {
		if cluster.NetworkSettings == nil {
			continue
		}
		// 获取overlay ip cidr
		ipNum, err := cidrtree.GetIPNumByCidr(cluster.NetworkSettings.ClusterIPv4CIDR)
		if err != nil {
			return err
		}
		overlayIPCidrs := []*cmproto.OverlayIPCidr{}
		overlayIPCidrs = append(overlayIPCidrs, &cmproto.OverlayIPCidr{
			Cidr:  cluster.NetworkSettings.ClusterIPv4CIDR,
			IpNum: ipNum,
		})
		for _, ipCidr := range cluster.NetworkSettings.MultiClusterCIDR {
			ipNum, err := cidrtree.GetIPNumByCidr(ipCidr)
			if err != nil {
				return err
			}
			overlayIPCidrs = append(overlayIPCidrs, &cmproto.OverlayIPCidr{
				Cidr:  ipCidr,
				IpNum: ipNum,
			})
		}
		la.vpcs = append(la.vpcs, &cmproto.CloudCluster{
			ClusterID:     cluster.ClusterID,
			OverlayIPCidr: overlayIPCidrs,
		})
	}
	la.resp.Total = uint32(total)

	return nil
}

// filterClusterList filter cluster list by cloud provider, region, vpcid
func (la *ListCloudVpcClusterAction) filterClusterList() (int64, []*cmproto.Cluster, error) {
	var (
		total       int64
		clusterList []*cmproto.Cluster
		err         error
	)

	condM := make(operator.M)
	if la.cloud != nil && la.cloud.CloudProvider != "" && len(la.cloud.CloudProvider) != 0 {
		condM["provider"] = la.cloud.CloudProvider
	}
	if len(la.req.Region) != 0 {
		condM["region"] = la.req.Region
	}
	if len(la.req.VpcID) != 0 {
		condM["vpcid"] = la.req.VpcID
	}
	condM["networktype"] = "overlay"
	branchCond := operator.NewLeafCondition(operator.Eq, condM)

	total, clusterList, err = la.model.ListClusterByPage(la.ctx, branchCond, &storeopt.ListOption{
		Offset: int64(la.req.Offset),
		Limit:  int64(la.req.Limit),
	})
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		return 0, clusterList, err
	}

	return total, clusterList, nil
}

// Handle list cloud vpcs
func (la *ListCloudVpcClusterAction) Handle(
	ctx context.Context, req *cmproto.ListCloudVpcClusterRequest, resp *cmproto.ListCloudVpcClusterResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list cloud vpc cluster failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := la.ListCloudVpcCluster(); err != nil {
		la.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
