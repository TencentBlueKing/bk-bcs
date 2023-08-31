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
	// BCSNetPoolInitializingStatus for BCSNetPool Initializing status
	BCSNetPoolInitializingStatus = "Initializing"
	// BCSNetPoolNormalStatus for BCSNetPool Normal status
	BCSNetPoolNormalStatus = "Normal"

	// BCSNetIPActiveStatus for BCSNetIP Active status
	BCSNetIPActiveStatus = "Active"
	// BCSNetIPAvailableStatus for BCSNetIP Available status
	BCSNetIPAvailableStatus = "Available"
	// BCSNetIPReservedStatus for BCSNetIP Reserved status
	BCSNetIPReservedStatus = "Reserved"

	// BCSNetIPClaimBoundedStatus for BCSNetIPClaim Bound status
	BCSNetIPClaimBoundedStatus = "Bound"
	// BCSNetIPClaimPendingStatus for BCSNetIPClaim Pending status
	BCSNetIPClaimPendingStatus = "Pending"
	// BCSNetIPClaimExpiredStatus for BCSNetIPClaim Expired status
	BCSNetIPClaimExpiredStatus = "Expired"

	// PodAnnotationKeyForIPClaim pod annotation key for IP claim
	PodAnnotationKeyForIPClaim = "netservicecontroller.bkbcs.tencent.com/ipclaim"

	// FixIPLabel label key for fix ip
	FixIPLabel = "fixed-ip"

	// FinalizerNameBcsNetserviceController finalizer name of bcs netservice controller
	FinalizerNameBcsNetserviceController = "netservicecontroller.bkbcs.tencent.com"
)
