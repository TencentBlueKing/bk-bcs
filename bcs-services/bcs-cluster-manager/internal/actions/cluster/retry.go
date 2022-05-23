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
 *
 */

package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/taskserver"
)

// RetryCreateAction action for retry create cluster
type RetryCreateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.RetryCreateClusterReq
	resp  *cmproto.RetryCreateClusterResp
}

// NewRetryCreateAction retry create cluster action
func NewRetryCreateAction(model store.ClusterManagerModel) *RetryCreateAction {
	return &RetryCreateAction{
		model: model,
	}
}

func (ra *RetryCreateAction) setResp(code uint32, msg string) {
	ra.resp.Code = code
	ra.resp.Message = msg
	ra.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ra *RetryCreateAction) validate(cluster *cmproto.Cluster) error {
	if cluster.Status == common.StatusInitialization {
		errMsg := fmt.Errorf("cluster[%s] current status %s, not allow retry create", cluster.ClusterID, cluster.Status)
		return errMsg
	}

	return nil
}

// Handle retry create cluster request
func (ra *RetryCreateAction) Handle(ctx context.Context, req *cmproto.RetryCreateClusterReq, resp *cmproto.RetryCreateClusterResp) {
	if req == nil || resp == nil {
		blog.Errorf("retry createCluster failed, req or resp is empty")
		return
	}
	ra.ctx = ctx
	ra.req = req
	ra.resp = resp

	cls, err := ra.model.GetCluster(ra.ctx, ra.req.ClusterID)
	if err != nil {
		blog.Errorf("get cluster %s failed: %v", ra.req.ClusterID, err)
		ra.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	err = ra.validate(cls)
	if err != nil {
		blog.Errorf(err.Error())
		ra.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	cloud, err := ra.model.GetCloud(ctx, cls.Provider)
	if err != nil {
		blog.Errorf("get cluster %s relative Cloud %s failed, %s",
			req.ClusterID, cls.ProjectID, err.Error(),
		)
		ra.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	// Create Cluster by CloudProvider, underlay cloud cluster manager interface
	provider, err := cloudprovider.GetClusterMgr(cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cluster %s relative cloud provider %s failed, %s",
			req.ClusterID, cloud.CloudProvider, err.Error())
		ra.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	cls.Status = common.StatusInitialization
	// step1: update cluster to save mongo
	// step2: call cloud provider cluster_manager feature to create cluster task
	if err = ra.model.UpdateCluster(ctx, cls); err != nil {
		blog.Errorf("update Cluster %s information to store failed, %s", cls.ClusterID, err.Error())
		//other db operation error
		ra.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	// project save cloud credential info
	// first, get cloud credentialInfo from project; second, get from cloud provider when failed to obtain
	coption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     cloud,
		AccountID: cls.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("Get Credential failed from Cloud %s: %s", cloud.CloudID, err.Error())
		ra.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	coption.Region = cls.Region

	// create cluster task by task manager
	task, err := provider.CreateCluster(cls, &cloudprovider.CreateClusterOption{
		CommonOption: *coption,
		Operator:     cls.Creator,
		Cloud:        cloud,
	})
	if err != nil {
		blog.Errorf("create Cluster %s by Cloud %s with provider %s failed, %s",
			req.ClusterID, cloud.CloudID, cloud.CloudProvider, err.Error())
		ra.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	// create task and dispatch task
	if err = ra.model.CreateTask(ra.ctx, task); err != nil {
		blog.Errorf("save create cluster task for cluster %s failed, %s", cls.ClusterName, err.Error())
		ra.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	if err = taskserver.GetTaskServer().Dispatch(task); err != nil {
		blog.Errorf("dispatch create cluster task for cluster %s failed, %s", cls.ClusterName, err.Error())
		ra.setResp(common.BcsErrClusterManagerTaskErr, err.Error())
		return
	}
	blog.Infof("retry create cluster[%s] task cloud[%s] provider[%s] successfully",
		cls.ClusterName, cloud.CloudID, cloud.CloudProvider)

	err = ra.model.CreateOperationLog(ra.ctx, &cmproto.OperationLog{
		ResourceType: common.Cluster.String(),
		ResourceID:   cls.ClusterID,
		TaskID:       task.TaskID,
		Message:      fmt.Sprintf("重试创建%s集群%s", cls.Provider, cls.ClusterID),
		OpUser:       req.Operator,
		CreateTime:   time.Now().String(),
	})
	if err != nil {
		blog.Errorf("create cluster[%s] CreateOperationLog failed: %v", cls.ClusterID, err)
	}

	ra.resp.Data = cls
	ra.resp.Task = task
	ra.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
