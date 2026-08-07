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

package azure

import (
	// NOCC:gas/crypto(误报 未用于创建密钥)
	"crypto/md5"
	"fmt"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/go-autorest/autorest/to"
	mapset "github.com/deckarep/golang-set"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// ensureLoadBalancerListener ensure load balancer listener, azure load balancer can only support layer4 listener
// LoadBalancerListener contains:
// - backendAddressPool
// - probe
// - loadBalancingRules (can be seen as listener)
// (LoadBalancer listener)   --> LoadBalancingRule --> backendAddressPool   --> backend1
//
//												   |--> backend2
//												   |--> ...
//	|--> ...
func (a *Alb) ensureLoadBalancerListener(region string, listeners []*networkextensionv1.Listener) (map[string]cloud.
	Result, error) {
	if len(listeners) == 0 {
		return nil, fmt.Errorf("listeners cannot be empty when ensure loadBalancer listeners")
	}
	blog.V(4).Infof("ensure load balancer listener[%d]", len(listeners))
	for _, listener := range listeners {
		// listener下的targetGroup中，所有backend必须有相同的port
		if !isRuleSamePort(listener) {
			return nil, errors.Wrapf(multiplePortInOneTargetGroupError, "listener '%s/%s' check failed",
				listener.GetNamespace(), listener.GetName())
		}
	}

	// 1. ensure backend address pool
	failedListenerMap := a.ensureAddrPoolForLB(listeners)

	// 筛选address pool更新成功的listener
	successListenerList := make([]*networkextensionv1.Listener, 0)
	for _, li := range listeners {
		if _, ok := failedListenerMap.Load(li.GetName()); ok {
			continue
		}
		successListenerList = append(successListenerList, li)
	}

	// 2. ensure listener
	if err := a.ensureLoadBalancer(region, successListenerList); err != nil {
		for _, li := range successListenerList {
			failedListenerMap.Store(li.GetName(), err)
		}
	}

	// ensure失败的监听器，需要将原因返回上层
	// 遍历全部listener而非successListenerList，否则address pool阶段失败的listener
	// 不会出现在retMap中，上层只能给出笼统的错误
	retMap := make(map[string]cloud.Result)
	for _, li := range listeners {
		if errI, ok := failedListenerMap.Load(li.GetName()); ok {
			err, isErr := errI.(error)
			if !isErr {
				err = fmt.Errorf("ensure listener '%s' failed", li.GetName())
			}
			retMap[li.GetName()] = cloud.Result{
				IsError: true,
				Err:     err,
			}
			continue
		}
		retMap[li.GetName()] = cloud.Result{
			IsError: false,
			Res:     li.GetName(),
		}
	}

	return retMap, nil
}

// ensureAddrPoolForLB ensure addr pool for load balancer
func (a *Alb) ensureAddrPoolForLB(listeners []*networkextensionv1.Listener) *sync.Map {
	failedListenerMap := &sync.Map{}
	// 通过channel限制同时启动的goroutine数量
	ch := make(chan struct{}, CreateGoroutineLimit)
	wg := sync.WaitGroup{}
	wg.Add(len(listeners))
	for _, listener := range listeners {
		ch <- struct{}{}
		// 不同AddrPool之间互不影响，goroutine创建加快效率
		go func(listener *networkextensionv1.Listener) {
			defer func() {
				wg.Done()
				<-ch
			}()
			lbName := listener.Spec.LoadbalancerID

			poolName := getLBRuleTgName(listener.Name, listener.Spec.Port)
			addrList := make([]*armnetwork.LoadBalancerBackendAddress, 0)

			// 根据listener.spec.targetGroup构建AddressPool
			if listener.Spec.TargetGroup != nil && len(listener.Spec.TargetGroup.Backends) != 0 {
				for _, backend := range listener.Spec.TargetGroup.Backends {
					addrList = append(addrList, &armnetwork.LoadBalancerBackendAddress{
						// NOCC:gas/crypto(误报 未用于创建密钥)
						Name: to.StringPtr(fmt.Sprintf("%x", md5.Sum([]byte(backend.IP)))),
						Properties: &armnetwork.LoadBalancerBackendAddressPropertiesFormat{
							IPAddress:      to.StringPtr(backend.IP),
							VirtualNetwork: &armnetwork.SubResource{ID: to.StringPtr(a.sdkWrapper.buildVNetID())},
						},
					})
				}
			}

			_, err := a.sdkWrapper.CreateOrUpdateBackendAddressPool(lbName, poolName, armnetwork.BackendAddressPool{
				Name: to.StringPtr(poolName),
				Properties: &armnetwork.BackendAddressPoolPropertiesFormat{
					LoadBalancerBackendAddresses: addrList,
				},
			})
			if err != nil {
				failedListenerMap.Store(listener.GetName(), err)
			}
		}(listener)
	}
	wg.Wait()
	return failedListenerMap
}

func (a *Alb) ensureLoadBalancer(region string, listeners []*networkextensionv1.Listener) error {
	if len(listeners) == 0 {
		return fmt.Errorf("empty listener list")
	}
	lbResp, err := a.sdkWrapper.GetLoadBalancer(region, listeners[0].Spec.LoadbalancerID)
	if err != nil {
		return err
	}

	lb := &lbResp.LoadBalancer
	// 1. ensure probe
	lb = a.ensureProbesForLB(lb, listeners)

	// 2. ensure loadBalancingRules
	lb, err = a.ensureLoadBalancingRule(lb, listeners)
	if err != nil {
		return err
	}

	if err = validateLBChildNamesUnique(lb); err != nil {
		return err
	}
	if err = validateLBRuleEndpointsUnique(lb); err != nil {
		return err
	}

	// 3. ensure loadBalancer
	_, err = a.sdkWrapper.CreateOrUpdateLoadBalancer(listeners[0].Spec.LoadbalancerID, *lb)
	if err != nil {
		return err
	}

	return nil
}

func (a *Alb) ensureProbesForLB(loadBalancer *armnetwork.LoadBalancer,
	listeners []*networkextensionv1.Listener) *armnetwork.LoadBalancer {
	newProbeList := make([]*armnetwork.Probe, 0)
	probeNameSet := mapset.NewThreadUnsafeSet()
	for _, listener := range listeners {
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

		// translate cr listenerAttribute to cloud request field
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
		probeNameSet.Add(probeName)
	}

	// 避免遗漏用户手动创建的probe
	for _, probe := range loadBalancer.Properties.Probes {
		if probe.Name != nil && probeNameSet.Contains(*probe.Name) {
			continue
		}

		newProbeList = append(newProbeList, probe)
	}

	loadBalancer.Properties.Probes = newProbeList

	return loadBalancer
}

func (a *Alb) ensureLoadBalancingRule(loadBalancer *armnetwork.LoadBalancer,
	listeners []*networkextensionv1.Listener) (*armnetwork.LoadBalancer, error) {
	if len(loadBalancer.Properties.FrontendIPConfigurations) == 0 {
		return nil, unknownFrontIPConfiguration
	}
	// select frontendIP[0] as default
	frontendIPConfigurationID := loadBalancer.Properties.FrontendIPConfigurations[0].ID
	newRules := make([]*armnetwork.LoadBalancingRule, 0)
	ruleNameSet := mapset.NewThreadUnsafeSet()

	for _, listener := range listeners {
		ruleName := getLBRuleTgName(listener.Name, listener.Spec.Port)
		port := getBackendPort(listener.Spec.TargetGroup)

		// translate cr field to cloud request field
		newRule := &armnetwork.LoadBalancingRule{
			Name: to.StringPtr(ruleName),
			Properties: &armnetwork.LoadBalancingRulePropertiesFormat{
				FrontendPort: to.Int32Ptr(int32(listener.Spec.Port)),
				Protocol:     transTransportProtocolPtr(listener.Spec.Protocol),
				BackendAddressPool: a.resourceHelper.genSubResource(ResourceProviderLoadBalancer, listener.Spec.LoadbalancerID,
					ResourceTypeBackendAddressPools, ruleName),
				BackendPort:             to.Int32Ptr(port),
				EnableFloatingIP:        to.BoolPtr(false),
				FrontendIPConfiguration: a.resourceHelper.getSubResourceByID(*frontendIPConfigurationID),
				Probe: a.resourceHelper.genSubResource(ResourceProviderLoadBalancer, listener.Spec.LoadbalancerID,
					ResourceTypeProbes, ruleName),
			},
		}

		if listener.Spec.ListenerAttribute != nil && listener.Spec.ListenerAttribute.SessionTime != 0 {
			sessionTime := listener.Spec.ListenerAttribute.SessionTime
			// sessionTime is in minutes here, matching azure's idleTimeoutInMinutes range of 4~30
			// which validateLBListenerAttribute enforces
			newRule.Properties.IdleTimeoutInMinutes = to.Int32Ptr(int32(sessionTime))
		}

		newRules = append(newRules, newRule)
		ruleNameSet.Add(ruleName)
	}

	// 避免遗漏用户手动创建的规则
	for _, rule := range loadBalancer.Properties.LoadBalancingRules {
		if rule.Name != nil && ruleNameSet.Contains(*rule.Name) {
			continue
		}

		newRules = append(newRules, rule)
	}
	loadBalancer.Properties.LoadBalancingRules = newRules

	return loadBalancer, nil
}

func (a *Alb) deleteLoadBalancerListener(region string, listeners []*networkextensionv1.Listener) error {
	if len(listeners) == 0 {
		return nil
	}
	poolNameSet := mapset.NewThreadUnsafeSet()
	// 一批listener属于同一个lb
	lbName := listeners[0].Spec.LoadbalancerID

	for _, listener := range listeners {
		poolNameSet.Add(getLBRuleTgName(listener.Name, listener.Spec.Port))
	}

	lbResp, err := a.sdkWrapper.GetLoadBalancer(region, lbName)
	if err != nil {
		return err
	}

	lb := lbResp.LoadBalancer
	// 1. delete probe
	newProbes := make([]*armnetwork.Probe, 0)
	for _, probe := range lb.Properties.Probes {
		if probe.Name != nil && poolNameSet.Contains(*probe.Name) {
			continue
		}

		newProbes = append(newProbes, probe)
	}
	lb.Properties.Probes = newProbes

	// 2. delete rule
	newRules := make([]*armnetwork.LoadBalancingRule, 0)
	for _, rule := range lb.Properties.LoadBalancingRules {
		if rule.Name != nil && poolNameSet.Contains(*rule.Name) {
			continue
		}

		newRules = append(newRules, rule)
	}

	lb.Properties.LoadBalancingRules = newRules

	if err = validateLBChildNamesUnique(&lb); err != nil {
		return err
	}

	// 3. 必须先解除loadBalancingRule/probe对addressPool的引用再删除addressPool，
	// 否则azure会拒绝删除仍被引用的addressPool
	if _, err = a.sdkWrapper.CreateOrUpdateLoadBalancer(lbName, lb); err != nil {
		return err
	}

	group := &errgroup.Group{}
	// 设置goroutine上限
	group.SetLimit(DeleteGoroutineLimit)

	for _, listener := range listeners {
		poolName := getLBRuleTgName(listener.Name, listener.Spec.Port)
		group.Go(func() error {
			if err := a.sdkWrapper.DeleteLoadBalanceAddressPool(lbName, poolName); err != nil {
				return err
			}
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return err
	}

	return nil
}

// ensureApplicationGatewayListener ensure listeners of ApplicationGateway.
// ApplicationGateway is layer7 load balancer in azure
// contains:
// - frontendPort
// - AddressPool
// - Probes
// - BackendSettings
// - HttpListener
// - URLPathMap
// - RequestRoutingRule
func (a *Alb) ensureApplicationGatewayListener(region string, listeners []*networkextensionv1.Listener) error {
	if len(listeners) == 0 {
		return nil
	}

	for _, listener := range listeners {
		if !isRuleSamePort(listener) {
			return errors.Wrapf(multiplePortInOneTargetGroupError, "listener '%s/%s' check failed",
				listener.GetNamespace(), listener.GetName())
		}
	}

	lbName := listeners[0].Spec.LoadbalancerID
	// 1. get raw application gateway
	appGatewayRsp, err := a.sdkWrapper.GetApplicationGateway(region, lbName)
	if err != nil {
		return err
	}

	appGateway := &appGatewayRsp.ApplicationGateway

	// 2. ensure frontend port
	appGateway = a.ensureFrontendPortForAg(appGateway, listeners)

	// 3. ensure addrPool
	appGateway = a.ensureAddrPoolForAg(appGateway, listeners)

	// 4. ensure probes
	appGateway = a.ensureProbeForAg(appGateway, listeners)

	// 5. backend settings
	appGateway = a.ensureBackendSettings(appGateway, listeners)

	// 6. listener
	appGateway, err = a.ensureHttpListenerForAg(appGateway, listeners)
	if err != nil {
		return err
	}

	// 7. URLPathMap
	appGateway = a.ensureUrlPathMap(appGateway, listeners)

	// 8. request routing rule
	appGateway, err = a.ensureRequestRoutingRule(appGateway, listeners)
	if err != nil {
		return err
	}

	// 9. drop routes that belong to these listeners but are no longer declared in their spec,
	// otherwise they keep referencing pools/settings that step 3~5 already removed
	appGateway = cleanupStaleAgRoutes(appGateway, listeners)

	// 10. a routing rule shared with another listener survives step 9, but its rule level backend
	// may still point at a pool this reconcile deleted, repair those references before sending
	appGateway = a.repairAgDanglingRefs(appGateway, listeners)

	if err = backfillRoutingRulePriorities(appGateway); err != nil {
		return err
	}
	if err = validateAgChildNamesUnique(appGateway); err != nil {
		return err
	}
	if err = validateAgNoDanglingRefs(appGateway, listeners); err != nil {
		return err
	}
	if err = validateAgRulePriorityUnique(appGateway, listeners); err != nil {
		return err
	}

	// 11. update application gateway
	_, err = a.sdkWrapper.CreateOrUpdateApplicationGateway(listeners[0].Spec.LoadbalancerID, *appGateway)
	if err != nil {
		return err
	}

	return nil
}

// repairAgDanglingRefs clears rule level backend references that this reconcile invalidated.
// Only path based routing rules are repaired. Azure routes a path based rule entirely through its
// URLPathMap, whose DefaultBackendAddressPool serves unmatched requests, so the rule level backend
// carries no routing meaning and dropping it is what the Azure reference model prescribes. Rules
// written by an older version of this controller still carry it, and a rule shared with another
// listener is not rewritten by ensure*, so the stale reference has to be cleared here.
// Basic rules are left alone: there the rule level backend is authoritative, so a dangling one is
// reported by validateAgNoDanglingRefs instead of being silently changed.
func (a *Alb) repairAgDanglingRefs(appGateway *armnetwork.ApplicationGateway,
	listeners []*networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	if appGateway == nil || appGateway.Properties == nil || len(listeners) == 0 {
		return appGateway
	}
	checker := newAgBackendRefChecker(appGateway, listeners)

	for _, routingRule := range appGateway.Properties.RequestRoutingRules {
		if !isPathBasedRoutingRule(routingRule) {
			continue
		}
		if checker.isDangling(subResourceName(routingRule.Properties.BackendAddressPool), checker.pools) {
			routingRule.Properties.BackendAddressPool = nil
		}
		if checker.isDangling(subResourceName(routingRule.Properties.BackendHTTPSettings), checker.settings) {
			routingRule.Properties.BackendHTTPSettings = nil
		}
	}

	return appGateway
}

func (a *Alb) ensureFrontendPortForAg(appGateway *armnetwork.ApplicationGateway,
	listeners []*networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	for _, listener := range listeners {
		listenPort := listener.Spec.Port
		portName := fmt.Sprintf("port_%d", listenPort)

		exists := false
		for _, port := range appGateway.Properties.FrontendPorts {
			if port.Name != nil && *port.Name == portName {
				exists = true
				break
			}
		}
		if exists {
			// continue so remaining listeners can still ensure their ports
			continue
		}

		appGateway.Properties.FrontendPorts = append(appGateway.Properties.FrontendPorts,
			&armnetwork.ApplicationGatewayFrontendPort{
				Name:       to.StringPtr(portName),
				Properties: &armnetwork.ApplicationGatewayFrontendPortPropertiesFormat{Port: to.Int32Ptr(int32(listenPort))},
			})
	}

	return appGateway
}

// azure中，addressPool只包含IP。 监听器具体的后端转发端口/协议由backendSetting指定
func (a *Alb) ensureAddrPoolForAg(appGateway *armnetwork.ApplicationGateway,
	listeners []*networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	newPools := make([]*armnetwork.ApplicationGatewayBackendAddressPool, 0)
	for _, listener := range listeners {
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
	}

	// default pool is shared; add once regardless of listener count
	newPools = append(newPools, &armnetwork.ApplicationGatewayBackendAddressPool{
		Name: to.StringPtr(DefaultBackendPoolName),
		Properties: &armnetwork.ApplicationGatewayBackendAddressPoolPropertiesFormat{
			BackendAddresses: make([]*armnetwork.ApplicationGatewayBackendAddress, 0),
		},
	})

	// keep pools not owned by listeners being reconciled (once each)
	for _, pool := range appGateway.Properties.BackendAddressPools {
		if pool == nil || pool.Name == nil {
			continue
		}
		if keepExistingAgResource(*pool.Name, listeners, DefaultBackendPoolName) {
			newPools = append(newPools, pool)
		}
	}

	appGateway.Properties.BackendAddressPools = newPools

	return appGateway
}

// backendSetting 用于确认后端对应的端口和协议
func (a *Alb) ensureBackendSettings(appGateway *armnetwork.ApplicationGateway,
	listeners []*networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	newSettings := make([]*armnetwork.ApplicationGatewayBackendHTTPSettings, 0)

	for _, listener := range listeners {
		for _, rule := range listener.Spec.Rules {
			settingName := getRuleTgName(listener.Name, rule.Domain, rule.Path, listener.Spec.Port)

			needProbe := false
			var probeResource *armnetwork.SubResource
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
					PickHostNameFromBackendAddress: to.BoolPtr(false),
					Port:                           to.Int32Ptr(int32(port)),
					Probe:                          probeResource,
					ProbeEnabled:                   &needProbe,
					Protocol:                       transAgProtocolPtr(protocol),
					RequestTimeout:                 to.Int32Ptr(DefaultRequestTimeout),
				},
			}
			newSettings = append(newSettings, newSetting)
		}
	}

	// add default settings
	newSettings = append(newSettings, &armnetwork.ApplicationGatewayBackendHTTPSettings{
		Name: to.StringPtr(DefaultBackendSettingName),
		Properties: &armnetwork.ApplicationGatewayBackendHTTPSettingsPropertiesFormat{
			Port:           to.Int32Ptr(80),
			Protocol:       transAgProtocolPtr(string(armnetwork.ApplicationGatewayProtocolHTTP)),
			RequestTimeout: to.Int32Ptr(DefaultRequestTimeout),
		},
	})

	// keep settings not owned by listeners being reconciled (once each)
	for _, setting := range appGateway.Properties.BackendHTTPSettingsCollection {
		if setting == nil || setting.Name == nil {
			continue
		}
		if keepExistingAgResource(*setting.Name, listeners, DefaultBackendSettingName) {
			newSettings = append(newSettings, setting)
		}
	}

	appGateway.Properties.BackendHTTPSettingsCollection = newSettings

	return appGateway
}

// if no need probe, return false and nil subResource
func (a *Alb) ensureProbeForAg(appGateway *armnetwork.ApplicationGateway,
	listeners []*networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	newProbes := make([]*armnetwork.ApplicationGatewayProbe, 0)

	for _, listener := range listeners {
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

			// translate cr field to cloud request field
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
					// azure only accepts 1~20 here, sending the unset 0 fails the whole gateway update
					UnhealthyThreshold: to.Int32Ptr(int32(DefaultProbeUnhealthyThreshold)),
				},
			}
			if healthCheck.UnHealthNum != 0 {
				newProbe.Properties.UnhealthyThreshold = to.Int32Ptr(int32(healthCheck.UnHealthNum))
			}
			// 用户未配置健康检查端口时，使用后端服务的端口
			if healthCheck.HealthCheckPort == 0 {
				newProbe.Properties.Port = to.Int32Ptr(getBackendPort(rule.TargetGroup))
			}
			// 用户未配置健康检查协议时，使用后端服务的协议
			if healthCheck.HealthCheckProtocol == "" {
				if rule.TargetGroup == nil || len(rule.TargetGroup.Backends) == 0 {
					// 空监听器使用HTTP作为默认（用于没有rs，实际不会被用到）
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
	}

	// keep probes not owned by listeners being reconciled (once each)
	for _, probe := range appGateway.Properties.Probes {
		if probe == nil || probe.Name == nil {
			continue
		}
		if keepExistingAgResource(*probe.Name, listeners) {
			newProbes = append(newProbes, probe)
		}
	}

	appGateway.Properties.Probes = newProbes

	return appGateway
}

func (a *Alb) ensureHttpListenerForAg(appGateway *armnetwork.ApplicationGateway,
	listeners []*networkextensionv1.Listener) (*armnetwork.ApplicationGateway, error) {
	if len(appGateway.Properties.FrontendIPConfigurations) != 1 {
		return nil, unknownFrontIPConfiguration
	}
	frontIPConfigurationID := appGateway.Properties.FrontendIPConfigurations[0].ID

	listenerNameSet := mapset.NewThreadUnsafeSet()
	newHttpListenerList := make([]*armnetwork.ApplicationGatewayHTTPListener, 0)

	for _, listener := range listeners {
		for _, rule := range listener.Spec.Rules {
			httpListenerName := getHttpListenerName(listener.Spec.Port, rule.Domain)
			// AGW allows one HTTP listener per port+hostname; multi-path rules share it
			if listenerNameSet.Contains(httpListenerName) {
				continue
			}

			listenPort := listener.Spec.Port

			// translate cr field to cloud request field
			newHttpListener := &armnetwork.ApplicationGatewayHTTPListener{
				Name: to.StringPtr(httpListenerName),
				Properties: &armnetwork.ApplicationGatewayHTTPListenerPropertiesFormat{
					FrontendIPConfiguration: a.resourceHelper.getSubResourceByID(*frontIPConfigurationID),
					FrontendPort: a.resourceHelper.genSubResource(ResourceProviderApplicationGateway,
						listener.Spec.LoadbalancerID, ResourceTypeFrontendPorts, fmt.Sprintf("port_%d", listenPort)),
					Protocol: transAgProtocolPtr(listener.Spec.Protocol),
				},
			}
			// HostName only takes one concrete name, a wildcard has to go through HostNames.
			// The two fields are mutually exclusive.
			switch {
			case rule.Domain == "":
			case isWildcardDomain(rule.Domain):
				newHttpListener.Properties.HostNames = []*string{to.StringPtr(rule.Domain)}
			default:
				newHttpListener.Properties.HostName = to.StringPtr(rule.Domain)
			}
			if strings.ToUpper(listener.Spec.Protocol) == AzureProtocolHTTPS {
				if listener.Spec.Certificate != nil {
					newHttpListener.Properties.SSLCertificate = a.resourceHelper.genSubResource(
						ResourceProviderApplicationGateway, listener.Spec.LoadbalancerID,
						ResourceTypeSSLCertificate, listener.Spec.Certificate.CertID)
				}
				// A https listener carrying a host name is a multi site listener, azure requires SNI
				// for those so that several domains can share one frontend port.
				if rule.Domain != "" {
					newHttpListener.Properties.RequireServerNameIndication = to.BoolPtr(true)
				}
			}

			newHttpListenerList = append(newHttpListenerList, newHttpListener)
			listenerNameSet.Add(httpListenerName)
		}
	}

	// 避免遗漏用户手动创建的监听器
	for _, httpListener := range appGateway.Properties.HTTPListeners {
		if httpListener.Name != nil && listenerNameSet.Contains(*httpListener.Name) {
			continue
		}
		newHttpListenerList = append(newHttpListenerList, httpListener)
	}

	appGateway.Properties.HTTPListeners = newHttpListenerList

	return appGateway, nil
}

func (a *Alb) ensureRequestRoutingRule(appGateway *armnetwork.ApplicationGateway,
	listeners []*networkextensionv1.Listener) (*armnetwork.ApplicationGateway, error) {
	routingRuleMap := make(map[string]*armnetwork.ApplicationGatewayRequestRoutingRule)

	for _, routingRule := range appGateway.Properties.RequestRoutingRules {
		if routingRule == nil || routingRule.Name == nil {
			continue
		}
		routingRuleMap[*routingRule.Name] = routingRule
	}

	// priority决定多站点匹配顺序，用户显式声明的值优先于自动分配，且自动分配不得占用这些值
	userPriorities := collectUserRulePriorities(listeners)
	reserved := reservedPrioritySet(userPriorities)

	ensuredRuleNames := mapset.NewThreadUnsafeSet()

	for _, listener := range listeners {
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

			if routingRule, ok := routingRuleMap[httpListenerName]; ok &&
				routingRule.Properties != nil && routingRule.Properties.RuleType != nil {
				if *routingRule.Properties.RuleType != ruleType {
					return nil, fmt.Errorf("conflict rule type in routingRule[%s], exists: %s, want: %s, "+
						"routingRule info :%s", httpListenerName, *routingRule.Properties.RuleType, ruleType,
						common.ToJsonString(routingRule))
				}
			}

			// one RequestRoutingRule per port+domain; extra paths merge via URLPathMap
			if ensuredRuleNames.Contains(httpListenerName) {
				continue
			}

			// Azure 规定一个ruleTg中的所有backend都必须是相同的port
			ruleTgName := getRuleTgName(listener.Name, rule.Domain, rule.Path, listener.Spec.Port)
			priority := resolveRulePriority(appGateway, routingRuleMap, httpListenerName,
				rule.Domain, userPriorities, reserved)
			if priority == 0 {
				// azure rejects priority 0, fail fast instead of sending an invalid request
				return nil, fmt.Errorf("no available request routing rule priority on gateway '%s', "+
					"all %d priorities are in use", listener.Spec.LoadbalancerID, MaxRoutingRulePriority)
			}

			newRoutingRule := &armnetwork.ApplicationGatewayRequestRoutingRule{
				Name: to.StringPtr(httpListenerName),
				Properties: &armnetwork.ApplicationGatewayRequestRoutingRulePropertiesFormat{
					HTTPListener: a.resourceHelper.genSubResource(ResourceProviderApplicationGateway,
						listener.Spec.LoadbalancerID, ResourceTypeHttpListeners, httpListenerName),
					LoadDistributionPolicy: nil,
					Priority:               &priority,
					RuleType:               &ruleType,
					URLPathMap:             pathMapResource,
				},
			}
			// A path based rule routes purely through its URLPathMap, whose DefaultBackendAddressPool
			// handles unmatched requests. Only a basic rule carries the backend on the rule itself.
			if ruleType == armnetwork.ApplicationGatewayRequestRoutingRuleTypeBasic {
				newRoutingRule.Properties.BackendAddressPool = a.resourceHelper.genSubResource(
					ResourceProviderApplicationGateway, listener.Spec.LoadbalancerID,
					ResourceTypeBackendAddressPools, ruleTgName)
				newRoutingRule.Properties.BackendHTTPSettings = a.resourceHelper.genSubResource(
					ResourceProviderApplicationGateway, listener.Spec.LoadbalancerID,
					ResourceTypeBackendHttpSettingsCollection, ruleTgName)
			}

			routingRuleMap[httpListenerName] = newRoutingRule
			ensuredRuleNames.Add(httpListenerName)

			// add into rule list for build priority
			appGateway.Properties.RequestRoutingRules = append(appGateway.Properties.RequestRoutingRules, newRoutingRule)
		}
	}

	appGateway.Properties.RequestRoutingRules = make([]*armnetwork.ApplicationGatewayRequestRoutingRule, 0)
	for _, routingRule := range routingRuleMap {
		appGateway.Properties.RequestRoutingRules = append(appGateway.Properties.RequestRoutingRules, routingRule)
	}

	return appGateway, nil
}

