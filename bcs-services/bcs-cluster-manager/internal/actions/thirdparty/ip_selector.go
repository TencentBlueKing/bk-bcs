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

package thirdparty

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/gse"
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
	moduleInfo := cmdb.NewIpSelector(cmdb.GetCmdbClient(), gse.GetGseClient()).GetCustomSettingModuleList(la.req.ModuleList)

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
	return
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

	for _, scope := range ga.req.ScopeList {
		if scope.ScopeType == common.Biz {
			bizID, _ := strconv.Atoi(scope.ScopeId)
			topoData, err := ipSelector.GetBizModuleTopoData(bizID)
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
	return
}

// GetTopologyNodesAction action for get biz topology nodes
type GetTopologyNodesAction struct {
	ctx  context.Context
	req  *cmproto.GetTopologyNodesRequest
	resp *cmproto.GetTopologyNodesResponse
}

// NewGetTopoNodesAction create action
func NewGetTopoNodesAction() *GetTopologyNodesAction {
	return &GetTopologyNodesAction{}
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
	if gt.req.PageSize <= 0 {
		gt.req.PageSize = 20
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

func (gt *GetTopologyNodesAction) listBizTopologyNodes() error {
	ipSelector := cmdb.NewIpSelector(cmdb.GetCmdbClient(), gse.GetGseClient())

	bizID, _ := strconv.Atoi(gt.req.ScopeId)
	modules := gt.buildModuleInfo()
	filter := gt.buildFilterCondition()

	topoNodes, err := ipSelector.GetBizTopoHostData(bizID, modules, filter)
	if err != nil {
		blog.Errorf("GetTopologyNodesAction GetBizTopoHostData[%v] failed: %v", bizID, err)
		return err
	}

	gt.resp.Data = &cmproto.GetTopologyNodesData{
		Start:    gt.req.Start,
		PageSize: gt.req.PageSize,
		Total:    uint64(len(topoNodes)),
	}

	data := make([]*cmproto.HostData, 0)
	endIndex := gt.req.Start + gt.req.PageSize
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
	return
}

// GetScopeHostCheckAction action for get scope host check
type GetScopeHostCheckAction struct {
	ctx  context.Context
	req  *cmproto.GetScopeHostCheckRequest
	resp *cmproto.GetScopeHostCheckResponse
}

// NewGetScopeHostCheckAction create action
func NewGetScopeHostCheckAction() *GetScopeHostCheckAction {
	return &GetScopeHostCheckAction{}
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

	ipSelector := cmdb.NewIpSelector(cmdb.GetCmdbClient(), gse.GetGseClient())

	bizID, _ := strconv.Atoi(gt.req.ScopeId)
	modules := gt.buildModuleInfo()
	filter := gt.buildFilterCondition()

	topoNodes, err := ipSelector.GetBizTopoHostData(bizID, modules, filter)
	if err != nil {
		blog.Errorf("GetTopologyNodesAction GetBizTopoHostData[%v] failed: %v", bizID, err)
		return err
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
	return
}
