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
	"reflect"
	"strconv"
	"strings"

	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	tcommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// convert clb health check info to crd fields
func convertHealthCheck(hc *tclb.HealthCheck) *networkextensionv1.ListenerHealthCheck {
	if hc == nil {
		return nil
	}
	healthCheck := &networkextensionv1.ListenerHealthCheck{}
	if hc.HealthSwitch == nil {
		healthCheck.Enabled = false
		return healthCheck
	}
	healthCheck.Enabled = true
	if hc.TimeOut != nil {
		healthCheck.Timeout = int(*hc.TimeOut)
	}
	if hc.IntervalTime != nil {
		healthCheck.IntervalTime = int(*hc.IntervalTime)
	}
	if hc.HealthNum != nil {
		healthCheck.HealthNum = int(*hc.HealthNum)
	}
	if hc.UnHealthNum != nil {
		healthCheck.UnHealthNum = int(*hc.UnHealthNum)
	}
	if hc.HttpCode != nil {
		healthCheck.HTTPCode = int(*hc.HttpCode)
	}
	if hc.HttpCheckPath != nil {
		healthCheck.HTTPCheckPath = *hc.HttpCheckPath
	}
	if hc.HttpCheckMethod != nil {
		healthCheck.HTTPCheckMethod = *hc.HttpCheckMethod
	}
	return healthCheck
}

// convert clb listener attribute to crd fields
func convertListenerAttribute(lis *tclb.Listener) *networkextensionv1.IngressListenerAttribute {
	if lis == nil {
		return nil
	}
	attr := &networkextensionv1.IngressListenerAttribute{}
	if lis.SessionExpireTime != nil {
		attr.SessionTime = int(*lis.SessionExpireTime)
	}
	if lis.Scheduler != nil {
		attr.LbPolicy = *lis.Scheduler
	}
	if lis.HealthCheck != nil {
		attr.HealthCheck = convertHealthCheck(lis.HealthCheck)
	}
	if lis.SniSwitch != nil {
		attr.SniSwitch = int(*lis.SniSwitch)
	}
	if lis.KeepaliveEnable != nil {
		attr.KeepAliveEnable = int(*lis.KeepaliveEnable)
	}
	return attr
}

// convert clb rule attribute to crd fields
func convertRuleAttribute(rule *tclb.RuleOutput) *networkextensionv1.IngressListenerAttribute {
	if rule == nil {
		return nil
	}
	attr := &networkextensionv1.IngressListenerAttribute{}
	if rule.SessionExpireTime != nil {
		attr.SessionTime = int(*rule.SessionExpireTime)
	}
	if rule.Scheduler != nil {
		attr.LbPolicy = *rule.Scheduler
	}
	if rule.HealthCheck != nil {
		attr.HealthCheck = convertHealthCheck(rule.HealthCheck)
	}
	return attr
}

// convert clb certificates info to crd fields
func convertCertificate(certs *tclb.CertificateOutput) *networkextensionv1.IngressListenerCertificate {
	if certs == nil {
		return nil
	}
	certificate := &networkextensionv1.IngressListenerCertificate{}
	if certs.SSLMode != nil {
		certificate.Mode = *certs.SSLMode
	}
	if certs.CertId != nil {
		certificate.CertID = *certs.CertId
	}
	if certs.CertCaId != nil {
		certificate.CertCaID = *certs.CertCaId
	}
	return certificate
}

// convert clb backends info to crd fields
func convertClbBackends(backends []*tclb.Backend) *networkextensionv1.ListenerTargetGroup {
	tg := &networkextensionv1.ListenerTargetGroup{}
	for _, backend := range backends {
		if len(backend.PrivateIpAddresses) == 0 {
			continue
		}
		tg.Backends = append(tg.Backends, networkextensionv1.ListenerBackend{
			IP:     *backend.PrivateIpAddresses[0],
			Port:   int(*backend.Port),
			Weight: int(*backend.Weight),
		})
	}
	return tg
}

