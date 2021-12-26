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

package perm_resource

// Provider is the interface for provider
type Provider interface {
	ListAttr() ([]AttrResource, error)
	ListAttrValue(filter *ListAttrValueFilter, page Page) (*ListAttrValueResult, error)

	// ListInstance and SearchInstance
	ListInstance(filter *ListInstanceFilter, page Page) (*ListInstanceResult, error)
	SearchInstance(filter *SearchInstanceFilter, page Page) (*ListInstanceResult, error)
	FetchInstanceInfo(filter *FetchInstanceInfoFilter) ([]map[string]interface{}, error)

	ListInstanceByPolicy(filter *ListInstanceByPolicyFilter, page Page) (*ListInstanceResult, error)
}

// Dispatcher is the interface of dispatcher, for callback
type Dispatcher interface {
	RegisterProvider(resourceType string, provider Provider)
	GetProvider(resourceType string) (provider Provider, exist bool)
}

// NewDispatcher will create a dispatcher
func NewDispatcher() Dispatcher {
	return &dispatcher{
		providers: make(map[string]Provider),
	}
}

type dispatcher struct {
	providers map[string]Provider
}

// RegisterProvider will register a provider
func (d *dispatcher) RegisterProvider(resourceType string, provider Provider) {
	d.providers[resourceType] = provider
}

// GetProvider get the provider by type
func (d *dispatcher) GetProvider(resourceType string) (Provider, bool) {
	provider, exist := d.providers[resourceType]
	return provider, exist
}
