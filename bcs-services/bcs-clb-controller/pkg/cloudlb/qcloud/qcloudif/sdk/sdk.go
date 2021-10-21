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

package sdk

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cloudListenerType "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/apis/network/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/cloudlb/qcloud/qcloudif"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	tcommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	terrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	tprofile "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	tvpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

const (
	// DescribeFilterNamePrivateIP filter name for private IP
	DescribeFilterNamePrivateIP = "private-ip-address"
	// ClbBackendTargetTypeCVM clb backend type is cvm
	// clb backend type defined the way to call register backend to listener or rule
	// when use "CVM": register backend with instance id which is get from cvm ip
	// when use "ENI": register backend with eni IP
	ClbBackendTargetTypeCVM = "CVM"
	// ClbBackendTargetTypeENI clb backend type is elastic network interface
	ClbBackendTargetTypeENI = "ENI"
	// TaskStatusDealing  task is dealing
	TaskStatusDealing = 2
	// TaskStatusFailed task is failed
	TaskStatusFailed = 1
	// TaskStatusSucceed task is successful
	TaskStatusSucceed = 0
	// ClbStatusCreating clb instance is creating
	ClbStatusCreating = 0
	// ClbStatusNormal clb instance is normal
	ClbStatusNormal = 1

	// EnvDomainTencentCloudClbDomain env name for tencent cloud clb domain
	EnvDomainTencentCloudClbDomain = "TENCENTCLOUD_CLB_DOMAIN"
	// EnvDomainTencentCloudCvmDomain env name for tencent cloud
	EnvDomainTencentCloudCvmDomain = "TENCENTCLOUD_CVM_DOMAIN"
)

// Config config for sdk client
type Config struct {
	// Region tencent cloud region
	Region string
	// ProjectID project id for tencent cloud
	ProjectID int
	// SubnetID subent id for tencent cloud vpc, only for creating private clb instance
	// it is useless when take over a existed private clb intance
	SubnetID string
	// VpcID vpc id for tencent cloud
	VpcID string
	// SecretID secret id for tencent cloud
	SecretID string
	// SecretKey secret key for tencent cloud
	SecretKey string
	// BackendType cvm or eni
	BackendType string
	// MaxTimeout times for retrying query asynchronous task result
	MaxTimeout int
	// WaitPeriodExceedLimit wait second when exceed api limit
	WaitPeriodExceedLimit int
	// WaitPeriodLBDealing wait second when retrying query asynchronous task result
	WaitPeriodLBDealing int
}

// Client client for call tencent cloud sdk
type Client struct {
	// client for operate clb
	clb *tclb.Client
	// client for query cvm
	cvm *tcvm.Client
	// client for query vpc
	vpc *tvpc.Client
	// config
	sdkConfig *Config
}

// NewClient create client for tencent cloud sdk
func NewClient(sc *Config) qcloudif.ClbAdapter {
	clbDomain := os.Getenv(EnvDomainTencentCloudClbDomain)
	cvmDomain := os.Getenv(EnvDomainTencentCloudCvmDomain)

	// referenced tencent cloud example
	credential := tcommon.NewCredential(sc.SecretID, sc.SecretKey)
	clbProfile := tprofile.NewClientProfile()
	if len(clbDomain) != 0 {
		clbProfile.HttpProfile.Endpoint = clbDomain
	}
	clbClient := &tclb.Client{}
	clbClient.Init(sc.Region).
		WithCredential(credential).
		WithProfile(clbProfile)

	cvmProfile := tprofile.NewClientProfile()
	if len(cvmDomain) != 0 {
		cvmProfile.HttpProfile.Endpoint = cvmDomain
	}
	cvmClient := &tcvm.Client{}
	cvmClient.Init(sc.Region).
		WithCredential(credential).
		WithProfile(cvmProfile)
	return &Client{
		sdkConfig: sc,
		clb:       clbClient,
		cvm:       cvmClient,
	}
}

// checkErrCode common method for check tencent cloud sdk err
func (c *Client) checkErrCode(err *terrors.TencentCloudSDKError) {
	if err.Code == RequestLimitExceededCode {
		blog.Warnf("request exceed limit, have a rest for %d second", c.sdkConfig.WaitPeriodExceedLimit)
		time.Sleep(time.Duration(c.sdkConfig.WaitPeriodLBDealing) * time.Second)
	} else if err.Code == WrongStatusCode {
		blog.Warnf("clb is dealing another action, have a rest for %d second", c.sdkConfig.WaitPeriodLBDealing)
		time.Sleep(time.Duration(c.sdkConfig.WaitPeriodLBDealing) * time.Second)
	}
}

// CreateLoadBalance call sdk to create clb, return clb id
// TODO: deal with vips
func (c *Client) CreateLoadBalance(lb *cloudListenerType.CloudLoadBalancer) (lbID string, vips []string, err error) {
	request := tclb.NewCreateLoadBalancerRequest()
	// loadbalance has application type and tradition type
	// we always use application type
	request.Forward = tcommon.Int64Ptr(LoadBalancerForwardApplication)
	// public
	if lb.NetworkType == cloudListenerType.ClbNetworkTypePublic {
		request.LoadBalancerType = tcommon.StringPtr(LoadBalancerNetworkPublic)
		// when create private loadbalance, need subentid
	} else {
		request.LoadBalancerType = tcommon.StringPtr(LoadBalancerNetworkInternal)
		request.SubnetId = tcommon.StringPtr(c.sdkConfig.SubnetID)
	}
	request.LoadBalancerName = tcommon.StringPtr(lb.Name)
	request.VpcId = tcommon.StringPtr(c.sdkConfig.VpcID)
	blog.Infof("create clb with request:\n%s", request.ToJsonString())
	response, err := c.clb.CreateLoadBalancer(request)
	if err != nil {
		if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
			c.checkErrCode(terr)
		}
		blog.Errorf("create loadbalance err %s", err.Error())
		return "", nil, fmt.Errorf("create loadbalance err %s", err.Error())
	}
	blog.Infof("create clb response:\n%s", response.ToJsonString())
	// wait util clb instance status is normal
	for counter := 0; counter < c.sdkConfig.MaxTimeout; counter++ {
		clb, err := c.doDescribeLoadBalance(lb.Name)
		if err != nil {
			return "", nil, fmt.Errorf("describe clb by name %s failed, err %s", lb.Name, err.Error())
		}
		if clb == nil {
			return "", nil, fmt.Errorf("describe clb by name %s return nil", lb.Name)
		}
		if clb.Status == nil {
			return "", nil, fmt.Errorf("clb status is empty")
		}
		if *clb.Status == ClbStatusCreating {
			time.Sleep(time.Duration(c.sdkConfig.WaitPeriodLBDealing) * time.Second)
			continue
		} else {
			var vips []string
			if len(clb.LoadBalancerVips) > 0 {
				for _, vipPtr := range clb.LoadBalancerVips {
					vips = append(vips, *vipPtr)
				}
			}
			return *clb.LoadBalancerId, vips, nil
		}
	}
	return "", nil, fmt.Errorf("waiting for loadbalance creating timeout")
}

// do actually describe loadbalance by loadbalance name
func (c *Client) doDescribeLoadBalance(name string) (*tclb.LoadBalancer, error) {
	request := tclb.NewDescribeLoadBalancersRequest()
	request.Forward = tcommon.Int64Ptr(LoadBalancerForwardApplication)
	request.LoadBalancerName = tcommon.StringPtr(name)
	blog.Infof("describe clb request:\n%s", request.ToJsonString())
	response, err := c.clb.DescribeLoadBalancers(request)
	if err != nil {
		if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
			c.checkErrCode(terr)
		}
		blog.Errorf("describe loadbalance err %s", err.Error())
		return nil, fmt.Errorf("describe loadbalance err %s", err.Error())
	}
	if response.Response == nil {
		blog.Warnf("describe clb empty response")
		return nil, nil
	}
	if *response.Response.TotalCount == 0 {
		blog.Warnf("describe clb return zero element")
		return nil, nil
	}
	for _, lb := range response.Response.LoadBalancerSet {
		if lb.LoadBalancerName != nil && *lb.LoadBalancerName == name {
			return lb, nil
		}
	}
	return nil, fmt.Errorf("loadbalance with name %s not found", name)
}

