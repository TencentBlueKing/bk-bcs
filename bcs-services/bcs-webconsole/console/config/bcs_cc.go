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

// BCSCCConf : bcs cc 接口配置, 调用项目信息使用
type BCSCCConf struct {
	Host  string `yaml:"host"`
	Stage string `yaml:"stage"`
}

func (c *BCSCCConf) Init() error {
	// only for development
	c.Host = ""
	c.Stage = "uat"
	return nil
}
