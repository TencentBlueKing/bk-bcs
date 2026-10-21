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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/taskserver"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// DeleteVirtualAction action for delete virtual cluster
type DeleteVirtualAction struct {
	ctx   context.Context
	req   *cmproto.DeleteVirtualClusterReq
	resp  *cmproto.DeleteVirtualClusterResp
	model store.ClusterManagerModel

	cluster     *cmproto.Cluster
	hostCluster *cmproto.Cluster
	namespace   *cmproto.NamespaceInfo
	cloud       *cmproto.Cloud

	cmOptions *cloudprovider.CommonOption
	task      *cmproto.Task
}

// NewDeleteVirtualAction delete virtual cluster action
func NewDeleteVirtualAction(model store.ClusterManagerModel) *DeleteVirtualAction {
	return &DeleteVirtualAction{
		model: model,
	}
}

func (da *DeleteVirtualAction) getClusterInfo() error {
	cluster, err := actions.GetClusterInfoByClusterID(da.model, da.req.ClusterID)
	if err != nil {
		blog.Errorf("get virtual cluster[%s] failed: %v", da.req.ClusterID, err)
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	da.cluster = cluster

	cloud, hostCluster, err := actions.GetCloudAndCluster(da.model, cluster.Provider, cluster.SystemID)
	if err != nil {
		blog.Errorf("get provider/cluster[%s:%s] failed: %v", cluster.Provider, cluster.SystemID, err)
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	da.cloud = cloud
	da.hostCluster = hostCluster

	// parse virtual cluster in host namespace
	namespaceStr := utils.GetValueFromMap(cluster.ExtraInfo, common.VClusterNamespaceInfo)
	if len(namespaceStr) == 0 {
		return fmt.Errorf("vcluster[%s] namespace empty", cluster.ClusterID)
	}
	nsInfo := &cmproto.NamespaceInfo{}
	json.Unmarshal([]byte(namespaceStr), nsInfo) // nolint

	da.namespace = nsInfo
	coption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     cloud,
		AccountID: da.cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get Credential failed when delete Cluster %s, %s. Cloud %s",
			da.cluster.ClusterID, da.cloud.CloudID, err.Error())
		da.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return err
	}
	coption.Region = da.cluster.Region
	da.cmOptions = coption

	return nil
}

func (da *DeleteVirtualAction) cleanLocalInformation() error {
	// async delete cluster dependency info
	go asyncDeleteImportedClusterInfo(da.ctx, da.model, da.cluster)

	// finally clean cluster
	da.cluster.Status = common.StatusDeleted
	if err := da.model.UpdateCluster(da.ctx, da.cluster); err != nil {
		return err
	}
	return nil
}

