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

// ClientConf Client 配置
type ClientConf struct {
	Cert    string `yaml:"cert" usage:"Client Cert"`
	CertPwd string `yaml:"certPwd" usage:"Client Cert Password"`
	Key     string `yaml:"key" usage:"Client Key"`
	Ca      string `yaml:"ca" usage:"Client CA"`
}

// Init etcd初始化默认值
func (c *ClientConf) Init() {
	c.Cert = ""
	c.CertPwd = ""
	c.Key = ""
	c.Ca = ""
}
