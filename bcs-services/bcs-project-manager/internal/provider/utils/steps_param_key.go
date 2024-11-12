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

// Package utils xxx
package utils

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
)

var (
	// UserNameKey user
	UserNameKey task.ParamKey = "user"
	// ProjectCodeKey projectCode
	ProjectCodeKey task.ParamKey = "projectCode"
	// ProjectIdKey projectID
	ProjectIdKey task.ParamKey = "projectId"
	// ClusterIDKey clusterID
	ClusterIDKey task.ParamKey = "clusterId"
	// ContentKey content
	ContentKey task.ParamKey = "content"
	// WaitTimeKey waitTime
	WaitTimeKey task.ParamKey = "waitTime"
	// WaitTypeKey waitType
	WaitTypeKey task.ParamKey = "waitType"
	// EndWaitTimeKey end wait time
	EndWaitTimeKey task.ParamKey = "endWaitTime"

	// QuotaIdKey quotaId
	QuotaIdKey task.ParamKey = "quotaId"
	// FederationQuotaDataKey federationQuota
	FederationQuotaDataKey task.ParamKey = "federationQuota"
)
