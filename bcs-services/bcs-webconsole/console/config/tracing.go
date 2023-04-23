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

// TracingConf 链路追踪配置
type TracingConf struct {
	Enabled       bool              `yaml:"enabled" usage:"enable trace"`
	Endpoint      string            `yaml:"endpoint" usage:"Collector service endpoint"`
	Token         string            `yaml:"token" usage:"token for collector sevice"`
	ResourceAttrs map[string]string `yaml:"resourceAttrs" usage:"attributes of traced service"`
}

// Init xxx
func (c *TracingConf) Init() error {
	// only for development
	c.Endpoint = ""
	c.Enabled = false
	c.Token = ""
	return nil
}
