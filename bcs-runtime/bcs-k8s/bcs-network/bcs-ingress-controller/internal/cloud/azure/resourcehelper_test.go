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
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/go-autorest/autorest/to"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

const (
	testSubscriptionID = "00000000-0000-0000-0000-000000000000"
	testResourceGroup  = "test-rg"
	testAppGatewayName = "test-appgw"
)

func testIDPrefix() string {
	return "/subscriptions/" + testSubscriptionID + "/resourceGroups/" + testResourceGroup +
		"/providers/Microsoft.Network/applicationGateways/" + testAppGatewayName
}

// TestGenSubResource ensure generated sub resource ID uses the resource type segment required by
// Azure ARM. The segment must match the corresponding property name on the parent application
// gateway, otherwise ARM rejects the whole request with 400 InvalidJsonReferenceFormat.
func TestGenSubResource(t *testing.T) {
	rh := NewResourceHelper(testSubscriptionID, testResourceGroup)

	testCases := []struct {
		name    string
		resType resourceType
		resName string
		want    string
	}{
		{
			name:    "ssl certificate",
			resType: ResourceTypeSSLCertificate,
			resName: "my-cert",
			want:    testIDPrefix() + "/sslCertificates/my-cert",
		},
		{
			name:    "ssl profile",
			resType: ResourceTypeSSLProfile,
			resName: "my-profile",
			want:    testIDPrefix() + "/sslProfiles/my-profile",
		},
		{
			name:    "frontend port",
			resType: ResourceTypeFrontendPorts,
			resName: "port_443",
			want:    testIDPrefix() + "/frontendPorts/port_443",
		},
		{
			name:    "http listener",
			resType: ResourceTypeHttpListeners,
			resName: "443",
			want:    testIDPrefix() + "/httpListeners/443",
		},
		{
			name:    "backend address pool",
			resType: ResourceTypeBackendAddressPools,
			resName: "pool-1",
			want:    testIDPrefix() + "/backendAddressPools/pool-1",
		},
		{
			name:    "url path map",
			resType: ResourceTypeURLPathMaps,
			resName: "map-1",
			want:    testIDPrefix() + "/urlPathMaps/map-1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := rh.genSubResource(ResourceProviderApplicationGateway, testAppGatewayName,
				tc.resType, tc.resName)
			if got == nil || got.ID == nil {
				t.Fatalf("genSubResource returned nil ID")
			}
			if *got.ID != tc.want {
				t.Errorf("unexpected resource ID\n got: %s\nwant: %s", *got.ID, tc.want)
			}
		})
	}
}

// TestAgHTTPSListenerCertID ensure an HTTPS listener references its certificate through a valid
// Azure sub resource ID.
// NOCC:tosa/fn_length(测试函数)
func TestAgHTTPSListenerCertID(t *testing.T) {
	alb := &Alb{
		resourceHelper: NewResourceHelper(testSubscriptionID, testResourceGroup),
	}

	appGateway := &armnetwork.ApplicationGateway{
		Properties: &armnetwork.ApplicationGatewayPropertiesFormat{
			FrontendIPConfigurations: []*armnetwork.ApplicationGatewayFrontendIPConfiguration{
				{ID: to.StringPtr(testIDPrefix() + "/frontendIPConfigurations/frontend-ip")},
			},
		},
	}

	listener := &networkextensionv1.Listener{
		Spec: networkextensionv1.ListenerSpec{
			LoadbalancerID: testAppGatewayName,
			Port:           443,
			Protocol:       AzureProtocolHTTPS,
			Certificate: &networkextensionv1.IngressListenerCertificate{
				CertID: "my-cert",
			},
			Rules: []networkextensionv1.ListenerRule{
				{Domain: "example.com"},
			},
		},
	}

	result, err := alb.ensureHttpListenerForAg(appGateway, []*networkextensionv1.Listener{listener})
	if err != nil {
		t.Fatalf("ensureHttpListenerForAg failed: %v", err)
	}
	if len(result.Properties.HTTPListeners) != 1 {
		t.Fatalf("expect 1 http listener, got %d", len(result.Properties.HTTPListeners))
	}

	sslCert := result.Properties.HTTPListeners[0].Properties.SSLCertificate
	if sslCert == nil || sslCert.ID == nil {
		t.Fatalf("expect ssl certificate reference on HTTPS listener, got nil")
	}

	want := testIDPrefix() + "/sslCertificates/my-cert"
	if *sslCert.ID != want {
		t.Errorf("unexpected ssl certificate ID\n got: %s\nwant: %s", *sslCert.ID, want)
	}
}
