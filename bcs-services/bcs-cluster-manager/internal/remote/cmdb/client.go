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

// Package cmdb xxx
package cmdb

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/i18n"
	"github.com/parnurzeal/gorequest"
	"github.com/patrickmn/go-cache"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/types"
	rutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// CmdbClient global cmdb client
var CmdbClient *Client

// SetCmdbClient set cmdb client
func SetCmdbClient(options Options) error {
	// init client
	cli, err := NewCmdbClient(options)
	if err != nil {
		return err
	}

	CmdbClient = cli
	// auth.SetCheckBizPerm(checkUserBizPerm)
	return nil
}

// GetCmdbClient get cmdb client
func GetCmdbClient() *Client {
	return CmdbClient
}

// 检查业务下的角色用户是否存在
func checkUserBizPerm(ctx context.Context, username string, businessID string) (bool, error) {
	if businessID == "" {
		return false, errors.New("permission denied")
	}
	bizID, err := strconv.Atoi(businessID)
	if err != nil {
		errMsg := fmt.Errorf("strconv BusinessID to int failed: %v", err)
		blog.Errorf(errMsg.Error())
		return false, errMsg
	}

	// query biz hosts
	businessData, err := GetCmdbClient().GetBusinessMaintainer(ctx, bizID)
	if err != nil {
		blog.Errorf("getUserHasPermHosts GetBusinessMaintainer failed: %v", err)
		return false, err
	}
	var userList []string
	userList = append(userList, strings.Split(businessData.BKBizMaintainer, ",")...)
	userList = append(userList, strings.Split(businessData.BkBizProductor, ",")...)
	userList = append(userList, strings.Split(businessData.BkBizTester, ",")...)
	userList = append(userList, strings.Split(businessData.BkBizDeveloper, ",")...)
	userList = append(userList, strings.Split(businessData.Operator, ",")...)
	if utils.StringInSlice(username, userList) {
		return true, nil
	}
	return false, errors.New("permission denied, need scope of blueking bkcc business permition:" +
		"[bk_biz_maintainer|bk_biz_productor|bk_biz_tester|bk_biz_developer]")
}

// NewCmdbClient create cmdb client
func NewCmdbClient(options Options) (*Client, error) {
	c := &Client{
		appCode:     options.AppCode,
		appSecret:   options.AppSecret,
		bkUserName:  options.BKUserName,
		server:      options.Server,
		serverDebug: options.Debug,
		// Create a cache with a default expiration time of 10 minutes, and which
		// purges expired items every 1 hour
		cache: cache.New(3*time.Minute, 60*time.Minute),
	}

	// disable cmdb
	if !options.Enable {
		return nil, nil
	}

	return c, nil
}

var (
	// defaultTimeOut default timeout
	defaultTimeOut = time.Second * 60
	// ErrServerNotInit server not init
	ErrServerNotInit = errors.New("server not inited")
)

// Options for cmdb client
type Options struct {
	// Enable enable client
	Enable bool
	// AppCode app code
	AppCode string
	// AppSecret app secret
	AppSecret string
	// BKUserName bk username
	BKUserName string
	// Server server
	Server string
	// Debug debug
	Debug bool
}

// AuthInfo auth user
type AuthInfo struct {
	// BkAppCode bk app code
	BkAppCode string `json:"bk_app_code"`
	// BkAppSecret bk app secret
	BkAppSecret string `json:"bk_app_secret"`
	// BkUserName bk username
	BkUserName string `json:"bk_username"`
}

// Client for cc
type Client struct {
	appCode     string
	appSecret   string
	bkUserName  string
	server      string
	serverDebug bool
	cache       *cache.Cache
}

