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
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/pkg/errors"
)

func (a *Alb) ensureLoadBalancerListener(region string, listener *networkextensionv1.Listener) (string, error) {
	if !isSamePort(listener.Spec.TargetGroup) {
		return "", multiplePortInOneTargetGroupError
	}
	// 1. ensure backend address pool
	if err := a.ensureAddrPoolForLB(listener); err != nil {
		return "", errors.Wrapf(err, "ensure lb address pool failed")
	}

	// 2. ensure listener
	if err := a.ensureLoadBalancer(region, listener); err != nil {
		return "", errors.Wrapf(err, "ensure lb failed")
	}

	return listener.Name, nil
}

func (a *Alb) ensureAddrPoolForLB(listener *networkextensionv1.Listener) error {
	lbName := listener.Spec.LoadbalancerID

	poolName := getLBRuleTgName(listener.Name, listener.Spec.Port)
	addrList := make([]*armnetwork.LoadBalancerBackendAddress, 0)

	if listener.Spec.TargetGroup != nil && len(listener.Spec.TargetGroup.Backends) != 0 {
		for _, backend := range listener.Spec.TargetGroup.Backends {
			addrList = append(addrList, &armnetwork.LoadBalancerBackendAddress{
				Name: to.StringPtr(fmt.Sprintf("%x", md5.Sum([]byte(backend.IP)))),
				Properties: &armnetwork.LoadBalancerBackendAddressPropertiesFormat{
					IPAddress:      to.StringPtr(backend.IP),
					VirtualNetwork: &armnetwork.SubResource{ID: to.StringPtr(a.sdkWrapper.buildVNetID())},
				},
			})
		}
	}

	_, err := a.sdkWrapper.CreateOrUpdateLoadBalanceBackendAddressPool(lbName, poolName, armnetwork.BackendAddressPool{
		Name: to.StringPtr(poolName),
		Properties: &armnetwork.BackendAddressPoolPropertiesFormat{
			LoadBalancerBackendAddresses: addrList,
		},
	})

	if err != nil {
		return err
	}
	return nil
}

func (a *Alb) ensureLoadBalancer(region string, listener *networkextensionv1.Listener) error {
	lbResp, err := a.sdkWrapper.GetLoadBalancer(region, listener.Spec.LoadbalancerID)
	if err != nil {
		return err
	}

	lb := &lbResp.LoadBalancer
	// 1. ensure probe
	lb = a.ensureProbesForLB(lb, listener)

	// 2. ensure loadBalancingRules
	lb, err = a.ensureLoadBalancingRule(lb, listener)
	if err != nil {
		return err
	}

	_, err = a.sdkWrapper.CreateOrUpdateLoadBalancer(listener.Spec.LoadbalancerID, *lb)
	if err != nil {
		return err
	}

	return nil
}

func (a *Alb) ensureProbesForLB(loadBalancer *armnetwork.LoadBalancer,
	listener *networkextensionv1.Listener) *armnetwork.LoadBalancer {
	newProbeList := make([]*armnetwork.Probe, 0)
	probeName := getLBRuleTgName(listener.Name, listener.Spec.Port)
	port := getBackendPort(listener.Spec.TargetGroup)

	newProbe := &armnetwork.Probe{
		Name: to.StringPtr(probeName),
		Properties: &armnetwork.ProbePropertiesFormat{
			Port:              to.Int32Ptr(port),
			Protocol:          transProbeProtocolPtr(listener.Spec.Protocol),
			IntervalInSeconds: to.Int32Ptr(DefaultLoadBalancerProbeInterval),
			NumberOfProbes:    to.Int32Ptr(DefaultLoadBalancerProbeNumberOfProbes),
		},
	}

	if listener.Spec.ListenerAttribute != nil && listener.Spec.ListenerAttribute.HealthCheck != nil && listener.
		Spec.ListenerAttribute.HealthCheck.Enabled == true {
		healthCheck := listener.Spec.ListenerAttribute.HealthCheck
		if healthCheck.IntervalTime != 0 {
			newProbe.Properties.IntervalInSeconds = to.Int32Ptr(int32(healthCheck.IntervalTime))
		}
		if healthCheck.HealthCheckProtocol != "" {
			newProbe.Properties.Protocol = transProbeProtocolPtr(healthCheck.HealthCheckProtocol)
		}
		if healthCheck.HealthCheckPort != 0 {
			newProbe.Properties.Port = to.Int32Ptr(int32(healthCheck.HealthCheckPort))
		}
	}

	newProbeList = append(newProbeList, newProbe)

	for _, probe := range loadBalancer.Properties.Probes {
		if strings.HasPrefix(*probe.Name, listener.Name) {
			continue
		}

		newProbeList = append(newProbeList, probe)
	}
	loadBalancer.Properties.Probes = newProbeList

	return loadBalancer
}

