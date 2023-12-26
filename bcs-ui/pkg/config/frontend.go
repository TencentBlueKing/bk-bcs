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

// HostConf :
type HostConf struct {
	SiteURL               string `yaml:"site_url"`                // 前端路由URL
	DevOpsHost            string `yaml:"devops_host"`             // 蓝盾
	DevOpsBCSAPIURL       string `yaml:"devops_bcs_api_url"`      // SaaS Backend api 地址
	DevOpsArtifactoryHost string `yaml:"devops_artifactory_host"` // 制品库地址
	BKPaaSHost            string `yaml:"bk_paas_host"`            // PaaS 地址
	BKIAMHost             string `yaml:"bk_iam_host"`             // 权限中心
	BKCCHost              string `yaml:"bk_cc_host"`              // cmdb
	BKMonitorHost         string `yaml:"bk_monitor_host"`         // 蓝鲸监控
	BKSREHOST             string `yaml:"bk_sre_host"`             // 申请服务器地址
	BKUserHost            string `yaml:"bk_user_host"`            // 用户中心地址
	BKLogHost             string `yaml:"bk_log_host"`             // 日志平台地址
	LoginFullURL          string `yaml:"login_full_url"`          // 登录跳转地址
}

// FrontendConf frontend config
type FrontendConf struct {
	Docs     map[string]string `yaml:"docs"`
	Host     *HostConf         `yaml:"hosts"`
	Features map[string]string `yaml:"features"`
}

// defaultFrontendConf 默认配置
func defaultFrontendConf() *FrontendConf {
	c := &FrontendConf{
		Docs:     map[string]string{},
		Host:     &HostConf{SiteURL: "/bcs"},
		Features: map[string]string{"zh_cn": ""},
	}
	return c
}
