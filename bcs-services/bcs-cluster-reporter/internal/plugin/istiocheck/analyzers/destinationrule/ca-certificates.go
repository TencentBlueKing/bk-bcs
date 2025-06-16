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

// Package destinationrule 提供 Istio 的目的地规则分析器
package destinationrule

import (
	"fmt"

	"istio.io/api/networking/v1alpha3"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/analysis"
	"istio.io/istio/pkg/config/analysis/analyzers/util"
	"istio.io/istio/pkg/config/resource"
	"istio.io/istio/pkg/config/schema/gvk"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/msg"
)

// CaCertificateAnalyzer checks if CaCertificate is set in case mode is SIMPLE/MUTUAL
type CaCertificateAnalyzer struct{}

var _ analysis.Analyzer = &CaCertificateAnalyzer{}

// Metadata 返回分析器的元数据
func (c *CaCertificateAnalyzer) Metadata() analysis.Metadata {
	return analysis.Metadata{
		Name:        "destinationrule.CaCertificateAnalyzer",
		Description: "Checks if caCertificates is set when TLS mode is SIMPLE/MUTUAL",
		Inputs: []config.GroupVersionKind{
			gvk.DestinationRule,
		},
	}
}

// Analyze 分析器的主要逻辑
func (c *CaCertificateAnalyzer) Analyze(ctx analysis.Context) {
	ctx.ForEach(gvk.DestinationRule, func(r *resource.Instance) bool {
		c.analyzeDestinationRule(r, ctx)
		return true
	})
}

func (c *CaCertificateAnalyzer) analyzeDestinationRule(r *resource.Instance, ctx analysis.Context) {
	dr := r.Message.(*v1alpha3.DestinationRule)
	drNs := r.Metadata.FullName.Namespace
	drName := r.Metadata.FullName.String()
	mode := dr.GetTrafficPolicy().GetTls().GetMode()

	if mode == v1alpha3.ClientTLSSettings_SIMPLE || mode == v1alpha3.ClientTLSSettings_MUTUAL {
		if dr.GetTrafficPolicy().GetTls().GetCaCertificates() == "" {
			m := msg.NewNoServerCertificateVerificationDestinationLevel(r, drName,
				drNs.String(), mode.String(), dr.GetHost())

			if line, ok := util.ErrorLine(r, fmt.Sprintf(util.DestinationRuleTLSCert)); ok {
				m.Line = line
			}
			ctx.Report(gvk.DestinationRule, m)
		}
	}
	portSettings := dr.TrafficPolicy.GetPortLevelSettings()

	for i, p := range portSettings {
		mode = p.GetTls().GetMode()
		if mode == v1alpha3.ClientTLSSettings_SIMPLE || mode == v1alpha3.ClientTLSSettings_MUTUAL {
			if p.GetTls().GetCaCertificates() == "" {
				m := msg.NewNoServerCertificateVerificationPortLevel(r, drName,
					drNs.String(), mode.String(), dr.GetHost(), p.GetPort().String())

				if line, ok := util.ErrorLine(r, fmt.Sprintf(util.DestinationRuleTLSPortLevelCert, i)); ok {
					m.Line = line
				}
				ctx.Report(gvk.DestinationRule, m)
			}
		}
	}
}