func (a *Alb) ensureLoadBalancingRule(loadBalancer *armnetwork.LoadBalancer,
	listener *networkextensionv1.Listener) (*armnetwork.LoadBalancer, error) {
	if len(loadBalancer.Properties.FrontendIPConfigurations) == 0 {
		return nil, unknownFrontIPConfiguration
	}
	// select frontendIP[0] as default
	frontendIPConfigurationID := loadBalancer.Properties.FrontendIPConfigurations[0].ID

	newRules := make([]*armnetwork.LoadBalancingRule, 0)
	ruleName := getLBRuleTgName(listener.Name, listener.Spec.Port)
	port := getBackendPort(listener.Spec.TargetGroup)

	newRule := &armnetwork.LoadBalancingRule{
		Name: to.StringPtr(ruleName),
		Properties: &armnetwork.LoadBalancingRulePropertiesFormat{
			FrontendPort: to.Int32Ptr(int32(listener.Spec.Port)),
			Protocol:     transTransportProtocolPtr(listener.Spec.Protocol),
			BackendAddressPool: a.resourceHelper.genSubResource(ResourceProviderLoadBalancer, listener.Spec.LoadbalancerID,
				ResourceTypeBackendAddressPools, ruleName),
			BackendAddressPools:     nil,
			BackendPort:             to.Int32Ptr(port),
			DisableOutboundSnat:     nil,
			EnableFloatingIP:        to.BoolPtr(false),
			EnableTCPReset:          nil,
			FrontendIPConfiguration: a.resourceHelper.getSubResourceByID(*frontendIPConfigurationID),
			IdleTimeoutInMinutes:    nil,
			LoadDistribution:        nil,
			Probe: a.resourceHelper.genSubResource(ResourceProviderLoadBalancer, listener.Spec.LoadbalancerID,
				ResourceTypeProbes, ruleName),
		},
	}

	if listener.Spec.ListenerAttribute != nil && listener.Spec.ListenerAttribute.SessionTime != 0 {
		sessionTime := listener.Spec.ListenerAttribute.SessionTime
		// sessionTime unit is seconds
		newRule.Properties.IdleTimeoutInMinutes = to.Int32Ptr(int32(sessionTime))
	}

	newRules = append(newRules, newRule)

	for _, rule := range loadBalancer.Properties.LoadBalancingRules {
		if strings.HasPrefix(*rule.Name, listener.Name) {
			continue
		}

		newRules = append(newRules, rule)
	}
	loadBalancer.Properties.LoadBalancingRules = newRules

	return loadBalancer, nil
}

func (a *Alb) deleteLoadBalancerListener(region string, listener *networkextensionv1.Listener) error {
	poolName := getLBRuleTgName(listener.Name, listener.Spec.Port)
	// 1. delete backend address pool
	if err := a.sdkWrapper.DeleteLoadBalanceAddressPool(listener.Spec.LoadbalancerID, poolName); err != nil {
		return err
	}

	lbResp, err := a.sdkWrapper.GetLoadBalancer(region, listener.Spec.LoadbalancerID)
	if err != nil {
		return err
	}

	lb := lbResp.LoadBalancer
	// 2. delete probe
	newProbes := make([]*armnetwork.Probe, 0)
	for _, probe := range lb.Properties.Probes {
		if probe.Name != nil && strings.HasPrefix(*probe.Name, listener.Name) {
			continue
		}

		newProbes = append(newProbes, probe)
	}
	lb.Properties.Probes = newProbes

	// 3. delete rule
	newRules := make([]*armnetwork.LoadBalancingRule, 0)
	for _, rule := range lb.Properties.LoadBalancingRules {
		if rule.Name != nil && strings.HasPrefix(*rule.Name, listener.Name) {
			continue
		}

		newRules = append(newRules, rule)
	}

	lb.Properties.LoadBalancingRules = newRules
	_, err = a.sdkWrapper.CreateOrUpdateLoadBalancer(listener.Spec.LoadbalancerID, lb)

	if err != nil {
		return err
	}

	return nil
}
func (a *Alb) ensureApplicationGatewayListener(region string, listener *networkextensionv1.Listener) (string, error) {

	if !isRuleSamePort(listener) {
		return "", multiplePortInOneTargetGroupError
	}

	// 1. get raw application gateway
	appGatewayRsp, err := a.sdkWrapper.GetApplicationGateway(region, listener.Spec.LoadbalancerID)
	if err != nil {
		return "", err
	}

	appGateway := &appGatewayRsp.ApplicationGateway

	// 2. ensure frontend port
	appGateway = a.ensureFrontendPortForAg(appGateway, listener)

	// 3. ensure addrPool
	appGateway = a.ensureAddrPoolForAg(appGateway, listener)

	// 4. ensure probes
	appGateway = a.ensureProbeForAg(appGateway, listener)

	// 5. backend settings
	appGateway = a.ensureBackendSettings(appGateway, listener)

	// 6. listener
	appGateway, err = a.ensureHttpListenerForAg(appGateway, listener)
	if err != nil {
		return "", err
	}

	// 7. URLPathMap
	appGateway = a.ensureUrlPathMap(appGateway, listener)

	// 7. request routing rule
	appGateway, err = a.ensureRequestRoutingRule(appGateway, listener)
	if err != nil {
		return "", err
	}

	// 8. update application gateway
	_, err = a.sdkWrapper.CreateOrUpdateApplicationGateway(listener.Spec.LoadbalancerID, *appGateway)
	if err != nil {
		return "", err
	}

	return listener.Name, nil
}