// convert heatlh check in crd to clb request field
func transIngressHealtchCheck(hc *networkextensionv1.ListenerHealthCheck) *tclb.HealthCheck {
	if hc == nil {
		return nil
	}
	healthCheck := &tclb.HealthCheck{}
	var heatlthSwitch int64
	if hc.Enabled {
		heatlthSwitch = 1
	} else {
		heatlthSwitch = 0
	}
	healthCheck.HealthSwitch = tcommon.Int64Ptr(heatlthSwitch)
	if hc.IntervalTime != 0 {
		healthCheck.IntervalTime = tcommon.Int64Ptr(int64(hc.IntervalTime))
	}
	if hc.HealthNum != 0 {
		healthCheck.HealthNum = tcommon.Int64Ptr(int64(hc.HealthNum))
	}
	if hc.UnHealthNum != 0 {
		healthCheck.UnHealthNum = tcommon.Int64Ptr(int64(hc.UnHealthNum))
	}
	if hc.Timeout != 0 {
		healthCheck.TimeOut = tcommon.Int64Ptr(int64(hc.Timeout))
	}
	if len(hc.HTTPCheckMethod) != 0 {
		healthCheck.HttpCheckMethod = tcommon.StringPtr(hc.HTTPCheckMethod)
	}
	if len(hc.HTTPCheckPath) != 0 {
		healthCheck.HttpCheckPath = tcommon.StringPtr(hc.HTTPCheckPath)
	}
	if hc.HTTPCode != 0 {
		healthCheck.HttpCode = tcommon.Int64Ptr(int64(hc.HTTPCode))
	}
	return healthCheck
}

// convert certificates in crd to clb request field
func transIngressCertificate(tc *networkextensionv1.IngressListenerCertificate) *tclb.CertificateInput {
	if tc == nil {
		return nil
	}
	certInput := &tclb.CertificateInput{}
	if len(tc.Mode) != 0 {
		certInput.SSLMode = tcommon.StringPtr(tc.Mode)
	}
	if len(tc.CertID) != 0 {
		certInput.CertId = tcommon.StringPtr(tc.CertID)
	}
	if len(tc.CertCaID) != 0 {
		certInput.CertCaId = tcommon.StringPtr(tc.CertCaID)
	}
	return certInput
}

// getIPPortKey return certain format
func getIPPortKey(ip string, port int) string {
	return ip + ":" + strconv.Itoa(port)
}

// getTargets transfer crd targetGroup to clb request field
func getTargets(tg *networkextensionv1.ListenerTargetGroup) []*tclb.Target {
	if tg == nil {
		return nil
	}
	var retTargets []*tclb.Target
	for _, backend := range tg.Backends {
		retTargets = append(retTargets, &tclb.Target{
			EniIp:  tcommon.StringPtr(backend.IP),
			Port:   tcommon.Int64Ptr(int64(backend.Port)),
			Weight: tcommon.Int64Ptr(int64(backend.Weight)),
		})
	}
	return retTargets
}

// getDiffBetweenTargetGroup compare targetGroup between cloud and local
// return addTarget/delTarget/updateTarget
// addTarget: in local but not in cloud
// delTarget: in cloud but not in local
// updateTarget: ip&port both in cloud and local, but weight different
func getDiffBetweenTargetGroup(existedTg, newTg *networkextensionv1.ListenerTargetGroup) (
	[]*tclb.Target, []*tclb.Target, []*tclb.Target) {

	existedBackendsMap := make(map[string]networkextensionv1.ListenerBackend)
	if existedTg != nil {
		for _, backend := range existedTg.Backends {
			existedBackendsMap[getIPPortKey(backend.IP, backend.Port)] = backend
		}
	}

	newBackendMap := make(map[string]networkextensionv1.ListenerBackend)
	if newTg != nil {
		for _, backend := range newTg.Backends {
			newBackendMap[getIPPortKey(backend.IP, backend.Port)] = backend
		}
	}

	var addTargets []*tclb.Target
	var delTargets []*tclb.Target
	var updateWeightTargets []*tclb.Target
	if newTg != nil {
		for _, backend := range newTg.Backends {
			existedBackend, ok := existedBackendsMap[getIPPortKey(backend.IP, backend.Port)]
			if !ok {
				addTargets = append(addTargets, &tclb.Target{
					EniIp:  tcommon.StringPtr(backend.IP),
					Port:   tcommon.Int64Ptr(int64(backend.Port)),
					Weight: tcommon.Int64Ptr(int64(backend.Weight)),
				})
			} else if backend.Weight != existedBackend.Weight {
				updateWeightTargets = append(updateWeightTargets, &tclb.Target{
					EniIp:  tcommon.StringPtr(backend.IP),
					Port:   tcommon.Int64Ptr(int64(backend.Port)),
					Weight: tcommon.Int64Ptr(int64(backend.Weight)),
				})
			}
		}
	}

	if existedTg != nil {
		for _, backend := range existedTg.Backends {
			if _, ok := newBackendMap[getIPPortKey(backend.IP, backend.Port)]; !ok {
				delTargets = append(delTargets, &tclb.Target{
					EniIp: tcommon.StringPtr(backend.IP),
					Port:  tcommon.Int64Ptr(int64(backend.Port)),
				})
			}
		}
	}

	return addTargets, delTargets, updateWeightTargets
}

