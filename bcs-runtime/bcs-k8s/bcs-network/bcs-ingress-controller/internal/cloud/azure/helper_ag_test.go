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
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/go-autorest/autorest/to"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

func newTestAlb() *Alb {
	return &Alb{
		resourceHelper: NewResourceHelper(testSubscriptionID, testResourceGroup),
	}
}

func newEmptyAppGateway() *armnetwork.ApplicationGateway {
	return &armnetwork.ApplicationGateway{
		Properties: &armnetwork.ApplicationGatewayPropertiesFormat{
			FrontendIPConfigurations: []*armnetwork.ApplicationGatewayFrontendIPConfiguration{
				{ID: to.StringPtr(testIDPrefix() + "/frontendIPConfigurations/frontend-ip")},
			},
			FrontendPorts:                 []*armnetwork.ApplicationGatewayFrontendPort{},
			BackendAddressPools:           []*armnetwork.ApplicationGatewayBackendAddressPool{},
			BackendHTTPSettingsCollection: []*armnetwork.ApplicationGatewayBackendHTTPSettings{},
			Probes:                        []*armnetwork.ApplicationGatewayProbe{},
			HTTPListeners:                 []*armnetwork.ApplicationGatewayHTTPListener{},
			URLPathMaps:                   []*armnetwork.ApplicationGatewayURLPathMap{},
			RequestRoutingRules:           []*armnetwork.ApplicationGatewayRequestRoutingRule{},
		},
	}
}

func dualPathListener(name, domain string, port int) *networkextensionv1.Listener {
	return &networkextensionv1.Listener{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test-ns", Name: name},
		Spec: networkextensionv1.ListenerSpec{
			LoadbalancerID: testAppGatewayName,
			Port:           port,
			Protocol:       AzureProtocolHTTPS,
			Certificate: &networkextensionv1.IngressListenerCertificate{
				CertID: "test-wildcard-cert",
			},
			Rules: []networkextensionv1.ListenerRule{
				{
					Domain: domain,
					Path:   "/api/specific/*",
					TargetGroup: &networkextensionv1.ListenerTargetGroup{
						TargetGroupProtocol: AzureProtocolHTTP,
						Backends: []networkextensionv1.ListenerBackend{
							{IP: "10.0.0.1", Port: 48777},
						},
					},
				},
				{
					Domain: domain,
					Path:   "/*",
					ListenerAttribute: &networkextensionv1.IngressListenerAttribute{
						HealthCheck: &networkextensionv1.ListenerHealthCheck{
							Enabled:             true,
							HealthCheckProtocol: AzureProtocolHTTP,
							HTTPCheckPath:       "/healthz",
							IntervalTime:        20,
							Timeout:             20,
							UnHealthNum:         3,
						},
					},
					TargetGroup: &networkextensionv1.ListenerTargetGroup{
						TargetGroupProtocol: AzureProtocolHTTP,
						Backends: []networkextensionv1.ListenerBackend{
							{IP: "10.0.0.2", Port: 48777},
						},
					},
				},
			},
		},
	}
}

// TestHttpListenerSameDomainDedup ensures same port+domain multi-path rules create one HTTP listener.
// NOCC:tosa/fn_length(测试函数)
func TestHttpListenerSameDomainDedup(t *testing.T) {
	alb := newTestAlb()
	appGateway := newEmptyAppGateway()
	domain := "app.example.com"
	listener := dualPathListener("agw-443", domain, 443)

	result, err := alb.ensureHttpListenerForAg(appGateway, []*networkextensionv1.Listener{listener})
	if err != nil {
		t.Fatalf("ensureHttpListenerForAg failed: %v", err)
	}
	if len(result.Properties.HTTPListeners) != 1 {
		t.Fatalf("expect 1 http listener for dual-path same domain, got %d",
			len(result.Properties.HTTPListeners))
	}

	wantName := getHttpListenerName(443, domain)
	gotName := *result.Properties.HTTPListeners[0].Name
	if gotName != wantName {
		t.Errorf("unexpected http listener name got=%s want=%s", gotName, wantName)
	}
	if err = validateAgChildNamesUnique(result); err != nil {
		t.Fatalf("duplicate child names after ensureHttpListenerForAg: %v", err)
	}
}

// TestAgResourcesUniqueMultiPath builds full AGW L7 objects for dual-path ingress and checks uniqueness.
// NOCC:tosa/fn_length(测试函数)
func TestAgResourcesUniqueMultiPath(t *testing.T) {
	alb := newTestAlb()
	appGateway := newEmptyAppGateway()
	domain := "app.example.com"
	listener := dualPathListener("agw-443", domain, 443)
	listeners := []*networkextensionv1.Listener{listener}

	appGateway = alb.ensureFrontendPortForAg(appGateway, listeners)
	appGateway = alb.ensureAddrPoolForAg(appGateway, listeners)
	appGateway = alb.ensureProbeForAg(appGateway, listeners)
	appGateway = alb.ensureBackendSettings(appGateway, listeners)

	var err error
	appGateway, err = alb.ensureHttpListenerForAg(appGateway, listeners)
	if err != nil {
		t.Fatalf("ensureHttpListenerForAg failed: %v", err)
	}
	appGateway = alb.ensureUrlPathMap(appGateway, listeners)
	appGateway, err = alb.ensureRequestRoutingRule(appGateway, listeners)
	if err != nil {
		t.Fatalf("ensureRequestRoutingRule failed: %v", err)
	}

	if err = validateAgChildNamesUnique(appGateway); err != nil {
		t.Fatalf("duplicate AGW child names: %v", err)
	}

	if len(appGateway.Properties.HTTPListeners) != 1 {
		t.Fatalf("expect 1 http listener, got %d", len(appGateway.Properties.HTTPListeners))
	}
	if len(appGateway.Properties.RequestRoutingRules) != 1 {
		t.Fatalf("expect 1 request routing rule, got %d", len(appGateway.Properties.RequestRoutingRules))
	}
	if len(appGateway.Properties.URLPathMaps) != 1 {
		t.Fatalf("expect 1 url path map, got %d", len(appGateway.Properties.URLPathMaps))
	}
	pathRules := appGateway.Properties.URLPathMaps[0].Properties.PathRules
	if len(pathRules) != 2 {
		t.Fatalf("expect 2 path rules under same domain, got %d", len(pathRules))
	}
	wantPaths := []string{"/api/specific/*", "/*"}
	for i, want := range wantPaths {
		if pathRules[i].Properties == nil || len(pathRules[i].Properties.Paths) != 1 ||
			pathRules[i].Properties.Paths[0] == nil || *pathRules[i].Properties.Paths[0] != want {
			t.Errorf("path rule %d changed configured order, want %q", i, want)
		}
	}
	// 2 rule pools + 1 default
	if len(appGateway.Properties.BackendAddressPools) != 3 {
		t.Fatalf("expect 3 backend pools (2 rules + default), got %d",
			len(appGateway.Properties.BackendAddressPools))
	}
}

// TestFrontendPortMultiListeners ensures different ports are all ensured (no early return).
// NOCC:tosa/fn_length(测试函数)
func TestFrontendPortMultiListeners(t *testing.T) {
	alb := newTestAlb()
	appGateway := newEmptyAppGateway()
	// seed existing port so first listener would historically trigger early return
	appGateway.Properties.FrontendPorts = []*armnetwork.ApplicationGatewayFrontendPort{
		{
			Name:       to.StringPtr("port_80"),
			Properties: &armnetwork.ApplicationGatewayFrontendPortPropertiesFormat{Port: to.Int32Ptr(80)},
		},
	}

	listeners := []*networkextensionv1.Listener{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "li-80"},
			Spec:       networkextensionv1.ListenerSpec{Port: 80, Protocol: AzureProtocolHTTP},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "li-443"},
			Spec:       networkextensionv1.ListenerSpec{Port: 443, Protocol: AzureProtocolHTTPS},
		},
	}

	appGateway = alb.ensureFrontendPortForAg(appGateway, listeners)
	if len(appGateway.Properties.FrontendPorts) != 2 {
		t.Fatalf("expect frontend ports 80 and 443, got %d", len(appGateway.Properties.FrontendPorts))
	}
	if err := validateAgChildNamesUnique(appGateway); err != nil {
		t.Fatalf("duplicate frontend port names: %v", err)
	}
}