func (a *Alb) ensureFrontendPortForAg(appGateway *armnetwork.ApplicationGateway,
	listener *networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	listenPort := listener.Spec.Port

	for _, port := range appGateway.Properties.FrontendPorts {
		if port.Name != nil && *port.Name == fmt.Sprintf("port_%d", listenPort) {
			return appGateway
		}
	}

	appGateway.Properties.FrontendPorts = append(appGateway.Properties.FrontendPorts,
		&armnetwork.ApplicationGatewayFrontendPort{
			Name:       to.StringPtr(fmt.Sprintf("port_%d", listenPort)),
			Properties: &armnetwork.ApplicationGatewayFrontendPortPropertiesFormat{Port: to.Int32Ptr(int32(listenPort))},
		})

	return appGateway
}

func (a *Alb) ensureAddrPoolForAg(appGateway *armnetwork.ApplicationGateway,
	listener *networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	newPools := make([]*armnetwork.ApplicationGatewayBackendAddressPool, 0)
	for _, rule := range listener.Spec.Rules {
		addrList := make([]*armnetwork.ApplicationGatewayBackendAddress, 0)
		if rule.TargetGroup != nil {
			for _, backend := range rule.TargetGroup.Backends {
				addrList = append(addrList, &armnetwork.ApplicationGatewayBackendAddress{
					IPAddress: to.StringPtr(backend.IP),
				})
			}
		}

		poolName := getRuleTgName(listener.Name, rule.Domain, rule.Path, listener.Spec.Port)
		newPools = append(newPools, &armnetwork.ApplicationGatewayBackendAddressPool{
			Name: to.StringPtr(poolName),
			Properties: &armnetwork.ApplicationGatewayBackendAddressPoolPropertiesFormat{
				BackendAddresses: addrList,
			},
		})
	}

	newPools = append(newPools, &armnetwork.ApplicationGatewayBackendAddressPool{
		Name: to.StringPtr(DefaultBackendPoolName),
		Properties: &armnetwork.ApplicationGatewayBackendAddressPoolPropertiesFormat{
			BackendAddresses: make([]*armnetwork.ApplicationGatewayBackendAddress, 0),
		},
	})

	// exclude pool relates to current listener
	for _, pool := range appGateway.Properties.BackendAddressPools {
		if strings.HasPrefix(*pool.Name, listener.Name) || *pool.Name == DefaultBackendPoolName {
			continue
		}

		newPools = append(newPools, pool)
	}
	appGateway.Properties.BackendAddressPools = newPools

	return appGateway
}

