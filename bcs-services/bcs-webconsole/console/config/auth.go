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

// AuthConf :
type AuthConf struct {
	Host        string `yaml:"host"`         // bkiam 地址, 获取 global.bkIAM.iamHost
	GatewayHost string `yaml:"gateway_host"` // 网关模式地址, 如果不为空，优先使用 gatway 模式; 获取 global.bkIAM.gateWayHost
	UseGateway  bool   `yaml:"use_gw"`       // 是否启用网关
}

// Init : init default auth config
func (c *AuthConf) Init() {
	// only for development
	c.Host = ""
	c.GatewayHost = ""
	c.UseGateway = true
}
