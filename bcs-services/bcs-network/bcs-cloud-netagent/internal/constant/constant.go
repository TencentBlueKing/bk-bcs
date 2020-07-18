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
	START_ROUTE_TABLE = 100
	ENI_PREFIX        = "eni"

	// annotations for pod
	ANNOTATION_POD_ENI_KEY         = "eni.cloud.bkbcs.tencent.com"
	ANNOTATION_POD_REQUESTIP_KEY   = "requestip.cloud.bkbcs.tencent.com"
	ANNOTATION_POD_ENI_VALUE_FIXED = "fixed"
	ANNOTATION_POD_ENI_VALUE_TRUE  = "true"

	// constant for cloud IP
	CRD_VERSION_V1        = "v1"
	CRD_NAME_CLOUD_SUBNET = "CloudSubnet"
	CRD_NAME_CLOUD_IP     = "CloudIP"
	ANNOTATION_CLOUD_IP_HOST = "host.cloud.bkbcs.tencent.com"
)
