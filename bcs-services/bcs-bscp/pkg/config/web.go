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

package config

import (
	"fmt"
	"net/url"
	"path"
)

// WebConf web 相关配置
type WebConf struct {
	Host                 string   `yaml:"host"`
	RoutePrefix          string   `yaml:"route_prefix"` // vue路由, 静态资源前缀
	PreferredDomains     string   `yaml:"preferred_domains"`
	BaseURL              *url.URL `yaml:"-"`
	BKSharedResURL       string   `yaml:"bk_shared_res_url"` // 对应运维公共变量bkSharedResUrl, PaaS环境变量BKPAAS_SHARED_RES_URL
	BKSharedResBaseJSURL string   `yaml:"-"`                 // 规则是${bkSharedResUrl}/${目录名 aks app_code}/base.js
}

// init 初始化
func (c *WebConf) init() error {
	u, err := url.Parse(c.Host)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, c.RoutePrefix)

	c.BaseURL = u
	return nil
}

func (c *WebConf) initResBaseJSURL(appCode string) error {
	if c.BKSharedResURL == "" {
		return nil
	}
	if appCode == "" {
		return fmt.Errorf("initResBaseJSURL: app_code is required")
	}

	u, err := url.Parse(c.BKSharedResURL)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, appCode, "base.js")

	c.BKSharedResBaseJSURL = u.String()
	return nil
}

// defaultWebConf 默认配置
func defaultWebConf() *WebConf {
	c := &WebConf{
		Host:             "",
		RoutePrefix:      "",
		PreferredDomains: "",
	}
	return c
}
