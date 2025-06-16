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

// Package analyzers 提供 Istio 的分析器
package analyzers

import (
	"istio.io/istio/pkg/config/analysis"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/analyzers/annotations"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/analyzers/authz"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/analyzers/deployment"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/analyzers/deprecation"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/analyzers/destinationrule"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/analyzers/envoyfilter"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/analyzers/gateway"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/analyzers/injection"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/analyzers/multicluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/analyzers/schema"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/analyzers/service"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/analyzers/serviceentry"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/analyzers/sidecar"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/analyzers/telemetry"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/analyzers/virtualservice"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/analyzers/webhook"
)

// AnalyzerNameMap 用于将analyzer的名称映射到analyzer对象
var AnalyzerNameMap = map[string]analysis.Analyzer{
	"annotations.K8sAnalyzer":                            &annotations.K8sAnalyzer{},
	"authz.AuthorizationPoliciesAnalyzer":                &authz.AuthorizationPoliciesAnalyzer{},
	"deployment.ServiceAssociationAnalyzer":              &deployment.ServiceAssociationAnalyzer{},
	"deployment.ApplicationUIDAnalyzer":                  &deployment.ApplicationUIDAnalyzer{},
	"deprecation.FieldAnalyzer":                          &deprecation.FieldAnalyzer{},
	"gateway.IngressGatewayPortAnalyzer":                 &gateway.IngressGatewayPortAnalyzer{},
	"gateway.CertificateAnalyzer":                        &gateway.CertificateAnalyzer{},
	"gateway.SecretAnalyzer":                             &gateway.SecretAnalyzer{},
	"gateway.ConflictingGatewayAnalyzer":                 &gateway.ConflictingGatewayAnalyzer{},
	"injection.Analyzer":                                 &injection.Analyzer{},
	"injection.ImageAnalyzer":                            &injection.ImageAnalyzer{},
	"injection.ImageAutoAnalyzer":                        &injection.ImageAutoAnalyzer{},
	"multicluster.MeshNetworksAnalyzer":                  &multicluster.MeshNetworksAnalyzer{},
	"service.PortNameAnalyzer":                           &service.PortNameAnalyzer{},
	"sidecar.DefaultSelectorAnalyzer":                    &sidecar.DefaultSelectorAnalyzer{},
	"sidecar.SelectorAnalyzer":                           &sidecar.SelectorAnalyzer{},
	"virtualservice.ConflictingMeshGatewayHostsAnalyzer": &virtualservice.ConflictingMeshGatewayHostsAnalyzer{},
	"virtualservice.DestinationHostAnalyzer":             &virtualservice.DestinationHostAnalyzer{},
	"virtualservice.DestinationRuleAnalyzer":             &virtualservice.DestinationRuleAnalyzer{},
	"virtualservice.GatewayAnalyzer":                     &virtualservice.GatewayAnalyzer{},
	"virtualservice.JWTClaimRouteAnalyzer":               &virtualservice.JWTClaimRouteAnalyzer{},
	"virtualservice.RegexAnalyzer":                       &virtualservice.RegexAnalyzer{},
	"destinationrule.CaCertificateAnalyzer":              &destinationrule.CaCertificateAnalyzer{},
	"serviceentry.ProtocolAddressesAnalyzer":             &serviceentry.ProtocolAddressesAnalyzer{},
	"webhook.Analyzer":                                   &webhook.Analyzer{},
	"envoyfilter.EnvoyPatchAnalyzer":                     &envoyfilter.EnvoyPatchAnalyzer{},
	"telemetry.ProviderAnalyzer":                         &telemetry.ProviderAnalyzer{},
	"telemetry.SelectorAnalyzer":                         &telemetry.SelectorAnalyzer{},
	"telemetry.DefaultSelectorAnalyzer":                  &telemetry.DefaultSelectorAnalyzer{},
	"telemetry.LightstepAnalyzer":                        &telemetry.LightstepAnalyzer{},
}

// All 返回所有分析器。
func All() []analysis.Analyzer {
	analyzers := []analysis.Analyzer{
		// Please keep this list sorted alphabetically by pkg.name for convenience
		&annotations.K8sAnalyzer{},
		&authz.AuthorizationPoliciesAnalyzer{},
		&deployment.ServiceAssociationAnalyzer{},
		&deployment.ApplicationUIDAnalyzer{},
		&deprecation.FieldAnalyzer{},
		&gateway.IngressGatewayPortAnalyzer{},
		&gateway.CertificateAnalyzer{},
		&gateway.SecretAnalyzer{},
		&gateway.ConflictingGatewayAnalyzer{},
		&injection.Analyzer{},
		&injection.ImageAnalyzer{},
		&injection.ImageAutoAnalyzer{},
		&multicluster.MeshNetworksAnalyzer{},
		&service.PortNameAnalyzer{},
		&sidecar.DefaultSelectorAnalyzer{},
		&sidecar.SelectorAnalyzer{},
		&virtualservice.ConflictingMeshGatewayHostsAnalyzer{},
		&virtualservice.DestinationHostAnalyzer{},
		&virtualservice.DestinationRuleAnalyzer{},
		&virtualservice.GatewayAnalyzer{},
		&virtualservice.JWTClaimRouteAnalyzer{},
		&virtualservice.RegexAnalyzer{},
		&destinationrule.CaCertificateAnalyzer{},
		&serviceentry.ProtocolAddressesAnalyzer{},
		&webhook.Analyzer{},
		&envoyfilter.EnvoyPatchAnalyzer{},
		&telemetry.ProviderAnalyzer{},
		&telemetry.SelectorAnalyzer{},
		&telemetry.DefaultSelectorAnalyzer{},
		&telemetry.LightstepAnalyzer{},
	}

	analyzers = append(analyzers, schema.AllValidationAnalyzers()...)

	return analyzers
}

// AllCombined 返回所有合并的分析器。
func AllCombined() *analysis.CombinedAnalyzer {
	return analysis.Combine("all", All()...)
}
