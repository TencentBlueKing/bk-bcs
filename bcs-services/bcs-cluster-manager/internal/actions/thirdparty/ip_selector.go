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

package thirdparty

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/tenant"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/gse"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// GetCustomSettingAction action for get custom setting
type GetCustomSettingAction struct {
	ctx  context.Context
	req  *cmproto.GetBatchCustomSettingRequest
	resp *cmproto.GetBatchCustomSettingResponse
}

// NewGetCustomSettingAction create action
func NewGetCustomSettingAction() *GetCustomSettingAction {
	return &GetCustomSettingAction{}
}

func (la *GetCustomSettingAction) validate() error {
	if err := la.req.Validate(); err != nil {
		return err
	}

	return nil
}

func (la *GetCustomSettingAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.ErrorMsg = msg
	la.resp.Success = (code == common.BcsErrClusterManagerSuccess)
}

func (la *GetCustomSettingAction) listModuleCustomSetting() error {
	moduleInfo := cmdb.NewIpSelector(cmdb.GetCmdbClient(),
		gse.GetGseClient()).GetCustomSettingModuleList(la.req.ModuleList)

	result, err := utils.MarshalInterfaceToValue(moduleInfo)
	if err != nil {
		blog.Errorf("marshal moduleInfo err, %s", err.Error())
		la.setResp(common.BcsErrClusterManagerCommonErr, err.Error())
		return err
	}

	la.resp.Data = result
	return nil
}

// Handle handles customSetting data
func (la *GetCustomSettingAction) Handle(ctx context.Context, req *cmproto.GetBatchCustomSettingRequest,
	resp *cmproto.GetBatchCustomSettingResponse) {
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := la.listModuleCustomSetting(); err != nil {
		la.setResp(common.BcsErrClusterManagerCommonErr, err.Error())
		return
	}

	blog.Infof("GetCustomSettingAction list module custom setting successfully")
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// GetBizInstanceTopoAction action for get biz topo
type GetBizInstanceTopoAction struct {
	ctx  context.Context
	req  *cmproto.GetBizTopologyHostRequest
	resp *cmproto.GetBizTopologyHostResponse
}

// NewGetBizInstanceTopoAction create action
func NewGetBizInstanceTopoAction() *GetBizInstanceTopoAction {
	return &GetBizInstanceTopoAction{}
}

func (ga *GetBizInstanceTopoAction) validate() error {
	if err := ga.req.Validate(); err != nil {
		return err
	}
	if len(ga.req.ScopeList) == 0 {
		return fmt.Errorf("GetBizInstanceTopoAction scopeList empty")
	}

	for _, scope := range ga.req.ScopeList {
		if scope.GetScopeType() != common.Biz && scope.GetScopeType() != common.BizSet {
			return fmt.Errorf("GetBizInstanceTopoAction scopeList not supported %s", scope.GetScopeType())
		}
	}

	return nil
}

func (ga *GetBizInstanceTopoAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.ErrorMsg = msg
	ga.resp.Success = (code == common.BcsErrClusterManagerSuccess)
}

func (ga *GetBizInstanceTopoAction) listBizHostTopo() error {
	bizTopos := make([]*cmdb.BizInstanceTopoData, 0)
	ipSelector := cmdb.NewIpSelector(cmdb.GetCmdbClient(), gse.GetGseClient())

	user := auth.GetAuthAndTenantInfoFromCtx(ga.ctx)
	ctx := tenant.WithTenantIdFromContext(ga.ctx, user.ResourceTenantId)

	for _, scope := range ga.req.ScopeList {
		if scope.ScopeType == common.Biz {
			bizID, _ := strconv.Atoi(scope.ScopeId)
			topoData, err := ipSelector.GetBizModuleTopoData(ctx, bizID)
			if err != nil {
				blog.Errorf("GetBizInstanceTopoAction GetBizModuleTopoData[%v] failed: %v", bizID, err)
				continue
			}

			bizTopos = append(bizTopos, topoData)
		}
	}

	result, err := utils.MarshalInterfaceToListValue(bizTopos)
	if err != nil {
		blog.Errorf("marshal BizInstanceTopoData err, %s", err.Error())
		ga.setResp(common.BcsErrClusterManagerCommonErr, err.Error())
		return err
	}

	ga.resp.Data = result
	return nil
}

// Handle handles customSetting data
func (ga *GetBizInstanceTopoAction) Handle(ctx context.Context, req *cmproto.GetBizTopologyHostRequest,
	resp *cmproto.GetBizTopologyHostResponse) {
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	if err := ga.validate(); err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ga.listBizHostTopo(); err != nil {
		ga.setResp(common.BcsErrClusterManagerCommonErr, err.Error())
		return
	}

	blog.Infof("GetBizInstanceTopoAction get biz[%v] instanceTopo successfully", ga.req.ScopeId)
	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// GetTopologyNodesAction action for get biz topology nodes
type GetTopologyNodesAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.GetTopologyNodesRequest
	resp  *cmproto.GetTopologyNodesResponse
}

// NewGetTopoNodesAction create action
func NewGetTopoNodesAction(model store.ClusterManagerModel) *GetTopologyNodesAction {
	return &GetTopologyNodesAction{model: model}
}

func (gt *GetTopologyNodesAction) validate() error {
	if err := gt.req.Validate(); err != nil {
		return err
	}

	if gt.req.ScopeType != common.Biz {
		return fmt.Errorf("GetTopologyNodesAction scopeType[%s] not supported", gt.req.ScopeType)
	}

	if gt.req.Start <= 0 {
		gt.req.Start = 0
	}

	return nil
}

func (gt *GetTopologyNodesAction) setResp(code uint32, msg string) {
	gt.resp.Code = code
	gt.resp.ErrorMsg = msg
	gt.resp.Success = (code == common.BcsErrClusterManagerSuccess)
}

func (gt *GetTopologyNodesAction) buildModuleInfo() []cmdb.HostModuleInfo {
	modules := make([]cmdb.HostModuleInfo, 0)
	for i := range gt.req.NodeList {
		modules = append(modules, cmdb.HostModuleInfo{
			ObjectID:   gt.req.NodeList[i].ObjectId,
			InstanceID: int64(gt.req.NodeList[i].InstanceId),
		})
	}

	return modules
}

func (gt *GetTopologyNodesAction) buildFilterCondition() cmdb.HostFilter {
	if gt.req.Alive == nil && gt.req.SearchContent == "" {
		return &cmdb.HostFilterEmpty{}
	}

	filter := &cmdb.HostFilterTopoNodes{Alive: nil, SearchContent: ""}
	if gt.req.Alive != nil {
		alive := int(gt.req.Alive.GetValue())
		filter.Alive = &alive
	}
	if gt.req.SearchContent != "" {
		filter.SearchContent = gt.req.SearchContent
	}

	return filter
}

func nodeExistClusterManager(model store.ClusterManagerModel, nodeIP string,
	masterIps map[string]cluster.ClusterInfo) bool {
	// check if exist masterIPs
	if _, ok := masterIps[nodeIP]; ok {
		return true
	}

	// check if exist nodeIPs
	_, err := model.GetNodeByIP(context.Background(), nodeIP)
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		return false
	}

	if errors.Is(err, drivers.ErrTableRecordNotFound) {
		return false
	}

	return true
}

