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

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// ListCloudRuntimeInfoAction list action for runtime info
type ListCloudRuntimeInfoAction struct {
	ctx   context.Context
	cloud *cmproto.Cloud
	model store.ClusterManagerModel
	req   *cmproto.ListCloudRuntimeInfoRequest
	resp  *cmproto.ListCloudRuntimeInfoResponse

	runtimeInfo map[string]*cmproto.RunTimeVersion
}

// NewListCloudRuntimeInfoAction create list action for runtime info
func NewListCloudRuntimeInfoAction(model store.ClusterManagerModel) *ListCloudRuntimeInfoAction {
	return &ListCloudRuntimeInfoAction{
		model: model,
	}
}

func (la *ListCloudRuntimeInfoAction) validate() error {
	if err := la.req.Validate(); err != nil {
		return err
	}
	err := la.getRelativeData()
	if err != nil {
		return err
	}

	return nil
}

func (la *ListCloudRuntimeInfoAction) getRelativeData() error {
	// get relative cluster for information injection
	cloud, err := actions.GetCloudByCloudID(la.model, la.req.CloudID)
	if err != nil {
		return err
	}
	la.cloud = cloud

	return nil
}

func (la *ListCloudRuntimeInfoAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.runtimeInfo
}

func (la *ListCloudRuntimeInfoAction) listCloudRuntimeInfo() error {
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: la.req.ClusterID,
		CloudID:   la.req.CloudID,
	})
	if err != nil {
		blog.Errorf("get dependBasicInfo %s clusterManager for list runtime info failed, %s",
			la.cloud.CloudProvider, err.Error())
		return err
	}

	clsMgr, err := cloudprovider.GetNodeMgr(la.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s clusterManager for list runtime info failed, %s",
			la.cloud.CloudProvider, err.Error())
		return err
	}

	runtimeinfo, err := clsMgr.ListRuntimeInfo(&cloudprovider.ListRuntimeInfoOption{
		CommonOption: *dependInfo.CmOption,
		Cluster:      dependInfo.Cluster,
	})
	if err != nil {
		return err
	}

	tmp := make(map[string]*cmproto.RunTimeVersion, 0)
	for k, v := range runtimeinfo {
		versions := v
		tmp[k] = &cmproto.RunTimeVersion{
			Version: versions,
		}
	}

	la.runtimeInfo = tmp

	return nil
}

// Handle handle list runtime info request
func (la *ListCloudRuntimeInfoAction) Handle(ctx context.Context,
	req *cmproto.ListCloudRuntimeInfoRequest, resp *cmproto.ListCloudRuntimeInfoResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list runtime info failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := la.listCloudRuntimeInfo(); err != nil {
		la.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
