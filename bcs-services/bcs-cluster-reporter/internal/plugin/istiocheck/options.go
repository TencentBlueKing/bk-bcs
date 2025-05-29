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

package istio

import (
	"fmt"

	"istio.io/istio/pkg/config/analysis"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/analyzers"
)

var defaultInterval = 60
var defaultEnabledAnalyzers = []string{
	"annotations.K8sAnalyzer",
	"authz.AuthorizationPoliciesAnalyzer",
	"deployment.ServiceAssociationAnalyzer",
	"deployment.ApplicationUIDAnalyzer",
	"deprecation.FieldAnalyzer",
	"gateway.IngressGatewayPortAnalyzer",
	"gateway.CertificateAnalyzer",
	"gateway.SecretAnalyzer",
	"gateway.ConflictingGatewayAnalyzer",
	"injection.ImageAnalyzer",
	"injection.ImageAutoAnalyzer",
	"multicluster.MeshNetworksAnalyzer",
	"service.PortNameAnalyzer",
	"sidecar.DefaultSelectorAnalyzer",
	"sidecar.SelectorAnalyzer",
	"virtualservice.ConflictingMeshGatewayHostsAnalyzer",
	"virtualservice.DestinationHostAnalyzer",
	"virtualservice.DestinationRuleAnalyzer",
	"virtualservice.GatewayAnalyzer",
	"virtualservice.JWTClaimRouteAnalyzer",
	"virtualservice.RegexAnalyzer",
	"destinationrule.CaCertificateAnalyzer",
	"serviceentry.ProtocolAddressesAnalyzer",
	"webhook.Analyzer",
	"envoyfilter.EnvoyPatchAnalyzer",
	"telemetry.ProviderAnalyzer",
	"telemetry.SelectorAnalyzer",
	"telemetry.DefaultSelectorAnalyzer",
	"telemetry.LightstepAnalyzer",
}

// Options is the options for the istio check plugin
type Options struct {
	EnabledAnalyzers []string `json:"enabledAnalyzers" yaml:"enabledAnalyzers"`
	Interval         int      `json:"interval" yaml:"interval"`
	IstioNamespace   string   `json:"istioNamespace" yaml:"istioNamespace"`

	enabledAnalyzersObject []analysis.Analyzer
}

// Validate validate options
func (o *Options) Validate() error {
	// 如果未指定analyzer，则使用默认的
	if len(o.EnabledAnalyzers) == 0 {
		klog.Warning("no analyzer specified, use default analyzers")
		o.EnabledAnalyzers = defaultEnabledAnalyzers
	}

	klog.Infof("enabledAnalyzers: %v", o.EnabledAnalyzers)
	for _, analyzer := range o.EnabledAnalyzers {
		if analyzerObj, ok := analyzers.AnalyzerNameMap[analyzer]; ok {
			o.enabledAnalyzersObject = append(o.enabledAnalyzersObject, analyzerObj)
		} else {
			return fmt.Errorf("analyzer %s not found", analyzer)
		}
	}

	klog.Infof("enabledAnalyzersObject: %v", o.enabledAnalyzersObject)
	if o.Interval <= 0 {
		klog.Warningf("interval is less than 0, set to 60")
		o.Interval = defaultInterval
	}
	if o.IstioNamespace == "" {
		klog.Warningf("istioNamespace is empty, set to istio-system")
		o.IstioNamespace = "istio-system"
	}
	return nil
}
