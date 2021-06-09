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

package api

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	qcloud "github.com/Tencent/bk-bcs/bcs-common/pkg/qcloud/clbv2"
	loadbalance "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/apis/network/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/cloudlb/qcloud/qcloudif"
)

//ClbAPI Clb api operator
type ClbAPI struct {
	// api for tencent cloud clb v2 api
	api qcloud.APIInterface
	// project id for tencent cloud account
	ProjectID int
	// region for clb
	Region string
	// subnet id for private clb instance
	SubnetID string
	// vpc id for clb instance
	VpcID string
	// secret id for tencent cloud account
	SecretID string
	// secret key for tencent cloud account
	SecretKey string
	// backend type: CVM or eni
	BackendType string
	// wait second for next query when task is dealing
	WaitPeriodLBDealing int
	// wait second for next query when exceeding api limit
	WaitPeriodExceedLimit    int
	ExpireTimeForHTTPSession int
}

//NewCloudClbAPI new clb api operator
func NewCloudClbAPI(projectID int, region, subnet, vpcID, secretID, secretKey, backendType string,
	waitPeriodLBDealing, waitPeriodExceedLimit, expireTime int) qcloudif.ClbAdapter {
	lbClient := qcloud.NewClient(QCloudLBURL, secretKey)
	cvmClient := qcloud.NewClient(QCloudCVMURLV3, secretKey)

	return &ClbAPI{
		ProjectID:                projectID,
		Region:                   region,
		VpcID:                    vpcID,
		SubnetID:                 subnet,
		SecretID:                 secretID,
		SecretKey:                secretKey,
		BackendType:              backendType,
		WaitPeriodLBDealing:      waitPeriodLBDealing,
		WaitPeriodExceedLimit:    waitPeriodExceedLimit,
		ExpireTimeForHTTPSession: expireTime,
		api:                      qcloud.NewAPI(lbClient, cvmClient),
	}
}

// CreateLoadBalance create clb by incoming cloud loadbalance info
// return clb instance id, vips and err
func (clb *ClbAPI) CreateLoadBalance(lb *loadbalance.CloudLoadBalancer) (string, []string, error) {
	networkType, ok := NetworkTypeBcs2QCloudMap[lb.NetworkType]
	if !ok {
		return "", nil, fmt.Errorf("unknown bcs network type %s", lb.NetworkType)
	}
	input := new(qcloud.CreateLBInput)
	input.Action = "CreateLoadBalancer"
	input.Nonce = uint(rand.Uint32())
	input.LoadBalancerName = lb.Name
	input.Region = clb.Region
	input.SecretID = clb.SecretID
	input.Timestamp = uint(time.Now().Unix())
	input.Forward = ClbApplicationType
	input.LoadBalancerType = networkType
	input.ProjectID = clb.ProjectID
	if networkType == ClbPrivate {
		input.SubnetID = clb.SubnetID
	}
	input.VpcID = clb.VpcID
	output, err := clb.api.CreateLoadBalance(input)
	if err != nil {
		return "", nil, fmt.Errorf("create loadbalance with input %v failed, err %s", input, err.Error())
	}

	if output.Code != 0 {
		blog.Errorf("create clb failed, response code %d, code desc %s, message %s",
			output.Code, output.CodeDesc, output.Message)
		return "", nil, fmt.Errorf("create clb failed, response code %d, code desc %s, message %s",
			output.Code, output.CodeDesc, output.Message)
	}
	blog.Infof("create clb request done, wait asynchronous task done")
	for _, id := range output.UnLoadBalancerIds {
		if len(id) != 1 {
			blog.Errorf("create clb result invalid, returned length of %s error, %v", id, output)
			return "", nil, fmt.Errorf("create clb result invalid, returned length of %s error, %v", id, output)
		}
		err := clb.waitForTaskResult(output.RequestID)
		if err != nil {
			return "", nil, fmt.Errorf("wait for task result %d failed, err %s", output.RequestID, err.Error())
		}
		count := 0
		for ; count <= ClbMaxTimeout; count++ {
			if count == ClbMaxTimeout {
				blog.Errorf("describe clb %s timeout after creation", lb.Name)
				return "", nil, fmt.Errorf("describe clb %s timeout after creation", lb.Name)
			}
			descResp, err := clb.doDescribeLoadBalance(lb.Name)
			if err != nil {
				blog.Errorf("describe clb %s failed after creation, err %s", lb.Name, err.Error())
				return "", nil, fmt.Errorf("describe clb %s failed after creation, err %s", lb.Name, err.Error())
			}
			if len(descResp.LoadBalances) != 1 {
				blog.Errorf("describe clb returned invalid length, result %v", descResp)
				return "", nil, fmt.Errorf("describe clb returned invalid length, result %v", descResp)
			}
			returnedLBOutput := descResp.LoadBalances[0]
			if returnedLBOutput.Status == ClbInstanceCreatingStatus {
				blog.Infof("clb instance %s is creating, describe it later", lb.Name)
				time.Sleep(time.Duration(clb.WaitPeriodLBDealing) * time.Second)
			} else if returnedLBOutput.Status == ClbInstanceRunningStatus {
				blog.Infof("describe clb %s successfully", lb.Name)
				return returnedLBOutput.UnLoadBalancerID, returnedLBOutput.LoadBalancerVips, nil
			} else {
				blog.Errorf("unknown clb instance status %d", returnedLBOutput.Status)
				return "", nil, fmt.Errorf("unknown clb instance status %d", returnedLBOutput.Status)
			}
		}
	}

	blog.Errorf("empty ids map in create clb output %v", output)
	return "", nil, fmt.Errorf("empty ids map in create clb output %v", output)
}

