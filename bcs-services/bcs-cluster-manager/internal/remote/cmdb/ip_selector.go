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

package cmdb

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/gse"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// IPSelector ip selector interface
type IPSelector interface {
	// GetCustomSettingModuleList get custom setting moduleList
	GetCustomSettingModuleList(moduleList []string) interface{}
	// GetBizModuleTopoData get topo modules by bizID
	GetBizModuleTopoData(ctx context.Context, bizID int) (*BizInstanceTopoData, error)
	// GetBizTopoHostData get host data by bizID
	GetBizTopoHostData(bizID int, info []HostModuleInfo, filter HostFilter) ([]HostDetailInfo, error)
}

// NewIpSelector init ipSelector client
func NewIpSelector(cmdb *Client, gse gse.Interface) IPSelector {
	return &ipSelectorClient{
		cmdb: cmdb,
		gse:  gse,
	}
}

type ipSelectorClient struct {
	cmdb *Client
	gse  gse.Interface
}

func (ipSelector *ipSelectorClient) GetCustomSettingModuleList(moduleList []string) interface{} {
	setting := make(map[string]interface{})

	for _, module := range moduleList {
		m := CustomSettingModule(module)

		moduleInfo, err := m.GetSettingColumn()
		if err != nil {
			blog.Errorf("GetCustomSettingModuleList failed: %v", err)
			continue
		}

		setting[m.String()] = moduleInfo
	}

	return setting
}

// GetBizModuleTopoData get biz module topo data
func (ipSelector *ipSelectorClient) GetBizModuleTopoData(ctx context.Context, bizID int) (*BizInstanceTopoData, error) {
	return GetBizModuleTopoData(ctx, ipSelector.cmdb, bizID)
}

func (ipSelector *ipSelectorClient) GetBizTopoHostData(
	bizID int, module []HostModuleInfo, filter HostFilter) ([]HostDetailInfo, error) {
	if len(module) == 0 {
		return nil, nil
	}

	hosts, err := GetBizHostDetailedData(ipSelector.cmdb, ipSelector.gse, bizID, module)
	if err != nil {
		return nil, err
	}

	if filter == nil {
		return hosts, nil
	}

	return filter.FilterHostByCondition(hosts), nil
}

// GetBizHostDetailedData xxx
func GetBizHostDetailedData(cmdb *Client, gseCli gse.Interface, // nolint
	bizID int, module []HostModuleInfo) ([]HostDetailInfo, error) {

	hostTopos, err := cmdb.FetchAllHostTopoRelationsByBizID(bizID)
	if err != nil {
		return nil, err
	}

	// filter module info and remove duplicates
	var (
		hostFilterResult = make([]HostTopoRelation, 0)
		hostIDMap        = make(map[int]struct{})
	)
	for _, hTopo := range hostTopos {
		if HostTopoInHostModule(hTopo, module) {
			if _, exists := hostIDMap[hTopo.BkHostID]; !exists {
				hostFilterResult = append(hostFilterResult, hTopo)
				hostIDMap[hTopo.BkHostID] = struct{}{}
			}
		}
	}

	hostData, err := cmdb.FetchAllHostsByBizID(bizID, true)
	if err != nil {
		return nil, err
	}

	hostDataMap := make(map[int64]HostData, 0)
	for _, host := range hostData {
		if _, ok := hostDataMap[host.BKHostID]; !ok {
			hostDataMap[host.BKHostID] = host
		}
	}

	// filter host info
	var (
		hosts    = make([]HostDetailInfo, 0)
		hostLock = &sync.RWMutex{}
	)
	pool := utils.NewRoutinePool(50)
	defer pool.Close()

	for _, result := range hostFilterResult {
		pool.Add(1)
		go func(r HostTopoRelation) {
			defer pool.Done()
			if h, ok := hostDataMap[int64(r.BkHostID)]; ok {
				detailInfo := HostDetailInfo{
					Ip:       h.BKHostInnerIP,
					Ipv6:     h.BKHostInnerIPV6,
					HostId:   int(h.BKHostID),
					HostName: h.BKHostName,
					Alive:    0,
					CloudArea: CloudArea{
						ID:                int64(h.BKHostCloudID),
						BKSupplierAccount: r.BkSupplierAccount,
					},
					OsName:  h.BKHostOsName,
					OsType:  h.BKHostOsType,
					AgentID: h.BkAgentID,
				}
				cloudArea, err := cmdb.SearchCloudAreaByCloudID(h.BKHostCloudID)
				if err != nil {
					blog.Errorf("GetAllBizHostTopos SearchCloudAreaByCloudID[%v] failed: %v", h.BKHostCloudID, err)
					detailInfo.CloudArea.Name = gse.DefaultBkCloudName
				} else {
					detailInfo.CloudArea.Name = cloudArea.CloudName
				}
				hostLock.Lock()
				hosts = append(hosts, detailInfo)
				hostLock.Unlock()
			}
		}(result)
	}
	pool.Wait()

	// get gse agent status
	hostAgentStatus := GetHostAgentStatus(gseCli, hosts)

	// append gse agent status
	for i := range hosts {
		hosts[i].Alive = 0

		agentFlag := gse.BKAgentKey(int(hosts[i].CloudArea.ID), hosts[i].Ip)
		if len(hosts[i].AgentID) > 0 {
			agentFlag = hosts[i].AgentID
		}
		s, ok := hostAgentStatus[agentFlag]
		if ok {
			hosts[i].Alive = s.Alive
			continue
		}
	}

	// sort hostList
	sort.Sort(HostDetailInfoList(hosts))

	return hosts, nil
}