// URLPath 指定了http监听器中具体路径和addressPool/backendSetting的对应关系
func (a *Alb) ensureUrlPathMap(appGateway *armnetwork.ApplicationGateway,
	listeners []*networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	urlPathMapMap := make(map[string]*armnetwork.ApplicationGatewayURLPathMap)
	for _, urlPathMap := range appGateway.Properties.URLPathMaps {
		if urlPathMap == nil || urlPathMap.Name == nil {
			continue
		}
		urlPathMapMap[*urlPathMap.Name] = urlPathMap
	}

	for _, listener := range listeners {
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

			// NOCC:gas/crypto(误报 未用于创建密钥)
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
	}

	urlPathMapList := make([]*armnetwork.ApplicationGatewayURLPathMap, 0)
	for _, urlPathMap := range urlPathMapMap {
		urlPathMapList = append(urlPathMapList, urlPathMap)
	}
	appGateway.Properties.URLPathMaps = urlPathMapList
	return appGateway

}

// agDesiredRoutes holds the child resource names the current listener specs would generate
type agDesiredRoutes struct {
	tgNames       mapset.Set
	listenerNames mapset.Set
	ports         map[int]struct{}
	listeners     []*networkextensionv1.Listener
}

func newAgDesiredRoutes(listeners []*networkextensionv1.Listener) *agDesiredRoutes {
	desired := &agDesiredRoutes{
		tgNames:       mapset.NewThreadUnsafeSet(),
		listenerNames: mapset.NewThreadUnsafeSet(),
		ports:         make(map[int]struct{}),
		listeners:     listeners,
	}
	for _, listener := range listeners {
		desired.ports[listener.Spec.Port] = struct{}{}
		for _, rule := range listener.Spec.Rules {
			desired.tgNames.Add(getRuleTgName(listener.Name, rule.Domain, rule.Path, listener.Spec.Port))
			desired.listenerNames.Add(getHttpListenerName(listener.Spec.Port, rule.Domain))
		}
	}
	return desired
}

