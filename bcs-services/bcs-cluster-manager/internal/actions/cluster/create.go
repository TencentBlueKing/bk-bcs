/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cluster

import (
	"context"
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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cidrmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/taskserver"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

const (
	// createClusterIDLockKey clusterID locker
	createClusterIDLockKey = "/bcs-services/bcs-cluster-manager/createClusterID"
)

// CreateAction action for create cluster
type CreateAction struct {
	ctx          context.Context
	locker       lock.DistributedLock
	model        store.ClusterManagerModel
	cloud        *cmproto.Cloud
	nodeTemplate *cmproto.NodeTemplate
	task         *cmproto.Task
	req          *cmproto.CreateClusterReq
	resp         *cmproto.CreateClusterResp
}

// NewCreateAction create cluster action
func NewCreateAction(model store.ClusterManagerModel, locker lock.DistributedLock) *CreateAction {
	return &CreateAction{
		model:  model,
		locker: locker,
	}
}

func (ca *CreateAction) applyClusterCIDR(cls *cmproto.Cluster) error {
	if len(cls.NetworkSettings.ClusterIPv4CIDR) > 0 || options.GetEditionInfo().IsCommunicationEdition() {
		return nil
	}
	// auto update set cluster cidr
	cidr, err := applyClusterCIDR(cls)
	if err != nil {
		return err
	}
	cls.NetworkSettings.ClusterIPv4CIDR = cidr

	return nil
}

func applyClusterCIDR(cls *cmproto.Cluster) (string, error) {
	cidrCli, conClose, err := cidrmanager.GetCidrClient().GetCidrManagerClient()
	if err != nil {
		return "", fmt.Errorf("获取组件cidr-manager客户端失败: %v", err)
	}
	defer func() {
		if conClose != nil {
			conClose()
		}
	}()

	timeOutCtx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	req := &cidrmanager.GetAllocatableCidrRequest{
		Region:   cls.Region,
		CidrType: utils.GlobalRouter.String(),
		VpcID:    cls.VpcID,
		CidrLen:  utils.CalMaskLen(float64(cls.NetworkSettings.CidrStep)),
	}
	resp, err := cidrCli.GetAllocatableCidr(timeOutCtx, req)
	if err != nil {
		return "", fmt.Errorf("地域[%s]vpc[%s]获取cidr资源失败: %s", cls.Region, cls.VpcID, err)
	}
	if resp.Code != 0 {
		return "", fmt.Errorf("地域[%s]vpc[%s]获取cidr资源失败: %s", cls.Region, cls.VpcID, resp.Message)
	}

	if resp.Data.Cidr == "" {
		return "", fmt.Errorf("vpc[%s] GlobalRouter cidr资源不足", cls.VpcID)
	}

	blog.Infof("createCluster[%s] apply cidr[%s] successful", cls.ClusterID, resp.Data.Cidr)
	return resp.Data.Cidr, nil
}

func (ca *CreateAction) constructCluster(cloud *cmproto.Cloud) (*cmproto.Cluster, error) {
	createTime := time.Now().Format(time.RFC3339)
	cls := &cmproto.Cluster{
		ClusterID:   ca.req.ClusterID,
		ClusterName: ca.req.ClusterName,
		SystemID:    ca.req.CloudID,
		NetworkType: ca.req.NetworkType,
		// associate cloud template cloudID
		Provider:                ca.req.Provider,
		Region:                  ca.req.Region,
		VpcID:                   ca.req.VpcID,
		ProjectID:               ca.req.ProjectID,
		BusinessID:              ca.req.BusinessID,
		Environment:             ca.req.Environment,
		EngineType:              ca.req.EngineType,
		IsExclusive:             ca.req.IsExclusive,
		ClusterType:             ca.req.ClusterType,
		FederationClusterID:     ca.req.FederationClusterID,
		Labels:                  ca.req.Labels,
		BcsAddons:               ca.req.BcsAddons,
		ExtraAddons:             ca.req.ExtraAddons,
		ManageType:              ca.req.ManageType,
		ClusterBasicSettings:    ca.req.ClusterBasicSettings,
		NetworkSettings:         ca.req.NetworkSettings,
		ClusterAdvanceSettings:  ca.req.ClusterAdvanceSettings,
		NodeSettings:            ca.req.NodeSettings,
		AutoGenerateMasterNodes: ca.req.AutoGenerateMasterNodes,
		Template:                ca.req.Instances,
		ExtraInfo:               ca.req.ExtraInfo,
		ModuleID:                ca.req.ModuleID,
		ExtraClusterID:          ca.req.ExtraClusterID,
		IsCommonCluster:         ca.req.IsCommonCluster,
		Description:             ca.req.Description,
		ClusterCategory:         ca.req.ClusterCategory,
		IsShared:                ca.req.IsShared,
		Creator:                 ca.req.Creator,
		CloudAccountID:          ca.req.CloudAccountID,
		CreateTime:              createTime,
		UpdateTime:              createTime,
		Status:                  common.StatusInitialization,
	}

	// set cloud default values
	cloudInfoMgr, err := cloudprovider.GetCloudInfoMgr(cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s CloudInfoMgr Cluster %s failed, %s",
			cloud.CloudProvider, ca.req.ClusterID, err.Error())
		return nil, err
	}
	err = cloudInfoMgr.InitCloudClusterDefaultInfo(cls, &cloudprovider.InitClusterConfigOption{
		Cloud: cloud,
		ClusterVersion: func() string {
			if ca.req.ClusterBasicSettings != nil {
				return ca.req.ClusterBasicSettings.Version
			}
			return ""
		}(),
	})
	if err != nil {
		blog.Errorf("Create Cloud[%s] Cluster set defaultInfo failed: %v", ca.cloud.CloudProvider, err)
		return cls, err
	}

	// check cloud master nodes
	err = ca.checkClusterMasterNodes(cls)
	if err != nil {
		return cls, err
	}

	return cls, err
}

