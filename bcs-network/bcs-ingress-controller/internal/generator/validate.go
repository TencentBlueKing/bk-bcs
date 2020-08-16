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

package generator

import (
	"fmt"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"
)

func validateIngress(ingress *networkextensionv1.Ingress) (bool, string) {
	ruleMap := make(map[int]*networkextensionv1.IngressRule)
	for index, rule := range ingress.Spec.Rules {
		existedRule, ok := portMap[rule.Port]
		if !ok {
			portMap[rule.Port] = ingress.Spec.Rules[index]
		}
		return false, fmt.Sprintf("%+v conflicts with %+v", rule, existedRule)
	}

	mappingMap := make(map[int]*networkextensionv1.IngressPortMapping)
	for i := 0; i < len(ingress.Spec.PortMappings); i++ {
		mapping := ingress.Spec.PortMappings[i]
		for port, rule := range ruleMap {
			if port >= mapping.StartPort+mapping.StartIndex && port < mapping.StartPort+mapping.EndIndex {
				return false, fmt.Sprintf("%+v port conflicts with %+v", mapping, rule)
			}
		}
		for j := i + 1; j < len(ingress.Spec.PortMappings); j++ {
			tmpMapping := ingress.Spec.PortMappings[j]
			if mapping.StartPort+mapping.StartIndex > tmpMapping.StartPort+tmpMapping.EndIndex ||
				mapping.StartPort+mapping.EndIndex < tmpMapping.StartPort+tmpMapping.StartIndex {
				continue
			}
			return false, fmt.Sprintf("%+v ports conflicts with %+v", mapping, tmpMapping)
		}
	}
	return true, ""
}

func validateIngressRule(rule *networkextensionv1.IngressRule) (bool, string) {

}

func validateIngressPortMapping(mapping *networkextensionv1.IngressPortMapping) (bool, string) {

}
