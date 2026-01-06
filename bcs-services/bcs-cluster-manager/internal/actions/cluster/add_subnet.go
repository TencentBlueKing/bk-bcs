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

// AddSubnetToClusterAction for cluster add subnet
type AddSubnetToClusterAction struct {
	ctx context.Context

	model store.ClusterManagerModel
	req   *cmproto.AddSubnetToClusterReq
	resp  *cmproto.AddSubnetToClusterResp

	cluster *cmproto.Cluster
	cloud   *cmproto.Cloud
	account *cmproto.CloudAccount // nolint
}

// NewAddSubnetToClusterAction create addSubnet action
func NewAddSubnetToClusterAction(model store.ClusterManagerModel) *AddSubnetToClusterAction {
	return &AddSubnetToClusterAction{
		model: model,
	}
}

func (ga *AddSubnetToClusterAction) validate() error {
	err := ga.req.Validate()
	if err != nil {
		return err
	}

	if ga.req.Subnet == nil || (len(ga.req.Subnet.GetNew()) == 0 && len(ga.req.Subnet.GetExisted().GetIds()) == 0) {
		return fmt.Errorf("addSubnetToClusterAction subnetInfo empty")
	}

	return nil
}

func (ga *AddSubnetToClusterAction) getRelativeData() error {
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

func (ga *AddSubnetToClusterAction) addSubnetToCluster() error {
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     ga.cloud,
		AccountID: ga.cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s addSubnetToCluster failed, %s",
			ga.cloud.CloudID, ga.cloud.CloudProvider, err.Error())
		return err
	}
	cmOption.Region = ga.cluster.Region

	clsMgr, err := cloudprovider.GetClusterMgr(ga.cloud.CloudProvider)
	if err != nil {
		return err
	}

	err = clsMgr.AddSubnetsToCluster(ga.ctx, ga.req.GetSubnet(), &cloudprovider.AddSubnetsToClusterOption{
		CommonOption: *cmOption,
		Cluster:      ga.cluster,
	})
	if err != nil {
		blog.Errorf("addSubnetToClusterAction call clsMgr addSubnetsToCluster failed: %v", err)
		return err
	}

	return nil
}

func (ga *AddSubnetToClusterAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle add subnet action
func (ga *AddSubnetToClusterAction) Handle(ctx context.Context,
	req *cmproto.AddSubnetToClusterReq, resp *cmproto.AddSubnetToClusterResp) {
	if req == nil || resp == nil {
		blog.Errorf("addSubnetToClusterAction failed, req or resp is empty")
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

	if err := ga.addSubnetToCluster(); err != nil {
		ga.setResp(common.BcsErrClusterManagerClsMgrCloudErr, err.Error())
		return
	}

	err := ga.model.CreateOperationLog(ga.ctx, &cmproto.OperationLog{
		ResourceType: common.Cluster.String(),
		ResourceID:   ga.req.ClusterID,
		TaskID:       "",
		Message:      fmt.Sprintf("集群[%s]添加子网资源", ga.req.ClusterID),
		OpUser:       auth.GetUserFromCtx(ctx),
		CreateTime:   time.Now().UTC().Format(time.RFC3339),
		ClusterID:    ga.req.ClusterID,
		ProjectID:    ga.cluster.ProjectID,
		ResourceName: ga.cluster.ClusterName,
	})
	if err != nil {
		blog.Errorf("AddSubnetToCluster[%s] CreateOperationLog failed: %v", ga.req.ClusterID, err)
	}

	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