// DescribeLoadBalance describe clb by name, return clb info, and return true if it is existed
func (c *Client) DescribeLoadBalance(name string) (*cloudListenerType.CloudLoadBalancer, bool, error) {
	lb, err := c.doDescribeLoadBalance(name)
	if err != nil {
		return nil, false, err
	}
	if lb == nil {
		return nil, false, nil
	}
	var lbNetworkType string
	if *lb.LoadBalancerType == LoadBalancerNetworkPublic {
		lbNetworkType = cloudListenerType.ClbNetworkTypePublic
	} else {
		lbNetworkType = cloudListenerType.ClbNetworkTypePrivate
	}
	var vips []string
	if len(lb.LoadBalancerVips) != 0 {
		for _, vipPtr := range lb.LoadBalancerVips {
			vips = append(vips, *vipPtr)
		}
	}
	return &cloudListenerType.CloudLoadBalancer{
		ID:          *lb.LoadBalancerId,
		Name:        *lb.LoadBalancerName,
		NetworkType: lbNetworkType,
		VIPS:        vips,
	}, true, nil
}

// CreateListener create listener
// create7LayerListener and create4LayerListener only return request object
// CreateListener does request action and waits for result
func (c *Client) CreateListener(listener *cloudListenerType.CloudListener) (listenerID string, err error) {
	protocol, ok := ProtocolBcs2SDKMap[listener.Spec.Protocol]
	if !ok {
		return "", fmt.Errorf("protocol %s cannot be recognized", listener.Spec.Protocol)
	}
	var request *tclb.CreateListenerRequest
	if protocol == ListenerProtocolHTTP || protocol == ListenerProtocolHTTPS {
		request, err = c.create7LayerListener(listener)
		if err != nil {
			return "", fmt.Errorf("create 7 layer listener failed, err %s", err.Error())
		}
	} else {
		request, err = c.create4LayerListener(listener)
		if err != nil {
			return "", fmt.Errorf("create 4 layer listener failed, err %s", err.Error())
		}
	}

	blog.Infof("create listener request:\n%s", request.ToJsonString())
	// wait util listener is created
	counter := 0
	var response *tclb.CreateListenerResponse
	for ; counter < c.sdkConfig.MaxTimeout; counter++ {
		response, err = c.clb.CreateListener(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				c.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("create listener failed with request %s, err %s", request.ToJsonString(), err.Error())
			return "", fmt.Errorf("create listener failed with request %s, err %s", request.ToJsonString(), err.Error())
		}
		// should return listener id in response
		if len(response.Response.ListenerIds) == 0 {
			blog.Errorf("create listener return zero length ids with request %s, err %s",
				request.ToJsonString(), err.Error())
			return "", fmt.Errorf("create listener return zero length ids with request %s, err %s",
				request.ToJsonString(), err.Error())
		}
		blog.Infof("create listener response:\n%s", response.ToJsonString())
		break
	}
	if counter >= c.sdkConfig.MaxTimeout {
		blog.Errorf("create listener with request %s timeout", request.ToJsonString())
		return "", fmt.Errorf("create listener with request %s timeout", request.ToJsonString())
	}
	// creating listener is asynchronous
	err = c.waitTaskDone(*response.Response.RequestId)
	if err != nil {
		return "", err
	}
	return *response.Response.ListenerIds[0], nil
}

// create7LayerListener create request for creating 7 layer listener
// create listener won't bind backends
func (c *Client) create7LayerListener(listener *cloudListenerType.CloudListener) (*tclb.CreateListenerRequest, error) {
	request := tclb.NewCreateListenerRequest()
	request.LoadBalancerId = tcommon.StringPtr(listener.Spec.LoadBalancerID)
	request.Ports = []*int64{
		tcommon.Int64Ptr(int64(listener.Spec.ListenPort)),
	}
	request.ListenerNames = []*string{
		tcommon.StringPtr(listener.GetName()),
	}
	protocol, ok := ProtocolBcs2SDKMap[listener.Spec.Protocol]
	if !ok {
		return nil, fmt.Errorf("protocol %s cannot be recognized", listener.Spec.Protocol)
	}
	request.Protocol = tcommon.StringPtr(protocol)
	// https tls config
	if protocol == ListenerProtocolHTTPS {
		if listener.Spec.TLS == nil {
			return nil, fmt.Errorf("tls config must be defined for protocol %s listener", protocol)
		}
		request.Certificate = &tclb.CertificateInput{}
		sslMode, ok := SSLModeBcs2SDKMap[listener.Spec.TLS.Mode]
		if !ok {
			return nil, fmt.Errorf("invalid ssl mode %s", listener.Spec.TLS.Mode)
		}
		request.Certificate.SSLMode = tcommon.StringPtr(sslMode)
		if len(listener.Spec.TLS.CertID) != 0 {
			request.Certificate.CertId = tcommon.StringPtr(listener.Spec.TLS.CertID)
		}
		if len(listener.Spec.TLS.CertServerName) != 0 {
			request.Certificate.CertName = tcommon.StringPtr(listener.Spec.TLS.CertServerName)
		}
		if len(listener.Spec.TLS.CertServerKey) != 0 {
			request.Certificate.CertKey = tcommon.StringPtr(listener.Spec.TLS.CertServerKey)
		}
		if len(listener.Spec.TLS.CertServerContent) != 0 {
			request.Certificate.CertContent = tcommon.StringPtr(listener.Spec.TLS.CertServerContent)
		}
		if len(listener.Spec.TLS.CertCaID) != 0 {
			request.Certificate.CertCaId = tcommon.StringPtr(listener.Spec.TLS.CertCaID)
		}
		if len(listener.Spec.TLS.CertClientCaName) != 0 {
			request.Certificate.CertCaName = tcommon.StringPtr(listener.Spec.TLS.CertClientCaName)
		}
		if len(listener.Spec.TLS.CertClientCaContent) != 0 {
			request.Certificate.CertCaContent = tcommon.StringPtr(listener.Spec.TLS.CertClientCaContent)
		}
	}
	// 7 layer listener has no target group, deal with target group in rule operations
	return request, nil
}

// create4LayerListener create request for creating 4 layer listener
// create listener won't bind backends
func (c *Client) create4LayerListener(listener *cloudListenerType.CloudListener) (*tclb.CreateListenerRequest, error) {
	request := tclb.NewCreateListenerRequest()
	request.LoadBalancerId = tcommon.StringPtr(listener.Spec.LoadBalancerID)
	request.Ports = []*int64{
		tcommon.Int64Ptr(int64(listener.Spec.ListenPort)),
	}
	request.ListenerNames = []*string{
		tcommon.StringPtr(listener.GetName()),
	}
	protocol, ok := ProtocolBcs2SDKMap[listener.Spec.Protocol]
	if !ok {
		return nil, fmt.Errorf("protocol %s cannot be recognized", listener.Spec.Protocol)
	}
	request.Protocol = tcommon.StringPtr(protocol)
	// 4 layer listener has target group
	if listener.Spec.TargetGroup != nil {
		request.SessionExpireTime = tcommon.Int64Ptr(int64(listener.Spec.TargetGroup.SessionExpire))
		lbPolicy := LBAlgorithmRoundRobin
		if validPolicy, ok := LBAlgorithmTypeBcs2SDKMap[listener.Spec.TargetGroup.LBPolicy]; ok {
			lbPolicy = validPolicy
		}
		request.Scheduler = tcommon.StringPtr(lbPolicy)
		if listener.Spec.TargetGroup.HealthCheck != nil {
			request.HealthCheck = &tclb.HealthCheck{}
			request.HealthCheck.HealthSwitch = tcommon.Int64Ptr(int64(listener.Spec.TargetGroup.HealthCheck.Enabled))
			request.HealthCheck.IntervalTime = tcommon.Int64Ptr(
				int64(listener.Spec.TargetGroup.HealthCheck.IntervalTime))
			request.HealthCheck.HealthNum = tcommon.Int64Ptr(int64(listener.Spec.TargetGroup.HealthCheck.HealthNum))
			request.HealthCheck.UnHealthNum = tcommon.Int64Ptr(int64(listener.Spec.TargetGroup.HealthCheck.UnHealthNum))
			request.HealthCheck.TimeOut = tcommon.Int64Ptr(int64(listener.Spec.TargetGroup.HealthCheck.Timeout))
		}
	}
	return request, nil
}

