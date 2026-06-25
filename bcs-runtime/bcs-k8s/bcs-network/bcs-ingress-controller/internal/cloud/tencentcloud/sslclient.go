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

// Package tencentcloud SSL client uses DescribeCertificate (singular) per certificate ID.
// The project keeps tencentcloud-sdk-go v1.0.132 monolithic package whose
// DescribeCertificatesRequest lacks CertIds; upgrading to submodule >= v1.0.1090
// would require CLB/VPC/CVM linkage and risks omitnil serialization regressions.
// SSLClient.DescribeCertificates is only an upper-layer batch semantic wrapper.
package tencentcloud

import (
	"fmt"
	"os"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	tcommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tprofile "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
)

const (
	sslEndpoint       = "ssl.tencentcloudapi.com"
	sslMaxPageSize    = 1000
	sslDescribeRetry  = 3
	certTimeLayout    = "2006-01-02 15:04:05"
	certTypeCA        = "CA"
)

var certLocation = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*3600)
	}
	return loc
}()

// SSLClient queries Tencent Cloud SSL certificate expiry information.
type SSLClient interface {
	DescribeCertificates(certIDs []string) (map[string]int64, error)
}

type sslAPI interface {
	DescribeCertificate(request *tssl.DescribeCertificateRequest) (*tssl.DescribeCertificateResponse, error)
}

type sslClientImpl struct {
	api sslAPI
}

// NewSSLClient creates an SSL client using controller global credentials from environment variables.
func NewSSLClient() (SSLClient, error) {
	secretID := os.Getenv(EnvNameTencentCloudAccessKeyID)
	secretKey := os.Getenv(EnvNameTencentCloudAccessKey)
	if secretID == "" || secretKey == "" {
		return nil, fmt.Errorf("missing %s or %s environment variable",
			EnvNameTencentCloudAccessKeyID, EnvNameTencentCloudAccessKey)
	}
	return NewSSLClientWithSecretIDKey(secretID, secretKey)
}

// resolveSSLEndpoint returns SSL API endpoint from TENCENTCLOUD_SSL_DOMAIN or the default.
func resolveSSLEndpoint() string {
	if domain := os.Getenv(EnvNameTencentCloudSslDomain); domain != "" {
		return domain
	}
	return sslEndpoint
}

// NewSSLClientWithSecretIDKey creates an SSL client with explicit credentials.
func NewSSLClientWithSecretIDKey(id, key string) (SSLClient, error) {
	credential := tcommon.NewCredential(id, key)
	cpf := tprofile.NewClientProfile()
	cpf.HttpProfile.Endpoint = resolveSSLEndpoint()
	client, err := tssl.NewClient(credential, "", cpf)
	if err != nil {
		return nil, fmt.Errorf("create ssl client failed, err %s", err.Error())
	}
	return &sslClientImpl{api: client}, nil
}

// NewSSLClientWithSecret creates an SSL client from per-namespace secret data.
func NewSSLClientWithSecret(data map[string][]byte) (SSLClient, error) {
	secretIDBytes, ok := data[EnvNameTencentCloudAccessKeyID]
	if !ok {
		return nil, fmt.Errorf("lost %s in secret", EnvNameTencentCloudAccessKeyID)
	}
	secretKeyBytes, ok := data[EnvNameTencentCloudAccessKey]
	if !ok {
		return nil, fmt.Errorf("lost %s in secret", EnvNameTencentCloudAccessKey)
	}
	return NewSSLClientWithSecretIDKey(string(secretIDBytes), string(secretKeyBytes))
}

// DescribeCertificates batch-queries certificate expiry timestamps keyed by certificate ID.
func (c *sslClientImpl) DescribeCertificates(certIDs []string) (map[string]int64, error) {
	result := make(map[string]int64)
	if len(certIDs) == 0 {
		return result, nil
	}
	blog.V(3).Infof("DescribeCertificates request: %d certIDs", len(certIDs))
	for start := 0; start < len(certIDs); start += sslMaxPageSize {
		end := start + sslMaxPageSize
		if end > len(certIDs) {
			end = len(certIDs)
		}
		page := certIDs[start:end]
		blog.V(3).Infof("DescribeCertificates (%d,%d)/%d", start, end-1, len(certIDs))
		pageResult, err := c.describeCertificatesPage(page)
		if err != nil {
			return nil, err
		}
		for id, ts := range pageResult {
			result[id] = ts
		}
	}
	return result, nil
}

