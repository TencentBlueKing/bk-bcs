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

// Package config xxx
package config

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

// HostConf host conf
type HostConf struct {
	BKIAMHost            string `yaml:"bk_iam_host"`       // 权限中心
	BKCMDBHost           string `yaml:"bk_cmdb_host"`      // 配置平台
	BSCPAPIURL           string `yaml:"bscp_api_url"`      // bscp api地址
	BKNODEMANHOST        string `yaml:"bk_nodeman_host"`   // 节点管理地址
	BKSharedResURL       string `yaml:"bk_shared_res_url"` // 对应运维公共变量bkSharedResUrl, PaaS环境变量BKPAAS_SHARED_RES_URL
	BKSharedResBaseJSURL string `yaml:"-"`                 // 规则是${bkSharedResUrl}/${目录名 aka app_code}/base.js
	UserManHost          string `yaml:"user_man_host"`     // 用户列表host
}

// FrontendConf docs and host conf
type FrontendConf struct {
	Docs           map[string]string `yaml:"docs"`
	Host           *HostConf         `yaml:"hosts"`
	Helper         string            `yaml:"helper"`           // 白名单对接人员
	EnableBKNotice bool              `yaml:"enable_bk_notice"` // 是否启用蓝鲸通知中心
}

// defaultFrontendConf 默认配置
func defaultUIConf() *FrontendConf {
	c := &FrontendConf{
		Docs: map[string]string{},
		Host: &HostConf{},
	}
	return c
}

func (c *FrontendConf) initResBaseJSURL(appCode string) error {
	if c.Host.BKSharedResURL == "" {
		return nil
	}
	if appCode == "" {
		return fmt.Errorf("initResBaseJSURL: app_code is required")
	}

	// 规范: 统一使用下划线做目录名
	appCode = strings.ReplaceAll(appCode, "-", "_")

	u, err := url.Parse(c.Host.BKSharedResURL)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, appCode, "base.js")

	c.Host.BKSharedResBaseJSURL = u.String()
	return nil
}
