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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	spb "google.golang.org/protobuf/types/known/structpb"
	corev1 "k8s.io/api/core/v1"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

// ListAction list action for cluster
type ListAction struct {
	ctx         context.Context
	model       store.ClusterManagerModel
	iam         iam.PermClient
	req         *cmproto.ListClusterReq
	resp        *cmproto.ListClusterResp
	clusterList []*cmproto.Cluster
}

// NewListAction create list action for cluster
func NewListAction(model store.ClusterManagerModel, iam iam.PermClient) *ListAction {
	return &ListAction{
		model: model,
		iam:   iam,
	}
}

func (la *ListAction) validate() error {
	if err := la.req.Validate(); err != nil {
		return err
	}

	if len(la.req.Environment) > 0 {
		_, ok := EnvironmentLookup[la.req.Environment]
		if !ok {
			return fmt.Errorf("request Environment invalid, must be [test/debug/prod]")
		}
	}

	if len(la.req.EngineType) > 0 {
		_, ok := EngineTypeLookup[la.req.EngineType]
		if !ok {
			return fmt.Errorf("request EngineType invalid, must be [k8s/mesos]")
		}
	}

	if len(la.req.ClusterType) > 0 {
		_, ok := ClusterTypeLookup[la.req.ClusterType]
		if !ok {
			return fmt.Errorf("request ClusterType invalid, must be [federation/signal]")
		}
	}

	return nil
}

func (la *ListAction) getSharedCluster() error {
	condM := make(operator.M)
	condM["isshared"] = true
	condCluster := operator.NewLeafCondition(operator.Eq, condM)
	condStatus := operator.NewLeafCondition(operator.Ne, operator.M{"status": common.StatusDeleted})

	branchCond := operator.NewBranchCondition(operator.And, condCluster, condStatus)
	clusterList, err := la.model.ListCluster(la.ctx, branchCond, &storeopt.ListOption{})
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		return err
	}

	clusterIDs := make([]string, 0)
	for i := range clusterList {
		la.clusterList = append(la.clusterList, shieldClusterInfo(&clusterList[i]))
		clusterIDs = append(clusterIDs, clusterList[i].ClusterID)
	}

	if la.resp.WebAnnotations == nil {
		la.resp.WebAnnotations = &cmproto.WebAnnotations{
			Perms: make(map[string]*spb.Struct),
		}
	} else {
		if la.resp.WebAnnotations.Perms == nil {
			la.resp.WebAnnotations.Perms = make(map[string]*spb.Struct)
		}
	}

	for _, clusterID := range clusterIDs {
		if _, ok := la.resp.WebAnnotations.Perms[clusterID]; !ok {
			actionPerm, err := spb.NewStruct(auth.GetV3SharedClusterPerm())
			if err != nil {
				continue
			}
			la.resp.WebAnnotations.Perms[clusterID] = actionPerm
		}
	}

	return nil
}

