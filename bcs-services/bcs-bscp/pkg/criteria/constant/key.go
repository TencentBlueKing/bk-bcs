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

package constant

// Note:
// This scope is used to define all the constant keys which is used inside and outside
// the BSCP system except sidecar.
const (
	// KitKey
	KitKey = "X-BSCP-KIT"

	// RidKey is request id header key.
	RidKey = "X-Bkapi-Request-Id"
	// RidKeyGeneric for generic header key
	RidKeyGeneric = "X-Request-Id"

	// LangKey is language key
	LangKey = "X-Bkapi-Language"

	// UserKey is operator name header key.
	UserKey = "X-Bkapi-User-Name"

	// AppCodeKey is blueking application code header key.
	AppCodeKey = "X-Bkapi-App-Code"

	// OperateWayKey is approve operate way header key.
	OperateWayKey = "X-Bscp-Operate-Way"

	// Space
	SpaceIDKey     = "X-Bkapi-Space-Id"
	SpaceTypeIDKey = "X-Bkapi-Space-Type-Id"
	BizIDKey       = "X-Bkapi-Biz-Id"
	AppIDKey       = "X-Bkapi-App-Id"

	// LanguageKey the language key word.
	LanguageKey = "HTTP_BLUEKING_LANGUAGE"

	// BKGWJWTTokenKey is blueking api gateway jwt header key.
	BKGWJWTTokenKey = "X-Bkapi-JWT" //nolint

	// BKTokenForTest is a token for test
	BKTokenForTest = "bk-token-for-test" //nolint:gosec

	// BKUserForTestPrefix is a user prefix for test
	BKUserForTestPrefix = "bk-user-for-test-"

	// BKSystemUser can be saved for user field in db when some operations come from bscp system itself
	BKSystemUser = "system"

	// ContentIDHeaderKey is common content sha256 id.
	ContentIDHeaderKey = "X-Bkapi-File-Content-Id"
	// PartNumHeaderKey is multipart upload part num key.
	PartNumHeaderKey = "X-Bscp-Part-Num"
	// MultipartUploadID is multipart upload id key.
	UploadIDHeaderKey = "X-Bscp-Upload-Id"
	// AppIDHeaderKey is app id.
	AppIDHeaderKey = "X-Bscp-App-Id"
	// TmplSpaceIDHeaderKey is template space id.
	//nolint:gosec
	TmplSpaceIDHeaderKey = "X-Bscp-Template-Space-Id"

	// TemplateVariablePrefix is the prefix for template variable name
	TemplateVariablePrefix = "bk_bscp_"

	// MaxRenderBytes is the max bytes to render for template config which is 2MB
	MaxRenderBytes = 2 * 1024 * 1024
)

// default resource
const (
	// DefaultTmplSpaceName is default template space name
	DefaultTmplSpaceName = "default_space"
	// DefaultTmplSpaceCNName is default template space chinese name
	DefaultTmplSpaceCNName = "默认空间"
	// DefaultTmplSpaceMemo is default template space memo
	DefaultTmplSpaceMemo = "this is default space"
	// DefaultTmplSetName is default template set name
	DefaultTmplSetName = "默认套餐"
	// DefaultTmplSetMemo is default template set memo
	DefaultTmplSetMemo = "当前空间下的所有模版"

	// DefaultLanguage is default language
	DefaultLanguage = "zh-cn"
)

// Note:
// 1. This scope defines keys which is used only by sidecar and feed server.
// 2. All the defined key should be prefixed with 'Side'.
const (
	// SidecarMetaKey defines the key to store the sidecar's metadata info.
	SidecarMetaKey = "sidecar-meta"
	// SideRidKey defines the incoming request id between sidecar and feed server.
	SideRidKey = "side-rid"
	// SideWorkspaceDir sidecar workspace dir name.
	SideWorkspaceDir = "bk-bscp"
)

const (
	// AuthLoginProviderKey is auth login provider
	AuthLoginProviderKey = "auth-login-provider"
	// AuthLoginUID is auth login uid
	AuthLoginUID = "auth-login-uid"
	// AuthLoginToken is auth login token
	AuthLoginToken = "auth-login-token" //nolint
)

var (
	// RidKeys support request_id keys
	RidKeys = []string{
		RidKey,
		RidKeyGeneric,
	}
)

// 文件状态，未命名版本服务配置项相比上一个版本的变化
const (
	// FileStateAdd 增加
	FileStateAdd = "ADD"
	// FileStateDelete 删除
	FileStateDelete = "DELETE"
	// FileStateRevise 修改
	FileStateRevise = "REVISE"
	// FileStateUnchange 不变
	FileStateUnchange = "UNCHANGE"
)

const (
	// MaxUploadTextFileSize 最大上传文件大小
	MaxUploadTextFileSize = 5 * 1024 * 1024
	// MaxConcurrentUpload 限制上传文件并发数
	MaxConcurrentUpload = 10
	// UploadBatchSize 上传时分批检测文件路冲突
	UploadBatchSize = 50
	// UploadTemporaryDirectory 上传的临时目录
	UploadTemporaryDirectory = "upload/files"
	// MB 字节
	MB = 1 << 20 // 1MB = 2^20 bytes
)

const (
	// LabelKeyAgentID is the key of agent id in bcs node labels.
	LabelKeyAgentID = "bkcmdb.tencent.com/bk-agent-id"
)

// itsm相关
const (
	// CreateCountSignApproveItsmServiceID used to create an itsm ticket
	// when creating a count sign approve in a shared cluster
	CreateCountSignApproveItsmServiceID = "create_count_sign_approve__itsm_service_id"
	// CreateOrSignApproveItsmServiceID used to create an itsm ticket
	// when creating an or sign approve in a shared cluster
	CreateOrSignApproveItsmServiceID = "create_or_sign_approve__itsm_service_id"

	// ItsmTicketStatusCreated enum string for created status
	ItsmTicketStatusCreated = "CREATED"
	// ItsmTicketStatusRevoked enum string for revoked status
	ItsmTicketStatusRevoked = "REVOKED"
	// ItsmTicketStatusRejected enum string for rejected status
	ItsmTicketStatusRejected = "REJECTED"
	// ItsmTicketStatusPassed enum string for passed status
	ItsmTicketStatusPassed = "PASSED"

	// ItsmTicketTypeCreate enum string for itsm ticket type create
	ItsmTicketTypeCreate = "CREATE"
	// ItsmTicketTypeUpdate enum string for itsm ticket type update
	ItsmTicketTypeUpdate = "UPDATE"
	// ItsmTicketTypeDelete enum string for itsm ticket type delete
	ItsmTicketTypeDelete = "DELETE"

	// ItsmApproveType 负责人审批workflow 节点类型
	ItsmApproveType = "APPROVAL"
	// ItsmCountSignServiceName 服务名称
	ItsmCountSignServiceName = "创建上线会签审批"
	// ItsmOrSignServiceName 服务名称
	ItsmOrSignServiceName = "创建上线或签审批"
	// ItsmApproveResult itsm已处理人的结果
	ItsmApproveResult = "已处理【负责人审批】(通过)"

	// 单据状态:

	// TicketRunningStatu 处理中
	TicketRunningStatu = "RUNNING"
	// TicketFinishedStatu 已结束
	TicketFinishedStatu = "FINISHED"
	// TicketTerminatedStatu 被终止
	TicketTerminatedStatu = "TERMINATED"
	// TicketSuspendedStatu 被挂起
	TicketSuspendedStatu = "SUSPENDED"
	// TicketRevokedStatu 被撤销
	TicketRevokedStatu = "REVOKED"
)