// FetchAllHostTopoRelationsByBizID fetch biz topo
func (c *Client) FetchAllHostTopoRelationsByBizID(ctx context.Context, bizID int) ([]HostTopoRelation, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	// get hostTopo from cache
	hostTopo, ok := GetBizHostTopoData(c.cache, bizID)
	if ok {
		blog.Infof("FetchAllHostTopoRelationsByBizID hit cache by bizID[%s]", bizID)
		return hostTopo, nil
	}
	blog.V(3).Infof("FetchAllHostTopoRelationsByBizID miss cache by bizID[%v]", bizID)

	// get all hostTopo counts
	counts, _, err := c.FindHostTopoRelation(ctx, bizID, Page{
		Start: 0,
		Limit: 1,
	})
	if err != nil {
		return nil, err
	}
	blog.Infof("FetchAllHostTopoRelationsByBizID count %d by bizID %d", counts, bizID)

	// split count to pages
	pages := splitCountToPage(counts, MaxLimits)
	var (
		hostTopoList = make([]HostTopoRelation, 0)
		hostLock     = &sync.RWMutex{}
	)

	con := utils.NewRoutinePool(20)
	defer con.Close()

	// routine pool handle
	for i := range pages {
		con.Add(1)
		go func(page Page) {
			defer con.Done()
			// find host topos
			_, hostTopos, errLocal := c.FindHostTopoRelation(ctx, bizID, page) // nolint
			if errLocal != nil {
				blog.Errorf("cmdb client FindHostTopoRelation %v failed, %s", bizID, err.Error())
				return
			}
			hostLock.Lock()
			hostTopoList = append(hostTopoList, hostTopos...)
			hostLock.Unlock()
		}(pages[i])
	}
	con.Wait()

	blog.Infof("FetchAllHostsByBizID successful %v", bizID)

	// set biz topo cache
	err = SetBizHostTopoData(c.cache, bizID, hostTopoList)
	if err != nil {
		blog.Errorf("FetchAllHostTopoRelationsByBizID[%v] SetBizHostTopoData failed: %v", bizID, err)
	}
	return hostTopoList, nil
}

// FetchAllHostsByBizID get allHosts by bizID
func (c *Client) FetchAllHostsByBizID(ctx context.Context, bizID int, cache bool) ([]HostData, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	// get bizHostData from cache
	if cache {
		hostData, ok := GetBizHostData(c.cache, bizID)
		if ok {
			blog.Infof("FetchAllHostsByBizID hit cache by bizID[%s]", bizID)
			return hostData, nil
		}
		blog.V(3).Infof("FetchAllHostsByBizID miss cache by bizID[%v]", bizID)
	}

	// get all host counts
	counts, _, err := c.QueryHostByBizID(ctx, bizID, Page{
		Start: 0,
		Limit: 1,
	})
	if err != nil {
		return nil, err
	}

	blog.Infof("FetchAllHostsByBizID count %d by bizID %d", counts, bizID)
	pages := splitCountToPage(counts, MaxLimits)
	var (
		hostList = make([]HostData, 0)
		hostLock = &sync.RWMutex{}
	)

	con := utils.NewRoutinePool(20)
	defer con.Close()

	// routine pool handle
	for i := range pages {
		con.Add(1)
		go func(page Page) {
			defer con.Done()
			// query hosts by biz
			_, hosts, err := c.QueryHostByBizID(ctx, bizID, page) // nolint
			if err != nil {
				blog.Errorf("cmdb client QueryHostByBizID %v failed, %s", bizID, err.Error())
				return
			}
			hostLock.Lock()
			hostList = append(hostList, hosts...)
			hostLock.Unlock()
		}(pages[i])
	}
	con.Wait()

	blog.Infof("FetchAllHostsByBizID successful %v", bizID)
	if len(hostList) == 0 {
		return nil, fmt.Errorf("FetchAllHostsByBizID[%d] failed: imageIDList empty", bizID)
	}

	// set biz hosts cache
	err = SetBizHostData(c.cache, bizID, hostList)
	if err != nil {
		blog.Errorf("FetchAllHostsByBizID[%v] SetBizHostData failed: %v", bizID, err)
	}

	return hostList, nil
}

// QueryAllHostInfoWithoutBiz get all host info by ips
func (c *Client) QueryAllHostInfoWithoutBiz(ctx context.Context, ips []string) ([]HostDetailData, error) {
	chunk := utils.SplitStringsChunks(ips, MaxLimits)
	list := make([]HostDetailData, 0)

	for _, v := range chunk {
		// query hostInfoIp without biz
		data, err := c.QueryHostInfoWithoutBiz(ctx, FieldHostIP, v, Page{Start: 0, Limit: MaxLimits})
		if err != nil {
			return nil, err
		}
		list = append(list, data...)
	}
	return list, nil
}