// GetHostAgentStatus by supplyName
func GetHostAgentStatus(gseCli gse.Interface, hosts []HostDetailInfo) map[string]gse.HostAgentStatus {
	supplyAccountHost := make(map[string][]HostDetailInfo, 0)
	for _, host := range hosts {
		if host.Ip == "" {
			continue
		}

		_, ok := supplyAccountHost[host.CloudArea.BKSupplierAccount]
		if ok {
			supplyAccountHost[host.CloudArea.BKSupplierAccount] =
				append(supplyAccountHost[host.CloudArea.BKSupplierAccount], host)
			continue
		}

		if supplyAccountHost[host.CloudArea.BKSupplierAccount] == nil {
			supplyAccountHost[host.CloudArea.BKSupplierAccount] = make([]HostDetailInfo, 0)
		}

		supplyAccountHost[host.CloudArea.BKSupplierAccount] =
			append(supplyAccountHost[host.CloudArea.BKSupplierAccount], host)
	}
	// filter agent host status
	var (
		agentStatus = make([]gse.HostAgentStatus, 0)
		agentLock   = &sync.RWMutex{}
	)
	pool := utils.NewRoutinePool(20)
	defer pool.Close()

	for account, hostList := range supplyAccountHost {
		pool.Add(1)
		go func(account string, hosts []HostDetailInfo) {
			defer pool.Done()
			gseHosts := make([]gse.Host, 0)
			for i := range hosts {

				gseHosts = append(gseHosts, gse.Host{
					IP:        hosts[i].Ip,
					BKCloudID: int(hosts[i].CloudArea.ID),
					AgentID:   hosts[i].AgentID,
				})
			}

			bkAgent, err := gseCli.GetHostsGseAgentStatus(account, gseHosts)
			if err != nil {
				blog.Errorf("GetHostsGseAgentStatus[%v] failed: %v", gseHosts, err)
				return
			}
			agentLock.Lock()
			agentStatus = append(agentStatus, bkAgent...)
			agentLock.Unlock()
		}(account, hostList)
	}
	pool.Wait()

	hostAgentStatus := make(map[string]gse.HostAgentStatus, 0)
	for i := range agentStatus {
		gseFlag := gse.BKAgentKey(agentStatus[i].BKCloudID, agentStatus[i].IP)
		if len(agentStatus[i].AgentID) > 0 {
			gseFlag = agentStatus[i].AgentID
		}
		hostAgentStatus[gseFlag] = agentStatus[i]
	}

	return hostAgentStatus
}

// HostTopoInHostModule host module
func HostTopoInHostModule(topo HostTopoRelation, info []HostModuleInfo) bool {
	moduleMap := make(map[int64]struct{}, 0)
	for i := range info {
		_, ok := moduleMap[info[i].InstanceID]
		if !ok {
			moduleMap[info[i].InstanceID] = struct{}{}
		}
	}

	// bizID
	_, ok := moduleMap[int64(topo.BkBizID)]
	if ok {
		return true
	}
	// setID
	_, ok = moduleMap[int64(topo.BkSetID)]
	if ok {
		return true
	}
	// moduleID
	_, ok = moduleMap[int64(topo.BkModuleID)]
	return ok
}

// Object "biz" "set" "module"
type Object struct {
	ObjectName string
	ObjectID   int
}

