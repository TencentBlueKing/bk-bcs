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
	"sort"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	spb "google.golang.org/protobuf/types/known/structpb"
	corev1 "k8s.io/api/core/v1"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	autils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/utils"
	iauth "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/gse"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
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

// validate request validation
func (la *ListAction) validate() error {
	if err := la.req.Validate(); err != nil {
		return err
	}

	// env
	if len(la.req.Environment) > 0 {
		_, ok := common.EnvironmentLookup[la.req.Environment]
		if !ok {
			return fmt.Errorf("request Environment invalid, must be [test/debug/prod]")
		}
	}

	// engineType
	if len(la.req.EngineType) > 0 {
		_, ok := common.EngineTypeLookup[la.req.EngineType]
		if !ok {
			return fmt.Errorf("request EngineType invalid, must be [k8s/mesos]")
		}
	}

	// clusterType
	if len(la.req.ClusterType) > 0 {
		_, ok := common.ClusterTypeLookup[la.req.ClusterType]
		if !ok {
			return fmt.Errorf("request ClusterType invalid, must be [federation/signal]")
		}
	}

	return nil
}

// getSharedCluster shared cluster
func (la *ListAction) getSharedCluster() error {
	conds := make([]*operator.Condition, 0)

	condM := make(operator.M)
	condM["isshared"] = true
	condCluster := operator.NewLeafCondition(operator.Eq, condM)
	conds = append(conds, condCluster)

	if !la.req.GetAll() {
		condStatus := operator.NewLeafCondition(operator.Ne, operator.M{"status": common.StatusDeleted})
		conds = append(conds, condStatus)
	}

	branchCond := operator.NewBranchCondition(operator.And, conds...)
	clusterList, err := la.model.ListCluster(la.ctx, branchCond, &storeopt.ListOption{})
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		return err
	}

	clusterIDs := make([]string, 0)
	for i := range clusterList {
		if clusterList[i].GetProjectID() == la.req.ProjectID {
			continue
		}
		la.clusterList = append(la.clusterList, shieldClusterInfo(clusterList[i]))
		clusterIDs = append(clusterIDs, clusterList[i].ClusterID)
	}

	// set webAnnotations
	if la.resp.WebAnnotations == nil {
		la.resp.WebAnnotations = &cmproto.WebAnnotations{
			Perms: make(map[string]*spb.Struct),
		}
	} else if la.resp.WebAnnotations.Perms == nil {
		la.resp.WebAnnotations.Perms = make(map[string]*spb.Struct)
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

func (la *ListAction) filterClusterList() ([]*cmproto.Cluster, bool, error) {
	var (
		sharedCluster = true
		clusterList   []*cmproto.Cluster
		err           error
	)

	conds := make([]*operator.Condition, 0)

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
		sharedCluster = false
	}
	if len(la.req.ClusterID) != 0 {
		condM["clusterid"] = la.req.ClusterID
	}
	condCluster := operator.NewLeafCondition(operator.Eq, condM)
	conds = append(conds, condCluster)

	if !la.req.All {
		condStatus := operator.NewLeafCondition(operator.Ne, operator.M{"status": common.StatusDeleted})
		conds = append(conds, condStatus)
	}
	branchCond := operator.NewBranchCondition(operator.And, conds...)

	clusterList, err = la.model.ListCluster(la.ctx, branchCond, &storeopt.ListOption{})
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		return nil, sharedCluster, err
	}

	return clusterList, sharedCluster, nil
}

// listCluster cluster list
func (la *ListAction) listCluster() error {
	clusterList, sharedCluster, err := la.filterClusterList()
	if err != nil {
		return err
	}

	clusterIDList := make([]string, 0)
	for i := range clusterList {
		if clusterList[i].IsShared {
			clusterList[i].IsShared = false
		}
		la.clusterList = append(la.clusterList, shieldClusterInfo(clusterList[i]))
		clusterIDList = append(clusterIDList, clusterList[i].ClusterID)
	}

	// return cluster extraInfo
	la.resp.ClusterExtraInfo = returnClusterExtraInfo(la.model, clusterList)

	// projectID / operator get user perm
	if la.req.ProjectID != "" && la.req.Operator != "" {
		v3Perm, err := la.GetProjectClustersV3Perm(actions.PermInfo{ // nolint
			ProjectID: la.req.ProjectID,
			UserID:    la.req.Operator,
		}, clusterIDList)
		if err != nil {
			blog.Errorf("listCluster GetUserPermListByProjectAndCluster failed: %v", err.Error())
			return err
		}
		la.resp.WebAnnotations = &cmproto.WebAnnotations{
			Perms: v3Perm,
		}
	}

	// default return shared cluster
	if sharedCluster {
		err = la.getSharedCluster()
		if err != nil {
			blog.Errorf("ListCluster getSharedCluster failed: %v", err)
			return err
		}
	}

	return nil
}

