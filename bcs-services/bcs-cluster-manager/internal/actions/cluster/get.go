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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	spb "google.golang.org/protobuf/types/known/structpb"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/tenant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// GetAction action for get cluster
type GetAction struct {
	ctx     context.Context
	model   store.ClusterManagerModel
	req     *cmproto.GetClusterReq
	resp    *cmproto.GetClusterResp
	cluster *cmproto.Cluster
	cloud   *cmproto.Cloud
}

// NewGetAction create get action
func NewGetAction(model store.ClusterManagerModel) *GetAction {
	return &GetAction{
		model: model,
	}
}

func (ga *GetAction) validate() error {
	return ga.req.Validate()
}

func (ga *GetAction) getCluster() error {
	cluster, err := ga.model.GetCluster(ga.ctx, ga.req.ClusterID)
	if err != nil {
		return err
	}
	cluster.CreateTime = utils.TransTimeFormat(cluster.CreateTime)
	cluster.UpdateTime = utils.TransTimeFormat(cluster.UpdateTime)
	ga.cluster = shieldClusterInfo(cluster)

	if ga.cluster != nil && ga.cluster.NetworkSettings != nil {
		if ga.cluster.NetworkSettings.ServiceIPv4CIDR != "" {
			step, _ := utils.ConvertCIDRToStep(ga.cluster.NetworkSettings.ServiceIPv4CIDR)
			ga.cluster.NetworkSettings.MaxServiceNum = step
		}
		if ga.cluster.NetworkSettings.ClusterIPv4CIDR != "" {
			step, _ := utils.ConvertCIDRToStep(ga.cluster.NetworkSettings.ClusterIPv4CIDR)
			ga.cluster.NetworkSettings.CidrStep = step
		}
	}

	// append apiServer info
	if cluster.ExtraInfo == nil {
		cluster.ExtraInfo = make(map[string]string, 0)
	}
	credential, exist, err := ga.model.GetClusterCredential(ga.ctx, ga.req.ClusterID)
	if err == nil && exist {
		cluster.ExtraInfo[common.ClusterApiServer] = credential.ServerAddress
	}

	// append project code
	if len(cluster.GetProjectID()) > 0 {
		pInfo, errLocal := project.GetProjectManagerClient().GetProjectInfo(ga.ctx, cluster.GetProjectID(), true)
		if errLocal == nil {
			cluster.ExtraInfo[common.ProjectCode] = pInfo.GetProjectCode()
		}
	}

	// sort cluster shared range, current project first
	ga.sortSharedRangeProjectIDorCodes()

	// append module info
	ga.appendModuleInfo()

	return nil
}

func (ga *GetAction) sortSharedRangeProjectIDorCodes() {
	sharedRanges := ga.cluster.GetSharedRanges()
	if sharedRanges == nil {
		return
	}

	if len(sharedRanges.GetProjectIdOrCodes()) == 0 {
		return
	}

	projectIDorCodes := sharedRanges.GetProjectIdOrCodes()
	currentProjectID := ga.cluster.GetProjectID()

	remainProjectIDorCodes := make([]string, 0, len(projectIDorCodes))
	for _, id := range projectIDorCodes {
		if id != currentProjectID {
			remainProjectIDorCodes = append(remainProjectIDorCodes, id)
		}
	}

	sorted := make([]string, 0, 1+len(remainProjectIDorCodes))
	sorted = append(sorted, append([]string{currentProjectID}, remainProjectIDorCodes...)...)

	ga.cluster.SharedRanges.ProjectIdOrCodes = sorted
}

