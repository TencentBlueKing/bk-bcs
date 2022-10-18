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
 *
 */

package cmdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// CmdbClient global cmdb client
var CmdbClient *Client

// SetCmdbClient set cmdb client
func SetCmdbClient(options Options) error {
	cli, err := NewCmdbClient(options)
	if err != nil {
		return err
	}

	CmdbClient = cli
	return nil
}

// GetCmdbClient get cmdb client
func GetCmdbClient() *Client {
	return CmdbClient
}

// NewCmdbClient create cmdb client
func NewCmdbClient(options Options) (*Client, error) {
	c := &Client{
		appCode:     options.AppCode,
		appSecret:   options.AppSecret,
		bkUserName:  options.BKUserName,
		server:      options.Server,
		serverDebug: options.Debug,
	}

	if !options.Enable {
		return nil, nil
	}

	auth, err := c.generateGateWayAuth()
	if err != nil {
		return nil, err
	}
	c.userAuth = auth
	return c, nil
}

var (
	defaultTimeOut = time.Second * 60
	// ErrServerNotInit server not init
	ErrServerNotInit = errors.New("server not inited")
)

// Options for cc client
type Options struct {
	Enable     bool
	AppCode    string
	AppSecret  string
	BKUserName string
	Server     string
	Debug      bool
}

// AuthInfo auth user
type AuthInfo struct {
	BkAppCode   string `json:"bk_app_code"`
	BkAppSecret string `json:"bk_app_secret"`
	BkUserName  string `json:"bk_username"`
}

// Client for cc
type Client struct {
	appCode     string
	appSecret   string
	bkUserName  string
	server      string
	serverDebug bool
	userAuth    string
}

