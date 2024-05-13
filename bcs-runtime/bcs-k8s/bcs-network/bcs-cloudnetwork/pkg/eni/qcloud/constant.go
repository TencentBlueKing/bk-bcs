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

package qcloud

const (
	// ENV_NAME_TENCENTCLOUD_REGION env name of tencent cloud region
	ENV_NAME_TENCENTCLOUD_REGION = "TENCENTCLOUD_REGION"
	// ENV_NAME_TENCENTCLOUD_VPC env name of tencent cloud vpc
	ENV_NAME_TENCENTCLOUD_VPC = "TENCENTCLOUD_VPC"
	// ENV_NAME_TENCENTCLOUD_SUBNETS env name of tencent cloud vpc
	ENV_NAME_TENCENTCLOUD_SUBNETS = "TENCENTCLOUD_SUBNETS"
	// ENV_NAME_TENCENTCLOUD_SECURITY_GROUPS env name of tencent cloud security groups used by enis
	ENV_NAME_TENCENTCLOUD_SECURITY_GROUPS = "TENCENTCLOUD_SECURITY_GROUPS"
	// ENV_NAME_TENCENTCLOUD_ACCESS_KEY_ID env name of tencent cloud secret id
	ENV_NAME_TENCENTCLOUD_ACCESS_KEY_ID = "TENCENTCLOUD_ACCESS_KEY_ID"
	// ENV_NAME_TENCENTCLOUD_ACCESS_KEY env name of tencent cloud secret key
	ENV_NAME_TENCENTCLOUD_ACCESS_KEY = "TENCENTCLOUD_ACCESS_KEY"

	// ENI_STATUS_PENDING eni pending status
	ENI_STATUS_PENDING = "PENDING"
	// ENI_STATUS_AVAILABLE eni available status
	ENI_STATUS_AVAILABLE = "AVAILABLE"
	// ENI_STATUS_ATTACHING eni attaching status
	ENI_STATUS_ATTACHING = "ATTACHING"
	// ENI_STATUS_DETACHING eni detaching status
	ENI_STATUS_DETACHING = "DETACHING"
	// ENI_STATUS_DELETING eni deleting status
	ENI_STATUS_DELETING = "DELETING"

	// status determined by eni Attachment field

	// ENI_STATUS_DETACHED eni detached status
	ENI_STATUS_DETACHED = "DETACHED"
	// ENI_STATUS_ATTACHED eni attached status
	ENI_STATUS_ATTACHED = "ATTACHED"

	// DEFAULT_CHECK_NUM default check number
	DEFAULT_CHECK_NUM = 5
	// DEFAULT_CHECK_INTERVAL default check interval, unit second
	DEFAULT_CHECK_INTERVAL = 3
)