func filterAvailableNodes(model store.ClusterManagerModel, bizNodes []cmdb.HostDetailInfo) []cmdb.HostDetailInfo {
	// get all masterIPs
	masterIPs := cluster.GetAllMasterIPs(model)

	availableNodes := make([]cmdb.HostDetailInfo, 0)

	for i := range bizNodes {
		exist := nodeExistClusterManager(model, bizNodes[i].Ip, masterIPs)
		if !exist {
			availableNodes = append(availableNodes, bizNodes[i])
		}
	}

	return availableNodes
}

func (gt *GetTopologyNodesAction) listBizTopologyNodes() error {
	ipSelector := cmdb.NewIpSelector(cmdb.GetCmdbClient(), gse.GetGseClient())

	bizID, _ := strconv.Atoi(gt.req.ScopeId)
	modules := gt.buildModuleInfo()
	filter := gt.buildFilterCondition()

	user := auth.GetAuthAndTenantInfoFromCtx(gt.ctx)
	ctx := tenant.WithTenantIdFromContext(gt.ctx, user.ResourceTenantId)

	var (
		topoNodes []cmdb.HostDetailInfo
		err       error
	)

	topoNodes, err = ipSelector.GetBizTopoHostData(ctx, bizID, modules, filter)
	if err != nil {
		blog.Errorf("GetTopologyNodesAction GetBizTopoHostData[%v] failed: %v", bizID, err)
		return err
	}

	// 过滤集群可用节点
	if gt.req.ShowAvailableNode {
		topoNodes = filterAvailableNodes(gt.model, topoNodes)
	}

	gt.resp.Data = &cmproto.GetTopologyNodesData{
		Start:    gt.req.Start,
		PageSize: gt.req.PageSize,
		Total:    uint64(len(topoNodes)),
	}

	data := make([]*cmproto.HostData, 0)

	var endIndex uint64
	if gt.req.PageSize <= 0 {
		endIndex = uint64(len(topoNodes))
	} else {
		endIndex = gt.req.Start + uint64(gt.req.PageSize)
	}

	for index, host := range topoNodes {
		if index >= int(gt.req.Start) && index < int(endIndex) {
			data = append(data, &cmproto.HostData{
				HostId:   uint64(host.HostId),
				Ip:       host.Ip,
				Ipv6:     host.Ipv6,
				HostName: host.HostName,
				Alive:    uint32(host.Alive),
				OsName:   host.OsName,
				CloudArea: &cmproto.HostCloudArea{
					Id:   uint32(host.CloudArea.ID),
					Name: host.CloudArea.Name,
				},
			})
		}
	}
	gt.resp.Data.Data = data

	return nil
}