//DescribeLoadBalance describe clb by name
func (clb *ClbAPI) DescribeLoadBalance(name string) (*loadbalance.CloudLoadBalancer, bool, error) {
	output, err := clb.doDescribeLoadBalance(name)
	if err != nil {
		return nil, false, fmt.Errorf("describe clb failed, err %s", err.Error())
	}
	if output.Code == 5000 {
		blog.Warnf("clb %s is not existed", name)
		return nil, false, nil
	}
	if len(output.LoadBalances) == 0 {
		blog.Warnf("describe clb %s info return nil", name)
		return nil, false, nil
	}
	networkType, ok := NetworkTypeQCloud2BcsMap[output.LoadBalances[0].LoadBalancerType]
	if !ok {
		return nil, false, fmt.Errorf("convert qcloud network type %d to bcs type failed, err %s",
			output.LoadBalances[0].LoadBalancerType, err.Error())
	}

	if output.LoadBalances[0].Forward != ClbApplicationType {
		return nil, false, fmt.Errorf("found lb with name %s but not Application Type", name)
	}

	vips := make([]string, 0)
	for _, ip := range output.LoadBalances[0].LoadBalancerVips {
		vips = append(vips, ip)
	}
	cloudLBInfo := &loadbalance.CloudLoadBalancer{
		Name:        output.LoadBalances[0].LoadBalancerName,
		ID:          output.LoadBalances[0].LoadBalancerID,
		NetworkType: networkType,
		VIPS:        vips,
	}

	return cloudLBInfo, true, nil
}

//CreateListener create listener
func (clb *ClbAPI) CreateListener(listener *loadbalance.CloudListener) (string, error) {
	protocol, ok := ProtocolTypeBcs2QCloudMap[listener.Spec.Protocol]
	if !ok {
		return "", fmt.Errorf("convert bcs protocol %s to qcloud type failed", listener.Spec.Protocol)
	}
	var listenerID string
	var err error
	// https and http
	if protocol == ClbListenerProtocolHTTP || protocol == ClbListenerProtocolHTTPS {
		listenerID, err = clb.create7LayerListener(listener)
		if err != nil {
			return "", fmt.Errorf("create 7 layer listener failed, err %s", err.Error())
		}
		// tcp and udp
	} else {
		listenerID, err = clb.create4LayerListener(listener)
		if err != nil {
			return "", fmt.Errorf("create 4 layer listener failed, err %s", err.Error())
		}
	}

	blog.Infof("create listener %s successfully, id %s", listener.GetName(), listenerID)
	return listenerID, nil
}