// isStaleTgName reports whether a target group name belongs to one of the listeners being reconciled
// but is no longer part of what their current spec asks for
func (d *agDesiredRoutes) isStaleTgName(name string) bool {
	return name != "" && isAgResourceOwnedByListener(name, d.listeners) && !d.tgNames.Contains(name)
}

// cleanupStaleAgRoutes removes pathRules / routingRules / httpListeners that were generated for the
// given listeners but no longer match their current spec. Resources created manually on the gateway
// reference other pools, so they are left untouched.
func cleanupStaleAgRoutes(appGateway *armnetwork.ApplicationGateway,
	listeners []*networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	if appGateway == nil || appGateway.Properties == nil {
		return appGateway
	}

	desired := newAgDesiredRoutes(listeners)
	remainingPathMaps := prunePathRules(appGateway, desired)
	usedListeners, droppedListeners := pruneRoutingRules(appGateway, desired, remainingPathMaps)
	pruneOrphanHttpListeners(appGateway, desired, usedListeners, droppedListeners)

	return appGateway
}

// prunePathRules drops stale pathRules and any path map emptied by that removal, returning the names
// of the path maps still present. A path map that never had a pathRule is left alone, it was not
// created by this controller.
func prunePathRules(appGateway *armnetwork.ApplicationGateway, desired *agDesiredRoutes) mapset.Set {
	urlPathMaps := make([]*armnetwork.ApplicationGatewayURLPathMap, 0, len(appGateway.Properties.URLPathMaps))
	remaining := mapset.NewThreadUnsafeSet()

	for _, pathMap := range appGateway.Properties.URLPathMaps {
		if pathMap == nil || pathMap.Properties == nil {
			continue
		}
		hadPathRule := len(pathMap.Properties.PathRules) != 0
		pathRules := keepFreshPathRules(pathMap.Properties.PathRules, desired)
		pathMap.Properties.PathRules = pathRules
		if hadPathRule && len(pathRules) == 0 {
			continue
		}
		urlPathMaps = append(urlPathMaps, pathMap)
		if pathMap.Name != nil {
			remaining.Add(*pathMap.Name)
		}
	}

	appGateway.Properties.URLPathMaps = urlPathMaps
	return remaining
}

