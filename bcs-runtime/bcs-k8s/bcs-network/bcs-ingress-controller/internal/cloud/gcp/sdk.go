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

package gcp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/throttle"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

// SdkWrapper wrapper for gcp sdk
type SdkWrapper struct {
	// credential for gcp
	credentials []byte

	// ratelimit
	ratelimitqps        int64
	ratelimitbucketSize int64
	// rate limiter for calling sdk
	throttler throttle.RateLimiter

	// clients
	computeService *compute.Service
}

const (
	// EnvNameGCPProject env name of gcp project
	EnvNameGCPProject = "GCP_PROJECT"
	// EnvNameGCPCredentials env name of gcp credentials
	EnvNameGCPCredentials = "GOOGLE_APPLICATION_CREDENTIALS"

	// EnvNameGCPRateLimitQPS env name for gcp api rate limit qps
	EnvNameGCPRateLimitQPS = "GCP_RATELIMIT_QPS"
	// EnvNameGCPRateLimitBucketSize env name for gcp api rate limit bucket size
	EnvNameGCPRateLimitBucketSize = "GCP_RATELIMIT_BUCKET_SIZE"
)

var (
	// If the delay caused by the frequency limit exceeds this value, it is recorded in the log
	maxLatency = 120 * time.Millisecond
	// the maximum number of retries caused by server error or API overrun
	maxRetry = 5
	// qps for rate limit
	defaultThrottleQPS = 50
	// bucket size for rate limit
	defaultBucketSize = 50
	// wait seconds when cloud api is busy
	waitPeriodLBDealing = 2
	// default wait seconds when cloud api is async
	defaultTimeout = 30 * time.Second
	// async operation polling interval
	defaultPollingInterval = 1 * time.Second

	// GlobalRegion means global region
	GlobalRegion = "global"
)

// NewSdkWrapper create a new gcp sdk wrapper
func NewSdkWrapper() (*SdkWrapper, error) {
	sw := &SdkWrapper{}
	err := sw.loadEnv()
	if err != nil {
		return nil, err
	}

	// init client
	c, err := google.DefaultClient(context.Background(), compute.CloudPlatformScope)
	if err != nil {
		return nil, err
	}
	sw.computeService, err = compute.New(c)
	if err != nil {
		return nil, err
	}

	sw.throttler = throttle.NewTokenBucket(sw.ratelimitqps, sw.ratelimitbucketSize)
	return sw, nil
}

// NewSdkWrapperWithSecretIDKey create a new gcp sdk wrapper with credential
func NewSdkWrapperWithSecretIDKey(credentials []byte) (*SdkWrapper, error) {
	sw := &SdkWrapper{}
	sw.credentials = credentials
	return NewSdkWrapper()
}

func (sw *SdkWrapper) loadEnv() error {
	if len(sw.credentials) == 0 {
		sw.credentials = []byte(os.Getenv(EnvNameGCPCredentials))
	}

	qpsStr := os.Getenv(EnvNameGCPRateLimitQPS)
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

	bucketSizeStr := os.Getenv(EnvNameGCPRateLimitBucketSize)
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
	return nil
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

// IsNotFound returns true if the error is resource not found
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	ae, ok := err.(*googleapi.Error)
	return ok && ae.Code == http.StatusNotFound
}

// Wait wait for cloud api async
func (sw *SdkWrapper) Wait(ctx context.Context, project string, op *compute.Operation) error {
	tk := time.NewTicker(defaultPollingInterval)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tk.C:
			var resp *compute.Operation
			var err error
			blog.Infof("wait for operation %s", op.Name)
			if op.Region != "" {
				regionStrs := strings.Split(op.Region, "/")
				region := regionStrs[len(regionStrs)-1]
				resp, err = sw.computeService.RegionOperations.Get(project, region, op.Name).Context(ctx).Do()
			}
			if op.Zone != "" {
				zoneStrs := strings.Split(op.Zone, "/")
				zone := zoneStrs[len(zoneStrs)-1]
				resp, err = sw.computeService.ZoneOperations.Get(project, zone, op.Name).Context(ctx).Do()
			} else {
				resp, err = sw.computeService.GlobalOperations.Get(project, op.Name).Context(ctx).Do()
			}
			if err != nil {
				blog.Errorf("wait for operation %s failed, err %s", op.Name, err.Error())
				return err
			}
			if resp == nil {
				return fmt.Errorf("operation %s not found", op.Name)
			}
			if resp.Status == "DONE" {
				if resp.Error != nil {
					e, err := resp.Error.MarshalJSON()
					if err != nil {
						return err
					}
					return fmt.Errorf("operation %s failed, error: %s", resp.Name, string(e))
				}
				return nil
			}
		}
	}
}

