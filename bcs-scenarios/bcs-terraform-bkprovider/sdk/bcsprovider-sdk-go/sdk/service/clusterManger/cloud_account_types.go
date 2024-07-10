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

// Package clusterManger cluster-service
package clusterManger

/*
	云凭证导入
*/

const (
	// createCloudAccountApi post ( cloudID )
	createCloudAccountApi = "/clustermanager/v1/clouds/%s/accounts"

	// deleteCloudAccountApi delete ( cloudID + accountID )
	deleteCloudAccountApi = "/clustermanager/v1/clouds/%s/accounts/%s"

	// updateCloudAccountApi put ( cloudID + accountID )
	updateCloudAccountApi = "/clustermanager/v1/clouds/%s/accounts/%s"

	// listCloudAccountApi  get ( cloudID )
	listCloudAccountApi = "/clustermanager/v1/clouds/%s/accounts"
)

// 云ID
const (
	// GcpCloud 谷歌云
	GcpCloud = "gcpCloud"

	// AzureCloud azure云
	AzureCloud = "azureCloud"

	// TencentCloud 腾讯云
	TencentCloud = "tencentPublicCloud"
)

// CreateCloudAccountRequest body
//
// CloudID 云区域ID
// ***不能为空***
//
// AccountName 账号名称
// ***不能为空***
//
// ProjectID 项目ID
// ***不能为空***
type CreateCloudAccountRequest struct {
	// RequestID 请求ID
	RequestID string

	// CloudID 云区域ID
	// ***不能为空***
	CloudID string `json:"cloudID,omitempty"`

	// AccountName 账号名称
	// ***不能为空***
	AccountName string `json:"accountName,omitempty"`

	// Account 账号
	// ***不能为空***
	Account *Account `json:"account,omitempty"`

	// ProjectID 项目ID
	// ***不能为空***
	ProjectID string `json:"projectID,omitempty"`

	// Desc 账号描述
	Desc string `json:"desc,omitempty"`

	// Enable 是否开启
	// 默认为true
	Enable bool `json:"enable,omitempty"`

	// Creator cloud账号创建者
	// 默认为当前username
	Creator string `json:"creator,omitempty"`
}

// DeleteCloudAccountRequest update body.
//
// CloudID 云区域ID
// ***不能为空***
//
// AccountID 账号ID
// ***不能为空***
type DeleteCloudAccountRequest struct {
	// RequestID 请求ID
	RequestID string

	// CloudID 云区域ID
	// ***不能为空***
	CloudID string `json:"cloudID,omitempty"`

	// AccountID 账号ID
	// ***不能为空***
	AccountID string `json:"accountID,omitempty"`
}

// UpdateCloudAccountRequest  body (请注意这是覆盖修改)
//
// CloudID 云区域ID
// ***不能为空***
//
// AccountID 账号ID
// ***不能为空***
//
// AccountName 账号名称
// ***不能为空***
//
// ProjectID 项目ID
// ***不能为空***
//
// Account 账号
// ***不能为空***
type UpdateCloudAccountRequest struct {
	// RequestID 请求ID
	RequestID string

	// CloudID 云区域ID
	// ***不能为空***
	CloudID string `json:"cloudID,omitempty"`

	// AccountID 账号ID
	// ***不能为空***
	AccountID string `json:"accountID,omitempty"`

	// AccountName 账号名称
	// ***不能为空***
	AccountName string `json:"accountName,omitempty"`

	// Account 账号
	// ***不能为空***
	Account *Account `json:"account,omitempty"`

	// ProjectID 项目ID
	// ***不能为空***
	ProjectID string `json:"projectID,omitempty"`

	// Desc 账号描述
	Desc string `json:"desc,omitempty"`

	// Enable 是否开启
	// 默认为true
	Enable bool `json:"enable,omitempty"`

	// Updater 更新者
	// 默认为当前username
	Updater string `json:"updater,omitempty"`
}

// ListCloudAccountRequest 查询云账号请求，如果填写了目标字段，则组合目标信息查询，如果全为空则为查询全量信息
type ListCloudAccountRequest struct {
	// RequestID 请求ID
	RequestID string

	// CloudID 云区域ID
	// ***不能为空***
	CloudID string `json:"cloudID,omitempty"`

	// AccountID 账号ID
	AccountID string `json:"accountID,omitempty"`

	// ProjectID 项目ID
	ProjectID string `json:"projectID,omitempty"`

	// Operator operator云账号权限列表
	Operator string `json:"operator,omitempty"`
}

// Account 用于存储不同cloud的账号信息, 不同cloud格式兼容处理
type Account struct {
	/*
		腾讯云
	*/
	// SecretID 仅用于导入腾讯云凭证, 与其他凭证字段为互斥关系
	SecretID string `json:"secretID,omitempty"`
	// SecretKey 仅用于导入腾讯云凭证, 与其他凭证字段为互斥关系
	SecretKey string `json:"secretKey,omitempty"`

	/*
		gcp
	*/
	// ServiceAccountSecret Google Cloud service account的json字符串秘钥；(仅用于导入谷歌云凭证, 与其他凭证字段为互斥关系)
	ServiceAccountSecret string `json:"serviceAccountSecret,omitempty"`
	//GkeProjectID string `json:"gkeProjectID,omitempty"` //略

	/*
		azure
	*/
	// SubscriptionID Azure 订阅ID；(仅用于导入微软云凭证, 与其他凭证字段为互斥关系)
	SubscriptionID string `json:"subscriptionID,omitempty"`
	// ClientID Azure 租户ID；(仅用于导入微软云凭证, 与其他凭证字段为互斥关系)
	ClientID string `json:"clientID,omitempty"`
	// TenantID Azure Service Principal ClientID(需要创建独立的应用, 并为其分配权限);(仅用于导入微软云凭证, 与其他凭证字段为互斥关系)
	TenantID string `json:"tenantID,omitempty"`
	// ClientSecret Service Principal ClientSecret(需要创建独立的应用, 并为其分配权限)(仅用于导入微软云凭证, 与其他凭证字段为互斥关系)
	ClientSecret string `json:"clientSecret,omitempty"`
	//ResourceGroupName string `json:"resourceGroupName,omitempty"` //略
}
