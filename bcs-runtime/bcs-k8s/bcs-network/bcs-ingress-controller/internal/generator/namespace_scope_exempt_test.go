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

package generator

import (
	"testing"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// TestRuleConverter_isIngressNamespaceExempt verifies the exemption check considers
// both the ingress namespace and whether an exempt set was configured.
func TestRuleConvNsExempt(t *testing.T) {
	cases := []struct {
		name             string
		ingressNamespace string
		exemptNamespaces map[string]struct{}
		want             bool
	}{
		{
			name:             "nil exempt map returns false",
			ingressNamespace: "bcs-system",
			exemptNamespaces: nil,
			want:             false,
		},
		{
			name:             "empty exempt map returns false",
			ingressNamespace: "bcs-system",
			exemptNamespaces: map[string]struct{}{},
			want:             false,
		},
		{
			name:             "ingress namespace in exempt list returns true",
			ingressNamespace: "bcs-system",
			exemptNamespaces: map[string]struct{}{"bcs-system": {}, "kube-system": {}},
			want:             true,
		},
		{
			name:             "ingress namespace not in exempt list returns false",
			ingressNamespace: "user-ns",
			exemptNamespaces: map[string]struct{}{"bcs-system": {}},
			want:             false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rc := &RuleConverter{
				ingressNamespace: c.ingressNamespace,
				exemptNamespaces: c.exemptNamespaces,
			}
			if got := rc.isIngressNamespaceExempt(); got != c.want {
				t.Errorf("isIngressNamespaceExempt() = %v, want %v", got, c.want)
			}
		})
	}
}

// TestRuleConverter_getServiceNamespace covers the combinations of
// isNamespaced x exemptNamespaces x ServiceRoute.ServiceNamespace that
// drive cross-namespace binding decisions.
func TestRuleConvSvcNamespace(t *testing.T) {
	cases := []struct {
		name             string
		ingressNamespace string
		isNamespaced     bool
		exemptNamespaces map[string]struct{}
		serviceNamespace string
		want             string
	}{
		{
			name:             "not namespaced, empty svc ns falls back to ingress ns",
			ingressNamespace: "user-ns",
			isNamespaced:     false,
			serviceNamespace: "",
			want:             "user-ns",
		},
		{
			name:             "not namespaced, explicit svc ns kept",
			ingressNamespace: "user-ns",
			isNamespaced:     false,
			serviceNamespace: "other-ns",
			want:             "other-ns",
		},
		{
			name:             "namespaced, not exempt, cross ns overridden to ingress ns",
			ingressNamespace: "user-ns",
			isNamespaced:     true,
			serviceNamespace: "other-ns",
			want:             "user-ns",
		},
		{
			name:             "namespaced, exempt, cross ns preserved",
			ingressNamespace: "bcs-system",
			isNamespaced:     true,
			exemptNamespaces: map[string]struct{}{"bcs-system": {}},
			serviceNamespace: "user-ns",
			want:             "user-ns",
		},
		{
			name:             "namespaced, exempt but svc ns empty still falls back to ingress ns",
			ingressNamespace: "bcs-system",
			isNamespaced:     true,
			exemptNamespaces: map[string]struct{}{"bcs-system": {}},
			serviceNamespace: "",
			want:             "bcs-system",
		},
		{
			name:             "namespaced, exempt set but ingress ns not in list, cross ns overridden",
			ingressNamespace: "user-ns",
			isNamespaced:     true,
			exemptNamespaces: map[string]struct{}{"bcs-system": {}},
			serviceNamespace: "other-ns",
			want:             "user-ns",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rc := &RuleConverter{
				ingressNamespace: c.ingressNamespace,
				isNamespaced:     c.isNamespaced,
				exemptNamespaces: c.exemptNamespaces,
			}
			svcRoute := &networkextensionv1.ServiceRoute{
				ServiceNamespace: c.serviceNamespace,
			}
			if got := rc.getServiceNamespace(svcRoute); got != c.want {
				t.Errorf("getServiceNamespace() = %q, want %q", got, c.want)
			}
		})
	}
}

// TestRuleConverter_SetExemptNamespaces checks the setter assigns the set verbatim.
func TestRuleConvSetExemptNs(t *testing.T) {
	rc := &RuleConverter{}
	exempt := map[string]struct{}{"bcs-system": {}}
	rc.SetExemptNamespaces(exempt)
	if _, ok := rc.exemptNamespaces["bcs-system"]; !ok {
		t.Fatalf("SetExemptNamespaces did not set the namespace set correctly")
	}
}

// TestMappingConverter_isIngressNamespaceExempt verifies the exemption check for MappingConverter.
func TestMappingConvNsExempt(t *testing.T) {
	cases := []struct {
		name             string
		ingressNamespace string
		exemptNamespaces map[string]struct{}
		want             bool
	}{
		{
			name:             "nil exempt map returns false",
			ingressNamespace: "bcs-system",
			exemptNamespaces: nil,
			want:             false,
		},
		{
			name:             "namespace in exempt list returns true",
			ingressNamespace: "bcs-system",
			exemptNamespaces: map[string]struct{}{"bcs-system": {}},
			want:             true,
		},
		{
			name:             "namespace not in exempt list returns false",
			ingressNamespace: "user-ns",
			exemptNamespaces: map[string]struct{}{"bcs-system": {}},
			want:             false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mg := &MappingConverter{
				ingressNamespace: c.ingressNamespace,
				exemptNamespaces: c.exemptNamespaces,
			}
			if got := mg.isIngressNamespaceExempt(); got != c.want {
				t.Errorf("isIngressNamespaceExempt() = %v, want %v", got, c.want)
			}
		})
	}
}

// TestMappingConverter_workloadNamespaceResolution simulates the logic in DoConvert
// to confirm that the exempt flag controls whether WorkloadNamespace is overridden.
func TestMappingConvWorkloadNs(t *testing.T) {
	cases := []struct {
		name              string
		ingressNamespace  string
		isNamespaced      bool
		exemptNamespaces  map[string]struct{}
		workloadNamespace string
		wantNamespace     string
	}{
		{
			name:              "not namespaced keeps workload namespace",
			ingressNamespace:  "user-ns",
			isNamespaced:      false,
			workloadNamespace: "other-ns",
			wantNamespace:     "other-ns",
		},
		{
			name:              "namespaced non-exempt overrides workload namespace",
			ingressNamespace:  "user-ns",
			isNamespaced:      true,
			workloadNamespace: "other-ns",
			wantNamespace:     "user-ns",
		},
		{
			name:              "namespaced exempt preserves workload namespace",
			ingressNamespace:  "bcs-system",
			isNamespaced:      true,
			exemptNamespaces:  map[string]struct{}{"bcs-system": {}},
			workloadNamespace: "user-ns",
			wantNamespace:     "user-ns",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mg := &MappingConverter{
				ingressNamespace: c.ingressNamespace,
				isNamespaced:     c.isNamespaced,
				exemptNamespaces: c.exemptNamespaces,
				mapping: &networkextensionv1.IngressPortMapping{
					WorkloadNamespace: c.workloadNamespace,
				},
			}

			workloadNamespace := mg.mapping.WorkloadNamespace
			if mg.isNamespaced && !mg.isIngressNamespaceExempt() {
				workloadNamespace = mg.ingressNamespace
			}

			if workloadNamespace != c.wantNamespace {
				t.Errorf("resolved workloadNamespace = %q, want %q", workloadNamespace, c.wantNamespace)
			}
		})
	}
}