// keepFreshPathRules filters out path rules whose backend was removed by the current spec
func keepFreshPathRules(all []*armnetwork.ApplicationGatewayPathRule,
	desired *agDesiredRoutes) []*armnetwork.ApplicationGatewayPathRule {
	kept := make([]*armnetwork.ApplicationGatewayPathRule, 0, len(all))
	for _, pathRule := range all {
		if pathRule == nil || pathRule.Properties == nil {
			continue
		}
		if desired.isStaleTgName(subResourceName(pathRule.Properties.BackendAddressPool)) {
			continue
		}
		kept = append(kept, pathRule)
	}
	return kept
}

// pruneRoutingRules drops stale routing rules and reports the http listener names still in use and
// the ones left orphaned. For path based rules the URLPathMap decides routing, so only a removed
// path map makes them stale: the rule level BackendAddressPool may point at any of the listeners
// sharing the path map, so dropping a rule based on it would break the other listeners.
func pruneRoutingRules(appGateway *armnetwork.ApplicationGateway, desired *agDesiredRoutes,
	remainingPathMaps mapset.Set) (used mapset.Set, dropped mapset.Set) {
	routingRules := make([]*armnetwork.ApplicationGatewayRequestRoutingRule, 0,
		len(appGateway.Properties.RequestRoutingRules))
	used = mapset.NewThreadUnsafeSet()
	dropped = mapset.NewThreadUnsafeSet()

	for _, routingRule := range appGateway.Properties.RequestRoutingRules {
		if routingRule == nil || routingRule.Properties == nil {
			continue
		}
		listenerName := subResourceName(routingRule.Properties.HTTPListener)
		pathMapName := subResourceName(routingRule.Properties.URLPathMap)

		stale := desired.isStaleTgName(subResourceName(routingRule.Properties.BackendAddressPool))
		if pathMapName != "" {
			stale = !remainingPathMaps.Contains(pathMapName)
		}
		if stale {
			if listenerName != "" {
				dropped.Add(listenerName)
			}
			continue
		}

		routingRules = append(routingRules, routingRule)
		if listenerName != "" {
			used.Add(listenerName)
		}
	}

	appGateway.Properties.RequestRoutingRules = routingRules
	return used, dropped
}