// Handle handles customSetting data
func (gt *GetTopologyNodesAction) Handle(ctx context.Context, req *cmproto.GetTopologyNodesRequest,
	resp *cmproto.GetTopologyNodesResponse) {
	gt.ctx = ctx
	gt.req = req
	gt.resp = resp

	if err := gt.validate(); err != nil {
		gt.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := gt.listBizTopologyNodes(); err != nil {
		gt.setResp(common.BcsErrClusterManagerCommonErr, err.Error())
		return
	}

	blog.Infof("GetTopologyNodesAction get biz[%v] topologyNodes successfully", gt.req.ScopeId)
	gt.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// GetScopeHostCheckAction action for get scope host check
type GetScopeHostCheckAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.GetScopeHostCheckRequest
	resp  *cmproto.GetScopeHostCheckResponse
}

// NewGetScopeHostCheckAction create action
func NewGetScopeHostCheckAction(model store.ClusterManagerModel) *GetScopeHostCheckAction {
	return &GetScopeHostCheckAction{model: model}
}

func (gt *GetScopeHostCheckAction) validate() error {
	if err := gt.req.Validate(); err != nil {
		return err
	}

	if gt.req.ScopeType != common.Biz {
		return fmt.Errorf("GetScopeHostCheckAction scopeType[%s] not supported", gt.req.ScopeType)
	}

	return nil
}

func (gt *GetScopeHostCheckAction) setResp(code uint32, msg string) {
	gt.resp.Code = code
	gt.resp.ErrorMsg = msg
	gt.resp.Success = (code == common.BcsErrClusterManagerSuccess)
}

func (gt *GetScopeHostCheckAction) buildModuleInfo() []cmdb.HostModuleInfo {
	modules := make([]cmdb.HostModuleInfo, 0)
	scopeId, _ := strconv.Atoi(gt.req.ScopeId)

	modules = append(modules, cmdb.HostModuleInfo{
		ObjectID:   gt.req.ScopeType,
		InstanceID: int64(scopeId),
	})
	return modules
}

func (gt *GetScopeHostCheckAction) buildFilterCondition() cmdb.HostFilter {
	return &cmdb.HostFilterCheckNodes{
		IpList:   gt.req.IpList,
		Ipv6List: gt.req.Ipv6List,
		KeyList:  gt.req.KeyList,
	}
}

func (gt *GetScopeHostCheckAction) listScopeHostInfo() error {
	if len(gt.req.IpList) == 0 && len(gt.req.Ipv6List) == 0 && len(gt.req.KeyList) == 0 {
		blog.Infof("GetScopeHostCheckAction listScopeHostInfo paras[ipList/ipv6List/keyList] empty")
		return nil
	}

	user := auth.GetAuthAndTenantInfoFromCtx(gt.ctx)
	ctx := tenant.WithTenantIdFromContext(gt.ctx, user.ResourceTenantId)

	ipSelector := cmdb.NewIpSelector(cmdb.GetCmdbClient(), gse.GetGseClient())

	bizID, _ := strconv.Atoi(gt.req.ScopeId)
	modules := gt.buildModuleInfo()
	filter := gt.buildFilterCondition()

	var (
		topoNodes []cmdb.HostDetailInfo
		err       error
	)

	topoNodes, err = ipSelector.GetBizTopoHostData(ctx, bizID, modules, filter)
	if err != nil {
		blog.Errorf("GetTopologyNodesAction GetBizTopoHostData[%v] failed: %v", bizID, err)
		return err
	}

	// 过滤集群可用节点
	if gt.req.ShowAvailableNode {
		topoNodes = filterAvailableNodes(gt.model, topoNodes)
	}

	// get topology nodes
	data := make([]*cmproto.HostData, 0)
	for _, host := range topoNodes {
		data = append(data, &cmproto.HostData{
			HostId:   uint64(host.HostId),
			Ip:       host.Ip,
			Ipv6:     host.Ipv6,
			HostName: host.HostName,
			Alive:    uint32(host.Alive),
			OsName:   host.OsName,
			CloudArea: &cmproto.HostCloudArea{
				Id:   uint32(host.CloudArea.ID),
				Name: host.CloudArea.Name,
			},
		})
	}
	gt.resp.Data = data

	return nil
}

// Handle handles customSetting data
func (gt *GetScopeHostCheckAction) Handle(ctx context.Context, req *cmproto.GetScopeHostCheckRequest,
	resp *cmproto.GetScopeHostCheckResponse) {
	gt.ctx = ctx
	gt.req = req
	gt.resp = resp

	if err := gt.validate(); err != nil {
		gt.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := gt.listScopeHostInfo(); err != nil {
		gt.setResp(common.BcsErrClusterManagerCommonErr, err.Error())
		return
	}

	blog.Infof("GetScopeHostCheckAction get biz[%v] scopeHost successfully", gt.req.ScopeId)
	gt.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// GetTopologyHostIdsNodesAction action for get biz topology hostIds
type GetTopologyHostIdsNodesAction struct {
	ctx  context.Context
	req  *cmproto.GetTopologyHostIdsNodesRequest
	resp *cmproto.GetTopologyHostIdsNodesResponse
}

// NewGetTopologyHostIdsNodesAction create action
func NewGetTopologyHostIdsNodesAction() *GetTopologyHostIdsNodesAction {
	return &GetTopologyHostIdsNodesAction{}
}

func (gt *GetTopologyHostIdsNodesAction) validate() error {
	if err := gt.req.Validate(); err != nil {
		return err
	}

	if gt.req.ScopeType != common.Biz {
		return fmt.Errorf("GetTopologyHostIdsNodesAction scopeType[%s] not supported", gt.req.ScopeType)
	}

	if gt.req.Start <= 0 {
		gt.req.Start = 0
	}

	return nil
}

func (gt *GetTopologyHostIdsNodesAction) setResp(code uint32, msg string) {
	gt.resp.Code = code
	gt.resp.ErrorMsg = msg
	gt.resp.Success = (code == common.BcsErrClusterManagerSuccess)
}

func (gt *GetTopologyHostIdsNodesAction) buildModuleInfo() []cmdb.HostModuleInfo {
	modules := make([]cmdb.HostModuleInfo, 0)
	for i := range gt.req.NodeList {
		modules = append(modules, cmdb.HostModuleInfo{
			ObjectID:   gt.req.NodeList[i].ObjectId,
			InstanceID: int64(gt.req.NodeList[i].InstanceId),
		})
	}

	return modules
}

func (gt *GetTopologyHostIdsNodesAction) buildFilterCondition() cmdb.HostFilter {
	if gt.req.Alive == nil && gt.req.SearchContent == "" {
		return &cmdb.HostFilterEmpty{}
	}

	filter := &cmdb.HostFilterTopoNodes{Alive: nil, SearchContent: ""}
	if gt.req.Alive != nil {
		alive := int(gt.req.Alive.GetValue())
		filter.Alive = &alive
	}
	if gt.req.SearchContent != "" {
		filter.SearchContent = gt.req.SearchContent
	}

	return filter
}

func (gt *GetTopologyHostIdsNodesAction) listBizTopologyHostIdsNodes() error {
	ipSelector := cmdb.NewIpSelector(cmdb.GetCmdbClient(), gse.GetGseClient())

	user := auth.GetAuthAndTenantInfoFromCtx(gt.ctx)
	ctx := tenant.WithTenantIdFromContext(gt.ctx, user.ResourceTenantId)

	bizID, _ := strconv.Atoi(gt.req.ScopeId)
	modules := gt.buildModuleInfo()
	filter := gt.buildFilterCondition()

	topoNodes, err := ipSelector.GetBizTopoHostData(ctx, bizID, modules, filter)
	if err != nil {
		blog.Errorf("GetTopologyHostIdsNodesAction GetBizTopoHostData[%v] failed: %v", bizID, err)
		return err
	}

	gt.resp.Data = &cmproto.GetTopologyHostIdsNodesData{
		Start:    gt.req.Start,
		PageSize: gt.req.PageSize,
		Total:    uint64(len(topoNodes)),
	}

	data := make([]*cmproto.HostIDsNodeData, 0)

	var endIndex uint64
	if gt.req.PageSize <= 0 {
		endIndex = uint64(len(topoNodes))
	} else {
		endIndex = gt.req.Start + uint64(gt.req.PageSize)
	}

	for index, host := range topoNodes {
		if index >= int(gt.req.Start) && index < int(endIndex) {
			data = append(data, &cmproto.HostIDsNodeData{
				HostId: uint64(host.HostId),
			})
		}
	}
	gt.resp.Data.Data = data

	return nil
}

// Handle handles customSetting data
func (gt *GetTopologyHostIdsNodesAction) Handle(ctx context.Context, req *cmproto.GetTopologyHostIdsNodesRequest,
	resp *cmproto.GetTopologyHostIdsNodesResponse) {
	gt.ctx = ctx
	gt.req = req
	gt.resp = resp

	if err := gt.validate(); err != nil {
		gt.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := gt.listBizTopologyHostIdsNodes(); err != nil {
		gt.setResp(common.BcsErrClusterManagerCommonErr, err.Error())
		return
	}

	blog.Infof("GetTopologyNodesAction get biz[%v] topologyNodes successfully", gt.req.ScopeId)
	gt.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// GetHostsDetailsAction action for get biz host details
type GetHostsDetailsAction struct {
	ctx     context.Context
	req     *cmproto.GetHostsDetailsRequest
	resp    *cmproto.GetHostsDetailsResponse
	hostIds []int
}

// NewGetHostsDetailsAction create action
func NewGetHostsDetailsAction() *GetHostsDetailsAction {
	return &GetHostsDetailsAction{
		hostIds: make([]int, 0),
	}
}

func (gt *GetHostsDetailsAction) validate() error {
	if err := gt.req.Validate(); err != nil {
		return err
	}

	if gt.req.ScopeType != common.Biz {
		return fmt.Errorf("GetTopologyHostIdsNodesAction scopeType[%s] not supported", gt.req.ScopeType)
	}
	if len(gt.req.GetHostList()) == 0 {
		return fmt.Errorf("GetTopologyHostIdsNodesAction hostList empty")
	}

	for i := range gt.req.GetHostList() {
		gt.hostIds = append(gt.hostIds, int(gt.req.GetHostList()[i].HostId))
	}

	return nil
}

func (gt *GetHostsDetailsAction) setResp(code uint32, msg string) {
	gt.resp.Code = code
	gt.resp.ErrorMsg = msg
	gt.resp.Success = (code == common.BcsErrClusterManagerSuccess)
}

func (gt *GetHostsDetailsAction) listBizTopologyHostIdsNodes() error {
	ipSelector := cmdb.NewIpSelector(cmdb.GetCmdbClient(), gse.GetGseClient())

	user := auth.GetAuthAndTenantInfoFromCtx(gt.ctx)
	ctx := tenant.WithTenantIdFromContext(gt.ctx, user.ResourceTenantId)

	bizID, _ := strconv.Atoi(gt.req.ScopeId)

	topoNodes, err := ipSelector.GetBizTopoHostData(ctx, bizID,
		[]cmdb.HostModuleInfo{{InstanceID: int64(bizID)}}, nil)
	if err != nil {
		blog.Errorf("GetHostsDetailsAction GetBizTopoHostData[%v] failed: %v", bizID, err)
		return err
	}

	gt.resp.Data = make([]*cmproto.HostDataWithMeta, 0)

	for _, host := range topoNodes {
		if utils.IntInSlice(host.HostId, gt.hostIds) {
			gt.resp.Data = append(gt.resp.Data, &cmproto.HostDataWithMeta{
				HostId:   uint64(host.HostId),
				Ip:       host.Ip,
				Ipv6:     host.Ipv6,
				HostName: host.HostName,
				Alive:    uint32(host.Alive),
				OsName:   host.OsName,
				CloudArea: &cmproto.HostCloudArea{
					Id:   uint32(host.CloudArea.ID),
					Name: host.CloudArea.Name,
				},
				Meta: nil,
			})
		}
	}

	return nil
}

// Handle handles customSetting data
func (gt *GetHostsDetailsAction) Handle(ctx context.Context, req *cmproto.GetHostsDetailsRequest,
	resp *cmproto.GetHostsDetailsResponse) {
	gt.ctx = ctx
	gt.req = req
	gt.resp = resp

	if err := gt.validate(); err != nil {
		gt.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := gt.listBizTopologyHostIdsNodes(); err != nil {
		gt.setResp(common.BcsErrClusterManagerCommonErr, err.Error())
		return
	}

	blog.Infof("GetTopologyNodesAction get biz[%v] topologyNodes successfully", gt.req.ScopeId)
	gt.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