// TestAddrPoolKeepExistingOnce ensures unrelated pools are kept once when ensuring multi listeners.
// NOCC:tosa/fn_length(测试函数)
func TestAddrPoolKeepExistingOnce(t *testing.T) {
	alb := newTestAlb()
	appGateway := newEmptyAppGateway()
	appGateway.Properties.BackendAddressPools = []*armnetwork.ApplicationGatewayBackendAddressPool{
		{
			Name: to.StringPtr("other-listener.deadbeef.80"),
			Properties: &armnetwork.ApplicationGatewayBackendAddressPoolPropertiesFormat{
				BackendAddresses: []*armnetwork.ApplicationGatewayBackendAddress{},
			},
		},
	}

	listeners := []*networkextensionv1.Listener{
		dualPathListener("li-a-443", "a.example.com", 443),
		dualPathListener("li-b-8443", "b.example.com", 8443),
	}

	appGateway = alb.ensureAddrPoolForAg(appGateway, listeners)
	if err := validateAgChildNamesUnique(appGateway); err != nil {
		t.Fatalf("duplicate backend pool names: %v", err)
	}

	defaultCount := 0
	otherCount := 0
	for _, pool := range appGateway.Properties.BackendAddressPools {
		if pool.Name == nil {
			continue
		}
		if *pool.Name == DefaultBackendPoolName {
			defaultCount++
		}
		if *pool.Name == "other-listener.deadbeef.80" {
			otherCount++
		}
	}
	if defaultCount != 1 {
		t.Fatalf("expect default backend pool once, got %d", defaultCount)
	}
	if otherCount != 1 {
		t.Fatalf("expect unrelated pool kept once, got %d", otherCount)
	}
}

// TestDeleteAgResourcesByListener ensures listener owned pool/probe/setting are actually removed
// while shared default and unrelated resources are preserved.
// NOCC:tosa/fn_length(测试函数)
func TestDeleteAgResourcesByListener(t *testing.T) {
	alb := newTestAlb()
	listener := dualPathListener("agw-443", "app.example.com", 443)
	listeners := []*networkextensionv1.Listener{listener}

	appGateway := newEmptyAppGateway()
	appGateway = alb.ensureAddrPoolForAg(appGateway, listeners)
	appGateway = alb.ensureProbeForAg(appGateway, listeners)
	appGateway = alb.ensureBackendSettings(appGateway, listeners)

	// unrelated resources owned by another listener must survive
	appGateway.Properties.BackendAddressPools = append(appGateway.Properties.BackendAddressPools,
		&armnetwork.ApplicationGatewayBackendAddressPool{Name: to.StringPtr("other-li.abcd.80")})
	appGateway.Properties.Probes = append(appGateway.Properties.Probes,
		&armnetwork.ApplicationGatewayProbe{Name: to.StringPtr("other-li.abcd.80")})
	appGateway.Properties.BackendHTTPSettingsCollection = append(
		appGateway.Properties.BackendHTTPSettingsCollection,
		&armnetwork.ApplicationGatewayBackendHTTPSettings{Name: to.StringPtr("other-li.abcd.80")})

	appGateway = alb.deleteAddrPoolForAg(appGateway, listeners)
	appGateway = alb.deleteProbeForAg(appGateway, listeners)
	appGateway = alb.deleteBackendSettingsForAg(appGateway, listeners)

	for _, pool := range appGateway.Properties.BackendAddressPools {
		if pool.Name != nil && strings.HasPrefix(*pool.Name, listener.Name) {
			t.Errorf("backend pool %s should have been deleted", *pool.Name)
		}
	}
	for _, probe := range appGateway.Properties.Probes {
		if probe.Name != nil && strings.HasPrefix(*probe.Name, listener.Name) {
			t.Errorf("probe %s should have been deleted", *probe.Name)
		}
	}
	for _, setting := range appGateway.Properties.BackendHTTPSettingsCollection {
		if setting.Name != nil && strings.HasPrefix(*setting.Name, listener.Name) {
			t.Errorf("backend setting %s should have been deleted", *setting.Name)
		}
	}

	if !containsResourceName(namesOfPools(appGateway), DefaultBackendPoolName) {
		t.Error("default backend pool should be preserved on delete")
	}
	if !containsResourceName(namesOfPools(appGateway), "other-li.abcd.80") {
		t.Error("unrelated backend pool should be preserved on delete")
	}
}

func namesOfPools(appGateway *armnetwork.ApplicationGateway) []string {
	names := make([]string, 0, len(appGateway.Properties.BackendAddressPools))
	for _, pool := range appGateway.Properties.BackendAddressPools {
		if pool != nil && pool.Name != nil {
			names = append(names, *pool.Name)
		}
	}
	return names
}

func containsResourceName(names []string, want string) bool {
	for _, name := range names {
		if name == want {
			return true
		}
	}
	return false
}

// TestListenerNamePrefixIsolation ensures listener "lb-80" never claims resources of "lb-8080".
// NOCC:tosa/fn_length(测试函数)
func TestListenerNamePrefixIsolation(t *testing.T) {
	alb := newTestAlb()
	shortLi := dualPathListener("appgw-80", "a.example.com", 80)
	longLi := dualPathListener("appgw-8080", "b.example.com", 8080)

	// build resources owned by the 8080 listener only
	appGateway := newEmptyAppGateway()
	appGateway = alb.ensureAddrPoolForAg(appGateway, []*networkextensionv1.Listener{longLi})
	appGateway = alb.ensureProbeForAg(appGateway, []*networkextensionv1.Listener{longLi})
	appGateway = alb.ensureBackendSettings(appGateway, []*networkextensionv1.Listener{longLi})

	longOwnedPools := 0
	for _, pool := range appGateway.Properties.BackendAddressPools {
		if pool.Name != nil && strings.HasPrefix(*pool.Name, longLi.Name+".") {
			longOwnedPools++
		}
	}
	if longOwnedPools == 0 {
		t.Fatal("test setup failed: no pool owned by appgw-8080")
	}

	// reconciling appgw-80 must not drop appgw-8080 resources
	shortOnly := []*networkextensionv1.Listener{shortLi}
	ensured := alb.ensureAddrPoolForAg(appGateway, shortOnly)
	kept := 0
	for _, pool := range ensured.Properties.BackendAddressPools {
		if pool.Name != nil && strings.HasPrefix(*pool.Name, longLi.Name+".") {
			kept++
		}
	}
	if kept != longOwnedPools {
		t.Errorf("ensure appgw-80 dropped appgw-8080 pools, kept=%d want=%d", kept, longOwnedPools)
	}

	// deleting appgw-80 must not delete appgw-8080 resources
	deleted := alb.deleteAddrPoolForAg(ensured, shortOnly)
	kept = 0
	for _, pool := range deleted.Properties.BackendAddressPools {
		if pool.Name != nil && strings.HasPrefix(*pool.Name, longLi.Name+".") {
			kept++
		}
	}
	if kept != longOwnedPools {
		t.Errorf("delete appgw-80 removed appgw-8080 pools, kept=%d want=%d", kept, longOwnedPools)
	}

	if !isAgResourceOwnedByListener(getRuleTgName(longLi.Name, "b.example.com", "/*", 8080),
		[]*networkextensionv1.Listener{longLi}) {
		t.Error("resource should be owned by its own listener")
	}
	if isAgResourceOwnedByListener(getRuleTgName(longLi.Name, "b.example.com", "/*", 8080), shortOnly) {
		t.Error("appgw-8080 resource must not be owned by appgw-80")
	}
}

// TestPriorityStableAcrossReconcile ensures repeated reconcile keeps the same priority. Priority
// drives multi site matching order and which certificate a non SNI client gets, so churning it
// would reconfigure the gateway on every pass.
// NOCC:tosa/fn_length(测试函数)
func TestPriorityStableAcrossReconcile(t *testing.T) {
	alb := newTestAlb()
	listeners := []*networkextensionv1.Listener{
		dualPathListener("agw-443", "a.example.com", 443),
	}

	appGateway := ensureFullAgListener(t, alb, listeners)
	first := *appGateway.Properties.RequestRoutingRules[0].Properties.Priority

	for round := 0; round < 3; round++ {
		appGateway = ensureFullAgListenerOn(t, alb, appGateway, listeners)
		got := *appGateway.Properties.RequestRoutingRules[0].Properties.Priority
		if got != first {
			t.Fatalf("priority drifted on round %d: got=%d want=%d", round, got, first)
		}
	}
}

