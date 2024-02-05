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

package cloudresource

import (
	"context"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// ListCloudInstancesAction list action for cloud instance
type ListCloudInstancesAction struct {
	ctx     context.Context
	cloud   *cmproto.Cloud
	account *cmproto.CloudAccount
	model   store.ClusterManagerModel

	req  *cmproto.ListCloudInstancesRequest
	resp *cmproto.ListCloudInstancesResponse

	ipList []string
	nodes  []*cmproto.CloudNode
}

// NewListCloudInstancesAction create list action for node type
func NewListCloudInstancesAction(model store.ClusterManagerModel) *ListCloudInstancesAction {
	return &ListCloudInstancesAction{
		model: model,
	}
}

func (la *ListCloudInstancesAction) validate() error {
	if err := la.req.Validate(); err != nil {
		return err
	}
	err := la.getRelativeData()
	if err != nil {
		return err
	}

	la.ipList = strings.Split(la.req.IpList, ",")

	validate, err := cloudprovider.GetCloudValidateMgr(la.cloud.CloudProvider)
	if err != nil {
		return err
	}

	err = validate.ListInstancesValidate(la.req, func() *cmproto.Account {
		if la.account == nil || la.account.Account == nil {
			return nil
		}
		return la.account.Account
	}())
	if err != nil {
		return err
	}

	return nil
}

func (la *ListCloudInstancesAction) getRelativeData() error {
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

func (la *ListCloudInstancesAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.nodes
}

func (la *ListCloudInstancesAction) listCloudInstancesByIPs() error {
	// create vpc client with cloudProvider
	nodeMgr, err := cloudprovider.GetNodeMgr(la.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s NodeManager for list instances failed, %s",
			la.cloud.CloudProvider, err.Error())
		return err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     la.cloud,
		AccountID: la.req.AccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s list instances failed, %s",
			la.cloud.CloudID, la.cloud.CloudProvider, err.Error())
		return err
	}
	cmOption.Region = la.req.Region

	// get region zones
	zoneMap := make(map[string]*cmproto.ZoneInfo, 0)
	zones, err := nodeMgr.GetZoneList(&cloudprovider.GetZoneListOption{
		CommonOption: *cmOption,
	})
	if err != nil {
		blog.Errorf("cloudprovider %s/%s get zones failed, %s",
			la.cloud.CloudID, la.cloud.CloudProvider, err.Error())
		return err
	}
	for i := range zones {
		zoneMap[zones[i].ZoneID] = zones[i]
	}

	// get instance nodes list
	instanceNodes, err := nodeMgr.ListNodesByIP(la.ipList, &cloudprovider.ListNodesOption{
		Common: cmOption,
	})
	if err != nil {
		blog.Errorf("ListCloudInstancesAction ListNodesByIP failed: %v", err)
		return err
	}
	instanceNodeMap := make(map[string]*cmproto.Node, 0)
	for i := range instanceNodes {
		instanceNodeMap[instanceNodes[i].InnerIP] = instanceNodes[i]
	}

	for _, ip := range la.ipList {
		n, ok := instanceNodeMap[ip]
		if ok {
			zoneID := fmt.Sprintf("%v", n.Zone)

			la.nodes = append(la.nodes, &cmproto.CloudNode{
				NodeID:       n.NodeID,
				InnerIP:      n.InnerIP,
				InstanceType: n.InstanceType,
				Cpu:          n.CPU,
				Mem:          n.Mem,
				Gpu:          n.GPU,
				Vpc:          n.VPC,
				Region:       n.Region,
				InnerIPv6:    n.InnerIPv6,
				ZoneID:       zoneID,
				Zone:         n.ZoneID,
				ZoneName: func() string {
					zone, ok := zoneMap[zoneID]
					if ok && zone != nil {
						return zone.ZoneName
					}
					return ""
				}(),
				CloudRegionNode: true,
			})

			continue
		}

		la.nodes = append(la.nodes, &cmproto.CloudNode{
			InnerIP:         ip,
			CloudRegionNode: false,
		})
	}

	return nil
}

// Handle list node type request
func (la *ListCloudInstancesAction) Handle(ctx context.Context,
	req *cmproto.ListCloudInstancesRequest, resp *cmproto.ListCloudInstancesResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list node instances failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := la.listCloudInstancesByIPs(); err != nil {
		la.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// ListKeyPairsAction action for list keypairs
type ListKeyPairsAction struct {
	ctx      context.Context
	cloud    *cmproto.Cloud
	account  *cmproto.CloudAccount
	model    store.ClusterManagerModel
	req      *cmproto.ListKeyPairsRequest
	resp     *cmproto.ListKeyPairsResponse
	keyPairs []*cmproto.KeyPair
}

// NewListKeyPairsAction create list action for key pairs
func NewListKeyPairsAction(model store.ClusterManagerModel) *ListKeyPairsAction {
	return &ListKeyPairsAction{
		model: model,
	}
}

func (la *ListKeyPairsAction) validate() error {
	if err := la.req.Validate(); err != nil {
		return err
	}

	// get cloud/account info
	err := la.getRelativeData()
	if err != nil {
		return err
	}

	validate, err := cloudprovider.GetCloudValidateMgr(la.cloud.CloudProvider)
	if err != nil {
		return err
	}

	err = validate.ListKeyPairsValidate(la.req, func() *cmproto.Account {
		if la.account == nil || la.account.Account == nil {
			return nil
		}

		return la.account.Account
	}())
	if err != nil {
		return err
	}

	return nil
}

func (la *ListKeyPairsAction) getRelativeData() error {
	cloud, err := actions.GetCloudByCloudID(la.model, la.req.CloudID)
	if err != nil {
		return err
	}
	la.cloud = cloud

	if len(la.req.AccountID) > 0 {
		account, errGet := la.model.GetCloudAccount(la.ctx, la.req.CloudID, la.req.AccountID, false)
		if errGet != nil {
			return errGet
		}

		la.account = account
	}

	return nil
}

func (la *ListKeyPairsAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.keyPairs
}

func (la *ListKeyPairsAction) listKeyPairs() error {
	// create node client with cloudProvider
	nodeMgr, err := cloudprovider.GetNodeMgr(la.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s NodeManager for list keyPairs failed, %s",
			la.cloud.CloudProvider, err.Error())
		return err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     la.cloud,
		AccountID: la.req.AccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s list SecurityGroups failed, %s",
			la.cloud.CloudID, la.cloud.CloudProvider, err.Error())
		return err
	}
	cmOption.Region = la.req.Region

	// get key list
	keys, err := nodeMgr.ListKeyPairs(&cloudprovider.ListNetworksOption{
		CommonOption:      *cmOption,
		ResourceGroupName: la.req.ResourceGroupName,
	})
	if err != nil {
		return err
	}
	la.keyPairs = keys

	return nil
}

// Handle list key pairs
func (la *ListKeyPairsAction) Handle(
	ctx context.Context, req *cmproto.ListKeyPairsRequest, resp *cmproto.ListKeyPairsResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list key pairs failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := la.listKeyPairs(); err != nil {
		la.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
