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
	"encoding/json"
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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
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
	auth.SetCheckBizPerm(checkUserBizPerm)
	return nil
}

// GetCmdbClient get cmdb client
func GetCmdbClient() *Client {
	return CmdbClient
}

// 检查业务下的角色用户是否存在
func checkUserBizPerm(username string, businessID string) (bool, error) {
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
	businessData, err := GetCmdbClient().GetBusinessMaintainer(bizID)
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

	// gateway auth
	auth, err := c.generateGateWayAuth()
	if err != nil {
		return nil, err
	}
	c.userAuth = auth
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
	userAuth    string
	cache       *cache.Cache
}

// generateGateWayAuth generate gateway auth
func (c *Client) generateGateWayAuth() (string, error) {
	if c == nil {
		return "", ErrServerNotInit
	}

	// auth info
	auth := &AuthInfo{
		BkAppCode:   c.appCode,
		BkAppSecret: c.appSecret,
		BkUserName:  c.bkUserName,
	}

	userAuth, err := json.Marshal(auth)
	if err != nil {
		return "", err
	}

	return string(userAuth), nil
}

// FetchAllHostTopoRelationsByBizID fetch biz topo
func (c *Client) FetchAllHostTopoRelationsByBizID(bizID int) ([]HostTopoRelation, error) {
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
	counts, _, err := c.FindHostTopoRelation(bizID, Page{
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
			_, hostTopos, errLocal := c.FindHostTopoRelation(bizID, page) // nolint
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
func (c *Client) FetchAllHostsByBizID(bizID int, cache bool) ([]HostData, error) {
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
	counts, _, err := c.QueryHostByBizID(bizID, Page{
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
			_, hosts, err := c.QueryHostByBizID(bizID, page) // nolint
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

// QueryHostInfoWithoutBiz query values string host info by field
func (c *Client) QueryHostInfoWithoutBiz(field string, values []string, page Page) ([]HostDetailData, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	// list_hosts_without_biz
	var (
		reqURL  = fmt.Sprintf("%s/api/c/compapi/v2/cc/list_hosts_without_biz/", c.server)
		request = &ListHostsWithoutBizRequest{
			Page:               page,
			HostPropertyFilter: buildFilterConditionByStrValues(field, values),
			Fields:             fieldHostDetailInfo,
		}
		respData = &ListHostsWithoutBizResponse{}
	)

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
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

// QueryAllHostInfoWithoutBiz get all host info by ips
func (c *Client) QueryAllHostInfoWithoutBiz(ips []string) ([]HostDetailData, error) {
	chunk := utils.SplitStringsChunks(ips, MaxLimits)
	list := make([]HostDetailData, 0)

	for _, v := range chunk {
		// query hostInfoIp without biz
		data, err := c.QueryHostInfoWithoutBiz(FieldHostIP, v, Page{Start: 0, Limit: MaxLimits})
		if err != nil {
			return nil, err
		}
		list = append(list, data...)
	}
	return list, nil
}

// QueryAllHostInfoByAssetIdWithoutBiz get all host info by assetIds
func (c *Client) QueryAllHostInfoByAssetIdWithoutBiz(assetIds []string) ([]HostDetailData, error) {
	chunk := utils.SplitStringsChunks(assetIds, MaxLimits)
	list := make([]HostDetailData, 0)
	for _, v := range chunk {
		// // query hostInfoAssertId without biz
		data, err := c.QueryHostInfoWithoutBiz(FieldAssetId, v, Page{Start: 0, Limit: MaxLimits})
		if err != nil {
			return nil, err
		}
		list = append(list, data...)
	}
	return list, nil
}

// FindHostTopoRelation find host topo
func (c *Client) FindHostTopoRelation(bizID int, page Page) (int, []HostTopoRelation, error) {
	if c == nil {
		return 0, nil, ErrServerNotInit
	}

	// find_host_topo_relation
	var (
		reqURL  = fmt.Sprintf("%s/api/c/compapi/v2/cc/find_host_topo_relation/", c.server)
		request = &HostTopoRelationReq{
			Page:    page,
			BkBizID: bizID,
		}
		respData = &HostTopoRelationResp{}
	)

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
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

// SearchCloudAreaByCloudID search cloudArea info by cloudID
func (c *Client) SearchCloudAreaByCloudID(cloudID int) (*SearchCloudAreaInfo, error) {
	cloudData, ok := GetCloudData(c.cache, cloudID)
	if ok {
		blog.Infof("SearchCloudAreaByCloudID hit cache by cloudID[%d]", cloudID)
		return cloudData, nil
	}
	blog.V(3).Infof("SearchCloudAreaByCloudID miss cache by cloudID[%v]", cloudID)

	// search_cloud_area
	var (
		reqURL  = fmt.Sprintf("%s/api/c/compapi/v2/cc/search_cloud_area/", c.server)
		request = &SearchCloudAreaRequest{
			Page: Page{
				Start: 0,
				Limit: MaxLimits,
			},
			Condition: BuildCloudAreaCondition(cloudID),
		}
		respData = &SearchCloudAreaResp{}
	)

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
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

	err := SetCloudData(c.cache, cloudID, respData.Data.Info[0])
	if err != nil {
		blog.Errorf("SearchCloudAreaByCloudID[%v] SetCloudData failed: %v", cloudID, err)
	}

	return respData.Data.Info[0], nil
}

// QueryHostByBizID query host by bizID
func (c *Client) QueryHostByBizID(bizID int, page Page) (int, []HostData, error) {
	if c == nil {
		return 0, nil, ErrServerNotInit
	}

	// list_biz_hosts
	var (
		reqURL  = fmt.Sprintf("%s/api/c/compapi/v2/cc/list_biz_hosts/", c.server)
		request = &ListBizHostRequest{
			Page:    page,
			BKBizID: bizID,
			Fields:  fieldHostIPSelectorInfo,
		}
		respData = &ListBizHostsResponse{}
	)

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
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

// FindHostBizRelations query host biz relations by hostID
func (c *Client) FindHostBizRelations(hostID []int) ([]HostBizRelations, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	// find_host_biz_relations
	var (
		reqURL  = fmt.Sprintf("%s/api/c/compapi/v2/cc/find_host_biz_relations/", c.server)
		request = &FindHostBizRelationsRequest{
			BkHostID: hostID,
		}
		respData = &FindHostBizRelationsResponse{}
	)

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
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

// TransHostToRecycleModule trans host to recycleModule
func (c *Client) TransHostToRecycleModule(bizID int, hostID []int) error {
	if c == nil {
		return ErrServerNotInit
	}

	// transfer_host_to_recyclemodule
	var (
		reqURL  = fmt.Sprintf("%s/api/c/compapi/v2/cc/transfer_host_to_recyclemodule/", c.server)
		request = &TransHostToERecycleModuleRequest{
			BkBizID:  bizID,
			BkHostID: hostID,
		}
		respData = &TransHostToERecycleModuleResponse{}
	)

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
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

// GetBizInternalModule get biz recycle module info
func (c *Client) GetBizInternalModule(ctx context.Context, bizID int) (*BizInternalModuleData, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	language := i18n.LanguageFromCtx(ctx)
	blog.Infof("cmdb client GetBizInternalModule language %s", language)

	// get_biz_internal_module
	var (
		reqURL  = fmt.Sprintf("%s/api/c/compapi/v2/cc/get_biz_internal_module/", c.server)
		request = &QueryBizInternalModuleRequest{
			BizID: bizID,
		}
		respData = &QueryBizInternalModuleResponse{}
	)

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("Blueking-Language", language).
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.serverDebug).
		Send(request).
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

// TransHostAcrossBiz trans host to other biz
func (c *Client) TransHostAcrossBiz(hostInfo TransHostAcrossBizInfo) error {
	if c == nil {
		return ErrServerNotInit
	}

	// transfer_host_across_biz
	var (
		reqURL  = fmt.Sprintf("%s/api/c/compapi/v2/cc/transfer_host_across_biz/", c.server)
		request = &TransferHostAcrossBizRequest{
			SrcBizID:   hostInfo.SrcBizID,
			BkHostID:   hostInfo.HostID,
			DstBizID:   hostInfo.DstBizID,
			BkModuleID: hostInfo.DstModuleID,
		}
		respData = &TransferHostAcrossBizResponse{}
	)

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
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

// GetBusinessMaintainer get maintainers by bizID
func (c *Client) GetBusinessMaintainer(bizID int) (*BusinessData, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	// search_business
	var (
		reqURL  = fmt.Sprintf("%s/api/c/compapi/v2/cc/search_business/", c.server)
		request = &SearchBusinessRequest{
			Fields: []string{},
			Condition: map[string]interface{}{
				conditionBkBizID: bizID,
			},
		}
		respData = &SearchBusinessResponse{}
	)

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
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

// GetBS2IDByBizID get bs2ID by bizID
func (c *Client) GetBS2IDByBizID(bizID int64) (int, error) {
	if c == nil {
		return 0, ErrServerNotInit
	}

	// search_business
	var (
		reqURL  = fmt.Sprintf("%s/api/c/compapi/v2/cc/search_business/", c.server)
		request = &SearchBusinessRequest{
			Fields: []string{fieldBS2NameID},
			Condition: map[string]interface{}{
				conditionBkBizID: bizID,
			},
		}
		respData = &SearchBusinessResponse{}
	)

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
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

// TransferHostToIdleModule transfer host to idle module
func (c *Client) TransferHostToIdleModule(bizID int, hostID []int) error {
	// transfer_host_to_idlemodule
	var (
		reqURL  = fmt.Sprintf("%s/api/c/compapi/v2/cc/transfer_host_to_idlemodule/", c.server)
		request = &TransferHostToIdleModuleRequest{
			BkBizID:  bizID,
			BkHostID: hostID,
		}
		respData = &TransferHostToIdleModuleResponse{}
	)

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
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

// TransferHostToResourceModule transfer host to resource module
func (c *Client) TransferHostToResourceModule(bizID int, hostID []int) error {
	// transfer_host_to_resourcemodule
	var (
		reqURL  = fmt.Sprintf("%s/api/c/compapi/v2/cc/transfer_host_to_resourcemodule/", c.server)
		request = &TransferHostToResourceModuleRequest{
			BkBizID:  bizID,
			BkHostID: hostID,
		}
		respData = &TransferHostToResourceModuleResponse{}
	)

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
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

// DeleteHost delete host
func (c *Client) DeleteHost(hostID []int) error {
	hostIDs := []string{}
	for _, v := range hostID {
		hostIDs = append(hostIDs, strconv.Itoa(v))
	}

	// delete_host
	var (
		reqURL  = fmt.Sprintf("%s/api/c/compapi/v2/cc/delete_host/", c.server)
		request = &DeleteHostRequest{
			BkHostID: strings.Join(hostIDs, ","),
		}
		respData = &DeleteHostResponse{}
	)

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
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

// SearchBizInstTopo search biz inst topo
func (c *Client) SearchBizInstTopo(bizID int) ([]SearchBizInstTopoData, error) {
	// search_biz_inst_topo
	var (
		reqURL   = fmt.Sprintf("%s/api/c/compapi/v2/cc/search_biz_inst_topo?bk_biz_id=%d", c.server, bizID)
		respData = &SearchBizInstTopoResponse{}
	)

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Get(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
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
	topos, err := c.SearchBizInstTopo(bizID)
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

// TransferHostModule transfer host to module
func (c *Client) TransferHostModule(bizID int, hostID []int, moduleID []int, isIncrement bool) error {
	// transfer_host_module
	var (
		reqURL  = fmt.Sprintf("%s/api/c/compapi/v2/cc/transfer_host_module/", c.server)
		request = &TransferHostModuleRequest{
			BKBizID:     bizID,
			BKHostID:    hostID,
			BKModuleID:  moduleID,
			IsIncrement: isIncrement,
		}
		respData = &BaseResponse{}
	)

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
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
func (c *Client) AddHostFromCmpy(svrIds []string, ips []string, assetIds []string) error {
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

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
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
func (c *Client) SyncHostInfoFromCmpy(bkCloudId int, bkHostIds []int64) error {
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

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
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

// GetBcsPod get pod
func (c *Client) GetBcsPod(req *GetBcsPodReq) (*[]Pod, error) {
	reqURL := fmt.Sprintf("%s/api/bk-cmdb/prod/api/v3/findmany/kube/pod", c.server)
	respData := &GetBcsPodResp{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.serverDebug).
		Send(req).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api list_pod failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api list_pod failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api list_pod with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, req, resp.Header.Get("X-Request-Id"))

	return respData.Data.Info, nil
}

// DeleteBcsPod delete pod
func (c *Client) DeleteBcsPod(req *DeleteBcsPodReq) error {
	reqURL := fmt.Sprintf("%s/api/bk-cmdb/prod/api/v3/deletemany/kube/pod", c.server)
	respData := &DeleteBcsPodResp{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Delete(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.serverDebug).
		Send(req).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_delete_kube_pod failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_delete_kube_pod failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_delete_kube_pod with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, req, resp.Header.Get("X-Request-Id"))

	return nil
}

// GetBcsWorkload get workload
func (c *Client) GetBcsWorkload(req *GetBcsWorkloadReq) (*[]interface{}, error) {
	reqURL := fmt.Sprintf("%s/api/bk-cmdb/prod/api/v3/findmany/kube/workload/%s", c.server, req.Kind)
	respData := &GetBcsWorkloadResp{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.serverDebug).
		Send(req).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api list_workload failed: %v, rid: %s", errs[0], respData.RequestID)
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api list_workload failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api list_workload with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, req, resp.Header.Get("X-Request-Id"))

	return &respData.Data.Info, nil
}

// DeleteBcsWorkload delete workload
func (c *Client) DeleteBcsWorkload(req *DeleteBcsWorkloadReq) error {
	reqURL := fmt.Sprintf("%s/api/bk-cmdb/prod/api/v3/deletemany/kube/workload/%s", c.server, *req.Kind)
	respData := &DeleteBcsWorkloadResp{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Delete(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.serverDebug).
		Send(req).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_delete_workload failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf(
			"call api batch_delete_workload failed: %v, rid: %s, request: bkbizid: %d, kind: %s, ids: %v",
			respData.Message, resp.Header.Get("X-Request-Id"), *req.BKBizID, *req.Kind, *req.IDs)
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_delete_workload with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, req, resp.Header.Get("X-Request-Id"))

	return nil
}

// GetBcsNamespace get namespace
func (c *Client) GetBcsNamespace(req *GetBcsNamespaceReq) (*[]Namespace, error) {
	reqURL := fmt.Sprintf("%s/api/bk-cmdb/prod/api/v3/findmany/kube/namespace", c.server)
	respData := &GetBcsNamespaceResp{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.serverDebug).
		Send(req).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api list_namespace failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api list_namespace failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api list_namespace with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, req, resp.Header.Get("X-Request-Id"))

	return respData.Data.Info, nil
}

// DeleteBcsNamespace delete namespace
func (c *Client) DeleteBcsNamespace(req *DeleteBcsNamespaceReq) error {
	reqURL := fmt.Sprintf("%s/api/bk-cmdb/prod/api/v3/deletemany/kube/namespace", c.server)
	respData := &DeleteBcsNamespaceResp{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Delete(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.serverDebug).
		Send(req).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_delete_namespace failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_delete_namespace failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_delete_namespace with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, req, resp.Header.Get("X-Request-Id"))

	return nil
}

// GetBcsNode get node
func (c *Client) GetBcsNode(req *GetBcsNodeReq) (*[]Node, error) {
	reqURL := fmt.Sprintf("%s/api/bk-cmdb/prod/api/v3/findmany/kube/node", c.server)
	respData := &GetBcsNodeResp{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.serverDebug).
		Send(req).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api list_kube_node failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api list_kube_node failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api list_kube_node with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, req, resp.Header.Get("X-Request-Id"))

	return respData.Data.Info, nil
}

// DeleteBcsNode delete node
func (c *Client) DeleteBcsNode(req *DeleteBcsNodeReq) error {
	reqURL := fmt.Sprintf("%s/api/bk-cmdb/prod/api/v3/deletemany/kube/node", c.server)
	respData := &DeleteBcsNodeResp{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Delete(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.serverDebug).
		Send(req).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_delete_kube_node failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_delete_kube_node failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_delete_kube_node with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, req, resp.Header.Get("X-Request-Id"))

	return nil
}

// GetBcsCluster get cluster
func (c *Client) GetBcsCluster(req *GetBcsClusterReq) (*[]Cluster, error) {
	// 如果没有通过数据库查询，则通过API调用获取集群信息
	reqURL := fmt.Sprintf("%s/api/bk-cmdb/prod/api/v3/findmany/kube/cluster", c.server)
	respData := &GetBcsClusterResp{}
	// 使用gorequest库发送POST请求，并处理响应
	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.serverDebug).
		Send(req).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	// 检查是否有错误发生
	if len(errs) > 0 {
		blog.Errorf("call api list_kube_cluster failed: %v", errs[0])
		return nil, errs[0]
	}

	// 检查API响应是否成功
	if !respData.Result {
		blog.Errorf("call api list_kube_cluster failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api list_kube_cluster with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, req, resp.Header.Get("X-Request-Id"))

	return &respData.Data.Info, nil
}

// DeleteBcsCluster delete bcs cluster
func (c *Client) DeleteBcsCluster(req *DeleteBcsClusterReq) error {
	// 构造请求的 URL
	reqURL := fmt.Sprintf("%s/api/bk-cmdb/prod/api/v3/delete/kube/cluster", c.server)
	// 初始化响应数据结构
	respData := &DeleteBcsClusterResp{}

	// 使用 gorequest 库发送 HTTP DELETE 请求
	// 设置请求超时时间、内容类型、接受类型、授权头等信息
	// 发送请求体并尝试重试，最多重试3次，每次间隔3秒，如果遇到429状态码也会重试
	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Delete(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.serverDebug).
		Send(req).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)

	// 如果请求过程中出现错误，则记录错误并返回
	if len(errs) > 0 {
		blog.Errorf("call api batch_delete_kube_cluster failed: %v", errs[0])
		return errs[0]
	}

	// 如果响应结果指示操作失败，则记录错误信息并返回错误
	if !respData.Result {
		blog.Errorf("call api batch_delete_kube_cluster failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return fmt.Errorf(respData.Message)
	}

	// 如果操作成功，则记录成功的日志信息
	blog.Infof("call api batch_delete_kube_cluster with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, req, resp.Header.Get("X-Request-Id"))

	return nil
}
