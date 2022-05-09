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
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"

	"github.com/parnurzeal/gorequest"
)

// CmdbInterface for cloud Tags
type CmdbInterface interface {
	// GetBusinessMaintainer get biz maintainer
	GetBusinessMaintainer(bizID int) (*BusinessData, error)
	// FetchAllHostsByBizID fetch all hosts by bizID
	FetchAllHostsByBizID(bizID int) ([]HostData, error)
}

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

	wgHost := sync.WaitGroup{}
	con := utils.NewBarrier(20)

	for i := range pages {
		con.Advance()
		wgHost.Add(1)
		go func(page Page) {
			defer wgHost.Done()
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
	wgHost.Wait()

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
	//successfully request
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
	//successfully request
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
	//successfully request
	blog.Infof("call api GetBS2IDByBizID with url(%s) successfully", reqURL)

	if len(respData.Data.Info) > 0 {
		return respData.Data.Info[0].BS2NameID, nil
	}

	return 0, fmt.Errorf("call api GetBS2IDByBizID failed")
}