// GetAddress get address by name, if region is Global, it will get global address
func (sw *SdkWrapper) GetAddress(project, region, name string) (*compute.Address, error) {
	blog.V(3).Infof("GetAddress input: project/%s, region/%s, name/%s", project, region, name)

	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"GetAddress", ret, startTime)
	}
	sw.tryThrottle()

	var addr *compute.Address
	var err error
	if region == GlobalRegion {
		addr, err = sw.computeService.GlobalAddresses.Get(project, name).Do()
	} else {
		addr, err = sw.computeService.Addresses.Get(project, region, name).Do()
	}

	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("GetAddress failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return nil, err
	}
	blog.V(3).Infof("GetAddress response: %s", common.ToJsonString(addr))
	mf(metrics.LibCallStatusOK)
	return addr, nil
}

// GetNetworkEndpointGroups get network endpoint groups by name
func (sw *SdkWrapper) GetNetworkEndpointGroups(project, zone, name string) (*compute.NetworkEndpointGroup, error) {
	blog.V(3).Infof("GetNetworkEndpointGroups input: project/%s, zone/%s, name/%s", project, zone, name)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"GetNetworkEndpointGroups", ret, startTime)
	}
	sw.tryThrottle()

	neg, err := sw.computeService.NetworkEndpointGroups.Get(project, zone, name).Do()

	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("GetNetworkEndpointGroups failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return nil, err
	}
	blog.V(3).Infof("GetNetworkEndpointGroups response: %s", common.ToJsonString(neg))
	mf(metrics.LibCallStatusOK)
	return neg, nil
}

// ListNetworkEndpointGroups list network endpoint group
func (sw *SdkWrapper) ListNetworkEndpointGroups(project string) ([]*compute.NetworkEndpointGroup, error) {
	blog.V(3).Infof("ListNetworkEndpointGroups input: project/%s", project)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"ListNetworkEndpointGroups", ret, startTime)
	}
	sw.tryThrottle()

	var negs []*compute.NetworkEndpointGroup
	var err error
	negList, err := sw.computeService.NetworkEndpointGroups.AggregatedList(project).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("ListNetworkEndpointGroups failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return nil, fmt.Errorf(errMsg)
	}
	blog.V(3).Infof("ListNetworkEndpointGroups response: %s", common.ToJsonString(negs))
	mf(metrics.LibCallStatusOK)
	for _, v := range negList.Items {
		negs = append(negs, v.NetworkEndpointGroups...)
	}
	return negs, nil
}

// CreateNetworkEndpointGroups create network endpoint groups
func (sw *SdkWrapper) CreateNetworkEndpointGroups(project, zone, name, network, subnetwork string) error {
	blog.V(3).Infof("CreateNetworkEndpointGroups input: project/%s, zone/%s, name/%s, network/%s, subnetwork/%s",
		project, zone, name, network, subnetwork)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"CreateNetworkEndpointGroups", ret, startTime)
	}
	sw.tryThrottle()
	op, err := sw.computeService.NetworkEndpointGroups.Insert(project, zone, &compute.NetworkEndpointGroup{
		Name:                name,
		NetworkEndpointType: "GCE_VM_IP_PORT",
		Network:             network,
		Subnetwork:          subnetwork,
	}).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("CreateNetworkEndpointGroups failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := sw.Wait(ctx, project, op); err != nil {
		mf(metrics.LibCallStatusTimeout)
		errMsg := fmt.Sprintf("CreateNetworkEndpointGroups failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}

	blog.V(3).Infof("CreateNetworkEndpointGroups response: %s", common.ToJsonString(op))
	mf(metrics.LibCallStatusOK)
	return nil
}

// DeleteNetworkEndpointGroups delete network endpoint groups
func (sw *SdkWrapper) DeleteNetworkEndpointGroups(project, zone, name string) error {
	blog.V(3).Infof("DeleteNetworkEndpointGroups input: project/%s, zone/%s, name/%s", project, zone, name)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"DeleteNetworkEndpointGroups", ret, startTime)
	}
	sw.tryThrottle()
	op, err := sw.computeService.NetworkEndpointGroups.Delete(project, zone, name).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("DeleteNetworkEndpointGroups failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := sw.Wait(ctx, project, op); err != nil {
		mf(metrics.LibCallStatusTimeout)
		errMsg := fmt.Sprintf("DeleteNetworkEndpointGroups failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return err
	}

	blog.V(3).Infof("DeleteNetworkEndpointGroups response: %s", common.ToJsonString(op))
	mf(metrics.LibCallStatusOK)
	return nil
}

