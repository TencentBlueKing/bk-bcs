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

// HostConf host conf
type HostConf struct {
	BKIAMHost  string `yaml:"bk_iam_host"`  // 权限中心
	BKCMDBHost string `yaml:"bk_cmdb_host"` // 配置平台
	BSCPAPIURL string `yaml:"bscp_api_url"` // bscp api地址
}

// FrontendConf docs and host conf
type FrontendConf struct {
	Docs   map[string]string `yaml:"docs"`
	Host   *HostConf         `yaml:"hosts"`
	Helper string            `yaml:"helper"` // 白名单对接人员
}

// defaultFrontendConf 默认配置
func defaultUIConf() *FrontendConf {
	c := &FrontendConf{
		Docs: map[string]string{},
		Host: &HostConf{},
	}
	return c
}