// GetProjectClustersV3Perm get iam v3 perm
func (la *ListAction) GetProjectClustersV3Perm(user actions.PermInfo, clusterList []string) (
	map[string]*spb.Struct, error) {
	var (
		v3Perm map[string]map[string]interface{}
		err    error
	)

	// get user clusterList perms
	v3Perm, err = getUserClusterPermList(la.iam, user, clusterList)
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

// setResp resp body
func (la *ListAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.clusterList
}

// Handle list cluster request
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
}

// ListProjectClusterAction list action for project clusters
type ListProjectClusterAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	iam   iam.PermClient

	req         *cmproto.ListProjectClusterReq
	resp        *cmproto.ListProjectClusterResp
	clusterList []*cmproto.Cluster
}

// NewListProjectClusterAction create list action for project cluster
func NewListProjectClusterAction(model store.ClusterManagerModel, iam iam.PermClient) *ListProjectClusterAction {
	return &ListProjectClusterAction{
		model: model,
		iam:   iam,
	}
}

func (la *ListProjectClusterAction) validate() error {
	if err := la.req.Validate(); err != nil {
		return err
	}

	return nil
}

// listProjectCluster get project clusters
func (la *ListProjectClusterAction) listProjectCluster() error {
	condM := make(operator.M)

	if len(la.req.ProjectID) != 0 {
		condM["projectid"] = la.req.ProjectID
	}
	if len(la.req.Provider) != 0 {
		condM["provider"] = la.req.Provider
	}
	if len(la.req.Region) != 0 {
		condM["region"] = la.req.Region
	}

	condCluster := operator.NewLeafCondition(operator.Eq, condM)
	condStatus := operator.NewLeafCondition(operator.Ne, operator.M{"status": common.StatusDeleted})

	branchCond := operator.NewBranchCondition(operator.And, condCluster, condStatus)
	clusterList, err := la.model.ListCluster(la.ctx, branchCond, &storeopt.ListOption{})
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		blog.Errorf("ListProjectClusterAction ListCluster failed: %v", err)
		return err
	}

	// cluster sort
	var (
		otherCluster   = make([]*cmproto.Cluster, 0)
		runningCluster = make([]*cmproto.Cluster, 0)
		clusterIDList  = make([]string, 0)
	)
	for i := range clusterList {
		if clusterList[i].IsShared {
			clusterList[i].IsShared = false
		}

		if clusterList[i].Status == common.StatusRunning {
			runningCluster = append(runningCluster, shieldClusterInfo(clusterList[i]))
		} else {
			otherCluster = append(otherCluster, shieldClusterInfo(clusterList[i]))
		}
		clusterIDList = append(clusterIDList, clusterList[i].ClusterID)
	}
	if len(otherCluster) > 0 {
		sort.Sort(utils.ClusterSlice(otherCluster))
	}
	if len(runningCluster) > 0 {
		sort.Sort(utils.ClusterSlice(runningCluster))
	}
	la.clusterList = append(la.clusterList, otherCluster...)
	la.clusterList = append(la.clusterList, runningCluster...)

	// return cluster extraInfo
	la.resp.ClusterExtraInfo = returnClusterExtraInfo(la.model, clusterList)

	// get shared cluster
	sharedClusters, err := getSharedCluster(la.req.ProjectID, la.req.GetBizId(), la.model)
	if err != nil {
		blog.Errorf("ListProjectClusterAction getSharedCluster failed: %v", err)
	} else {
		la.clusterList = append(la.clusterList, sharedClusters...)
	}

	// return project user cluster perms & shared cluster perms
	la.resp.WebAnnotations = la.getWebAnnotations(la.req.ProjectID, clusterIDList, sharedClusters)

	return nil
}