func (ca *CreateAction) checkClusterWorkerNodes(cls *cmproto.Cluster) error {
	for _, nodeIP := range ca.req.Nodes {
		n, err := ca.transNodeIPToCloudNode(nodeIP)
		if err != nil {
			blog.Errorf("createCluster checkClusterWorkerNodes[%s] failed: %v", nodeIP, err)
			continue
		}
		n.ClusterID = cls.ClusterID
		n.Status = common.StatusInitialization
		n.NodeTemplateID = ca.req.NodeTemplateID

		err = importClusterNode(ca.model, n)
		if err != nil {
			blog.Errorf("createCluster checkClusterWorkerNodes[%s] failed: %v", nodeIP, err)
			continue
		}
	}

	return nil
}

// checkClusterMasterNodes for check cloud node
func (ca *CreateAction) checkClusterMasterNodes(cls *cmproto.Cluster) error {
	// setting master node for storage
	cls.Master = make(map[string]*cmproto.Node)
	for _, masterIP := range ca.req.Master {
		node, err := ca.transNodeIPToCloudNode(masterIP)
		if err != nil {
			errMsg := fmt.Errorf("createCluster transNodeIPToCloudNode[%s] failed: %v", masterIP, err)
			blog.Errorf(errMsg.Error())
			return errMsg
		}
		cls.Master[masterIP] = node
	}

	return nil
}

