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

// Package errorx xxx
package errorx

import commErr "github.com/Tencent/bk-bcs/bcs-common/common"

const (
	// Success 正常返回
	Success = 0
	// SuccessMsg 正常返回的消息
	SuccessMsg = "success"
	// ParamErr 参数校验失败
	ParamErr = commErr.AdditionErrorCode + 400
	// ParamErrMsg 参数校验失败消息
	ParamErrMsg = "params error"
	// InnerErr 内部服务异常
	InnerErr = commErr.AdditionErrorCode + 500
	// InnerErrMsg 内部服务异常消息
	InnerErrMsg = "inner error"
	// DBErr DB操作失败
	DBErr = commErr.AdditionErrorCode + 501
	// DBErrMsg DB操作失败消息
	DBErrMsg = "db error"
	// ClusterMsg 集群操作失败消息
	ClusterMsg = "cluster error"
	// UnauthErr 未认证/认证失败
	UnauthErr = commErr.AdditionErrorCode + 401
	// UnauthErrMsg 认证失败消息
	UnauthErrMsg = "auth error"
	// PermDeniedErr 无权限
	PermDeniedErr = commErr.AdditionErrorCode + 403
	// PermDeniedErrMsg 无权限消息
	PermDeniedErrMsg = "no permission"
	// IAMClientErr 构建 iam client异常
	IAMClientErr = commErr.AdditionErrorCode + 506
	// IAMClientErrMsg 构建 iam client 异常消息
	IAMClientErrMsg = "make iam client error"
	// IAMOPErr 错误的iam operation
	IAMOPErr = commErr.AdditionErrorCode + 507
	// IAMOPErrMsg 错误的 iam operation 异常信息
	IAMOPErrMsg = "iam op error"
	// RequestIAMErr 请求 IAM api 异常
	RequestIAMErr = commErr.AdditionErrorCode + 508
	// RequestIAMErrMsg 请求 IAM api 异常信息
	RequestIAMErrMsg = "request iam api error"
	// NotFoundHeaderUserErr header中没有发现username
	NotFoundHeaderUserErr = commErr.AdditionErrorCode + 406
	// NotFoundHeaderUserErrMsg header 中没有发现 username 异常信息
	NotFoundHeaderUserErrMsg = "not found username from header"
	// RequestCMDBErr 请求 cmdb api 异常
	RequestCMDBErr = commErr.AdditionErrorCode + 509
	// RequestCMDBErrMsg 请求 cmdb api 异常信息
	RequestCMDBErrMsg = "request cmdb api error"
	// NoMaintainerRoleErr 用户不为运维角色
	NoMaintainerRoleErr = commErr.AdditionErrorCode + 407
	// NoMaintainerRoleErrMsg 用户不为运维角色异常信息
	NoMaintainerRoleErrMsg = "user is not biz maintainer role"
	// RequestBKSSMErr 请求 bk-ssm api 异常
	RequestBKSSMErr = commErr.AdditionErrorCode + 510
	// RequestBKSSMMsg 请求 bk-ssm api 异常信息
	RequestBKSSMMsg = "request bk ssm api error"
	// RequestBCSCCErr 请求 bcs cc api 异常
	RequestBCSCCErr = commErr.AdditionErrorCode + 511
	// RequestBCSCCErrMsg 请求 bcs cc api 异常信息
	RequestBCSCCErrMsg = "request bcs cc api error"
	// RequestITSMErr 请求 bk itsm api 异常
	RequestITSMErr = commErr.AdditionErrorCode + 512
	// RequestBkMonitorErr 请求 bk monitor api 异常
	RequestBkMonitorErr = commErr.AdditionErrorCode + 513
	// RequestBkMonitorErrMsg 请求 bk monitor api 异常信息
	RequestBkMonitorErrMsg = "request bk monitor api error"
	// RequestITSMErrMsg 请求 bk itsm api 异常信息
	RequestITSMErrMsg = "request bk itsm api error"
	// RequestTaskErr 构建任务异常
	RequestTaskErr = commErr.AdditionErrorCode + 514
	// RequestTaskErrMsg 构建任务异常信息
	RequestTaskErrMsg = "build task error"
	// RequestCheckQuotaStatusErr 检测额度状态异常
	RequestCheckQuotaStatusErr = commErr.AdditionErrorCode + 515
	// RequestQuotaStatusErrMsg 检测额度状态异常信息
	RequestQuotaStatusErrMsg = "check quota status error"
	// NoPermissionErr 无权限
	NoPermissionErr = 40403
	// ProjectNotExistsErr 项目不存在
	ProjectNotExistsErr = 40404
	// ProjectQuotaNotExistsErr 项目配额不存在
	ProjectQuotaNotExistsErr = 40405
)