// QueryAllHostInfoByAssetIdWithoutBiz get all host info by assetIds
func (c *Client) QueryAllHostInfoByAssetIdWithoutBiz(ctx context.Context, assetIds []string) ([]HostDetailData, error) {
	chunk := utils.SplitStringsChunks(assetIds, MaxLimits)
	list := make([]HostDetailData, 0)
	for _, v := range chunk {
		// // query hostInfoAssertId without biz
		data, err := c.QueryHostInfoWithoutBiz(ctx, FieldAssetId, v, Page{Start: 0, Limit: MaxLimits})
		if err != nil {
			return nil, err
		}
		list = append(list, data...)
	}
	return list, nil
}

// ListTopology list topology
func (c *Client) ListTopology(ctx context.Context, bizID int, filterInter bool, cache bool) (
	*SearchBizInstTopoData, error) {
	// get bizTopoData from cache
	if cache {
		bizTopo, ok := GetBizTopoData(c.cache, bizID)
		if ok {
			blog.Infof("ListTopology hit cache by bizID[%s]", bizID)
			return bizTopo, nil
		}
		blog.V(3).Infof("ListTopology miss cache by bizID[%v]", bizID)
	}

	// get biz internal module
	internalModules, err := c.GetBizInternalModule(ctx, bizID)
	if err != nil {
		return nil, err
	}
	// internalModules.ReplaceName()

	// search biz inst topo
	topos, err := c.SearchBizInstTopo(ctx, bizID)
	if err != nil {
		return nil, err
	}
	var topo *SearchBizInstTopoData
	for i := range topos {
		if topos[i].BKInstID == bizID {
			topo = &topos[i]
		}
	}
	if topo == nil {
		return nil, fmt.Errorf("topology is empty")
	}
	childs := make([]SearchBizInstTopoData, 0)

	if !filterInter {
		child := SearchBizInstTopoData{
			BKInstID:   internalModules.SetID,
			BKInstName: internalModules.SetName,
			BKObjID:    "set",
			BKObjName:  "set",
			Child:      make([]SearchBizInstTopoData, 0),
		}
		for _, v := range internalModules.ModuleInfo {
			child.Child = append(child.Child, SearchBizInstTopoData{
				BKInstID:   v.ModuleID,
				BKInstName: v.ModuleName,
				BKObjID:    "module",
				BKObjName:  "module",
				Child:      make([]SearchBizInstTopoData, 0),
			})
		}
		childs = append(childs, child)
	}

	childs = append(childs, topo.Child...)
	topo.Child = childs

	if cache {
		err = SetBizTopoData(c.cache, bizID, topo)
		if err != nil {
			blog.Errorf("ListTopology[%v] SetBizTopoData failed: %v", bizID, err)
		}
	}

	return topo, nil
}