// GetBizModuleTopoData get biz topo data
func GetBizModuleTopoData(ctx context.Context, cli *Client, bizID int) (*BizInstanceTopoData, error) {
	bizTopo, err := cli.ListTopology(ctx, bizID, false, false)
	if err != nil {
		return nil, err
	}

	bizHostCnt, err := GetHostCountByObject(cli, bizID, Object{
		ObjectName: "biz",
		ObjectID:   bizID,
	})
	if err != nil {
		return nil, err
	}

	// 默认 expand = true
	var topo = &BizInstanceTopoData{
		BKInstID:   bizTopo.BKInstID,
		BKInstName: bizTopo.BKInstName,
		BKObjID:    bizTopo.BKObjID,
		BKObjName:  bizTopo.BKObjName,
		Expanded:   true,
		Count:      bizHostCnt,
	}

	// set childs
	var (
		sets = make([]BizInstanceTopoData, 0)
	)

	for _, set := range bizTopo.Child {
		s := BizInstanceTopoData{
			BKInstID:   set.BKInstID,
			BKInstName: set.BKInstName,
			BKObjID:    set.BKObjID,
			BKObjName:  set.BKObjName,
			Expanded:   true,
		}
		hostCnt, err := GetHostCountByObject(cli, bizID, Object{
			ObjectName: set.BKObjID,
			ObjectID:   set.BKInstID,
		})
		if err != nil {
			blog.Errorf("GetBizModuleTopoData GetHostCountByObject failed: %v", err)
		}
		s.Count = hostCnt

		modules := GetSetModuleChild(cli, bizID, set.Child)
		s.Child = modules

		sets = append(sets, s)
	}

	topo.Child = sets

	return topo, nil
}

// GetSetModuleChild module child
func GetSetModuleChild(cli *Client, bizID int, childs []SearchBizInstTopoData) []BizInstanceTopoData {
	var (
		modules    = make([]BizInstanceTopoData, 0)
		moduleLock = &sync.RWMutex{}
	)
	pool := utils.NewRoutinePool(20)
	defer pool.Close()

	for _, c := range childs {
		pool.Add(1)
		go func(data SearchBizInstTopoData) {
			defer pool.Done()
			module := BizInstanceTopoData{
				BKInstID:   data.BKInstID,
				BKInstName: data.BKInstName,
				BKObjID:    data.BKObjID,
				BKObjName:  data.BKObjName,
				Expanded:   false,
				Child:      make([]BizInstanceTopoData, 0),
			}
			cnt, err := GetHostCountByObject(cli, bizID, Object{
				ObjectName: data.BKObjID,
				ObjectID:   data.BKInstID,
			})
			if err != nil {
				blog.Errorf("GetSetModuleChild GetHostCountByObject failed: %v", err)
			}
			module.Count = cnt

			moduleLock.Lock()
			modules = append(modules, module)
			moduleLock.Unlock()
		}(c)
	}
	pool.Wait()

	return modules
}

// GetHostCountByObject get host cnt by bizID
func GetHostCountByObject(cli *Client, bizID int, object Object) (int, error) {
	hostTopos, err := cli.FetchAllHostTopoRelationsByBizID(bizID)
	if err != nil {
		return 0, err
	}

	cnt := 0
	for _, host := range hostTopos {
		switch object.ObjectName {
		case "biz":
			return len(hostTopos), nil
		case "set":
			if host.BkSetID == object.ObjectID {
				cnt++
			}
		case "module":
			if host.BkModuleID == object.ObjectID {
				cnt++
			}
		default:
		}
	}

	return cnt, nil
}

// CustomSettingModule xxx
type CustomSettingModule string

var (
	// IpSelectorHostList xxx
	IpSelectorHostList CustomSettingModule = "ip_selector_host_list"
	hostListColumn                         = []string{"ip", "ipv6", "coludArea", "alive", "hostName", "osName",
		"osType", "hostId"}
	hostListColumnSort = []string{"ip", "ipv6", "coludArea", "alive", "hostName", "osName",
		"osType", "hostId"}
)

// String xxx
func (csm CustomSettingModule) String() string {
	return string(csm)
}

// GetSettingColumn get module setting
func (csm CustomSettingModule) GetSettingColumn() (map[string]interface{}, error) {
	switch csm {
	case IpSelectorHostList:
		return map[string]interface{}{
			"hostListColumn":     hostListColumn,
			"hostListColumnSort": hostListColumnSort,
		}, nil
	default:
	}

	return nil, fmt.Errorf("not supported CustomSettingModule type[%s]", csm)
}

