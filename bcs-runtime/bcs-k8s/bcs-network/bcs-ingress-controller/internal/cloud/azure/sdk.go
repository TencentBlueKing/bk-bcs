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
 *
 */

package azure

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/throttle"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
	"github.com/pkg/errors"
)

const (
	envNameAzureClientID            = "AZURE_CLIENT_ID"
	envNameAzureClientSecret        = "AZURE_CLIENT_SECRET"
	envNameAzureTenantID            = "AZURE_TENANT_ID"
	envNameAzureSubscriptionID      = "AZURE_SUBSCRIPTION_ID"
	envNameAzureResourceGroup       = "AZURE_RESOURCE_GROUP"
	envNameAzureVNetID              = "AZURE_VNET_NAME"
	envNameAzureVNetResourceGroup   = "AZURE_VNET_RESOURCE_GROUP"
	envNameAzureRateLimitQPS        = "AZURE_RATE_LIMIT_QPS"
	envNameAzureRateLimitBucketSize = "AZURE_RATE_LIMIT_BUCKET_SIZE"
)

var (
	// If the delay caused by the frequency limit exceeds this value, it is recorded in the log
	maxLatency = 120 * time.Millisecond
	// qps for rate limit
	defaultThrottleQPS = 50
	// bucket size for rate limit
	defaultBucketSize = 50
)

// SdkWrapper sdk wrapper for azure
type SdkWrapper struct {
	ctx context.Context

	clientID              string
	clientSecret          string
	tenantID              string
	subscriptionID        string
	resourceGroupName     string
	vNetName              string
	vNetResourceGroupName string

	agFrontIPName string
	lbFrontIPName string

	credential    *azidentity.DefaultAzureCredential
	lbCli         *armnetwork.LoadBalancersClient
	lbAddrPoolCli *armnetwork.LoadBalancerBackendAddressPoolsClient
	appGatewayCli *armnetwork.ApplicationGatewaysClient

	ratelimitqps        int64
	ratelimitbucketSize int64
	// rate limiter for calling sdk
	throttler throttle.RateLimiter
}

// NewSdkWrapper create sdk wrapper
func NewSdkWrapper() (*SdkWrapper, error) {
	sw := &SdkWrapper{}
	sw.loadEnv()

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, errors.Wrapf(err, "create azure cred failed")
	}

	sw.credential = cred
	sw.lbCli, err = armnetwork.NewLoadBalancersClient(sw.subscriptionID, sw.credential, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "create azure lb client failed")
	}
	sw.lbAddrPoolCli, err = armnetwork.NewLoadBalancerBackendAddressPoolsClient(sw.subscriptionID,
		sw.credential, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "create azure lb address pool client failed")
	}

	sw.appGatewayCli, err = armnetwork.NewApplicationGatewaysClient(sw.subscriptionID, sw.credential, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "create azure application gateway client failed")
	}
	sw.throttler = throttle.NewTokenBucket(sw.ratelimitqps, sw.ratelimitbucketSize)
	return sw, nil
}

// NewSdkWrapperWithSecretIDKey create a new aws sdk wrapper with secret id and key
func NewSdkWrapperWithSecretIDKey(secretID, secretKey string) (*SdkWrapper, error) {
	sw, err := NewSdkWrapper()
	if err != nil {
		return nil, err
	}
	sw.clientID = secretID
	sw.clientSecret = secretKey
	return sw, nil
}

// loadEnv load env parameters todo
func (sw *SdkWrapper) loadEnv() {
	sw.clientID = os.Getenv(envNameAzureClientID)
	sw.clientSecret = os.Getenv(envNameAzureClientSecret)
	sw.tenantID = os.Getenv(envNameAzureTenantID)
	sw.subscriptionID = os.Getenv(envNameAzureSubscriptionID)
	sw.resourceGroupName = os.Getenv(envNameAzureResourceGroup)
	sw.vNetName = os.Getenv(envNameAzureVNetID)
	sw.vNetResourceGroupName = os.Getenv(envNameAzureVNetResourceGroup)

	// if not set, use resourceGroupName
	if sw.vNetResourceGroupName == "" {
		sw.vNetResourceGroupName = sw.resourceGroupName
	}

	qpsStr := os.Getenv(envNameAzureRateLimitQPS)
	if len(qpsStr) != 0 {
		qps, err := strconv.ParseInt(qpsStr, 10, 64)
		if err != nil {
			blog.Warnf("parse rate limit qps %s failed, err %s, use default %d",
				qpsStr, err.Error(), defaultThrottleQPS)
			sw.ratelimitqps = int64(defaultThrottleQPS)
		} else {
			sw.ratelimitqps = qps
		}
	} else {
		sw.ratelimitqps = int64(defaultThrottleQPS)
	}

	bucketSizeStr := os.Getenv(envNameAzureRateLimitBucketSize)
	if len(bucketSizeStr) != 0 {
		bucketSize, err := strconv.ParseInt(bucketSizeStr, 10, 64)
		if err != nil {
			blog.Warnf("parse rate limit bucket size %s failed, err %s, use default %d",
				bucketSizeStr, err.Error(), defaultBucketSize)
			sw.ratelimitbucketSize = int64(defaultBucketSize)
		} else {
			sw.ratelimitbucketSize = bucketSize
		}
	} else {
		sw.ratelimitbucketSize = int64(defaultBucketSize)
	}
}

