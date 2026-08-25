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
	// NOCC:gas/crypto(未用于生成密钥)
	"crypto/md5"
	"fmt"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	mapset "github.com/deckarep/golang-set"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

func transTransportProtocolPtr(str string) *armnetwork.TransportProtocol {
	var protocol armnetwork.TransportProtocol
	switch strings.ToUpper(str) {
	case AzureProtocolTCP:
		protocol = armnetwork.TransportProtocolTCP
	case AzureProtocolUDP:
		protocol = armnetwork.TransportProtocolUDP
	}

	return &protocol
}

func transAgProtocolPtr(str string) *armnetwork.ApplicationGatewayProtocol {
	var protocol armnetwork.ApplicationGatewayProtocol
	switch strings.ToUpper(str) {
	case AzureProtocolHTTP:
		protocol = armnetwork.ApplicationGatewayProtocolHTTP
	case AzureProtocolHTTPS:
		protocol = armnetwork.ApplicationGatewayProtocolHTTPS
	case AzureProtocolTCP:
		protocol = armnetwork.ApplicationGatewayProtocolTCP
	case AzureProtocolTLS:
		protocol = armnetwork.ApplicationGatewayProtocolTLS
	}

	return &protocol
}

func transProbeProtocolPtr(str string) *armnetwork.ProbeProtocol {
	var protocol armnetwork.ProbeProtocol
	switch strings.ToUpper(str) {
	case AzureProtocolHTTP:
		protocol = armnetwork.ProbeProtocolHTTP
	case AzureProtocolHTTPS:
		protocol = armnetwork.ProbeProtocolHTTPS
	case AzureProtocolTCP, AzureProtocolUDP: // azure不支持UDP协议的健康检查，当使用UDP时，转为TCP
		protocol = armnetwork.ProbeProtocolTCP
	}
	return &protocol
}

// transAgProbeMatch translate healthCheck to azure entity
func transAgProbeMatch(healthCheck *networkextensionv1.ListenerHealthCheck) *armnetwork.
	ApplicationGatewayProbeHealthResponseMatch {
	if healthCheck == nil || healthCheck.HTTPCode < 1 || healthCheck.HTTPCode > 31 {
		return nil
	}
	match := &armnetwork.ApplicationGatewayProbeHealthResponseMatch{}
	httpCode := healthCheck.HTTPCode
	cnt := 1
	for httpCode != 0 && cnt <= 5 {
		if httpCode&1 != 0 {
			matchCode := fmt.Sprintf("%d-%d", cnt*100, (cnt+1)*100-1)
			match.StatusCodes = append(match.StatusCodes, to.StringPtr(matchCode))
		}
		httpCode = httpCode >> 1
		cnt++
	}
	return match
}

// listenerName.md5(listenerName+domain+path)
func getRuleTgName(listenerName, domain, path string, listenPort int) string {
	// NOCC:gas/crypto(未用于生成密钥)
	return fmt.Sprintf("%s%s%x%s%d", listenerName, agResourceNameSep,
		md5.Sum([]byte(listenerName+domain+path)), agResourceNameSep, listenPort)
}

// listenPort.md5(domain)
func getHttpListenerName(listenPort int, domain string) string {
	// NOCC:gas/crypto(未用于生成密钥)
	return fmt.Sprintf("%d.%x", listenPort, md5.Sum([]byte(domain)))
}

// isAgResourceOwnedByListener reports whether an AGW child resource belongs to one of the given
// listeners. getRuleTgName and getLBRuleTgName always append a "." separator after the listener
// name, so the separator must be part of the prefix. Matching on the bare listener name would let
// listener "lb-80" claim resources of listener "lb-8080".
func isAgResourceOwnedByListener(name string, listeners []*networkextensionv1.Listener) bool {
	for _, listener := range listeners {
		if strings.HasPrefix(name, listener.Name+agResourceNameSep) {
			return true
		}
	}
	return false
}

// keepExistingAgResource reports whether an existing AGW child resource should be preserved
// when reconciling the given listeners. Resources owned by listeners in the reconcile batch
// or matching defaultNames are rebuilt by ensure* and must not be kept.
func keepExistingAgResource(name string, listeners []*networkextensionv1.Listener,
	defaultNames ...string) bool {
	for _, defaultName := range defaultNames {
		if name == defaultName {
			return false
		}
	}
	return !isAgResourceOwnedByListener(name, listeners)
}