// DeleteListener delete listener by loadbalance id and listener id
func (c *Client) DeleteListener(lbID, listenerID string) error {
	request := tclb.NewDeleteListenerRequest()
	request.LoadBalancerId = tcommon.StringPtr(lbID)
	request.ListenerId = tcommon.StringPtr(listenerID)
	blog.Infof("delete listener request:\n%s", request.ToJsonString())
	counter := 0
	var err error
	var response *tclb.DeleteListenerResponse
	for ; counter < c.sdkConfig.MaxTimeout; counter++ {
		response, err = c.clb.DeleteListener(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				c.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("delete listener with request %s failed, err %s", request.ToJsonString(), err.Error())
			return fmt.Errorf("delete listener with request %s failed, err %s", request.ToJsonString(), err.Error())
		}
		blog.Infof("delete listener response:\n%s", response.ToJsonString())
		break
	}
	if counter >= c.sdkConfig.MaxTimeout {
		blog.Errorf("delete listener with request %s timeout", request.ToJsonString())
		return fmt.Errorf("delete listener with request %s timeout", request.ToJsonString())
	}
	return c.waitTaskDone(*response.Response.RequestId)
}

// DescribeListener describe listener
// by loadbalance id, either listener id or listener port
// this function does different conversion for 4 layer listener and 7 layer listener.
func (c *Client) DescribeListener(lbID, listenerID string, port int) (
	listener *cloudListenerType.CloudListener, isExisted bool, err error) {
	request := tclb.NewDescribeListenersRequest()
	request.LoadBalancerId = tcommon.StringPtr(lbID)
	if len(listenerID) != 0 {
		request.ListenerIds = []*string{
			tcommon.StringPtr(listenerID),
		}
	}
	if port > 0 {
		request.Port = tcommon.Int64Ptr(int64(port))
	}
	// do request
	blog.Infof("describe listener request:\n%s", request.ToJsonString())
	for counter := 0; counter < c.sdkConfig.MaxTimeout; counter++ {
		response, err := c.clb.DescribeListeners(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				c.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("describe listener failed, err %s", err.Error())
			return nil, false, fmt.Errorf("describe listener failed, err %s", err.Error())
		}
		blog.Infof("describe listener response:\n%s", response.ToJsonString())
		if len(response.Response.Listeners) == 0 {
			blog.Warnf("describe response return zero listener")
			return nil, false, nil
		}
		if len(response.Response.Listeners) != 1 {
			blog.Errorf("describe response invalid listeners length %d", len(response.Response.Listeners))
			return nil, false, fmt.Errorf("describe response invalid listeners length %d",
				len(response.Response.Listeners))
		}
		listener := response.Response.Listeners[0]
		protocol, ok := ProtocolSDK2BcsMap[*listener.Protocol]
		if !ok {
			return nil, false, fmt.Errorf("unrecognized protocol %s", *listener.Protocol)
		}
		//TODO: get full information from tencent cloud
		retListener := &cloudListenerType.CloudListener{
			ObjectMeta: metav1.ObjectMeta{
				Name: *listener.ListenerName,
			},
			Spec: cloudListenerType.CloudListenerSpec{
				ListenerID:     *listener.ListenerId,
				LoadBalancerID: lbID,
				Protocol:       protocol,
				ListenPort:     int(*listener.Port),
				Rules:          make([]*cloudListenerType.Rule, 0),
			},
		}
		// convert 4 layer listener
		if *listener.Protocol == ListenerProtocolTCP || *listener.Protocol == ListenerProtocolUDP {
			lbPolicy := LBAlgorithmRoundRobin
			if validPolicy, ok := LBAlgorithmTypeSDK2BcsMap[*listener.Scheduler]; ok {
				lbPolicy = validPolicy
			}
			// convert target group
			retListener.Spec.TargetGroup = &cloudListenerType.TargetGroup{
				SessionExpire: int(*listener.SessionExpireTime),
				LBPolicy:      lbPolicy,
			}
			retListener.Spec.TargetGroup.HealthCheck = &cloudListenerType.TargetGroupHealthCheck{}
			if listener.HealthCheck != nil {
				if listener.HealthCheck.HealthSwitch != nil {
					retListener.Spec.TargetGroup.HealthCheck.Enabled = int(*listener.HealthCheck.HealthSwitch)
				}
				if listener.HealthCheck.IntervalTime != nil {
					retListener.Spec.TargetGroup.HealthCheck.IntervalTime = int(*listener.HealthCheck.IntervalTime)
				}
				if listener.HealthCheck.HealthNum != nil {
					retListener.Spec.TargetGroup.HealthCheck.HealthNum = int(*listener.HealthCheck.HealthNum)
				}
				if listener.HealthCheck.UnHealthNum != nil {
					retListener.Spec.TargetGroup.HealthCheck.UnHealthNum = int(*listener.HealthCheck.UnHealthNum)
				}
				if listener.HealthCheck.TimeOut != nil {
					retListener.Spec.TargetGroup.HealthCheck.Timeout = int(*listener.HealthCheck.TimeOut)
				}
			} else {
				retListener.Spec.TargetGroup.HealthCheck.Enabled = 0
			}
			return retListener, true, nil
		}
		// convert tls config
		if *listener.Protocol == ListenerProtocolHTTPS {
			sslMode, _ := SSLModeSDK2BcsMap[*listener.Certificate.SSLMode]
			retListener.Spec.TLS = &cloudListenerType.CloudListenerTls{
				Mode:   sslMode,
				CertID: *listener.Certificate.CertId,
			}
			if listener.Certificate.CertCaId != nil {
				retListener.Spec.TLS.CertCaID = *listener.Certificate.CertCaId
			}
		}
		// convert 7 layer rules
		for _, rule := range listener.Rules {
			newRule := &cloudListenerType.Rule{
				ID:     *rule.LocationId,
				Domain: *rule.Domain,
				URL:    *rule.Url,
				TargetGroup: &cloudListenerType.TargetGroup{
					SessionExpire: int(*rule.SessionExpireTime),
					LBPolicy:      *rule.Scheduler,
				},
			}
			// convert target group for every rule
			newRule.TargetGroup.HealthCheck = &cloudListenerType.TargetGroupHealthCheck{}
			if rule.HealthCheck != nil {
				newRule.TargetGroup.HealthCheck.Enabled = 1
				newRule.TargetGroup.HealthCheck.IntervalTime = int(*rule.HealthCheck.IntervalTime)
				newRule.TargetGroup.HealthCheck.HealthNum = int(*rule.HealthCheck.HealthNum)
				newRule.TargetGroup.HealthCheck.UnHealthNum = int(*rule.HealthCheck.UnHealthNum)
				newRule.TargetGroup.HealthCheck.Timeout = int(*rule.HealthCheck.TimeOut)
				newRule.TargetGroup.HealthCheck.HTTPCode = int(*rule.HealthCheck.HttpCode)
				newRule.TargetGroup.HealthCheck.HTTPCheckPath = *rule.HealthCheck.HttpCheckPath
			} else {
				newRule.TargetGroup.HealthCheck.Enabled = 0
			}
			retListener.Spec.Rules = append(retListener.Spec.Rules, newRule)
		}
		return retListener, true, nil
	}
	blog.Errorf("describe listener with request %s timeout", request.ToJsonString())
	return nil, false, fmt.Errorf("describe listener with request %s timeout", request.ToJsonString())
}

