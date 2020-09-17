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

package v2

const (
	// CniAnnotationKey bkbcs CNI plugin annotation key
	CniAnnotationKey = "tke.cloud.tencent.com/networks"
	// FixedIpAnnotationKey bkbcs fixed ip request annotation key
	FixedIpAnnotationKey = "eni.cloud.bkbcs.tencent.com"
	// CniAnnotationValue CNI plugin annotation value
	CniAnnotationValue = "bcs-eni-cni"
	// FixedIpAnnotationValue fixed ip request annotation value
	FixedIpAnnotationValue = "fixed"
	// BcsSystem system namespace for scheduler
	BcsSystem = "bcs-system"
	// IP_LABEL_KEY_FOR_HOST key in ip annotations for host, used to search cloud ip
	IP_LABEL_KEY_FOR_HOST = "host.cloud.bkbcs.tencent.com"
	// IP_LABEL_KEY_FOR_IS_FIXED key in ip annotations for if it is fixed
	IP_LABEL_KEY_FOR_IS_FIXED = "fixed.cloud.bkbcs.tencent.com"
)
