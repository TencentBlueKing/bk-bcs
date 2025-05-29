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

// Package sidecar 提供 Istio 的 sidecar 分析器
package sidecar

import (
	"fmt"

	"k8s.io/apimachinery/pkg/labels"

	"istio.io/api/networking/v1alpha3"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/analysis"
	"istio.io/istio/pkg/config/analysis/analyzers/util"
	"istio.io/istio/pkg/config/resource"
	"istio.io/istio/pkg/config/schema/gvk"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/msg"
)

// SelectorAnalyzer validates, per namespace, that:
// * sidecar resources that define a workload selector match at least one pod
// * there aren't multiple sidecar resources that select overlapping pods
type SelectorAnalyzer struct{}

var _ analysis.Analyzer = &SelectorAnalyzer{}

// Metadata implements Analyzer
func (a *SelectorAnalyzer) Metadata() analysis.Metadata {
	return analysis.Metadata{
		Name: "sidecar.SelectorAnalyzer",
		Description: "Validates that sidecars that define a workload selector " +
			"match at least one pod, and that there aren't multiple sidecar resources that select overlapping pods",
		Inputs: []config.GroupVersionKind{
			gvk.Sidecar,
			gvk.Pod,
		},
	}
}

// Analyze implements Analyzer
func (a *SelectorAnalyzer) Analyze(c analysis.Context) {
	podsToSidecars := make(map[resource.FullName][]*resource.Instance)

	// This is using an unindexed approach for matching selectors.
	// Using an index for selectoes is problematic because selector != label
	// We can match a label to a selector, but we can't generate a selector from a label.
	c.ForEach(gvk.Sidecar, func(rs *resource.Instance) bool {
		s := rs.Message.(*v1alpha3.Sidecar)

		// For this analysis, ignore Sidecars with no workload selectors specified at all.
		if s.WorkloadSelector == nil || len(s.WorkloadSelector.Labels) == 0 {
			return true
		}

		sNs := rs.Metadata.FullName.Namespace
		sel := labels.SelectorFromSet(s.WorkloadSelector.Labels)

		foundPod := false
		c.ForEach(gvk.Pod, func(rp *resource.Instance) bool {
			pNs := rp.Metadata.FullName.Namespace
			podLabels := labels.Set(rp.Metadata.Labels)

			// Only attempt to match in the same namespace
			if pNs != sNs {
				return true
			}

			if sel.Matches(podLabels) {
				foundPod = true
				podsToSidecars[rp.Metadata.FullName] = append(podsToSidecars[rp.Metadata.FullName], rs)
			}

			return true
		})

		if !foundPod {
			m := msg.NewReferencedResourceNotFound(rs, "selector", sel.String())

			label := util.ExtractLabelFromSelectorString(sel.String())
			if line, ok := util.ErrorLine(rs, fmt.Sprintf(util.WorkloadSelector, label)); ok {
				m.Line = line
			}

			c.Report(gvk.Sidecar, m)
		}

		return true
	})

	for p, sList := range podsToSidecars {
		if len(sList) == 1 {
			continue
		}

		sNames := getNames(sList)

		for _, rs := range sList {

			m := msg.NewConflictingSidecarWorkloadSelectors(rs, sNames,
				p.Namespace.String(), p.Name.String())

			if line, ok := util.ErrorLine(rs, fmt.Sprintf(util.MetadataName)); ok {
				m.Line = line
			}

			c.Report(gvk.Sidecar, m)
		}
	}
}