func (a *Alb) ensureBackendSettings(appGateway *armnetwork.ApplicationGateway,
	listener *networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	newSettings := make([]*armnetwork.ApplicationGatewayBackendHTTPSettings, 0)
	for _, rule := range listener.Spec.Rules {
		settingName := getRuleTgName(listener.Name, rule.Domain, rule.Path, listener.Spec.Port)

		needProbe := false
		var probeResource *armnetwork.SubResource = nil
		if rule.ListenerAttribute != nil && rule.ListenerAttribute.HealthCheck != nil && rule.ListenerAttribute.
			HealthCheck.Enabled {
			needProbe = true
			probeResource = a.resourceHelper.genSubResource(ResourceProviderApplicationGateway,
				listener.Spec.LoadbalancerID, ResourceTypeProbes, settingName)
		}

		// if no backends, use default port and protocol
		port := 80
		protocol := AzureProtocolHTTP
		if rule.TargetGroup != nil && len(rule.TargetGroup.Backends) != 0 {
			port = rule.TargetGroup.Backends[0].Port
			protocol = rule.TargetGroup.TargetGroupProtocol
		}
		newSetting := &armnetwork.ApplicationGatewayBackendHTTPSettings{
			Name: to.StringPtr(settingName),
			Properties: &armnetwork.ApplicationGatewayBackendHTTPSettingsPropertiesFormat{
				AffinityCookieName:             nil,
				AuthenticationCertificates:     nil,
				ConnectionDraining:             nil,
				CookieBasedAffinity:            nil,
				HostName:                       nil,
				Path:                           nil,
				PickHostNameFromBackendAddress: to.BoolPtr(false),
				Port:                           to.Int32Ptr(int32(port)),
				Probe:                          probeResource,
				ProbeEnabled:                   &needProbe,
				Protocol:                       transAgProtocolPtr(protocol),
				RequestTimeout:                 to.Int32Ptr(DefaultRequestTimeout),
				TrustedRootCertificates:        nil,
				ProvisioningState:              nil,
			},
		}
		newSettings = append(newSettings, newSetting)
	}

	// add default settings
	newSettings = append(newSettings, &armnetwork.ApplicationGatewayBackendHTTPSettings{
		Name: to.StringPtr(DefaultBackendSettingName),
		Properties: &armnetwork.ApplicationGatewayBackendHTTPSettingsPropertiesFormat{
			AffinityCookieName:             nil,
			AuthenticationCertificates:     nil,
			ConnectionDraining:             nil,
			CookieBasedAffinity:            nil,
			HostName:                       nil,
			Path:                           nil,
			PickHostNameFromBackendAddress: nil,
			Port:                           to.Int32Ptr(80),
			Protocol:                       transAgProtocolPtr(string(armnetwork.ApplicationGatewayProtocolHTTP)),
			RequestTimeout:                 to.Int32Ptr(DefaultRequestTimeout),
			TrustedRootCertificates:        nil,
			ProvisioningState:              nil,
		},
	})

	// exclude settings relates to current listener
	for _, setting := range appGateway.Properties.BackendHTTPSettingsCollection {
		if strings.HasPrefix(*setting.Name, listener.Name) || *setting.Name == DefaultBackendSettingName {
			continue
		}

		newSettings = append(newSettings, setting)
	}
	appGateway.Properties.BackendHTTPSettingsCollection = newSettings

	return appGateway
}

// if no need probe, return false and nil subResource
func (a *Alb) ensureProbeForAg(appGateway *armnetwork.ApplicationGateway,
	listener *networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	newProbes := make([]*armnetwork.ApplicationGatewayProbe, 0)
	for _, rule := range listener.Spec.Rules {
		if rule.ListenerAttribute == nil || rule.ListenerAttribute.HealthCheck == nil || !rule.ListenerAttribute.
			HealthCheck.Enabled {
			continue
		}

		probeName := getRuleTgName(listener.Name, rule.Domain, rule.Path, listener.Spec.Port)

		healthCheck := rule.ListenerAttribute.HealthCheck
		var probeHost string
		if rule.Domain != "" {
			probeHost = rule.Domain
		} else {
			probeHost = "127.0.0.1"
		}

		newProbe := &armnetwork.ApplicationGatewayProbe{
			Name: to.StringPtr(probeName),
			Properties: &armnetwork.ApplicationGatewayProbePropertiesFormat{
				Host:                                to.StringPtr(probeHost),
				Interval:                            to.Int32Ptr(int32(DefaultLoadBalancerProbeInterval)),
				Match:                               transAgProbeMatch(healthCheck),
				Path:                                to.StringPtr(healthCheck.HTTPCheckPath),
				PickHostNameFromBackendHTTPSettings: to.BoolPtr(false),
				PickHostNameFromBackendSettings:     to.BoolPtr(false),
				Port:                                to.Int32Ptr(int32(healthCheck.HealthCheckPort)),
				Protocol:                            transAgProtocolPtr(healthCheck.HealthCheckProtocol),
				Timeout:                             to.Int32Ptr(int32(DefaultRequestTimeout)),
				UnhealthyThreshold:                  to.Int32Ptr(int32(healthCheck.UnHealthNum)),
			},
		}
		if healthCheck.HealthCheckPort == 0 {
			newProbe.Properties.Port = to.Int32Ptr(getBackendPort(rule.TargetGroup))
		}
		if healthCheck.HealthCheckProtocol == "" {
			if rule.TargetGroup == nil || len(rule.TargetGroup.Backends) == 0 {
				newProbe.Properties.Protocol = transAgProtocolPtr(AzureProtocolHTTP)
			} else {
				newProbe.Properties.Protocol = transAgProtocolPtr(rule.TargetGroup.TargetGroupProtocol)
			}
		}
		if healthCheck.HTTPCheckPath == "" {
			if rule.Path != "" {
				newProbe.Properties.Path = to.StringPtr(rule.Path)
			} else {
				newProbe.Properties.Path = to.StringPtr("/")
			}
		}
		if healthCheck.Timeout != 0 {
			newProbe.Properties.Timeout = to.Int32Ptr(int32(healthCheck.Timeout))
		}
		if healthCheck.IntervalTime != 0 {
			newProbe.Properties.Interval = to.Int32Ptr(int32(healthCheck.IntervalTime))
		}

		newProbes = append(newProbes, newProbe)
	}

	// exclude probe relates to current listener
	for _, probe := range appGateway.Properties.Probes {
		if strings.HasPrefix(*probe.Name, listener.Name) {
			continue
		}

		newProbes = append(newProbes, probe)
	}
	appGateway.Properties.Probes = newProbes

	return appGateway
}

