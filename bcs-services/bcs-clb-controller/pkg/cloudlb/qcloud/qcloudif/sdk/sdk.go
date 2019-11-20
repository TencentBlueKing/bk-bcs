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
	"fmt"
	"time"

	"bk-bcs/bcs-common/common/blog"

	cloudListenerType "bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/network/v1"
	"bk-bcs/bcs-services/bcs-clb-controller/pkg/cloudlb/qcloud/qcloudif"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	tcommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	terrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	tprofile "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	tvpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

const (
	DescribeFilterNamePrivateIP = "private-ip-address"
	ClbBackendTargetTypeCVM     = "CVM"
	ClbBackendTargetTypeENI     = "ENI"
	TaskStatusDealing           = 2
	TaskStatusFailed            = 1
	TaskStatusSucceed           = 0
	ClbStatusCreating           = 0
	ClbStatusNormal             = 1
)

type SdkConfig struct {
	Region                string
	ProjectID             int
	SubnetID              string
	VpcID                 string
	SecretID              string
	SecretKey             string
	BackendType           string
	MaxTimeout            int
	WaitPeriodExceedLimit int
	WaitPeriodLBDealing   int
}

type Client struct {
	clb       *tclb.Client
	cvm       *tcvm.Client
	vpc       *tvpc.Client
	sdkConfig *SdkConfig
}

func NewClient(sc *SdkConfig) qcloudif.ClbAdapter {
	credential := tcommon.NewCredential(sc.SecretID, sc.SecretKey)
	profile := tprofile.NewClientProfile()
	clbClient := &tclb.Client{}
	clbClient.Init(sc.Region).
		WithCredential(credential).
		WithProfile(profile)
	cvmClient := &tcvm.Client{}
	cvmClient.Init(sc.Region).
		WithCredential(credential).
		WithProfile(profile)
	return &Client{
		sdkConfig: sc,
		clb:       clbClient,
		cvm:       cvmClient,
	}
}

func (c *Client) checkErrCode(err *terrors.TencentCloudSDKError) {
	if err.Code == "4400" {
		blog.Warnf("request exceed limit, have a rest for %d second", c.sdkConfig.WaitPeriodExceedLimit)
		time.Sleep(time.Duration(c.sdkConfig.WaitPeriodLBDealing) * time.Second)
	} else if err.Code == "4000" {
		blog.Warnf("clb is dealing another action, have a rest for %d second", c.sdkConfig.WaitPeriodLBDealing)
		time.Sleep(time.Duration(c.sdkConfig.WaitPeriodLBDealing) * time.Second)
	}
}