// GetForwardingRules get forwarding rules by name
func (sw *SdkWrapper) GetForwardingRules(project, name string) (*compute.ForwardingRule, error) {
	blog.V(3).Infof("GetForwardingRules input: project/%s, name/%s", project, name)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"GetForwardingRules", ret, startTime)
	}
	sw.tryThrottle()

	fr, err := sw.computeService.GlobalForwardingRules.Get(project, name).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("GetForwardingRules failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return nil, err
	}
	blog.V(3).Infof("GetForwardingRules response: %s", common.ToJsonString(fr))
	mf(metrics.LibCallStatusOK)
	return fr, nil
}

// CreateForwardingRules create forwarding rules
func (sw *SdkWrapper) CreateForwardingRules(project, name, targetProxy, address string, port int) error {
	blog.V(3).Infof("CreateForwardingRules input: name/%s, targetProxy/%s, address/%s, port/%d", name, targetProxy, address, port)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"CreateForwardingRules", ret, startTime)
	}
	sw.tryThrottle()

	targetProxies := "targetHttpProxies"
	if port == 443 {
		targetProxies = "targetHttpsProxies"
	}
	target := fmt.Sprintf("global/%s/%s", targetProxies, targetProxy)
	resp, err := sw.computeService.GlobalForwardingRules.Insert(project,
		&compute.ForwardingRule{
			Name:                name,
			Target:              target,
			IPAddress:           "global/addresses/" + address,
			PortRange:           strconv.Itoa(port),
			IPProtocol:          "TCP",
			LoadBalancingScheme: "EXTERNAL",
		}).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("CreateForwardingRules failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}

	blog.V(3).Infof("CreateForwardingRules response: %s", common.ToJsonString(resp))
	mf(metrics.LibCallStatusOK)
	return nil
}

// GetHealthChecks get health checks by name
func (sw *SdkWrapper) GetHealthChecks(project, name string) (*compute.HealthCheck, error) {
	blog.V(3).Infof("GetHealthChecks input: project/%s, name/%s", project, name)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"GetHealthChecks", ret, startTime)
	}
	sw.tryThrottle()

	hc, err := sw.computeService.HealthChecks.Get(project, name).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("GetHealthChecks failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return nil, err
	}
	blog.V(3).Infof("GetHealthChecks response: %s", common.ToJsonString(hc))
	mf(metrics.LibCallStatusOK)
	return hc, nil
}

// CreateHealthChecks create health checks
func (sw *SdkWrapper) CreateHealthChecks(project string, healthCheck *compute.HealthCheck) error {
	blog.V(3).Infof("CreateHealthChecks input: project/%s, healthCheck/%v", project, healthCheck)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"CreateHealthChecks", ret, startTime)
	}
	sw.tryThrottle()

	resp, err := sw.computeService.HealthChecks.Insert(project, healthCheck).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("CreateHealthChecks failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := sw.Wait(ctx, project, resp); err != nil {
		mf(metrics.LibCallStatusTimeout)
		errMsg := fmt.Sprintf("CreateHealthChecks failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}

	blog.V(3).Infof("CreateHealthChecks response: %s", common.ToJsonString(resp))
	mf(metrics.LibCallStatusOK)
	return nil
}

// UpdateHealthChecks update health checks
func (sw *SdkWrapper) UpdateHealthChecks(project string, healthCheck *compute.HealthCheck) error {
	blog.V(3).Infof("UpdateHealthChecks input: project/%s, healthCheck/%v", project, healthCheck)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"UpdateHealthChecks", ret, startTime)
	}
	sw.tryThrottle()

	resp, err := sw.computeService.HealthChecks.Update(project, healthCheck.Name, healthCheck).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("UpdateHealthChecks failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := sw.Wait(ctx, project, resp); err != nil {
		mf(metrics.LibCallStatusTimeout)
		errMsg := fmt.Sprintf("UpdateHealthChecks failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}

	blog.V(3).Infof("UpdateHealthChecks response: %s", common.ToJsonString(resp))
	mf(metrics.LibCallStatusOK)
	return nil
}

// GetBackendServices get backend services by name
func (sw *SdkWrapper) GetBackendServices(project, name string) (*compute.BackendService, error) {
	blog.V(3).Infof("GetBackendServices input: project/%s, name/%s", project, name)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"GetBackendServices", ret, startTime)
	}
	sw.tryThrottle()

	bs, err := sw.computeService.BackendServices.Get(project, name).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("GetBackendServices failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return nil, err
	}
	blog.V(3).Infof("GetBackendServices response: %s", common.ToJsonString(bs))
	mf(metrics.LibCallStatusOK)
	return bs, nil
}

