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
	"strings"

	k8scorev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// IngressConverter listener generator
type IngressConverter struct {
	ingress *networkextensionv1.Ingress
	cli     client.Client
}

// NewIngressConverter create ingress generator
func NewIngressConverter(ingress *networkextensionv1.Ingress, cli client.Client) *IngressConverter {

}

// ProcessUpdateIngress process newly added or updated ingress
func (g *IngressConverter) ProcessUpdateIngress() error {
	lbIDStrs := g.ingress.Annotations[networkextensionv1.AnnotationKeyForLoadbalanceIDs]
	lbIDs = strings.Split(lbIDStrs, ",")
	for _, rule := range g.ingress.Spec.Rules {
		ruleConverter := NewRuleConverter(g.cli, lbIDs, g.ingress.GetName(), g.ingress.GetNamespace(), rule)
		listeners, err := ruleConverter.DoConvert()
		if err != nil {
			blog.Errorf("convert rule +v failed, err %s", rule, err.Error())
			return fmt.Errorf("convert rule +v failed, err %s", rule, err.Error())
		}
	}
	for _, mapping := range g.ingress.Spec.PortMappings {
		MappingConverter := NewMappingConverter(g.cli, lbIDs, g.ingress.GetName(), g.ingress.GetNamespace(), mapping)
		listeners, err := MappingConverter.DoConvert()
		if err != nil {
			blog.Errorf("convert mapping +v failed, err %s", mapping, err.Error())
			return fmt.Errorf("convert mapping +v failed, err %s", mapping, err.Error())
		}
	}
}

// ProcessDeleteIngress  process deleted ingress
func (g *IngressConverter) ProcessDeleteIngress(ingressName, ingressNamespaces string) error {

}

func (g *IngressConverter) validate() error {

}

func (g *IngressConverter) generate() error {

}

func (g *IngressConverter) syncListener() error {

}
