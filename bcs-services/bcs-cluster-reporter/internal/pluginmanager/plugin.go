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

// Package pluginmanager xxx
package pluginmanager

import (
	"sync"
)

// Plugin xxx
type Plugin interface {
	// return plugin name
	Name() string
	// setup plugin work flow
	Setup(configFilePath string, runMode string) error
	// stop plugin work flow
	Stop() error
	// check if plugin result is ready
	Ready(string) bool
	// get serilized check result
	GetResult(string) CheckResult
	// get check details for further analysiss
	GetDetail() interface{}
	// Check function for one time execute
	Check(checkOption CheckOption)
}

// CheckOption check options
type CheckOption struct {
	// 是否触发深度检查
	DeepCheck bool `json:"deepCheck" form:"deepCheck"`
	// 需要返回的plugin
	PluginStr  string   `json:"pluginStr" form:"pluginStr"`
	ClusterIDs []string `json:"clusterIDs" form:"clusterIDs"`
}

// CheckItem struct to store check result
type CheckItem struct {
	// 检查项的名字 集群可用性 . etc 或者说大类
	ItemName string
	// ItemTarget 该诊断的对象
	ItemTarget string
	// 检查的其它相关信息，用来提示用户以及匹配文档
	Detail string
	// tag 用来聚合输出
	Tags map[string]string
	// level
	Level  string
	Normal bool

	// 需要对接生成不同状态的metric，所以需要status属性
	Status string
}

// InfoItem store check info
type InfoItem struct {
	// 检查项的名字 集群可用性 . etc 或者说大类
	ItemName string
	// label，检查项的相关信息,会输出在报告和metric中，主要用于nodeagent
	Labels map[string]string
	// 检查的结果
	Result interface{}
}

// SetTags xxx
func (i CheckItem) SetTags(key, value string) CheckItem {
	i.Tags[key] = value
	return i
}

// SetItemTarget xxx
func (i CheckItem) SetItemTarget(target string) CheckItem {
	i.ItemTarget = target
	return i
}

// SetDetail xxx
func (i CheckItem) SetDetail(detail string) CheckItem {
	i.Detail = detail
	return i
}

// SetLevel xxx
func (i CheckItem) SetLevel(level string) CheckItem {
	i.Level = level
	return i
}

// CheckResult xxx
type CheckResult struct {
	Items        []CheckItem `yaml:"items"`
	InfoItemList []InfoItem  `yaml:"infoItems"`
}

// BasePlugin xxx
type BasePlugin struct {
	PluginName string
	StopChan   chan int
	CheckLock  sync.Mutex
	WriteLock  sync.Mutex
}

// NodePlugin xxx
type NodePlugin struct {
	BasePlugin
	Result CheckResult
}

// ClusterPlugin xxx
type ClusterPlugin struct {
	BasePlugin
	ReadyMap map[string]bool
	Result   map[string]CheckResult
}

// GetDetail get check detail
func (p *ClusterPlugin) GetDetail() interface{} {
	return false
}

// PluginInfo xxx
type PluginInfo struct {
	Result CheckResult `yaml:"result"`
	Detail interface{} `yaml:"detail"`
}