// transNodeIPToCloudNode by req nodeIPs trans to cloud node
func (ca *CreateAction) transNodeIPToCloudNode(ip string) (*cmproto.Node, error) {
	nodeMgr, err := cloudprovider.GetNodeMgr(ca.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s NodeManager Cluster %s failed, %s",
			ca.cloud.CloudProvider, ca.req.ClusterID, err.Error())
		return nil, err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     ca.cloud,
		AccountID: ca.req.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s cluster %s failed, %s",
			ca.cloud.CloudID, ca.cloud.CloudProvider, ca.req.ClusterID, err.Error())
		return nil, err
	}
	cmOption.Region = ca.req.Region

	// cluster check instance if exist, validate nodes existence
	node, err := nodeMgr.GetNodeByIP(ip, &cloudprovider.GetNodeOption{
		Common:       cmOption,
		ClusterVPCID: ca.req.VpcID,
	})
	if err != nil {
		blog.Errorf("validate nodes %s existence failed, %s", ip, err.Error())
		return nil, err
	}

	blog.Infof("get cloud[%s] IP[%s] to Node successfully", ca.cloud.CloudProvider, ip)
	return node, nil
}

func (ca *CreateAction) validate() error {
	if err := ca.req.Validate(); err != nil {
		return err
	}
	// kubernetes version
	if len(ca.req.ClusterBasicSettings.Version) == 0 {
		return fmt.Errorf("lost kubernetes version in request")
	}

	// check masterIP
	if ca.req.ManageType == common.ClusterManageTypeIndependent && len(ca.req.Master) == 0 {
		return fmt.Errorf("lost kubernetes cluster masterIP")
	}

	// default not handle systemReinstall
	ca.req.SystemReinstall = true

	// auto generate master nodes
	if ca.req.AutoGenerateMasterNodes && len(ca.req.Instances) == 0 {
		return fmt.Errorf("invalid instanceTemplate config when AutoGenerateMasterNodes=true")
	}

	// use existed instances
	if !ca.req.AutoGenerateMasterNodes {
		switch ca.req.ManageType {
		case common.ClusterManageTypeManaged:
			if len(ca.req.Nodes) == 0 {
				return fmt.Errorf("invalid node config when AutoGenerateMasterNodes false in MANAGED_CLUSTER")
			}
		default:
			if len(ca.req.Master) == 0 {
				return fmt.Errorf("invalid master config when AutoGenerateMasterNodes false in INDEPENDENT_CLUSTER")
			}
		}
	}

	// masterIP check
	ipList := getAllIPList(ca.req.Provider, ca.model)
	for _, ip := range ca.req.Master {
		if _, ok := ipList[ip]; ok {
			errMsg := fmt.Errorf("create cluster masterIP[%s] already be used, please input other Nodes", ip)
			blog.Errorf(errMsg.Error())
			return errMsg
		}
	}
	// cluster category
	if len(ca.req.ClusterCategory) == 0 {
		ca.req.ClusterCategory = Builder
	}

	// check operator host permission
	nodes := make([]string, 0)
	nodes = append(nodes, ca.req.Master...)
	nodes = append(nodes, ca.req.Nodes...)

	if len(nodes) == 0 {
		return nil
	}

	// check operator host permission
	canUse := CheckUseNodesPermForUser(ca.req.BusinessID, ca.req.Creator, nodes)
	if !canUse {
		errMsg := fmt.Errorf("create cluster failed: user[%s] no perm to use nodes[%v] in bizID[%s]",
			ca.req.Creator, nodes, ca.req.BusinessID)
		blog.Errorf(errMsg.Error())
		return errMsg
	}

	return nil
}

func (ca *CreateAction) setResp(code uint32, msg string) {
	ca.resp.Code = code
	ca.resp.Message = msg
	ca.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ca *CreateAction) getCloudInfo(ctx context.Context, req *cmproto.CreateClusterReq) error {
	cloud, err := actions.GetCloudByCloudID(ca.model, req.Provider)
	if err != nil {
		blog.Errorf("get cluster %s relative Cloud %s failed, %s", req.ClusterID, req.CloudID, err.Error())
		return err
	}

	ca.cloud = cloud
	return nil
}

func (ca *CreateAction) importClusterData(cls *cmproto.Cluster) error {
	err := ca.generateClusterID(cls)
	if err != nil {
		return err
	}
	blog.Infof("Cluster %s only create information", ca.req.ClusterID)
	cls.Status = common.StatusRunning

	// save clusterInfo to DB
	err = importClusterData(ca.model, cls)
	if err != nil {
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return err
	}
	ca.resp.Data = cls
	// import cluster info to extra system
	importClusterExtraOperation(cls)

	ca.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)

	return nil
}