func (c *sslClientImpl) describeCertificatesPage(certIDs []string) (map[string]int64, error) {
	var lastErr error
	for attempt := 1; attempt <= sslDescribeRetry; attempt++ {
		blog.V(3).Infof("DescribeCertificates try %d/%d", attempt, sslDescribeRetry)
		pageResult, err := c.doDescribeCertificates(certIDs)
		if err == nil {
			return pageResult, nil
		}
		lastErr = err
	}
	blog.Errorf("DescribeCertificates out of maxRetry %d, err %s", sslDescribeRetry, lastErr.Error())
	return nil, fmt.Errorf("DescribeCertificates out of maxRetry %d, err %s", sslDescribeRetry, lastErr.Error())
}

func (c *sslClientImpl) doDescribeCertificates(certIDs []string) (map[string]int64, error) {
	result := make(map[string]int64, len(certIDs))
	for _, certID := range certIDs {
		trySharedThrottle()
		startTime := time.Now()
		req := tssl.NewDescribeCertificateRequest()
		req.CertificateId = tcommon.StringPtr(certID)
		blog.V(3).Infof("DescribeCertificate request: %s", req.ToJsonString())
		resp, err := c.api.DescribeCertificate(req)
		if err != nil {
			metrics.ReportLibRequestMetric(
				SystemNameInMetricTencentCloud,
				HandlerNameInMetricTencentCloudSDK,
				"DescribeCertificate", metrics.LibCallStatusErr, startTime)
			blog.Errorf("DescribeCertificate failed, err %s", err.Error())
			return nil, err
		}
		metrics.ReportLibRequestMetric(
			SystemNameInMetricTencentCloud,
			HandlerNameInMetricTencentCloudSDK,
			"DescribeCertificate", metrics.LibCallStatusOK, startTime)
		if resp == nil || resp.Response == nil {
			blog.V(3).Infof("DescribeCertificate response: empty for certID %s", certID)
			continue
		}
		blog.V(3).Infof("DescribeCertificate response: %s", resp.ToJsonString())
		endUnix, ok := parseCertExpiryUnix(resp.Response.CertificateType, resp.Response.CertEndTime, nil)
		if !ok {
			blog.Infof("DescribeCertificate certID %s has no valid expiry time, skip", certID)
			continue
		}
		result[certID] = endUnix
	}
	return result, nil
}

// certExpiryInput holds fields used to resolve certificate expiry time.
type certExpiryInput struct {
	certificateType *string
	certEndTime     *string
	caEndTimes      []string
}

func parseCertExpiryUnix(certType, certEndTime *string, caEndTimes []string) (int64, bool) {
	info := certExpiryInput{
		certificateType: certType,
		certEndTime:     certEndTime,
		caEndTimes:      caEndTimes,
	}
	return parseCertExpiryFromInput(info)
}

func parseCertExpiryFromInput(info certExpiryInput) (int64, bool) {
	if info.certEndTime != nil && *info.certEndTime != "" {
		return parseCertEndTimeString(*info.certEndTime)
	}
	if info.certificateType != nil && *info.certificateType == certTypeCA && len(info.caEndTimes) > 0 {
		return earliestCAEndTime(info.caEndTimes)
	}
	return 0, false
}

func parseCertEndTimeString(value string) (int64, bool) {
	t, err := time.ParseInLocation(certTimeLayout, value, certLocation)
	if err != nil {
		blog.Errorf("parse CertEndTime %q failed, err %s", value, err.Error())
		return 0, false
	}
	return t.Unix(), true
}

func earliestCAEndTime(values []string) (int64, bool) {
	var earliest int64
	found := false
	for _, v := range values {
		if v == "" {
			continue
		}
		ts, ok := parseCertEndTimeString(v)
		if !ok {
			continue
		}
		if !found || ts < earliest {
			earliest = ts
			found = true
		}
	}
	return earliest, found
}
