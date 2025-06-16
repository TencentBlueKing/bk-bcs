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

package virtualservice

import (
	"fmt"
	"strings"

	"istio.io/api/networking/v1alpha3"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/analysis"
	"istio.io/istio/pkg/config/analysis/analyzers/util"
	"istio.io/istio/pkg/config/host"
	"istio.io/istio/pkg/config/resource"
	"istio.io/istio/pkg/config/schema/gvk"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/msg"
)

// GatewayAnalyzer checks the gateways associated with each virtual service
type GatewayAnalyzer struct{}

var _ analysis.Analyzer = &GatewayAnalyzer{}

// Metadata implements Analyzer
func (s *GatewayAnalyzer) Metadata() analysis.Metadata {
	return analysis.Metadata{
		Name:        "virtualservice.GatewayAnalyzer",
		Description: "Checks the gateways associated with each virtual service",
		Inputs: []config.GroupVersionKind{
			gvk.Gateway,
			gvk.VirtualService,
		},
	}
}

// Analyze implements Analyzer
func (s *GatewayAnalyzer) Analyze(c analysis.Context) {
	c.ForEach(gvk.VirtualService, func(r *resource.Instance) bool {
		s.analyzeVirtualService(r, c)
		return true
	})
}

func (s *GatewayAnalyzer) analyzeVirtualService(r *resource.Instance, c analysis.Context) {
	vs := r.Message.(*v1alpha3.VirtualService)
	vsNs := r.Metadata.FullName.Namespace
	vsName := r.Metadata.FullName

	for i, gwName := range vs.Gateways {
		// This is a special-case accepted value
		if gwName == util.MeshGateway {
			continue
		}

		gwFullName := resource.NewShortOrFullName(vsNs, gwName)

		if !c.Exists(gvk.Gateway, gwFullName) {
			m := msg.NewReferencedResourceNotFound(r, "gateway", gwName)

			if line, ok := util.ErrorLine(r, fmt.Sprintf(util.VSGateway, i)); ok {
				m.Line = line
			}

			c.Report(gvk.VirtualService, m)
		}

		if !vsHostInGateway(c, gwFullName, vs.Hosts, vsNs.String()) {
			m := msg.NewVirtualServiceHostNotFoundInGateway(r, vs.Hosts, vsName.String(), gwFullName.String())

			if line, ok := util.ErrorLine(r, fmt.Sprintf(util.VSGateway, i)); ok {
				m.Line = line
			}

			c.Report(gvk.VirtualService, m)
		}
	}
}

func vsHostInGateway(c analysis.Context, gateway resource.FullName, vsHosts []string, vsNamespace string) bool {
	var gatewayHosts []string
	var gatewayNs string

	c.ForEach(gvk.Gateway, func(r *resource.Instance) bool {
		if r.Metadata.FullName == gateway {
			s := r.Message.(*v1alpha3.Gateway)
			gatewayNs = r.Metadata.FullName.Namespace.String()
			for _, v := range s.Servers {
				sanitizeServerHostNamespace(v, gatewayNs)
				gatewayHosts = append(gatewayHosts, v.Hosts...)
			}
		}

		return true
	})

	gatewayHostNames := host.NamesForNamespace(gatewayHosts, vsNamespace)
	for _, gh := range gatewayHostNames {
		for _, vsh := range vsHosts {
			gatewayHost := gh
			vsHost := host.Name(vsh)

			if gatewayHost.Matches(vsHost) {
				return true
			}
		}
	}

	return false
}

// convert ./host to currentNamespace/Host
// */host to just host
// */* to just *
func sanitizeServerHostNamespace(server *v1alpha3.Server, namespace string) {
	for i, h := range server.Hosts {
		if strings.Contains(h, "/") {
			parts := strings.Split(h, "/")
			if parts[0] == "." {
				server.Hosts[i] = fmt.Sprintf("%s/%s", namespace, parts[1])
			} else if parts[0] == "*" {
				if parts[1] == "*" {
					server.Hosts = []string{"*"}
					return
				}
				server.Hosts[i] = parts[1]
			}
		}
	}
}