func (a *Alb) ensureHttpListenerForAg(appGateway *armnetwork.ApplicationGateway,
	listener *networkextensionv1.Listener) (*armnetwork.ApplicationGateway, error) {
	if len(appGateway.Properties.FrontendIPConfigurations) != 1 {
		return nil, unknownFrontIPConfiguration
	}
	frontIPConfigurationID := appGateway.Properties.FrontendIPConfigurations[0].ID

	httpListenerMap := make(map[string]struct{})
	for _, httpListener := range appGateway.Properties.HTTPListeners {
		httpListenerMap[*httpListener.Name] = struct{}{}
	}
	for _, rule := range listener.Spec.Rules {
		httpListenerName := getHttpListenerName(listener.Spec.Port, rule.Domain)
		if _, ok := httpListenerMap[httpListenerName]; ok {
			// 不更新原有listener，避免影响原有业务
			continue
		}

		listenPort := listener.Spec.Port
		var hostNamePtr *string = nil
		if rule.Domain != "" {
			hostNamePtr = to.StringPtr(rule.Domain)
		}

		newHttpListener := &armnetwork.ApplicationGatewayHTTPListener{
			Name: to.StringPtr(httpListenerName),
			Properties: &armnetwork.ApplicationGatewayHTTPListenerPropertiesFormat{
				FrontendIPConfiguration: a.resourceHelper.getSubResourceByID(*frontIPConfigurationID),
				FrontendPort: a.resourceHelper.genSubResource(ResourceProviderApplicationGateway,
					listener.Spec.LoadbalancerID, ResourceTypeFrontendPorts, fmt.Sprintf("port_%d", listenPort)),
				HostName:                    hostNamePtr,
				Protocol:                    transAgProtocolPtr(listener.Spec.Protocol),
				RequireServerNameIndication: nil,
				SSLCertificate:              nil,
				SSLProfile:                  nil,
			},
		}
		if strings.ToLower(listener.Spec.Protocol) == "https" && listener.Spec.Certificate != nil {
			newHttpListener.Properties.SSLCertificate = a.resourceHelper.genSubResource(
				ResourceProviderApplicationGateway, listener.Spec.LoadbalancerID, ResourceTypeSSLCertificate,
				listener.Spec.Certificate.CertID)
		}

		appGateway.Properties.HTTPListeners = append(appGateway.Properties.HTTPListeners, newHttpListener)
		httpListenerMap[httpListenerName] = struct{}{}
	}

	return appGateway, nil
}