// compareTargetGroup
func compareTargetGroup(existedTg, newTg *networkextensionv1.ListenerTargetGroup) (
	[]networkextensionv1.ListenerBackend, []networkextensionv1.ListenerBackend, []networkextensionv1.ListenerBackend) {

	existedBackendsMap := make(map[string]networkextensionv1.ListenerBackend)
	if existedTg != nil {
		for _, backend := range existedTg.Backends {
			existedBackendsMap[getIPPortKey(backend.IP, backend.Port)] = backend
		}
	}

	newBackendMap := make(map[string]networkextensionv1.ListenerBackend)
	if newTg != nil {
		for _, backend := range newTg.Backends {
			newBackendMap[getIPPortKey(backend.IP, backend.Port)] = backend
		}
	}

	var addBackends []networkextensionv1.ListenerBackend
	var delBackends []networkextensionv1.ListenerBackend
	var updateWeightBackends []networkextensionv1.ListenerBackend
	if newTg != nil {
		for _, backend := range newTg.Backends {
			existedBackend, ok := existedBackendsMap[getIPPortKey(backend.IP, backend.Port)]
			if !ok {
				addBackends = append(addBackends, backend)
			} else if backend.Weight != existedBackend.Weight {
				updateWeightBackends = append(updateWeightBackends, backend)
			}
		}
	}

	if existedTg != nil {
		for _, backend := range existedTg.Backends {
			if _, ok := newBackendMap[getIPPortKey(backend.IP, backend.Port)]; !ok {
				delBackends = append(delBackends, backend)
			}
		}
	}

	return addBackends, delBackends, updateWeightBackends
}

// getDomainPathKey return certain format
func getDomainPathKey(domain, path string) string {
	return domain + path
}

// to see if the attribute should be update
func needUpdateAttribute(oldAttr, newAttr *networkextensionv1.IngressListenerAttribute) bool {
	if newAttr == nil {
		return false
	}
	if oldAttr == nil {
		return true
	}
	if (len(newAttr.LbPolicy) != 0 && newAttr.LbPolicy != oldAttr.LbPolicy) ||
		newAttr.SessionTime != oldAttr.SessionTime {
		return true
	}
	if newAttr.HealthCheck == nil {
		return false
	}
	if oldAttr.HealthCheck == nil {
		return true
	}

	return needUpdateHealthCheck(newAttr.HealthCheck, oldAttr.HealthCheck)
}

// needUpdateHealthCheck return true if health check need update
func needUpdateHealthCheck(newHealth, oldHealth *networkextensionv1.ListenerHealthCheck) bool {
	if newHealth.Enabled != oldHealth.Enabled {
		return true
	}
	if (len(newHealth.HTTPCheckMethod) != 0 && newHealth.HTTPCheckMethod != oldHealth.HTTPCheckMethod) ||
		(len(newHealth.HTTPCheckPath) != 0 && newHealth.HTTPCheckPath != oldHealth.HTTPCheckPath) ||
		(newHealth.HTTPCode != 0 && newHealth.HTTPCode != oldHealth.HTTPCode) ||
		(newHealth.HealthNum != 0 && newHealth.HealthNum != oldHealth.HealthNum) ||
		(newHealth.UnHealthNum != 0 && newHealth.UnHealthNum != oldHealth.UnHealthNum) ||
		(newHealth.IntervalTime != 0 && newHealth.IntervalTime != oldHealth.IntervalTime) ||
		(newHealth.Timeout != 0 && newHealth.Timeout != oldHealth.Timeout) {
		return true
	}

	return false
}

