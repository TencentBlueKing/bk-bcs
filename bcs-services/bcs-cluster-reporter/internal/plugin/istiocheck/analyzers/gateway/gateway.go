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

// Package gateway 提供 Istio 的网关分析器
package gateway

import (
	"fmt"

	"istio.io/api/networking/v1alpha3"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/analysis"
	"istio.io/istio/pkg/config/analysis/analyzers/util"
	"istio.io/istio/pkg/config/resource"
	"istio.io/istio/pkg/config/schema/gvk"
	v1 "k8s.io/api/core/v1"
	klabels "k8s.io/apimachinery/pkg/labels"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/msg"
)

// IngressGatewayPortAnalyzer checks a gateway's ports against the gateway's Kubernetes service ports.
type IngressGatewayPortAnalyzer struct{}

// (compile-time check that we implement the interface)
var _ analysis.Analyzer = &IngressGatewayPortAnalyzer{}

// Metadata implements analysis.Analyzer
func (*IngressGatewayPortAnalyzer) Metadata() analysis.Metadata {
	return analysis.Metadata{
		Name:        "gateway.IngressGatewayPortAnalyzer",
		Description: "Checks a gateway's ports against the gateway's Kubernetes service ports",
		Inputs: []config.GroupVersionKind{
			gvk.Gateway,
			gvk.Pod,
			gvk.Service,
		},
	}
}

// Analyze implements analysis.Analyzer
func (s *IngressGatewayPortAnalyzer) Analyze(c analysis.Context) {
	c.ForEach(gvk.Gateway, func(r *resource.Instance) bool {
		s.analyzeGateway(r, c)
		return true
	})
}

func (*IngressGatewayPortAnalyzer) analyzeGateway(r *resource.Instance, c analysis.Context) {
	gw := r.Message.(*v1alpha3.Gateway)

	// Typically there will be a single istio-ingressgateway service, which will select
	// the same ingress gateway pod workload as the Gateway resource.  If there are multiple
	// Kubernetes services, and they offer different TCP port combinations, this validator will
	// not report a problem if *any* selecting service exposes the Gateway's port.
	servicePorts := map[uint32]bool{}
	gwSelectorMatches := 0

	// For pods selected by gw.Selector, find Services that select them and remember those ports
	gwSelector := klabels.SelectorFromSet(gw.Selector)
	c.ForEach(gvk.Pod, func(rPod *resource.Instance) bool {
		podLabels := klabels.Set(rPod.Metadata.Labels)
		if gwSelector.Matches(podLabels) {
			gwSelectorMatches++
			c.ForEach(gvk.Service, func(rSvc *resource.Instance) bool {
				nsSvc := string(rSvc.Metadata.FullName.Namespace)
				if nsSvc != rPod.Metadata.FullName.Namespace.String() {
					return true // Services only select pods in their namespace
				}

				service := rSvc.Message.(*v1.ServiceSpec)
				// TODO I want to match service.Namespace to pod.ObjectMeta.Namespace
				svcSelector := klabels.SelectorFromSet(service.Selector)
				if svcSelector.Matches(podLabels) {
					for _, port := range service.Ports {
						if port.Protocol == "TCP" {
							servicePorts[uint32(port.Port)] = true
						}
					}
				}
				return true
			})
		}
		return true
	})

	// Report if we found no pods matching this gateway's selector
	if gwSelectorMatches == 0 {
		m := msg.NewReferencedResourceNotFound(r, "selector", gwSelector.String())

		label := util.ExtractLabelFromSelectorString(gwSelector.String())
		if line, ok := util.ErrorLine(r, fmt.Sprintf(util.GatewaySelector, label)); ok {
			m.Line = line
		}

		c.Report(gvk.Gateway, m)
		return
	}

	// Check each Gateway port against what the workload ingress service offers
	for _, server := range gw.Servers {
		if server.Port != nil {
			_, ok := servicePorts[server.Port.Number]
			if !ok {
				m := msg.NewGatewayPortNotOnWorkload(r, gwSelector.String(), int(server.Port.Number))

				label := util.ExtractLabelFromSelectorString(gwSelector.String())
				if line, ok := util.ErrorLine(r, fmt.Sprintf(util.GatewaySelector, label)); ok {
					m.Line = line
				}

				c.Report(gvk.Gateway, m)
			}
		}
	}
}
