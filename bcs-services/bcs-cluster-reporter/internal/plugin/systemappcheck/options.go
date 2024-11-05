/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package systemappcheck xxx
package systemappcheck

import (
	"strings"
)

// Options bcs log options
type Options struct {
	Components           []Component            `json:"components" yaml:"components"`
	ComponentVersionConf []ComponentVersionConf `json:"componentVersionConf" yaml:"componentVersionConf"`
	Interval             int                    `json:"interval" yaml:"interval"`
	Namespaces           []string               `json:"namespaces" yaml:"namespaces"`
}

// Component xxx
type Component struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Name      string `json:"name" yaml:"name"`
	Resource  string `json:"resource" yaml:"resource"`
}

// ComponentVersionConf component version config
type ComponentVersionConf struct {
	Name string `json:"name" yaml:"name"`

	NiceToUpgrade string `json:"niceToUpgrade" yaml:"niceToUpgrade"`
	NeedUpgrade   string `json:"needUpgrade" yaml:"needUpgrade"`
}

// Validate validate options
func (o *Options) Validate() error {
	// if len(o.KubeMaster) == 0 {
	//	return fmt.Errorf("kube_master cannot be empty")
	// }
	// if len(o.Kubeconfig) == 0 {
	//	return fmt.Errorf("kubeconfig cannot be empty")
	// }

	if o.Namespaces == nil || len(o.Namespaces) == 0 {
		o.Namespaces = []string{"kube-system", "default", "bk-system", "bcs-system", "bkmonitor-operator"}
	}

	if o.Components != nil {
		components := []Component{
			{
				Namespace: "kube-system",
				Name:      "kube-proxy",
				Resource:  "daemonset",
			}, {
				Namespace: "kube-system",
				Name:      "kube-dns",
				Resource:  "deployment",
			}, {
				Namespace: "kube-system",
				Name:      "coredns",
				Resource:  "deployment",
			},
		}
		for _, component := range components {
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
		o.Components = []Component{
			{
				Namespace: "kube-system",
				Name:      "kube-proxy",
				Resource:  "daemonset",
			}, {
				Namespace: "kube-system",
				Name:      "kube-dns",
				Resource:  "deployment",
			}, {
				Namespace: "kube-system",
				Name:      "coredns",
				Resource:  "deployment",
			},
		}
	}

	return nil
}
