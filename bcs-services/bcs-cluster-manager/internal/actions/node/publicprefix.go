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

package node

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// ListCloudNodePublicPrefixAction action for update node public prefix
type ListCloudNodePublicPrefixAction struct {
	ctx   context.Context
	model store.ClusterManagerModel

	cloud   *cmproto.Cloud
	account *cmproto.CloudAccount

	req  *cmproto.ListCloudNodePublicPrefixRequest
	resp *cmproto.ListCloudNodePublicPrefixResponse

	publicPrefixs []*cmproto.NodePublicPrefix
}

// NewListCloudNodePublicPrefixAction create update action
func NewListCloudNodePublicPrefixAction(model store.ClusterManagerModel) *ListCloudNodePublicPrefixAction {
	return &ListCloudNodePublicPrefixAction{
		model: model,
	}
}

func (ua *ListCloudNodePublicPrefixAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}

	err := ua.getRelativeData()
	if err != nil {
		return err
	}

	if len(ua.req.AccountID) > 0 {
		validate, errGet := cloudprovider.GetCloudValidateMgr(ua.cloud.CloudProvider)
		if errGet != nil {
			return errGet
		}
		err := validate.ImportCloudAccountValidate(ua.account.Account)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ua *ListCloudNodePublicPrefixAction) listPubilcPrefix() error { // nolint
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     ua.cloud,
		AccountID: ua.req.AccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s list CloudBwpsResource failed, %s",
			ua.cloud.CloudID, ua.cloud.CloudProvider, err.Error())
		return err
	}

	// create vpc client with cloudProvider
	clsMgr, err := cloudprovider.GetNodeMgr(ua.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s VPCManager for list CloudBwpsResource failed, %s",
			ua.cloud.CloudProvider, err.Error())
		return err
	}

	ua.publicPrefixs, err = clsMgr.ListNodePublicPrefixs(&cloudprovider.ListNodePublicPrefixesOption{
		CommonOption:      *cmOption,
		ResourceGroupName: ua.req.ResourceGroupName,
	})
	if err != nil {
		return err
	}

	return nil
}

func (ua *ListCloudNodePublicPrefixAction) getRelativeData() error {
	cloud, err := actions.GetCloudByCloudID(ua.model, ua.req.CloudID)
	if err != nil {
		return err
	}

	if len(ua.req.AccountID) > 0 {
		account, errGet := ua.model.GetCloudAccount(ua.ctx, ua.req.CloudID, ua.req.AccountID, false)
		if errGet != nil {
			return errGet
		}
		ua.account = account
	}

	ua.cloud = cloud
	return nil
}

func (ua *ListCloudNodePublicPrefixAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	ua.resp.Data = ua.publicPrefixs
}

// Handle handles update node taints
func (ua *ListCloudNodePublicPrefixAction) Handle(ctx context.Context, req *cmproto.ListCloudNodePublicPrefixRequest,
	resp *cmproto.ListCloudNodePublicPrefixResponse) {
	if req == nil || resp == nil {
		blog.Errorf("update node taints failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ua.listPubilcPrefix(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