func (ca *CreateAction) generateClusterID(cls *cmproto.Cluster) error {
	if cls.ClusterID == "" {
		clusterID, clusterNum, err := generateClusterID(cls, ca.model)
		if err != nil {
			blog.Errorf("generate clusterId failed when create cluster")
			ca.resp.Data = cls
			ca.setResp(common.BcsErrClusterManagerClusterIDBuildErr, err.Error())
			return err
		}

		blog.Infof("generate clusterID[%v:%s] successful when create cluster", clusterNum, clusterID)
		cls.ClusterID = clusterID
	}

	return nil
}

// Handle create cluster request
func (ca *CreateAction) Handle(ctx context.Context, req *cmproto.CreateClusterReq, resp *cmproto.CreateClusterResp) {
	if req == nil || resp == nil {
		blog.Errorf("create cluster failed, req or resp is empty")
		return
	}
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp

	if err := ca.req.Validate(); err != nil {
		ca.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	// get cluster cloud Info
	err := ca.getCloudInfo(ctx, req)
	if err != nil {
		blog.Errorf("get cluster %s relative Cloud/Project %s failed, %s", req.ClusterID, req.ProjectID, err.Error())
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

	ca.locker.Lock(createClusterIDLockKey, []lock.LockOption{lock.LockTTL(time.Second * 10)}...)
	defer ca.locker.Unlock(createClusterIDLockKey)

	// create validate cluster
	if err = ca.validate(); err != nil {
		ca.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	// generate clusterID
	err = ca.generateClusterID(cls)
	if err != nil {
		blog.Errorf("generateClusterID failed: %v", err)
		return
	}

	// import cluster nodes
	_ = ca.checkClusterWorkerNodes(cls)

	// apply cluster CIDR Info
	err = ca.applyClusterCIDR(cls)
	if err != nil {
		ca.resp.Data = cls
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	// create cluster save to mongoDB
	// generate cluster task and dispatch it
	err = ca.createClusterTask(ctx, cls)
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
		Message:      fmt.Sprintf("创建%s集群%s", cls.Provider, cls.ClusterID),
		OpUser:       cls.Creator,
		CreateTime:   time.Now().String(),
		ClusterID:    cls.ClusterID,
		ProjectID:    req.ProjectID,
	})
	if err != nil {
		blog.Errorf("create cluster[%s] CreateOperationLog failed: %v", cls.ClusterID, err)
	}

	ca.resp.Data = cls
	ca.resp.Task = ca.task
	ca.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}

func (ca *CreateAction) createClusterTask(ctx context.Context, cls *cmproto.Cluster) error {
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
		//other db operation error
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
		AccountID: ca.req.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("Get Credential failed from Project %s and Cloud %s: %s",
			ca.req.ProjectID, ca.cloud.CloudID, err.Error())
		ca.resp.Data = cls
		ca.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		// if clean stored cluster information
		return err
	}
	coption.Region = ca.req.Region

	// create cluster task by task manager
	task, err := provider.CreateCluster(cls, &cloudprovider.CreateClusterOption{
		CommonOption: *coption,
		Reinstall:    ca.req.SystemReinstall,
		InitPassword: ca.req.InitLoginPassword,
		Operator:     ca.req.Creator,
		Cloud:        ca.cloud,
		Nodes:        ca.req.Nodes,
		NodeTemplate: func() *cmproto.NodeTemplate {
			if ca.req.NodeTemplateID == "" {
				return nil
			}

			template, errGet := ca.model.GetNodeTemplateByID(ca.ctx, ca.req.NodeTemplateID)
			if errGet != nil {
				blog.Errorf("CreateAction GetNodeTemplateByID[%s] failed: %v", cls.ClusterID, errGet)
				return nil
			}

			return template
		}(),
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
