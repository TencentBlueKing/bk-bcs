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

// EtcdConf etcd配置
type EtcdConf struct {
	Endpoints string `yaml:"endpoints"`
	Ca        string `yaml:"ca"`
	Cert      string `yaml:"cert"`
	Key       string `yaml:"key"`
}

// Init etcd初始化默认值
func (c *EtcdConf) Init() error {
	c.Endpoints = ""
	c.Cert = ""
	c.Ca = ""
	c.Key = ""
	return nil
}
