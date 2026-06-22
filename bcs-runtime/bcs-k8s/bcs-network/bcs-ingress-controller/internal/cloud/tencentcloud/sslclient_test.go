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

package tencentcloud

import (
	"fmt"
	"os"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	tcommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tprofile "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
)

type mockSSLAPI struct {
	calls       int32
	failUntil   int32
	certEndTime map[string]string
	certType    map[string]string
}

func (m *mockSSLAPI) DescribeCertificate(req *tssl.DescribeCertificateRequest) (
	*tssl.DescribeCertificateResponse, error) {
	call := atomic.AddInt32(&m.calls, 1)
	if req == nil || req.CertificateId == nil {
		return nil, fmt.Errorf("missing certificate id")
	}
	certID := *req.CertificateId
	if call <= m.failUntil {
		return nil, fmt.Errorf("mock api error")
	}
	endTime := m.certEndTime[certID]
	certType := m.certType[certID]
	resp := tssl.NewDescribeCertificateResponse()
	jsonStr := fmt.Sprintf(`{"Response":{"CertificateId":"%s","CertEndTime":"%s","CertificateType":"%s"}}`,
		certID, endTime, certType)
	if err := resp.FromJsonString(jsonStr); err != nil {
		return nil, err
	}
	return resp, nil
}

func TestDescribeCertificates(t *testing.T) {
	future := time.Now().In(certLocation).Add(30 * 24 * time.Hour).Format(certTimeLayout)
	mock := &mockSSLAPI{
		certEndTime: map[string]string{"cert-1": future},
		certType:    map[string]string{"cert-1": "SVR"},
	}
	client := &sslClientImpl{api: mock}
	result, err := client.DescribeCertificates([]string{"cert-1"})
	if err != nil {
		t.Fatalf("DescribeCertificates failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if _, ok := result["cert-1"]; !ok {
		t.Fatalf("missing cert-1 in result")
	}
}

func TestDescribeCertificatesPagination(t *testing.T) {
	future := time.Now().In(certLocation).Add(24 * time.Hour).Format(certTimeLayout)
	mock := &mockSSLAPI{
		certEndTime: make(map[string]string),
		certType:    make(map[string]string),
	}
	ids := make([]string, 0, 1001)
	for i := 0; i < 1001; i++ {
		id := fmt.Sprintf("cert-%d", i)
		ids = append(ids, id)
		mock.certEndTime[id] = future
		mock.certType[id] = "SVR"
	}
	client := &sslClientImpl{api: mock}
	result, err := client.DescribeCertificates(ids)
	if err != nil {
		t.Fatalf("DescribeCertificates pagination failed: %v", err)
	}
	if len(result) != 1001 {
		t.Fatalf("expected 1001 results, got %d", len(result))
	}
	if atomic.LoadInt32(&mock.calls) != 1001 {
		t.Fatalf("expected 1001 api calls, got %d", mock.calls)
	}
}

func TestDescribeCertificatesRetry(t *testing.T) {
	future := time.Now().In(certLocation).Add(24 * time.Hour).Format(certTimeLayout)
	mock := &mockSSLAPI{
		failUntil:   2,
		certEndTime: map[string]string{"cert-1": future},
		certType:    map[string]string{"cert-1": "SVR"},
	}
	client := &sslClientImpl{api: mock}
	result, err := client.DescribeCertificates([]string{"cert-1"})
	if err != nil {
		t.Fatalf("DescribeCertificates retry failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result after retry, got %d", len(result))
	}
	if atomic.LoadInt32(&mock.calls) != 3 {
		t.Fatalf("expected 3 api calls, got %d", mock.calls)
	}
}

func TestDescribeCertsRetryExhaust(t *testing.T) {
	mock := &mockSSLAPI{failUntil: 10}
	client := &sslClientImpl{api: mock}
	_, err := client.DescribeCertificates([]string{"cert-1"})
	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}
	if atomic.LoadInt32(&mock.calls) != sslDescribeRetry {
		t.Fatalf("expected %d retries, got %d", sslDescribeRetry, mock.calls)
	}
}

func TestDescribeCertsThrottleAcc(t *testing.T) {
	future := time.Now().In(certLocation).Add(24 * time.Hour).Format(certTimeLayout)
	mock := &mockSSLAPI{
		certEndTime: map[string]string{"cert-1": future, "cert-2": future},
		certType:    map[string]string{"cert-1": "SVR", "cert-2": "SVR"},
	}
	client := &sslClientImpl{api: mock}

	var accepts int32
	oldThrottle := trySharedThrottle
	trySharedThrottle = func() { atomic.AddInt32(&accepts, 1) }
	defer func() { trySharedThrottle = oldThrottle }()

	_, err := client.DescribeCertificates([]string{"cert-1", "cert-2"})
	if err != nil {
		t.Fatalf("DescribeCertificates failed: %v", err)
	}
	if got := atomic.LoadInt32(&accepts); got < 2 {
		t.Fatalf("expected at least 2 throttle accepts, got %d", got)
	}
}

