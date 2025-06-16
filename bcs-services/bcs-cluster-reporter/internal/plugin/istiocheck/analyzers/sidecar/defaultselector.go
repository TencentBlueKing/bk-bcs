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

package sidecar

import (
	"istio.io/api/networking/v1alpha3"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/analysis"
	"istio.io/istio/pkg/config/resource"
	"istio.io/istio/pkg/config/schema/gvk"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/msg"
)

// DefaultSelectorAnalyzer validates, per namespace, that there aren't multiple
// sidecar resources that have no selector. This is distinct from
// SelectorAnalyzer because it does not require pods, so it can run even if that
// collection is unavailable.
type DefaultSelectorAnalyzer struct{}

var _ analysis.Analyzer = &DefaultSelectorAnalyzer{}

// Metadata implements Analyzer
func (a *DefaultSelectorAnalyzer) Metadata() analysis.Metadata {
	return analysis.Metadata{
		Name:        "sidecar.DefaultSelectorAnalyzer",
		Description: "Validates that there aren't multiple sidecar resources that have no selector",
		Inputs: []config.GroupVersionKind{
			gvk.Sidecar,
		},
	}
}

// Analyze implements Analyzer
func (a *DefaultSelectorAnalyzer) Analyze(c analysis.Context) {
	nsToSidecars := make(map[resource.Namespace][]*resource.Instance)

	c.ForEach(gvk.Sidecar, func(r *resource.Instance) bool {
		s := r.Message.(*v1alpha3.Sidecar)

		ns := r.Metadata.FullName.Namespace

		if s.WorkloadSelector == nil {
			nsToSidecars[ns] = append(nsToSidecars[ns], r)
		}
		return true
	})

	// Check for more than one selector-less sidecar instance, per namespace
	for ns, sList := range nsToSidecars {
		if len(sList) > 1 {
			sNames := getNames(sList)
			for _, r := range sList {
				c.Report(gvk.Sidecar, msg.NewMultipleSidecarsWithoutWorkloadSelectors(r, sNames, string(ns)))
			}
		}
	}
}
