/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package option

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

// ControllerOption options for controller
type ControllerOption struct {
	// Address address for server
	Address string

	// Port port for server
	Port int

	// MetricPort port for metric server
	MetricPort int

	// Cloud cloud mod
	Cloud string

	// Region cloud region
	Region string

	// ElectionNamespace election namespace
	ElectionNamespace string

	// IsNamespaceScope if the ingress can only be associated with the service and workload in the same namespace
	IsNamespaceScope bool

	// LogConfig for blog
	conf.LogConfig

	// IsTCPUDPReuse if the loadbalancer provider support tcp udp port reuse
	// if enabled, we will find protocol info in 4 layer listener name
	IsTCPUDPPortReuse bool

	// IsBulkMode if use bulk interface for cloud lb
	IsBulkMode bool
}
