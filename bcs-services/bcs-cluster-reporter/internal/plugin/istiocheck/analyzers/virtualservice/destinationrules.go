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

	"istio.io/api/networking/v1alpha3"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/analysis"
	"istio.io/istio/pkg/config/analysis/analyzers/util"
	"istio.io/istio/pkg/config/resource"
	"istio.io/istio/pkg/config/schema/gvk"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/msg"
)

// DestinationRuleAnalyzer checks the destination rules associated with each virtual service
type DestinationRuleAnalyzer struct{}

var _ analysis.Analyzer = &DestinationRuleAnalyzer{}

// Metadata implements Analyzer
func (d *DestinationRuleAnalyzer) Metadata() analysis.Metadata {
	return analysis.Metadata{
		Name:        "virtualservice.DestinationRuleAnalyzer",
		Description: "Checks the destination rules associated with each virtual service",
		Inputs: []config.GroupVersionKind{
			gvk.VirtualService,
			gvk.DestinationRule,
		},
	}
}

// Analyze implements Analyzer
func (d *DestinationRuleAnalyzer) Analyze(ctx analysis.Context) {
	// To avoid repeated iteration, precompute the set of existing destination host+subset combinations
	destHostsAndSubsets := initDestHostsAndSubsets(ctx)

	ctx.ForEach(gvk.VirtualService, func(r *resource.Instance) bool {
		d.analyzeVirtualService(r, ctx, destHostsAndSubsets)
		return true
	})
}

func (d *DestinationRuleAnalyzer) analyzeVirtualService(r *resource.Instance, ctx analysis.Context,
	destHostsAndSubsets map[hostAndSubset]bool,
) {
	vs := r.Message.(*v1alpha3.VirtualService)
	ns := r.Metadata.FullName.Namespace

	for _, ad := range getRouteDestinations(vs) {
		if !d.checkDestinationSubset(ns, ad.Destination, destHostsAndSubsets) {

			m := msg.NewReferencedResourceNotFound(r, "host+subset in destinationrule",
				fmt.Sprintf("%s+%s", ad.Destination.GetHost(), ad.Destination.GetSubset()))

			key := fmt.Sprintf(util.DestinationHost, ad.RouteRule, ad.ServiceIndex, ad.DestinationIndex)
			if line, ok := util.ErrorLine(r, key); ok {
				m.Line = line
			}

			ctx.Report(gvk.VirtualService, m)
		}
	}

	for _, ad := range getHTTPMirrorDestinations(vs) {
		if !d.checkDestinationSubset(ns, ad.Destination, destHostsAndSubsets) {

			m := msg.NewReferencedResourceNotFound(r, "mirror+subset in destinationrule",
				fmt.Sprintf("%s+%s", ad.Destination.GetHost(), ad.Destination.GetSubset()))

			key := fmt.Sprintf(util.MirrorHost, ad.ServiceIndex)
			if line, ok := util.ErrorLine(r, key); ok {
				m.Line = line
			}

			ctx.Report(gvk.VirtualService, m)
		}
	}
}

func (d *DestinationRuleAnalyzer) checkDestinationSubset(vsNamespace resource.Namespace, destination *v1alpha3.Destination,
	destHostsAndSubsets map[hostAndSubset]bool,
) bool {
	name := util.GetResourceNameFromHost(vsNamespace, destination.GetHost())
	subset := destination.GetSubset()

	// if there's no subset specified, we're done
	if subset == "" {
		return true
	}

	hs := hostAndSubset{
		host:   name,
		subset: subset,
	}
	if _, ok := destHostsAndSubsets[hs]; ok {
		return true
	}

	return false
}

func initDestHostsAndSubsets(ctx analysis.Context) map[hostAndSubset]bool {
	hostsAndSubsets := make(map[hostAndSubset]bool)
	ctx.ForEach(gvk.DestinationRule, func(r *resource.Instance) bool {
		dr := r.Message.(*v1alpha3.DestinationRule)
		drNamespace := r.Metadata.FullName.Namespace

		for _, ss := range dr.GetSubsets() {
			hs := hostAndSubset{
				host:   util.GetResourceNameFromHost(drNamespace, dr.GetHost()),
				subset: ss.GetName(),
			}
			hostsAndSubsets[hs] = true
		}
		return true
	})
	return hostsAndSubsets
}