// TestAutoPriorityFollowsAgicScheme locks the auto assignment scheme to AGIC's: multi site rules
// (those carrying a domain) start at 19000, basic rules at 19500, stepping by 5, so a basic listener
// never outranks a multi site one.
// NOCC:tosa/fn_length(测试函数)
func TestAutoPriorityFollowsAgicScheme(t *testing.T) {
	alb := newTestAlb()

	firstSite := dualPathListener("agw-a-443", "app.example.com", 443)
	secondSite := dualPathListener("agw-b-443", "*.example.com", 443)
	basic := dualPathListener("agw-c-443", "", 443)

	// reconcile the basic listener first, so ordering cannot come from creation order
	all := []*networkextensionv1.Listener{basic, firstSite, secondSite}
	appGateway := ensureFullAgListener(t, alb, all)

	byName := make(map[string]int32)
	for _, rule := range appGateway.Properties.RequestRoutingRules {
		byName[*rule.Name] = *rule.Properties.Priority
	}
	firstPriority := byName[getHttpListenerName(443, "app.example.com")]
	secondPriority := byName[getHttpListenerName(443, "*.example.com")]
	basicPriority := byName[getHttpListenerName(443, "")]

	if firstPriority != MultiSiteRulePriorityStart {
		t.Errorf("first multi site rule = %d, want %d", firstPriority, MultiSiteRulePriorityStart)
	}
	if secondPriority != MultiSiteRulePriorityStart+RulePriorityJump {
		t.Errorf("second multi site rule = %d, want %d", secondPriority,
			MultiSiteRulePriorityStart+RulePriorityJump)
	}
	if basicPriority != BasicRulePriorityStart {
		t.Errorf("basic rule = %d, want %d", basicPriority, BasicRulePriorityStart)
	}
	if basicPriority < firstPriority || basicPriority < secondPriority {
		t.Errorf("basic listener must not outrank a multi site one: basic=%d sites=%d/%d",
			basicPriority, firstPriority, secondPriority)
	}

	// an already assigned rule must keep its value
	appGateway = ensureFullAgListenerOn(t, alb, appGateway, all)
	for _, rule := range appGateway.Properties.RequestRoutingRules {
		if got := *rule.Properties.Priority; got != byName[*rule.Name] {
			t.Errorf("rule %s drifted from %d to %d", *rule.Name, byName[*rule.Name], got)
		}
	}
}

// TestAutoPriorityLeavesRoomForUsers covers the cross batch case: azure reconciles http and https
// listeners of one gateway separately, so a batch cannot see the priorities declared by the other.
// Auto assignment walks its band top down precisely so the small numbers users reach for stay free.
// NOCC:tosa/fn_length(测试函数)
func TestAutoPriorityLeavesRoomForUsers(t *testing.T) {
	alb := newTestAlb()

	// batch 1: http listener, no declaration
	httpLi := dualPathListener("agw-80", "a.example.com", 80)
	httpLi.Spec.Protocol = AzureProtocolHTTP
	httpLi.Spec.Certificate = nil
	appGateway := ensureFullAgListener(t, alb, []*networkextensionv1.Listener{httpLi})

	autoAssigned := *appGateway.Properties.RequestRoutingRules[0].Properties.Priority
	for _, low := range []int32{1, 10, 100} {
		if autoAssigned == low {
			t.Errorf("auto assignment took %d, the low numbers must stay free for users", low)
		}
	}

	// batch 2: https listener on the same gateway declaring a small value
	httpsLi := dualPathListener("agw-443", "b.example.com", 443)
	for i := range httpsLi.Spec.Rules {
		httpsLi.Spec.Rules[i].ListenerAttribute = &networkextensionv1.IngressListenerAttribute{
			Priority: 1,
		}
	}
	appGateway = ensureFullAgListenerOn(t, alb, appGateway, []*networkextensionv1.Listener{httpsLi})

	byName := make(map[string]int32)
	for _, rule := range appGateway.Properties.RequestRoutingRules {
		byName[*rule.Name] = *rule.Properties.Priority
	}
	if got := byName[getHttpListenerName(443, "b.example.com")]; got != 1 {
		t.Errorf("declared priority not honoured across batches: got=%d want=1", got)
	}
	if err := validateAgRulePriorityUnique(appGateway, []*networkextensionv1.Listener{httpsLi}); err != nil {
		t.Errorf("cross batch declaration collided with auto assignment: %v", err)
	}
}

// TestWildcardDomainUsesHostNames ensures a wildcard domain goes through HostNames, azure's HostName
// field only accepts a single concrete name.
// NOCC:tosa/fn_length(测试函数)
func TestWildcardDomainUsesHostNames(t *testing.T) {
	alb := newTestAlb()

	cases := []struct {
		name          string
		domain        string
		wantHostName  string
		wantHostNames []string
	}{
		{name: "concrete domain", domain: "app.example.com", wantHostName: "app.example.com"},
		{name: "wildcard domain", domain: "*.example.com", wantHostNames: []string{"*.example.com"}},
		{name: "no domain", domain: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			listener := dualPathListener("agw-443", tc.domain, 443)
			appGateway, err := alb.ensureHttpListenerForAg(newEmptyAppGateway(),
				[]*networkextensionv1.Listener{listener})
			if err != nil {
				t.Fatalf("ensureHttpListenerForAg failed: %v", err)
			}
			props := appGateway.Properties.HTTPListeners[0].Properties

			gotHostName := ""
			if props.HostName != nil {
				gotHostName = *props.HostName
			}
			if gotHostName != tc.wantHostName {
				t.Errorf("HostName = %q, want %q", gotHostName, tc.wantHostName)
			}
			if len(props.HostNames) != len(tc.wantHostNames) {
				t.Fatalf("HostNames = %d entries, want %d", len(props.HostNames), len(tc.wantHostNames))
			}
			for i, want := range tc.wantHostNames {
				if *props.HostNames[i] != want {
					t.Errorf("HostNames[%d] = %q, want %q", i, *props.HostNames[i], want)
				}
			}
			// azure rejects a listener carrying both fields
			if props.HostName != nil && len(props.HostNames) != 0 {
				t.Error("HostName and HostNames are mutually exclusive")
			}
		})
	}
}

// TestUserPriorityWins ensures a priority declared through listenerAttribute is what reaches azure,
// and that it stays put across reconciles.
// NOCC:tosa/fn_length(测试函数)
func TestUserPriorityWins(t *testing.T) {
	alb := newTestAlb()
	listener := dualPathListener("agw-443", "a.example.com", 443)
	for i := range listener.Spec.Rules {
		listener.Spec.Rules[i].ListenerAttribute = &networkextensionv1.IngressListenerAttribute{
			Priority: 100,
		}
	}
	listeners := []*networkextensionv1.Listener{listener}

	appGateway := ensureFullAgListener(t, alb, listeners)
	if got := *appGateway.Properties.RequestRoutingRules[0].Properties.Priority; got != 100 {
		t.Fatalf("user declared priority ignored: got=%d want=100", got)
	}

	appGateway = ensureFullAgListenerOn(t, alb, appGateway, listeners)
	if got := *appGateway.Properties.RequestRoutingRules[0].Properties.Priority; got != 100 {
		t.Errorf("user declared priority drifted: got=%d want=100", got)
	}
}