// getDiffBetweenListenerRule compare listener Rule in cloud and local
// return  addRules, delRules, updateOldRules, updatedRules
// - addRule: in local but not in cloud
// - delRule: in cloud but not in local
// - updatedRule: both in cloud and local, but attr different
// - updateOldRules: localRule before update, have same order of updatedRule
func getDiffBetweenListenerRule(existedListener, newListener *networkextensionv1.Listener) (
	[]networkextensionv1.ListenerRule, []networkextensionv1.ListenerRule,
	[]networkextensionv1.ListenerRule, []networkextensionv1.ListenerRule) {

	existedRuleMap := make(map[string]networkextensionv1.ListenerRule)
	for _, rule := range existedListener.Spec.Rules {
		existedRuleMap[getDomainPathKey(rule.Domain, rule.Path)] = rule
	}
	newRuleMap := make(map[string]networkextensionv1.ListenerRule)
	for _, rule := range newListener.Spec.Rules {
		newRuleMap[getDomainPathKey(rule.Domain, rule.Path)] = rule
	}

	var addRules []networkextensionv1.ListenerRule
	var delRules []networkextensionv1.ListenerRule
	var updateOldRules []networkextensionv1.ListenerRule
	var updatedRules []networkextensionv1.ListenerRule
	for _, rule := range newListener.Spec.Rules {
		existedRule, ok := existedRuleMap[getDomainPathKey(rule.Domain, rule.Path)]
		if !ok {
			addRules = append(addRules, rule)
			continue
		}
		addBackends, delBackends, updateWeightBackends := getDiffBetweenTargetGroup(
			existedRule.TargetGroup, rule.TargetGroup)
		if len(addBackends) != 0 || len(delBackends) != 0 || len(updateWeightBackends) != 0 ||
			(rule.ListenerAttribute != nil &&
				needUpdateAttribute(existedRule.ListenerAttribute, rule.ListenerAttribute)) ||
			!reflect.DeepEqual(existedRule.Certificate, rule.Certificate) {
			updateRule := networkextensionv1.ListenerRule{}
			updateRule.Domain = rule.Domain
			updateRule.Path = rule.Path
			// here should add ruleid for a update rule
			// UpdateRule interface need ruleid
			updateRule.RuleID = existedRule.RuleID
			updateRule.ListenerAttribute = rule.ListenerAttribute
			updateRule.TargetGroup = rule.TargetGroup
			updatedRules = append(updatedRules, updateRule)
			updateOldRules = append(updateOldRules, existedRule)
		}
	}
	for _, rule := range existedListener.Spec.Rules {
		if _, ok := newRuleMap[getDomainPathKey(rule.Domain, rule.Path)]; !ok {
			delRules = append(delRules, rule)
		}
	}
	return addRules, delRules, updateOldRules, updatedRules
}

// splitListenersToDiffProtocol split listener by its protocol
func splitListenersToDiffProtocol(listenerList []*networkextensionv1.Listener) [][]*networkextensionv1.Listener {
	retMap := make(map[string][]*networkextensionv1.Listener)
	for _, li := range listenerList {
		if _, ok := retMap[li.Spec.Protocol]; ok {
			retMap[li.Spec.Protocol] = append(retMap[li.Spec.Protocol], li)
		} else {
			retMap[li.Spec.Protocol] = make([]*networkextensionv1.Listener, 0)
			retMap[li.Spec.Protocol] = append(retMap[li.Spec.Protocol], li)
		}
	}
	retList := make([][]*networkextensionv1.Listener, 0)
	for _, list := range retMap {
		retList = append(retList, list)
	}
	return retList
}

// splitListenersToDiffBatch split listeners by its 'listenerAttribute' and 'certificate'
func splitListenersToDiffBatch(listenerList []*networkextensionv1.Listener) [][]*networkextensionv1.Listener {
	attrList := make([]struct {
		attr *networkextensionv1.IngressListenerAttribute
		cert *networkextensionv1.IngressListenerCertificate
	}, 0)
	retList := make([][]*networkextensionv1.Listener, 0)
	for _, li := range listenerList {
		found := false
		for index, attr := range attrList {
			if reflect.DeepEqual(attr.attr, li.Spec.ListenerAttribute) &&
				reflect.DeepEqual(attr.cert, li.Spec.Certificate) {
				retList[index] = append(retList[index], li)
				found = true
				break
			}
		}
		if found {
			continue
		}
		attrList = append(attrList, struct {
			attr *networkextensionv1.IngressListenerAttribute
			cert *networkextensionv1.IngressListenerCertificate
		}{
			attr: li.Spec.ListenerAttribute,
			cert: li.Spec.Certificate,
		})
		tmpList := make([]*networkextensionv1.Listener, 0)
		tmpList = append(tmpList, li)
		retList = append(retList, tmpList)
	}
	return retList
}

