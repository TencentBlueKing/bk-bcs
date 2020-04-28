/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package strategy

import (
	"encoding/json"

	pbcommon "bk-bscp/internal/protocol/common"
)

const (
	// EmptySidecarLabels is empty json string for sidecar labels.
	EmptySidecarLabels = "{}"

	// EmptyStrategy is empty json string for release.
	EmptyStrategy = "{}"
)

// SidecarLabels is app sidecar instance labels struct.
type SidecarLabels struct {
	Labels map[string]string
}

/*
	example:
	{
		"Labels":
			{
				"k1":"v1",
				"k2":"v2",
				"k3":"v3"
			}
	}
*/

// Strategy is struct of strategies in release.
type Strategy struct {
	Appid      string
	Clusterids []string
	Zoneids    []string
	Dcs        []string
	IPs        []string
	Labels     map[string]string
}

/*
	example:
	{
		"Appid":"appid",
		"Clusterids":
			["clusterid01","clusterid02","clusterid03"],
		"Zoneids":
			["zoneid01","zoneid02","zoneid03"],
		"Dcs":
			["dc01","dc02","dc03"],
		"IPs":
			["X.X.X.1","X.X.X.2","X.X.X.3"],
		"Labels":
			{
				"k1":"v1",
				"k2":"v2",
				"k3":"v3"
			}
	}
*/

// Matcher matches the strategy base on information in app sidecar instance.
type Matcher func(strategy *Strategy, instance *pbcommon.AppInstance) bool

// Handler handles the strategy matching.
type Handler struct {
	matcher Matcher
}

// NewHandler creates a new strategy handler.
func NewHandler(matcher Matcher) *Handler {
	if matcher == nil {
		matcher = defaultMatcher
	}
	return &Handler{matcher: matcher}
}

// Matcher returns matcher in handler.
func (h *Handler) Matcher() Matcher {
	return h.matcher
}

// default matcher impl.
var defaultMatcher = func(strategy *Strategy, instance *pbcommon.AppInstance) bool {
	return matchWhitelist(strategy, instance)
}

// matchWhitelist matchs strategy.
func matchWhitelist(strategy *Strategy, instance *pbcommon.AppInstance) bool {
	// #lizard forgives
	if strategy.Appid != instance.Appid {
		return false
	}

	if strategy.Clusterids != nil && len(strategy.Clusterids) != 0 {
		for _, clusterid := range strategy.Clusterids {
			if clusterid == instance.Clusterid {
				return true
			}
		}
	}

	if strategy.Zoneids != nil && len(strategy.Zoneids) != 0 {
		for _, zoneid := range strategy.Zoneids {
			if zoneid == instance.Zoneid {
				return true
			}
		}
	}

	if strategy.Dcs != nil && len(strategy.Dcs) != 0 {
		for _, dc := range strategy.Dcs {
			if dc == instance.Dc {
				return true
			}
		}
	}

	if strategy.IPs != nil && len(strategy.IPs) != 0 {
		for _, ip := range strategy.IPs {
			if ip == instance.IP {
				return true
			}
		}
	}

	if strategy.Labels != nil && len(strategy.Labels) != 0 {
		labels := SidecarLabels{}
		if err := json.Unmarshal([]byte(instance.Labels), &labels); err != nil {
			return false
		}

		for k, v := range strategy.Labels {
			if labels.Labels[k] == v {
				return true
			}
		}
	}

	return false
}
