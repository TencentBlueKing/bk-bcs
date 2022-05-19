/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
	// DbErrMsg DB操作失败消息
	DBErrMsg = "db error"
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
	// IAMClientErrMsg ...
	IAMClientErrMsg = "make iam client error"
	// IAMOPErr 错误的iam operation
	IAMOPErr = commErr.AdditionErrorCode + 507
	// IAMOPErrMsg ...
	IAMOPErrMsg = "iam op error"
	// RequestIAMErr 请求 IAM api 异常
	RequestIAMErr = commErr.AdditionErrorCode + 508
	// RequestIAMErrMsg ...
	RequestIAMErrMsg = "request iam api error"
	// NotFoundHeaderUserErr header中没有发现username
	NotFoundHeaderUserErr = commErr.AdditionErrorCode + 406
	// NotFoundUserFromHeaderMsg
	NotFoundHeaderUserErrMsg = "not found username from header"
	// RequestCMDBErr 请求 cmdb api 异常
	RequestCMDBErr = commErr.AdditionErrorCode + 509
	// RequestCMDBErrMsg ...
	RequestCMDBErrMsg = "request iam api error"
	// NoMaintainerRoleErr 用户不为运维角色
	NoMaintainerRoleErr = commErr.AdditionErrorCode + 407
	// NoMaintainerRoleErrMsg ...
	NoMaintainerRoleErrMsg = "user is not biz maintainer role"
	// RequestBCSCCErr ...
	RequestBCSCCErr = commErr.AdditionErrorCode + 510
	// RequestBCSCCErrMsg ...
	RequestBCSCCErrMsg = "request bcs cc api error"
)
