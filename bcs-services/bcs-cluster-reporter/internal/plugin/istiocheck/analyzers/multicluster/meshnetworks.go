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

// Package multicluster 提供 Istio 的多集群分析器
package multicluster

import (
	"fmt"

	v1 "k8s.io/api/core/v1"

	"istio.io/api/mesh/v1alpha1"
	"istio.io/istio/pilot/pkg/serviceregistry/provider"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/analysis"
	"istio.io/istio/pkg/config/analysis/analyzers/util"
	"istio.io/istio/pkg/config/resource"
	"istio.io/istio/pkg/config/schema/gvk"
	"istio.io/istio/pkg/kube/multicluster"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/msg"
)

// MeshNetworksAnalyzer validates MeshNetworks configuration in multi-cluster.
type MeshNetworksAnalyzer struct{}

var _ analysis.Analyzer = &MeshNetworksAnalyzer{}

// Service Registries that are known to istio.
var serviceRegistries = []provider.ID{
	provider.Mock,
	provider.Kubernetes,
	provider.External,
}

// Metadata implements Analyzer
func (s *MeshNetworksAnalyzer) Metadata() analysis.Metadata {
	return analysis.Metadata{
		Name:        "meshnetworks.MeshNetworksAnalyzer",
		Description: "Check the validity of MeshNetworks in the cluster",
		Inputs: []config.GroupVersionKind{
			gvk.MeshNetworks,
			gvk.Secret,
		},
	}
}

// Analyze implements Analyzer
func (s *MeshNetworksAnalyzer) Analyze(c analysis.Context) {
	c.ForEach(gvk.Secret, func(r *resource.Instance) bool {
		if r.Metadata.Labels[multicluster.MultiClusterSecretLabel] == "true" {
			s := r.Message.(*v1.Secret)
			for c := range s.Data {
				serviceRegistries = append(serviceRegistries, provider.ID(c))
			}
		}
		return true
	})

	// only one meshnetworks config should exist.
	c.ForEach(gvk.MeshNetworks, func(r *resource.Instance) bool {
		mn := r.Message.(*v1alpha1.MeshNetworks)
		for i, n := range mn.Networks {
			for j, e := range n.Endpoints {
				switch re := e.Ne.(type) {
				case *v1alpha1.Network_NetworkEndpoints_FromRegistry:
					found := false
					for _, s := range serviceRegistries {
						if provider.ID(re.FromRegistry) == s {
							found = true
						}
					}
					if !found {
						m := msg.NewUnknownMeshNetworksServiceRegistry(r, re.FromRegistry, i)

						if line, ok := util.ErrorLine(r, fmt.Sprintf(util.FromRegistry, i, j)); ok {
							m.Line = line
						}

						c.Report(gvk.MeshNetworks, m)
					}
				}
			}
		}
		return true
	})
}