// ModifyListenerAttribute modify listener attribute
// modify7LayerListenerAttribute can only modify https listener tls config, leave other modify actions to ModifyRuleAttribute
// modify4LayerListenerAttribute can modify health check, session, lb policy for 4 layer listener
func (c *Client) ModifyListenerAttribute(listener *cloudListenerType.CloudListener) error {
	protocol, ok := ProtocolBcs2SDKMap[listener.Spec.Protocol]
	if !ok {
		return fmt.Errorf("unrecognized protocol %s", listener.Spec.Protocol)
	}
	if protocol == ListenerProtocolHTTPS {
		if listener.Spec.TLS == nil {
			return fmt.Errorf("https with nil tls config")
		}
		return c.modify7LayerListenerAttribute(listener)
	}
	if listener.Spec.TargetGroup == nil {
		return fmt.Errorf("listener.spec.targetgroup is nil")
	}
	if listener.Spec.TargetGroup.HealthCheck == nil {
		return fmt.Errorf("listener spec.targetgroup.healthcheck is nil")
	}
	return c.modify4LayerListenerAttribute(listener)
}

// modify7LayerListenerAttribute modify tls configs for https listener
// http listener don't support modify attribute
func (c *Client) modify7LayerListenerAttribute(listener *cloudListenerType.CloudListener) error {
	request := tclb.NewModifyListenerRequest()
	request.ListenerId = tcommon.StringPtr(listener.Spec.ListenerID)
	request.LoadBalancerId = tcommon.StringPtr(listener.Spec.LoadBalancerID)
	request.Certificate = &tclb.CertificateInput{}
	sslMode, ok := SSLModeBcs2SDKMap[listener.Spec.TLS.Mode]
	if !ok {
		return fmt.Errorf("invalid ssl mode %s", listener.Spec.TLS.Mode)
	}
	request.Certificate.SSLMode = tcommon.StringPtr(sslMode)
	if len(listener.Spec.TLS.CertID) != 0 {
		request.Certificate.CertId = tcommon.StringPtr(listener.Spec.TLS.CertID)
	}
	if len(listener.Spec.TLS.CertCaID) != 0 {
		request.Certificate.CertCaId = tcommon.StringPtr(listener.Spec.TLS.CertCaID)
	}
	if len(listener.Spec.TLS.CertServerName) != 0 &&
		len(listener.Spec.TLS.CertServerKey) != 0 &&
		len(listener.Spec.TLS.CertServerContent) != 0 {
		request.Certificate.CertName = tcommon.StringPtr(listener.Spec.TLS.CertServerName)
		request.Certificate.CertKey = tcommon.StringPtr(listener.Spec.TLS.CertServerKey)
		request.Certificate.CertContent = tcommon.StringPtr(listener.Spec.TLS.CertServerContent)
	}
	if len(listener.Spec.TLS.CertClientCaName) != 0 &&
		len(listener.Spec.TLS.CertClientCaContent) != 0 {
		request.Certificate.CertCaName = tcommon.StringPtr(listener.Spec.TLS.CertClientCaName)
		request.Certificate.CertCaContent = tcommon.StringPtr(listener.Spec.TLS.CertClientCaContent)
	}
	counter := 0
	var err error
	var response *tclb.ModifyListenerResponse
	for ; counter < c.sdkConfig.MaxTimeout; counter++ {
		response, err = c.clb.ModifyListener(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				c.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("modify 7 layer listener failed, err %s", err.Error())
			return fmt.Errorf("modify 7 layer listener failed, err %s", err.Error())
		}
		blog.Infof("modify 7 layer listener response:\n%s", response.ToJsonString())
		break
	}
	if counter >= c.sdkConfig.MaxTimeout {
		blog.Errorf("modify 7 layer listener timeout")
		return fmt.Errorf("modify 7 layer listener timeout")
	}
	return c.waitTaskDone(*response.Response.RequestId)
}

// modify4LayerListenerAttribute modify health check, session config, lb policy for 4 layer listener
func (c *Client) modify4LayerListenerAttribute(listener *cloudListenerType.CloudListener) error {
	request := tclb.NewModifyListenerRequest()
	request.ListenerId = tcommon.StringPtr(listener.Spec.ListenerID)
	request.LoadBalancerId = tcommon.StringPtr(listener.Spec.LoadBalancerID)
	if listener.Spec.TargetGroup == nil {
		return fmt.Errorf("target group for 4 layer listener cannot be empty, error listener %v", listener)
	}
	request.SessionExpireTime = tcommon.Int64Ptr(int64(listener.Spec.TargetGroup.SessionExpire))
	lbPolicy := LBAlgorithmRoundRobin
	if validPolicy, ok := LBAlgorithmTypeBcs2SDKMap[listener.Spec.TargetGroup.LBPolicy]; ok {
		lbPolicy = validPolicy
	}
	request.Scheduler = tcommon.StringPtr(lbPolicy)
	if listener.Spec.TargetGroup.HealthCheck != nil {
		request.HealthCheck = &tclb.HealthCheck{}
		request.HealthCheck.HealthSwitch = tcommon.Int64Ptr(int64(listener.Spec.TargetGroup.HealthCheck.Enabled))
		if listener.Spec.TargetGroup.HealthCheck.Enabled == 1 {
			request.HealthCheck.HealthNum = tcommon.Int64Ptr(int64(listener.Spec.TargetGroup.HealthCheck.HealthNum))
			request.HealthCheck.UnHealthNum = tcommon.Int64Ptr(int64(listener.Spec.TargetGroup.HealthCheck.UnHealthNum))
			request.HealthCheck.IntervalTime = tcommon.Int64Ptr(
				int64(listener.Spec.TargetGroup.HealthCheck.IntervalTime))
			request.HealthCheck.TimeOut = tcommon.Int64Ptr(int64(listener.Spec.TargetGroup.HealthCheck.Timeout))
		}
	}
	counter := 0
	var err error
	var response *tclb.ModifyListenerResponse
	for ; counter < c.sdkConfig.MaxTimeout; counter++ {
		response, err = c.clb.ModifyListener(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				c.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("modify 4 layer listener failed, err %s", err.Error())
			return fmt.Errorf("modify 4 layer listener failed, err %s", err.Error())
		}
		blog.Infof("modify 4 layer listener response:\n%s", response.ToJsonString())
		break
	}
	if counter >= c.sdkConfig.MaxTimeout {
		blog.Errorf("modify 4 layer listener timeout")
		return fmt.Errorf("modify 4 layer listener timeout")
	}
	return c.waitTaskDone(*response.Response.RequestId)
}