// CreateBackendService create backend service
func (sw *SdkWrapper) CreateBackendService(project string, backendService *compute.BackendService) error {
	blog.V(3).Infof("CreateBackendService input: project/%s, backendService/%v", project, backendService)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"CreateBackendService", ret, startTime)
	}
	sw.tryThrottle()

	resp, err := sw.computeService.BackendServices.Insert(project, backendService).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("CreateBackendService failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := sw.Wait(ctx, project, resp); err != nil {
		mf(metrics.LibCallStatusTimeout)
		errMsg := fmt.Sprintf("CreateBackendService failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}

	blog.V(3).Infof("CreateBackendService response: %s", common.ToJsonString(resp))
	mf(metrics.LibCallStatusOK)
	return nil
}

// ListComputeZones list compute zones
func (sw *SdkWrapper) ListComputeZones(project string) ([]*compute.Zone, error) {
	blog.V(3).Infof("ListComputeZones input: project/%s", project)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"ListComputeZones", ret, startTime)
	}
	sw.tryThrottle()

	zones, err := sw.computeService.Zones.List(project).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("ListComputeZones failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return nil, fmt.Errorf(errMsg)
	}
	blog.V(3).Infof("ListComputeZones response: %s", common.ToJsonString(zones))
	mf(metrics.LibCallStatusOK)
	return zones.Items, nil
}

// GetURLMaps get url maps by name
func (sw *SdkWrapper) GetURLMaps(project, name string) (*compute.UrlMap, error) {
	blog.V(3).Infof("GetURLMaps input: project/%s, name/%s", project, name)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"GetURLMaps", ret, startTime)
	}
	sw.tryThrottle()

	urlMap, err := sw.computeService.UrlMaps.Get(project, name).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("GetURLMaps failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return nil, err
	}
	blog.V(3).Infof("GetURLMaps response: %s", common.ToJsonString(urlMap))
	mf(metrics.LibCallStatusOK)
	return urlMap, nil
}

// CreateURLMap create url map
func (sw *SdkWrapper) CreateURLMap(project string, urlMap *compute.UrlMap) error {
	blog.V(3).Infof("CreateURLMap input: project/%s, urlMap/%v", project, urlMap)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"CreateURLMap", ret, startTime)
	}
	sw.tryThrottle()

	resp, err := sw.computeService.UrlMaps.Insert(project, urlMap).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("CreateURLMap failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := sw.Wait(ctx, project, resp); err != nil {
		mf(metrics.LibCallStatusTimeout)
		errMsg := fmt.Sprintf("CreateURLMap failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}

	blog.V(3).Infof("CreateURLMap response: %s", common.ToJsonString(resp))
	mf(metrics.LibCallStatusOK)
	return nil
}

// GetTargetHTTPProxies get target http proxies by name
func (sw *SdkWrapper) GetTargetHTTPProxies(project, name string) (*compute.TargetHttpProxy, error) {
	blog.V(3).Infof("GetTargetHttpProxies input: project/%s, name/%s", project, name)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"GetTargetHttpProxies", ret, startTime)
	}
	sw.tryThrottle()

	targetHTTPProxy, err := sw.computeService.TargetHttpProxies.Get(project, name).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("GetTargetHttpProxies failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return nil, err
	}
	blog.V(3).Infof("GetTargetHttpProxies response: %s", common.ToJsonString(targetHTTPProxy))
	mf(metrics.LibCallStatusOK)
	return targetHTTPProxy, nil
}

