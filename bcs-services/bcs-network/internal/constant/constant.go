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
	// RouteTableStartIndex start index for linux route table
	RouteTableStartIndex = 100
	// EniPrefix prefix for elastic network interface
	EniPrefix = "eni"

	// NodeAnnotationKeyForNodeNetwork key in node annotation for cloud node network
	NodeAnnotationKeyForNodeNetwork = "nodenetwork.bkbcs.tencent.com"

	// PodAnnotationKeyForEni key in pod annotation for elastic network interface
	PodAnnotationKeyForEni = "eni.cloud.bkbcs.tencent.com"
	// PodAnnotationKeyForEniRequestIP key in pod annotation for request ip in eni network mode
	PodAnnotationKeyForEniRequestIP = "requestip.cloud.bkbcs.tencent.com"
	// PodAnnotationValueForFixedIP value pod pod annotation for fixed ip
	PodAnnotationValueForFixedIP = "fixed"
	// PodAnnotationKeyForHost key in pod annotations for host, used to search cloud ip
	PodAnnotationKeyForHost = "host.cloud.bkbcs.tencent.com"

	// IPAnnotationKeyForHost key in ip annotations for host, used to search cloud ip
	IPAnnotationKeyForHost = "host.cloud.bkbcs.tencent.com"
	// IPAnnotationKeyForWorkloadName key in ip annotations for workload name
	IPAnnotationKeyForWorkloadName = "workload.cloud.bkbcs.tencent.com/name"
	// IPAnnotationKeyForWorkloadKind key in ip annotations for workload kind
	IPAnnotationKeyForWorkloadKind = "workload.cloud.bkbcs.tencent.com/kind"
	// IPAnnotationKeyForContainerID key in ip annotations for container id
	IPAnnotationKeyForContainerID = "container.cloud.bkbcs.tencent.com"
	// IPAnnotationKeyForIsFixed key in ip annotations for if it is fixed
	IPAnnotationKeyForIsFixed = "fixed.cloud.bkbcs.tencent.com"
	// IPAnnotationKeyForStatus key in ip annotations for status
	IPAnnotationKeyForStatus = "status.cloud.bkbcs.tencent.com"
	// IPAnnotationKeyForIsClusterLayer key in ip annotations for if ip is cluster layer
	IPAnnotationKeyForIsClusterLayer = "clusterlayer.cloud.bkbcs.tencent.com"

	// IndexForFloatingIPEni index for floating ip eni
	IndexForFloatingIPEni = 99
	// FinalizerNameForNetController finalizer name for net controller
	FinalizerNameForNetController = "netcontroller.cloud.bkbcs.tencent.com"
	// FinalizerNameForNetAgent finalizer name for net agent
	FinalizerNameForNetAgent = "netagent.cloud.bkbcs.tencent.com"

	// CloudCrdVersionV1 version for cloud crd
	CloudCrdVersionV1 = "v1"
	// CloudCrdNamespace namespace for cloud crd
	CloudCrdNamespace = "bcs-system"
	// CloudCrdNameSubnet crd name for cloud subnet
	CloudCrdNameSubnet = "CloudSubnet"
	// CloudCrdNameIP crd nama for cloud ip
	CloudCrdNameIP = "CloudIP"

	// CloudTencent cloud provider name of tencent cloud
	CloudTencent = "tencentcloud"
	// CloudAws cloud provider name of tencent aws
	CloudAws = "aws"

	// StatusIPActive ip is active
	StatusIPActive = "active"
	// StatusIPAvailable ip is available
	StatusIPAvailable = "available"
	// StatusIPDeleting ip is deleting
	StatusIPDeleting = "deleting"
)
