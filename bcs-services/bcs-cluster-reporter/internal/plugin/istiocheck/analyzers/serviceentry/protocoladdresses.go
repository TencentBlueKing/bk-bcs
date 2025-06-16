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

// Package serviceentry 提供 Istio 的服务入口分析器
package serviceentry

import (
	"fmt"

	meshconfig "istio.io/api/mesh/v1alpha1"
	"istio.io/api/networking/v1alpha3"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/analysis"
	"istio.io/istio/pkg/config/analysis/analyzers/util"
	"istio.io/istio/pkg/config/resource"
	"istio.io/istio/pkg/config/schema/gvk"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/msg"
)

// ProtocolAddressesAnalyzer 协议地址分析器
type ProtocolAddressesAnalyzer struct{}

var _ analysis.Analyzer = &ProtocolAddressesAnalyzer{}

// Metadata 返回分析器的元数据
func (serviceEntry *ProtocolAddressesAnalyzer) Metadata() analysis.Metadata {
	return analysis.Metadata{
		Name:        "serviceentry.Analyzer",
		Description: "Checks the validity of ServiceEntry",
		Inputs: []config.GroupVersionKind{
			gvk.ServiceEntry,
			gvk.MeshConfig,
		},
	}
}

// Analyze 分析器的主要逻辑
func (serviceEntry *ProtocolAddressesAnalyzer) Analyze(context analysis.Context) {
	autoAllocated := false
	context.ForEach(gvk.MeshConfig, func(r *resource.Instance) bool {
		mc := r.Message.(*meshconfig.MeshConfig)
		if v, ok := mc.DefaultConfig.ProxyMetadata["ISTIO_META_DNS_CAPTURE"]; !ok || v != "true" {
			return true
		}
		if v, ok := mc.DefaultConfig.ProxyMetadata["ISTIO_META_DNS_AUTO_ALLOCATE"]; ok && v == "true" {
			autoAllocated = true
		}
		return true
	})

	context.ForEach(gvk.ServiceEntry, func(resource *resource.Instance) bool {
		serviceEntry.analyzeProtocolAddresses(resource, context, autoAllocated)
		return true
	})
}

func (serviceEntry *ProtocolAddressesAnalyzer) analyzeProtocolAddresses(r *resource.Instance, ctx analysis.Context, metaDNSAutoAllocated bool) {
	se := r.Message.(*v1alpha3.ServiceEntry)
	if se.Addresses == nil && !metaDNSAutoAllocated {
		for index, port := range se.Ports {
			if port.Protocol == "" || port.Protocol == "TCP" {
				message := msg.NewServiceEntryAddressesRequired(r)

				if line, ok := util.ErrorLine(r, fmt.Sprintf(util.ServiceEntryPort, index)); ok {
					message.Line = line
				}

				ctx.Report(gvk.ServiceEntry, message)
			}
		}
	}
}
