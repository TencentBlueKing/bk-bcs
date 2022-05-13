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
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/lock"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/taskserver"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"

	"github.com/golang/protobuf/ptypes/wrappers"
)

// ImportAction action for import cluster
type ImportAction struct {
	ctx     context.Context
	model   store.ClusterManagerModel
	locker  lock.DistributedLock
	project *cmproto.Project
	cloud   *cmproto.Cloud
	task    *cmproto.Task
	cluster *cmproto.Cluster
	req     *cmproto.ImportClusterReq
	resp    *cmproto.ImportClusterResp
}

// NewImportAction import cluster action
func NewImportAction(model store.ClusterManagerModel, locker lock.DistributedLock) *ImportAction {
	return &ImportAction{
		model:  model,
		locker: locker,
	}
}

func (ia *ImportAction) constructCluster() *cmproto.Cluster {
	createTime := time.Now().Format(time.RFC3339)
	cls := &cmproto.Cluster{
		ClusterID:   ia.req.ClusterID,
		ClusterName: ia.req.ClusterName,
		Provider:    ia.req.Provider,
		Region:      ia.req.Region,
		ProjectID:   ia.req.ProjectID,
		Description: ia.req.Description,
		BusinessID:  ia.req.BusinessID,
		Environment: ia.req.Environment,
		EngineType:  ia.req.EngineType,
		IsExclusive: ia.req.IsExclusive.GetValue(),
		ClusterType: ia.req.ClusterType,
		ManageType:  ia.req.ManageType,
		NetworkType: ia.req.NetworkType,
		// associate cloud template cloudID
		Labels:          ia.req.Labels,
		ExtraInfo:       ia.req.ExtraInfo,
		CreateTime:      createTime,
		UpdateTime:      createTime,
		Creator:         ia.req.Creator,
		Updater:         ia.req.Creator,
		ClusterCategory: ia.req.ClusterCategory,
		// import cluster category
		ImportCategory: func() string {
			if ia.req.CloudMode.KubeConfig != "" {
				return KubeConfig
			}
			return Cloud
		}(),
		IsShared: ia.req.IsShared,
	}

	return cls
}

func (ia *ImportAction) syncClusterInfoToDB(cls *cmproto.Cluster) error {
	// generate ClusterID
	err := ia.generateClusterID(cls)
	if err != nil {
		return err
	}

	// save kubeConfig
	if ia.req.CloudMode.KubeConfig != "" {
		kubeRet := base64.StdEncoding.EncodeToString([]byte(ia.req.CloudMode.KubeConfig))
		cls.KubeConfig = kubeRet
	}

	// update imported cluster status
	cls.Status = common.StatusInitialization
	ia.cluster = cls

	err = ia.model.CreateCluster(ia.ctx, cls)
	if err != nil {
		blog.Errorf("save Cluster %s information to store failed, %s", cls.ClusterID, err.Error())
		if errors.Is(err, drivers.ErrTableRecordDuplicateKey) {
			return err
		}
		return err
	}

	return nil
}

func (ia *ImportAction) syncClusterCloudConfig(cls *cmproto.Cluster) error {
	cloudInfoMgr, err := cloudprovider.GetCloudInfoMgr(ia.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s CloudInfoMgr Cluster %s failed, %s",
			ia.cloud.CloudProvider, ia.req.ClusterID, err.Error())
		return err
	}
	cmOption, err := cloudprovider.GetCredential(ia.project, ia.cloud)
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s cluster %s failed, %s",
			ia.cloud.CloudID, ia.cloud.CloudProvider, ia.req.ClusterID, err.Error())
		return err
	}
	cmOption.Region = ia.req.Region

	// sync cluster cloud related info: vpc、systemID、network、clusterBasicSetting
	err = cloudInfoMgr.SyncClusterCloudInfo(cls, &cloudprovider.SyncClusterCloudInfoOption{
		Common:         cmOption,
		Cloud:          ia.cloud,
		ImportMode:     ia.req.CloudMode,
		ClusterVersion: ia.req.Version,
	})
	if err != nil {
		return err
	}

	return nil
}

func (ia *ImportAction) setResp(code uint32, msg string) {
	ia.resp.Code = code
	ia.resp.Message = msg
	ia.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	ia.setResponseData(ia.resp.Result)
}

func (ia *ImportAction) setResponseData(result bool) {
	if !result {
		return
	}

	respData := map[string]interface{}{
		"cluster": ia.cluster,
		"task":    ia.task,
	}

	data, err := utils.MapToProtobufStruct(respData)
	if err != nil {
		blog.Errorf("ImportAction[%s] trans Data failed: %v", ia.cluster.ClusterID, err)
		return
	}
	ia.resp.Data = data

	return
}

