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

package pkgs

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/cmd/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/pkg/consumer"
)

// ResourceSwitch xxx
type ResourceSwitch string

const (
	// ResourceSubOn on sub handler
	ResourceSubOn ResourceSwitch = "on"
	// ResourceSubOff off sub handler
	ResourceSubOff ResourceSwitch = "off"
)

// GetFactoryConsumers init consumer object according to resourceSubInfo
func GetFactoryConsumers(options *config.AlertManagerOptions) []consumer.Consumer {
	resourceSubInfo := parseResourceSubs(options)
	var consumers []consumer.Consumer

	for resource, switchKey := range resourceSubInfo {
		if strings.EqualFold(switchKey, string(ResourceSubOn)) {
			con := handlerFactory(resource, options)
			if con != nil {
				consumers = append(consumers, con)
			}
		}
	}

	return consumers
}

func handlerFactory(resourceKind string, options *config.AlertManagerOptions) consumer.Consumer {
	switch resourceKind {
	case msgqueue.EventSubscribeType:
		return GetEventSyncHandler(options)
	}
	return nil
}

func parseResourceSubs(options *config.AlertManagerOptions) map[string]string {
	resourceSubs := make(map[string]string)

	for _, resource := range options.ResourceSubs {
		resourceKind := resource.Category
		switchKey := resource.Switch

		if _, ok := resourceSubs[resourceKind]; !ok {
			resourceSubs[resourceKind] = switchKey
		}
	}

	blog.Infof("parseResourceSubs %v", resourceSubs)
	return resourceSubs
}