// CreateRules create rules for clb listener
func (c *Client) CreateRules(lbID, listenerID string, rules cloudListenerType.RuleList) error {
	request := tclb.NewCreateRuleRequest()
	request.LoadBalancerId = tcommon.StringPtr(lbID)
	request.ListenerId = tcommon.StringPtr(listenerID)
	// convert rules in CloudListener
	for _, rule := range rules {
		ruleInput := &tclb.RuleInput{}
		ruleInput.Domain = tcommon.StringPtr(rule.Domain)
		ruleInput.Url = tcommon.StringPtr(rule.URL)
		if rule.TargetGroup != nil {
			lbPolicy := LBAlgorithmRoundRobin
			if validPolicy, ok := LBAlgorithmTypeBcs2SDKMap[rule.TargetGroup.LBPolicy]; ok {
				lbPolicy = validPolicy
			}
			ruleInput.Scheduler = tcommon.StringPtr(lbPolicy)
			ruleInput.SessionExpireTime = tcommon.Int64Ptr(int64(rule.TargetGroup.SessionExpire))
			if rule.TargetGroup.HealthCheck != nil {
				ruleInput.HealthCheck = &tclb.HealthCheck{}
				ruleInput.HealthCheck.HealthSwitch = tcommon.Int64Ptr(int64(rule.TargetGroup.HealthCheck.Enabled))
				if rule.TargetGroup.HealthCheck.Enabled == 1 {
					ruleInput.HealthCheck.HealthNum = tcommon.Int64Ptr(int64(rule.TargetGroup.HealthCheck.HealthNum))
					ruleInput.HealthCheck.UnHealthNum = tcommon.Int64Ptr(
						int64(rule.TargetGroup.HealthCheck.UnHealthNum))
					ruleInput.HealthCheck.IntervalTime = tcommon.Int64Ptr(
						int64(rule.TargetGroup.HealthCheck.IntervalTime))
					ruleInput.HealthCheck.TimeOut = tcommon.Int64Ptr(int64(rule.TargetGroup.HealthCheck.Timeout))
					ruleInput.HealthCheck.HttpCheckPath = tcommon.StringPtr(rule.TargetGroup.HealthCheck.HTTPCheckPath)
					ruleInput.HealthCheck.HttpCode = tcommon.Int64Ptr(int64(rule.TargetGroup.HealthCheck.HTTPCode))
				}
			}
		}
		request.Rules = append(request.Rules, ruleInput)
	}
	blog.Infof("create rules request:\n%s", request.ToJsonString())
	counter := 0
	var err error
	var response *tclb.CreateRuleResponse
	for ; counter < c.sdkConfig.MaxTimeout; counter++ {
		response, err = c.clb.CreateRule(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				c.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("create rule failed, err %s", err.Error())
			return fmt.Errorf("create rule failed, err %s", err.Error())
		}
		blog.Infof("create rules response:\n%s", response.ToJsonString())
		break
	}
	if counter >= c.sdkConfig.MaxTimeout {
		blog.Errorf("create rules timeout")
		return fmt.Errorf("create rules timeout")
	}
	return c.waitTaskDone(*response.Response.RequestId)
}

// DeleteRule delete rule of clb listener by domain and url, loadbalance id and listener id
func (c *Client) DeleteRule(lbID, listenerID, domain, url string) error {
	request := tclb.NewDeleteRuleRequest()
	request.LoadBalancerId = tcommon.StringPtr(lbID)
	request.ListenerId = tcommon.StringPtr(listenerID)
	request.Domain = tcommon.StringPtr(domain)
	request.Url = tcommon.StringPtr(url)
	blog.Infof("delete rule request:\n%s", request.ToJsonString())
	counter := 0
	var err error
	var response *tclb.DeleteRuleResponse
	for ; counter < c.sdkConfig.MaxTimeout; counter++ {
		response, err = c.clb.DeleteRule(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				c.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("delete rule failed, err %s", err.Error())
			return fmt.Errorf("delete rule failed, err %s", err.Error())
		}
		blog.Infof("delete rule response:\n%s", response.ToJsonString())
		break
	}
	if counter >= c.sdkConfig.MaxTimeout {
		blog.Errorf("delete rule timeout")
		return fmt.Errorf("delete rule timeout")
	}
	return c.waitTaskDone(*response.Response.RequestId)
}

// DescribeRuleByDomainAndURL describe rule by domain and url
// call DescribeListener api to find the certain listener
// traverse all the rules in listener with certain domain and url
func (c *Client) DescribeRuleByDomainAndURL(loadBalanceID, listenerID, domain, url string) (
	rule *cloudListenerType.Rule, isExisted bool, err error) {
	request := tclb.NewDescribeListenersRequest()
	request.LoadBalancerId = tcommon.StringPtr(loadBalanceID)
	request.ListenerIds = []*string{
		tcommon.StringPtr(listenerID),
	}

	blog.Infof("describe listener request:\n%s", request.ToJsonString())
	for counter := 0; counter < c.sdkConfig.MaxTimeout; counter++ {
		response, err := c.clb.DescribeListeners(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				c.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("describe listener failed, err %s", err.Error())
			return nil, false, fmt.Errorf("describe listener failed, err %s", err.Error())
		}
		blog.Infof("describe listener response:\n%s", response.ToJsonString())
		if len(response.Response.Listeners) != 1 {
			blog.Errorf("describe response invalid listeners length %d", len(response.Response.Listeners))
			return nil, false, fmt.Errorf("describe response invalid listeners length %d",
				len(response.Response.Listeners))
		}
		listener := response.Response.Listeners[0]
		// find rule with domain and url
		for _, ruleOutput := range listener.Rules {
			if *ruleOutput.Domain == domain && *ruleOutput.Url == url {
				retRule := &cloudListenerType.Rule{
					ID:     *ruleOutput.LocationId,
					Domain: domain,
					URL:    url,
				}
				lbPolicy := LBAlgorithmRoundRobin
				if validPolicy, ok := LBAlgorithmTypeSDK2BcsMap[*ruleOutput.Scheduler]; ok {
					lbPolicy = validPolicy
				}
				retRule.TargetGroup = &cloudListenerType.TargetGroup{
					SessionExpire: int(*ruleOutput.SessionExpireTime),
					LBPolicy:      lbPolicy,
				}
				// health check config
				if ruleOutput.HealthCheck != nil {
					retHealth := &cloudListenerType.TargetGroupHealthCheck{}
					if ruleOutput.HealthCheck.HealthSwitch != nil {
						retHealth.Enabled = int(*ruleOutput.HealthCheck.HealthSwitch)
					}
					if ruleOutput.HealthCheck.IntervalTime != nil {
						retHealth.IntervalTime = int(*ruleOutput.HealthCheck.IntervalTime)
					}
					if ruleOutput.HealthCheck.HealthNum != nil {
						retHealth.HealthNum = int(*ruleOutput.HealthCheck.HealthNum)
					}
					if ruleOutput.HealthCheck.UnHealthNum != nil {
						retHealth.UnHealthNum = int(*ruleOutput.HealthCheck.UnHealthNum)
					}
					if ruleOutput.HealthCheck.HttpCode != nil {
						retHealth.HTTPCode = int(*ruleOutput.HealthCheck.HttpCode)
					}
					if ruleOutput.HealthCheck.HttpCheckPath != nil {
						retHealth.HTTPCheckPath = string(*ruleOutput.HealthCheck.HttpCheckPath)
					}
					retRule.TargetGroup.HealthCheck = retHealth
				}
				return retRule, true, nil
			}
		}
		blog.Infof("rule %s %s no found with %s %s", domain, url, listenerID, loadBalanceID)
		return nil, false, nil
	}
	blog.Errorf("describe rule with request %s timeout", request.ToJsonString())
	return nil, false, fmt.Errorf("describe rule with request %s timeout", request.ToJsonString())
}