// Handle create cluster request
func (ia *ImportAction) Handle(ctx context.Context, req *cmproto.ImportClusterReq, resp *cmproto.ImportClusterResp) {
	if req == nil || resp == nil {
		blog.Errorf("import cluster failed, req or resp is empty")
		return
	}
	ia.ctx = ctx
	ia.req = req
	ia.resp = resp

	// parameters check
	if err := ia.req.Validate(); err != nil {
		ia.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	// get cluster cloud and project info
	err := ia.getCloudProjectInfo(ctx, req)
	if err != nil {
		blog.Errorf("get cluster %s relative Cloud/Project %s failed, %s", req.ClusterID, req.ProjectID, err.Error())
		ia.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ia.locker.Lock(createClusterIDLockKey, []lock.LockOption{lock.LockTTL(time.Second * 10)}...)
	defer ia.locker.Unlock(createClusterIDLockKey)

	// import validate cluster
	if err = ia.validate(); err != nil {
		ia.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	// init cluster and set cloud default info
	cls := ia.constructCluster()
	err = ia.syncClusterCloudConfig(cls)
	if err != nil {
		ia.setResp(common.BcsErrClusterManagerSyncCloudErr, err.Error())
		return
	}

	// create cluster save to mongoDB
	err = ia.syncClusterInfoToDB(cls)
	if err != nil {
		ia.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	// generate cluster task and dispatch it
	err = ia.importClusterTask(ctx, cls)
	if err != nil {
		return
	}
	blog.Infof("create cluster[%s] task cloud[%s] provider[%s] successfully",
		cls.ClusterName, ia.cloud.CloudID, ia.cloud.CloudProvider)

	// import cluster info to extra system
	importClusterExtraOperation(cls)

	// build operationLog
	err = ia.model.CreateOperationLog(ctx, &cmproto.OperationLog{
		ResourceType: common.Cluster.String(),
		ResourceID:   cls.ClusterID,
		TaskID:       ia.task.TaskID,
		Message:      fmt.Sprintf("导入%s集群%s", cls.Provider, cls.ClusterID),
		OpUser:       cls.Creator,
		CreateTime:   time.Now().String(),
	})
	if err != nil {
		blog.Errorf("import cluster[%s] CreateOperationLog failed: %v", cls.ClusterID, err)
	}

	ia.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}

func (ia *ImportAction) importClusterTask(ctx context.Context, cls *cmproto.Cluster) error {
	// call cloud provider cluster_manager feature to import cluster task
	// Import Cluster by CloudProvider, underlay cloud cluster manager interface
	provider, err := cloudprovider.GetClusterMgr(ia.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cluster %s relative cloud provider %s failed, %s", ia.req.ClusterID, ia.cloud.CloudProvider, err.Error())
		ia.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return err
	}

	// first, get cloud credentialInfo from project; second, get from cloud provider when failed to obtain
	coption, err := cloudprovider.GetCredential(ia.project, ia.cloud)
	if err != nil {
		blog.Errorf("Get Credential failed from Project %s and Cloud %s: %s",
			ia.project.ProjectID, ia.cloud.CloudID, err.Error())
		ia.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return err
	}
	coption.Region = ia.req.Region

	// create cluster task by task manager
	task, err := provider.ImportCluster(cls, &cloudprovider.ImportClusterOption{
		CommonOption: *coption,
		Cloud:        ia.cloud,
		CloudMode:    ia.req.CloudMode,
		Operator:     ia.req.Creator,
	})
	if err != nil {
		blog.Errorf("create Cluster %s by Cloud %s with provider %s failed, %s",
			ia.req.ClusterID, ia.cloud.CloudID, ia.cloud.CloudProvider, err.Error())
		ia.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return err
	}

	// create task and dispatch task
	if err := ia.model.CreateTask(ctx, task); err != nil {
		blog.Errorf("save create cluster task for cluster %s failed, %s",
			cls.ClusterName, err.Error(),
		)
		ia.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	if err := taskserver.GetTaskServer().Dispatch(task); err != nil {
		blog.Errorf("dispatch create cluster task for cluster %s failed, %s",
			cls.ClusterName, err.Error(),
		)
		ia.setResp(common.BcsErrClusterManagerTaskErr, err.Error())
		return err
	}

	ia.task = task
	return nil
}

func (ia *ImportAction) getCloudProjectInfo(ctx context.Context, req *cmproto.ImportClusterReq) error {
	cloud, err := ia.model.GetCloud(ctx, req.Provider)
	if err != nil {
		blog.Errorf("get cluster %s relative Cloud %s failed, %s", req.ClusterID, req.Provider, err.Error())
		return err
	}
	ia.cloud = cloud

	project, err := ia.model.GetProject(ctx, req.ProjectID)
	if err != nil {
		blog.Errorf("get cluster %s relative Project %s failed, %s", req.ClusterID, req.ProjectID, err.Error())
		return err
	}
	ia.project = project

	return nil
}

func (ia *ImportAction) generateClusterID(cls *cmproto.Cluster) error {
	if cls.ClusterID == "" {
		clusterID, clusterNum, err := generateClusterID(cls, ia.model)
		if err != nil {
			blog.Errorf("generate clusterId failed when import cluster")
			return err
		}

		blog.Infof("generate clusterID[%v:%s] successful when impport cluster", clusterNum, clusterID)
		cls.ClusterID = clusterID
	}

	return nil
}

// commonValidate importCluster common validate
func commonValidate(req *cmproto.ImportClusterReq) error {
	if req.GetEngineType() == "" {
		req.EngineType = common.ClusterEngineTypeK8s
	}
	if req.GetIsExclusive() == nil {
		req.IsExclusive = &wrappers.BoolValue{Value: true}
	}
	if req.ClusterType == "" {
		req.ClusterType = common.ClusterManageTypeIndependent
	}
	if req.NetworkType == "" {
		req.NetworkType = common.ClusterOverlayNetwork
	}
	req.ClusterCategory = Importer

	if req.CloudMode == nil {
		return fmt.Errorf("ImportCluster CommonValidate failed: CloudMode empty")
	}

	if req.CloudMode.CloudID == "" && req.CloudMode.KubeConfig == "" {
		return fmt.Errorf("ImportCluster CommonValidate CloudMode cloudID&kubeConfig empty")
	}

	return nil
}

func (ia *ImportAction) validate() error {
	// common validate
	if err := commonValidate(ia.req); err != nil {
		return err
	}
	// cloud validate
	cloudValidate, err := cloudprovider.GetCloudValidateMgr(ia.cloud.CloudProvider)
	if err != nil {
		return err
	}
	// first, get cloud credentialInfo from project; second, get from cloud provider when failed to obtain
	cOption, err := cloudprovider.GetCredential(ia.project, ia.cloud)
	if err != nil {
		blog.Errorf("Get Credential failed from Project %s and Cloud %s: %s", ia.project.ProjectID,
			ia.cloud.CloudID, err.Error())
		return err
	}
	cOption.Region = ia.req.Region

	err = cloudValidate.ImportClusterValidate(ia.req, cOption)
	if err != nil {
		return err
	}
	return nil
}

// CheckKubeAction action for check cluster kubeConfig
type CheckKubeAction struct {
	ctx  context.Context
	req  *cmproto.KubeConfigReq
	resp *cmproto.KubeConfigResp
}

// NewCheckKubeAction check cluster kubeConfig action
func NewCheckKubeAction() *CheckKubeAction {
	return &CheckKubeAction{}
}

func (ka *CheckKubeAction) setResp(code uint32, msg string) {
	ka.resp.Code = code
	ka.resp.Message = msg
	ka.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle create cluster request
func (ka *CheckKubeAction) Handle(ctx context.Context, req *cmproto.KubeConfigReq, resp *cmproto.KubeConfigResp) {
	if req == nil || resp == nil {
		blog.Errorf("check cluster kubeConfig failed, req or resp is empty")
		return
	}
	ka.ctx = ctx
	ka.req = req
	ka.resp = resp

	// import validate cluster
	if err := req.Validate(); err != nil {
		ka.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	err := checkKubeConfig(req.KubeConfig)
	if err != nil {
		ka.setResp(common.BcsErrClusterManagerCheckKubeErr, err.Error())
		return
	}

	ka.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}

func checkKubeConfig(kubeConfig string) error {
	_, err := types.GetKubeConfigFromYAMLBody(false, types.YamlInput{
		FileName:    "",
		YamlContent: kubeConfig,
	})
	if err != nil {
		return fmt.Errorf("checkKubeConfig validate failed: %v", err)
	}

	kubeRet := base64.StdEncoding.EncodeToString([]byte(kubeConfig))
	kubeCli, err := clusterops.NewKubeClient(kubeRet)
	if err != nil {
		return fmt.Errorf("checkKubeConfig validate failed: %v", err)
	}

	_, err = kubeCli.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("checkKubeConfig connect cluster failed: %v", err)
	}
	blog.Infof("checkKubeConfig YAMLStyle and connectCluster success")

	return nil
}
