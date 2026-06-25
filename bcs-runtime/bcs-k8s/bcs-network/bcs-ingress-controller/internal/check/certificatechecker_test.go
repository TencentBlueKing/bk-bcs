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

package check

import (
	"context"
	"fmt"
	"testing"
	"time"

	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud/namespacedssl"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud/tencentcloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/option"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

type fakeSSLClient struct {
	result map[string]int64
	err    error
	calls  int
}

func (f *fakeSSLClient) DescribeCertificates(certIDs []string) (map[string]int64, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	out := make(map[string]int64, len(certIDs))
	for _, id := range certIDs {
		if ts, ok := f.result[id]; ok {
			out[id] = ts
		}
	}
	return out, nil
}

func newCertCheckScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = networkextensionv1.AddToScheme(s)
	_ = k8scorev1.AddToScheme(s)
	return s
}

func makeNSSecret(ns string) *k8scorev1.Secret {
	return &k8scorev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            namespacedssl.IDKeySecretName,
			Namespace:       ns,
			ResourceVersion: "1",
		},
		Data: map[string][]byte{
			tencentcloud.EnvNameTencentCloudAccessKeyID: []byte("id"),
			tencentcloud.EnvNameTencentCloudAccessKey:   []byte("key"),
		},
	}
}

func makeHTTPSIngress(ns, name, certID string, mutual bool) *networkextensionv1.Ingress {
	cert := &networkextensionv1.IngressListenerCertificate{CertID: certID}
	if mutual {
		cert.Mode = certModeMutual
		cert.CertCaID = certID + "-ca"
	}
	return &networkextensionv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name},
		Spec: networkextensionv1.IngressSpec{
			Rules: []networkextensionv1.IngressRule{{
				Port: 443, Protocol: constant.ProtocolHTTPS, Certificate: cert,
			}},
		},
	}
}

func TestCertificateChecker(t *testing.T) {
	future := time.Now().Add(30 * 24 * time.Hour).Unix()
	past := time.Now().Add(-3 * 24 * time.Hour).Unix()
	ssl := &fakeSSLClient{result: map[string]int64{"cert-1": future, "cert-expired": past}}

	scheme := newCertCheckScheme()
	cli := k8sfake.NewFakeClientWithScheme(scheme,
		makeHTTPSIngress("default", "ing-1", "cert-1", false),
		makeHTTPSIngress("default", "ing-mutual", "cert-m", true),
	)
	opts := &option.ControllerOption{Cloud: constant.CloudTencent}
	checker := NewCertificateChecker(cli, ssl, nil, opts)
	checker.Run()

	if got := testutil.ToFloat64(metrics.CertificateBindingsTotal); got != 3 {
		t.Fatalf("bindings_total = %v, want 3", got)
	}
	if got := testutil.ToFloat64(metrics.CertificateQuerySuccess.WithLabelValues(
		constant.KindIngress, "default", "ing-1", "cert-1", CertRoleServer, CertScopeRule, constant.ProtocolHTTPS, "443", "")); got != 1 {
		t.Fatalf("query_success for ing-1 = %v, want 1", got)
	}
	if got := testutil.ToFloat64(metrics.CertificateDaysUntilExpiry.WithLabelValues(
		constant.KindIngress, "default", "ing-1", "cert-1", CertRoleServer, CertScopeRule, constant.ProtocolHTTPS, "443", "")); got <= 0 {
		t.Fatalf("expected positive days_until_expiry, got %v", got)
	}
	if got := testutil.ToFloat64(metrics.CertificateQuerySuccess.WithLabelValues(
		constant.KindIngress, "default", "ing-mutual", "cert-m-ca", CertRoleClientCA, CertScopeRule, constant.ProtocolHTTPS, "443", "")); got != 0 {
		t.Fatalf("missing cert should have query_success=0, got %v", got)
	}
}

func TestCertificateCheckerExpiredCert(t *testing.T) {
	past := time.Now().Add(-3 * 24 * time.Hour).Unix()
	ssl := &fakeSSLClient{result: map[string]int64{"cert-old": past}}
	scheme := newCertCheckScheme()
	cli := k8sfake.NewFakeClientWithScheme(scheme, makeHTTPSIngress("default", "ing-old", "cert-old", false))
	checker := NewCertificateChecker(cli, ssl, nil, &option.ControllerOption{})
	checker.Run()

	if got := testutil.ToFloat64(metrics.CertificateDaysUntilExpiry.WithLabelValues(
		constant.KindIngress, "default", "ing-old", "cert-old", CertRoleServer, CertScopeRule, constant.ProtocolHTTPS, "443", "")); got >= 0 {
		t.Fatalf("expected negative days_until_expiry, got %v", got)
	}
}

func TestCertificateCheckerAPIFailure(t *testing.T) {
	ssl := &fakeSSLClient{err: fmt.Errorf("api down")}
	scheme := newCertCheckScheme()
	cli := k8sfake.NewFakeClientWithScheme(scheme, makeHTTPSIngress("default", "ing-fail", "cert-f", false))
	checker := NewCertificateChecker(cli, ssl, nil, &option.ControllerOption{})
	checker.Run()

	if got := testutil.ToFloat64(metrics.CertificateDaysUntilExpiry.WithLabelValues(
		constant.KindIngress, "default", "ing-fail", "cert-f", CertRoleServer, CertScopeRule, constant.ProtocolHTTPS, "443", "")); got != 0 {
		t.Fatalf("days_until_expiry should be deleted on failure, got %v", got)
	}
	if got := testutil.ToFloat64(metrics.CertificateQuerySuccess.WithLabelValues(
		constant.KindIngress, "default", "ing-fail", "cert-f", CertRoleServer, CertScopeRule, constant.ProtocolHTTPS, "443", "")); got != 0 {
		t.Fatalf("query_success should be 0 on failure, got %v", got)
	}
}