// pruneOrphanHttpListeners drops only the generated listeners left without a routing rule by
// pruneRoutingRules, so listeners created manually on the gateway are never touched
func pruneOrphanHttpListeners(appGateway *armnetwork.ApplicationGateway, desired *agDesiredRoutes,
	used mapset.Set, dropped mapset.Set) {
	httpListeners := make([]*armnetwork.ApplicationGatewayHTTPListener, 0,
		len(appGateway.Properties.HTTPListeners))

	for _, httpListener := range appGateway.Properties.HTTPListeners {
		if httpListener == nil {
			continue
		}
		if httpListener.Name != nil && isOrphanAgHttpListener(*httpListener.Name, desired, used, dropped) {
			continue
		}
		httpListeners = append(httpListeners, httpListener)
	}

	appGateway.Properties.HTTPListeners = httpListeners
}

func isOrphanAgHttpListener(name string, desired *agDesiredRoutes, used mapset.Set,
	dropped mapset.Set) bool {
	return dropped.Contains(name) && !desired.listenerNames.Contains(name) && !used.Contains(name) &&
		isGeneratedHttpListenerName(name, desired.ports)
}

func (a *Alb) deleteApplicationGatewayListener(region string, listeners []*networkextensionv1.Listener) error {
	if len(listeners) == 0 {
		return nil
	}

	appGatewayRsp, err := a.sdkWrapper.GetApplicationGateway(region, listeners[0].Spec.LoadbalancerID)
	if err != nil {
		return err
	}

	appGateway := &appGatewayRsp.ApplicationGateway

	// addrPool
	appGateway = a.deleteAddrPoolForAg(appGateway, listeners)

	// probes
	appGateway = a.deleteProbeForAg(appGateway, listeners)

	// backend settings
	appGateway = a.deleteBackendSettingsForAg(appGateway, listeners)

	// delete order : urlPathMap -> routingRule -> listener
	// URLPathMap
	appGateway = a.deleteURLPathMapForAg(appGateway, listeners)

	// request routing rule
	appGateway = a.deleteRoutingRuleForAg(appGateway, listeners)

	// listener
	appGateway = a.deleteHttpListenerForAg(appGateway, listeners)

	// a routing rule shared with another listener survives the steps above, repair its rule level
	// backend so it no longer points at the pool/setting just deleted
	appGateway = a.repairAgDanglingRefs(appGateway, listeners)

	if err = backfillRoutingRulePriorities(appGateway); err != nil {
		return err
	}
	if err = validateAgChildNamesUnique(appGateway); err != nil {
		return err
	}
	if err = validateAgNoDanglingRefs(appGateway, listeners); err != nil {
		return err
	}
	if err = validateAgRulePriorityUnique(appGateway, listeners); err != nil {
		return err
	}

	_, err = a.sdkWrapper.CreateOrUpdateApplicationGateway(listeners[0].Spec.LoadbalancerID, *appGateway)
	if err != nil {
		return err
	}
	return nil
}