func (a *Alb) ensureRequestRoutingRule(appGateway *armnetwork.ApplicationGateway,
	listener *networkextensionv1.Listener) (*armnetwork.ApplicationGateway, error) {
	routingRuleMap := make(map[string]*armnetwork.ApplicationGatewayRequestRoutingRule)

	for _, routingRule := range appGateway.Properties.RequestRoutingRules {
		routingRuleMap[*routingRule.Name] = routingRule
	}

	for _, rule := range listener.Spec.Rules {
		httpListenerName := getHttpListenerName(listener.Spec.Port, rule.Domain)

		var pathMapResource *armnetwork.SubResource
		var ruleType armnetwork.ApplicationGatewayRequestRoutingRuleType
		if rule.Path == "" {
			ruleType = armnetwork.ApplicationGatewayRequestRoutingRuleTypeBasic
			pathMapResource = nil
		} else {
			ruleType = armnetwork.ApplicationGatewayRequestRoutingRuleTypePathBasedRouting
			pathMapResource = a.resourceHelper.genSubResource(ResourceProviderApplicationGateway,
				listener.Spec.LoadbalancerID, ResourceTypeURLPathMaps, httpListenerName)
		}

		if routingRule, ok := routingRuleMap[httpListenerName]; ok {
			if *routingRule.Properties.RuleType != ruleType {
				return nil, fmt.Errorf("conflict rule type in routingRule[%s], exists: %s, want: %s, "+
					"routingRule info :%s", httpListenerName, *routingRule.Properties.RuleType, ruleType,
					common.ToJsonString(routingRule))
			}
		}

		// Azure request all backends have same port
		ruleTgName := getRuleTgName(listener.Name, rule.Domain, rule.Path, listener.Spec.Port)
		priority := generatePriority(appGateway)

		newRoutingRule := &armnetwork.ApplicationGatewayRequestRoutingRule{
			Name: to.StringPtr(httpListenerName),
			Properties: &armnetwork.ApplicationGatewayRequestRoutingRulePropertiesFormat{
				BackendAddressPool: a.resourceHelper.genSubResource(ResourceProviderApplicationGateway,
					listener.Spec.LoadbalancerID, ResourceTypeBackendAddressPools, ruleTgName),
				BackendHTTPSettings: a.resourceHelper.genSubResource(ResourceProviderApplicationGateway,
					listener.Spec.LoadbalancerID, ResourceTypeBackendHttpSettingsCollection, ruleTgName),
				HTTPListener: a.resourceHelper.genSubResource(ResourceProviderApplicationGateway,
					listener.Spec.LoadbalancerID, ResourceTypeHttpListeners, httpListenerName),
				LoadDistributionPolicy: nil,
				Priority:               &priority,
				RuleType:               &ruleType,
				URLPathMap:             pathMapResource,
			},
		}

		routingRuleMap[httpListenerName] = newRoutingRule

		// add into rule list for build priority
		appGateway.Properties.RequestRoutingRules = append(appGateway.Properties.RequestRoutingRules, newRoutingRule)
	}

	appGateway.Properties.RequestRoutingRules = make([]*armnetwork.ApplicationGatewayRequestRoutingRule, 0)
	for _, routingRule := range routingRuleMap {
		appGateway.Properties.RequestRoutingRules = append(appGateway.Properties.RequestRoutingRules, routingRule)
	}

	return appGateway, nil
}