func (da *DeleteVirtualAction) setResp(code uint32, msg string) {
	da.resp.Code = code
	da.resp.Message = msg
	da.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (da *DeleteVirtualAction) validate(req *cmproto.DeleteVirtualClusterReq) error {
	if err := req.Validate(); err != nil {
		return err
	}

	if len(req.Operator) == 0 {
		return fmt.Errorf("operator empty when delete cluster")
	}

	return nil
}

// Handle delete virtual cluster handler
func (da *DeleteVirtualAction) Handle(ctx context.Context,
	req *cmproto.DeleteVirtualClusterReq, resp *cmproto.DeleteVirtualClusterResp) {
	if req == nil || resp == nil {
		blog.Errorf("delete virtual cluster failed, req or resp is empty")
		return
	}
	da.ctx = ctx
	da.req = req
	da.resp = resp

	// delete parameter validate check
	if err := da.validate(req); err != nil {
		da.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	// delete cluster info
	if da.req.OnlyDeleteInfo {
		// clean all relative resource then update cluster deleted status finally
		if err := da.cleanLocalInformation(); err != nil {
			blog.Errorf("only delete Cluster %s local information err, %s", req.ClusterID, err.Error())
			da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
			return
		}

		blog.Infof("only Delete Cluster %s local information successfully", req.ClusterID)
		da.resp.Data = da.cluster
		da.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
		return
	}

	// get cluster relative info
	if err := da.getClusterInfo(); err != nil {
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			da.setResp(common.BcsErrClusterManagerDatabaseRecordNotFound, err.Error())
			return
		}
		return
	}

	// cluster is deleting or already deleted
	if da.cluster.Status == common.StatusDeleting || da.cluster.Status == common.StatusDeleted {
		blog.Warnf("Cluster %s is under %s and is not force deleting, simply return",
			req.ClusterID, da.cluster.Status)
		da.setResp(common.BcsErrClusterManagerTaskErr, "cluster is under deleting/deleted")
		return
	}

	blog.Infof("try to delete cluster %s", req.ClusterID)

	// create delete cluster task
	err := da.createDeleteClusterTask(req)
	if err != nil {
		return
	}
	blog.Infof("delete virtual cluster[%s] task cloud[%s] provider[%s] successfully",
		da.cluster.ClusterName, da.cloud.CloudID, da.cloud.CloudProvider)

	// build operation log
	err = da.model.CreateOperationLog(da.ctx, &cmproto.OperationLog{
		ResourceType: common.Cluster.String(),
		ResourceID:   da.cluster.ClusterID,
		TaskID:       da.task.TaskID,
		Message:      fmt.Sprintf("删除%s虚拟集群%s", da.cluster.Provider, da.cluster.ClusterID),
		OpUser:       da.req.Operator,
		CreateTime:   time.Now().UTC().Format(time.RFC3339),
		ClusterID:    da.cluster.ClusterID,
		ProjectID:    da.cluster.ProjectID,
		ResourceName: da.cluster.ClusterName,
	})
	if err != nil {
		blog.Errorf("delete vCluster[%s] CreateOperationLog failed: %v", da.cluster.ClusterID, err)
	}

	da.resp.Data = da.cluster
	da.resp.Task = da.task
	da.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

func (da *DeleteVirtualAction) createDeleteClusterTask(req *cmproto.DeleteVirtualClusterReq) error {
	clsMgr, err := cloudprovider.GetClusterMgr(da.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get vCluster %s real CloudProvider %s manager failed, %s",
			req.ClusterID, da.cloud.CloudProvider, err)
		da.setResp(common.BcsErrClusterManagerCommonErr, err.Error())
		return err
	}

	// update cluster deleting status
	da.cluster.Status = common.StatusDeleting
	if err = da.model.UpdateCluster(da.ctx, da.cluster); err != nil {
		blog.Errorf("update vCluster %s to status DELETING failed, %s", req.ClusterID, err.Error())
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}

	// call cloud provider api to delete virtual cluster by async tasks
	task, err := clsMgr.DeleteVirtualCluster(da.cluster, &cloudprovider.DeleteVirtualClusterOption{
		CommonOption: *da.cmOptions,
		Operator:     req.Operator,
		Cloud:        da.cloud,
		HostCluster:  da.hostCluster,
		Namespace:    da.namespace,
	})
	if err != nil {
		blog.Errorf("delete vCluster %s by cloudprovider %s failed, %s",
			da.cluster.ClusterID, da.cloud.CloudID, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerTaskErr, err.Error())
		return err
	}

	// create task and dispatch task
	if err := da.model.CreateTask(da.ctx, task); err != nil {
		blog.Errorf("save delete vCluster task for cluster %s failed, %s",
			da.cluster.ClusterName, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	if err := taskserver.GetTaskServer().Dispatch(task); err != nil {
		blog.Errorf("dispatch delete vCluster task for cluster %s failed, %s",
			da.cluster.ClusterName, err.Error(),
		)
		da.setResp(common.BcsErrClusterManagerTaskErr, err.Error())
		return err
	}

	da.task = task
	return nil
}