func (la *ListAction) listCluster() error {
	getSharedCluster := true
	condM := make(operator.M)
	if len(la.req.ClusterName) != 0 {
		condM["clustername"] = la.req.ClusterName
	}
	if len(la.req.Provider) != 0 {
		condM["provider"] = la.req.Provider
	}
	if len(la.req.Region) != 0 {
		condM["region"] = la.req.Region
	}
	if len(la.req.VpcID) != 0 {
		condM["vpcid"] = la.req.VpcID
	}
	if len(la.req.ProjectID) != 0 {
		condM["projectid"] = la.req.ProjectID
	}
	if len(la.req.BusinessID) != 0 {
		condM["businessid"] = la.req.BusinessID
	}
	if len(la.req.Environment) != 0 {
		condM["environment"] = la.req.Environment
	}
	if len(la.req.EngineType) != 0 {
		condM["enginetype"] = la.req.EngineType
	}
	if len(la.req.SystemID) != 0 {
		condM["systemid"] = la.req.SystemID
	}
	if len(la.req.ExtraClusterID) != 0 {
		condM["extraclusterid"] = la.req.ExtraClusterID
		getSharedCluster = false
	}
	if len(la.req.ClusterID) != 0 {
		condM["clusterid"] = la.req.ClusterID
	}

	condCluster := operator.NewLeafCondition(operator.Eq, condM)
	condStatus := operator.NewLeafCondition(operator.Ne, operator.M{"status": common.StatusDeleted})

	branchCond := operator.NewBranchCondition(operator.And, condCluster, condStatus)
	clusterList, err := la.model.ListCluster(la.ctx, branchCond, &storeopt.ListOption{})
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		return err
	}

	clusterIDList := make([]string, 0)
	for i := range clusterList {
		if clusterList[i].IsShared {
			clusterList[i].IsShared = false
		}
		la.clusterList = append(la.clusterList, shieldClusterInfo(&clusterList[i]))
		clusterIDList = append(clusterIDList, clusterList[i].ClusterID)
	}

	// return cluster extraInfo
	la.returnClusterExtraInfo(clusterList)

	// projectID / operator get user perm
	if la.req.ProjectID != "" && la.req.Operator != "" {
		v3Perm, err := la.GetProjectClustersV3Perm(actions.PermInfo{
			ProjectID: la.req.ProjectID,
			UserID:    la.req.Operator,
		}, clusterIDList)
		if err != nil {
			blog.Errorf("listCluster GetUserPermListByProjectAndCluster failed: %v", err.Error())
		}
		la.resp.WebAnnotations = &cmproto.WebAnnotations{
			Perms: v3Perm,
		}
	}

	// default return shared cluster
	if getSharedCluster {
		err = la.getSharedCluster()
		if err != nil {
			blog.Errorf("ListCluster getSharedCluster failed: %v", err)
			return err
		}
	}

	return nil
}

// GetProjectClustersV3Perm get iam v3 perm
func (la *ListAction) GetProjectClustersV3Perm(user actions.PermInfo, clusterList []string) (map[string]*spb.Struct,
	error) {
	var (
		v3Perm map[string]map[string]interface{}
		err    error
	)

	v3Perm, err = la.getUserClusterPermList(user, clusterList)
	if err != nil {
		blog.Errorf("listCluster GetUserClusterPermList failed: %v", err.Error())
		return nil, err
	}

	// trans result for adapt front
	v3ResultPerm := make(map[string]*spb.Struct)
	for clsID := range v3Perm {
		actionPerm, err := spb.NewStruct(v3Perm[clsID])
		if err != nil {
			return nil, err
		}

		v3ResultPerm[clsID] = actionPerm
	}

	return v3ResultPerm, nil
}

func (la *ListAction) getUserClusterPermList(user actions.PermInfo, clusterList []string) (
	map[string]map[string]interface{}, error) {
	permissions := make(map[string]map[string]interface{})
	clusterPerm := cluster.NewBCSClusterPermClient(la.iam)

	actionIDs := []string{cluster.ClusterView.String(), cluster.ClusterManage.String(), cluster.ClusterDelete.String()}
	perms, err := clusterPerm.GetMultiClusterMultiActionPermission(user.UserID, user.ProjectID, clusterList, actionIDs)
	if err != nil {
		return nil, err
	}

	for clusterID, perm := range perms {
		if permissions[clusterID] == nil {
			permissions[clusterID] = make(map[string]interface{})
		}
		for action, res := range perm {
			permissions[clusterID][action] = res
		}
	}

	return permissions, nil
}

// GetCloudProviderEngine get cloud engineType
func (la *ListAction) GetCloudProviderEngine(cls cmproto.Cluster) string {
	cloud, err := la.model.GetCloud(la.ctx, cls.Provider)
	if err != nil {
		blog.Errorf("listCluster GetCloudProviderEngine failed: %v", err)
		return ""
	}

	return cloud.GetEngineType()
}

func (la *ListAction) returnClusterExtraInfo(clusterList []cmproto.Cluster) {
	if la.resp.ClusterExtraInfo == nil {
		la.resp.ClusterExtraInfo = make(map[string]*cmproto.ExtraInfo)
	}

	// cluster extra info
	for i := range clusterList {
		la.resp.ClusterExtraInfo[clusterList[i].ClusterID] = &cmproto.ExtraInfo{
			CanDeleted:   true,
			ProviderType: la.GetCloudProviderEngine(clusterList[i]),
		}
	}

	return
}