// QueryHostInfoWithoutBiz query values string host info by field 没有业务ID的主机查询
func (c *Client) QueryHostInfoWithoutBiz(ctx context.Context, field string, values []string,
	page Page) ([]HostDetailData, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	// list_hosts_without_biz
	var (
		reqURL  = fmt.Sprintf("%s/api/v3/hosts/list_hosts_without_app", c.server)
		request = &ListHostsWithoutBizRequest{
			Page:               page,
			HostPropertyFilter: buildFilterConditionByStrValues(field, values),
			Fields:             fieldHostDetailInfo,
		}
		respData = &ListHostsWithoutBizResponse{}
	)

	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return nil, err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		SetDebug(c.serverDebug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("cmdb", "QueryHostInfoWithoutBiz", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api QueryHostInfoWithoutBiz failed: %v", errs[0])
		return nil, errs[0]
	}
	metrics.ReportLibRequestMetric("cmdb", "QueryHostInfoWithoutBiz", "http", metrics.LibCallStatusOK, start)

	if !respData.Result {
		blog.Errorf("call api QueryHostInfoWithoutBiz failed: %v", respData.Message)
		return nil, errors.New(respData.Message)
	}
	// successfully request
	blog.Infof("call api QueryHostInfoWithoutBiz with url(%s) successfully", reqURL)

	return respData.Data.Info, nil
}

// FindHostTopoRelation find host topo 获取主机拓扑关系
func (c *Client) FindHostTopoRelation(ctx context.Context, bizID int, page Page) (int, []HostTopoRelation, error) {
	if c == nil {
		return 0, nil, ErrServerNotInit
	}

	// find_host_topo_relation
	var (
		reqURL  = fmt.Sprintf("%s/api/v3/host/topo/relation/read", c.server)
		request = &HostTopoRelationReq{
			Page:    page,
			BkBizID: bizID,
		}
		respData = &HostTopoRelationResp{}
	)

	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return 0, nil, err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		SetDebug(c.serverDebug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("cmdb", "FindHostTopoRelation", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api FindHostTopoRelation failed: %v", errs[0])
		return 0, nil, errs[0]
	}
	metrics.ReportLibRequestMetric("cmdb", "FindHostTopoRelation", "http", metrics.LibCallStatusOK, start)

	if !respData.Result {
		blog.Errorf("call api FindHostTopoRelation failed: %v", respData.Message)
		return 0, nil, errors.New(respData.Message)
	}
	// successfully request
	blog.Infof("call api FindHostTopoRelation with url(%s) successfully", reqURL)

	if len(respData.Data.Data) > 0 {
		return respData.Data.Count, respData.Data.Data, nil
	}

	return 0, nil, fmt.Errorf("call api FindHostTopoRelation failed")
}

// SearchCloudAreaByCloudID search cloudArea info by cloudID 查询管控区域
func (c *Client) SearchCloudAreaByCloudID(ctx context.Context, cloudID int) (*SearchCloudAreaInfo, error) {
	cloudData, ok := GetCloudData(c.cache, cloudID)
	if ok {
		blog.Infof("SearchCloudAreaByCloudID hit cache by cloudID[%d]", cloudID)
		return cloudData, nil
	}
	blog.V(3).Infof("SearchCloudAreaByCloudID miss cache by cloudID[%v]", cloudID)

	// search_cloud_area
	var (
		reqURL  = fmt.Sprintf("%s/api/v3/findmany/cloudarea", c.server)
		request = &SearchCloudAreaRequest{
			Page: Page{
				Start: 0,
				Limit: MaxLimits,
			},
			Condition: BuildCloudAreaCondition(cloudID),
		}
		respData = &SearchCloudAreaResp{}
	)

	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return nil, err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		SetDebug(c.serverDebug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("cmdb", "SearchCloudAreaByCloudID", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api SearchCloudAreaByCloudID failed: %v", errs[0])
		return nil, errs[0]
	}
	metrics.ReportLibRequestMetric("cmdb", "SearchCloudAreaByCloudID", "http", metrics.LibCallStatusOK, start)

	if !respData.Result {
		blog.Errorf("call api SearchCloudAreaByCloudID failed: %v", respData.Message)
		return nil, errors.New(respData.Message)
	}

	// successfully request
	blog.Infof("call api SearchCloudAreaByCloudID with url(%s) successfully", reqURL)

	if len(respData.Data.Info) == 0 {
		return nil, fmt.Errorf("SearchCloudAreaByCloudID not exist %v", cloudID)
	}

	err = SetCloudData(c.cache, cloudID, respData.Data.Info[0])
	if err != nil {
		blog.Errorf("SearchCloudAreaByCloudID[%v] SetCloudData failed: %v", cloudID, err)
	}

	return respData.Data.Info[0], nil
}

// QueryHostByBizID query host by bizID 查询业务下主机
func (c *Client) QueryHostByBizID(ctx context.Context, bizID int, page Page) (int, []HostData, error) {
	if c == nil {
		return 0, nil, ErrServerNotInit
	}

	// list_biz_hosts
	var (
		reqURL  = fmt.Sprintf("%s/api/v3/hosts/app/%v/list_hosts", c.server, bizID)
		request = &ListBizHostRequest{
			Page:    page,
			BKBizID: bizID,
			Fields:  fieldHostIPSelectorInfo,
		}
		respData = &ListBizHostsResponse{}
	)

	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return 0, nil, err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		SetDebug(c.serverDebug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("cmdb", "QueryHostByBizID", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api QueryHostNumByBizID failed: %v", errs[0])
		return 0, nil, errs[0]
	}
	metrics.ReportLibRequestMetric("cmdb", "QueryHostByBizID", "http", metrics.LibCallStatusOK, start)

	if !respData.Result {
		blog.Errorf("call api QueryHostNumByBizID failed: %v", respData.Message)
		return 0, nil, errors.New(respData.Message)
	}
	// successfully request
	blog.Infof("call api QueryHostNumByBizID with url(%s) successfully", reqURL)

	if len(respData.Data.Info) > 0 {
		return respData.Data.Count, respData.Data.Info, nil
	}

	return 0, nil, fmt.Errorf("call api GetBS2IDByBizID failed")
}

// FindHostBizRelations query host biz relations by hostID 查询主机业务关系信息
func (c *Client) FindHostBizRelations(ctx context.Context, hostID []int) ([]HostBizRelations, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	// find_host_biz_relations
	var (
		reqURL  = fmt.Sprintf("%s/api/v3/hosts/modules/read", c.server)
		request = &FindHostBizRelationsRequest{
			BkHostID: hostID,
		}
		respData = &FindHostBizRelationsResponse{}
	)

	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return nil, err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		SetDebug(c.serverDebug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("cmdb", "FindHostBizRelations", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api FindHostBizRelations failed: %v", errs[0])
		return nil, errs[0]
	}
	metrics.ReportLibRequestMetric("cmdb", "FindHostBizRelations", "http", metrics.LibCallStatusOK, start)

	if !respData.Result {
		blog.Errorf("call api FindHostBizRelations failed: %v", respData.Message)
		return nil, errors.New(respData.Message)
	}
	// successfully request
	blog.Infof("call api FindHostBizRelations with url(%s) successfully", reqURL)

	return respData.Data, nil
}

// GetBizInternalModule get biz recycle module info 查询业务的空闲机/故障机/待回收模块
func (c *Client) GetBizInternalModule(ctx context.Context, bizID int) (*BizInternalModuleData, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	language := i18n.LanguageFromCtx(ctx)
	blog.Infof("cmdb client GetBizInternalModule language %s", language)

	// get_biz_internal_module
	var (
		reqURL   = fmt.Sprintf("%s/api/v3/topo/internal/default/%v", c.server, bizID)
		respData = &QueryBizInternalModuleResponse{}
	)

	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return nil, err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Get(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("Blueking-Language", language).
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		SetDebug(c.serverDebug).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("cmdb", "GetBizInternalModule", "http", metrics.LibCallStatusErr, start)
		return nil, errs[0]
	}
	metrics.ReportLibRequestMetric("cmdb", "GetBizInternalModule", "http", metrics.LibCallStatusOK, start)

	if !respData.Result {
		return nil, errors.New(respData.Message)
	}

	return &respData.Data, nil
}

// TransHostAcrossBiz trans host to other biz 跨业务转移主机
func (c *Client) TransHostAcrossBiz(ctx context.Context, hostInfo TransHostAcrossBizInfo) error {
	if c == nil {
		return ErrServerNotInit
	}

	// transfer_host_across_biz
	var (
		reqURL  = fmt.Sprintf("%s/api/v3/hosts/modules/across/biz", c.server)
		request = &TransferHostAcrossBizRequest{
			SrcBizID:   hostInfo.SrcBizID,
			BkHostID:   hostInfo.HostID,
			DstBizID:   hostInfo.DstBizID,
			BkModuleID: hostInfo.DstModuleID,
		}
		respData = &TransferHostAcrossBizResponse{}
	)

	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		SetDebug(c.serverDebug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("cmdb", "TransHostAcrossBiz", "http", metrics.LibCallStatusErr, start)
		return errs[0]
	}
	metrics.ReportLibRequestMetric("cmdb", "TransHostAcrossBiz", "http", metrics.LibCallStatusOK, start)

	if !respData.Result {
		return errors.New(respData.Message)
	}

	return nil
}

// GetBusinessMaintainer get maintainers by bizID 查询业务
func (c *Client) GetBusinessMaintainer(ctx context.Context, bizID int) (*BusinessData, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	// search_business
	var (
		reqURL  = fmt.Sprintf("%s/api/v3/biz/search/default", c.server)
		request = &SearchBusinessRequest{
			Fields: []string{},
			Condition: map[string]interface{}{
				conditionBkBizID: bizID,
			},
		}
		respData = &SearchBusinessResponse{}
	)

	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return nil, err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		SetDebug(c.serverDebug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("cmdb", "GetBusinessMaintainer", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api GetBS2IDByBizID failed: %v", errs[0])
		return nil, errs[0]
	}
	metrics.ReportLibRequestMetric("cmdb", "GetBusinessMaintainer", "http", metrics.LibCallStatusOK, start)

	if !respData.Result {
		blog.Errorf("call api GetBS2IDByBizID failed: %v", respData.Message)
		return nil, errors.New(respData.Message)
	}
	// successfully request
	blog.Infof("call api GetBS2IDByBizID with url(%s) successfully", reqURL)

	if len(respData.Data.Info) > 0 {
		return &respData.Data.Info[0], nil
	}

	return nil, fmt.Errorf("call api GetBS2IDByBizID failed")
}

// GetBS2IDByBizID get bs2ID by bizID 查询bkcc对应公司的二级业务ID
func (c *Client) GetBS2IDByBizID(ctx context.Context, bizID int64) (int, error) {
	if c == nil {
		return 0, ErrServerNotInit
	}

	// search_business
	var (
		reqURL  = fmt.Sprintf("%s/api/v3/biz/search/default", c.server)
		request = &SearchBusinessRequest{
			Fields: []string{fieldBS2NameID},
			Condition: map[string]interface{}{
				conditionBkBizID: bizID,
			},
		}
		respData = &SearchBusinessResponse{}
	)

	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return 0, err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		SetDebug(c.serverDebug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("cmdb", "GetBS2IDByBizID", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api GetBS2IDByBizID failed: %v", errs[0])
		return 0, errs[0]
	}
	metrics.ReportLibRequestMetric("cmdb", "GetBS2IDByBizID", "http", metrics.LibCallStatusOK, start)

	if !respData.Result {
		blog.Errorf("call api GetBS2IDByBizID failed: %v", respData.Message)
		return 0, errors.New(respData.Message)
	}
	// successfully request
	blog.Infof("call api GetBS2IDByBizID with url(%s) successfully", reqURL)

	if len(respData.Data.Info) > 0 {
		return respData.Data.Info[0].BS2NameID, nil
	}

	return 0, fmt.Errorf("call api GetBS2IDByBizID failed")
}

// TransHostToRecycleModule trans host to recycleModule 上交主机到业务的待回收模块
func (c *Client) TransHostToRecycleModule(ctx context.Context, bizID int, hostID []int) error {
	if c == nil {
		return ErrServerNotInit
	}

	// transfer_host_to_recyclemodule
	var (
		reqURL  = fmt.Sprintf("%s/api/v3/hosts/modules/recycle", c.server)
		request = &TransHostToERecycleModuleRequest{
			BkBizID:  bizID,
			BkHostID: hostID,
		}
		respData = &TransHostToERecycleModuleResponse{}
	)

	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		SetDebug(c.serverDebug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("cmdb", "TransHostToRecycleModule", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api TransHostToRecycleModule failed: %v", errs[0])
		return errs[0]
	}
	metrics.ReportLibRequestMetric("cmdb", "TransHostToRecycleModule", "http", metrics.LibCallStatusOK, start)

	if !respData.Result {
		blog.Errorf("call api TransHostToRecycleModule failed: %v", respData.Message)
		return errors.New(respData.Message)
	}
	// successfully request
	blog.Infof("call api TransHostToRecycleModule with url(%s) successfully", reqURL)

	return nil
}

// TransferHostToIdleModule transfer host to idle module 上交主机到业务的空闲机模块
func (c *Client) TransferHostToIdleModule(ctx context.Context, bizID int, hostID []int) error {
	// transfer_host_to_idlemodule
	var (
		reqURL  = fmt.Sprintf("%s/api/v3/hosts/modules/idle", c.server)
		request = &TransferHostToIdleModuleRequest{
			BkBizID:  bizID,
			BkHostID: hostID,
		}
		respData = &TransferHostToIdleModuleResponse{}
	)

	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		SetDebug(c.serverDebug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("cmdb", "TransferHostToIdleModule", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api TransferHostToIdleModule failed: %v", errs[0])
		return errs[0]
	}
	metrics.ReportLibRequestMetric("cmdb", "TransferHostToIdleModule", "http", metrics.LibCallStatusOK, start)

	if !respData.Result || respData.Code != 0 {
		blog.Errorf("call api TransferHostToIdleModule failed: %v", respData)
		return fmt.Errorf("call api TransferHostToIdleModule failed: %v", respData)
	}
	blog.Infof("call api TransferHostToIdleModule with url(%s) successfully", reqURL)

	return nil
}

// TransferHostToResourceModule transfer host to resource module 上交主机至资源池
func (c *Client) TransferHostToResourceModule(ctx context.Context, bizID int, hostID []int) error {
	// transfer_host_to_resourcemodule
	var (
		reqURL  = fmt.Sprintf("%s/api/v3/hosts/modules/resource", c.server)
		request = &TransferHostToResourceModuleRequest{
			BkBizID:  bizID,
			BkHostID: hostID,
		}
		respData = &TransferHostToResourceModuleResponse{}
	)

	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		SetDebug(c.serverDebug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("cmdb", "TransferHostToResourceModule", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api TransferHostToResourceModule failed: %v", errs[0])
		return errs[0]
	}
	metrics.ReportLibRequestMetric("cmdb", "TransferHostToResourceModule", "http", metrics.LibCallStatusOK, start)

	if !respData.Result || respData.Code != 0 {
		blog.Errorf("call api TransferHostToResourceModule failed: %v", respData)
		return fmt.Errorf("call api TransferHostToResourceModule failed: %v", respData)
	}
	blog.Infof("call api TransferHostToResourceModule with url(%s) successfully", reqURL)

	return nil
}

// DeleteHost delete host 删除主机
func (c *Client) DeleteHost(ctx context.Context, hostID []int) error {
	hostIDs := []string{}
	for _, v := range hostID {
		hostIDs = append(hostIDs, strconv.Itoa(v))
	}

	// delete_host
	var (
		reqURL  = fmt.Sprintf("%s/api/v3/hosts/batch", c.server)
		request = &DeleteHostRequest{
			BkHostID: strings.Join(hostIDs, ","),
		}
		respData = &DeleteHostResponse{}
	)

	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Delete(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		SetDebug(c.serverDebug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("cmdb", "DeleteHost", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api DeleteHost failed: %v", errs[0])
		return errs[0]
	}
	metrics.ReportLibRequestMetric("cmdb", "DeleteHost", "http", metrics.LibCallStatusOK, start)

	if !respData.Result || respData.Code != 0 {
		blog.Errorf("call api DeleteHost failed: %v", respData)
		return fmt.Errorf("call api DeleteHost failed: %v", respData)
	}
	blog.Infof("call api DeleteHost with url(%s) successfully", reqURL)

	return nil
}

// SearchBizInstTopo search biz inst topo 查询业务实例拓扑
func (c *Client) SearchBizInstTopo(ctx context.Context, bizID int) ([]SearchBizInstTopoData, error) {
	// search_biz_inst_topo
	var (
		reqURL   = fmt.Sprintf("%s/api/v3/find/topoinst/biz/%d", c.server, bizID)
		respData = &SearchBizInstTopoResponse{}
	)

	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return nil, err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		SetDebug(c.serverDebug).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("cmdb", "SearchBizInstTopo", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api SearchBizInstTopo failed: %v", errs[0])
		return nil, errs[0]
	}
	metrics.ReportLibRequestMetric("cmdb", "SearchBizInstTopo", "http", metrics.LibCallStatusOK, start)

	if !respData.Result || respData.Code != 0 {
		blog.Errorf("call api SearchBizInstTopo failed: %v", respData)
		return nil, fmt.Errorf("call api SearchBizInstTopo failed: %v", respData)
	}
	blog.Infof("call api SearchBizInstTopo with url(%s) successfully", reqURL)

	return respData.Data, nil
}

// TransferHostModule transfer host to module 业务内主机转移模块
func (c *Client) TransferHostModule(ctx context.Context, bizID int, hostID []int,
	moduleID []int, isIncrement bool) error {
	// transfer_host_module
	var (
		reqURL  = fmt.Sprintf("%s/api/v3/hosts/modules", c.server)
		request = &TransferHostModuleRequest{
			BKBizID:     bizID,
			BKHostID:    hostID,
			BKModuleID:  moduleID,
			IsIncrement: isIncrement,
		}
		respData = &BaseResponse{}
	)

	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		SetDebug(c.serverDebug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("cmdb", "TransferHostModule", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api TransferHostModule failed: %v", errs[0])
		return errs[0]
	}
	metrics.ReportLibRequestMetric("cmdb", "TransferHostModule", "http", metrics.LibCallStatusOK, start)

	if !respData.Result || respData.Code != 0 {
		blog.Errorf("call api TransferHostModule failed: %v", respData)
		return fmt.Errorf("call api TransferHostModule failed: %v", respData)
	}
	blog.Infof("call api TransferHostModule with url(%s) successfully", reqURL)

	return nil
}

// AddHostFromCmpy add host from cmpy
func (c *Client) AddHostFromCmpy(ctx context.Context, svrIds []string, ips []string, assetIds []string) error {
	if c == nil {
		return ErrServerNotInit
	}

	// add_host_from_cmpy
	var (
		reqURL  = fmt.Sprintf("%s/api/c/compapi/v2/cc/add_host_from_cmpy/", c.server)
		request = &AddHostFromCmpyReq{
			SvrIds:   svrIds,
			AssetIds: assetIds,
			InnerIps: ips,
		}
		respData = &AddHostFromCmpyResp{}
	)

	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		SetDebug(c.serverDebug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("cmdb", "AddHostFromCmpy", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api AddHostFromCmpy failed: %v", errs[0])
		return errs[0]
	}
	metrics.ReportLibRequestMetric("cmdb", "AddHostFromCmpy", "http", metrics.LibCallStatusOK, start)

	if !respData.Result {
		blog.Errorf("call api AddHostFromCmpy failed: %v", respData.Message)
		return errors.New(respData.Message)
	}
	// successfully request
	blog.Infof("call api AddHostFromCmpy with url(%s) successfully", reqURL)

	return nil
}

// SyncHostInfoFromCmpy update host info from cmpy
func (c *Client) SyncHostInfoFromCmpy(ctx context.Context, bkCloudId int, bkHostIds []int64) error {
	if c == nil {
		return ErrServerNotInit
	}

	// sync_host_info_from_cmpy
	var (
		reqURL  = fmt.Sprintf("%s/api/c/compapi/v2/cc/sync_host_info_from_cmpy/", c.server)
		request = &SyncHostInfoFromCmpyReq{
			BkHostIds: bkHostIds,
			BkCloudId: bkCloudId,
		}

		respData = &SyncHostInfoFromCmpyResp{}
	)

	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		SetDebug(c.serverDebug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("cmdb", "SyncHostInfoFromCmpy", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api SyncHostInfoFromCmpy failed: %v", errs[0])
		return errs[0]
	}
	metrics.ReportLibRequestMetric("cmdb", "SyncHostInfoFromCmpy", "http", metrics.LibCallStatusOK, start)

	if !respData.Result {
		blog.Errorf("call api SyncHostInfoFromCmpy failed: %v", respData.Message)
		return errors.New(respData.Message)
	}
	blog.Infof("call api SyncHostInfoFromCmpy with url(%s) successfully", reqURL)

	return nil
}
