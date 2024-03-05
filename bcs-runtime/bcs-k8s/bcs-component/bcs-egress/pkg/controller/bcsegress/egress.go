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

package bcsegress

import (
	"fmt"
	"sort"
)

// SimpleHTTPConfig init one simple config
func SimpleHTTPConfig(name, domain string, port uint) *HTTPConfig {
	c := &HTTPConfig{
		Name:            name,
		Domain:          domain,
		DestinationPort: port,
		Label:           make(map[string]string),
	}
	c.Key()
	return c
}

// HTTPConfig configuration for http proxy
type HTTPConfig struct {
	indexer string
	Name    string
	Domain  string
	// backend ip list, reserved for extension
	IPs             []string
	DestinationPort uint
	// Label use for custom information storage
	// all control informations are depend on Label,
	// ! Label is reqired
	Label map[string]string
}

// Key indexer for cache storage
func (config *HTTPConfig) Key() string {
	if len(config.indexer) == 0 {
		if config.DestinationPort != 80 {
			config.indexer = fmt.Sprintf("%s:%d", config.Domain, config.DestinationPort)
		} else {
			config.indexer = config.Domain
		}
	}
	return config.indexer
}

// IsChanged check if destination Config changed
func (config *HTTPConfig) IsChanged(dest *HTTPConfig) bool {
	if config.Name != dest.Name {
		return true
	}
	if config.Domain != dest.Domain {
		return true
	}
	if config.DestinationPort != dest.DestinationPort {
		return true
	}
	return false
}

// LabelFilter find specified Config, if filter match exactly
// then return true, otherwise false
func (config *HTTPConfig) LabelFilter(filter map[string]string) bool {
	if len(filter) == 0 {
		return false
	}
	if len(config.Label) == 0 {
		return false
	}
	for k, v := range filter {
		if value, ok := config.Label[k]; ok && v == value {
			continue
		} else {
			return false
		}
	}
	return true
}

// TCPConfig configuration for tcp proxy
// indexer: name_port
type TCPConfig struct {
	indexer string
	// Name for management & first indexer
	Name string
	// port for second indexer
	ProxyPort uint
	// check if we need to create backend
	HasBackend bool
	// domain for destination
	Domain string
	// specified destination
	IPs []string
	// algorithm only use for IP mode
	Algorithm string
	// DestinationPort work for domain & iplist
	DestinationPort uint
	// Label use for custom information storage
	// all control informations are depend on Label,
	// ! Label is reqired
	Label map[string]string
}

// IsChanged check if destination Config changed
func (config *TCPConfig) IsChanged(dest *TCPConfig) bool {
	if config.Domain != dest.Domain {
		return true
	}
	if config.DestinationPort != dest.DestinationPort {
		return true
	}
	if config.Algorithm != dest.Algorithm {
		return true
	}
	if len(config.IPs) != len(dest.IPs) {
		return true
	}
	for index, ip := range config.IPs {
		if ip != dest.IPs[index] {
			return true
		}
	}
	return false
}

// Key config key for cache
func (config *TCPConfig) Key() string {
	if len(config.indexer) == 0 {
		config.indexer = fmt.Sprintf("%s_%d", config.Name, config.ProxyPort)
	}
	return config.indexer
}

// LabelFilter find specified Config, if filter match exactly
// then return true, otherwise false
func (config *TCPConfig) LabelFilter(filter map[string]string) bool {
	if len(filter) == 0 {
		return false
	}
	if len(config.Label) == 0 {
		return false
	}
	for k, v := range filter {
		if value, ok := config.Label[k]; ok && v == value {
			continue
		} else {
			return false
		}
	}
	return true
}

// SortIPs sort IP list when for destination
func (config *TCPConfig) SortIPs() {
	sort.Strings(config.IPs)
}

// TCPList list for sort tcpConfig
type TCPList []*TCPConfig

// Len return list length
func (l TCPList) Len() int {
	return len(l)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (l TCPList) Less(i, j int) bool {
	if l[i].ProxyPort < l[j].ProxyPort {
		return true
	}
	return l[i].Name < l[j].Name
}

// Swap swaps the elements with indexes i and j.
func (l TCPList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// HTTPList list for sort http config
type HTTPList []*HTTPConfig

// Len return list length
func (l HTTPList) Len() int {
	return len(l)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (l HTTPList) Less(i, j int) bool {
	if l[i].DestinationPort < l[j].DestinationPort {
		return true
	}
	return l[i].Name < l[j].Name
}

// Swap swaps the elements with indexes i and j.
func (l HTTPList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
