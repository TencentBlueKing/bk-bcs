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

package metrics

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
)

var wantCertLabels = []string{
	"owner_kind", "owner_namespace", "owner_name", "cert_id", "cert_role",
	"cert_scope", "protocol", "port", "domain",
}

func TestCertificateMetricsRegistration(t *testing.T) {
	cases := []struct {
		name       string
		collector  prometheus.Collector
		metricName string
		labels     []string
	}{
		{
			name:       "days_until_expiry",
			collector:  CertificateDaysUntilExpiry,
			metricName: "bkbcs_ingressctrl_certificate_days_until_expiry",
			labels:     wantCertLabels,
		},
		{
			name:       "query_success",
			collector:  CertificateQuerySuccess,
			metricName: "bkbcs_ingressctrl_certificate_query_success",
			labels:     wantCertLabels,
		},
		{
			name:       "bindings_total",
			collector:  CertificateBindingsTotal,
			metricName: "bkbcs_ingressctrl_certificate_bindings_total",
			labels:     nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			descCh := make(chan *prometheus.Desc, 1)
			c.collector.Describe(descCh)
			close(descCh)
			desc := <-descCh
			if desc == nil {
				t.Fatal("expected metric descriptor")
			}
			fqName := desc.String()
			if !strings.Contains(fqName, c.metricName) {
				t.Fatalf("descriptor %q does not contain metric name %q", fqName, c.metricName)
			}
			if c.labels == nil {
				return
			}
			for _, label := range c.labels {
				if !strings.Contains(fqName, label) {
					t.Fatalf("descriptor %q missing label %q", fqName, label)
				}
			}
		})
	}
}

func TestCertificateMetricHelpers(t *testing.T) {
	SetCertificateDaysUntilExpiry(constant.KindIngress, "ns1", "ing1", "cert1", "server", "rule", "HTTPS", "443", "", 30)
	SetCertificateQuerySuccess(constant.KindIngress, "ns1", "ing1", "cert1", "server", "rule", "HTTPS", "443", "", 1)
	SetCertificateBindingsTotal(1)

	if got := testutil.ToFloat64(CertificateDaysUntilExpiry.WithLabelValues(
		constant.KindIngress, "ns1", "ing1", "cert1", "server", "rule", "HTTPS", "443", "")); got != 30 {
		t.Fatalf("days_until_expiry = %v, want 30", got)
	}
	if got := testutil.ToFloat64(CertificateQuerySuccess.WithLabelValues(
		constant.KindIngress, "ns1", "ing1", "cert1", "server", "rule", "HTTPS", "443", "")); got != 1 {
		t.Fatalf("query_success = %v, want 1", got)
	}
	if got := testutil.ToFloat64(CertificateBindingsTotal); got != 1 {
		t.Fatalf("bindings_total = %v, want 1", got)
	}

	DeleteCertificateDaysUntilExpiry(constant.KindIngress, "ns1", "ing1", "cert1", "server", "rule", "HTTPS", "443", "")
	if got := testutil.ToFloat64(CertificateDaysUntilExpiry.WithLabelValues(
		constant.KindIngress, "ns1", "ing1", "cert1", "server", "rule", "HTTPS", "443", "")); got != 0 {
		t.Fatalf("days_until_expiry after delete = %v, want 0", got)
	}

	DeleteCertificateMetrics(constant.KindIngress, "ns1", "ing1", "cert1", "server", "rule", "HTTPS", "443", "")
	if got := testutil.ToFloat64(CertificateQuerySuccess.WithLabelValues(
		constant.KindIngress, "ns1", "ing1", "cert1", "server", "rule", "HTTPS", "443", "")); got != 0 {
		t.Fatalf("query_success after delete = %v, want 0", got)
	}
}
