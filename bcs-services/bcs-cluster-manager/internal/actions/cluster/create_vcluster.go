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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/lock"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/taskserver"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// CreateVirtualClusterAction action for create virtual cluster
type CreateVirtualClusterAction struct {
	ctx    context.Context
	locker lock.DistributedLock
	model  store.ClusterManagerModel

	cloud       *cmproto.Cloud
	hostCluster *cmproto.Cluster

	task *cmproto.Task
	req  *cmproto.CreateVirtualClusterReq
	resp *cmproto.CreateVirtualClusterResp
}

// NewCreateVirtualClusterAction create virtual cluster action
func NewCreateVirtualClusterAction(model store.ClusterManagerModel,
	locker lock.DistributedLock) *CreateVirtualClusterAction {
	return &CreateVirtualClusterAction{
		model:  model,
		locker: locker,
	}
}

func (ca *CreateVirtualClusterAction) constructCluster(cloud *cmproto.Cloud) (*cmproto.Cluster, error) {
	createTime := time.Now().Format(time.RFC3339)
	cls := &cmproto.Cluster{
		ClusterID:   ca.req.ClusterID,
		ClusterName: ca.req.ClusterName,
		Provider: func() string {
			if ca.req.Provider != "" {
				return ca.req.Provider
			}

			return cloud.CloudID
		}(),
		Region:                 ca.req.Region,
		VpcID:                  ca.hostCluster.VpcID,
		ProjectID:              ca.req.ProjectID,
		BusinessID:             ca.req.BusinessID,
		Environment:            ca.req.Environment,
		EngineType:             ca.req.EngineType,
		IsExclusive:            ca.req.IsExclusive,
		ClusterType:            common.ClusterTypeVirtual,
		Labels:                 ca.req.Labels,
		Creator:                ca.req.Creator,
		CreateTime:             createTime,
		UpdateTime:             createTime,
		SystemID:               ca.hostCluster.ClusterID,
		ManageType:             common.ClusterManageTypeIndependent,
		NetworkSettings:        ca.req.NetworkSettings,
		ClusterBasicSettings:   ca.req.ClusterBasicSettings,
		ClusterAdvanceSettings: ca.req.ClusterAdvanceSettings,
		NodeSettings:           ca.req.NodeSettings,
		Status:                 common.StatusInitialization,
		Updater:                ca.req.Creator,
		NetworkType:            ca.hostCluster.NetworkType,
		IsCommonCluster:        false,
		Description:            ca.req.Description,
		IsShared:               false,
		ClusterCategory:        common.Builder,
	}

	err := ca.generateClusterID(cls)
	if err != nil {
		blog.Errorf("generateClusterID failed: %v", err)
		return nil, err
	}
	// inject namespace info & host cluster data
	cls.ExtraInfo = func() map[string]string {
		extraInfo := make(map[string]string, 0)

		ca.req.Ns.Name = utils.GenerateNamespaceName(options.GetGlobalCMOptions().PrefixVcluster,
			ca.req.ProjectCode, cls.ClusterID)
		if ca.req.Ns.Annotations == nil {
			ca.req.Ns.Annotations = make(map[string]string, 0)
		}
		ca.req.Ns.Annotations[utils.NamespaceVcluster] = cls.ClusterID

		nsInfo, _ := json.Marshal(ca.req.Ns)

		extraInfo[common.VClusterNetworkKey] = ca.req.HostClusterNetwork
		extraInfo[common.VClusterNamespaceInfo] = string(nsInfo)

		for k, v := range ca.req.ExtraInfo {
			extraInfo[k] = v
		}
		return extraInfo
	}()

	// check cloud master nodes
	err = ca.checkClusterMasterNodes(cls)
	if err != nil {
		return cls, err
	}

	return cls, err
}

// checkClusterMasterNodes for check cloud node
func (ca *CreateVirtualClusterAction) checkClusterMasterNodes(cls *cmproto.Cluster) error { // nolint
	// vcluster not init master ip
	cls.Master = make(map[string]*cmproto.Node)
	return nil
}

