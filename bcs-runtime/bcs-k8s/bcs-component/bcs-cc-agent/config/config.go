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

// Package config xxx
package config

// BcsCcAgentConfig is a configuration of  bcs-cc-agent
type BcsCcAgentConfig struct {
	MetricPort uint
	EngineType string
	Kubeconfig string
	KubeMaster string
	EsbUrl     string
	AppCode    string
	AppSecret  string
	BkUsername string
}

// NewBcsCcAgentConfig create a config object
func NewBcsCcAgentConfig() *BcsCcAgentConfig {
	return &BcsCcAgentConfig{}
}
