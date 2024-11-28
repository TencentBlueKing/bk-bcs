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

package clustercredential

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// UpdateKubeconfigAction action for update kubeconfig of cluster
type UpdateKubeconfigAction struct {
	ctx     context.Context
	model   store.ClusterManagerModel
	cloud   *cmproto.Cloud
	cluster *cmproto.Cluster
	req     *cmproto.UpdateClusterKubeConfigReq
	resp    *cmproto.UpdateClusterKubeConfigResp
}

// NewUpdateKubeconfigAction create cluster action
func NewUpdateKubeconfigAction(model store.ClusterManagerModel) *UpdateKubeconfigAction {
	return &UpdateKubeconfigAction{
		model: model,
	}
}

// getClusterBasicInfo get cluster/cloud/project info
func (ua *UpdateKubeconfigAction) getClusterBasicInfo() error {
	cluster, err := ua.model.GetCluster(ua.ctx, ua.req.ClusterID)
	if err != nil {
		blog.Errorf("get Cluster %s failed when AddNodesToCluster, %s", ua.req.ClusterID, err.Error())
		return err
	}
	ua.cluster = cluster

	cloud, err := actions.GetCloudByCloudID(ua.model, ua.cluster.Provider)
	if err != nil {
		blog.Errorf("get Cluster %s provider %s and Project %s failed, %s",
			ua.cluster.ClusterID, ua.cluster.Provider, ua.cluster.ProjectID, err.Error(),
		)
		return err
	}
	ua.cloud = cloud

	return nil
}

// validate check
func (ua *UpdateKubeconfigAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}

	// get cluster basic info(project/cluster/cloud)
	err := ua.getClusterBasicInfo()
	if err != nil {
		return err
	}

	return nil
}

func (ua *UpdateKubeconfigAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle create cluster request
func (ua *UpdateKubeconfigAction) Handle(ctx context.Context, req *cmproto.UpdateClusterKubeConfigReq, // nolint
	resp *cmproto.UpdateClusterKubeConfigResp) {
	if req == nil || resp == nil {
		blog.Errorf("create cluster failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	// check node if exist in cloud_provider
	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	// Create Cluster by CloudProvider, underlay cloud cluster manager interface
	provider, err := cloudprovider.GetClusterMgr(ua.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cluster %s relative cloud provider %s failed, %s",
			req.ClusterID, ua.cloud.CloudProvider, err.Error())
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     ua.cloud,
		AccountID: ua.cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloud provider %s/%s failed, %s",
			ua.cloud.CloudID, ua.cloud.CloudProvider, err.Error())
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	ok, err := provider.UpdateCloudKubeConfig(ua.req.KubeConfig,
		&cloudprovider.UpdateCloudKubeConfigOption{
			Cluster:      ua.cluster,
			CommonOption: *cmOption,
		},
	)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerCheckKubeConnErr, err.Error())
		return
	}

	if !ok {
		ua.setResp(common.BcsErrClusterManagerCheckKubeConnErr, "update cluster kubeConfig failed")
		return
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
