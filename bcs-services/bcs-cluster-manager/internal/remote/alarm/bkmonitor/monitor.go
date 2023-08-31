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

package bkmonitor

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/alarm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/utils"
	"github.com/parnurzeal/gorequest"
)

// Options monitor options
type Options struct {
	AppCode   string
	AppSecret string
	Enable    bool
	Server    string
	Debug     bool
}

// MonitorClient global bkMonitor client
var MonitorClient *Client

// SetMonitorClient set bkMonitor client
func SetMonitorClient(options Options) error {
	cli, err := NewClient(options)
	if err != nil {
		return err
	}

	MonitorClient = cli
	return nil
}

// GetBkMonitorClient get bkMonitor client
func GetBkMonitorClient() *Client {
	return MonitorClient
}

// NewClient create bkMonitor client
func NewClient(options Options) (*Client, error) {
	c := &Client{
		appCode:     options.AppCode,
		appSecret:   options.AppSecret,
		server:      options.Server,
		enable:      options.Enable,
		serverDebug: options.Debug,
	}

	if !options.Enable {
		return nil, nil
	}

	return c, nil
}

// Client for bkMonitor
type Client struct {
	appCode     string
	appSecret   string
	enable      bool
	server      string
	serverDebug bool
}

func (c *Client) getAccessToken(clientAuth *auth.ClientAuth) (string, error) {
	if c == nil {
		return "", alarm.ErrServerNotInit
	}

	if clientAuth != nil {
		return clientAuth.GetAccessToken(utils.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		})
	}

	return auth.GetAccessClient().GetAccessToken(utils.BkAppUser{
		BkAppCode:   c.appCode,
		BkAppSecret: c.appSecret,
	})
}

// ShieldHostAlarmConfig shield host alarm
func (c *Client) ShieldHostAlarmConfig(user string, config *alarm.ShieldHost) error {
	if c == nil {
		return alarm.ErrServerNotInit
	}

	var (
		reqURL   = "/add_shield"
		respData = &utils.BaseResponse{}
	)

	token, err := c.getAccessToken(nil)
	if err != nil {
		blog.Errorf("ShieldHostAlarmConfig getAccessToken failed: %v", err)
		return err
	}

	userAuth, err := utils.BuildGateWayAuth(&utils.AuthInfo{
		BkAppUser: utils.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		AccessToken: token,
	}, "")
	if err != nil {
		blog.Errorf("call api ShieldHostAlarmConfig BuildGateWayAuth failed: %v", err)
		return err
	}
	req, err := buildBizHostAlarmConfig(config)
	if err != nil {
		blog.Errorf("call api ShieldHostAlarmConfig buildBizHostAlarmConfig failed: %v", err)
		return err
	}

	_, _, errs := gorequest.New().
		Timeout(alarm.DefaultTimeOut).
		Post(c.server+reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		SetDebug(c.serverDebug).
		Send(req).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api ShieldHostAlarmConfig failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api ShieldHostAlarmConfig failed: %v", respData.Message)
		return fmt.Errorf(respData.Message)
	}

	// successfully request
	blog.Infof("call api ShieldHostAlarmConfig with shields successfully")
	return nil
}

// Name for client name
func (c *Client) Name() string {
	return "bk_monitor"
}