// TestAutoPriorityAvoidsUserValue ensures an auto assigned rule never takes a priority that another
// rule of the same batch explicitly asked for.
// NOCC:tosa/fn_length(测试函数)
func TestAutoPriorityAvoidsUserValue(t *testing.T) {
	alb := newTestAlb()

	// auto assigned rule would normally take priority 1, but the peer reserves it
	auto := dualPathListener("agw-a-443", "auto.example.com", 443)
	reservedLi := dualPathListener("agw-b-443", "fixed.example.com", 443)
	for i := range reservedLi.Spec.Rules {
		reservedLi.Spec.Rules[i].ListenerAttribute = &networkextensionv1.IngressListenerAttribute{
			Priority: 1,
		}
	}

	both := []*networkextensionv1.Listener{auto, reservedLi}
	appGateway := ensureFullAgListener(t, alb, both)

	byName := make(map[string]int32)
	for _, rule := range appGateway.Properties.RequestRoutingRules {
		byName[*rule.Name] = *rule.Properties.Priority
	}
	if got := byName[getHttpListenerName(443, "fixed.example.com")]; got != 1 {
		t.Errorf("reserved priority not honoured: got=%d want=1", got)
	}
	if got := byName[getHttpListenerName(443, "auto.example.com")]; got == 1 {
		t.Error("auto assigned rule stole the reserved priority 1")
	}
	if err := validateAgRulePriorityUnique(appGateway, both); err != nil {
		t.Errorf("priorities must stay unique: %v", err)
	}
}

// TestUserPriorityConflictDegrades ensures a bad declaration falls back to auto assignment instead
// of failing the whole batch. The webhook is where such config is rejected; at reconcile time one
// stale object must not stop every layer 7 listener of a shared gateway from syncing.
// NOCC:tosa/fn_length(测试函数)
func TestUserPriorityConflictDegrades(t *testing.T) {
	withPriority := func(name, domain string, priorities ...int) *networkextensionv1.Listener {
		li := dualPathListener(name, domain, 443)
		for i := range li.Spec.Rules {
			li.Spec.Rules[i].ListenerAttribute = &networkextensionv1.IngressListenerAttribute{
				Priority: priorities[i%len(priorities)],
			}
		}
		return li
	}

	cases := []struct {
		name      string
		listeners []*networkextensionv1.Listener
		// how many rule names should end up with a usable declared priority
		wantKept int
	}{
		{
			name:      "same domain different priority keeps the first",
			listeners: []*networkextensionv1.Listener{withPriority("agw-443", "a.example.com", 10, 20)},
			wantKept:  1,
		},
		{
			name: "two domains same priority keeps one",
			listeners: []*networkextensionv1.Listener{
				withPriority("agw-a-443", "a.example.com", 10),
				withPriority("agw-b-443", "b.example.com", 10),
			},
			wantKept: 1,
		},
		{
			name:      "out of range is dropped",
			listeners: []*networkextensionv1.Listener{withPriority("agw-443", "a.example.com", 20001)},
			wantKept:  0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := collectUserRulePriorities(tc.listeners)
			if len(got) != tc.wantKept {
				t.Errorf("kept %d declarations (%v), want %d", len(got), got, tc.wantKept)
			}
			for _, priority := range got {
				if priority < 1 || priority > MaxRoutingRulePriority {
					t.Errorf("kept an unusable priority %d", priority)
				}
			}
		})
	}
}

// TestConflictingDeclarationStillSyncs ensures the whole ensure chain still produces a valid gateway
// when an ingress carries a conflicting declaration, mirroring a portMapping fanned out over ports.
// NOCC:tosa/fn_length(测试函数)
func TestConflictingDeclarationStillSyncs(t *testing.T) {
	alb := newTestAlb()
	domain := "app.example.com"

	// same domain, same declared priority, two ports: exactly what a portMapping expands into
	first := dualPathListener("agw-30000", domain, 30000)
	second := dualPathListener("agw-30001", domain, 30001)
	for _, li := range []*networkextensionv1.Listener{first, second} {
		for i := range li.Spec.Rules {
			li.Spec.Rules[i].ListenerAttribute = &networkextensionv1.IngressListenerAttribute{
				Priority: 100,
			}
		}
	}
	both := []*networkextensionv1.Listener{first, second}

	appGateway := ensureFullAgListener(t, alb, both)

	if len(appGateway.Properties.RequestRoutingRules) != 2 {
		t.Fatalf("expect 2 routing rules, got %d", len(appGateway.Properties.RequestRoutingRules))
	}
	if err := validateAgRulePriorityUnique(appGateway, both); err != nil {
		t.Errorf("conflicting declaration must degrade, not produce an invalid gateway: %v", err)
	}
}

// TestValidateIngressPriorities covers the admission side checks for declared priorities, including
// duplicates that sit on different ports of the same gateway.
// NOCC:tosa/fn_length(测试函数)
func TestValidateIngressPriorities(t *testing.T) {
	validater := &AlbValidater{}
	route := func(domain, path string, priority int) networkextensionv1.Layer7Route {
		r := networkextensionv1.Layer7Route{Domain: domain, Path: path}
		if priority != 0 {
			r.ListenerAttribute = &networkextensionv1.IngressListenerAttribute{Priority: priority}
		}
		return r
	}
	ingress := func(rules ...networkextensionv1.IngressRule) *networkextensionv1.Ingress {
		return &networkextensionv1.Ingress{Spec: networkextensionv1.IngressSpec{Rules: rules}}
	}
	rule := func(port int, routes ...networkextensionv1.Layer7Route) networkextensionv1.IngressRule {
		return networkextensionv1.IngressRule{Port: port, Protocol: AzureProtocolHTTPS, Routes: routes}
	}

	cases := []struct {
		name    string
		ingress *networkextensionv1.Ingress
		valid   bool
	}{
		{
			name:    "same domain same priority is fine",
			ingress: ingress(rule(443, route("a.com", "/x", 10), route("a.com", "/y", 10))),
			valid:   true,
		},
		{
			name:    "only one path declares it is fine",
			ingress: ingress(rule(443, route("a.com", "/x", 10), route("a.com", "/y", 0))),
			valid:   true,
		},
		{
			name:    "same domain conflicting priority",
			ingress: ingress(rule(443, route("a.com", "/x", 10), route("a.com", "/y", 20))),
			valid:   false,
		},
		{
			name:    "two domains share a priority",
			ingress: ingress(rule(443, route("a.com", "/x", 10), route("b.com", "/y", 10))),
			valid:   false,
		},
		{
			name: "duplicate across ports is rejected too",
			ingress: ingress(rule(80, route("a.com", "/x", 10)),
				rule(8080, route("b.com", "/y", 10))),
			valid: false,
		},
		{
			name: "different ports different priorities is fine",
			ingress: ingress(rule(80, route("a.com", "/x", 10)),
				rule(8080, route("b.com", "/y", 11))),
			valid: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, msg := validater.validateIngressPriorities(tc.ingress)
			if got != tc.valid {
				t.Errorf("validateIngressPriorities = %v (%s), want %v", got, msg, tc.valid)
			}
		})
	}
}

// TestPriorityAttributeRangeAndScope covers the range check and the places priority is not allowed.
// NOCC:tosa/fn_length(测试函数)
func TestPriorityAttributeRangeAndScope(t *testing.T) {
	validater := &AlbValidater{}

	if ok, _ := validater.validateAgListenerAttribute(
		&networkextensionv1.IngressListenerAttribute{Priority: 20001}); ok {
		t.Error("priority above 20000 must be rejected")
	}
	if ok, msg := validater.validateAgListenerAttribute(
		&networkextensionv1.IngressListenerAttribute{Priority: 20000}); !ok {
		t.Errorf("priority 20000 must be accepted: %s", msg)
	}
	if ok, _ := validater.validateAgListenerAttribute(
		&networkextensionv1.IngressListenerAttribute{Priority: -1}); ok {
		t.Error("negative priority must be rejected")
	}

	// a portMapping fans out to many ports, one priority cannot serve them all
	ok, msg := validater.validatePortMappingRoute(&networkextensionv1.IngressPortMappingLayer7Route{
		ListenerAttribute: &networkextensionv1.IngressListenerAttribute{Priority: 100},
	})
	if ok {
		t.Error("priority in portMappings must be rejected")
	}
	if !strings.Contains(msg, "spec.rules") {
		t.Errorf("error should point at the supported place, got: %s", msg)
	}
}