func (ca *CreateVirtualClusterAction) validate() error {
	err := ca.req.Validate()
	if err != nil {
		return err
	}

	// default clusterType virtual
	ca.req.ClusterType = common.ClusterTypeVirtual

	// check version
	if ca.req.GetClusterBasicSettings().GetVersion() == "" {
		return fmt.Errorf("vcluster version empty")
	}

	// check project code
	if ca.req.GetProjectCode() == "" {
		return fmt.Errorf("CreateVirtualClusterAction projectCode empty")
	}

	// cluster host network type
	if len(ca.req.HostClusterNetwork) == 0 {
		ca.req.HostClusterNetwork = utils.DEVNET.String()
	}

	// auto select host cluster by different methods
	if len(ca.req.HostClusterID) == 0 {
		ca.req.HostClusterID, err = selectVclusterHostCluster(ca.model, VClusterHostFilterInfo{
			Provider: ca.req.Provider,
			Region:   ca.req.Region,
			Version: func() string {
				if ca.req.ClusterBasicSettings == nil {
					return ""
				}

				return ca.req.GetClusterBasicSettings().GetVersion()
			}(),
		})
		if err != nil {
			return err
		}
	}

	// init namespace info
	if ca.req.Ns == nil || ca.req.Ns.Quota == nil {
		return fmt.Errorf("CreateVirtualClusterAction namespace quota is empty")
	}
	if ca.req.Ns.Annotations == nil {
		ca.req.Ns.Annotations = map[string]string{
			utils.ProjectCode:      ca.req.ProjectCode,
			utils.NamespaceCreator: ca.req.Creator,
		}
	} else {
		ca.req.Ns.Annotations[utils.ProjectCode] = ca.req.ProjectCode
		ca.req.Ns.Annotations[utils.NamespaceCreator] = ca.req.Creator
	}

	return nil
}

func (ca *CreateVirtualClusterAction) setResp(code uint32, msg string) {
	ca.resp.Code = code
	ca.resp.Message = msg
	ca.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ca *CreateVirtualClusterAction) importClusterData(cls *cmproto.Cluster) error {
	blog.Infof("Cluster %s only create information", ca.req.ClusterID)
	cls.Status = common.StatusRunning

	// save clusterInfo to DB
	if err := ca.model.CreateCluster(ca.ctx, cls); err != nil {
		if errors.Is(err, drivers.ErrTableRecordDuplicateKey) {
			ca.setResp(common.BcsErrClusterManagerDatabaseRecordDuplicateKey, err.Error())
			return err
		}
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	ca.resp.Data = cls
	// import cluster info to extra system
	importClusterExtraOperation(cls)

	ca.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)

	return nil
}

func (ca *CreateVirtualClusterAction) generateClusterID(cls *cmproto.Cluster) error {
	if cls.ClusterID == "" {
		clusterID, clusterNum, err := generateClusterID(cls, ca.model)
		if err != nil {
			blog.Errorf("generate clusterId failed when create cluster")
			return err
		}

		blog.Infof("generate clusterID[%v:%s] successful when create cluster", clusterNum, clusterID)
		cls.ClusterID = clusterID
	}

	return nil
}

