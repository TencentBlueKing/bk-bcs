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

package ingresscontroller

import (
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"
)

func isServiceInIngress(ingress *networkextensionv1.Ingress, svcName, svcNamespace string) bool {
	for _, rule := range ingress.Spec.Rules {
		if strings.ToLower(rule.Protocol) == "tcp" || strings.ToLower(rule.Protocol) == "udp" {
			for _, route := range rule.Services {
				if svcName == route.ServiceName && svcNamespace == route.ServiceNamespace {
					blog.V(2).Infof("service %s/%s found in ingress %s/%s",
						svcNamespace, svcName, ingress.GetNamespace(), ingress.GetName())
					return true
				}
			}
		}
		if strings.ToLower(rule.Protocol) == "http" || strings.ToLower(rule.Protocol) == "https" {
			for _, httpRoute := range rule.Routes {
				for _, route := range httpRoute.Services {
					if svcName == route.ServiceName && svcNamespace == route.ServiceNamespace {
						blog.V(2).Infof("service %s/%s found in ingress %s/%s",
							svcNamespace, svcName, ingress.GetNamespace(), ingress.GetName())
						return true
					}
				}
			}
		}
	}
	return false
}

func findIngressesByService(svcName, svcNamespace string,
	ingressList *networkextensionv1.IngressList) []*networkextensionv1.Ingress {
	var retIngressList []*networkextensionv1.Ingress
	for index, ingress := range ingressList.Items {
		if isServiceInIngress(&ingress, svcName, svcNamespace) {
			retIngressList = append(retIngressList, &ingressList.Items[index])
		}
	}
	return retIngressList
}

func findIngressesByWorkload(kind, name, ns string,
	ingressList *networkextensionv1.IngressList) []*networkextensionv1.Ingress {
	var retIngressList []*networkextensionv1.Ingress
	for index, ingress := range ingressList.Items {
		for _, mapping := range ingress.Spec.PortMappings {
			if strings.ToLower(mapping.WorkloadKind) == strings.ToLower(kind) &&
				mapping.WorkloadName == name &&
				mapping.WorkloadNamespace == ns {
				blog.V(2).Infof("workload %s/%s/%s found in ingress %s/%s",
					kind, name, ns, ingress.GetNamespace(), ingress.GetName())
				retIngressList = append(retIngressList, &ingressList.Items[index])
				break
			}
		}
	}
	return retIngressList
}

func deduplicateIngresses(ingresses []*networkextensionv1.Ingress) []*networkextensionv1.Ingress {
	var retList []*networkextensionv1.Ingress
	ingressMap := make(map[string]*networkextensionv1.Ingress)
	for index, ingress := range ingresses {
		if _, ok := ingressMap[ingress.GetNamespace()+"/"+ingress.GetName()]; !ok {
			ingressMap[ingress.GetNamespace()+"/"+ingress.GetName()] = ingresses[index]
			retList = append(retList, ingresses[index])
		}
	}
	return retList
}