// TestExistingPriorityYieldsToUser ensures a sibling rule gives up its auto assigned value when the
// user claims it, otherwise both rules would collide and the gateway update would be rejected
// forever.
// NOCC:tosa/fn_length(测试函数)
func TestExistingPriorityYieldsToUser(t *testing.T) {
	alb := newTestAlb()
	peer := dualPathListener("agw-a-443", "a.example.com", 443)
	claimer := dualPathListener("agw-b-443", "b.example.com", 443)
	both := []*networkextensionv1.Listener{peer, claimer}

	// first round: both get auto assigned priorities
	appGateway := ensureFullAgListener(t, alb, both)
	peerName := getHttpListenerName(443, "a.example.com")
	claimerName := getHttpListenerName(443, "b.example.com")

	peerAuto := int32(0)
	for _, r := range appGateway.Properties.RequestRoutingRules {
		if *r.Name == peerName {
			peerAuto = *r.Properties.Priority
		}
	}
	if peerAuto == 0 {
		t.Fatal("test setup failed: peer rule has no priority")
	}

	// second round: the user claims exactly the value the peer holds
	for i := range claimer.Spec.Rules {
		claimer.Spec.Rules[i].ListenerAttribute = &networkextensionv1.IngressListenerAttribute{
			Priority: int(peerAuto),
		}
	}
	appGateway = ensureFullAgListenerOn(t, alb, appGateway, both)

	byName := make(map[string]int32)
	for _, r := range appGateway.Properties.RequestRoutingRules {
		byName[*r.Name] = *r.Properties.Priority
	}
	if byName[claimerName] != peerAuto {
		t.Errorf("user claim not honoured: got=%d want=%d", byName[claimerName], peerAuto)
	}
	if byName[peerName] == peerAuto {
		t.Errorf("peer rule should have moved off priority %d", peerAuto)
	}
	if err := validateAgRulePriorityUnique(appGateway, both); err != nil {
		t.Errorf("priorities must stay unique: %v", err)
	}
}

// TestBackfillPriorityForLegacyRules ensures a gateway that still holds rules without priority is
// not left in a state azure rejects with ApplicationGatewayRequestRoutingRulePartialPriorityDefined.
// NOCC:tosa/fn_length(测试函数)
func TestBackfillPriorityForLegacyRules(t *testing.T) {
	appGateway := newEmptyAppGateway()
	appGateway.Properties.RequestRoutingRules = []*armnetwork.ApplicationGatewayRequestRoutingRule{
		{Name: to.StringPtr("legacy-no-priority"), Properties: &armnetwork.
			ApplicationGatewayRequestRoutingRulePropertiesFormat{}},
		{Name: to.StringPtr("managed"), Properties: &armnetwork.
			ApplicationGatewayRequestRoutingRulePropertiesFormat{Priority: to.Int32Ptr(1)}},
		{Name: to.StringPtr("legacy-no-priority-2"), Properties: &armnetwork.
			ApplicationGatewayRequestRoutingRulePropertiesFormat{}},
	}

	if err := backfillRoutingRulePriorities(appGateway); err != nil {
		t.Fatalf("backfill failed: %v", err)
	}

	seen := make(map[int32]string)
	for _, rule := range appGateway.Properties.RequestRoutingRules {
		if rule.Properties.Priority == nil {
			t.Fatalf("rule %s still has no priority", *rule.Name)
		}
		priority := *rule.Properties.Priority
		if priority < 1 || priority > MaxRoutingRulePriority {
			t.Errorf("rule %s got out of range priority %d", *rule.Name, priority)
		}
		if other, dup := seen[priority]; dup {
			t.Errorf("rules %s and %s share priority %d", other, *rule.Name, priority)
		}
		seen[priority] = *rule.Name
	}
}

// TestHTTPSListenerEnablesSNI ensures a multi site https listener asks azure for SNI, without it
// several domains cannot share one frontend port.
// NOCC:tosa/fn_length(测试函数)
func TestHTTPSListenerEnablesSNI(t *testing.T) {
	alb := newTestAlb()
	httpsListener := dualPathListener("agw-443", "a.example.com", 443)
	appGateway, err := alb.ensureHttpListenerForAg(newEmptyAppGateway(),
		[]*networkextensionv1.Listener{httpsListener})
	if err != nil {
		t.Fatalf("ensureHttpListenerForAg failed: %v", err)
	}
	sni := appGateway.Properties.HTTPListeners[0].Properties.RequireServerNameIndication
	if sni == nil || !*sni {
		t.Error("multi site https listener must enable RequireServerNameIndication")
	}

	// http listeners must not carry the flag, azure only accepts it for https
	httpListener := dualPathListener("agw-80", "a.example.com", 80)
	httpListener.Spec.Protocol = AzureProtocolHTTP
	httpListener.Spec.Certificate = nil
	appGateway, err = alb.ensureHttpListenerForAg(newEmptyAppGateway(),
		[]*networkextensionv1.Listener{httpListener})
	if err != nil {
		t.Fatalf("ensureHttpListenerForAg failed: %v", err)
	}
	if appGateway.Properties.HTTPListeners[0].Properties.RequireServerNameIndication != nil {
		t.Error("http listener must not set RequireServerNameIndication")
	}
}

// TestProbeUnhealthyThresholdDefault ensures an enabled health check without unHealthNum still
// sends a value azure accepts (1~20) instead of 0.
// NOCC:tosa/fn_length(测试函数)
func TestProbeUnhealthyThresholdDefault(t *testing.T) {
	alb := newTestAlb()
	listener := dualPathListener("agw-443", "a.example.com", 443)
	listener.Spec.Rules[0].ListenerAttribute = &networkextensionv1.IngressListenerAttribute{
		HealthCheck: &networkextensionv1.ListenerHealthCheck{Enabled: true},
	}

	appGateway := alb.ensureProbeForAg(newEmptyAppGateway(),
		[]*networkextensionv1.Listener{listener})

	found := false
	for _, probe := range appGateway.Properties.Probes {
		found = true
		threshold := *probe.Properties.UnhealthyThreshold
		if threshold < 1 || threshold > 20 {
			t.Errorf("probe %s has unhealthyThreshold %d, azure accepts 1~20",
				*probe.Name, threshold)
		}
	}
	if !found {
		t.Fatal("test setup failed: no probe generated")
	}
}

// TestGeneratePriorityTolerateNil ensures rules without priority do not panic priority allocation.
// NOCC:tosa/fn_length(测试函数)
func TestGeneratePriorityTolerateNil(t *testing.T) {
	appGateway := newEmptyAppGateway()
	appGateway.Properties.RequestRoutingRules = []*armnetwork.ApplicationGatewayRequestRoutingRule{
		// user created rule on a v1 SKU gateway has no priority
		{Name: to.StringPtr("legacy"), Properties: &armnetwork.
			ApplicationGatewayRequestRoutingRulePropertiesFormat{}},
		// out of range priority must be ignored instead of panicking
		{Name: to.StringPtr("weird"), Properties: &armnetwork.
			ApplicationGatewayRequestRoutingRulePropertiesFormat{Priority: to.Int32Ptr(99999)}},
		{Name: to.StringPtr("normal"), Properties: &armnetwork.
			ApplicationGatewayRequestRoutingRulePropertiesFormat{Priority: to.Int32Ptr(1)}},
		nil,
	}

	if got := generatePriority(appGateway); got != 2 {
		t.Errorf("unexpected priority got=%d want=2", got)
	}
}

// TestDeleteHttpListenerNilSafe ensures delete tolerates rules/listeners without IDs.
// NOCC:tosa/fn_length(测试函数)
func TestDeleteHttpListenerNilSafe(t *testing.T) {
	alb := newTestAlb()
	appGateway := newEmptyAppGateway()
	appGateway.Properties.RequestRoutingRules = []*armnetwork.ApplicationGatewayRequestRoutingRule{
		{Name: to.StringPtr("no-props")},
		{Name: to.StringPtr("no-listener"), Properties: &armnetwork.
			ApplicationGatewayRequestRoutingRulePropertiesFormat{}},
	}
	appGateway.Properties.HTTPListeners = []*armnetwork.ApplicationGatewayHTTPListener{
		{Name: to.StringPtr("443.abcd")},
	}

	listeners := []*networkextensionv1.Listener{
		dualPathListener("li-443", "a.example.com", 443),
	}

	result := alb.deleteHttpListenerForAg(appGateway, listeners)
	if len(result.Properties.HTTPListeners) != 1 {
		t.Fatalf("expect listener without ID kept, got %d", len(result.Properties.HTTPListeners))
	}
}