func (a *Alb) ensureUrlPathMap(appGateway *armnetwork.ApplicationGateway,
	listener *networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	urlPathMapMap := make(map[string]*armnetwork.ApplicationGatewayURLPathMap)
	for _, urlPathMap := range appGateway.Properties.URLPathMaps {
		urlPathMapMap[*urlPathMap.Name] = urlPathMap
	}

	for _, rule := range listener.Spec.Rules {
		if rule.Path == "" {
			continue
		}

		ruleTgName := getRuleTgName(listener.Name, rule.Domain, rule.Path, listener.Spec.Port)
		URLPathMapName := getHttpListenerName(listener.Spec.Port, rule.Domain)
		var urlPathMap *armnetwork.ApplicationGatewayURLPathMap
		if pathMap, ok := urlPathMapMap[URLPathMapName]; ok {
			urlPathMap = pathMap
		} else {
			urlPathMap = &armnetwork.ApplicationGatewayURLPathMap{
				Name: to.StringPtr(URLPathMapName),
				Properties: &armnetwork.ApplicationGatewayURLPathMapPropertiesFormat{
					DefaultBackendAddressPool: a.resourceHelper.genSubResource(ResourceProviderApplicationGateway,
						listener.Spec.LoadbalancerID, ResourceTypeBackendAddressPools, DefaultBackendPoolName),
					DefaultBackendHTTPSettings: a.resourceHelper.genSubResource(ResourceProviderApplicationGateway,
						listener.Spec.LoadbalancerID, ResourceTypeBackendHttpSettingsCollection,
						DefaultBackendSettingName),
					PathRules: make([]*armnetwork.ApplicationGatewayPathRule, 0),
				},
			}
		}

		pathRuleName := fmt.Sprintf("%x", md5.Sum([]byte(rule.Path)))
		redundant := false
		for _, pathRule := range urlPathMap.Properties.PathRules {
			if pathRule.Name != nil && *pathRule.Name == pathRuleName {
				// 不应该更新已有的pathRule
				redundant = true
				break
			}
		}
		if redundant == true {
			continue
		}

		urlPathMap.Properties.PathRules = append(urlPathMap.Properties.PathRules,
			&armnetwork.ApplicationGatewayPathRule{
				Name: to.StringPtr(pathRuleName),
				Properties: &armnetwork.ApplicationGatewayPathRulePropertiesFormat{
					BackendAddressPool: a.resourceHelper.genSubResource(ResourceProviderApplicationGateway,
						listener.Spec.LoadbalancerID, ResourceTypeBackendAddressPools, ruleTgName),
					BackendHTTPSettings: a.resourceHelper.genSubResource(ResourceProviderApplicationGateway,
						listener.Spec.LoadbalancerID, ResourceTypeBackendHttpSettingsCollection, ruleTgName),
					Paths: []*string{to.StringPtr(rule.Path)},
				},
			})

		urlPathMapMap[URLPathMapName] = urlPathMap
	}

	urlPathMapList := make([]*armnetwork.ApplicationGatewayURLPathMap, 0)
	for _, urlPathMap := range urlPathMapMap {
		urlPathMapList = append(urlPathMapList, urlPathMap)
	}
	appGateway.Properties.URLPathMaps = urlPathMapList
	return appGateway

}

func (a *Alb) deleteApplicationGatewayListener(region string, listener *networkextensionv1.Listener) error {
	appGatewayRsp, err := a.sdkWrapper.GetApplicationGateway(region, listener.Spec.LoadbalancerID)
	if err != nil {
		return err
	}

	appGateway := &appGatewayRsp.ApplicationGateway

	// addrPool
	appGateway = a.deleteAddrPoolForAg(appGateway, listener)

	// probes
	appGateway = a.deleteProbeForAg(appGateway, listener)

	// backend settings
	appGateway = a.deleteBackendSettingsForAg(appGateway, listener)

	// delete order : urlPathMap -> routingRule -> listener
	// URLPathMap
	appGateway = a.deleteURLPathMapForAg(appGateway, listener)

	// request routing rule
	appGateway = a.deleteRoutingRuleForAg(appGateway, listener)

	// listener
	appGateway = a.deleteHttpListenerForAg(appGateway, listener)

	_, err = a.sdkWrapper.CreateOrUpdateApplicationGateway(listener.Spec.LoadbalancerID, *appGateway)
	if err != nil {
		return err
	}
	return nil
}

// remove listener related backendAddressPool
func (a *Alb) deleteAddrPoolForAg(appGateway *armnetwork.ApplicationGateway,
	listener *networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	list := make([]*armnetwork.ApplicationGatewayBackendAddressPool, 0)
	for _, obj := range appGateway.Properties.BackendAddressPools {
		if obj.Name != nil && strings.HasPrefix(*obj.Name, listener.Name) {
			continue
		}
		list = append(list, obj)
	}
	appGateway.Properties.BackendAddressPools = list

	return appGateway
}

// remove listener related probe
func (a *Alb) deleteProbeForAg(appGateway *armnetwork.ApplicationGateway,
	listener *networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	list := make([]*armnetwork.ApplicationGatewayProbe, 0)
	for _, obj := range appGateway.Properties.Probes {
		if obj.Name != nil && strings.HasPrefix(*obj.Name, listener.Name) {
			continue
		}
		list = append(list, obj)
	}
	appGateway.Properties.Probes = list

	return appGateway
}

// remove listener related http backend setting
func (a *Alb) deleteBackendSettingsForAg(appGateway *armnetwork.ApplicationGateway,
	listener *networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	list := make([]*armnetwork.ApplicationGatewayBackendHTTPSettings, 0)
	for _, obj := range appGateway.Properties.BackendHTTPSettingsCollection {
		if obj.Name != nil && strings.HasPrefix(*obj.Name, listener.Name) {
			continue
		}
		list = append(list, obj)
	}
	appGateway.Properties.BackendHTTPSettingsCollection = list

	return appGateway
}