// Handle create virtual cluster request
func (ca *CreateVirtualClusterAction) Handle(ctx context.Context, req *cmproto.CreateVirtualClusterReq,
	resp *cmproto.CreateVirtualClusterResp) {
	if req == nil || resp == nil {
		blog.Errorf("create virtual cluster failed, req or resp is empty")
		return
	}
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp

	var err error

	// create validate cluster
	if err = ca.validate(); err != nil {
		ca.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	// get host cluster info
	ca.cloud, ca.hostCluster, err = actions.GetCloudAndCluster(ca.model, ca.req.Provider, ca.req.HostClusterID)
	if err != nil {
		blog.Errorf("CreateVirtualClusterAction getCloudAndCluster failed: %v", err)
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	// init cluster and set cloud default info
	cls, err := ca.constructCluster(ca.cloud)
	if err != nil {
		blog.Errorf("CreateCluster constructCluster failed: %v", err)
		ca.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	// only create cluster information, for that cluster already exists
	if ca.req.OnlyCreateInfo {
		_ = ca.importClusterData(cls)
		return
	}

	ca.locker.Lock(createClusterIDLockKey, []lock.LockOption{lock.LockTTL(time.Second * 10)}...) // nolint
	defer ca.locker.Unlock(createClusterIDLockKey)                                               // nolint

	// create cluster save to mongoDB
	// generate cluster task and dispatch it
	err = ca.createVirtualClusterTask(ctx, cls)
	if err != nil {
		return
	}
	blog.Infof("create cluster[%s] task cloud[%s] provider[%s] successfully",
		cls.ClusterName, ca.cloud.CloudID, ca.cloud.CloudProvider)

	// build operationLog
	err = ca.model.CreateOperationLog(ca.ctx, &cmproto.OperationLog{
		ResourceType: common.Cluster.String(),
		ResourceID:   cls.ClusterID,
		TaskID:       ca.task.TaskID,
		Message:      fmt.Sprintf("创建%s虚拟集群%s", cls.Provider, cls.ClusterID),
		OpUser:       cls.Creator,
		CreateTime:   time.Now().String(),
		ClusterID:    cls.ClusterID,
		ProjectID:    ca.req.ProjectID,
	})
	if err != nil {
		blog.Errorf("create cluster[%s] CreateOperationLog failed: %v", cls.ClusterID, err)
	}

	ca.resp.Data = cls
	ca.resp.Task = ca.task
	ca.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

func (ca *CreateVirtualClusterAction) createVirtualClusterTask(ctx context.Context, cls *cmproto.Cluster) error {
	// step1: create cluster to save mongo
	// step2: call cloud provider cluster_manager feature to create cluster task
	err := ca.model.CreateCluster(ctx, cls)
	if err != nil {
		blog.Errorf("save Cluster %s information to store failed, %s", cls.ClusterID, err.Error())
		if errors.Is(err, drivers.ErrTableRecordDuplicateKey) {
			ca.resp.Data = cls
			ca.setResp(common.BcsErrClusterManagerDatabaseRecordDuplicateKey, err.Error())
			return err
		}
		// other db operation error
		ca.resp.Data = cls
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}

	// Create Cluster by CloudProvider, underlay cloud cluster manager interface
	provider, err := cloudprovider.GetClusterMgr(ca.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cluster %s relative cloud provider %s failed, %s",
			ca.req.ClusterID, ca.cloud.CloudProvider, err.Error())
		ca.resp.Data = cls
		ca.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return err
	}

	// first, get cloud credentialInfo from cloud; second, get cloud credentialInfo from cluster
	coption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     ca.cloud,
		AccountID: ca.hostCluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("Get Credential failed from Project %s and Cloud %s: %s",
			ca.req.ProjectID, ca.cloud.CloudID, err.Error())
		ca.resp.Data = cls
		ca.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return err
	}
	coption.Region = ca.req.Region

	// create cluster task by task manager
	task, err := provider.CreateVirtualCluster(cls, &cloudprovider.CreateVirtualClusterOption{
		CommonOption: *coption,
		Operator:     ca.req.Creator,
		Cloud:        ca.cloud,
		HostCluster:  ca.hostCluster,
		Namespace:    ca.req.Ns,
	})
	if err != nil {
		blog.Errorf("create Cluster %s by Cloud %s with provider %s failed, %s",
			ca.req.ClusterID, ca.cloud.CloudID, ca.cloud.CloudProvider, err.Error())
		ca.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return err
	}

	// create task and dispatch task
	if err := ca.model.CreateTask(ca.ctx, task); err != nil {
		blog.Errorf("save create cluster task for cluster %s failed, %s",
			cls.ClusterName, err.Error(),
		)
		ca.resp.Data = cls
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	if err := taskserver.GetTaskServer().Dispatch(task); err != nil {
		blog.Errorf("dispatch create cluster task for cluster %s failed, %s",
			cls.ClusterName, err.Error(),
		)
		ca.resp.Data = cls
		ca.setResp(common.BcsErrClusterManagerTaskErr, err.Error())
		return err
	}

	ca.task = task
	return nil
}
