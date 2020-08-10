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

package controllers

import (
	"strings"

	k8sappsv1 "k8s.io/api/apps/v1"
	k8scorev1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"
)

func findIngressesByService(service *k8scorev1.Service,
	ingressList *networkextensionv1.IngressList) []*networkextensionv1.Ingress {
	var retIngressList []*networkextensionv1.Ingress
	for _, ingress := range ingressList.Items {
		found := false
		for _, rule := range ingress.Spec.Rules {
			if strings.ToLower(rule.Protocol) == "tcp" || strings.ToLower(rule.Protocol) == "udp" {
				for _, route := range rule.Services {
					if service.GetName() == route.ServiceName && service.GetNamespace() == route.ServiceNamespace {
						blog.V(2).Infof("service %s/%s found in ingress %s/%s",
							service.GetNamespace(), service.GetName(), ingress.GetNamespace(), ingress.GetName())
						retIngressList = append(retIngressList, &ingress)
						found = true
					}
				}
			}
			if found {
				break
			}
			if strings.ToLower(rule.Protocol) == "http" || strings.ToLower(rule.Protocol) == "https" {
				for _, httpRoute := range rule.Routes {
					for _, route := range httpRoute.Services {
						if service.GetName() == route.ServiceName && service.GetNamespace() == route.ServiceNamespace {
							blog.V(2).Infof("service %s/%s found in ingress %s/%s",
								service.GetNamespace(), service.GetName(), ingress.GetNamespace(), ingress.GetName())
							retIngressList = append(retIngressList, &ingress)
							found = true
						}
					}
				}
			}
			if found {
				break
			}
		}
	}
	return retIngressList
}

func findIngressesByStatefulSet(sts *k8sappsv1.StatefulSet,
	ingressList *networkextensionv1.IngressList) []*networkextensionv1.Ingress {
	var retIngressList []*networkextensionv1.Ingress
	for _, ingress := range ingressList.Items {
		for _, mapping := range ingress.Spec.PortMappings {
			if strings.ToLower(mapping.WorkloadKind) == "statefulset" &&
				mapping.WorkloadName == sts.GetName() &&
				mapping.WorkloadNamespace == sts.GetNamespace() {

				retIngressList = append(retIngressList, &ingress)
				break
			}
		}
	}
	return retIngressList
}