func TestCertCheckerCleanupStale(t *testing.T) {
	future := time.Now().Add(10 * 24 * time.Hour).Unix()
	ssl := &fakeSSLClient{result: map[string]int64{"cert-a": future}}
	scheme := newCertCheckScheme()
	ing := makeHTTPSIngress("default", "ing-a", "cert-a", false)
	cli := k8sfake.NewFakeClientWithScheme(scheme, ing)
	checker := NewCertificateChecker(cli, ssl, nil, &option.ControllerOption{})

	checker.Run()
	if got := testutil.ToFloat64(metrics.CertificateQuerySuccess.WithLabelValues(
		constant.KindIngress, "default", "ing-a", "cert-a", CertRoleServer, CertScopeRule, constant.ProtocolHTTPS, "443", "")); got != 1 {
		t.Fatalf("expected initial query_success=1, got %v", got)
	}

	if err := cli.Delete(context.TODO(), ing); err != nil {
		t.Fatalf("delete ingress failed: %v", err)
	}
	checker.Run()
	if got := testutil.ToFloat64(metrics.CertificateQuerySuccess.WithLabelValues(
		constant.KindIngress, "default", "ing-a", "cert-a", CertRoleServer, CertScopeRule, constant.ProtocolHTTPS, "443", "")); got != 0 {
		t.Fatalf("stale query_success should be cleaned, got %v", got)
	}
	if got := testutil.ToFloat64(metrics.CertificateBindingsTotal); got != 0 {
		t.Fatalf("bindings_total should be 0, got %v", got)
	}
}

func TestCertificateCheckerListFailure(t *testing.T) {
	future := time.Now().Add(10 * 24 * time.Hour).Unix()
	ssl := &fakeSSLClient{result: map[string]int64{"cert-x": future}}
	scheme := newCertCheckScheme()
	cli := k8sfake.NewFakeClientWithScheme(scheme, makeHTTPSIngress("default", "ing-x", "cert-x", false))
	checker := NewCertificateChecker(cli, ssl, nil, &option.ControllerOption{})

	checker.Run()
	before := testutil.ToFloat64(metrics.CertificateBindingsTotal)

	broken := k8sfake.NewFakeClientWithScheme(runtime.NewScheme())
	checker.cli = broken
	checker.Run()
	if got := testutil.ToFloat64(metrics.CertificateBindingsTotal); got != before {
		t.Fatalf("list failure should not update metrics, before=%v got=%v", before, got)
	}
}

func TestCertCheckerNamespaceScope(t *testing.T) {
	future := time.Now().Add(20 * 24 * time.Hour).Unix()
	defaultSSL := &fakeSSLClient{result: map[string]int64{"cert-global": future}, calls: 0}
	nsSSL := &fakeSSLClient{result: map[string]int64{"cert-ns": future}, calls: 0}

	scheme := newCertCheckScheme()
	cli := k8sfake.NewFakeClientWithScheme(scheme,
		makeHTTPSIngress("bcs-system", "ing-exempt", "cert-global", false),
		makeHTTPSIngress("user-ns", "ing-user", "cert-ns", false),
		makeNSSecret("user-ns"),
	)

	nsRouter := namespacedssl.NewNamespacedSSLForTest(cli, func(data map[string][]byte) (tencentcloud.SSLClient, error) {
		return nsSSL, nil
	}, defaultSSL, map[string]struct{}{"bcs-system": {}})

	opts := &option.ControllerOption{IsNamespaceScope: true}
	checker := NewCertificateChecker(cli, defaultSSL, nsRouter, opts)
	checker.Run()

	if got := testutil.ToFloat64(metrics.CertificateQuerySuccess.WithLabelValues(
		constant.KindIngress, "bcs-system", "ing-exempt", "cert-global", CertRoleServer, CertScopeRule, constant.ProtocolHTTPS, "443", "")); got != 1 {
		t.Fatalf("exempt ns query_success = %v, want 1", got)
	}
	if nsSSL.calls == 0 && defaultSSL.calls == 0 {
		t.Fatalf("expected ssl client calls in namespace scope mode")
	}
}

func TestShouldRegisterCertChecker(t *testing.T) {
	if !ShouldRegisterCertificateChecker(constant.CloudTencent, true) {
		t.Fatal("tencentcloud with enabled=true should register certificate checker")
	}
	if ShouldRegisterCertificateChecker(constant.CloudTencent, false) {
		t.Fatal("tencentcloud with enabled=false should not register certificate checker")
	}
	if ShouldRegisterCertificateChecker(constant.CloudAWS, true) {
		t.Fatal("aws should not register certificate checker even when enabled=true")
	}
	if ShouldRegisterCertificateChecker(constant.CloudAWS, false) {
		t.Fatal("aws should not register certificate checker when enabled=false")
	}
}

func TestCertCheckerRegInterval(t *testing.T) {
	if !ShouldRegisterCertificateChecker(constant.CloudTencent, true) {
		t.Fatal("tencentcloud with enabled=true should register certificate checker")
	}
	if CertificateCheckerRegisterInterval != CheckPer60Min {
		t.Fatalf("CertificateCheckerRegisterInterval = %v, want CheckPer60Min", CertificateCheckerRegisterInterval)
	}
	if CertificateCheckerRegisterInterval == CheckPer10Min {
		t.Fatal("certificate checker must not register with CheckPer10Min")
	}
}