// remove listener related http backend setting
func (a *Alb) deleteHttpListenerForAg(appGateway *armnetwork.ApplicationGateway,
	listener *networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	usedHttpListenerMap := make(map[string]struct{})
	toDeleteHttpListenerMap := make(map[string]struct{})
	for _, routingRule := range appGateway.Properties.RequestRoutingRules {
		usedHttpListenerMap[*routingRule.Properties.HTTPListener.ID] = struct{}{}
	}
	// 仅删除rules相关httpListener
	for _, rule := range listener.Spec.Rules {
		httpListenerName := getHttpListenerName(listener.Spec.Port, rule.Domain)
		httpListenerID := *a.resourceHelper.genSubResource(ResourceProviderApplicationGateway,
			listener.Spec.LoadbalancerID, ResourceTypeHttpListeners, httpListenerName).ID
		if _, ok := usedHttpListenerMap[httpListenerID]; !ok {
			// if not use, delete it
			toDeleteHttpListenerMap[httpListenerID] = struct{}{}
		}
	}

	httpListenerList := make([]*armnetwork.ApplicationGatewayHTTPListener, 0)
	for _, httpListener := range appGateway.Properties.HTTPListeners {
		if _, ok := toDeleteHttpListenerMap[*httpListener.ID]; ok {
			continue
		}
		httpListenerList = append(httpListenerList, httpListener)
	}

	appGateway.Properties.HTTPListeners = httpListenerList
	return appGateway
}

// remove listener related http backend setting
func (a *Alb) deleteURLPathMapForAg(appGateway *armnetwork.ApplicationGateway,
	listener *networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	urlPathMapsMap := make(map[string]*armnetwork.ApplicationGatewayURLPathMap)
	for _, obj := range appGateway.Properties.URLPathMaps {
		urlPathMapsMap[*obj.Name] = obj
	}

	for _, rule := range listener.Spec.Rules {
		// empty path not have urlPathMap
		if rule.Path == "" {
			continue
		}
		pathName := fmt.Sprintf("%x", md5.Sum([]byte(rule.Path)))

		urlPathMapName := getHttpListenerName(listener.Spec.Port, rule.Domain)
		urlPathMap, ok := urlPathMapsMap[urlPathMapName]
		if !ok {
			continue
		}
		newPathRule := make([]*armnetwork.ApplicationGatewayPathRule, 0)
		for _, pathRule := range urlPathMap.Properties.PathRules {
			if pathRule.Name != nil && *pathRule.Name == pathName {
				continue
			}

			newPathRule = append(newPathRule, pathRule)
		}
		urlPathMap.Properties.PathRules = newPathRule
		if len(newPathRule) == 0 {
			delete(urlPathMapsMap, urlPathMapName)
		} else {
			urlPathMapsMap[urlPathMapName] = urlPathMap
		}
	}

	urlPathMapList := make([]*armnetwork.ApplicationGatewayURLPathMap, 0)
	for _, urlPathMap := range urlPathMapsMap {
		urlPathMapList = append(urlPathMapList, urlPathMap)
	}
	appGateway.Properties.URLPathMaps = urlPathMapList
	return appGateway
}

// remove listener related http backend setting
func (a *Alb) deleteRoutingRuleForAg(appGateway *armnetwork.ApplicationGateway,
	listener *networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	urlPathMapsMap := make(map[string]*armnetwork.ApplicationGatewayURLPathMap)
	for _, obj := range appGateway.Properties.URLPathMaps {
		urlPathMapsMap[*obj.Name] = obj
	}

	routingRuleMap := make(map[string]*armnetwork.ApplicationGatewayRequestRoutingRule)
	for _, obj := range appGateway.Properties.RequestRoutingRules {
		routingRuleMap[*obj.Name] = obj
	}

	for _, rule := range listener.Spec.Rules {
		httpListenerName := getHttpListenerName(listener.Spec.Port, rule.Domain)
		// RuleTypeBasic, delete routingRule directly
		if rule.Path == "" {
			delete(routingRuleMap, httpListenerName)
			continue
		}

		// if urlPathMap is empty, delete it
		if _, ok := urlPathMapsMap[httpListenerName]; !ok {
			delete(routingRuleMap, httpListenerName)
		}

	}

	routingRuleList := make([]*armnetwork.ApplicationGatewayRequestRoutingRule, 0)
	for _, routingRule := range routingRuleMap {
		routingRuleList = append(routingRuleList, routingRule)
	}
	appGateway.Properties.RequestRoutingRules = routingRuleList

	return appGateway
}
