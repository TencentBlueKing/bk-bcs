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

package nodeman

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// NodeManClient global nodeman client
var NodeManClient *Client

// SetNodeManClient set nodeman client
func SetNodeManClient(options Options) error {
	cli, err := NewNodeManClient(options)
	if err != nil {
		return err
	}

	NodeManClient = cli
	return nil
}

// GetNodeManClient get nodeman client
func GetNodeManClient() *Client {
	return NodeManClient
}

// NewNodeManClient create nodeman client
func NewNodeManClient(options Options) (*Client, error) {
	c := &Client{
		appCode:     options.AppCode,
		appSecret:   options.AppSecret,
		bkUserName:  options.BKUserName,
		server:      options.Server,
		serverDebug: options.Debug,
	}

	auth, err := c.generateGateWayAuth()
	if err != nil {
		return nil, err
	}
	c.userAuth = auth
	return c, nil
}

var (
	defaultTimeOut  = time.Second * 60
	defaultPage     = 1
	defaultPageSize = 200
)

// Options for client
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

// Client for nodeman
type Client struct {
	appCode     string
	appSecret   string
	bkUserName  string
	server      string
	serverDebug bool
	userAuth    string
}

func (c *Client) generateGateWayAuth() (string, error) {
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

// CloudList get cloud list
func (c *Client) CloudList() ([]CloudListData, error) {
	var (
		reqURL   = fmt.Sprintf("%s/api/cloud?RUN_VER=open&with_default_area=true", c.server)
		respData = &CloudListResponse{}
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
		blog.Errorf("call api CloudList failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api CloudList failed: %v", respData)
		return nil, fmt.Errorf("call api CloudList failed: %v", respData)
	}
	blog.Infof("call api CloudList with url(%s) successfully", reqURL)

	return respData.Data, nil
}

// JobInstall job install
func (c *Client) JobInstall(jobType JobType, hosts []JobInstallHost) (*JobInstallData, error) {
	var (
		reqURL  = fmt.Sprintf("%s/api/job/install/", c.server)
		request = &JobInstallRequest{
			JobType:       jobType,
			Hosts:         hosts,
			Retention:     1,
			ReplaceHostID: 1,
		}
		respData = &JobInstallResponse{}
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
		blog.Errorf("call api JobInstall failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api JobInstall failed: %v", respData)
		return nil, fmt.Errorf("call api JobInstall failed: %v", respData)
	}
	blog.Infof("call api JobInstall with url(%s) successfully", reqURL)

	return &respData.Data, nil
}

// JobDetails get job detail
func (c *Client) JobDetails(jobID int) (*JobDetailsData, error) {
	var (
		reqURL  = fmt.Sprintf("%s/api/job/details/", c.server)
		request = &JobDetailsRequest{
			JobID:    jobID,
			Page:     defaultPage,
			PageSize: defaultPageSize,
		}
		respData = &JobDetailsResponse{}
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
		blog.Errorf("call api JobDetails failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api JobDetails failed: %v", respData)
		return nil, fmt.Errorf("call api JobDetails failed: %v", respData)
	}
	blog.Infof("call api JobDetails with url(%s) successfully", reqURL)

	return &respData.Data, nil
}

// ListHosts list hosts with bk_biz_id
func (c *Client) ListHosts(bkBizID, page, pageSize int) (*ListHostsData, error) {
	var (
		reqURL  = fmt.Sprintf("%s/api/host/search/", c.server)
		request = &ListHostsRequest{
			BKBizIDs: []int{bkBizID},
			Page:     page,
			PageSize: pageSize,
		}
		respData = &ListHostsResponse{}
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
		blog.Errorf("call api ListHosts failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api ListHosts failed: %v", respData)
		return nil, fmt.Errorf("call api ListHosts failed: %v", respData)
	}
	blog.Infof("call api ListHosts with url(%s) successfully", reqURL)

	return &respData.Data, nil
}

// ListAllHosts list all hosts with bk_biz_id
func (c *Client) ListAllHosts(bkBizID int) ([]HostInfo, error) {
	// get all host counts
	result, err := c.ListHosts(bkBizID, defaultPage, defaultPageSize)
	if err != nil {
		return nil, err
	}

	blog.Infof("ListAllHosts count %d by bizID %d", result.Total, bkBizID)
	var (
		hostList = make([]HostInfo, 0)
		hostLock = &sync.RWMutex{}
	)

	con := utils.NewRoutinePool(20)
	defer con.Close()

	page := (result.Total-1)/defaultPageSize + 1
	for i := 1; i <= page; i++ {
		con.Add(1)
		go func(page int) {
			defer con.Done()
			hosts, err := c.ListHosts(bkBizID, page, defaultPageSize)
			if err != nil {
				blog.Errorf("ListAllHosts %v failed, %s", bkBizID, err.Error())
				return
			}
			hostLock.Lock()
			hostList = append(hostList, hosts.List...)
			hostLock.Unlock()
		}(i)
	}
	con.Wait()

	blog.Infof("ListAllHosts successful %v", bkBizID)
	return hostList, nil
}

// GetHostIDByIPs get host id by ips
func (c *Client) GetHostIDByIPs(bkBizID int, ips []string) ([]int, error) {
	hostIDs := make([]int, 0)
	hosts, err := c.ListAllHosts(bkBizID)
	if err != nil {
		return nil, fmt.Errorf("list nodeman hosts err %s", err.Error())
	}

	for _, v := range hosts {
		for _, ip := range ips {
			if v.InnerIP == ip {
				hostIDs = append(hostIDs, v.BKHostID)
				break
			}
		}
	}
	return hostIDs, nil
}