// getWebAnnotations get cluster perms
func (la *ListProjectClusterAction) getWebAnnotations(projectID string, clusterIDs []string,
	sharedClusters []*cmproto.Cluster) *cmproto.WebAnnotationsV2 {
	username := iauth.GetUserFromCtx(la.ctx)

	blog.Infof("ListProjectClusterAction GetWebAnnotations user[%s]", username)
	// default use request operator
	if la.req.Operator == "" {
		la.req.Operator = username
	}

	perms := make(map[string]map[string]interface{}, 0)

	// shared cluster perms
	sharedClusterIDs := make([]string, 0)
	for i := range sharedClusters {
		sharedClusterIDs = append(sharedClusterIDs, sharedClusters[i].ClusterID)
	}
	// shared cluster perms
	for _, clusterID := range sharedClusterIDs {
		if _, ok := perms[clusterID]; !ok {
			perms[clusterID] = auth.GetV3SharedClusterPerm()
		}
	}

	signalPerms, err := getUserClusterPermList(la.iam, actions.PermInfo{
		ProjectID: projectID,
		UserID:    la.req.Operator,
	}, clusterIDs)
	if err != nil {
		blog.Errorf("ListProjectClusterAction GetWebAnnotations user %s cluster perms failed, err: %s",
			username, err.Error())
	}
	for id := range signalPerms {
		perms[id] = signalPerms[id]
	}

	// marshal data
	s, err := utils.MarshalInterfaceToValue(perms)
	if err != nil {
		blog.Errorf("MarshalInterfaceToValue failed, perms %v, err: %s", perms, err.Error())
		return nil
	}
	webAnnotations := &cmproto.WebAnnotationsV2{
		Perms: s,
	}

	return webAnnotations
}