//DeleteListener delete listener
func (clb *ClbAPI) DeleteListener(lbID, listenerID string) error {
	return clb.doDeleteListener(lbID, listenerID)
}

//ModifyListenerAttribute modify listener attribute
func (clb *ClbAPI) ModifyListenerAttribute(listener *loadbalance.CloudListener) error {
	protocol, ok := ProtocolTypeBcs2QCloudMap[listener.Spec.Protocol]
	if !ok {
		return fmt.Errorf("convert bcs protocol %s to qcloud type failed", listener.Spec.Protocol)
	}
	// tls config modification is only available for https listener
	if protocol == ClbListenerProtocolHTTPS {
		if listener.Spec.TLS == nil {
			return fmt.Errorf("https with nil tls config, error https listener cannot be modified")
		}
		return clb.doModify7LayerListenerAttribute(listener)
	} else if protocol == ClbListenerProtocolHTTP {
		return fmt.Errorf("listener with http protocol cannot modify ssl attribute")
	}
	if listener.Spec.TargetGroup == nil {
		return fmt.Errorf("listener spec.targetgroup is nil")
	}
	if listener.Spec.TargetGroup.HealthCheck == nil {
		return fmt.Errorf("listener spec.targetgroup.healthcheck is nil")
	}
	return clb.doModify4LayerListenerAttribute(listener)
}

// DescribeListener describe listener
// response does not contains backends info
func (clb *ClbAPI) DescribeListener(lbID, listenerID string, port int) (*loadbalance.CloudListener, bool, error) {
	var listenerInfo *qcloud.ListenerInfo
	var err error
	if len(listenerID) != 0 {
		listenerInfo, err = clb.doDescribeListener(lbID, listenerID)
	} else if port > 0 {
		listenerInfo, err = clb.doDescribeListenerByPort(lbID, port)
	} else {
		blog.Errorf("describe listener need id or port")
		return nil, false, fmt.Errorf("describe listener need id or port")
	}
	if listenerInfo == nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("describe listener %s of lb %s", listenerID, lbID)
	}
	qcloudProtocol, ok := ProtocolTypeQCloud2BcsMap[listenerInfo.Protocol]
	if !ok {
		return nil, false, fmt.Errorf("convert qcloud protocol type %d to bcs type failed", listenerInfo.Protocol)
	}
	listener := &loadbalance.CloudListener{
		Spec: loadbalance.CloudListenerSpec{
			ListenerID:     listenerInfo.ListenerID,
			LoadBalancerID: lbID,
			Protocol:       qcloudProtocol,
			ListenPort:     int(listenerInfo.LoadBalancerPort),
			Rules:          make([]*loadbalance.Rule, 0),
		},
	}
	if listener.Spec.Protocol == loadbalance.ClbListenerProtocolTCP ||
		listener.Spec.Protocol == loadbalance.ClbListenerProtocolUDP {
		return listener, true, nil
	}
	if listener.Spec.Protocol == loadbalance.ClbListenerProtocolHTTPS {
		listener.Spec.TLS = &loadbalance.CloudListenerTls{
			Mode:     listenerInfo.SSLMode,
			CertID:   listenerInfo.CertID,
			CertCaID: listenerInfo.CertCaID,
		}
	}
	for _, rule := range listenerInfo.Rules {
		newRule := &loadbalance.Rule{
			ID:     rule.LocationID,
			Domain: rule.Domain,
			URL:    rule.URL,
			TargetGroup: &loadbalance.TargetGroup{
				SessionExpire: rule.SessionExpire,
				LBPolicy:      rule.HTTPHash,
				HealthCheck: &loadbalance.TargetGroupHealthCheck{
					Enabled:       rule.HealthSwitch,
					IntervalTime:  rule.IntervalTime,
					HealthNum:     rule.HealthNum,
					UnHealthNum:   rule.UnhealthNum,
					HTTPCode:      rule.HTTPCode,
					HTTPCheckPath: rule.HTTPCheckPath,
				},
			},
		}
		listener.Spec.Rules = append(listener.Spec.Rules, newRule)
	}

	return listener, true, nil
}