// ModifyRuleAttribute modify rule attributes
func (c *Client) ModifyRuleAttribute(loadBalanceID, listenerID string, rule *cloudListenerType.Rule) error {
	request := tclb.NewModifyRuleRequest()
	request.LoadBalancerId = tcommon.StringPtr(loadBalanceID)
	request.ListenerId = tcommon.StringPtr(listenerID)
	request.LocationId = tcommon.StringPtr(rule.ID)
	request.SessionExpireTime = tcommon.Int64Ptr(int64(rule.TargetGroup.SessionExpire))
	lbPolicy := LBAlgorithmRoundRobin
	if validPolicy, ok := LBAlgorithmTypeBcs2SDKMap[rule.TargetGroup.LBPolicy]; ok {
		lbPolicy = validPolicy
	}
	request.Scheduler = tcommon.StringPtr(lbPolicy)
	// health check config
	if rule.TargetGroup.HealthCheck != nil {
		request.HealthCheck = &tclb.HealthCheck{}
		request.HealthCheck.HealthSwitch = tcommon.Int64Ptr(int64(rule.TargetGroup.HealthCheck.Enabled))
		if rule.TargetGroup.HealthCheck.Enabled == 1 {
			request.HealthCheck.HealthNum = tcommon.Int64Ptr(int64(rule.TargetGroup.HealthCheck.HealthNum))
			request.HealthCheck.UnHealthNum = tcommon.Int64Ptr(int64(rule.TargetGroup.HealthCheck.UnHealthNum))
			request.HealthCheck.IntervalTime = tcommon.Int64Ptr(int64(rule.TargetGroup.HealthCheck.IntervalTime))
			request.HealthCheck.TimeOut = tcommon.Int64Ptr(int64(rule.TargetGroup.HealthCheck.Timeout))
			request.HealthCheck.HttpCode = tcommon.Int64Ptr(int64(rule.TargetGroup.HealthCheck.HTTPCode))
			request.HealthCheck.HttpCheckPath = tcommon.StringPtr(rule.TargetGroup.HealthCheck.HTTPCheckPath)
		}
	}
	// do request
	blog.Infof("modify rule with %v", request.ToJsonString())
	counter := 0
	var err error
	var response *tclb.ModifyRuleResponse
	for ; counter < c.sdkConfig.MaxTimeout; counter++ {
		response, err = c.clb.ModifyRule(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				c.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("modify rule failed, err %s", err.Error())
			return fmt.Errorf("modify rule failed, err %s", err.Error())
		}
		blog.Infof("modify rule response:\n%s", response.ToJsonString())
		break
	}
	if counter >= c.sdkConfig.MaxTimeout {
		blog.Errorf("modify rule timeout")
		return fmt.Errorf("modify rule timeout")
	}
	return c.waitTaskDone(*response.Response.RequestId)
}

// inStringPtrSlice see if StringPtr array contains key
func inStringPtrSlice(key string, array []*string) bool {
	for _, e := range array {
		if key == *e {
			return true
		}
	}
	return false
}

// getCVMInstanceIDMapByIP get (ip, CVM InstanceID) from ip array
// when use backend is cvm, we use cvm instance id to regsiter listener
func (c *Client) getCVMInstanceIDMapByIP(ips []string) (map[string]string, error) {
	request := tcvm.NewDescribeInstancesRequest()
	// construct filter
	privateIPFilter := &tcvm.Filter{
		Name: tcommon.StringPtr(DescribeFilterNamePrivateIP),
	}
	for _, ip := range ips {
		privateIPFilter.Values = append(privateIPFilter.Values, tcommon.StringPtr(ip))
	}
	request.Filters = []*tcvm.Filter{
		privateIPFilter,
	}
	blog.Infof("describe instance id by ips request:\n%s", request.ToJsonString())
	// do request
	for counter := 0; counter < c.sdkConfig.MaxTimeout; counter++ {
		response, err := c.cvm.DescribeInstances(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				c.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("describe instances by ips failed, err %s", err.Error())
			return nil, fmt.Errorf("describe instances by ips failed, err %s", err.Error())
		}
		if *response.Response.TotalCount == 0 {
			blog.Errorf("describe instances by ip return zero element")
			return nil, fmt.Errorf("describe instances by ip return zero element")
		}
		ipMap := make(map[string]string)
		for _, ip := range ips {
			for _, ins := range response.Response.InstanceSet {
				if inStringPtrSlice(ip, ins.PrivateIpAddresses) {
					ipMap[ip] = *ins.InstanceId
					break
				}
			}
		}
		return ipMap, nil
	}
	blog.Errorf("describe instances by ips with request %s timeout", request.ToJsonString())
	return nil, fmt.Errorf("describe instances by ips with request %s timeout", request.ToJsonString())
}

// registerBackends register backendList to rules
// when backend type is CVM: use instance id to register
// when backend type is eni: use eni ip to register
func (c *Client) registerBackends(lbID, listenerID, ruleID string,
	backendsRegister cloudListenerType.BackendList) error {
	request := tclb.NewRegisterTargetsRequest()
	request.LoadBalancerId = tcommon.StringPtr(lbID)
	request.ListenerId = tcommon.StringPtr(listenerID)
	if len(ruleID) != 0 {
		request.LocationId = tcommon.StringPtr(ruleID)
	}
	if len(backendsRegister) == 0 {
		blog.Infof("lb %s, listener %s, rule %s has no backend, no need to register", lbID, listenerID, ruleID)
		return nil
	}
	// cvm backend
	if c.sdkConfig.BackendType == ClbBackendTargetTypeCVM {
		var ips []string
		for _, backend := range backendsRegister {
			ips = append(ips, backend.IP)
		}
		ipMap, err := c.getCVMInstanceIDMapByIP(ips)
		if err != nil {
			return err
		}
		for _, backend := range backendsRegister {
			request.Targets = append(request.Targets, &tclb.Target{
				InstanceId: tcommon.StringPtr(ipMap[backend.IP]),
				Port:       tcommon.Int64Ptr(int64(backend.Port)),
				Type:       tcommon.StringPtr(ClbBackendTargetTypeCVM),
				Weight:     tcommon.Int64Ptr(int64(backend.Weight)),
			})
		}
		// eni backend
	} else {
		for _, backend := range backendsRegister {
			request.Targets = append(request.Targets, &tclb.Target{
				EniIp:  tcommon.StringPtr(backend.IP),
				Port:   tcommon.Int64Ptr(int64(backend.Port)),
				Type:   tcommon.StringPtr(ClbBackendTargetTypeENI),
				Weight: tcommon.Int64Ptr(int64(backend.Weight)),
			})
		}
	}
	// do request
	blog.Infof("register backend request:\n%s", request.ToJsonString())
	counter := 0
	var err error
	var response *tclb.RegisterTargetsResponse
	for ; counter < c.sdkConfig.MaxTimeout; counter++ {
		response, err = c.clb.RegisterTargets(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				c.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("register backend failed, err %s", err.Error())
			return fmt.Errorf("register backend failed, err %s", err.Error())
		}
		blog.Infof("register backend response:\n%s", response.ToJsonString())
		break
	}
	if counter >= c.sdkConfig.MaxTimeout {
		blog.Errorf("register backend with request %s timeout", request.ToJsonString())
		return fmt.Errorf("register backend with request %s timeout", request.ToJsonString())
	}
	return c.waitTaskDone(*response.Response.RequestId)
}

// deRegisterBackends deregister backendList to rules
// when backend type is CVM: use instance id to deregister
// when backend type is eni: use eni ip to deregister
func (c *Client) deRegisterBackends(lbID, listenerID, ruleID string,
	backendsDeregister cloudListenerType.BackendList) error {
	request := tclb.NewDeregisterTargetsRequest()
	request.LoadBalancerId = tcommon.StringPtr(lbID)
	request.ListenerId = tcommon.StringPtr(listenerID)
	if len(ruleID) != 0 {
		request.LocationId = tcommon.StringPtr(ruleID)
	}
	// backend type is cvm
	if c.sdkConfig.BackendType == ClbBackendTargetTypeCVM {
		var ips []string
		for _, backend := range backendsDeregister {
			ips = append(ips, backend.IP)
		}
		// get cvm instance id from cvm ip
		ipMap, err := c.getCVMInstanceIDMapByIP(ips)
		if err != nil {
			return err
		}
		for _, backend := range backendsDeregister {
			request.Targets = append(request.Targets, &tclb.Target{
				InstanceId: tcommon.StringPtr(ipMap[backend.IP]),
				Port:       tcommon.Int64Ptr(int64(backend.Port)),
				Type:       tcommon.StringPtr(ClbBackendTargetTypeCVM),
			})
		}
		// backend type is eni
	} else {
		for _, backend := range backendsDeregister {
			request.Targets = append(request.Targets, &tclb.Target{
				EniIp: tcommon.StringPtr(backend.IP),
				Port:  tcommon.Int64Ptr(int64(backend.Port)),
				Type:  tcommon.StringPtr(ClbBackendTargetTypeENI),
			})
		}
	}
	// do request
	blog.Infof("de register backend request:\n%s", request.ToJsonString())
	counter := 0
	var err error
	var response *tclb.DeregisterTargetsResponse
	for ; counter < c.sdkConfig.MaxTimeout; counter++ {
		response, err = c.clb.DeregisterTargets(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				c.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("de register backend failed, err %s", err.Error())
			return fmt.Errorf("de register backend failed, err %s", err.Error())
		}
		blog.Infof("de register backend response:\n%s", response.ToJsonString())
		break
	}
	if counter >= c.sdkConfig.MaxTimeout {
		blog.Errorf("de register backend with request %s timeout", request.ToJsonString())
		return fmt.Errorf("de register backend with request %s timeout", request.ToJsonString())
	}
	return c.waitTaskDone(*response.Response.RequestId)
}

