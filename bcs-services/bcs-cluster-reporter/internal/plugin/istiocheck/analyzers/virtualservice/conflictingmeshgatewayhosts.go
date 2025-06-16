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

// Package virtualservice 提供 Istio 的虚拟服务分析器
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
	"istio.io/istio/pkg/util/sets"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/msg"
)

// ConflictingMeshGatewayHostsAnalyzer checks if multiple virtual services
// associated with the mesh gateway have conflicting hosts. The behavior is
// undefined if conflicts exist.
type ConflictingMeshGatewayHostsAnalyzer struct{}

var _ analysis.Analyzer = &ConflictingMeshGatewayHostsAnalyzer{}

// Metadata implements Analyzer
func (c *ConflictingMeshGatewayHostsAnalyzer) Metadata() analysis.Metadata {
	return analysis.Metadata{
		Name:        "virtualservice.ConflictingMeshGatewayHostsAnalyzer",
		Description: "Checks if multiple virtual services associated with the mesh gateway have conflicting hosts",
		Inputs: []config.GroupVersionKind{
			gvk.VirtualService,
		},
	}
}

// Analyze implements Analyzer
func (c *ConflictingMeshGatewayHostsAnalyzer) Analyze(ctx analysis.Context) {
	hs := initMeshGatewayHosts(ctx)
	reported := make(map[resource.FullName]bool)
	for scopedFqdn, vsList := range hs {
		scope, _ := scopedFqdn.GetScopeAndFqdn()
		if scope != util.ExportToAllNamespaces {
			noScopedVSList := getExportToAllNamespacesVSListForScopedHost(scopedFqdn, hs)
			vsList = append(vsList, noScopedVSList...)
		}
		if len(vsList) > 1 {
			vsNames := combineResourceEntryNames(vsList)
			for i := range vsList {
				if reported[vsList[i].Metadata.FullName] {
					continue
				}
				reported[vsList[i].Metadata.FullName] = true
				m := msg.NewConflictingMeshGatewayVirtualServiceHosts(vsList[i], vsNames, string(scopedFqdn))

				if line, ok := util.ErrorLine(vsList[i], fmt.Sprintf(util.MetadataName)); ok {
					m.Line = line
				}

				ctx.Report(gvk.VirtualService, m)
			}
		}

	}
}

// getExportToAllNamespacesVSListForScopedHost 获取所有导出到所有命名空间的虚拟服务
// NOCC:tosa/fn_length(设计如此)
func getExportToAllNamespacesVSListForScopedHost(sh util.ScopedFqdn, meshGatewayHosts map[util.ScopedFqdn][]*resource.Instance) []*resource.Instance {
	_, h := sh.GetScopeAndFqdn()
	vss := make([]*resource.Instance, 0)
	for sf, resources := range meshGatewayHosts {
		mghScope, mgh := sf.GetScopeAndFqdn()
		hName := host.Name(h)
		mghName := host.Name(mgh)
		if mghScope != util.ExportToAllNamespaces || !hName.Matches(mghName) {
			continue
		}
		vss = append(vss, resources...)
	}
	return vss
}

func combineResourceEntryNames(rList []*resource.Instance) string {
	names := make([]string, 0, len(rList))
	for _, r := range rList {
		names = append(names, r.Metadata.FullName.String())
	}
	return strings.Join(names, ",")
}

func initMeshGatewayHosts(ctx analysis.Context) map[util.ScopedFqdn][]*resource.Instance {
	hostsVirtualServices := map[util.ScopedFqdn][]*resource.Instance{}
	ctx.ForEach(gvk.VirtualService, func(r *resource.Instance) bool {
		vs := r.Message.(*v1alpha3.VirtualService)
		vsNamespace := r.Metadata.FullName.Namespace
		vsAttachedToMeshGateway := false
		// No entry in gateways imply "mesh" by default
		if len(vs.Gateways) == 0 {
			vsAttachedToMeshGateway = true
		} else {
			for _, g := range vs.Gateways {
				if g == util.MeshGateway {
					vsAttachedToMeshGateway = true
				}
			}
		}
		if vsAttachedToMeshGateway {
			// determine the scope of hosts i.e. local to VirtualService namespace or
			// all namespaces
			hostsNamespaceScope := make([]string, 0)
			exportToAllNamespaces := util.IsExportToAllNamespaces(vs.ExportTo)
			if exportToAllNamespaces {
				hostsNamespaceScope = append(hostsNamespaceScope, util.ExportToAllNamespaces)
			} else {
				nss := sets.New[string]()
				for _, et := range vs.ExportTo {
					if et == util.ExportToNamespaceLocal {
						nss.Insert(vsNamespace.String())
					} else {
						nss.Insert(et)
					}
				}
				hostsNamespaceScope = nss.UnsortedList()
			}

			for _, nsScope := range hostsNamespaceScope {
				for _, h := range vs.Hosts {
					scopedFqdn := util.NewScopedFqdn(nsScope, vsNamespace, h)
					vsNames := hostsVirtualServices[scopedFqdn]
					if len(vsNames) == 0 {
						hostsVirtualServices[scopedFqdn] = []*resource.Instance{r}
					} else {
						hostsVirtualServices[scopedFqdn] = append(hostsVirtualServices[scopedFqdn], r)
					}
				}
			}
		}
		return true
	})
	return hostsVirtualServices
}
