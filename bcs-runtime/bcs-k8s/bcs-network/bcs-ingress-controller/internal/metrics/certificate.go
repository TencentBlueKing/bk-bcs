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
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// Certificate metric label names shared by days_until_expiry and query_success.
var certificateLabelNames = []string{
	"owner_kind", "owner_namespace", "owner_name", "cert_id", "cert_role",
	"cert_scope", "protocol", "port", "domain",
}

var (
	// CertificateDaysUntilExpiry tracks remaining days until SSL certificate expiry per binding.
	CertificateDaysUntilExpiry = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "certificate",
		Name:      "days_until_expiry",
		Help:      "Remaining days until SSL certificate expiry for a certificate binding",
	}, certificateLabelNames)

	// CertificateQuerySuccess indicates whether certificate expiry was queried successfully (1) or not (0).
	CertificateQuerySuccess = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "certificate",
		Name:      "query_success",
		Help:      "Whether certificate expiry query succeeded for a certificate binding (1=success, 0=failure)",
	}, certificateLabelNames)

	// CertificateBindingsTotal is the total number of certificate bindings in the current check round.
	CertificateBindingsTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "certificate",
		Name:      "bindings_total",
		Help:      "Total number of SSL certificate bindings in the current check round",
	})
)

func init() {
	metrics.Registry.MustRegister(CertificateDaysUntilExpiry)
	metrics.Registry.MustRegister(CertificateQuerySuccess)
	metrics.Registry.MustRegister(CertificateBindingsTotal)
}

// SetCertificateDaysUntilExpiry sets days_until_expiry for a certificate binding.
func SetCertificateDaysUntilExpiry(ownerKind, ownerNS, ownerName, certID, certRole, certScope, protocol, port,
	domain string, days float64) {
	CertificateDaysUntilExpiry.WithLabelValues(
		ownerKind, ownerNS, ownerName, certID, certRole, certScope, protocol, port, domain).Set(days)
}

// SetCertificateQuerySuccess sets query_success for a certificate binding.
func SetCertificateQuerySuccess(ownerKind, ownerNS, ownerName, certID, certRole, certScope, protocol, port,
	domain string, success float64) {
	CertificateQuerySuccess.WithLabelValues(
		ownerKind, ownerNS, ownerName, certID, certRole, certScope, protocol, port, domain).Set(success)
}

// SetCertificateBindingsTotal sets the total binding count gauge.
func SetCertificateBindingsTotal(total float64) {
	CertificateBindingsTotal.Set(total)
}

// DeleteCertificateDaysUntilExpiry removes days_until_expiry series for a binding label set.
func DeleteCertificateDaysUntilExpiry(ownerKind, ownerNS, ownerName, certID, certRole, certScope, protocol, port,
	domain string) {
	CertificateDaysUntilExpiry.DeleteLabelValues(
		ownerKind, ownerNS, ownerName, certID, certRole, certScope, protocol, port, domain)
}

// DeleteCertificateMetrics removes both days_until_expiry and query_success series for a binding.
func DeleteCertificateMetrics(ownerKind, ownerNS, ownerName, certID, certRole, certScope, protocol, port,
	domain string) {
	DeleteCertificateDaysUntilExpiry(ownerKind, ownerNS, ownerName, certID, certRole, certScope, protocol, port, domain)
	CertificateQuerySuccess.DeleteLabelValues(
		ownerKind, ownerNS, ownerName, certID, certRole, certScope, protocol, port, domain)
}