// CreateLoadBalance call sdk to create clb, return clb id
// TODO: deal with vips
func (c *Client) CreateLoadBalance(lb *cloudListenerType.CloudLoadBalancer) (lbID string, vips []string, err error) {
	request := tclb.NewCreateLoadBalancerRequest()
	request.Forward = tcommon.Int64Ptr(LoadBalancerForwardApplication)
	if lb.NetworkType == cloudListenerType.ClbNetworkTypePublic {
		request.LoadBalancerType = tcommon.StringPtr(LoadBalancerNetworkPublic)
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
	return response.Response.LoadBalancerSet[0], nil
}

// DescribeLoadBalance describe clb by name, return clb info, and return if it is existed
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
	counter := 0
	var response *tclb.CreateListenerResponse
	for ; counter < c.sdkConfig.MaxTimeout; counter++ {
		response, err = c.clb.CreateListener(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				c.checkErrCode(terr)
				if terr.Code == "4400" || terr.Code == "4000" {
					continue
				}
			}
			blog.Errorf("create listener failed with request %s, err %s", request.ToJsonString(), err.Error())
			return "", fmt.Errorf("create listener failed with request %s, err %s", request.ToJsonString(), err.Error())
		}
		if len(response.Response.ListenerIds) == 0 {
			blog.Errorf("create listener return zero length ids with request %s, err %s", request.ToJsonString(), err.Error())
			return "", fmt.Errorf("create listener return zero length ids with request %s, err %s", request.ToJsonString(), err.Error())
		}
		blog.Infof("create listener response:\n%s", response.ToJsonString())
		break
	}
	if counter >= c.sdkConfig.MaxTimeout {
		blog.Errorf("create listener with request %s timeout", request.ToJsonString())
		return "", fmt.Errorf("create listener with request %s timeout", request.ToJsonString())
	}
	err = c.waitTaskDone(*response.Response.RequestId)
	if err != nil {
		return "", err
	}
	return *response.Response.ListenerIds[0], nil
}

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
	return request, nil
}

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
			request.HealthCheck.IntervalTime = tcommon.Int64Ptr(int64(listener.Spec.TargetGroup.HealthCheck.IntervalTime))
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
				if terr.Code == "4400" || terr.Code == "4000" {
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
func (c *Client) DescribeListener(lbID, listenerID string, port int) (listener *cloudListenerType.CloudListener, isExisted bool, err error) {
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

	blog.Infof("describe listener request:\n%s", request.ToJsonString())
	for counter := 0; counter < c.sdkConfig.MaxTimeout; counter++ {
		response, err := c.clb.DescribeListeners(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				c.checkErrCode(terr)
				if terr.Code == "4400" || terr.Code == "4000" {
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
			return nil, false, fmt.Errorf("describe response invalid listeners length %d", len(response.Response.Listeners))
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
		if *listener.Protocol == ListenerProtocolTCP || *listener.Protocol == ListenerProtocolUDP {
			lbPolicy := LBAlgorithmRoundRobin
			if validPolicy, ok := LBAlgorithmTypeSDK2BcsMap[*listener.Scheduler]; ok {
				lbPolicy = validPolicy
			}
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
				if terr.Code == "4400" || terr.Code == "4000" {
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

func (c *Client) modify4LayerListenerAttribute(listener *cloudListenerType.CloudListener) error {
	request := tclb.NewModifyListenerRequest()
	request.ListenerId = tcommon.StringPtr(listener.Spec.ListenerID)
	request.LoadBalancerId = tcommon.StringPtr(listener.Spec.LoadBalancerID)
	if listener.Spec.TargetGroup == nil {
		return fmt.Errorf("target group for 4 layer listener cannot be emtpy, error listener %v", listener)
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
			request.HealthCheck.IntervalTime = tcommon.Int64Ptr(int64(listener.Spec.TargetGroup.HealthCheck.IntervalTime))
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
				if terr.Code == "4400" || terr.Code == "4000" {
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
					ruleInput.HealthCheck.UnHealthNum = tcommon.Int64Ptr(int64(rule.TargetGroup.HealthCheck.UnHealthNum))
					ruleInput.HealthCheck.IntervalTime = tcommon.Int64Ptr(int64(rule.TargetGroup.HealthCheck.IntervalTime))
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
				if terr.Code == "4400" || terr.Code == "4000" {
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

// DeleteRule delete rule of clb listener by domain and url
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
				if terr.Code == "4400" || terr.Code == "4000" {
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
func (c *Client) DescribeRuleByDomainAndURL(loadBalanceID, listenerID, Domain, URL string) (rule *cloudListenerType.Rule, isExisted bool, err error) {
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
				if terr.Code == "4400" || terr.Code == "4000" {
					continue
				}
			}
			blog.Errorf("describe listener failed, err %s", err.Error())
			return nil, false, fmt.Errorf("describe listener failed, err %s", err.Error())
		}
		blog.Infof("describe listener response:\n%s", response.ToJsonString())
		if len(response.Response.Listeners) != 1 {
			blog.Errorf("describe response invalid listeners length %d", len(response.Response.Listeners))
			return nil, false, fmt.Errorf("describe response invalid listeners length %d", len(response.Response.Listeners))
		}
		listener := response.Response.Listeners[0]
		for _, ruleOutput := range listener.Rules {
			if *ruleOutput.Domain == Domain && *ruleOutput.Url == URL {
				retRule := &cloudListenerType.Rule{
					ID:     *ruleOutput.LocationId,
					Domain: Domain,
					URL:    URL,
				}
				lbPolicy := LBAlgorithmRoundRobin
				if validPolicy, ok := LBAlgorithmTypeSDK2BcsMap[*ruleOutput.Scheduler]; ok {
					lbPolicy = validPolicy
				}
				retRule.TargetGroup = &cloudListenerType.TargetGroup{
					SessionExpire: int(*ruleOutput.SessionExpireTime),
					LBPolicy:      lbPolicy,
				}
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
		blog.Infof("rule %s %s no found with %s %s", Domain, URL, listenerID, loadBalanceID)
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

	blog.Infof("modify rule with %v", request.ToJsonString())
	counter := 0
	var err error
	var response *tclb.ModifyRuleResponse
	for ; counter < c.sdkConfig.MaxTimeout; counter++ {
		response, err = c.clb.ModifyRule(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				c.checkErrCode(terr)
				if terr.Code == "4400" || terr.Code == "4000" {
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

func inStringPtrSlice(key string, array []*string) bool {
	for _, e := range array {
		if key == *e {
			return true
		}
	}
	return false
}

func (c *Client) getCVMInstanceIDMapByIP(ips []string) (map[string]string, error) {
	request := tcvm.NewDescribeInstancesRequest()
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

	for counter := 0; counter < c.sdkConfig.MaxTimeout; counter++ {
		response, err := c.cvm.DescribeInstances(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				c.checkErrCode(terr)
				if terr.Code == "4400" || terr.Code == "4000" {
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

func (c *Client) registerBackends(lbID, listenerID, ruleID string, backendsRegister cloudListenerType.BackendList) error {
	request := tclb.NewRegisterTargetsRequest()
	request.LoadBalancerId = tcommon.StringPtr(lbID)
	request.ListenerId = tcommon.StringPtr(listenerID)
	if len(ruleID) != 0 {
		request.LocationId = tcommon.StringPtr(ruleID)
	}
	if len(backendsRegister) == 0 {
		blog.Infof("lb %s, listener %s, rule %s has no backend, no need to register")
		return nil
	}
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

	blog.Infof("register backend request:\n%s", request.ToJsonString())
	counter := 0
	var err error
	var response *tclb.RegisterTargetsResponse
	for ; counter < c.sdkConfig.MaxTimeout; counter++ {
		response, err = c.clb.RegisterTargets(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				c.checkErrCode(terr)
				if terr.Code == "4400" || terr.Code == "4000" {
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

func (c *Client) deRegisterBackends(lbID, listenerID, ruleID string, backendsDeregister cloudListenerType.BackendList) error {
	request := tclb.NewDeregisterTargetsRequest()
	request.LoadBalancerId = tcommon.StringPtr(lbID)
	request.ListenerId = tcommon.StringPtr(listenerID)
	if len(ruleID) != 0 {
		request.LocationId = tcommon.StringPtr(ruleID)
	}
	if c.sdkConfig.BackendType == ClbBackendTargetTypeCVM {
		var ips []string
		for _, backend := range backendsDeregister {
			ips = append(ips, backend.IP)
		}
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
	} else {
		for _, backend := range backendsDeregister {
			request.Targets = append(request.Targets, &tclb.Target{
				EniIp: tcommon.StringPtr(backend.IP),
				Port:  tcommon.Int64Ptr(int64(backend.Port)),
				Type:  tcommon.StringPtr(ClbBackendTargetTypeENI),
			})
		}
	}
	blog.Infof("de register backend request:\n%s", request.ToJsonString())
	counter := 0
	var err error
	var response *tclb.DeregisterTargetsResponse
	for ; counter < c.sdkConfig.MaxTimeout; counter++ {
		response, err = c.clb.DeregisterTargets(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				c.checkErrCode(terr)
				if terr.Code == "4400" || terr.Code == "4000" {
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
func (c *Client) Register7LayerBackends(lbID, listenerID, ruleID string, backendsRegister cloudListenerType.BackendList) error {
	return c.registerBackends(lbID, listenerID, ruleID, backendsRegister)
}

// DeRegister7LayerBackends deregister 7 layer backend
func (c *Client) DeRegister7LayerBackends(lbID, listenerID, ruleID string, backendsDeRegister cloudListenerType.BackendList) error {
	return c.deRegisterBackends(lbID, listenerID, ruleID, backendsDeRegister)
}

// Register4LayerBackends 4 layer backend
func (c *Client) Register4LayerBackends(lbID, listenerID string, backendsRegister cloudListenerType.BackendList) error {
	return c.registerBackends(lbID, listenerID, "", backendsRegister)
}

// DeRegister4LayerBackends deregister 4 layer
func (c *Client) DeRegister4LayerBackends(lbID, listenerID string, backendsDeRegister cloudListenerType.BackendList) error {
	return c.deRegisterBackends(lbID, listenerID, "", backendsDeRegister)
}

func (c *Client) waitTaskDone(taskID string) error {
	blog.Infof("start waiting for task %s", taskID)
	request := tclb.NewDescribeTaskStatusRequest()
	request.TaskId = tcommon.StringPtr(taskID)
	blog.Infof("describe task status request:\n%s", request.ToJsonString())
	for counter := 0; counter < c.sdkConfig.MaxTimeout; counter++ {
		response, err := c.clb.DescribeTaskStatus(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				c.checkErrCode(terr)
				if terr.Code == "4400" || terr.Code == "4000" {
					continue
				}
			}
			blog.Errorf("describe task status failed, err %s", err.Error())
			return fmt.Errorf("describe task status failed, err %s", err.Error())
		}
		blog.Infof("describe task status response:\n%s", response.ToJsonString())
		if *response.Response.Status == TaskStatusDealing {
			blog.Infof("task %s is dealing", taskID)
			time.Sleep(time.Duration(c.sdkConfig.WaitPeriodLBDealing) * time.Second)
			continue
		} else if *response.Response.Status == TaskStatusFailed {
			blog.Errorf("task %s is failed", taskID)
			return fmt.Errorf("task %s is failed", taskID)
		} else if *response.Response.Status == TaskStatusSucceed {
			blog.Infof("task %s is done", taskID)
			return nil
		}
		return fmt.Errorf("error status of task %d", *response.Response.Status)
	}
	blog.Errorf("describe task status with request %s timeout", request.ToJsonString())
	return fmt.Errorf("describe task status with request %s timeout", request.ToJsonString())
}
