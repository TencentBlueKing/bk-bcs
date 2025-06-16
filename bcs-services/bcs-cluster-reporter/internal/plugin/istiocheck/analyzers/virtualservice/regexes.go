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
	"regexp"

	"istio.io/api/networking/v1alpha3"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/analysis"
	"istio.io/istio/pkg/config/analysis/analyzers/util"
	"istio.io/istio/pkg/config/resource"
	"istio.io/istio/pkg/config/schema/gvk"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/msg"
)

// RegexAnalyzer checks all regexes in a virtual service
type RegexAnalyzer struct{}

var _ analysis.Analyzer = &RegexAnalyzer{}

// Metadata implements Analyzer
func (a *RegexAnalyzer) Metadata() analysis.Metadata {
	return analysis.Metadata{
		Name:        "virtualservice.RegexAnalyzer",
		Description: "Checks regex syntax",
		Inputs: []config.GroupVersionKind{
			gvk.VirtualService,
		},
	}
}

// Analyze implements Analyzer
func (a *RegexAnalyzer) Analyze(ctx analysis.Context) {
	ctx.ForEach(gvk.VirtualService, func(r *resource.Instance) bool {
		a.analyzeVirtualService(r, ctx)
		return true
	})
}

func (a *RegexAnalyzer) analyzeVirtualService(r *resource.Instance, ctx analysis.Context) {
	vs := r.Message.(*v1alpha3.VirtualService)

	for i, route := range vs.GetHttp() {
		for j, m := range route.GetMatch() {

			analyzeStringMatch(r, m.GetUri(), ctx, "uri",
				fmt.Sprintf(util.URISchemeMethodAuthorityRegexMatch, i, j, "uri"))
			analyzeStringMatch(r, m.GetScheme(), ctx, "scheme",
				fmt.Sprintf(util.URISchemeMethodAuthorityRegexMatch, i, j, "scheme"))
			analyzeStringMatch(r, m.GetMethod(), ctx, "method",
				fmt.Sprintf(util.URISchemeMethodAuthorityRegexMatch, i, j, "method"))
			analyzeStringMatch(r, m.GetAuthority(), ctx, "authority",
				fmt.Sprintf(util.URISchemeMethodAuthorityRegexMatch, i, j, "authority"))
			for key, h := range m.GetHeaders() {
				analyzeStringMatch(r, h, ctx, "headers",
					fmt.Sprintf(util.HeaderAndQueryParamsRegexMatch, i, j, "headers", key))
			}
			for key, qp := range m.GetQueryParams() {
				analyzeStringMatch(r, qp, ctx, "queryParams",
					fmt.Sprintf(util.HeaderAndQueryParamsRegexMatch, i, j, "queryParams", key))
			}
			// We don't validate withoutHeaders, because they are undocumented
		}
		for j, origin := range route.GetCorsPolicy().GetAllowOrigins() {
			analyzeStringMatch(r, origin, ctx, "corsPolicy.allowOrigins",
				fmt.Sprintf(util.AllowOriginsRegexMatch, i, j))
		}
	}
}

func analyzeStringMatch(r *resource.Instance, sm *v1alpha3.StringMatch, ctx analysis.Context, where string, key string) {
	re := sm.GetRegex()
	if re == "" {
		return
	}

	_, err := regexp.Compile(re)
	if err == nil {
		return
	}

	m := msg.NewInvalidRegexp(r, where, re, err.Error())

	// Get line number for different match field
	if line, ok := util.ErrorLine(r, key); ok {
		m.Line = line
	}

	ctx.Report(gvk.VirtualService, m)
}
