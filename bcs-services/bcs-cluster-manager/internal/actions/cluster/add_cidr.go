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

package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// AddClusterCidrAction for cluster add cidr
type AddClusterCidrAction struct {
	ctx context.Context

	model store.ClusterManagerModel
	req   *cmproto.AddClusterCidrReq
	resp  *cmproto.AddClusterCidrResp

	cluster *cmproto.Cluster
	cloud   *cmproto.Cloud
	account *cmproto.CloudAccount // nolint
}

// NewAddClusterCidrAction create add cluster cidr action
func NewAddClusterCidrAction(model store.ClusterManagerModel) *AddClusterCidrAction {
	return &AddClusterCidrAction{
		model: model,
	}
}

func (ga *AddClusterCidrAction) validate() error {
	err := ga.req.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (ga *AddClusterCidrAction) getRelativeData() error {
	cluster, err := actions.GetClusterInfoByClusterID(ga.model, ga.req.GetClusterID())
	if err != nil {
		return err
	}
	ga.cluster = cluster

	cloud, err := actions.GetCloudByCloudID(ga.model, ga.cluster.GetProvider())
	if err != nil {
		return err
	}
	ga.cloud = cloud

	return nil
}

func (ga *AddClusterCidrAction) addClusterCidr() error {
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     ga.cloud,
		AccountID: ga.cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s addClusterCidr failed, %s",
			ga.cloud.CloudID, ga.cloud.CloudProvider, err.Error())
		return err
	}
	cmOption.Region = ga.cluster.Region

	clsMgr, err := cloudprovider.GetClusterMgr(ga.cloud.CloudProvider)
	if err != nil {
		return err
	}

	err = clsMgr.AddClusterCidr(ga.ctx, ga.req.GetCidrs(), &cloudprovider.AddSubnetsToClusterOption{
		CommonOption: *cmOption,
		Cluster:      ga.cluster,
	})
	if err != nil {
		blog.Errorf("addClusterCidrAction call clsMgr addSubnetsToCluster failed: %v", err)
		return err
	}

	return nil
}

func (ga *AddClusterCidrAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle add cluster cidr action
func (ga *AddClusterCidrAction) Handle(ctx context.Context,
	req *cmproto.AddClusterCidrReq, resp *cmproto.AddClusterCidrResp) {
	if req == nil || resp == nil {
		blog.Errorf("addClusterCidrAction failed, req or resp is empty")
		return
	}
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	if err := ga.validate(); err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ga.getRelativeData(); err != nil {
		ga.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	if err := ga.addClusterCidr(); err != nil {
		ga.setResp(common.BcsErrClusterManagerClsMgrCloudErr, err.Error())
		return
	}

	err := ga.model.CreateOperationLog(ga.ctx, &cmproto.OperationLog{
		ResourceType: common.Cluster.String(),
		ResourceID:   ga.req.ClusterID,
		TaskID:       "",
		Message:      fmt.Sprintf("集群[%s]添加cidr", ga.req.ClusterID),
		OpUser:       auth.GetUserFromCtx(ctx),
		CreateTime:   time.Now().Format(time.RFC3339),
		ClusterID:    ga.req.ClusterID,
		ProjectID:    ga.cluster.ProjectID,
		ResourceName: ga.cluster.ClusterName,
	})
	if err != nil {
		blog.Errorf("AddClusterCidr[%s] CreateOperationLog failed: %v", ga.req.ClusterID, err)
	}

	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
