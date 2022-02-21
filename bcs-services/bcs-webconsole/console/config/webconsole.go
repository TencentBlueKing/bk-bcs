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

const (
	InternalMode = "internal" // 用户自己集群 inCluster 模式
	ExternalMode = "external" // 平台集群, 外部模式, 需要设置 AdminClusterId
)

type WebConsoleConf struct {
	Image          string `yaml:"image"`
	AdminClusterId string `yaml:"admin_cluster_id"`
	Mode           string `yaml:"mode"` // internal , external
}

func (c *WebConsoleConf) Init() error {
	// only for development
	c.Image = ""
	c.AdminClusterId = ""
	c.Mode = InternalMode

	return nil
}
