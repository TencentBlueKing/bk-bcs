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

package util

import (
	"strings"

	"istio.io/api/annotation"
	"istio.io/api/networking/v1alpha3"
	"istio.io/istio/pkg/config/analysis"
	"istio.io/istio/pkg/config/resource"
	"istio.io/istio/pkg/config/schema/gvk"
	corev1 "k8s.io/api/core/v1"
)

// InitServiceEntryHostMap 初始化服务入口主机映射
func InitServiceEntryHostMap(ctx analysis.Context) map[ScopedFqdn]*v1alpha3.ServiceEntry {
	result := make(map[ScopedFqdn]*v1alpha3.ServiceEntry)

	ctx.ForEach(gvk.ServiceEntry, func(r *resource.Instance) bool {
		s := r.Message.(*v1alpha3.ServiceEntry)
		hostsNamespaceScope := string(r.Metadata.FullName.Namespace)
		if IsExportToAllNamespaces(s.ExportTo) {
			hostsNamespaceScope = ExportToAllNamespaces
		}
		for _, h := range s.GetHosts() {
			result[NewScopedFqdn(hostsNamespaceScope, r.Metadata.FullName.Namespace, h)] = s
		}
		return true
	})

	// converts k8s service to serviceEntry since destinationHost
	// validation is performed against serviceEntry
	ctx.ForEach(gvk.Service, func(r *resource.Instance) bool {
		s := r.Message.(*corev1.ServiceSpec)
		var se *v1alpha3.ServiceEntry
		var ports []*v1alpha3.ServicePort
		for _, p := range s.Ports {
			ports = append(ports, &v1alpha3.ServicePort{
				Number:   uint32(p.Port),
				Name:     p.Name,
				Protocol: string(p.Protocol),
			})
		}
		host := ConvertHostToFQDN(r.Metadata.FullName.Namespace, r.Metadata.FullName.Name.String())
		se = &v1alpha3.ServiceEntry{
			Hosts: []string{host},
			Ports: ports,
		}
		visibleNamespaces := getVisibleNamespacesFromExportToAnno(
			r.Metadata.Annotations[annotation.NetworkingExportTo.Name], r.Metadata.FullName.Namespace.String())
		for _, scope := range visibleNamespaces {
			result[NewScopedFqdn(scope, r.Metadata.FullName.Namespace, r.Metadata.FullName.Name.String())] = se
		}
		return true
	})
	return result
}

// NOCC:tosa/fn_length(设计如此)
func getVisibleNamespacesFromExportToAnno(anno, resourceNamespace string) []string {
	scopes := make([]string, 0)
	if anno == "" {
		scopes = append(scopes, ExportToAllNamespaces)
	} else {
		for _, ns := range strings.Split(anno, ",") {
			if ns == ExportToNamespaceLocal {
				scopes = append(scopes, resourceNamespace)
			} else {
				scopes = append(scopes, ns)
			}
		}
	}
	return scopes
}

// GetDestinationHost 获取目标主机
func GetDestinationHost(sourceNs resource.Namespace, exportTo []string, host string,
	serviceEntryHosts map[ScopedFqdn]*v1alpha3.ServiceEntry,
) *v1alpha3.ServiceEntry {
	// Check explicitly defined ServiceEntries as well as services discovered from the platform

	// Check ServiceEntries which are exposed to all namespaces
	allNsScopedFqdn := NewScopedFqdn(ExportToAllNamespaces, sourceNs, host)
	if s, ok := serviceEntryHosts[allNsScopedFqdn]; ok {
		return s
	}

	// ServiceEntries can be either namespace scoped or exposed to different/all namespaces
	if len(exportTo) == 0 {
		nsScopedFqdn := NewScopedFqdn(string(sourceNs), sourceNs, host)
		if s, ok := serviceEntryHosts[nsScopedFqdn]; ok {
			return s
		}
	} else {
		for _, e := range exportTo {
			if e == ExportToNamespaceLocal {
				e = sourceNs.String()
			}
			nsScopedFqdn := NewScopedFqdn(e, sourceNs, host)
			if s, ok := serviceEntryHosts[nsScopedFqdn]; ok {
				return s
			}
		}
	}

	// Now check wildcard matches, namespace scoped or all namespaces
	// (This more expensive checking left for last)
	// Assumes the wildcard entries are correctly formatted ("*<dns suffix>")
	for seHostScopedFqdn, s := range serviceEntryHosts {
		scope, seHost := seHostScopedFqdn.GetScopeAndFqdn()

		// Skip over non-wildcard entries
		if !strings.HasPrefix(seHost, Wildcard) {
			continue
		}

		// Skip over entries not visible to the current virtual service namespace
		if scope != ExportToAllNamespaces && scope != string(sourceNs) {
			continue
		}

		seHostWithoutWildcard := strings.TrimPrefix(seHost, Wildcard)
		hostWithoutWildCard := strings.TrimPrefix(host, Wildcard)

		if strings.HasSuffix(hostWithoutWildCard, seHostWithoutWildcard) {
			return s
		}
	}

	return nil
}
