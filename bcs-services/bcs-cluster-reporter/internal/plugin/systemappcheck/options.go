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

// Package systemappcheck 系统应用检查插件，检查集群中系统组件的部署状态、镜像版本和配置
package systemappcheck

import (
	"strings"
)

// defaultComponents 默认需要检查的系统组件列表
var defaultComponents = []Component{
	{
		Namespace: "kube-system",
		Name:      "kube-proxy",
		Resource:  "daemonset",
	},
	{
		Namespace: "kube-system",
		Name:      "coredns",
		Resource:  "deployment",
	},
}

// Options 系统应用检查插件的配置选项
type Options struct {
	Components           []Component            `json:"components" yaml:"components"`
	ComponentVersionConf []ComponentVersionConf `json:"componentVersionConf" yaml:"componentVersionConf"`
	Interval             int                    `json:"interval" yaml:"interval"`
	Namespaces           []string               `json:"namespaces" yaml:"namespaces"`
}

// Component 需要检查的组件定义
type Component struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Name      string `json:"name" yaml:"name"`
	Resource  string `json:"resource" yaml:"resource"`
}

// ComponentVersionConf 组件版本配置，用于检查镜像版本是否需要升级
type ComponentVersionConf struct {
	Name string `json:"name" yaml:"name"`

	NiceToUpgrade string `json:"niceToUpgrade" yaml:"niceToUpgrade"`
	NeedUpgrade   string `json:"needUpgrade" yaml:"needUpgrade"`
}

// Validate 校验并补全配置选项
func (o *Options) Validate() error {
	if o.Namespaces == nil || len(o.Namespaces) == 0 {
		o.Namespaces = []string{"kube-system", "bk-system", "bcs-system", "bkmonitor-operator", "istio-system"}
	}

	if o.Components != nil {
		for _, component := range defaultComponents {
			setFlag := false
			for _, optionComponent := range o.Components {
				if optionComponent.Name == component.Name &&
					optionComponent.Namespace == component.Namespace && strings.EqualFold(optionComponent.Resource, component.Resource) {
					setFlag = true
					break
				}
			}

			if !setFlag {
				o.Components = append(o.Components, component)
			}
		}
	} else {
		o.Components = make([]Component, len(defaultComponents))
		copy(o.Components, defaultComponents)
	}

	return nil
}