// BizInstanceTopoData cmdb topo data
type BizInstanceTopoData struct {
	BKInstID   int                   `json:"instanceId"`
	BKInstName string                `json:"instanceName"`
	BKObjID    string                `json:"objectId"`
	BKObjName  string                `json:"objectName"`
	Expanded   bool                  `json:"expanded"`
	Count      int                   `json:"count"`
	Child      []BizInstanceTopoData `json:"child"`
}

// HostModuleInfo host module(biz/set/module)
type HostModuleInfo struct {
	ObjectID   string
	InstanceID int64
}

// CloudArea xxx
type CloudArea struct {
	ID                int64
	Name              string
	BKSupplierAccount string
}

// HostDetailInfo "ip", "ipv6", "coludArea", "alive", "hostName", "osName", "osType", "hostId"
type HostDetailInfo struct {
	Ip        string
	Ipv6      string
	HostId    int
	HostName  string
	Alive     int
	CloudArea CloudArea
	OsName    string
	OsType    string
	AgentID   string
}

// HostDetailInfoList xxx
type HostDetailInfoList []HostDetailInfo

func (hs HostDetailInfoList) Len() int { return len(hs) }
func (hs HostDetailInfoList) Less(i, j int) bool {
	if hs[i].Ip != "" && hs[j].Ip != "" {
		return hs[i].Ip < hs[j].Ip
	}
	if hs[i].Ipv6 != "" && hs[j].Ipv6 != "" {
		return hs[i].Ipv6 < hs[j].Ipv6
	}

	return hs[i].HostId < hs[j].HostId
}
func (hs HostDetailInfoList) Swap(i, j int) { hs[i], hs[j] = hs[j], hs[i] }

// HostFilter host filter interface for support different scene hostFilter
type HostFilter interface {
	// FilterHostByCondition filter hosts by different condition
	FilterHostByCondition(source []HostDetailInfo) []HostDetailInfo
}

// HostFilterEmpty empty host filter
type HostFilterEmpty struct{}

// FilterHostByCondition xxx
func (h *HostFilterEmpty) FilterHostByCondition(source []HostDetailInfo) []HostDetailInfo {
	return source
}

// HostFilterTopoNodes host filter info
type HostFilterTopoNodes struct {
	Alive         *int   //  alive：0为Agent异常, 1为Agent正常, nil 则不筛选
	SearchContent string // 模糊搜索内容（支持同时对主机IP/主机名/操作系统/云区域名称进行模糊搜索）
}

// FilterHostByCondition xxx
func (h *HostFilterTopoNodes) FilterHostByCondition(source []HostDetailInfo) []HostDetailInfo {
	filterResult := make([]HostDetailInfo, 0)
	for _, host := range source {
		if h.Alive != nil && host.Alive != *h.Alive {
			continue
		}

		if h.SearchContent != "" &&
			(!utils.StringContainInSlice(h.SearchContent,
				[]string{host.Ip, host.Ipv6, host.HostName, host.OsName, host.CloudArea.Name})) {
			continue
		}

		filterResult = append(filterResult, host)
	}

	return filterResult
}

// HostFilterCheckNodes host filter when check nodes
type HostFilterCheckNodes struct {
	IpList   []string // ipv4地址 支持如: 1.2.3.4, 0:1.2.3.4 格式
	Ipv6List []string // ipv6地址 支持如: x:x:x:x:x:x:x:x, 0:x:x:x:x:x:x:x:x 格式
	KeyList  []string // KeyList 支持主机名称模糊过滤
}

// FilterHostByCondition xxx
func (h *HostFilterCheckNodes) FilterHostByCondition(source []HostDetailInfo) []HostDetailInfo {
	filterResult := make([]HostDetailInfo, 0)
	for _, host := range source {
		if utils.StringInSlice(host.Ip, h.IpList) ||
			utils.StringInSlice(getCloudIPAddress(host.CloudArea.ID, host.Ip), h.IpList) ||
			utils.StringInSlice(host.Ipv6, h.Ipv6List) ||
			utils.StringInSlice(getCloudIPAddress(host.CloudArea.ID, host.Ipv6), h.Ipv6List) ||
			utils.SliceContainInString(h.KeyList, host.HostName) {
			filterResult = append(filterResult, host)
		}
	}

	return filterResult
}

func getCloudIPAddress(cloudID int64, ip string) string {
	return fmt.Sprintf("%v:%s", cloudID, ip)
}