// Register7LayerBackends register 7 layer backend
func (c *Client) Register7LayerBackends(lbID, listenerID, ruleID string,
	backendsRegister cloudListenerType.BackendList) error {
	return c.registerBackends(lbID, listenerID, ruleID, backendsRegister)
}

// DeRegister7LayerBackends deregister 7 layer backend
func (c *Client) DeRegister7LayerBackends(lbID, listenerID, ruleID string,
	backendsDeRegister cloudListenerType.BackendList) error {
	return c.deRegisterBackends(lbID, listenerID, ruleID, backendsDeRegister)
}

// Register4LayerBackends 4 layer backend
func (c *Client) Register4LayerBackends(lbID, listenerID string, backendsRegister cloudListenerType.BackendList) error {
	return c.registerBackends(lbID, listenerID, "", backendsRegister)
}

// DeRegister4LayerBackends deregister 4 layer
func (c *Client) DeRegister4LayerBackends(lbID, listenerID string,
	backendsDeRegister cloudListenerType.BackendList) error {
	return c.deRegisterBackends(lbID, listenerID, "", backendsDeRegister)
}

// waitTaskDone wait asynchronous task done
func (c *Client) waitTaskDone(taskID string) error {
	blog.Infof("start waiting for task %s", taskID)
	request := tclb.NewDescribeTaskStatusRequest()
	request.TaskId = tcommon.StringPtr(taskID)
	blog.Infof("describe task status request:\n%s", request.ToJsonString())
	for counter := 0; counter < c.sdkConfig.MaxTimeout; counter++ {
		// it may exceed limit when describe task result
		response, err := c.clb.DescribeTaskStatus(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				c.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("describe task status failed, err %s", err.Error())
			return fmt.Errorf("describe task status failed, err %s", err.Error())
		}
		blog.Infof("describe task status response:\n%s", response.ToJsonString())
		// dealing
		if *response.Response.Status == TaskStatusDealing {
			blog.Infof("task %s is dealing", taskID)
			time.Sleep(time.Duration(c.sdkConfig.WaitPeriodLBDealing) * time.Second)
			continue
			// failed
		} else if *response.Response.Status == TaskStatusFailed {
			blog.Errorf("task %s is failed", taskID)
			return fmt.Errorf("task %s is failed", taskID)
			// succeed
		} else if *response.Response.Status == TaskStatusSucceed {
			blog.Infof("task %s is done", taskID)
			return nil
		}
		return fmt.Errorf("error status of task %d", *response.Response.Status)
	}
	blog.Errorf("describe task status with request %s timeout", request.ToJsonString())
	return fmt.Errorf("describe task status with request %s timeout", request.ToJsonString())
}

// ListListener listener listeners in tencent cloud
// 1. list listener infos
// 2. list backends
// 3. list listener health status info
func (c *Client) ListListener(lbID string) ([]*cloudListenerType.CloudListener, error) {
	tclbListeners, err := c.doListListenerWithoutBackends(lbID)
	if err != nil {
		return nil, err
	}
	tclbListenerBackendMap, err := c.doListBackends(lbID)
	if err != nil {
		return nil, err
	}
	tclbListenerHealthMap, err := c.doListenerHealthStatus(lbID)
	if err != nil {
		return nil, err
	}
	var retListenerList []*cloudListenerType.CloudListener
	for _, tlistener := range tclbListeners {
		cloudListener, err := c.convertTclbListenerToCloudListener(
			tlistener, tclbListenerBackendMap[*tlistener.ListenerId], tclbListenerHealthMap[*tlistener.ListenerId])
		if err != nil {
			return nil, err
		}
		retListenerList = append(retListenerList, cloudListener)
	}
	return retListenerList, nil
}

// mashal backends to string
func backendToJSONString(b *tclb.Backend) string {
	data, _ := json.Marshal(b)
	return string(data)
}

// convert tclb backend to cloud listener backend
func (c *Client) convertToCloudListenerBackend(backendList []*tclb.Backend) ([]*cloudListenerType.Backend, error) {
	var retBackends []*cloudListenerType.Backend
	for _, backend := range backendList {
		// TODO: is there backend with no private ip address?
		if len(backend.PrivateIpAddresses) == 0 {
			return nil, fmt.Errorf("invalid backend %v", backendToJSONString(backend))
		}
		retBackends = append(retBackends, &cloudListenerType.Backend{
			IP:     *backend.PrivateIpAddresses[0],
			Port:   int(*backend.Port),
			Weight: int(*backend.Weight),
		})
	}
	return retBackends, nil
}

// convert tclb health check to local type
func (c *Client) convertToCloudListenerHealthCheck(hc *tclb.HealthCheck) (
	*cloudListenerType.TargetGroupHealthCheck, error) {
	if hc == nil {
		return nil, fmt.Errorf("cannot covert empty health check struct")
	}
	retHealthCheck := cloudListenerType.NewTargetGroupHealthCheck()
	// string ptr may be nil for all field
	if hc.HealthSwitch != nil {
		retHealthCheck.Enabled = int(*hc.HealthSwitch)
	}
	if hc.IntervalTime != nil {
		retHealthCheck.IntervalTime = int(*hc.IntervalTime)
	}
	if hc.HealthNum != nil {
		retHealthCheck.HealthNum = int(*hc.HealthNum)
	}
	if hc.UnHealthNum != nil {
		retHealthCheck.UnHealthNum = int(*hc.UnHealthNum)
	}
	if hc.HttpCode != nil {
		retHealthCheck.HTTPCode = int(*hc.HttpCode)
	}
	if hc.HttpCheckPath != nil {
		retHealthCheck.HTTPCheckPath = string(*hc.HttpCheckPath)
	}
	return retHealthCheck, nil
}

// convert rules targets to local target group
func (c *Client) convertRuleTargetsToTargetGroup(tclbRule *tclb.RuleOutput, tclbRuleTargets *tclb.RuleTargets) (
	*cloudListenerType.Rule, error) {
	rule := cloudListenerType.NewRule(*tclbRule.Domain, *tclbRule.Url)
	rule.ID = *tclbRule.LocationId
	hc, err := c.convertToCloudListenerHealthCheck(tclbRule.HealthCheck)
	if err != nil {
		return nil, err
	}
	rule.TargetGroup.HealthCheck = hc
	rule.TargetGroup.LBPolicy = LBAlgorithmTypeSDK2BcsMap[*tclbRule.Scheduler]
	if tclbRule.SessionExpireTime != nil {
		rule.TargetGroup.SessionExpire = int(*tclbRule.SessionExpireTime)
	}
	if tclbRuleTargets != nil && len(tclbRuleTargets.Targets) != 0 {
		backends, err := c.convertToCloudListenerBackend(tclbRuleTargets.Targets)
		if err != nil {
			return nil, err
		}
		rule.TargetGroup.Backends = backends
	}
	return nil, nil
}