// GetProjectClustersV3Perm get iam v3 perm
func (la *ListProjectClusterAction) GetProjectClustersV3Perm(user actions.PermInfo,
	clusterList []string) (map[string]*spb.Struct, error) {
	var (
		v3Perm map[string]map[string]interface{}
		err    error
	)

	// get user perms
	v3Perm, err = getUserClusterPermList(la.iam, user, clusterList)
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

func (la *ListProjectClusterAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.clusterList
}

// Handle list project cluster request
func (la *ListProjectClusterAction) Handle(ctx context.Context,
	req *cmproto.ListProjectClusterReq, resp *cmproto.ListProjectClusterResp) {
	if req == nil || resp == nil {
		blog.Errorf("list project cluster failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listProjectCluster(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
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

// listCluster cluster list
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
		// vcluster 使用的共享集群
		if la.req.GetShowVCluster() {
			extra := clusterList[i].GetExtraInfo()
			_, ok := extra[common.ShowSharedCluster]
			if !ok {
				continue
			}
		}

		// 用户共享集群
		if clusterList[i].GetSharedRanges() != nil &&
			(len(clusterList[i].GetSharedRanges().GetProjectIdOrCodes()) > 0 ||
				len(clusterList[i].GetSharedRanges().GetBizs()) > 0) {
			continue
		}

		// 平台共享集群
		la.clusterList = append(la.clusterList, shieldClusterInfo(clusterList[i]))
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

// setResp resp body
func (la *ListCommonClusterAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.clusterList
}

// Handle list common cluster request
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
}

// ListNodesInClusterAction list action for cluster
type ListNodesInClusterAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.ListNodesInClusterRequest
	resp  *cmproto.ListNodesInClusterResponse

	cluster *cmproto.Cluster
	cloud   *cmproto.Cloud
	k8sOp   *clusterops.K8SOperator
	nodes   []*cmproto.ClusterNode
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
	var err error

	if err = la.req.Validate(); err != nil {
		return err
	}

	// get cluster & cloud
	la.cluster, err = la.model.GetCluster(la.ctx, la.req.ClusterID)
	if err != nil {
		return err
	}
	la.cloud, err = la.model.GetCloud(la.ctx, la.cluster.Provider)
	if err != nil {
		return err
	}

	return nil
}

// listNodes merge cluster and db nodes
func (la *ListNodesInClusterAction) listNodes() error {
	condM := make(operator.M)
	condM["clusterid"] = la.req.ClusterID

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

	// remove passwd
	if !la.req.ShowPwd {
		removeNodeSensitiveInfo(nodes)
	}

	cmNodes := make([]*cmproto.ClusterNode, 0)
	for i := range nodes {
		cmNodes = append(cmNodes, transNodeToClusterNode(la.model, nodes[i]))
	}

	// get cluster nodes
	k8sNodes := autils.FilterNodesRole(la.getK8sNodes(cmNodes), false)
	la.nodes = mergeClusterNodes(la.cluster, cmNodes, k8sNodes)

	return nil
}

func (la *ListNodesInClusterAction) getK8sNodes(cmNodes []*cmproto.ClusterNode) []*corev1.Node {
	if !autils.CheckIfGetNodesFromCluster(la.cluster, la.cloud, cmNodes) {
		blog.Infof("ListNodesInClusterAction[%s] getK8sNodes clusterNodes empty", la.req.ClusterID)
		return nil
	}

	k8sNodes, err := la.k8sOp.ListClusterNodes(la.ctx, la.req.ClusterID)
	if err != nil {
		blog.Warnf("ListClusterNodes %s failed, %s", la.req.ClusterID, err.Error())
		return nil
	}
	return k8sNodes
}

// setResp resp body
func (la *ListNodesInClusterAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = code == common.BcsErrClusterManagerSuccess
	la.resp.Data = la.nodes
}

// Handle list cluster request
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
	// cloud nodes addition features
	la.appendCloudNodeInfo()

	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

func (la *ListNodesInClusterAction) appendCloudNodeInfo() {
	clsMgr, err := cloudprovider.GetClusterMgr(la.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("ListNodesInClusterAction[%s] %s GetClusterMgr failed, %s", la.cluster.ClusterID,
			la.cloud.CloudProvider, err.Error())
		return
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     la.cloud,
		AccountID: la.cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("ListNodesInClusterAction[%s] GetCredential[%s] failed, %s", la.cluster.ClusterID,
			la.cloud.CloudID, err.Error())
		return
	}
	cmOption.Region = la.cluster.Region

	err = clsMgr.AppendCloudNodeInfo(la.ctx, la.nodes, cmOption)
	if err != nil {
		blog.Errorf("ListNodesInClusterAction[%s] AppendCloudNodeInfo failed: %v", la.cluster.ClusterID, err)
		return
	}
}

// nolint
func (la *ListNodesInClusterAction) handleNodes() {
	// get all nodes instance cloud info
	ips := make([]string, 0)
	for i := range la.nodes {
		if la.nodes[i].InnerIP != "" {
			ips = append(ips, la.nodes[i].InnerIP)
		}
	}
	nodes, err := autils.GetCloudInstanceList(ips, la.cluster, la.cloud)
	if err != nil {
		blog.Errorf("GetCloudInstanceList[%s] handleNodes failed: %v", la.req.ClusterID, err)
		return
	}
	instanceMap := make(map[string]*cmproto.Node, 0)
	for i := range nodes {
		instanceMap[nodes[i].InnerIP] = nodes[i]
	}
	// 获取语言
	lang := i18n.LanguageFromCtx(la.ctx)

	// get node zoneName
	for i := range la.nodes {
		node, ok := instanceMap[la.nodes[i].InnerIP]
		if ok {
			la.nodes[i].ZoneName = node.ZoneName
			if lang != "zh" {
				la.nodes[i].ZoneName = node.ZoneID
			}
		}
	}
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

// listNodes list cluster nodes
func (la *ListMastersInClusterAction) listNodes() error {
	cls, err := la.model.GetCluster(la.ctx, la.req.ClusterID)
	if err != nil {
		blog.Errorf("get cluster %s failed, %s", la.req.ClusterID, err.Error())
		return err
	}
	cloud, err := la.model.GetCloud(la.ctx, cls.Provider)
	if err != nil {
		blog.Errorf("get cloud %s failed, %s", cls.Provider, err.Error())
		return err
	}
	clsNodes, err := getClusterNodes(la.model, cls)
	if err != nil {
		blog.Errorf("get cluster %s nodes failed, %s", cls.ClusterID, err.Error())
	}

	if !autils.CheckIfGetNodesFromCluster(cls, cloud, clsNodes) {
		cmNodes := make([]*cmproto.ClusterNode, 0)
		for node := range cls.GetMaster() {
			cmNodes = append(cmNodes, transNodeToClusterNode(la.model, cls.GetMaster()[node]))
		}
		la.nodes = cmNodes
	} else {
		// get cluster masters
		masters, err := la.k8sOp.ListClusterNodes(la.ctx, la.req.ClusterID)
		if err != nil {
			blog.Warnf("ListClusterNodes %s failed, %s", la.req.ClusterID, err.Error())
			return err
		}

		masters = autils.FilterNodesRole(masters, true)
		la.nodes = transK8sNodesToClusterNodes(la.req.ClusterID, masters)
	}

	// append cmdb host info
	la.appendHostInfo()
	// la.appendNodeAgent()
	return nil
}

// appendNodeAgent appedn node agentInfo
func (la *ListMastersInClusterAction) appendNodeAgent() { // nolint
	gseClient := gse.GetGseClient()
	hosts := make([]gse.Host, 0)
	for _, v := range la.nodes {
		hosts = append(hosts, gse.Host{IP: v.InnerIP, BKCloudID: int(v.BkCloudID)})
	}
	if len(hosts) == 0 {
		return
	}
	_, err := gseClient.GetAgentStatusV1(&gse.GetAgentStatusReq{
		Hosts: hosts,
	})
	if err != nil {
		blog.Warnf("GetAgentStatus for %s failed, %s", utils.ToJSONString(hosts), err.Error())
		return
	}
	/*
		for i := range la.nodes {
			la.nodes[i].Agent = uint32(resp.Data[gse.BKAgentKey(gse.DefaultBKCloudID,
				la.nodes[i].InnerIP)].BKAgentAlive)
		}
	*/
}

// appendHostInfo host info
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
				la.nodes[i].BkCloudID = uint32(v.BkCloudID)
			}
		}
	}
}