// CreateTargetHTTPProxy create target http proxy
func (sw *SdkWrapper) CreateTargetHTTPProxy(project string, targetHTTPProxy *compute.TargetHttpProxy) error {
	blog.V(3).Infof("CreateTargetHTTPProxy input: project/%s, targetHTTPProxy/%v", project, targetHTTPProxy)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"CreateTargetHTTPProxy", ret, startTime)
	}
	sw.tryThrottle()

	resp, err := sw.computeService.TargetHttpProxies.Insert(project, targetHTTPProxy).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("CreateTargetHTTPProxy failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := sw.Wait(ctx, project, resp); err != nil {
		mf(metrics.LibCallStatusTimeout)
		errMsg := fmt.Sprintf("CreateTargetHTTPProxy failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}

	blog.V(3).Infof("CreateTargetHTTPProxy response: %s", common.ToJsonString(resp))
	mf(metrics.LibCallStatusOK)
	return nil
}

// GetTargetHTTPSProxies get target https proxies by name
func (sw *SdkWrapper) GetTargetHTTPSProxies(project, name string) (*compute.TargetHttpsProxy, error) {
	blog.V(3).Infof("GetTargetHTTPSProxies input: project/%s, name/%s", project, name)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"GetTargetHTTPSProxies", ret, startTime)
	}
	sw.tryThrottle()

	targetHTTPSProxy, err := sw.computeService.TargetHttpsProxies.Get(project, name).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("GetTargetHTTPSProxies failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return nil, err
	}
	blog.V(3).Infof("GetTargetHTTPSProxies response: %s", common.ToJsonString(targetHTTPSProxy))
	mf(metrics.LibCallStatusOK)
	return targetHTTPSProxy, nil
}

// CreateTargetHTTPSProxy create target https proxy
func (sw *SdkWrapper) CreateTargetHTTPSProxy(project string, targetHTTPSProxy *compute.TargetHttpsProxy) error {
	blog.V(3).Infof("CreateTargetHTTPSProxy input: project/%s, targetHTTPSProxy/%v", project, targetHTTPSProxy)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"CreateTargetHTTPSProxy", ret, startTime)
	}
	sw.tryThrottle()

	resp, err := sw.computeService.TargetHttpsProxies.Insert(project, targetHTTPSProxy).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("CreateTargetHTTPSProxy failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := sw.Wait(ctx, project, resp); err != nil {
		mf(metrics.LibCallStatusTimeout)
		errMsg := fmt.Sprintf("CreateTargetHTTPSProxy failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}

	blog.V(3).Infof("CreateTargetHTTPSProxy response: %s", common.ToJsonString(resp))
	mf(metrics.LibCallStatusOK)
	return nil
}

// PatchBackendService patch backend service
// only specify the fields that you want to update
func (sw *SdkWrapper) PatchBackendService(project, name string, backendService *compute.BackendService) error {
	blog.V(3).Infof("PatchBackendService input: project/%s, name/%s, backendService/%v", project, name, backendService)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"PatchBackendService", ret, startTime)
	}
	sw.tryThrottle()

	resp, err := sw.computeService.BackendServices.Patch(project, name, backendService).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("PatchBackendService failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := sw.Wait(ctx, project, resp); err != nil {
		mf(metrics.LibCallStatusTimeout)
		errMsg := fmt.Sprintf("PatchBackendService failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}

	blog.V(3).Infof("PatchBackendService response: %s", common.ToJsonString(resp))
	mf(metrics.LibCallStatusOK)
	return nil
}

// ListInstances list instances
func (sw *SdkWrapper) ListInstances(project string) (*compute.InstanceAggregatedList, error) {
	blog.V(3).Infof("ListInstances input: project/%s", project)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"ListInstances", ret, startTime)
	}
	sw.tryThrottle()

	instances, err := sw.computeService.Instances.AggregatedList(project).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("ListInstances failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return nil, fmt.Errorf(errMsg)
	}
	blog.V(3).Infof("ListInstances response: %s", common.ToJsonString(instances))
	mf(metrics.LibCallStatusOK)
	return instances, nil
}