// subResourceName returns the resource name of an azure sub resource reference, which is the last
// segment of its ID. Returns empty string when the reference or ID is missing.
func subResourceName(sub *armnetwork.SubResource) string {
	if sub == nil || sub.ID == nil {
		return ""
	}
	id := *sub.ID
	if idx := strings.LastIndex(id, "/"); idx >= 0 {
		return id[idx+1:]
	}
	return id
}

// isGeneratedHttpListenerName reports whether name matches the "{port}.{md5hex}" pattern produced
// by getHttpListenerName for one of the given ports. Used to avoid touching listeners that were
// created manually on the gateway.
func isGeneratedHttpListenerName(name string, ports map[int]struct{}) bool {
	idx := strings.Index(name, agResourceNameSep)
	if idx <= 0 {
		return false
	}
	port, err := strconv.Atoi(name[:idx])
	if err != nil {
		return false
	}
	if _, ok := ports[port]; !ok {
		return false
	}
	return isMD5Hex(name[idx+1:])
}

func isMD5Hex(s string) bool {
	if len(s) != md5.Size*2 {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

// isPathBasedRoutingRule reports whether a routing rule forwards through a url path map
func isPathBasedRoutingRule(rule *armnetwork.ApplicationGatewayRequestRoutingRule) bool {
	if rule == nil || rule.Properties == nil {
		return false
	}
	if rule.Properties.RuleType != nil {
		return *rule.Properties.RuleType ==
			armnetwork.ApplicationGatewayRequestRoutingRuleTypePathBasedRouting
	}
	return subResourceName(rule.Properties.URLPathMap) != ""
}

// agPoolNameSet collects existing backend address pool names of an application gateway
func agPoolNameSet(appGateway *armnetwork.ApplicationGateway) mapset.Set {
	names := mapset.NewThreadUnsafeSet()
	for _, pool := range appGateway.Properties.BackendAddressPools {
		if pool != nil && pool.Name != nil {
			names.Add(*pool.Name)
		}
	}
	return names
}

// agSettingNameSet collects existing backend http settings names of an application gateway
func agSettingNameSet(appGateway *armnetwork.ApplicationGateway) mapset.Set {
	names := mapset.NewThreadUnsafeSet()
	for _, setting := range appGateway.Properties.BackendHTTPSettingsCollection {
		if setting != nil && setting.Name != nil {
			names.Add(*setting.Name)
		}
	}
	return names
}

// agBackendRefChecker reports references to backends that belong to the listeners being reconciled
// but are no longer present on the gateway
type agBackendRefChecker struct {
	pools     mapset.Set
	settings  mapset.Set
	listeners []*networkextensionv1.Listener
}

func newAgBackendRefChecker(appGateway *armnetwork.ApplicationGateway,
	listeners []*networkextensionv1.Listener) *agBackendRefChecker {
	return &agBackendRefChecker{
		pools:     agPoolNameSet(appGateway),
		settings:  agSettingNameSet(appGateway),
		listeners: listeners,
	}
}

func (c *agBackendRefChecker) isDangling(name string, existing mapset.Set) bool {
	return name != "" && !existing.Contains(name) && isAgResourceOwnedByListener(name, c.listeners)
}

// check validates both backend references of one route
func (c *agBackendRefChecker) check(owner string, pool, setting *armnetwork.SubResource) error {
	if name := subResourceName(pool); c.isDangling(name, c.pools) {
		return fmt.Errorf("%s references missing backend address pool '%s'", owner, name)
	}
	if name := subResourceName(setting); c.isDangling(name, c.settings) {
		return fmt.Errorf("%s references missing backend http setting '%s'", owner, name)
	}
	return nil
}

// validateAgNoDanglingRefs ensures no route still references a backend that belongs to the
// listeners being reconciled but is no longer present on the gateway. Such a request is rejected by
// Azure, failing here keeps the error actionable. References to resources outside this reconcile
// are not checked, they may legitimately be managed elsewhere.
func validateAgNoDanglingRefs(appGateway *armnetwork.ApplicationGateway,
	listeners []*networkextensionv1.Listener) error {
	if appGateway == nil || appGateway.Properties == nil {
		return nil
	}
	checker := newAgBackendRefChecker(appGateway, listeners)

	for _, routingRule := range appGateway.Properties.RequestRoutingRules {
		if routingRule == nil || routingRule.Properties == nil {
			continue
		}
		owner := "request routing rule"
		if routingRule.Name != nil {
			owner = fmt.Sprintf("request routing rule '%s'", *routingRule.Name)
		}
		if err := checker.check(owner, routingRule.Properties.BackendAddressPool,
			routingRule.Properties.BackendHTTPSettings); err != nil {
			return err
		}
	}

	return validateAgPathRuleRefs(appGateway, checker)
}

func validateAgPathRuleRefs(appGateway *armnetwork.ApplicationGateway,
	checker *agBackendRefChecker) error {
	for _, pathMap := range appGateway.Properties.URLPathMaps {
		if pathMap == nil || pathMap.Properties == nil {
			continue
		}
		for _, pathRule := range pathMap.Properties.PathRules {
			if pathRule == nil || pathRule.Properties == nil {
				continue
			}
			owner := "url path rule"
			if pathMap.Name != nil && pathRule.Name != nil {
				owner = fmt.Sprintf("url path rule '%s/%s'", *pathMap.Name, *pathRule.Name)
			}
			if err := checker.check(owner, pathRule.Properties.BackendAddressPool,
				pathRule.Properties.BackendHTTPSettings); err != nil {
				return err
			}
		}
	}

	return nil
}

// findDuplicateNames returns names that appear more than once in the slice.
func findDuplicateNames(names []string) []string {
	seen := make(map[string]int, len(names))
	for _, name := range names {
		if name == "" {
			continue
		}
		seen[name]++
	}
	var dups []string
	for name, cnt := range seen {
		if cnt > 1 {
			dups = append(dups, name)
		}
	}
	return dups
}

func appendChildResourceName(names []string, name *string) []string {
	if name == nil || *name == "" {
		return names
	}
	return append(names, *name)
}

// validateAgChildNamesUnique ensures each AGW child resource collection has unique names.
// Azure rejects PUT with DuplicateResourceName when the same type has duplicate names.
func validateAgChildNamesUnique(appGateway *armnetwork.ApplicationGateway) error {
	if appGateway == nil || appGateway.Properties == nil {
		return nil
	}
	props := appGateway.Properties

	var frontendPortNames []string
	for _, item := range props.FrontendPorts {
		if item != nil {
			frontendPortNames = appendChildResourceName(frontendPortNames, item.Name)
		}
	}
	var poolNames []string
	for _, item := range props.BackendAddressPools {
		if item != nil {
			poolNames = appendChildResourceName(poolNames, item.Name)
		}
	}
	var settingNames []string
	for _, item := range props.BackendHTTPSettingsCollection {
		if item != nil {
			settingNames = appendChildResourceName(settingNames, item.Name)
		}
	}
	var probeNames []string
	for _, item := range props.Probes {
		if item != nil {
			probeNames = appendChildResourceName(probeNames, item.Name)
		}
	}
	var httpListenerNames []string
	for _, item := range props.HTTPListeners {
		if item != nil {
			httpListenerNames = appendChildResourceName(httpListenerNames, item.Name)
		}
	}
	var urlPathMapNames []string
	for _, item := range props.URLPathMaps {
		if item != nil {
			urlPathMapNames = appendChildResourceName(urlPathMapNames, item.Name)
		}
	}
	var routingRuleNames []string
	for _, item := range props.RequestRoutingRules {
		if item != nil {
			routingRuleNames = appendChildResourceName(routingRuleNames, item.Name)
		}
	}

	checks := []struct {
		kind  string
		names []string
	}{
		{kind: "frontendPorts", names: frontendPortNames},
		{kind: "backendAddressPools", names: poolNames},
		{kind: "backendHttpSettingsCollection", names: settingNames},
		{kind: "probes", names: probeNames},
		{kind: "httpListeners", names: httpListenerNames},
		{kind: "urlPathMaps", names: urlPathMapNames},
		{kind: "requestRoutingRules", names: routingRuleNames},
	}

	for _, check := range checks {
		if dups := findDuplicateNames(check.names); len(dups) > 0 {
			return fmt.Errorf("application gateway %s has duplicate child resource names: %v",
				check.kind, dups)
		}
		if err := validateResourceNameLength(check.kind, check.names); err != nil {
			return err
		}
	}
	return nil
}

// validateLBChildNamesUnique ensures each load balancer child collection has unique names that fit
// the azure name limit. The layer 4 path builds names from the load balancer name too, so it can hit
// the same limits as the application gateway path.
func validateLBChildNamesUnique(loadBalancer *armnetwork.LoadBalancer) error {
	if loadBalancer == nil || loadBalancer.Properties == nil {
		return nil
	}

	var probeNames, ruleNames, poolNames []string
	for _, item := range loadBalancer.Properties.Probes {
		if item != nil {
			probeNames = appendChildResourceName(probeNames, item.Name)
		}
	}
	for _, item := range loadBalancer.Properties.LoadBalancingRules {
		if item != nil {
			ruleNames = appendChildResourceName(ruleNames, item.Name)
		}
	}
	for _, item := range loadBalancer.Properties.BackendAddressPools {
		if item != nil {
			poolNames = appendChildResourceName(poolNames, item.Name)
		}
	}

	return checkChildNameCollections("load balancer", map[string][]string{
		"probes":              probeNames,
		"loadBalancingRules":  ruleNames,
		"backendAddressPools": poolNames,
	})
}

// checkChildNameCollections rejects duplicate and over long names in each collection
func checkChildNameCollections(owner string, collections map[string][]string) error {
	for kind, names := range collections {
		if dups := findDuplicateNames(names); len(dups) > 0 {
			return fmt.Errorf("%s %s has duplicate child resource names: %v", owner, kind, dups)
		}
		if err := validateResourceNameLength(kind, names); err != nil {
			return err
		}
	}
	return nil
}

// validateLBRuleEndpointsUnique ensures no two load balancing rules share a frontend, protocol and
// port. There is no priority or evaluation order on a load balancer: azure requires the combination
// to be unique and answers with "The frontend, protocol and port combination of each load balancing
// rule and inbound NAT rule on a load balancer must be unique" without naming the offenders.
func validateLBRuleEndpointsUnique(loadBalancer *armnetwork.LoadBalancer) error {
	if loadBalancer == nil || loadBalancer.Properties == nil {
		return nil
	}
	owner := make(map[string]string)

	for _, rule := range loadBalancer.Properties.LoadBalancingRules {
		if rule == nil || rule.Properties == nil || rule.Properties.FrontendPort == nil {
			continue
		}
		protocol := ""
		if rule.Properties.Protocol != nil {
			protocol = string(*rule.Properties.Protocol)
		}
		key := fmt.Sprintf("%s/%s/%d", subResourceName(rule.Properties.FrontendIPConfiguration),
			protocol, *rule.Properties.FrontendPort)

		name := ""
		if rule.Name != nil {
			name = *rule.Name
		}
		if other, ok := owner[key]; ok {
			return fmt.Errorf("load balancing rules '%s' and '%s' both serve %s protocol %s port %d, "+
				"azure requires the frontend, protocol and port combination to be unique", other, name,
				subResourceName(rule.Properties.FrontendIPConfiguration), protocol,
				*rule.Properties.FrontendPort)
		}
		owner[key] = name
	}
	return nil
}

// validateResourceNameLength enforces the azure resource name limit. Generated names embed the load
// balancer name, so a long gateway name can push them past the limit and azure would answer with an
// opaque 400. See https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/resource-name-rules
func validateResourceNameLength(kind string, names []string) error {
	for _, name := range names {
		if len(name) > MaxAzureResourceNameLen {
			return fmt.Errorf("application gateway %s name '%s' is %d characters, azure allows at "+
				"most %d, please use a shorter load balancer name", kind, name, len(name),
				MaxAzureResourceNameLen)
		}
	}
	return nil
}

// listenerName.port
func getLBRuleTgName(listenerName string, listenPort int) string {
	return fmt.Sprintf("%s%s%d", listenerName, agResourceNameSep, listenPort)
}

// isSamePort check if all backends have same port, return true if all ports are same
func isSamePort(targetGroup *networkextensionv1.ListenerTargetGroup) bool {
	if targetGroup == nil || len(targetGroup.Backends) == 0 {
		return true
	}
	for i := 1; i < len(targetGroup.Backends); i++ {
		if targetGroup.Backends[i].Port != targetGroup.Backends[i-1].Port {
			return false
		}
	}

	return true
}

// isRuleSamePort check backends' port are same in one rule, return true if all ports are same
func isRuleSamePort(listener *networkextensionv1.Listener) bool {
	for _, rule := range listener.Spec.Rules {
		if !isSamePort(rule.TargetGroup) {
			return false
		}
	}
	return true
}

// collectUserRulePriorities maps a routing rule name to the priority declared by the user through
// listenerAttribute.priority. Azure creates one routing rule per port+domain, so every rule of the
// same port+domain has to agree on the value, and two different rules cannot claim the same one.
//
// Bad declarations are dropped with a warning instead of failing: the webhook is where they get
// rejected, and one stale object must not stop every layer 7 listener of a shared gateway from
// syncing. A dropped declaration simply falls back to auto assignment.
func collectUserRulePriorities(listeners []*networkextensionv1.Listener) map[string]int32 {
	byRule := make(map[string]int32)
	owner := make(map[int32]string)

	for _, listener := range listeners {
		for _, rule := range listener.Spec.Rules {
			if rule.ListenerAttribute == nil || rule.ListenerAttribute.Priority == 0 {
				continue
			}
			recordUserRulePriority(listener, rule, byRule, owner)
		}
	}

	return byRule
}

func recordUserRulePriority(listener *networkextensionv1.Listener, rule networkextensionv1.ListenerRule,
	byRule map[string]int32, owner map[int32]string) {
	// compare before narrowing, a large int would otherwise wrap into the valid range
	declared := rule.ListenerAttribute.Priority
	label := describeRuleTarget(listener.Spec.Port, rule.Domain)
	if declared < 1 || declared > int(MaxRoutingRulePriority) {
		blog.Warnf("azure: listener '%s' declares priority %d for %s, azure only accepts 1~%d, "+
			"falling back to auto assignment", listener.Name, declared, label, MaxRoutingRulePriority)
		return
	}
	priority := int32(declared)

	name := getHttpListenerName(listener.Spec.Port, rule.Domain)
	if exist, ok := byRule[name]; ok && exist != priority {
		blog.Warnf("azure: conflicting priority for %s: %d and %d. azure keeps one routing rule per "+
			"port and domain, so all paths of a domain share a single priority, keeping %d",
			label, exist, priority, exist)
		return
	}
	if other, ok := owner[priority]; ok && other != label {
		blog.Warnf("azure: priority %d is declared by both %s and %s, azure requires a unique "+
			"priority per routing rule, keeping it on %s and auto assigning the other",
			priority, other, label, other)
		return
	}

	byRule[name] = priority
	owner[priority] = label
}

// describeRuleTarget renders the port+domain a routing rule serves, used in operator facing messages
func describeRuleTarget(port int, domain string) string {
	if domain == "" {
		return fmt.Sprintf("port %d (no domain)", port)
	}
	return fmt.Sprintf("port %d domain '%s'", port, domain)
}

func reservedPrioritySet(userPriorities map[string]int32) mapset.Set {
	reserved := mapset.NewThreadUnsafeSet()
	for _, priority := range userPriorities {
		reserved.Add(priority)
	}
	return reserved
}

// agRuleLabels maps generated routing rule names to a "port/domain" label. Rule names are
// port.md5(domain), so an error naming only the rule cannot be acted on by the user.
func agRuleLabels(listeners []*networkextensionv1.Listener) map[string]string {
	labels := make(map[string]string)
	for _, listener := range listeners {
		for _, rule := range listener.Spec.Rules {
			name := getHttpListenerName(listener.Spec.Port, rule.Domain)
			labels[name] = describeRuleTarget(listener.Spec.Port, rule.Domain)
		}
	}
	return labels
}

func describeAgRule(name string, labels map[string]string) string {
	if label, ok := labels[name]; ok {
		return fmt.Sprintf("%s (%s)", label, name)
	}
	// a reconcile only sees its own batch, azure splits http and https listeners of one gateway
	// into separate batches, so this covers another ingress and manually created rules alike
	return fmt.Sprintf("'%s' (another ingress, another port, or created outside this controller)",
		name)
}

// validateAgRulePriorityUnique ensures no two routing rules share a priority. Azure answers
// "Priority must be unique across all the request routing rules" without naming them, and a user
// declared priority can collide with a rule this controller does not manage.
func validateAgRulePriorityUnique(appGateway *armnetwork.ApplicationGateway,
	listeners []*networkextensionv1.Listener) error {
	if appGateway == nil || appGateway.Properties == nil {
		return nil
	}
	labels := agRuleLabels(listeners)
	owner := make(map[int32]string)

	for _, rule := range appGateway.Properties.RequestRoutingRules {
		if rule == nil || rule.Properties == nil || rule.Properties.Priority == nil {
			continue
		}
		name := ""
		if rule.Name != nil {
			name = *rule.Name
		}
		priority := *rule.Properties.Priority
		if other, ok := owner[priority]; ok {
			return fmt.Errorf("priority %d is used by both %s and %s, azure requires it to be "+
				"unique across the whole gateway", priority, other, describeAgRule(name, labels))
		}
		owner[priority] = describeAgRule(name, labels)
	}
	return nil
}

// isWildcardDomain reports whether a domain uses azure's wildcard syntax. Such a domain has to go
// into the listener's HostNames list, HostName only accepts a single concrete name.
func isWildcardDomain(domain string) bool {
	return strings.ContainsAny(domain, "*?")
}

// autoPriorityStart returns where auto assignment starts looking for this rule. A rule carrying a
// domain drives a multi site listener and has to be evaluated before a basic one, which matches
// what a request without a matching host header should fall back to.
func autoPriorityStart(domain string) int32 {
	if domain != "" {
		return MultiSiteRulePriorityStart
	}
	return BasicRulePriorityStart
}

// resolveRulePriority picks the priority of one routing rule. Order is: what the user declared,
// then the value the rule already has, then a freshly allocated one. Keeping the existing value
// matters because priority drives multi site matching order and which certificate a client without
// SNI receives, so reallocating every reconcile would keep reconfiguring the gateway. A rule has to
// give up its existing value when another rule of the batch declared it, otherwise both would land
// on the same priority and azure would reject every update. Returns 0 when nothing is available.
func resolveRulePriority(appGateway *armnetwork.ApplicationGateway,
	existing map[string]*armnetwork.ApplicationGatewayRequestRoutingRule, name, domain string,
	userPriorities map[string]int32, reserved mapset.Set) int32 {
	if priority, ok := userPriorities[name]; ok && priority != 0 {
		return priority
	}
	if priority := existingRulePriority(existing, name); priority != 0 && !reserved.Contains(priority) {
		return priority
	}
	return findNextFreePriority(appGateway, reserved, autoPriorityStart(domain))
}

// findNextFreePriority walks up from start in steps of RulePriorityJump, the scheme AGIC uses to fill
// in priorities the user did not declare. Starting high keeps the low numbers free for users, which
// also matters because a reconcile only sees the declarations of its own batch and azure splits the
// http and https listeners of one gateway into separate batches.
// AGIC returns the maximum priority once the scan runs out, which can produce a duplicate; fall back
// to any free value instead so the gateway update stays valid.
func findNextFreePriority(appGateway *armnetwork.ApplicationGateway, reserved mapset.Set,
	start int32) int32 {
	used := usedPrioritySlots(appGateway)
	for priority := start; priority <= MaxRoutingRulePriority; priority += RulePriorityJump {
		if used[priority] || (reserved != nil && reserved.Contains(priority)) {
			continue
		}
		return priority
	}

	blog.Warnf("azure: no free priority left from %d upwards, falling back to any free value, "+
		"a basic listener may end up outranking a multi site one", start)
	return generatePriorityExcluding(appGateway, reserved)
}

// existingRulePriority returns the priority already assigned to a routing rule, or 0 when the rule
// is new or carries no priority yet
func existingRulePriority(rules map[string]*armnetwork.ApplicationGatewayRequestRoutingRule,
	name string) int32 {
	rule, ok := rules[name]
	if !ok || rule.Properties == nil || rule.Properties.Priority == nil {
		return 0
	}
	priority := *rule.Properties.Priority
	if priority < 1 || priority > MaxRoutingRulePriority {
		return 0
	}
	return priority
}

// backfillRoutingRulePriorities gives a priority to every rule that has none. Since api-version
// 2021-08-01 azure rejects a gateway where only some rules carry a priority with
// ApplicationGatewayRequestRoutingRulePartialPriorityDefined, which would otherwise block every
// update on a gateway that already has rules created before that api-version.
func backfillRoutingRulePriorities(appGateway *armnetwork.ApplicationGateway) error {
	if appGateway == nil || appGateway.Properties == nil {
		return nil
	}
	for _, rule := range appGateway.Properties.RequestRoutingRules {
		if rule == nil || rule.Properties == nil || rule.Properties.Priority != nil {
			continue
		}
		priority := generatePriority(appGateway)
		if priority == 0 {
			return fmt.Errorf("no available request routing rule priority to backfill, "+
				"all %d priorities are in use", MaxRoutingRulePriority)
		}
		rule.Properties.Priority = to.Int32Ptr(priority)
	}
	return nil
}

// return 0 if no available priority, callers must treat 0 as exhausted
func generatePriority(appGateway *armnetwork.ApplicationGateway) int32 {
	return generatePriorityExcluding(appGateway, nil)
}

// usedPrioritySlots marks which of the 1~20000 priorities the gateway already uses
func usedPrioritySlots(appGateway *armnetwork.ApplicationGateway) []bool {
	used := make([]bool, MaxRoutingRulePriority+1)
	used[0] = true
	for _, requestRule := range appGateway.Properties.RequestRoutingRules {
		// priority is optional on v1 SKU gateways and on user created rules
		if requestRule == nil || requestRule.Properties == nil || requestRule.Properties.Priority == nil {
			continue
		}
		priority := *requestRule.Properties.Priority
		if priority < 0 || priority > MaxRoutingRulePriority {
			continue
		}
		used[priority] = true
	}
	return used
}

// generatePriorityExcluding picks the smallest priority that is neither in use on the gateway nor
// reserved. Reserved values are the ones the user declared for rules of the current batch that have
// not been built yet, taking one of them would make that rule fail later.
// return 0 if no available priority, callers must treat 0 as exhausted
func generatePriorityExcluding(appGateway *armnetwork.ApplicationGateway, reserved mapset.Set) int32 {
	usedPriority := usedPrioritySlots(appGateway)

	for i, used := range usedPriority {
		if used {
			continue
		}
		if reserved != nil && reserved.Contains(int32(i)) {
			continue
		}
		return int32(i)
	}

	return 0
}

// getBackendPort return targetGroup's backend port, assume all ports are same. return 80 as default
func getBackendPort(targetGroup *networkextensionv1.ListenerTargetGroup) int32 {
	port := 80
	if targetGroup != nil && len(targetGroup.Backends) != 0 {
		port = targetGroup.Backends[0].Port
	}

	return int32(port)
}

func splitListenersToDiffProtocol(listeners []*networkextensionv1.Listener) [][]*networkextensionv1.Listener {
	retMap := make(map[string][]*networkextensionv1.Listener)
	for _, li := range listeners {
		var listenerList []*networkextensionv1.Listener
		if _, ok := retMap[li.Spec.Protocol]; ok {
			listenerList = retMap[li.Spec.Protocol]
		} else {
			listenerList = make([]*networkextensionv1.Listener, 0)
		}

		if li.Spec.EndPort != 0 {
			listenerList = append(listenerList, splitSegListener([]*networkextensionv1.
				Listener{li})...)
		} else {
			listenerList = append(listenerList, li)
		}

		retMap[li.Spec.Protocol] = listenerList
	}

	retList := make([][]*networkextensionv1.Listener, 0)
	for _, list := range retMap {
		retList = append(retList, list)
	}
	return retList
}

func splitSegListener(listenerList []*networkextensionv1.Listener) []*networkextensionv1.Listener {
	newListenerList := make([]*networkextensionv1.Listener, 0)

	for _, listener := range listenerList {
		if listener.Spec.EndPort == 0 {
			newListenerList = append(newListenerList, listener)
		} else {
			portIndex := 0
			for i := listener.Spec.Port; i <= listener.Spec.EndPort; i++ {
				// generate single port listener to ensure listener
				li := listener.DeepCopy()
				li.Spec.Port = i
				li.Spec.EndPort = 0
				if li.Spec.TargetGroup != nil {
					for j := range li.Spec.TargetGroup.Backends {
						li.Spec.TargetGroup.Backends[j].Port += portIndex
					}
				}
				portIndex++
				newListenerList = append(newListenerList, li)
			}
		}
	}
	return newListenerList
}
