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

// Package tmp xxx
package tmp

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/alarm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/utils"
)

// Options bkops options
type Options struct {
	AppCode    string
	AppSecret  string
	BkUserName string
	Enable     bool
	Server     string
	Debug      bool
}

// BKAlarmClient global bkAlarm client
var BKAlarmClient *Client

// SetBKAlarmClient set bkAlarm client
func SetBKAlarmClient(options Options) error {
	cli, err := NewClient(options)
	if err != nil {
		return err
	}

	BKAlarmClient = cli
	return nil
}

// GetBKAlarmClient get bkalarm client
func GetBKAlarmClient() *Client {
	return BKAlarmClient
}

// NewClient create bkalarm client
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

// Client for bksops
type Client struct {
	appCode     string
	appSecret   string
	enable      bool
	server      string
	serverDebug bool
}

// ShieldHostAlarmConfig shield host alarm, user biz managers
func (c *Client) ShieldHostAlarmConfig(ctx context.Context, bkUserName string, config *alarm.ShieldHost) error {
	if c == nil {
		return alarm.ErrServerNotInit
	}

	var (
		reqURL   = "/add_alarm_shield_config"
		respData = &ShieldHostAlarmResponse{}
	)

	userAuth, tenant, err := utils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: "",
	}, bkUserName)
	if err != nil {
		blog.Errorf("call api ShieldHostAlarmConfig generateGateWayAuth failed: %v", err)
		return err
	}

	ipList := make([]string, 0)
	for i := range config.HostList {
		ipList = append(ipList, config.HostList[i].IP)
	}
	req := &ShieldHostAlarmRequest{
		AppID:       config.BizID,
		IPList:      strings.Join(ipList, ","),
		ShieldStart: time.Now().Format("2006-01-02 15:04"),
		ShieldEnd:   time.Now().Add(time.Minute * 30).Format("2006-01-02 15:04"),
		Operator:    bkUserName,
	}
	_, _, errs := gorequest.New().
		Timeout(alarm.DefaultTimeOut).
		Post(c.server+reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		SetDebug(c.serverDebug).
		Send(req).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api ShieldHostAlarmConfig failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api ShieldHostAlarmConfig failed: %v", respData.Message)
		return errors.New(respData.Message)
	}

	// successfully request
	blog.Infof("call api ShieldHostAlarmConfig with shields(%v) successfully", respData.Data.ShieldID)
	return nil
}

// Name for client name
func (c *Client) Name() string {
	return "bk_tmp"
}