//CreateRules create rules
func (clb *ClbAPI) CreateRules(lbID, listenerID string, rules loadbalance.RuleList) error {
	return clb.doCreateRules(lbID, listenerID, rules)
}

//DescribeRuleByDomainAndURL query rule by domain and url
func (clb *ClbAPI) DescribeRuleByDomainAndURL(loadBalanceID, listenerID, Domain, URL string) (
	*loadbalance.Rule, bool, error) {
	listenerInfo, err := clb.doDescribeListener(loadBalanceID, listenerID)
	if err != nil {
		return nil, false, fmt.Errorf("describe listener %s of lb %s", listenerID, loadBalanceID)
	}

	if listenerInfo.Protocol != ClbListenerProtocolHTTP && listenerInfo.Protocol != ClbListenerProtocolHTTPS {
		return nil, false, fmt.Errorf("listener %s of lb %s is not 7 layer listener", listenerID, loadBalanceID)
	}

	for _, rule := range listenerInfo.Rules {
		if rule.Domain == Domain && rule.URL == URL {
			return &loadbalance.Rule{
				ID:     rule.LocationID,
				Domain: rule.Domain,
				URL:    rule.URL,
				TargetGroup: &loadbalance.TargetGroup{
					HealthCheck: &loadbalance.TargetGroupHealthCheck{
						Enabled:       rule.HealthSwitch,
						IntervalTime:  rule.IntervalTime,
						HealthNum:     rule.HealthNum,
						UnHealthNum:   rule.UnhealthNum,
						HTTPCode:      rule.HTTPCode,
						HTTPCheckPath: rule.HTTPCheckPath,
					},
					SessionExpire: rule.SessionExpire,
					LBPolicy:      rule.HTTPHash,
				},
			}, true, nil
		}
	}
	return nil, false, nil
}

//DeleteRule delete rule by domain and url
func (clb *ClbAPI) DeleteRule(lbID, listenerID, domain, url string) error {

	validDomain := domain
	if strings.Contains(validDomain, ":") {
		validDomains := strings.Split(domain, ":")
		validDomain = validDomains[0]
	}
	return clb.doDeleteRule(lbID, listenerID, validDomain, url)
}

// ModifyRuleAttribute modify rule attribute
func (clb *ClbAPI) ModifyRuleAttribute(loadBalanceID, listenerID string, rule *loadbalance.Rule) error {
	return clb.doModifyRule(loadBalanceID, listenerID, rule)
}

// when backend mode is cvm, describe cvm instance id by ips,
// when backend mode is eni, use original ip
func (clb *ClbAPI) getBackends(backends loadbalance.BackendList) (qcloud.BackendTargetList, error) {
	if len(backends) == 0 {
		return nil, fmt.Errorf("no backends")
	}
	var bList qcloud.BackendTargetList
	if clb.BackendType == "cvm" {
		instanceIDs, err := clb.getCVMInstanceIDs(backends)
		if err != nil {
			return nil, fmt.Errorf("get cvm instance ids failed, err %s", err.Error())
		}
		for i, backend := range backends {
			newBack := qcloud.BackendTarget{
				BackendsInstanceID: instanceIDs[i],
				BackendsPort:       int(backend.Port),
				BackendsWeight:     backend.Weight,
			}
			bList = append(bList, newBack)
		}
	} else {
		for _, backend := range backends {
			newBack := qcloud.BackendTarget{
				BackendsIP:     backend.IP,
				BackendsPort:   int(backend.Port),
				BackendsWeight: backend.Weight,
			}
			bList = append(bList, newBack)
		}
	}
	return bList, nil
}