// remove listener related backendAddressPool
func (a *Alb) deleteAddrPoolForAg(appGateway *armnetwork.ApplicationGateway,
	listeners []*networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	list := make([]*armnetwork.ApplicationGatewayBackendAddressPool, 0)
	for _, obj := range appGateway.Properties.BackendAddressPools {
		if obj == nil {
			continue
		}
		if obj.Name != nil && isAgResourceOwnedByListener(*obj.Name, listeners) {
			continue
		}
		list = append(list, obj)
	}
	appGateway.Properties.BackendAddressPools = list

	return appGateway
}

// remove listener related probe
func (a *Alb) deleteProbeForAg(appGateway *armnetwork.ApplicationGateway,
	listeners []*networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	list := make([]*armnetwork.ApplicationGatewayProbe, 0)
	for _, obj := range appGateway.Properties.Probes {
		if obj == nil {
			continue
		}
		if obj.Name != nil && isAgResourceOwnedByListener(*obj.Name, listeners) {
			continue
		}
		list = append(list, obj)
	}
	appGateway.Properties.Probes = list

	return appGateway
}

// remove listener related http backend setting
func (a *Alb) deleteBackendSettingsForAg(appGateway *armnetwork.ApplicationGateway,
	listeners []*networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	list := make([]*armnetwork.ApplicationGatewayBackendHTTPSettings, 0)
	for _, obj := range appGateway.Properties.BackendHTTPSettingsCollection {
		if obj == nil {
			continue
		}
		if obj.Name != nil && isAgResourceOwnedByListener(*obj.Name, listeners) {
			continue
		}
		list = append(list, obj)
	}
	appGateway.Properties.BackendHTTPSettingsCollection = list

	return appGateway
}