// ListNetworkEndpoints list network endpoint groups endpoints
func (sw *SdkWrapper) ListNetworkEndpoints(project, zone, name string) (*compute.NetworkEndpointGroupsListNetworkEndpoints, error) {
	blog.V(3).Infof("ListNetworkEndpoints input: project/%s, zone/%s, name/%s", project, zone, name)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"ListNetworkEndpoints", ret, startTime)
	}
	sw.tryThrottle()

	endpoints, err := sw.computeService.NetworkEndpointGroups.ListNetworkEndpoints(project, zone, name, nil).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("ListNetworkEndpoints failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return nil, fmt.Errorf(errMsg)
	}
	blog.V(3).Infof("ListNetworkEndpoints response: %s", common.ToJsonString(endpoints))
	mf(metrics.LibCallStatusOK)
	return endpoints, nil
}

// AttachNetworkEndpoints attach network endpoint groups endpoints
func (sw *SdkWrapper) AttachNetworkEndpoints(project, zone, name string, endpoints []*compute.NetworkEndpoint) error {
	blog.V(3).Infof("AttachNetworkEndpoints input: project/%s, zone/%s, name/%s, endpoints/%v", project, zone, name, endpoints)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"AttachNetworkEndpoints", ret, startTime)
	}
	sw.tryThrottle()

	resp, err := sw.computeService.NetworkEndpointGroups.AttachNetworkEndpoints(project, zone, name, &compute.NetworkEndpointGroupsAttachEndpointsRequest{
		NetworkEndpoints: endpoints,
	}).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("AttachNetworkEndpoints failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := sw.Wait(ctx, project, resp); err != nil {
		mf(metrics.LibCallStatusTimeout)
		errMsg := fmt.Sprintf("AttachNetworkEndpoints failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}

	blog.V(3).Infof("AttachNetworkEndpoints response: %s", common.ToJsonString(resp))
	mf(metrics.LibCallStatusOK)
	return nil
}

// DetachNetworkEndpoints detach network endpoint groups endpoints
func (sw *SdkWrapper) DetachNetworkEndpoints(project, zone, name string, endpoints []*compute.NetworkEndpoint) error {
	blog.V(3).Infof("DetachNetworkEndpoints input: project/%s, zone/%s, name/%s, endpoints/%v", project, zone, name, endpoints)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"DetachNetworkEndpoints", ret, startTime)
	}
	sw.tryThrottle()

	resp, err := sw.computeService.NetworkEndpointGroups.DetachNetworkEndpoints(project, zone, name, &compute.NetworkEndpointGroupsDetachEndpointsRequest{
		NetworkEndpoints: endpoints,
	}).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("DetachNetworkEndpoints failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := sw.Wait(ctx, project, resp); err != nil {
		mf(metrics.LibCallStatusTimeout)
		errMsg := fmt.Sprintf("DetachNetworkEndpoints failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}

	blog.V(3).Infof("DetachNetworkEndpoints response: %s", common.ToJsonString(resp))
	mf(metrics.LibCallStatusOK)
	return nil
}

// PatchURLMaps patch url maps
func (sw *SdkWrapper) PatchURLMaps(project, name string, urlMap *compute.UrlMap) error {
	blog.V(3).Infof("PatchURLMaps input: project/%s, name/%s, urlMap/%s", project, name, common.ToJsonString(urlMap))

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"PatchURLMaps", ret, startTime)
	}
	sw.tryThrottle()

	resp, err := sw.computeService.UrlMaps.Patch(project, name, urlMap).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("PatchURLMaps failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := sw.Wait(ctx, project, resp); err != nil {
		mf(metrics.LibCallStatusTimeout)
		errMsg := fmt.Sprintf("PatchURLMaps failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}

	blog.V(3).Infof("PatchURLMaps response: %s", common.ToJsonString(resp))
	mf(metrics.LibCallStatusOK)
	return nil
}

// DeleteForwardingRules delete forwarding rules
func (sw *SdkWrapper) DeleteForwardingRules(project, name string) error {
	blog.V(3).Infof("DeleteForwardingRules input: project/%s, name/%s", project, name)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"DeleteForwardingRules", ret, startTime)
	}
	sw.tryThrottle()

	resp, err := sw.computeService.GlobalForwardingRules.Delete(project, name).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("DeleteForwardingRules failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := sw.Wait(ctx, project, resp); err != nil {
		mf(metrics.LibCallStatusTimeout)
		errMsg := fmt.Sprintf("DeleteForwardingRules failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return err
	}

	blog.V(3).Infof("DeleteForwardingRules response: %s", common.ToJsonString(resp))
	mf(metrics.LibCallStatusOK)
	return nil
}