// call tryThrottle before each api call
func (sw *SdkWrapper) tryThrottle() {
	now := time.Now()
	sw.throttler.Accept()
	if latency := time.Since(now); latency > maxLatency {
		pc, _, _, _ := runtime.Caller(2)
		callerName := runtime.FuncForPC(pc).Name()
		blog.Infof("Throttling request took %d ms, function: %s", latency, callerName)
	}
}

// GetLoadBalancer get azure load balancer
func (sw *SdkWrapper) GetLoadBalancer(region string, loadBalancerName string) (*armnetwork.
	LoadBalancersClientGetResponse, error) {
	blog.V(3).Infof("GetLoadBalancer(region=%s,lbName=%s)", region, loadBalancerName)

	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(SystemNameInMetricAzure, HandlerNameInMetricAzureSDK,
			"GetLoadBalancer", ret, startTime)
	}
	sw.tryThrottle()
	lbResp, err := sw.lbCli.Get(context.TODO(), sw.resourceGroupName, loadBalancerName, nil)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, err
		}
		mf(metrics.LibCallStatusErr) // or time out
		return nil, errors.Wrapf(err, "GetLoadBalancer(region=%s,lbName=%s) failed", region, loadBalancerName)
	}

	blog.V(3).Infof("GetLoadBalancer(region=%s,lbName=%s) response: %s", region, loadBalancerName,
		common.ToJsonString(lbResp))
	mf(metrics.LibCallStatusOK)

	return &lbResp, nil
}

// GetApplicationGateway get azure application gateway
func (sw *SdkWrapper) GetApplicationGateway(region string, appGatewayName string) (*armnetwork.
	ApplicationGatewaysClientGetResponse, error) {
	blog.V(3).Infof("GetApplicationGateway(region=%s,appGatewayName=%s)", region, appGatewayName)

	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(SystemNameInMetricAzure, HandlerNameInMetricAzureSDK,
			"GetApplicationGateway", ret, startTime)
	}
	sw.tryThrottle()
	appGatewayRsp, err := sw.appGatewayCli.Get(context.TODO(), sw.resourceGroupName, appGatewayName, nil)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, err
		}
		mf(metrics.LibCallStatusErr) // or time out
		return nil, errors.Wrapf(err, "GetApplicationGateway(region=%s,appGatewayName=%s) failed", region,
			appGatewayName)
	}

	blog.V(3).Infof("GetApplicationGateway(region=%s,appGatewayName=%s) response: %s", region, appGatewayName,
		common.ToJsonString(appGatewayRsp))
	mf(metrics.LibCallStatusOK)

	return &appGatewayRsp, nil
}

// CreateOrUpdateApplicationGateway create or update azure application gateway
func (sw *SdkWrapper) CreateOrUpdateApplicationGateway(appGatewayName string,
	parameters armnetwork.ApplicationGateway) (*armnetwork.ApplicationGatewaysClientCreateOrUpdateResponse, error) {
	blog.V(3).Infof("CreateOrUpdateApplicationGateway[%s] request: %s", appGatewayName,
		common.ToJsonString(parameters))

	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(SystemNameInMetricAzure, HandlerNameInMetricAzureSDK,
			"CreateOrUpdateApplicationGateway", ret, startTime)
	}
	sw.tryThrottle()
	pollerResp, err := sw.appGatewayCli.BeginCreateOrUpdate(context.TODO(), sw.resourceGroupName, appGatewayName, parameters,
		nil)
	if err != nil || pollerResp == nil {
		mf(metrics.LibCallStatusErr) // or time out
		return nil, errors.Wrapf(err, "CreateOrUpdateApplicationGateway failed")
	}

	appGatewayRsp, err := pollerResp.PollUntilDone(context.TODO(), nil)
	if err != nil {
		mf(metrics.LibCallStatusErr) // or time out
		return nil, errors.Wrapf(err, "CreateOrUpdateApplicationGateway failed")
	}

	mf(metrics.LibCallStatusOK)
	blog.V(3).Infof("CreateOrUpdateApplicationGateway response: %s", common.ToJsonString(appGatewayRsp))

	return &appGatewayRsp, nil
}

// CreateOrUpdateLoadBalancer create or update azure load balancer
func (sw *SdkWrapper) CreateOrUpdateLoadBalancer(loadBalancerName string,
	parameters armnetwork.LoadBalancer) (*armnetwork.LoadBalancersClientCreateOrUpdateResponse, error) {
	blog.V(3).Infof("CreateOrUpdateBalancer[%s] request: %s", loadBalancerName, common.ToJsonString(parameters))

	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(SystemNameInMetricAzure, HandlerNameInMetricAzureSDK,
			"CreateOrUpdateBalancer", ret, startTime)
	}
	sw.tryThrottle()
	pollerResp, err := sw.lbCli.BeginCreateOrUpdate(context.TODO(), sw.resourceGroupName, loadBalancerName, parameters,
		nil)
	if err != nil {
		mf(metrics.LibCallStatusErr) // or time out
		return nil, errors.Wrapf(err, "CreateOrUpdateBalancer failed")
	}

	lb, err := pollerResp.PollUntilDone(context.TODO(), nil)
	if err != nil {
		mf(metrics.LibCallStatusErr) // or time out
		return nil, errors.Wrapf(err, "CreateOrUpdateBalancer failed")
	}

	mf(metrics.LibCallStatusOK)
	blog.V(3).Infof("CreateOrUpdateBalancer response: %s", common.ToJsonString(lb))

	return &lb, nil
}