func (c *Client) generateGateWayAuth() (string, error) {
	if c == nil {
		return "", ErrServerNotInit
	}

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

// FetchAllHostsByBizID get allHosts by bizID
func (c *Client) FetchAllHostsByBizID(bizID int) ([]HostData, error) {
	if c == nil {
		return nil, ErrServerNotInit
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

	for i := range pages {
		con.Add(1)
		go func(page Page) {
			defer con.Done()
			_, hosts, err := c.QueryHostByBizID(bizID, page)
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

	return hostList, nil
}

// QueryHostByBizID query host by bizID
func (c *Client) QueryHostByBizID(bizID int, page Page) (int, []HostData, error) {
	if c == nil {
		return 0, nil, ErrServerNotInit
	}

	var (
		reqURL  = fmt.Sprintf("%s/api/c/compapi/v2/cc/list_biz_hosts/", c.server)
		request = &ListBizHostRequest{
			Page:    page,
			BKBizID: bizID,
			Fields:  []string{fieldHostIP, fieldHostID, fieldOperator, fieldBakOperator},
		}
		respData = &ListBizHostsResponse{}
	)

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
		blog.Errorf("call api QueryHostNumByBizID failed: %v", errs[0])
		return 0, nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api QueryHostNumByBizID failed: %v", respData.Message)
		return 0, nil, fmt.Errorf(respData.Message)
	}
	// successfully request
	blog.Infof("call api QueryHostNumByBizID with url(%s) successfully", reqURL)

	if len(respData.Data.Info) > 0 {
		return respData.Data.Count, respData.Data.Info, nil
	}

	return 0, nil, fmt.Errorf("call api GetBS2IDByBizID failed")
}

// GetBusinessMaintainer get maintainers by bizID
func (c *Client) GetBusinessMaintainer(bizID int) (*BusinessData, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

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
		blog.Errorf("call api GetBS2IDByBizID failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api GetBS2IDByBizID failed: %v", respData.Message)
		return nil, fmt.Errorf(respData.Message)
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
		blog.Errorf("call api GetBS2IDByBizID failed: %v", errs[0])
		return 0, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api GetBS2IDByBizID failed: %v", respData.Message)
		return 0, fmt.Errorf(respData.Message)
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
	var (
		reqURL  = fmt.Sprintf("%s/api/c/compapi/v2/cc/transfer_host_to_idlemodule/", c.server)
		request = &TransferHostToIdleModuleRequest{
			BkBizID:  bizID,
			BkHostID: hostID,
		}
		respData = &TransferHostToIdleModuleResponse{}
	)

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
		blog.Errorf("call api TransferHostToIdleModule failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result || respData.Code != 0 {
		blog.Errorf("call api TransferHostToIdleModule failed: %v", respData)
		return fmt.Errorf("call api TransferHostToIdleModule failed: %v", respData)
	}
	blog.Infof("call api TransferHostToIdleModule with url(%s) successfully", reqURL)

	return nil
}

// TransferHostToResourceModule transfer host to resource module
func (c *Client) TransferHostToResourceModule(bizID int, hostID []int) error {
	var (
		reqURL  = fmt.Sprintf("%s/api/c/compapi/v2/cc/transfer_host_to_resourcemodule/", c.server)
		request = &TransferHostToResourceModuleRequest{
			BkBizID:  bizID,
			BkHostID: hostID,
		}
		respData = &TransferHostToResourceModuleResponse{}
	)

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
		blog.Errorf("call api TransferHostToResourceModule failed: %v", errs[0])
		return errs[0]
	}

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
	var (
		reqURL  = fmt.Sprintf("%s/api/c/compapi/v2/cc/delete_host/", c.server)
		request = &DeleteHostRequest{
			BkHostID: strings.Join(hostIDs, ","),
		}
		respData = &DeleteHostResponse{}
	)

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
		blog.Errorf("call api DeleteHost failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result || respData.Code != 0 {
		blog.Errorf("call api DeleteHost failed: %v", respData)
		return fmt.Errorf("call api DeleteHost failed: %v", respData)
	}
	blog.Infof("call api DeleteHost with url(%s) successfully", reqURL)

	return nil
}

// GetBizInternalModule get biz internal module
func (c *Client) GetBizInternalModule(bizID int) (*GetBizInternalModuleData, error) {
	var (
		reqURL   = fmt.Sprintf("%s/api/c/compapi/v2/cc/get_biz_internal_module?bk_biz_id=%d", c.server, bizID)
		respData = &GetBizInternalModuleResponse{}
	)

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Get(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.serverDebug).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api GetBizInternalModule failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result || respData.Code != 0 {
		blog.Errorf("call api GetBizInternalModule failed: %v", respData)
		return nil, fmt.Errorf("call api GetBizInternalModule failed: %v", respData)
	}
	blog.Infof("call api GetBizInternalModule with url(%s) successfully", reqURL)

	return &respData.Data, nil
}

// SearchBizInstTopo search biz inst topo
func (c *Client) SearchBizInstTopo(bizID int) ([]SearchBizInstTopoData, error) {
	var (
		reqURL   = fmt.Sprintf("%s/api/c/compapi/v2/cc/search_biz_inst_topo?bk_biz_id=%d", c.server, bizID)
		respData = &SearchBizInstTopoResponse{}
	)

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Get(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.serverDebug).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api SearchBizInstTopo failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result || respData.Code != 0 {
		blog.Errorf("call api SearchBizInstTopo failed: %v", respData)
		return nil, fmt.Errorf("call api SearchBizInstTopo failed: %v", respData)
	}
	blog.Infof("call api SearchBizInstTopo with url(%s) successfully", reqURL)

	return respData.Data, nil
}

// ListTopology list topology
func (c *Client) ListTopology(bizID int) (*SearchBizInstTopoData, error) {
	internalModules, err := c.GetBizInternalModule(bizID)
	internalModules.ReplaceName()
	if err != nil {
		return nil, err
	}
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
	child := SearchBizInstTopoData{
		BKInstID:   internalModules.BKSetID,
		BKInstName: internalModules.BKSetName,
		BKObjID:    "set",
		BKObjName:  "set",
		Child:      make([]SearchBizInstTopoData, 0),
	}
	for _, v := range internalModules.Modules {
		child.Child = append(child.Child, SearchBizInstTopoData{
			BKInstID:   v.BKModuleID,
			BKInstName: v.BKModuleName,
			BKObjID:    "module",
			BKObjName:  "module",
			Child:      make([]SearchBizInstTopoData, 0),
		})
	}
	childs = append(childs, child)
	childs = append(childs, topo.Child...)
	topo.Child = childs
	return topo, nil
}

// TransferHostModule transfer host to module
func (c *Client) TransferHostModule(bizID int, hostID []int, moduleID []int, isIncrement bool) error {
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
		blog.Errorf("call api TransferHostModule failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result || respData.Code != 0 {
		blog.Errorf("call api TransferHostModule failed: %v", respData)
		return fmt.Errorf("call api TransferHostModule failed: %v", respData)
	}
	blog.Infof("call api TransferHostModule with url(%s) successfully", reqURL)

	return nil
}