func (ga *GetAction) appendModuleInfo() {
	if ga.cluster.GetClusterBasicSettings().GetModule() == nil {
		ga.cluster.GetClusterBasicSettings().Module = &cmproto.ClusterModule{}
	}

	ctx, err := tenant.WithTenantIdByResourceForContext(ga.ctx,
		tenant.ResourceMetaData{ProjectId: ga.cluster.GetProjectID()})
	if err != nil {
		blog.Errorf("withTenantIdByResourceForContext failed: %s", err.Error())
	}

	// cluster business id
	bkBizID, _ := strconv.Atoi(ga.cluster.GetBusinessID())
	if ga.cluster.GetClusterBasicSettings().GetModule().GetMasterModuleID() != "" {
		bkModuleID, _ := strconv.Atoi(ga.cluster.GetClusterBasicSettings().GetModule().GetMasterModuleID())
		ga.cluster.GetClusterBasicSettings().Module.MasterModuleName = cloudprovider.GetModuleName(ctx,
			bkBizID, bkModuleID)
	}

	if ga.cluster.GetClusterBasicSettings().GetModule().GetWorkerModuleID() != "" {
		bkModuleID, _ := strconv.Atoi(ga.cluster.GetClusterBasicSettings().GetModule().GetWorkerModuleID())
		ga.cluster.GetClusterBasicSettings().Module.WorkerModuleName = cloudprovider.GetModuleName(ctx,
			bkBizID, bkModuleID)

		return
	}

	// compatible with autoscaling config
	autoScalingOption, err := ga.model.GetAutoScalingOption(ga.ctx, ga.req.ClusterID)
	if err == nil && autoScalingOption != nil && autoScalingOption.GetModule().GetScaleOutBizID() != "" &&
		autoScalingOption.GetModule().GetScaleOutModuleID() != "" {
		ga.cluster.GetClusterBasicSettings().Module.WorkerModuleID = autoScalingOption.GetModule().GetScaleOutModuleID()
		ga.cluster.GetClusterBasicSettings().Module.WorkerModuleName = autoScalingOption.GetModule().GetScaleOutModuleName()
		return
	}
}

func (ga *GetAction) updateClusterInfoByCloud() error {
	cloud, err := ga.model.GetCloud(ga.ctx, ga.cluster.Provider)
	if err != nil {
		return err
	}
	ga.cloud = cloud

	if ga.req.CloudInfo && ga.cluster.ClusterType != common.ClusterTypeVirtual && ga.cluster.GetSystemID() != "" {
		cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
			Cloud:     ga.cloud,
			AccountID: ga.cluster.CloudAccountID,
		})
		if err != nil {
			blog.Errorf("get credential for cloudprovider %s/%s updateClusterInfoByCloud failed, %s",
				ga.cloud.CloudID, ga.cloud.CloudProvider, err.Error())
			return err
		}
		cmOption.Region = ga.cluster.Region

		clsMgr, err := cloudprovider.GetClusterMgr(cloud.CloudProvider)
		if err != nil {
			return err
		}
		cluster, err := clsMgr.GetCluster(ga.cluster.SystemID, &cloudprovider.GetClusterOption{
			CommonOption: *cmOption,
			Cluster:      ga.cluster,
		})
		if err != nil {
			return err
		}
		ga.cluster = cluster
	}

	return nil
}

func (ga *GetAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	ga.resp.Data = ga.cluster
	if ga.resp.Extra == nil {
		ga.resp.Extra = &cmproto.ExtraClusterInfo{}
	}
	ga.resp.Extra.ProviderType = ga.cloud.GetEngineType()
}

