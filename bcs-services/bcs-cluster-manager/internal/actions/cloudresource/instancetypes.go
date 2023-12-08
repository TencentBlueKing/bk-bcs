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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/i18n"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// ListNodeTypeAction list action for node type
type ListNodeTypeAction struct {
	ctx          context.Context
	cloud        *cmproto.Cloud
	account      *cmproto.CloudAccount
	model        store.ClusterManagerModel
	req          *cmproto.ListCloudInstanceTypeRequest
	resp         *cmproto.ListCloudInstanceTypeResponse
	nodeTypeList []*cmproto.InstanceType
}

// NewListNodeTypeAction create list action for node type
func NewListNodeTypeAction(model store.ClusterManagerModel) *ListNodeTypeAction {
	return &ListNodeTypeAction{
		model: model,
	}
}

func (la *ListNodeTypeAction) validate() error {
	if err := la.req.Validate(); err != nil {
		return err
	}
	err := la.getRelativeData()
	if err != nil {
		return err
	}

	validate, err := cloudprovider.GetCloudValidateMgr(la.cloud.CloudProvider)
	if err != nil {
		return err
	}

	err = validate.ListInstanceTypeValidate(la.req, func() *cmproto.Account {
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

func (la *ListNodeTypeAction) getRelativeData() error {
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

func (la *ListNodeTypeAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.nodeTypeList
}

func (la *ListNodeTypeAction) listCloudInstancetypes() error {
	// create vpc client with cloudProvider
	nodeMgr, err := cloudprovider.GetNodeMgr(la.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s VPCManager for list subnets failed, %s", la.cloud.CloudProvider, err.Error())
		return err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     la.cloud,
		AccountID: la.req.AccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s list subnets failed, %s",
			la.cloud.CloudID, la.cloud.CloudProvider, err.Error())
		return err
	}
	cmOption.Region = la.req.Region

	// get instance types list
	insTypes, err := nodeMgr.ListNodeInstanceType(cloudprovider.InstanceInfo{
		Region:       la.req.Region,
		Zone:         la.req.Zone,
		NodeFamily:   la.req.NodeFamily,
		Cpu:          la.req.Cpu,
		Memory:       la.req.Memory,
		BizID:        la.req.BizID,
		Provider:     la.req.Provider,
		ResourceType: la.req.ResourceType,
	}, cmOption)
	if err != nil {
		return err
	}
	if len(insTypes) != 0 {
		for i := range insTypes {
			insTypes[i].TypeName = translate(la.ctx, insTypes[i].NodeFamily, insTypes[i].TypeName)
		}
	}
	la.nodeTypeList = insTypes

	return nil
}

// Handle handle list node type request
func (la *ListNodeTypeAction) Handle(ctx context.Context,
	req *cmproto.ListCloudInstanceTypeRequest, resp *cmproto.ListCloudInstanceTypeResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list node type failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := la.listCloudInstancetypes(); err != nil {
		la.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

func translate(ctx context.Context, nodeFamily, typeName string) string {
	switch nodeFamily {
	case "S1", "S2", "S3", "S4", "S5", "S6", "SA1", "SA2", "SA3", "SR1", "BMSA2", "BMS5", "BMS4":
		return i18n.Tf(ctx, "{{.Standard}}", nodeFamily)
	case "S5se":
		return i18n.Tf(ctx, "{{.StandardStorage}}", nodeFamily)
	case "S2ne", "SN3ne":
		return i18n.Tf(ctx, "{{.StandardNetwork}}", nodeFamily)
	case "M1", "M2", "M3", "M4", "M5", "M6", "MA2", "MA3", "M6mp", "M6p":
		return i18n.Tf(ctx, "{{.MemoryOptimized}}", nodeFamily)
	case "M6ce":
		return i18n.Tf(ctx, "{{.SEMemoryOptimized}}", nodeFamily)
	case "I3", "IT3", "IT5", "BMI5", "BMIA2":
		return i18n.Tf(ctx, "{{.HighIO}}", nodeFamily)
	case "D2", "D3", "D1", "BMDA2":
		return i18n.Tf(ctx, "{{.BigData}}", nodeFamily)
	case "BMGNV4", "BMG5t", "HCCPNV4h", "HCCG5v", "BMG5v":
		return i18n.Tf(ctx, "{{.GType}}", nodeFamily)
	case "C2", "C3", "C4", "C5", "C6":
		return i18n.Tf(ctx, "{{.ComputeOptimized}}", nodeFamily)
	case "CN3":
		return i18n.Tf(ctx, "{{.ComputeNetwork}}", nodeFamily)
	case "GN6S", "GN7", "GN8", "GN10X", "GT4", "PNV4", "GN10Xp":
		return i18n.Tf(ctx, "{{.GComputeOptimized}}", nodeFamily)
	case "GN7vi":
		return i18n.Tf(ctx, "{{.GVideoEnhanced}}", nodeFamily)
	case "GI3X":
		return i18n.Tf(ctx, "{{.GReasoning}}", nodeFamily)
	case "GNV4":
		return i18n.Tf(ctx, "{{.GRendering}}", nodeFamily)
	default:
		return typeName
	}
}