// setResp resp body
func (la *ListMastersInClusterAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.nodes
}

// Handle list cluster request
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
}

// getUserClusterPermList get user clusters perm
func getUserClusterPermList(iam iam.PermClient, user actions.PermInfo,
	clusterList []string) (map[string]map[string]interface{}, error) {

	permissions := make(map[string]map[string]interface{}, 0)
	clusterPerm := cluster.NewBCSClusterPermClient(iam)

	actionIDs := []string{cluster.ClusterView.String(), cluster.ClusterManage.String(), cluster.ClusterDelete.String()}
	perms, err := clusterPerm.GetMultiClusterMultiActionPerm(user.UserID, user.ProjectID, clusterList, actionIDs)
	if err != nil {
		blog.Errorf("getUserClusterPermList GetMultiClusterMultiActionPermission failed: %v", err)
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

// getCloudProviderEngine get cluster cloud engineType
func getCloudProviderEngine(model store.ClusterManagerModel, cls *cmproto.Cluster) string {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cloud, err := model.GetCloud(ctx, cls.Provider)
	if err != nil {
		blog.Errorf("GetCluster[%s] GetCloudProviderEngine failed: %v", cls.ClusterID, err)
		return ""
	}

	return cloud.GetEngineType()
}

// returnClusterExtraInfo return cluster extra info
func returnClusterExtraInfo(model store.ClusterManagerModel,
	clusterList []*cmproto.Cluster) map[string]*cmproto.ExtraInfo {
	extraInfo := make(map[string]*cmproto.ExtraInfo, 0)

	// cluster extra info
	for i := range clusterList {
		extraInfo[clusterList[i].ClusterID] = &cmproto.ExtraInfo{
			CanDeleted:   true,
			ProviderType: getCloudProviderEngine(model, clusterList[i]),
			AutoScale:    IsSupportAutoScale(model, clusterList[i]),
		}
	}

	return extraInfo
}

// getSharedCluster get shared clusters
func getSharedCluster(projectId string, bizId string, model store.ClusterManagerModel) ([]*cmproto.Cluster, error) {
	condM := make(operator.M)
	condM["isshared"] = true
	condCluster := operator.NewLeafCondition(operator.Eq, condM)
	condStatus := operator.NewLeafCondition(operator.Ne, operator.M{"status": common.StatusDeleted})

	branchCond := operator.NewBranchCondition(operator.And, condCluster, condStatus)
	clusterList, err := model.ListCluster(context.Background(), branchCond, &storeopt.ListOption{})
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		return nil, err
	}

	clusters := make([]*cmproto.Cluster, 0)

	for i := range clusterList {
		if clusterList[i].ProjectID == projectId {
			continue
		}
		// 是否共享给当前项目/业务
		if clusterList[i].SharedRanges != nil && ((len(clusterList[i].SharedRanges.GetProjectIdOrCodes()) > 0 &&
			utils.StringContainInSlice(projectId, clusterList[i].SharedRanges.ProjectIdOrCodes)) ||
			(len(clusterList[i].SharedRanges.GetBizs()) > 0 && len(bizId) > 0 &&
				utils.StringContainInSlice(bizId, clusterList[i].SharedRanges.GetBizs()))) {
			clusters = append(clusters, shieldClusterInfo(clusterList[i]))

			continue
		}

		// 共享给所有项目/业务
		if clusterList[i].SharedRanges == nil || (clusterList[i].SharedRanges != nil &&
			(len(clusterList[i].SharedRanges.GetProjectIdOrCodes()) == 0 &&
				len(clusterList[i].SharedRanges.GetBizs()) == 0)) {
			clusters = append(clusters, shieldClusterInfo(clusterList[i]))
			continue
		}
	}

	if len(clusters) > 0 {
		sort.Sort(utils.ClusterSlice(clusters))
	}

	return clusters, nil
}

// ListBasicInfoAction list action for cluster basic info
type ListBasicInfoAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.ListClusterV2Req
	resp  *cmproto.ListClusterV2Resp

	clusterBasicInfoList []*cmproto.ClusterBasicInfo
}

