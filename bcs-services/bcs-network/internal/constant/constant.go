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
	// ROUTE_TABLE_START_INDEX start index for linux route table
	ROUTE_TABLE_START_INDEX = 100
	// ENI_PREFIX prefix for elastic network interface
	ENI_PREFIX = "eni"

	// NODE_LABEL_KEY_FOR_NODE_NETWORK key in node annotation for cloud node network
	NODE_LABEL_KEY_FOR_NODE_NETWORK = "nodenetwork.bkbcs.tencent.com"

	// POD_ANNOTATION_KEY_FOR_ENI key in pod annotation for elastic network interface
	POD_ANNOTATION_KEY_FOR_ENI = "eni.cloud.bkbcs.tencent.com"
	// POD_ANNOTATION_KEY_FOR_ENI_REQUEST_IP key in pod annotation for request ip in eni network mode
	POD_ANNOTATION_KEY_FOR_ENI_REQUEST_IP = "requestip.cloud.bkbcs.tencent.com"
	// POD_ANNOTATION_VALUE_FOR_FIXED_IP value pod pod annotation for fixed ip
	POD_ANNOTATION_VALUE_FOR_FIXED_IP = "fixed"

	// IP_LABEL_KEY_FOR_HOST key in ip annotations for host, used to search cloud ip
	IP_LABEL_KEY_FOR_HOST = "host.cloud.bkbcs.tencent.com"
	// IP_LABEL_KEY_FOR_WORKLOAD_KIND key in ip annotations for workload kind
	IP_LABEL_KEY_FOR_WORKLOAD_KIND = "workload.cloud.bkbcs.tencent.com/kind"
	// IP_LABEL_KEY_FOR_IS_FIXED key in ip annotations for if it is fixed
	IP_LABEL_KEY_FOR_IS_FIXED = "fixed.cloud.bkbcs.tencent.com"
	// IP_LABEL_KEY_FOR_STATUS key in ip annotations for status
	IP_LABEL_KEY_FOR_STATUS = "status.cloud.bkbcs.tencent.com"
	// IP_LABEL_KEY_FOR_IS_CLUSTER_LAYER key in ip annotations for if ip is cluster layer
	IP_LABEL_KEY_FOR_IS_CLUSTER_LAYER = "clusterlayer.cloud.bkbcs.tencent.com"

	// INDEX_FOR_FLOATING_IP_ENI index for floating ip eni
	INDEX_FOR_FLOATING_IP_ENI = 99
	// FINALIZER_NAME_FOR_NETCONTROLLER finalizer name for net controller
	FINALIZER_NAME_FOR_NETCONTROLLER = "netcontroller.cloud.bkbcs.tencent.com"
	// FINALIZER_NAME_FOR_NETAGENT finalizer name for net agent
	FINALIZER_NAME_FOR_NETAGENT = "netagent.cloud.bkbcs.tencent.com"

	// CLOUD_CRD_VERSION_V1 version for cloud crd
	CLOUD_CRD_VERSION_V1 = "v1"
	// CLOUD_CRD_NAMESPACE_BCS_SYSTEM namespace for cloud crd
	CLOUD_CRD_NAMESPACE_BCS_SYSTEM = "bcs-system"
	// CLOUD_CRD_NAME_CLOUD_SUBNET crd name for cloud subnet
	CLOUD_CRD_NAME_CLOUD_SUBNET = "CloudSubnet"
	// CLOUD_CRD_NAME_CLOUD_IP crd nama for cloud ip
	CLOUD_CRD_NAME_CLOUD_IP = "CloudIP"

	// CLOUD_KIND_TENCENT cloud provider name of tencent cloud
	CLOUD_KIND_TENCENT = "tencentcloud"
	// CLOUD_KIND_AWS cloud provider name of tencent aws
	CLOUD_KIND_AWS = "aws"

	// IP_STATUS_ACTIVE ip is active
	IP_STATUS_ACTIVE = "active"
	// IP_STATUS_AVAILABLE ip is available
	IP_STATUS_AVAILABLE = "available"
	// IP_STATUS_DELETING ip is deleting
	IP_STATUS_DELETING = "deleting"
)
