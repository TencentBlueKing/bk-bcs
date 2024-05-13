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

package tencentcloud

const (
	// EnvNameTencentCloudVpcDomain env name of tencent cloud domain
	EnvNameTencentCloudVpcDomain = "TENCENTCLOUD_VPC_DOMAIN"
	// EnvNameTencentCloudCvmDomain env name of tencent cloud cvm domain
	EnvNameTencentCloudCvmDomain = "TENCENTCLOUD_CVM_DOMAIN"
	// EnvNameTencentCloudRegion env name of tencent cloud region
	EnvNameTencentCloudRegion = "TENCENTCLOUD_REGION"
	// EnvNameTencentCloudSecurityGroups env name of tencent cloud security groups
	EnvNameTencentCloudSecurityGroups = "TENCENTCLOUD_SECURITY_GROUPS"
	// EnvNameTencentCloudAccessKeyID env name of tencent cloud secret id
	EnvNameTencentCloudAccessKeyID = "TENCENTCLOUD_ACCESS_KEY_ID"
	// EnvNameTencentCloudAccessKey env name of tencent cloud secret key
	EnvNameTencentCloudAccessKey = "TENCENTCLOUD_ACCESS_KEY"

	// EnvNameTencentCloudEniPendingStatus eni pending status
	EnvNameTencentCloudEniPendingStatus = "PENDING"
	// EnvNameTencentCloudEniAvailableStatus eni available status
	EnvNameTencentCloudEniAvailableStatus = "AVAILABLE"
	// EnvNameTencentCloudEniAttachingStatus eni attaching status
	EnvNameTencentCloudEniAttachingStatus = "ATTACHING"
	// EnvNameTencentCloudEniDetachingStatus eni pending status
	EnvNameTencentCloudEniDetachingStatus = "DETACHING"
	// EnvNameTencentCloudEniDeletingStatus eni pending status
	EnvNameTencentCloudEniDeletingStatus = "DELETING"
	// EnvNameTencentCloudEniDetachedStatus eni pending status
	EnvNameTencentCloudEniDetachedStatus = "DETACHED"
	// EnvNameTencentCloudEniAttachedStatus eni pending status
	EnvNameTencentCloudEniAttachedStatus = "ATTACHED"

	// DefaultCheckNum default check number for asynchronous cloud task
	DefaultCheckNum = 5
	// DefaultCheckInterval default check interval for asynchrounous cloud task
	DefaultCheckInterval = 3
)