func TestDescribeCertsMetricMethod(t *testing.T) {
	future := time.Now().In(certLocation).Add(24 * time.Hour).Format(certTimeLayout)
	mock := &mockSSLAPI{
		certEndTime: map[string]string{"cert-metric": future},
		certType:    map[string]string{"cert-metric": "SVR"},
	}
	client := &sslClientImpl{api: mock}

	oldThrottle := trySharedThrottle
	trySharedThrottle = func() {}
	defer func() { trySharedThrottle = oldThrottle }()

	before := testutil.ToFloat64(metrics.LibRequestTotal.WithLabelValues(
		SystemNameInMetricTencentCloud,
		HandlerNameInMetricTencentCloudSDK,
		"DescribeCertificate",
		metrics.LibCallStatusOK,
	))

	_, err := client.DescribeCertificates([]string{"cert-metric"})
	if err != nil {
		t.Fatalf("DescribeCertificates failed: %v", err)
	}

	after := testutil.ToFloat64(metrics.LibRequestTotal.WithLabelValues(
		SystemNameInMetricTencentCloud,
		HandlerNameInMetricTencentCloudSDK,
		"DescribeCertificate",
		metrics.LibCallStatusOK,
	))
	if after-before != 1 {
		t.Fatalf("expected lib request counter +1, before=%v after=%v", before, after)
	}
}

func TestResolveSSLEndpoint(t *testing.T) {
	t.Setenv(EnvNameTencentCloudSslDomain, "")
	if got := resolveSSLEndpoint(); got != sslEndpoint {
		t.Fatalf("expected default %q, got %q", sslEndpoint, got)
	}

	t.Setenv(EnvNameTencentCloudSslDomain, "ssl.internal.tencentcloudapi.com")
	if got := resolveSSLEndpoint(); got != "ssl.internal.tencentcloudapi.com" {
		t.Fatalf("expected custom endpoint, got %q", got)
	}

	t.Setenv(EnvNameTencentCloudSslDomain, "")
}

func sslClientHTTPEndpoint(t *testing.T, client SSLClient) string {
	t.Helper()
	impl, ok := client.(*sslClientImpl)
	if !ok {
		t.Fatal("expected *sslClientImpl")
	}
	raw, ok := impl.api.(*tssl.Client)
	if !ok {
		t.Fatal("expected *tssl.Client")
	}
	v := reflect.ValueOf(raw).Elem().FieldByName("httpProfile")
	if !v.IsValid() || v.IsNil() {
		t.Fatal("missing httpProfile on ssl client")
	}
	endpoint := v.Elem().FieldByName("Endpoint")
	if !endpoint.IsValid() {
		t.Fatal("missing Endpoint on httpProfile")
	}
	return endpoint.String()
}

func TestSSLClientEndpoint(t *testing.T) {
	const internalDomain = "ssl.internal.tencentcloudapi.com"

	t.Run("default when env unset", func(t *testing.T) {
		os.Unsetenv(EnvNameTencentCloudSslDomain)
		client, err := NewSSLClientWithSecretIDKey("test-id", "test-key")
		if err != nil {
			t.Fatalf("NewSSLClientWithSecretIDKey failed: %v", err)
		}
		if got := sslClientHTTPEndpoint(t, client); got != sslEndpoint {
			t.Fatalf("expected endpoint %q, got %q", sslEndpoint, got)
		}
	})

	t.Run("custom when env set", func(t *testing.T) {
		t.Setenv(EnvNameTencentCloudSslDomain, internalDomain)
		client, err := NewSSLClientWithSecretIDKey("test-id", "test-key")
		if err != nil {
			t.Fatalf("NewSSLClientWithSecretIDKey failed: %v", err)
		}
		if got := sslClientHTTPEndpoint(t, client); got != internalDomain {
			t.Fatalf("expected endpoint %q, got %q", internalDomain, got)
		}
	})

	t.Run("profile endpoint matches resolveSSLEndpoint", func(t *testing.T) {
		t.Setenv(EnvNameTencentCloudSslDomain, internalDomain)
		cpf := tprofile.NewClientProfile()
		cpf.HttpProfile.Endpoint = resolveSSLEndpoint()
		if cpf.HttpProfile.Endpoint != internalDomain {
			t.Fatalf("expected profile endpoint %q, got %q", internalDomain, cpf.HttpProfile.Endpoint)
		}
	})
}

func TestParseCertExpiryFromInput(t *testing.T) {
	endTime := time.Now().In(certLocation).Add(48 * time.Hour).Format(certTimeLayout)
	ts, ok := parseCertExpiryFromInput(certExpiryInput{
		certEndTime: tcommon.StringPtr(endTime),
	})
	if !ok {
		t.Fatal("expected valid parse")
	}
	if ts <= time.Now().Unix() {
		t.Fatalf("expected future timestamp, got %d", ts)
	}

	caType := certTypeCA
	_, ok = parseCertExpiryFromInput(certExpiryInput{
		certificateType: &caType,
		caEndTimes:      []string{"invalid", "also-invalid"},
	})
	if ok {
		t.Fatal("expected invalid CAEndTimes to fail")
	}

	earliest := time.Now().In(certLocation).Add(24 * time.Hour).Format(certTimeLayout)
	later := time.Now().In(certLocation).Add(72 * time.Hour).Format(certTimeLayout)
	ts, ok = parseCertExpiryFromInput(certExpiryInput{
		certificateType: &caType,
		caEndTimes:      []string{later, earliest},
	})
	if !ok {
		t.Fatal("expected CAEndTimes fallback")
	}
	want, _ := parseCertEndTimeString(earliest)
	if ts != want {
		t.Fatalf("expected earliest CAEndTime %d, got %d", want, ts)
	}
}
