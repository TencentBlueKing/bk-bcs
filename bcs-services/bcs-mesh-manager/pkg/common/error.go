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

package common

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
)