// NewListBasicInfoAction create list action for cluster basic info
func NewListBasicInfoAction(model store.ClusterManagerModel) *ListBasicInfoAction {
	return &ListBasicInfoAction{
		model: model,
	}
}

// validate request validation
func (la *ListBasicInfoAction) validate() error {
	if err := la.req.Validate(); err != nil {
		return err
	}

	// env
	if len(la.req.Environment) > 0 {
		_, ok := common.EnvironmentLookup[la.req.Environment]
		if !ok {
			return fmt.Errorf("request Environment invalid, must be [test/debug/prod]")
		}
	}

	// engineType
	if len(la.req.EngineType) > 0 {
		_, ok := common.EngineTypeLookup[la.req.EngineType]
		if !ok {
			return fmt.Errorf("request EngineType invalid, must be [k8s/mesos]")
		}
	}

	// clusterType
	if len(la.req.ClusterType) > 0 {
		_, ok := common.ClusterTypeLookup[la.req.ClusterType]
		if !ok {
			return fmt.Errorf("request ClusterType invalid, must be [federation/signal]")
		}
	}

	return nil
}

// filterClusterBasicInfoList filter cluster list
func (la *ListBasicInfoAction) filterClusterBasicInfoList() ([]*cmproto.Cluster, error) {
	var (
		clusterList []*cmproto.Cluster
		err         error
	)

	conds := make([]*operator.Condition, 0)
	condM := make(operator.M)

	// filter
	if len(la.req.ProjectID) != 0 {
		condM["projectid"] = la.req.ProjectID
	}
	if len(la.req.BusinessID) != 0 {
		condM["businessid"] = la.req.BusinessID
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
	if len(la.req.Environment) != 0 {
		condM["environment"] = la.req.Environment
	}
	if len(la.req.EngineType) != 0 {
		condM["enginetype"] = la.req.EngineType
	}
	if len(la.req.ClusterType) != 0 {
		condM["clustertype"] = la.req.ClusterType
	}
	if len(la.req.Status) != 0 {
		condM["status"] = la.req.Status
	}
	if len(la.req.SystemID) != 0 {
		condM["systemid"] = la.req.SystemID
	}
	if len(la.req.ClusterID) != 0 {
		condM["clusterid"] = la.req.ClusterID
	}
	condCluster := operator.NewLeafCondition(operator.Eq, condM)
	conds = append(conds, condCluster)

	// if not all, filter deleted
	if !la.req.All {
		condStatus := operator.NewLeafCondition(operator.Ne, operator.M{"status": common.StatusDeleted})
		conds = append(conds, condStatus)
	}
	branchCond := operator.NewBranchCondition(operator.And, conds...)

	listOpt := &storeopt.ListOption{
		Offset: int64(la.req.Offset),
		Limit:  int64(la.req.Limit),
	}
	clusterList, err = la.model.ListCluster(la.ctx, branchCond, listOpt)
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		return nil, err
	}

	return clusterList, nil
}

// listCluster cluster list
func (la *ListBasicInfoAction) listClusterBasicInfo() error {
	clusterList, err := la.filterClusterBasicInfoList()
	if err != nil {
		return err
	}

	for i := range clusterList {
		if clusterList[i].IsShared {
			clusterList[i].IsShared = false
		}

		la.clusterBasicInfoList = append(la.clusterBasicInfoList, clusterToClusterBasicInfo(clusterList[i]))
	}

	return nil
}

// setResp resp body
func (la *ListBasicInfoAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.clusterBasicInfoList
}

// Handle list cluster request
func (la *ListBasicInfoAction) Handle(ctx context.Context, req *cmproto.ListClusterV2Req, resp *cmproto.ListClusterV2Resp) {
	if req == nil || resp == nil {
		blog.Errorf("list cluster failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp
	// validate
	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	// list cluster
	if err := la.listClusterBasicInfo(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