// Handle get cluster request
func (ga *GetAction) Handle(ctx context.Context, req *cmproto.GetClusterReq, resp *cmproto.GetClusterResp) {
	if req == nil || resp == nil {
		blog.Errorf("get cluster failed, req or resp is empty")
		return
	}
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	if err := ga.validate(); err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ga.getCluster(); err != nil {
		ga.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	// default get clusterInfo by db; if cloudInfo = true, update cluster by cloud
	if err := ga.updateClusterInfoByCloud(); err != nil {
		ga.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// GetNodeAction action for get cluster
type GetNodeAction struct {
	ctx   context.Context // nolint
	model store.ClusterManagerModel
	resp  *cmproto.GetNodeResponse
}

// NewGetNodeAction create get action
func NewGetNodeAction(model store.ClusterManagerModel) *GetNodeAction {
	return &GetNodeAction{
		model: model,
	}
}

func (ga *GetNodeAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle get node request, attention innerIP same in different cluster
func (ga *GetNodeAction) Handle(ctx context.Context, req *cmproto.GetNodeRequest, resp *cmproto.GetNodeResponse) {
	if req == nil || resp == nil {
		blog.Errorf("get node failed, req or resp is empty")
		return
	}
	ga.resp = resp

	if err := req.Validate(); err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	node, err := ga.model.GetNodeByIP(ctx, req.InnerIP)
	if err != nil {
		ga.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	node.Passwd = ""

	resp.Data = append(resp.Data, node)
	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// CheckNodeAction action for check node in cluster
type CheckNodeAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	k8sOp *clusterops.K8SOperator

	req  *cmproto.CheckNodesRequest
	resp *cmproto.CheckNodesResponse

	nodeResult map[string]*cmproto.NodeResult
}

// NewCheckNodeAction create checkNode action
func NewCheckNodeAction(model store.ClusterManagerModel, k8sOp *clusterops.K8SOperator) *CheckNodeAction {
	return &CheckNodeAction{
		model:      model,
		nodeResult: make(map[string]*cmproto.NodeResult),
		k8sOp:      k8sOp,
	}
}

func (ca *CheckNodeAction) validate() error {
	if err := ca.req.Validate(); err != nil {
		return err
	}

	return nil
}

func (ca *CheckNodeAction) checkNodesInCluster() error { // nolint
	if ca.nodeResult == nil {
		ca.nodeResult = make(map[string]*cmproto.NodeResult)
	}
	// get all masterIPs
	masterIPs := GetAllMasterIPs(ca.model)

	var (
		barrier = utils.NewRoutinePool(10)
		lock    = sync.Mutex{}
	)
	defer barrier.Close()

	for i := range ca.req.InnerIPs {
		barrier.Add(1)
		go func(nodeIP string) {
			defer func() {
				barrier.Done()
			}()

			nodeResult, err := ca.getNodeResultByNodeIP(nodeIP, masterIPs)
			if err != nil {
				blog.Errorf("CheckNodeAction getNodeResultByNodeIP failed: %v", err)
				return
			}

			lock.Lock()
			ca.nodeResult[nodeIP] = nodeResult
			lock.Unlock()

		}(ca.req.InnerIPs[i])
	}
	barrier.Wait()

	return nil
}

func (ca *CheckNodeAction) getNodeResultByNodeIP(nodeIP string, masterMapIPs map[string]ClusterInfo) (
	*cmproto.NodeResult, error) {
	nodeResult := &cmproto.NodeResult{
		IsExist:     false,
		ClusterID:   "",
		ClusterName: "",
	}

	// check if exist masterIPs
	if cls, ok := masterMapIPs[nodeIP]; ok {
		nodeResult.IsExist = true
		nodeResult.ClusterID = cls.ClusterID
		nodeResult.ClusterName = cls.ClusterName

		return nodeResult, nil
	}

	// check if exist nodeIPs
	node, err := ca.model.GetNodeByIP(ca.ctx, nodeIP)
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		return nodeResult, err
	}

	if errors.Is(err, drivers.ErrTableRecordNotFound) {
		return nodeResult, nil
	}

	nodeResult.IsExist = true
	nodeResult.ClusterID = node.ClusterID

	// only handle not ca nodes
	if len(node.ClusterID) != 0 && node.NodeGroupID == "" && node.Status == common.StatusRunning {
		cluster, err := ca.model.GetCluster(ca.ctx, node.ClusterID)
		if err != nil {
			return nodeResult, nil
		}
		nodeResult.ClusterName = cluster.GetClusterName()

		// check node exist in cluster
		if cluster.Status == common.StatusDeleted || !ca.checkNodeIPInCluster(node.ClusterID, node.InnerIP) {
			blog.Infof("checkNodeIPInCluster[%s:%s:%s] ip not in cluster", node.ClusterID,
				node.InnerIP, node.NodeID)

			err = ca.model.DeleteClusterNodeByIP(ca.ctx, node.ClusterID, nodeIP)
			if err != nil {
				blog.Errorf("CheckNodeAction[%s] getNodeResultByNodeIP failed: %v", nodeIP, err)
			} else {
				nodeResult.IsExist = false
			}
		}
	}

	return nodeResult, nil
}

func (ca *CheckNodeAction) setResp(code uint32, msg string) {
	ca.resp.Code = code
	ca.resp.Message = msg
	ca.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	if ca.resp.Data == nil {
		ca.resp.Data = make(map[string]*cmproto.NodeResult)
	}
	ca.resp.Data = ca.nodeResult
}

func (ca *CheckNodeAction) checkNodeIPInCluster(clusterID string, nodeIP string) bool {
	ctx, cancel := context.WithTimeout(ca.ctx, time.Second*5)
	defer cancel()

	_, err := ca.k8sOp.GetClusterNode(ctx, clusterops.QueryNodeOption{
		ClusterID: clusterID,
		NodeIP:    nodeIP,
	})
	if err != nil && strings.Contains(err.Error(), "not found") {
		blog.Errorf("CheckNodeAction[%s:%s] nodeIPInCluster %v", clusterID, nodeIP, err)
		return false
	}

	if err == nil {
		blog.Infof("CheckNodeAction[%s:%s] nodeIPInCluster", clusterID, nodeIP)
		return true
	}

	// other unknown errors
	return true
}

// Handle handles check nodes in cluster request
func (ca *CheckNodeAction) Handle(ctx context.Context, req *cmproto.CheckNodesRequest,
	resp *cmproto.CheckNodesResponse) {
	if req == nil || resp == nil {
		blog.Errorf("check cluster node failed, req or resp is empty")
		return
	}
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp

	if err := ca.validate(); err != nil {
		ca.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ca.checkNodesInCluster(); err != nil {
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ca.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// GetNodeInfoAction action for get cluster
type GetNodeInfoAction struct {
	ctx   context.Context
	model store.ClusterManagerModel

	req  *cmproto.GetNodeInfoRequest
	resp *cmproto.GetNodeInfoResponse
}

// NewGetNodeInfoAction create get action
func NewGetNodeInfoAction(model store.ClusterManagerModel) *GetNodeInfoAction {
	return &GetNodeInfoAction{
		model: model,
	}
}

func (ga *GetNodeInfoAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ga *GetNodeInfoAction) getNodeInfoByIP() error {
	node, err := ga.model.GetNodeByIP(ga.ctx, ga.req.InnerIP)
	if err != nil {
		return err
	}

	ga.resp.Data = &cmproto.NodeInfo{
		NodeName:       "",
		NodeType:       "",
		NodeID:         node.NodeID,
		InnerIP:        node.InnerIP,
		ClusterID:      node.ClusterID,
		VPC:            node.VPC,
		Region:         node.Region,
		DeviceID:       node.DeviceID,
		Status:         node.Status,
		InstanceConfig: nil,
		ZoneInfo: &cmproto.ZoneInfo{
			ZoneID: fmt.Sprintf("%d", node.Zone),
			Zone:   node.ZoneID,
		},
	}

	if len(node.NodeTemplateID) > 0 {
		template, err := ga.model.GetNodeTemplateByID(ga.ctx, node.NodeTemplateID)
		if err != nil {
			blog.Errorf("GetNodeInfoAction GetNodeTemplateByID[%s] failed: %v", node.NodeTemplateID, err)
			return err
		}

		ga.resp.Data.NodeTemplate = template
	}
	if len(node.NodeGroupID) > 0 {
		group, err := ga.model.GetNodeGroup(ga.ctx, node.NodeGroupID)
		if err != nil {
			blog.Errorf("GetNodeInfoAction GetNodeGroup[%s] failed: %v", node.NodeGroupID, err)
			return err
		}

		ga.resp.Data.Group = group
	}

	return nil
}

// Handle get nodeInfo request, attention innerIP same in different cluster
func (ga *GetNodeInfoAction) Handle(ctx context.Context, req *cmproto.GetNodeInfoRequest,
	resp *cmproto.GetNodeInfoResponse) {
	if req == nil || resp == nil {
		blog.Errorf("get nodeInfo failed, req or resp is empty")
		return
	}
	ga.req = req
	ga.resp = resp

	if err := req.Validate(); err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	err := ga.getNodeInfoByIP()
	if err != nil {
		ga.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// GetClustersMetaDataAction action for get cluster meta
type GetClustersMetaDataAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.GetClustersMetaDataRequest
	resp  *cmproto.GetClustersMetaDataResponse
	k8sOp *clusterops.K8SOperator

	clustersMeta []*cmproto.ClusterMeta
}

// NewGetClustersMetaDataAction get clusters meta action
func NewGetClustersMetaDataAction(model store.ClusterManagerModel,
	k8sOp *clusterops.K8SOperator) *GetClustersMetaDataAction {
	return &GetClustersMetaDataAction{
		model: model,
		k8sOp: k8sOp,
	}
}

func (ga *GetClustersMetaDataAction) getClusterNodeNum(clusterId string) uint32 {
	listNodesAction := NewListNodesInClusterAction(ga.model, ga.k8sOp)

	var (
		listNodesReq = &cmproto.ListNodesInClusterRequest{
			ClusterID: clusterId,
		}
		listNodesResp = &cmproto.ListNodesInClusterResponse{}
	)

	// list nodes
	listNodesAction.Handle(ga.ctx, listNodesReq, listNodesResp)

	if listNodesResp.GetCode() != 0 || !listNodesResp.GetResult() {
		blog.Errorf("GetClustersMetaDataAction[%s] getClusterNodeNum failed: %v",
			clusterId, listNodesResp.GetMessage())
		return 0
	}

	return uint32(len(listNodesResp.GetData()))
}

func (ga *GetClustersMetaDataAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)

	ga.resp.Data = ga.clustersMeta
}

func (ga *GetClustersMetaDataAction) validate() error {
	return ga.req.Validate()
}

func (ga *GetClustersMetaDataAction) getClustersMeta() {
	clusterIds := ga.req.GetClusters()

	var (
		lock = sync.Mutex{}
	)

	if ga.clustersMeta == nil {
		ga.clustersMeta = make([]*cmproto.ClusterMeta, 0)
	}
	concurency := utils.NewRoutinePool(20)
	defer concurency.Close()

	for i := range clusterIds {
		concurency.Add(1)
		go func(clusterId string) {
			defer utils.RecoverPrintStack("GetClustersMetaDataAction")
			defer concurency.Done()
			clusterMeta := &cmproto.ClusterMeta{
				ClusterId: clusterId,
			}

			nodeNum := ga.getClusterNodeNum(clusterId)
			clusterMeta.ClusterNodeNum = nodeNum

			lock.Lock()
			defer lock.Unlock()
			ga.clustersMeta = append(ga.clustersMeta, clusterMeta)
		}(clusterIds[i])
	}
	concurency.Wait()
}

// Handle delete cluster nodes request
func (ga *GetClustersMetaDataAction) Handle(ctx context.Context, req *cmproto.GetClustersMetaDataRequest,
	resp *cmproto.GetClustersMetaDataResponse) {
	if req == nil || resp == nil {
		blog.Errorf("get clusters meta failed, req or resp is empty")
		return
	}
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	// check request parameter validate
	err := ga.validate()
	if err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	ga.getClustersMeta()

	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// GetClusterSharedProjectAction action for get cluster project info
type GetClusterSharedProjectAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.GetClusterSharedProjectRequest
	resp  *cmproto.GetClusterSharedProjectResponse
}

// NewGetClusterSharedProjectAction get clusters cluster info action
func NewGetClusterSharedProjectAction(model store.ClusterManagerModel) *GetClusterSharedProjectAction {
	return &GetClusterSharedProjectAction{
		model: model,
	}
}

func (ga *GetClusterSharedProjectAction) validate() error {
	return ga.req.Validate()
}

func (ga *GetClusterSharedProjectAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle get cluster shared project request
func (ga *GetClusterSharedProjectAction) Handle(ctx context.Context, req *cmproto.GetClusterSharedProjectRequest,
	resp *cmproto.GetClusterSharedProjectResponse) {
	if req == nil || resp == nil {
		blog.Errorf("get cluster project info failed, req or resp is empty")
		return
	}
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	// check request parameter validate
	err := ga.validate()
	if err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	err = ga.getSharedProject()
	if err != nil {
		ga.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

func (ga *GetClusterSharedProjectAction) getSharedProject() error {
	cluster, err := ga.model.GetCluster(ga.ctx, ga.req.ClusterID)
	if err != nil {
		return err
	}

	sharedRanges := cluster.GetSharedRanges()
	var projectIDorCodes []string
	if sharedRanges != nil {
		projectIDorCodes = sharedRanges.GetProjectIdOrCodes()
	}

	ga.resp.Data = &spb.ListValue{
		Values: make([]*spb.Value, 0),
	}

	// if projectIDorCodes is empty, use cluster's projectID for sharedproject
	// only one value not use goroutine
	if len(projectIDorCodes) == 0 {
		projectID := cluster.GetProjectID()
		if projectID != "" {
			pInfo, err := project.GetProjectManagerClient().GetProjectInfo(ga.ctx, projectID, true)
			if err != nil {
				blog.Errorf("get project info by project manager client failed, %s", err.Error())
				return err
			}
			result, err := utils.MarshalInterfaceToValue(pInfo)
			if err != nil {
				blog.Errorf("marshal projectGroupsQuotaData err, %s", err.Error())
				return err
			}
			ga.resp.Data.Values = append(
				ga.resp.Data.Values,
				spb.NewStructValue(result),
			)
		}
		return nil
	}

	var (
		lock = sync.Mutex{}
	)

	// projectIDorCodes is not empty and have more than one value
	barrier := utils.NewRoutinePool(20)
	defer barrier.Close()

	for i := range projectIDorCodes {
		barrier.Add(1)
		go func(projectIDorCode string) {
			defer utils.RecoverPrintStack("GetClusterSharedProjectAction")
			defer barrier.Done()

			pInfo, err := project.GetProjectManagerClient().GetProjectInfo(ga.ctx, projectIDorCode, true)
			if err != nil {
				blog.Errorf("get project info by project manager client failed, %s", err.Error())
				return
			}
			result, err := utils.MarshalInterfaceToValue(pInfo)
			if err != nil {
				blog.Errorf("marshal projectGroupsQuotaData err, %s", err.Error())
				return
			}

			lock.Lock()
			defer lock.Unlock()
			ga.resp.Data.Values = append(
				ga.resp.Data.Values,
				spb.NewStructValue(result),
			)
		}(projectIDorCodes[i])
	}
	barrier.Wait()

	return nil
}
