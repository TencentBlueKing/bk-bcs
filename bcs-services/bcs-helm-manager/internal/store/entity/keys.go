/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package entity

// 定义一批统一的key, 用在db相关的字段
const (
	FieldKeyProjectID   = "projectID"
	FieldKeyProjectCode = "projectCode"
	FieldKeyClusterID   = "clusterID"
	FieldKeyName        = "name"
	FieldKeyDisplayName = "displayName"
	FieldKeyPublic      = "public"
	FieldKeyNamespace   = "namespace"
	FieldKeyType        = "type"
	FieldKeyRevision    = "revision"

	FieldKeyRemote         = "remote"
	FieldKeyRemoteURL      = "remoteURL"
	FieldKeyRemoteUsername = "remoteUsername"
	FieldKeyRemotePassword = "remotePassword"
	FieldKeyRepoURL        = "repoURL"
	FieldKeyUsername       = "username"
	FieldKeyPassword       = "password"

	FieldKeyRepoName     = "repo"
	FieldKeyChartName    = "chartName"
	FieldKeyChartVersion = "chartVersion"
	FieldKeyValueFile    = "valueFile"
	FieldKeyValues       = "values"
	FieldKeyArgs         = "args"

	FieldKeyCreateBy   = "createBy"
	FieldKeyUpdateBy   = "updateBy"
	FieldKeyCreateTime = "createTime"
	FieldKeyUpdateTime = "updateTime"

	FieldKeyStatus  = "status"
	FieldKeyMessage = "message"
)
