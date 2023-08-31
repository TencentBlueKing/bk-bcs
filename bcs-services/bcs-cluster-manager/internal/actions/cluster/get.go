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
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
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
	ga.cluster = shieldClusterInfo(cluster)
	return nil
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

	return
}

// GetNodeAction action for get cluster
type GetNodeAction struct {
	ctx   context.Context
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

	return
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

func (ca *CheckNodeAction) checkNodesInCluster() error {
	if ca.nodeResult == nil {
		ca.nodeResult = make(map[string]*cmproto.NodeResult)
	}
	// get all masterIPs
	masterIPs := getAllMasterIPs(ca.model)

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

func (ca *CheckNodeAction) getNodeResultByNodeIP(nodeIP string, masterMapIPs map[string]clusterInfo) (
	*cmproto.NodeResult, error) {
	nodeResult := &cmproto.NodeResult{
		IsExist:     false,
		ClusterID:   "",
		ClusterName: "",
	}

	// check if exist masterIPs
	if cls, ok := masterMapIPs[nodeIP]; ok {
		nodeResult.IsExist = true
		nodeResult.ClusterID = cls.clusterID
		nodeResult.ClusterName = cls.clusterName

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
	if len(node.ClusterID) != 0 && node.NodeGroupID == "" {
		cluster, err := ca.model.GetCluster(ca.ctx, node.ClusterID)
		if err == nil {
			nodeResult.ClusterName = cluster.GetClusterName()
		}

		// check node exist in cluster
		if cluster.Status == common.StatusDeleted || !ca.checkNodeIPInCluster(node.ClusterID, node.InnerIP) {
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
	return
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

	return
}