// convert tclb listener type to local cloud listener type
func (c *Client) convertTclbListenerToCloudListener(
	listener *tclb.Listener, listenerBackend *tclb.ListenerBackend, listenerHeath *tclb.ListenerHealth) (
	*cloudListenerType.CloudListener, error) {
	if listener == nil {
		return nil, fmt.Errorf("cannot convert empty tclb listener object to cloud listener")
	}
	cloudListener := &cloudListenerType.CloudListener{
		ObjectMeta: metav1.ObjectMeta{
			Name: *listener.ListenerName,
		},
		Spec: cloudListenerType.CloudListenerSpec{
			ListenerID: *listener.ListenerId,
			Protocol:   ProtocolSDK2BcsMap[*listener.Protocol],
			ListenPort: int(*listener.Port),
		},
	}
	// only https listener has tls config
	if *listener.Protocol == ListenerProtocolHTTPS {
		sslMode, _ := SSLModeSDK2BcsMap[*listener.Certificate.SSLMode]
		cloudListener.Spec.TLS = &cloudListenerType.CloudListenerTls{
			Mode:   sslMode,
			CertID: *listener.Certificate.CertId,
		}
		if listener.Certificate.CertCaId != nil {
			cloudListener.Spec.TLS.CertCaID = *listener.Certificate.CertCaId
		}
	}
	switch *listener.Protocol {
	// convert http https listener
	case ListenerProtocolHTTP, ListenerProtocolHTTPS:
		ruleBackendsMap := make(map[string]*tclb.RuleTargets)
		for _, ruleTarget := range listenerBackend.Rules {
			ruleBackendsMap[*ruleTarget.LocationId] = ruleTarget
		}
		var retRules []*cloudListenerType.Rule
		for _, tclbRule := range listener.Rules {
			// convertRuleTargetsToTargetGroup will deal with health check and backends info
			rule, err := c.convertRuleTargetsToTargetGroup(tclbRule, ruleBackendsMap[*tclbRule.LocationId])
			if err != nil {
				return nil, err
			}
			retRules = append(retRules, rule)
		}
	// convert tcp udp listener
	case ListenerProtocolTCP, ListenerProtocolUDP:
		cloudListener.Spec.TargetGroup = cloudListenerType.NewTargetGroup(
			"", "", SSLModeSDK2BcsMap[*listener.Protocol], int(*listener.Port))
		hc, err := c.convertToCloudListenerHealthCheck(listener.HealthCheck)
		if err != nil {
			return nil, err
		}
		cloudListener.Spec.TargetGroup.HealthCheck = hc
		cloudListener.Spec.TargetGroup.LBPolicy = LBAlgorithmTypeSDK2BcsMap[*listener.Scheduler]
		if listener.SessionExpireTime != nil {
			cloudListener.Spec.TargetGroup.SessionExpire = int(*listener.SessionExpireTime)
		}
		if len(listenerBackend.Targets) != 0 {
			backends, err := c.convertToCloudListenerBackend(listenerBackend.Targets)
			if err != nil {
				return nil, err
			}
			cloudListener.Spec.TargetGroup.Backends = backends
		}
	}
	// deal with listener health info
	healthStatus := &cloudListenerType.CloudListenerHealthStatus{}
	for _, ruleHealth := range listenerHeath.Rules {
		tmpRuleHealthStatus := &cloudListenerType.CloudListenerRuleHealthStatus{}
		// domain and url is nil for tcp and udp listener
		if ruleHealth.Domain != nil {
			tmpRuleHealthStatus.Domain = *ruleHealth.Domain
		}
		if ruleHealth.Url != nil {
			tmpRuleHealthStatus.URL = *ruleHealth.Url
		}
		if len(ruleHealth.Targets) != 0 {
			for _, target := range ruleHealth.Targets {
				tmpRuleHealthStatus.Backends = append(tmpRuleHealthStatus.Backends,
					&cloudListenerType.CloudListenerBackendHealthStatus{
						IP:                 *target.IP,
						Port:               int(*target.Port),
						HealthStatus:       *target.HealthStatus,
						HealthStatusDetail: *target.HealthStatusDetial,
						TargetID:           *target.TargetId,
					})
			}
		}
		healthStatus.RulesHealth = append(healthStatus.RulesHealth, tmpRuleHealthStatus)
	}
	cloudListener.Status.HealthStatus = healthStatus

	return cloudListener, nil
}

// call tencent sdk DescribeListeners
// there is no backends in response
func (c *Client) doListListenerWithoutBackends(lbID string) ([]*tclb.Listener, error) {
	request := tclb.NewDescribeListenersRequest()
	request.LoadBalancerId = tcommon.StringPtr(lbID)
	blog.Infof("describe listeners request:\n%s", request.ToJsonString())
	for counter := 0; counter < c.sdkConfig.MaxTimeout; counter++ {
		response, err := c.clb.DescribeListeners(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				c.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("describe listener failed, err %s", err.Error())
			return nil, fmt.Errorf("describe listener failed, err %s", err.Error())
		}
		blog.Infof("describe listener response:\n%s", response.ToJsonString())
		if len(response.Response.Listeners) == 0 {
			blog.Warnf("describe listeners return no listener")
			return nil, nil
		}
		return response.Response.Listeners, nil
	}
	return nil, fmt.Errorf("describe listeners timeout")
}

// call tencent sdk DescribeTargets
func (c *Client) doListBackends(lbID string) (map[string]*tclb.ListenerBackend, error) {
	request := tclb.NewDescribeTargetsRequest()
	request.LoadBalancerId = tcommon.StringPtr(lbID)
	blog.Infof("describe backends request:\n%s", request.ToJsonString())
	for counter := 0; counter < c.sdkConfig.MaxTimeout; counter++ {
		response, err := c.clb.DescribeTargets(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				c.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("describe backends failed, err %s", err.Error())
			return nil, fmt.Errorf("backends targets failed, err %s", err.Error())
		}
		blog.Infof("describe backends response:\n%s", response.ToJsonString())
		if len(response.Response.Listeners) == 0 {
			blog.Warnf("describe listener backends return no listener")
			return nil, nil
		}
		retMap := make(map[string]*tclb.ListenerBackend)
		for _, listenerBackend := range response.Response.Listeners {
			retMap[*listenerBackend.ListenerId] = listenerBackend
		}
		return retMap, nil
	}
	return nil, fmt.Errorf("describe listener backend timeout")
}

// get listener health status by loadbalance id
// return map[key]*tclb.ListenerHealth, key is listenerid
func (c *Client) doListenerHealthStatus(lbID string) (map[string]*tclb.ListenerHealth, error) {
	request := tclb.NewDescribeTargetHealthRequest()
	request.LoadBalancerIds = []*string{tcommon.StringPtr(lbID)}
	blog.Infof("DescribeTargetsHealth request:\n%s", request.ToJsonString())
	for counter := 0; counter < c.sdkConfig.MaxTimeout; counter++ {
		response, err := c.clb.DescribeTargetHealth(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				c.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("DescribeTargetsHealth failed, err %s", err.Error())
			return nil, fmt.Errorf("DescribeTargetsHealth failed, err %s", err.Error())
		}
		blog.Infof("DescribeTargetsHealth response:\n%s", response.ToJsonString())
		if len(response.Response.LoadBalancers) == 0 {
			blog.Warnf("DescribeTargetsHealth return no loadbalancerHealth")
			return nil, nil
		}
		// expect one loadbalance info
		if len(response.Response.LoadBalancers) != 1 {
			blog.Errorf("DescribeTargetsHealth return loadbalancerHealth array with %d element, more than 1",
				len(response.Response.LoadBalancers))
			return nil, fmt.Errorf("DescribeTargetsHealth return loadbalancerHealth array with %d element, more than 1",
				len(response.Response.LoadBalancers))
		}
		lbHealth := response.Response.LoadBalancers[0]
		if len(lbHealth.Listeners) == 0 {
			blog.Warnf("DescribeTargetsHealth return no listenerHealth")
			return nil, nil
		}
		listenerHealthMap := make(map[string]*tclb.ListenerHealth)
		for _, listenerHealth := range lbHealth.Listeners {
			listenerHealthMap[*listenerHealth.ListenerId] = listenerHealth
		}
		return listenerHealthMap, nil
	}
	return nil, fmt.Errorf("describe listener health timeout")
}