func (la *ListAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.clusterList
}

// Handle handle list cluster request
func (la *ListAction) Handle(ctx context.Context, req *cmproto.ListClusterReq, resp *cmproto.ListClusterResp) {
	if req == nil || resp == nil {
		blog.Errorf("list cluster failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listCluster(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}

// ListCommonClusterAction list action for cluster
type ListCommonClusterAction struct {
	ctx         context.Context
	model       store.ClusterManagerModel
	req         *cmproto.ListCommonClusterReq
	resp        *cmproto.ListCommonClusterResp
	clusterList []*cmproto.Cluster
}

// NewListCommonClusterAction create list action for cluster
func NewListCommonClusterAction(model store.ClusterManagerModel) *ListCommonClusterAction {
	return &ListCommonClusterAction{
		model: model,
	}
}

func (la *ListCommonClusterAction) validate() error {
	if err := la.req.Validate(); err != nil {
		return err
	}

	return nil
}

func (la *ListCommonClusterAction) listCluster() error {
	condCluster := operator.NewLeafCondition(operator.Eq, operator.M{
		"isshared": true,
	})
	condStatus := operator.NewLeafCondition(operator.Ne, operator.M{"status": common.StatusDeleted})
	branchCond := operator.NewBranchCondition(operator.And, condCluster, condStatus)

	clusterList, err := la.model.ListCluster(la.ctx, branchCond, &storeopt.ListOption{})
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		return err
	}

	clusterIDList := make([]string, 0)
	for i := range clusterList {
		la.clusterList = append(la.clusterList, shieldClusterInfo(&clusterList[i]))
		clusterIDList = append(clusterIDList, clusterList[i].ClusterID)
	}

	// get common cluster permission
	if len(clusterIDList) > 0 {
		v3Perm, err := GetProjectCommonClustersPerm(clusterIDList)
		if err != nil {
			blog.Errorf("listCluster GetUserPermListByProjectAndCluster failed: %v", err)
		}
		la.resp.WebAnnotations = &cmproto.WebAnnotations{
			Perms: v3Perm,
		}
	}

	return nil
}

func (la *ListCommonClusterAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.clusterList
}

// Handle handle list common cluster request
func (la *ListCommonClusterAction) Handle(ctx context.Context,
	req *cmproto.ListCommonClusterReq, resp *cmproto.ListCommonClusterResp) {
	if req == nil || resp == nil {
		blog.Errorf("list common cluster failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listCluster(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}

// ListNodesInClusterAction list action for cluster
type ListNodesInClusterAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.ListNodesInClusterRequest
	resp  *cmproto.ListNodesInClusterResponse
	k8sOp *clusterops.K8SOperator
	nodes []*cmproto.ClusterNode
}

// NewListNodesInClusterAction create list action for cluster
func NewListNodesInClusterAction(model store.ClusterManagerModel,
	k8sOp *clusterops.K8SOperator) *ListNodesInClusterAction {
	return &ListNodesInClusterAction{
		model: model,
		k8sOp: k8sOp,
	}
}

func (la *ListNodesInClusterAction) validate() error {
	if err := la.req.Validate(); err != nil {
		return err
	}
	return nil
}

func (la *ListNodesInClusterAction) listNodes() error {
	condM := make(operator.M)
	if len(la.req.ClusterID) != 0 {
		condM["clusterid"] = la.req.ClusterID
	}
	if len(la.req.Region) != 0 {
		condM["region"] = la.req.Region
	}
	if len(la.req.VpcID) != 0 {
		condM["vpcid"] = la.req.VpcID
	}
	if len(la.req.NodeGroupID) != 0 {
		condM["nodegroupid"] = la.req.NodeGroupID
	}
	if len(la.req.InstanceType) != 0 {
		condM["instancetype"] = la.req.InstanceType
	}
	if len(la.req.Status) != 0 {
		condM["status"] = la.req.Status
	}
	cond := operator.NewLeafCondition(operator.Eq, condM)
	nodes, err := la.model.ListNode(la.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		blog.Errorf("list nodes in cluster %s failed, %s", la.req.ClusterID, err.Error())
		return err
	}

	if !la.req.ShowPwd {
		removeNodeSensitiveInfo(nodes)
	}
	cmNodes := make([]*cmproto.ClusterNode, 0)
	for i := range nodes {
		cmNodes = append(cmNodes, transNodeToClusterNode(la.model, nodes[i]))
	}

	k8sNodes := filterNodesRole(la.getK8sNodes(), false)
	la.nodes = mergeClusterNodes(la.req.ClusterID, cmNodes, k8sNodes)

	return nil
}

func (la *ListNodesInClusterAction) getK8sNodes() []*corev1.Node {
	k8sNodes, err := la.k8sOp.ListClusterNodes(la.ctx, la.req.ClusterID)
	if err != nil {
		blog.Warnf("ListClusterNodes %s failed, %s", la.req.ClusterID, err.Error())
		return nil
	}
	return k8sNodes
}

func (la *ListNodesInClusterAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.nodes
}

// Handle handle list cluster request
func (la *ListNodesInClusterAction) Handle(ctx context.Context,
	req *cmproto.ListNodesInClusterRequest, resp *cmproto.ListNodesInClusterResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list cluster nodes failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listNodes(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}

// ListMastersInClusterAction list action for cluster
type ListMastersInClusterAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.ListMastersInClusterRequest
	resp  *cmproto.ListMastersInClusterResponse
	k8sOp *clusterops.K8SOperator
	nodes []*cmproto.ClusterNode
}

// NewListMastersInClusterAction create list action for cluster
func NewListMastersInClusterAction(model store.ClusterManagerModel,
	k8sOp *clusterops.K8SOperator) *ListMastersInClusterAction {
	return &ListMastersInClusterAction{
		model: model,
		k8sOp: k8sOp,
	}
}

func (la *ListMastersInClusterAction) validate() error {
	if err := la.req.Validate(); err != nil {
		return err
	}
	return nil
}

func (la *ListMastersInClusterAction) listNodes() error {
	_, err := la.model.GetCluster(la.ctx, la.req.ClusterID)
	if err != nil {
		blog.Errorf("get cluster %s failed, %s", la.req.ClusterID, err.Error())
		return err
	}
	masters, err := la.k8sOp.ListClusterNodes(la.ctx, la.req.ClusterID)
	if err != nil {
		blog.Warnf("ListClusterNodes %s failed, %s", la.req.ClusterID, err.Error())
		return err
	}

	masters = filterNodesRole(masters, true)
	la.nodes = transK8sNodesToClusterNodes(la.req.ClusterID, masters)

	la.appendHostInfo()
	return nil
}

func (la *ListMastersInClusterAction) appendHostInfo() {
	ips := make([]string, 0)
	for _, v := range la.nodes {
		ips = append(ips, v.InnerIP)
	}
	if len(ips) == 0 {
		return
	}

	cmdbClient := cmdb.GetCmdbClient()
	hosts, err := cmdbClient.QueryAllHostInfoWithoutBiz(ips)
	if err != nil {
		blog.Warnf("GetHostInfo for %s failed, %s", la.req.ClusterID, err.Error())
		return
	}

	for i := range la.nodes {
		for _, v := range hosts {
			if v.BKHostInnerIP == la.nodes[i].InnerIP {
				la.nodes[i].Idc = v.IDCName
				la.nodes[i].Rack = v.Rack
				la.nodes[i].DeviceClass = v.SCMDeviceType
			}
		}
	}
}

func (la *ListMastersInClusterAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.nodes
}

// Handle handle list cluster request
func (la *ListMastersInClusterAction) Handle(ctx context.Context,
	req *cmproto.ListMastersInClusterRequest, resp *cmproto.ListMastersInClusterResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list cluster masters failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listNodes(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