// remove listener related http backend setting
func (a *Alb) deleteHttpListenerForAg(appGateway *armnetwork.ApplicationGateway,
	listeners []*networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	usedHttpListenerMap := make(map[string]struct{})
	toDeleteHttpListenerMap := make(map[string]struct{})
	for _, routingRule := range appGateway.Properties.RequestRoutingRules {
		if routingRule == nil || routingRule.Properties == nil ||
			routingRule.Properties.HTTPListener == nil || routingRule.Properties.HTTPListener.ID == nil {
			continue
		}
		usedHttpListenerMap[*routingRule.Properties.HTTPListener.ID] = struct{}{}
	}

	for _, listener := range listeners {
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
	}

	httpListenerList := make([]*armnetwork.ApplicationGatewayHTTPListener, 0)
	for _, httpListener := range appGateway.Properties.HTTPListeners {
		if httpListener == nil {
			continue
		}
		if httpListener.ID != nil {
			if _, ok := toDeleteHttpListenerMap[*httpListener.ID]; ok {
				continue
			}
		}
		httpListenerList = append(httpListenerList, httpListener)
	}

	appGateway.Properties.HTTPListeners = httpListenerList
	return appGateway
}

// remove listener related http backend setting
func (a *Alb) deleteURLPathMapForAg(appGateway *armnetwork.ApplicationGateway,
	listeners []*networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	urlPathMapsMap := make(map[string]*armnetwork.ApplicationGatewayURLPathMap)
	for _, obj := range appGateway.Properties.URLPathMaps {
		if obj == nil || obj.Name == nil {
			continue
		}
		urlPathMapsMap[*obj.Name] = obj
	}

	for _, listener := range listeners {
		for _, rule := range listener.Spec.Rules {
			// empty path not have urlPathMap
			if rule.Path == "" {
				continue
			}
			// NOCC:gas/crypto(误报 未用于创建密钥)
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
	listeners []*networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	urlPathMapsMap := make(map[string]*armnetwork.ApplicationGatewayURLPathMap)
	for _, obj := range appGateway.Properties.URLPathMaps {
		if obj == nil || obj.Name == nil {
			continue
		}
		urlPathMapsMap[*obj.Name] = obj
	}

	routingRuleMap := make(map[string]*armnetwork.ApplicationGatewayRequestRoutingRule)
	for _, obj := range appGateway.Properties.RequestRoutingRules {
		if obj == nil || obj.Name == nil {
			continue
		}
		routingRuleMap[*obj.Name] = obj
	}

	for _, listener := range listeners {
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
	}

	routingRuleList := make([]*armnetwork.ApplicationGatewayRequestRoutingRule, 0)
	for _, routingRule := range routingRuleMap {
		routingRuleList = append(routingRuleList, routingRule)
	}
	appGateway.Properties.RequestRoutingRules = routingRuleList

	return appGateway
}
