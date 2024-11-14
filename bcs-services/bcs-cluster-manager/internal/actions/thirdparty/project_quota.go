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

// Package thirdparty xxx
package thirdparty

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// GetProjectResourceQuotaUsageAction action for get project resource quota
type GetProjectResourceQuotaUsageAction struct {
	ctx   context.Context
	model store.ClusterManagerModel

	req  *cmproto.GetProjectResourceQuotaUsageRequest
	resp *cmproto.GetProjectResourceQuotaUsageResponse

	projectId string
	cloud     *cmproto.Cloud

	groups []cmproto.NodeGroup

	regionInsTypes map[string][]string
}

// NewGetProjectResourceQuotaUsageAction create action
func NewGetProjectResourceQuotaUsageAction(model store.ClusterManagerModel) *GetProjectResourceQuotaUsageAction {
	return &GetProjectResourceQuotaUsageAction{
		model: model,
	}
}

// validate for check project or cloud
func (ga *GetProjectResourceQuotaUsageAction) validate() error {
	if err := ga.req.Validate(); err != nil {
		return err
	}

	// check projectId or Code
	proInfo, errLocal := project.GetProjectManagerClient().GetProjectInfo(ga.req.GetProjectID(), false)
	if errLocal != nil {
		return errLocal
	}
	ga.projectId = proInfo.GetProjectID()

	// check cloud
	cloud, err := ga.model.GetCloud(ga.ctx, ga.req.GetProviderID())
	if err != nil {
		return fmt.Errorf("not support provider[%s]", ga.req.GetProviderID())
	}
	ga.cloud = cloud

	return nil
}

func (ga *GetProjectResourceQuotaUsageAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ga *GetProjectResourceQuotaUsageAction) getProjectNodeGroups() error {
	condGroup := operator.NewLeafCondition(operator.Eq, operator.M{
		"status":    common.StatusRunning,
		"projectid": ga.projectId,
		"provider":  ga.cloud.GetCloudID(),
	})
	groupList, err := ga.model.ListNodeGroup(context.Background(), condGroup, &options.ListOption{All: true})
	if err != nil {
		return err
	}

	ga.groups = groupList

	return nil
}

func (ga *GetProjectResourceQuotaUsageAction) getProjectGroupsQuota() error {
	mgr, err := cloudprovider.GetNodeGroupMgr(ga.cloud.GetCloudProvider())
	if err != nil {
		blog.Errorf("get NodeGroup Manager cloudprovider %s/%s for getProjectGroupsQuota failed, %s",
			ga.cloud.CloudID, ga.cloud.CloudProvider, err.Error())
		return err
	}

	groupQuotas, err := mgr.GetProjectCaResourceQuota(ga.groups, nil)
	if err != nil {
		return err
	}

	result, err := utils.MarshalInterfaceToListValue(groupQuotas)
	if err != nil {
		blog.Errorf("marshal projectGroupsQuotaData err, %s", err.Error())
		return err
	}

	ga.resp.Data = result
	return nil
}

// Handle handles resource quota usage
func (ga *GetProjectResourceQuotaUsageAction) Handle(ctx context.Context,
	req *cmproto.GetProjectResourceQuotaUsageRequest, resp *cmproto.GetProjectResourceQuotaUsageResponse) {
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	if err := ga.validate(); err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ga.getProjectNodeGroups(); err != nil {
		ga.setResp(common.BcsErrClusterManagerCommonErr, err.Error())
		return
	}

	if err := ga.getProjectGroupsQuota(); err != nil {
		ga.setResp(common.BcsErrClusterManagerCommonErr, err.Error())
		return
	}

	blog.Infof("GetProjectResourceQuotaUsageAction get project groups resource quota successfully")
	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