// TestCleanupStaleRoutesOnShrink ensures removing a path from the spec also removes its pathRule,
// so no route keeps referencing a backend pool that ensure* already dropped.
// NOCC:tosa/fn_length(测试函数)
func TestCleanupStaleRoutesOnShrink(t *testing.T) {
	alb := newTestAlb()
	domain := "app.example.com"
	full := dualPathListener("agw-443", domain, 443)

	appGateway := ensureFullAgListener(t, alb, []*networkextensionv1.Listener{full})
	if got := len(appGateway.Properties.URLPathMaps[0].Properties.PathRules); got != 2 {
		t.Fatalf("test setup failed: expect 2 path rules, got %d", got)
	}
	staleTgName := getRuleTgName(full.Name, domain, "/api/specific/*", 443)

	// spec shrinks to a single path, reconciled against the gateway state built above
	shrunk := full.DeepCopy()
	shrunk.Spec.Rules = shrunk.Spec.Rules[1:]
	appGateway = ensureFullAgListenerOn(t, alb, appGateway, []*networkextensionv1.Listener{shrunk})

	pathRules := appGateway.Properties.URLPathMaps[0].Properties.PathRules
	if len(pathRules) != 1 {
		t.Fatalf("expect 1 path rule after shrink, got %d", len(pathRules))
	}
	for _, pathRule := range pathRules {
		if subResourceName(pathRule.Properties.BackendAddressPool) == staleTgName {
			t.Errorf("path rule still references removed pool %s", staleTgName)
		}
	}

	poolNames := namesOfPools(appGateway)
	if containsResourceName(poolNames, staleTgName) {
		t.Errorf("stale backend pool %s should have been removed", staleTgName)
	}

	assertNoDanglingBackendRefs(t, appGateway)
}

// TestCleanupKeepsSharedRouteOfPeer ensures shrinking one listener does not tear down the routing
// rule / http listener shared with another listener on the same port and domain.
// NOCC:tosa/fn_length(测试函数)
func TestCleanupKeepsSharedRouteOfPeer(t *testing.T) {
	alb := newTestAlb()
	domain := "shared.example.com"

	liA := dualPathListener("agw-a-443", domain, 443)
	liA.Spec.Rules = liA.Spec.Rules[:1] // only /api/specific/*
	liB := dualPathListener("agw-b-443", domain, 443)
	liB.Spec.Rules = liB.Spec.Rules[1:] // only /*

	// both listeners reconciled onto the same gateway, sharing port+domain
	appGateway := ensureFullAgListener(t, alb, []*networkextensionv1.Listener{liB})
	appGateway = ensureFullAgListenerOn(t, alb, appGateway, []*networkextensionv1.Listener{liA})

	sharedName := getHttpListenerName(443, domain)
	peerTgName := getRuleTgName(liB.Name, domain, "/*", 443)

	// listener A drops all of its routes, then reconciles alone
	emptyA := liA.DeepCopy()
	emptyA.Spec.Rules = nil
	appGateway = ensureFullAgListenerOn(t, alb, appGateway, []*networkextensionv1.Listener{emptyA})

	if !containsResourceName(namesOfRoutingRules(appGateway), sharedName) {
		t.Errorf("shared routing rule %s was removed while peer listener still uses it", sharedName)
	}
	if !containsResourceName(namesOfHTTPListeners(appGateway), sharedName) {
		t.Errorf("shared http listener %s was removed while peer listener still uses it", sharedName)
	}

	peerPathRuleKept := false
	for _, pathMap := range appGateway.Properties.URLPathMaps {
		for _, pathRule := range pathMap.Properties.PathRules {
			if subResourceName(pathRule.Properties.BackendAddressPool) == peerTgName {
				peerPathRuleKept = true
			}
		}
	}
	if !peerPathRuleKept {
		t.Errorf("peer path rule referencing %s should be kept", peerTgName)
	}

	// the shared rule must not keep pointing at listener A's deleted pool
	assertNoDanglingBackendRefs(t, appGateway)
	for _, rule := range appGateway.Properties.RequestRoutingRules {
		if rule.Name != nil && *rule.Name != sharedName {
			continue
		}
		if isAgResourceOwnedByListener(subResourceName(rule.Properties.BackendAddressPool),
			[]*networkextensionv1.Listener{liA}) {
			t.Errorf("shared rule still points at removed listener A pool %s",
				subResourceName(rule.Properties.BackendAddressPool))
		}
	}
}

func namesOfRoutingRules(appGateway *armnetwork.ApplicationGateway) []string {
	names := make([]string, 0, len(appGateway.Properties.RequestRoutingRules))
	for _, rule := range appGateway.Properties.RequestRoutingRules {
		if rule != nil && rule.Name != nil {
			names = append(names, *rule.Name)
		}
	}
	return names
}

func namesOfHTTPListeners(appGateway *armnetwork.ApplicationGateway) []string {
	names := make([]string, 0, len(appGateway.Properties.HTTPListeners))
	for _, httpListener := range appGateway.Properties.HTTPListeners {
		if httpListener != nil && httpListener.Name != nil {
			names = append(names, *httpListener.Name)
		}
	}
	return names
}

// TestCleanupKeepsManualPathMap ensures a manually created path map without pathRules survives.
// NOCC:tosa/fn_length(测试函数)
func TestCleanupKeepsManualPathMap(t *testing.T) {
	alb := newTestAlb()
	listener := dualPathListener("agw-443", "a.example.com", 443)
	appGateway := ensureFullAgListener(t, alb, []*networkextensionv1.Listener{listener})

	appGateway.Properties.URLPathMaps = append(appGateway.Properties.URLPathMaps,
		&armnetwork.ApplicationGatewayURLPathMap{
			Name: to.StringPtr("manual-path-map"),
			Properties: &armnetwork.ApplicationGatewayURLPathMapPropertiesFormat{
				PathRules: []*armnetwork.ApplicationGatewayPathRule{},
			},
		})

	appGateway = cleanupStaleAgRoutes(appGateway, []*networkextensionv1.Listener{listener})

	found := false
	for _, pathMap := range appGateway.Properties.URLPathMaps {
		if pathMap.Name != nil && *pathMap.Name == "manual-path-map" {
			found = true
		}
	}
	if !found {
		t.Error("manually created url path map without path rules must not be removed")
	}
}

// TestCleanupKeepsManualListener ensures manually created gateway listeners survive cleanup.
// NOCC:tosa/fn_length(测试函数)
func TestCleanupKeepsManualListener(t *testing.T) {
	alb := newTestAlb()
	listener := dualPathListener("agw-443", "a.example.com", 443)
	appGateway := ensureFullAgListener(t, alb, []*networkextensionv1.Listener{listener})

	manual := &armnetwork.ApplicationGatewayHTTPListener{Name: to.StringPtr("my-manual-listener")}
	appGateway.Properties.HTTPListeners = append(appGateway.Properties.HTTPListeners, manual)

	appGateway = cleanupStaleAgRoutes(appGateway, []*networkextensionv1.Listener{listener})

	found := false
	for _, httpListener := range appGateway.Properties.HTTPListeners {
		if httpListener.Name != nil && *httpListener.Name == "my-manual-listener" {
			found = true
		}
	}
	if !found {
		t.Error("manually created http listener must not be removed by cleanup")
	}
}

// ensureFullAgListener runs the whole L7 ensure chain and asserts child names stay unique.
func ensureFullAgListener(t *testing.T, alb *Alb,
	listeners []*networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	t.Helper()
	return ensureFullAgListenerOn(t, alb, newEmptyAppGateway(), listeners)
}

