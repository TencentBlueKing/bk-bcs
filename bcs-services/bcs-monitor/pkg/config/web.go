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

package config

import (
	"net/url"
	"path"
)

// WebConf web 相关配置
type WebConf struct {
	Host        string   `yaml:"host"`
	RoutePrefix string   `yaml:"route_prefix"`
	BaseURL     *url.URL `yaml:"-"`
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

// defaultWebConf 默认配置
func defaultWebConf() *WebConf {
	c := &WebConf{
		Host:        "http://127.0.0.1:8083",
		RoutePrefix: "/monitor",
	}
	return c
}