// CreateOrUpdateLoadBalanceBackendAddressPool create or update azure backend address pool
func (sw *SdkWrapper) CreateOrUpdateLoadBalanceBackendAddressPool(loadBalancerName string,
	backendAddressPoolName string, parameters armnetwork.BackendAddressPool) (*armnetwork.
	LoadBalancerBackendAddressPoolsClientCreateOrUpdateResponse, error) {
	blog.V(3).Infof("createOrUpdateBackendAddressPool[%s] request: %s", loadBalancerName,
		common.ToJsonString(parameters))

	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(SystemNameInMetricAzure, HandlerNameInMetricAzureSDK,
			"createOrUpdateBackendAddressPool", ret, startTime)
	}
	sw.tryThrottle()
	pollerResp, err := sw.lbAddrPoolCli.BeginCreateOrUpdate(context.TODO(), sw.resourceGroupName,
		loadBalancerName, backendAddressPoolName, parameters, nil)
	if err != nil {
		mf(metrics.LibCallStatusErr) // or time out
		return nil, errors.Wrapf(err, "createOrUpdateBackendAddressPool failed")
	}

	lb, err := pollerResp.PollUntilDone(context.TODO(), nil)
	if err != nil {
		mf(metrics.LibCallStatusErr) // or time out
		return nil, errors.Wrapf(err, "createOrUpdateBackendAddressPool failed")
	}

	mf(metrics.LibCallStatusOK)
	blog.V(3).Infof("createOrUpdateBackendAddressPool response: %s", common.ToJsonString(lb))

	return &lb, nil
}

// DeleteLoadBalanceAddressPool delete azure lb adddress pool
func (sw *SdkWrapper) DeleteLoadBalanceAddressPool(loadBalancerName string, poolName string) error {
	blog.V(3).Infof("deleteLoadBalanceAddressPool[%s] request: %s", loadBalancerName, poolName)

	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(SystemNameInMetricAzure, HandlerNameInMetricAzureSDK,
			"deleteLoadBalanceAddressPool", ret, startTime)
	}
	sw.tryThrottle()
	pollerResp, err := sw.lbAddrPoolCli.BeginDelete(context.TODO(), sw.resourceGroupName,
		loadBalancerName, poolName, nil)
	if err != nil {
		mf(metrics.LibCallStatusErr) // or time out
		return errors.Wrapf(err, "deleteLoadBalanceAddressPool failed")
	}

	lb, err := pollerResp.PollUntilDone(context.TODO(), nil)
	if err != nil {
		mf(metrics.LibCallStatusErr) // or time out
		return errors.Wrapf(err, "deleteLoadBalanceAddressPool failed")
	}

	mf(metrics.LibCallStatusOK)
	blog.V(3).Infof("deleteLoadBalanceAddressPool response: %s", common.ToJsonString(lb))

	return nil
}

// GetLoadBalanceBackendAddressPool get azure lb address pool
func (sw *SdkWrapper) GetLoadBalanceBackendAddressPool(loadBalancerName,
	addrPoolName string) (*armnetwork.LoadBalancerBackendAddressPoolsClientGetResponse, error) {

	blog.V(3).Infof("GetLoadBalanceBackendAddressPool(loadBalancer:%s, backendAddressPool:%s)", loadBalancerName,
		addrPoolName)

	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(SystemNameInMetricAzure, HandlerNameInMetricAzureSDK,
			"GetLoadBalanceBackendAddressPool", ret, startTime)
	}
	sw.tryThrottle()
	backendAddressPool, err := sw.lbAddrPoolCli.Get(context.TODO(), sw.resourceGroupName, loadBalancerName,
		addrPoolName, nil)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, err
		}
		mf(metrics.LibCallStatusErr) // or time out
		return nil, errors.Wrapf(err, "GetLoadBalanceBackendAddressPool(loadBalancer:%s, "+
			"backendAddressPool:%s) failed", loadBalancerName, addrPoolName)
	}

	mf(metrics.LibCallStatusOK)
	blog.V(3).Infof("GetLoadBalanceBackendAddressPool(loadBalancer:%s, backendAddressPool:%s) response: %s",
		loadBalancerName, *backendAddressPool.Name, common.ToJsonString(backendAddressPool))
	return &backendAddressPool, nil
}

func (sw *SdkWrapper) buildVNetID() string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/%s",
		sw.subscriptionID, sw.vNetResourceGroupName, sw.vNetName)
}