func ensureFullAgListenerOn(t *testing.T, alb *Alb, appGateway *armnetwork.ApplicationGateway,
	listeners []*networkextensionv1.Listener) *armnetwork.ApplicationGateway {
	t.Helper()

	appGateway = alb.ensureFrontendPortForAg(appGateway, listeners)
	appGateway = alb.ensureAddrPoolForAg(appGateway, listeners)
	appGateway = alb.ensureProbeForAg(appGateway, listeners)
	appGateway = alb.ensureBackendSettings(appGateway, listeners)

	appGateway, err := alb.ensureHttpListenerForAg(appGateway, listeners)
	if err != nil {
		t.Fatalf("ensureHttpListenerForAg failed: %v", err)
	}
	appGateway = alb.ensureUrlPathMap(appGateway, listeners)
	appGateway, err = alb.ensureRequestRoutingRule(appGateway, listeners)
	if err != nil {
		t.Fatalf("ensureRequestRoutingRule failed: %v", err)
	}
	appGateway = cleanupStaleAgRoutes(appGateway, listeners)
	appGateway = alb.repairAgDanglingRefs(appGateway, listeners)

	if err = validateAgChildNamesUnique(appGateway); err != nil {
		t.Fatalf("duplicate AGW child names: %v", err)
	}
	if err = validateAgNoDanglingRefs(appGateway, listeners); err != nil {
		t.Fatalf("dangling AGW reference: %v", err)
	}
	return appGateway
}

// assertNoDanglingBackendRefs checks every route points at a backend pool and setting that exists.
func assertNoDanglingBackendRefs(t *testing.T, appGateway *armnetwork.ApplicationGateway) {
	t.Helper()
	poolNames := namesOfPools(appGateway)
	settingNames := namesOfSettings(appGateway)

	check := func(owner, kind, name string, existing []string) {
		if name != "" && !containsResourceName(existing, name) {
			t.Errorf("%s references missing %s %s", owner, kind, name)
		}
	}

	for _, rule := range appGateway.Properties.RequestRoutingRules {
		check("routing rule", "pool", subResourceName(rule.Properties.BackendAddressPool), poolNames)
		check("routing rule", "setting", subResourceName(rule.Properties.BackendHTTPSettings), settingNames)
	}
	for _, pathMap := range appGateway.Properties.URLPathMaps {
		for _, pathRule := range pathMap.Properties.PathRules {
			check("path rule", "pool", subResourceName(pathRule.Properties.BackendAddressPool), poolNames)
			check("path rule", "setting",
				subResourceName(pathRule.Properties.BackendHTTPSettings), settingNames)
		}
	}
}

func namesOfSettings(appGateway *armnetwork.ApplicationGateway) []string {
	names := make([]string, 0, len(appGateway.Properties.BackendHTTPSettingsCollection))
	for _, setting := range appGateway.Properties.BackendHTTPSettingsCollection {
		if setting != nil && setting.Name != nil {
			names = append(names, *setting.Name)
		}
	}
	return names
}

// TestDeleteSharedListenerKeepsPeer ensures deleting one of two listeners sharing a port+domain
// leaves the peer route intact and free of dangling backend references.
// NOCC:tosa/fn_length(测试函数)
func TestDeleteSharedListenerKeepsPeer(t *testing.T) {
	alb := newTestAlb()
	domain := "shared.example.com"

	liA := dualPathListener("agw-a-443", domain, 443)
	liA.Spec.Rules = liA.Spec.Rules[:1]
	liB := dualPathListener("agw-b-443", domain, 443)
	liB.Spec.Rules = liB.Spec.Rules[1:]

	appGateway := ensureFullAgListener(t, alb, []*networkextensionv1.Listener{liB})
	appGateway = ensureFullAgListenerOn(t, alb, appGateway, []*networkextensionv1.Listener{liA})

	sharedName := getHttpListenerName(443, domain)
	peerTgName := getRuleTgName(liB.Name, domain, "/*", 443)

	// listener A is deleted while its spec still declares rules
	onlyA := []*networkextensionv1.Listener{liA}
	appGateway = alb.deleteAddrPoolForAg(appGateway, onlyA)
	appGateway = alb.deleteProbeForAg(appGateway, onlyA)
	appGateway = alb.deleteBackendSettingsForAg(appGateway, onlyA)
	appGateway = alb.deleteURLPathMapForAg(appGateway, onlyA)
	appGateway = alb.deleteRoutingRuleForAg(appGateway, onlyA)
	appGateway = alb.deleteHttpListenerForAg(appGateway, onlyA)
	appGateway = alb.repairAgDanglingRefs(appGateway, onlyA)

	if err := validateAgNoDanglingRefs(appGateway, onlyA); err != nil {
		t.Fatalf("delete left a dangling reference: %v", err)
	}
	assertNoDanglingBackendRefs(t, appGateway)

	if !containsResourceName(namesOfRoutingRules(appGateway), sharedName) {
		t.Errorf("shared routing rule %s removed while peer listener still uses it", sharedName)
	}
	if !containsResourceName(namesOfHTTPListeners(appGateway), sharedName) {
		t.Errorf("shared http listener %s removed while peer listener still uses it", sharedName)
	}
	if !containsResourceName(namesOfPools(appGateway), peerTgName) {
		t.Errorf("peer backend pool %s must survive deletion of listener A", peerTgName)
	}
}

// TestRuleBackendMatchesAzureModel locks the Azure reference model: a path based routing rule
// routes only through its URLPathMap and must not carry a rule level backend, while a basic rule
// must carry one. See https://learn.microsoft.com/en-us/azure/application-gateway/url-route-overview
// NOCC:tosa/fn_length(测试函数)
func TestRuleBackendMatchesAzureModel(t *testing.T) {
	alb := newTestAlb()
	domain := "a.example.com"

	pathBased := dualPathListener("agw-443", domain, 443)
	appGateway := ensureFullAgListener(t, alb, []*networkextensionv1.Listener{pathBased})
	if len(appGateway.Properties.RequestRoutingRules) != 1 {
		t.Fatalf("expect 1 routing rule, got %d", len(appGateway.Properties.RequestRoutingRules))
	}
	rule := appGateway.Properties.RequestRoutingRules[0]
	if *rule.Properties.RuleType != armnetwork.ApplicationGatewayRequestRoutingRuleTypePathBasedRouting {
		t.Fatalf("expect path based rule, got %s", *rule.Properties.RuleType)
	}
	if rule.Properties.BackendAddressPool != nil {
		t.Errorf("path based rule must not set rule level backend pool, got %s",
			subResourceName(rule.Properties.BackendAddressPool))
	}
	if rule.Properties.BackendHTTPSettings != nil {
		t.Errorf("path based rule must not set rule level backend setting, got %s",
			subResourceName(rule.Properties.BackendHTTPSettings))
	}
	if subResourceName(rule.Properties.URLPathMap) != getHttpListenerName(443, domain) {
		t.Errorf("path based rule must reference its url path map, got %s",
			subResourceName(rule.Properties.URLPathMap))
	}
	// the url path map must still provide a default backend for unmatched requests
	pathMap := appGateway.Properties.URLPathMaps[0]
	if subResourceName(pathMap.Properties.DefaultBackendAddressPool) != DefaultBackendPoolName {
		t.Errorf("url path map must keep a default backend pool, got %s",
			subResourceName(pathMap.Properties.DefaultBackendAddressPool))
	}

	// a rule without path stays basic and keeps its authoritative backend
	basic := dualPathListener("agw-80", domain, 80)
	basic.Spec.Rules = basic.Spec.Rules[1:]
	basic.Spec.Rules[0].Path = ""
	basicGateway := ensureFullAgListener(t, alb, []*networkextensionv1.Listener{basic})
	basicRule := basicGateway.Properties.RequestRoutingRules[0]
	if *basicRule.Properties.RuleType != armnetwork.ApplicationGatewayRequestRoutingRuleTypeBasic {
		t.Fatalf("expect basic rule, got %s", *basicRule.Properties.RuleType)
	}
	if basicRule.Properties.BackendAddressPool == nil || basicRule.Properties.BackendHTTPSettings == nil {
		t.Error("basic rule must carry its rule level backend")
	}
	if basicRule.Properties.URLPathMap != nil {
		t.Error("basic rule must not reference a url path map")
	}
}