// DeleteURLMaps delete url maps
func (sw *SdkWrapper) DeleteURLMaps(project, name string) error {
	blog.V(3).Infof("DeleteURLMaps input: project/%s, name/%s", project, name)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"DeleteURLMaps", ret, startTime)
	}
	sw.tryThrottle()

	resp, err := sw.computeService.UrlMaps.Delete(project, name).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("DeleteURLMaps failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := sw.Wait(ctx, project, resp); err != nil {
		mf(metrics.LibCallStatusTimeout)
		errMsg := fmt.Sprintf("DeleteURLMaps failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return err
	}

	blog.V(3).Infof("DeleteURLMaps response: %s", common.ToJsonString(resp))
	mf(metrics.LibCallStatusOK)
	return nil
}

// DeleteTargetHTTPProxy delete target http proxy
func (sw *SdkWrapper) DeleteTargetHTTPProxy(project, name string) error {
	blog.V(3).Infof("DeleteTargetHTTPProxy input: project/%s, name/%s", project, name)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"DeleteTargetHTTPProxy", ret, startTime)
	}
	sw.tryThrottle()

	resp, err := sw.computeService.TargetHttpProxies.Delete(project, name).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("DeleteTargetHTTPProxy failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := sw.Wait(ctx, project, resp); err != nil {
		mf(metrics.LibCallStatusTimeout)
		errMsg := fmt.Sprintf("DeleteTargetHTTPProxy failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return err
	}

	blog.V(3).Infof("DeleteTargetHTTPProxy response: %s", common.ToJsonString(resp))
	mf(metrics.LibCallStatusOK)
	return nil
}

// DeleteTargetHTTPSProxy delete target https proxy
func (sw *SdkWrapper) DeleteTargetHTTPSProxy(project, name string) error {
	blog.V(3).Infof("DeleteTargetHTTPSProxy input: project/%s, name/%s", project, name)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"DeleteTargetHTTPSProxy", ret, startTime)
	}
	sw.tryThrottle()

	resp, err := sw.computeService.TargetHttpsProxies.Delete(project, name).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("DeleteTargetHTTPSProxy failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := sw.Wait(ctx, project, resp); err != nil {
		mf(metrics.LibCallStatusTimeout)
		errMsg := fmt.Sprintf("DeleteTargetHTTPSProxy failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return err
	}

	blog.V(3).Infof("DeleteTargetHTTPSProxy response: %s", common.ToJsonString(resp))
	mf(metrics.LibCallStatusOK)
	return nil
}

// DeleteHealthCheck delete health check
func (sw *SdkWrapper) DeleteHealthCheck(project, name string) error {
	blog.V(3).Infof("DeleteHealthCheck input: project/%s, name/%s", project, name)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"DeleteHealthCheck", ret, startTime)
	}
	sw.tryThrottle()

	resp, err := sw.computeService.HttpHealthChecks.Delete(project, name).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("DeleteHealthCheck failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := sw.Wait(ctx, project, resp); err != nil {
		mf(metrics.LibCallStatusTimeout)
		errMsg := fmt.Sprintf("DeleteHealthCheck failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return err
	}

	blog.V(3).Infof("DeleteHealthCheck response: %s", common.ToJsonString(resp))
	mf(metrics.LibCallStatusOK)
	return nil
}

// DeleteBackendService delete backend service
func (sw *SdkWrapper) DeleteBackendService(project, name string) error {
	blog.V(3).Infof("DeleteBackendService input: project/%s, name/%s", project, name)

	startTime := time.Now()

	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetric,
			HandlerNameInMetricSDK,
			"DeleteBackendService", ret, startTime)
	}
	sw.tryThrottle()

	resp, err := sw.computeService.BackendServices.Delete(project, name).Do()
	if err != nil {
		mf(metrics.LibCallStatusErr)
		errMsg := fmt.Sprintf("DeleteBackendService failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := sw.Wait(ctx, project, resp); err != nil {
		mf(metrics.LibCallStatusTimeout)
		errMsg := fmt.Sprintf("DeleteBackendService failed, err %s", err.Error())
		blog.Errorf(errMsg)
		return err
	}

	blog.V(3).Infof("DeleteBackendService response: %s", common.ToJsonString(resp))
	mf(metrics.LibCallStatusOK)
	return nil
}