//Register7LayerBackends register 7 layer backends
func (clb *ClbAPI) Register7LayerBackends(
	lbID, listenerID, ruleID string, backendsRegister loadbalance.BackendList) error {
	bList, err := clb.getBackends(backendsRegister)
	if err != nil {
		return err
	}
	return clb.registerInsWith7thLayerListener(lbID, listenerID, ruleID, bList)
}

//_Register7LayerBackends register 7 layer backends
func (clb *ClbAPI) _Register7LayerBackends(
	lbID, listenerID, ruleID string, backendsRegister loadbalance.BackendList) error {
	if len(backendsRegister) == 0 {
		return fmt.Errorf("no backends in request")
	}
	instanceIDs, err := clb.getCVMInstanceIDs(backendsRegister)
	if err != nil {
		return fmt.Errorf("get cvm instance ids failed, err %s", err.Error())
	}
	var bList qcloud.BackendTargetList
	for i, backend := range backendsRegister {
		newBack := qcloud.BackendTarget{
			BackendsInstanceID: instanceIDs[i],
			BackendsPort:       int(backend.Port),
			BackendsWeight:     backend.Weight,
		}
		bList = append(bList, newBack)
	}

	return clb.registerInsWith7thLayerListener(lbID, listenerID, ruleID, bList)
}

// DeRegister7LayerBackends de register backends for 7 layer
func (clb *ClbAPI) DeRegister7LayerBackends(
	lbID, listenerID, ruleID string, backendsDeRegister loadbalance.BackendList) error {
	if len(backendsDeRegister) == 0 {
		blog.Infof("lb %s, listener %s, rule %s has no backend, no need to register", lbID, listenerID, ruleID)
		return nil
	}
	bList, err := clb.getBackends(backendsDeRegister)
	if err != nil {
		return err
	}
	return clb.deRegisterInstances7thListener(lbID, listenerID, ruleID, bList)
}

//Register4LayerBackends register 4 layer backends
func (clb *ClbAPI) Register4LayerBackends(lbID, listenerID string, backendsRegister loadbalance.BackendList) error {
	if len(backendsRegister) == 0 {
		blog.Infof("lb %s, listener %s has no backend, no need to register")
		return nil
	}
	bList, err := clb.getBackends(backendsRegister)
	if err != nil {
		return err
	}
	return clb.registerInsWith4thLayerListener(lbID, listenerID, bList)
}

//_Register4LayerBackends register 4 layer backends
// deprecated
func (clb *ClbAPI) _Register4LayerBackends(lbID, listenerID string, backendsRegister loadbalance.BackendList) error {
	if len(backendsRegister) == 0 {
		return fmt.Errorf("no backends in request")
	}
	instanceIDs, err := clb.getCVMInstanceIDs(backendsRegister)
	if err != nil {
		return fmt.Errorf("get cvm instance ids failed, err %s", err.Error())
	}
	var bList qcloud.BackendTargetList
	for i, backend := range backendsRegister {
		newBack := qcloud.BackendTarget{
			BackendsInstanceID: instanceIDs[i],
			BackendsPort:       int(backend.Port),
			BackendsWeight:     backend.Weight,
		}
		bList = append(bList, newBack)
	}
	return clb.registerInsWith4thLayerListener(lbID, listenerID, bList)
}

// DeRegister4LayerBackends de register backends for 4 layer
// deprecated
func (clb *ClbAPI) DeRegister4LayerBackends(lbID, listenerID string, backendsDeRegister loadbalance.BackendList) error {
	if len(backendsDeRegister) == 0 {
		return fmt.Errorf("zero length of backends to be deRegister")
	}
	bList, err := clb.getBackends(backendsDeRegister)
	if err != nil {
		return err
	}
	return clb.deRegisterInstances4thListener(lbID, listenerID, bList)
}

// ListListener list all listeners
func (clb *ClbAPI) ListListener(lbID string) ([]*loadbalance.CloudListener, error) {
	blog.Errorf("Not implemented")
	return nil, nil
}
