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
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

// ResourceHelper help generate azure sub resource
type ResourceHelper struct {
	subscriptionsID   string
	resourceGroupName string

	idPrefix string
}

type azureResource struct {
	Id                        string
	Name                      string
	SubscriptionsID           string
	ResourceGroupName         string
	ResourceProviderNamespace string
	ResourceType              string
}

type resourceProvider string
type resourceType string

const (
	// ResourceProviderLoadBalancer resource provider of load balancer
	ResourceProviderLoadBalancer resourceProvider = "loadBalancers"
	// ResourceProviderApplicationGateway resource provider of application gateway
	ResourceProviderApplicationGateway resourceProvider = "applicationGateways"

	// ResourceTypeProbes resource type of probes
	ResourceTypeProbes resourceType = "probes"
	// ResourceTypeBackendAddressPools resource type of BackendAddressPools
	ResourceTypeBackendAddressPools resourceType = "backendAddressPools"
	// ResourceTypeFrontendIpConfiguration resource type of FrontEndIpConfiguration
	ResourceTypeFrontendIpConfiguration resourceType = "frontendIPConfigurations"

	// ResourceTypeFrontendPorts resource type of frontendPorts
	ResourceTypeFrontendPorts resourceType = "frontendPorts"
	// ResourceTypeHttpListeners resource type of httpListeners
	ResourceTypeHttpListeners resourceType = "httpListeners"
	// ResourceTypeRequestRoutingRules resource type of requestRoutingRules
	ResourceTypeRequestRoutingRules resourceType = "requestRoutingRules"
	// ResourceTypeBackendHttpSettingsCollection resource type of backendHttpSettingsCollection
	ResourceTypeBackendHttpSettingsCollection resourceType = "backendHttpSettingsCollection"
	// ResourceTypeURLPathMaps resource type of urlPathMaps
	ResourceTypeURLPathMaps resourceType = "urlPathMaps"
	// ResourceTypeSSLCertificate resource type of sslCertificate
	ResourceTypeSSLCertificate resourceType = "sslCertificate"
	// ResourceTypeSSLProfile resource type of sslProfile
	ResourceTypeSSLProfile resourceType = "sslProfile"
)

// NewResourceHelper return resource helper
func NewResourceHelper(subscriptionsID string, resourceGroupName string) *ResourceHelper {
	return &ResourceHelper{
		subscriptionsID:   subscriptionsID,
		resourceGroupName: resourceGroupName,

		idPrefix: fmt.Sprintf("subscriptions/%s/resourceGroups/%s/providers/Microsoft."+
			"Network", subscriptionsID, resourceGroupName),
	}
}

// resourceID format: /subscriptions/{subID}/resourceGroups/{resource-group-name}/{resource-provider-namespace
// }/{resource-type}/{resource-name}
func (rh *ResourceHelper) transResourceID(resourceId string) (*azureResource, error) {
	splitStrs := strings.Split(resourceId, "/")
	if len(splitStrs) != 7 {
		return nil, fmt.Errorf("invalid azure resource ID: %s", resourceId)
	}

	return &azureResource{
		Id:                        resourceId,
		Name:                      splitStrs[6],
		SubscriptionsID:           splitStrs[1],
		ResourceGroupName:         splitStrs[3],
		ResourceProviderNamespace: splitStrs[4],
		ResourceType:              splitStrs[5],
	}, nil
}

func (rh *ResourceHelper) genSubResource(resProviderType resourceProvider, resProviderName string,
	resType resourceType,
	resName string) *armnetwork.SubResource {
	id := fmt.Sprintf("/%s/%s/%s/%s/%s", rh.idPrefix, resProviderType, resProviderName, resType, resName)
	return &armnetwork.SubResource{ID: &id}
}

func (rh *ResourceHelper) getSubResourceByID(id string) *armnetwork.SubResource {
	return &armnetwork.SubResource{ID: &id}
}
