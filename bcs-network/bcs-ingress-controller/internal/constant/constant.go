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

package constant

const (
	// FinalizerNameBcsIngressController finalizer name of bcs ingress controller
	FinalizerNameBcsIngressController = "ingresscontroller.bkbcs.tencent.com"
	// CloudTencent tencent cloud
	CloudTencent = "tencentcloud"
	// CloudAWS aws cloud
	CloudAWS = "aws"

	// EnvNameIsTCPUDPPortReuse env name for option if the loadbalancer provider support tcp udp port reuse
	// if enabled, we will find protocol info in 4 layer listener name
	EnvNameIsTCPUDPPortReuse = "TCP_UDP_PORT_REUSE"
	// EnvNameIsBulkMode env name for option if use bulk interface for cloud lb
	EnvNameIsBulkMode = "IS_BULK_MODE"
)