// TestRepairSkipsBasicRoutingRule ensures a basic rule's authoritative backend is never silently
// rewritten to the default pool, the dangling reference is reported instead.
// NOCC:tosa/fn_length(测试函数)
func TestRepairSkipsBasicRoutingRule(t *testing.T) {
	alb := newTestAlb()
	listener := dualPathListener("agw-443", "a.example.com", 443)
	listeners := []*networkextensionv1.Listener{listener}
	goneTgName := getRuleTgName(listener.Name, "a.example.com", "/gone", 443)

	appGateway := newEmptyAppGateway()
	appGateway = alb.ensureAddrPoolForAg(appGateway, listeners)
	appGateway = alb.ensureBackendSettings(appGateway, listeners)

	basicRule := &armnetwork.ApplicationGatewayRequestRoutingRule{
		Name: to.StringPtr("443.basic"),
		Properties: &armnetwork.ApplicationGatewayRequestRoutingRulePropertiesFormat{
			RuleType: ruleTypeBasicPtr(),
			BackendAddressPool: alb.resourceHelper.genSubResource(ResourceProviderApplicationGateway,
				testAppGatewayName, ResourceTypeBackendAddressPools, goneTgName),
		},
	}
	appGateway.Properties.RequestRoutingRules = []*armnetwork.ApplicationGatewayRequestRoutingRule{basicRule}

	appGateway = alb.repairAgDanglingRefs(appGateway, listeners)

	if subResourceName(basicRule.Properties.BackendAddressPool) != goneTgName {
		t.Errorf("basic rule backend was silently rewritten to %s",
			subResourceName(basicRule.Properties.BackendAddressPool))
	}
	if err := validateAgNoDanglingRefs(appGateway, listeners); err == nil {
		t.Error("expect dangling reference of basic rule to be reported")
	}
}

func ruleTypeBasicPtr() *armnetwork.ApplicationGatewayRequestRoutingRuleType {
	ruleType := armnetwork.ApplicationGatewayRequestRoutingRuleTypeBasic
	return &ruleType
}

// TestValidateAgNameLength fails fast when a long gateway name pushes generated child names past
// the azure 80 character limit, instead of letting azure answer with an opaque 400.
// NOCC:tosa/fn_length(测试函数)
func TestValidateAgNameLength(t *testing.T) {
	alb := newTestAlb()
	longLbID := strings.Repeat("a", 45)
	listener := dualPathListener(longLbID+"-443", "a.example.com", 443)
	listener.Spec.LoadbalancerID = longLbID
	listeners := []*networkextensionv1.Listener{listener}

	appGateway := newEmptyAppGateway()
	appGateway = alb.ensureAddrPoolForAg(appGateway, listeners)

	tgName := getRuleTgName(listener.Name, "a.example.com", "/api/specific/*", 443)
	if len(tgName) <= MaxAzureResourceNameLen {
		t.Fatalf("test setup failed: generated name is only %d characters", len(tgName))
	}

	err := validateAgChildNamesUnique(appGateway)
	if err == nil {
		t.Fatal("expect over long child resource name to be rejected")
	}
	if !strings.Contains(err.Error(), "azure allows at most") {
		t.Errorf("unexpected error message: %v", err)
	}

	// a normal length gateway name stays valid
	normal := dualPathListener("agw-443", "a.example.com", 443)
	normalGateway := ensureFullAgListener(t, alb, []*networkextensionv1.Listener{normal})
	if err = validateAgChildNamesUnique(normalGateway); err != nil {
		t.Errorf("normal gateway name should pass: %v", err)
	}
}

// TestValidateLBRuleEndpointsUnique covers the layer 4 constraint that replaces priority: an azure
// load balancer has no evaluation order, it requires frontend+protocol+port to be unique. TCP and
// UDP may share a port because the protocol tells them apart, which is what port reuse relies on.
// NOCC:tosa/fn_length(测试函数)
func TestValidateLBRuleEndpointsUnique(t *testing.T) {
	alb := newTestAlb()
	frontend := alb.resourceHelper.getSubResourceByID(testIDPrefix() +
		"/frontendIPConfigurations/frontend-ip")
	tcp := armnetwork.TransportProtocolTCP
	udp := armnetwork.TransportProtocolUDP

	rule := func(name string, protocol armnetwork.TransportProtocol, port int32,
	) *armnetwork.LoadBalancingRule {
		return &armnetwork.LoadBalancingRule{
			Name: to.StringPtr(name),
			Properties: &armnetwork.LoadBalancingRulePropertiesFormat{
				FrontendIPConfiguration: frontend,
				Protocol:                &protocol,
				FrontendPort:            to.Int32Ptr(port),
			},
		}
	}

	cases := []struct {
		name  string
		rules []*armnetwork.LoadBalancingRule
		valid bool
	}{
		{
			name:  "tcp and udp may share a port",
			rules: []*armnetwork.LoadBalancingRule{rule("a", tcp, 8000), rule("b", udp, 8000)},
			valid: true,
		},
		{
			name:  "different ports are fine",
			rules: []*armnetwork.LoadBalancingRule{rule("a", tcp, 8000), rule("b", tcp, 8001)},
			valid: true,
		},
		{
			name:  "same frontend protocol and port collide",
			rules: []*armnetwork.LoadBalancingRule{rule("a", tcp, 8000), rule("b", tcp, 8000)},
			valid: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lb := &armnetwork.LoadBalancer{
				Properties: &armnetwork.LoadBalancerPropertiesFormat{LoadBalancingRules: tc.rules},
			}
			err := validateLBRuleEndpointsUnique(lb)
			if tc.valid && err != nil {
				t.Errorf("expect valid, got %v", err)
			}
			if !tc.valid && err == nil {
				t.Error("expect the duplicate endpoint to be rejected")
			}
		})
	}
}

// TestValidateLBChildNames rejects duplicate and over long load balancer child names before the PUT.
// NOCC:tosa/fn_length(测试函数)
func TestValidateLBChildNames(t *testing.T) {
	lbWith := func(probes []*armnetwork.Probe) *armnetwork.LoadBalancer {
		return &armnetwork.LoadBalancer{
			Properties: &armnetwork.LoadBalancerPropertiesFormat{Probes: probes},
		}
	}

	dup := lbWith([]*armnetwork.Probe{
		{Name: to.StringPtr("lb-8000.8000")},
		{Name: to.StringPtr("lb-8000.8000")},
	})
	if err := validateLBChildNamesUnique(dup); err == nil {
		t.Error("expect duplicate probe names to be rejected")
	}

	tooLong := getLBRuleTgName(strings.Repeat("b", 75)+"-8000", 8000)
	if len(tooLong) <= MaxAzureResourceNameLen {
		t.Fatalf("test setup failed: name is only %d characters", len(tooLong))
	}
	long := lbWith([]*armnetwork.Probe{{Name: to.StringPtr(tooLong)}})
	err := validateLBChildNamesUnique(long)
	if err == nil {
		t.Fatal("expect over long probe name to be rejected")
	}
	if !strings.Contains(err.Error(), "azure allows at most") {
		t.Errorf("unexpected error message: %v", err)
	}

	ok := lbWith([]*armnetwork.Probe{{Name: to.StringPtr(getLBRuleTgName("lb-8000", 8000))}})
	if err = validateLBChildNamesUnique(ok); err != nil {
		t.Errorf("normal names should pass: %v", err)
	}
}

// TestValidateAgChildNamesUnique detects duplicate names within one child collection.
// NOCC:tosa/fn_length(测试函数)
func TestValidateAgChildNamesUnique(t *testing.T) {
	appGateway := newEmptyAppGateway()
	dupName := "443.deadbeef"
	appGateway.Properties.HTTPListeners = []*armnetwork.ApplicationGatewayHTTPListener{
		{Name: to.StringPtr(dupName)},
		{Name: to.StringPtr(dupName)},
	}
	if err := validateAgChildNamesUnique(appGateway); err == nil {
		t.Fatal("expect duplicate httpListeners error, got nil")
	}

	appGateway.Properties.HTTPListeners = []*armnetwork.ApplicationGatewayHTTPListener{
		{Name: to.StringPtr(dupName)},
	}
	// same name across different resource types is allowed by AGW
	appGateway.Properties.RequestRoutingRules = []*armnetwork.ApplicationGatewayRequestRoutingRule{
		{Name: to.StringPtr(dupName)},
	}
	if err := validateAgChildNamesUnique(appGateway); err != nil {
		t.Fatalf("cross-type same name should be allowed: %v", err)
	}
}