// getListenerNames return []string of listener name
func getListenerNames(listenerList []*networkextensionv1.Listener) []string {
	var retList []string
	for _, li := range listenerList {
		retList = append(retList, li.GetName())
	}
	return retList
}

// convertHealthStatus transfer cloud status to local
func convertHealthStatus(status string) string {
	var statusStr string
	switch status {
	case ClbBackendAlive:
		statusStr = cloud.BackendHealthStatusHealthy
	case ClbBackendDead:
		statusStr = cloud.BackendHealthStatusUnhealthy
	default:
		statusStr = cloud.BackendHealthStatusUnknown
	}
	return statusStr
}

// transferCloudListener transfer cloud listener to local listener
func transferCloudListener(lbID string, cloudLiResp *tclb.DescribeListenersResponse, portMap map[int]struct{}) (
	[]string, map[string]*networkextensionv1.Listener, map[string]*networkextensionv1.IngressListenerAttribute,
	map[string]*networkextensionv1.IngressListenerCertificate) {
	var listenerIDs []string
	retListenerMap := make(map[string]*networkextensionv1.Listener)
	ruleIDAttrMap := make(map[string]*networkextensionv1.IngressListenerAttribute)
	ruleIDCertMap := make(map[string]*networkextensionv1.IngressListenerCertificate)

	for _, cloudLi := range cloudLiResp.Response.Listeners {
		// only care about listener with given ports
		if _, ok := portMap[int(*cloudLi.Port)]; !ok {
			continue
		}
		listenerIDs = append(listenerIDs, *cloudLi.ListenerId)
		li := &networkextensionv1.Listener{}
		li.Spec.LoadbalancerID = lbID
		li.Spec.Port = int(*cloudLi.Port)
		// get segment listener end port
		if cloudLi.EndPort != nil && *cloudLi.EndPort > 0 {
			li.Spec.EndPort = int(*cloudLi.EndPort)
		}
		li.Spec.Protocol = strings.ToLower(*cloudLi.Protocol)
		li.Spec.Certificate = convertCertificate(cloudLi.Certificate)
		li.Spec.ListenerAttribute = convertListenerAttribute(cloudLi)
		if len(cloudLi.Rules) != 0 {
			for _, respRule := range cloudLi.Rules {
				if respRule.LocationId != nil {
					ruleIDAttrMap[*respRule.LocationId] = convertRuleAttribute(respRule)
					ruleIDCertMap[*respRule.LocationId] = convertCertificate(respRule.Certificate)
				}
			}
		}
		li.Status.ListenerID = *cloudLi.ListenerId
		retListenerMap[common.GetListenerNameWithProtocol(lbID, li.Spec.Protocol, li.Spec.Port, li.Spec.EndPort)] = li
	}

	return listenerIDs, retListenerMap, ruleIDAttrMap, ruleIDCertMap
}

// compareListener compare listener of cloud and local
func compareListener(lbID string, cloudListenerMap map[string]*networkextensionv1.Listener,
	localListener []*networkextensionv1.Listener) ([]*networkextensionv1.Listener, []*networkextensionv1.Listener, []*networkextensionv1.Listener) {
	addListeners := make([]*networkextensionv1.Listener, 0)
	updatedListeners := make([]*networkextensionv1.Listener, 0)
	deleteCloudListeners := make([]*networkextensionv1.Listener, 0)
	for _, li := range localListener {
		cloudLi, ok := cloudListenerMap[common.GetListenerNameWithProtocol(
			lbID, li.Spec.Protocol, li.Spec.Port, li.Spec.EndPort)]
		if !ok {
			addListeners = append(addListeners, li)
		} else {
			if strings.ToLower(cloudLi.Spec.Protocol) != strings.ToLower(li.Spec.Protocol) {
				deleteCloudListeners = append(deleteCloudListeners, cloudLi)
				addListeners = append(addListeners, li)
			} else {
				updatedListeners = append(updatedListeners, li)
			}
		}
	}
	return addListeners, updatedListeners, deleteCloudListeners
}
