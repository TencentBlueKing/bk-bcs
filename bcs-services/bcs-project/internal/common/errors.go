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

// TODO: 标识错误码及消息

package common

import (
	errorCode "github.com/Tencent/bk-bcs/bcs-common/common"
)

const (
	// BcsProjectSuccess 正常返回
	BcsProjectSuccess = 0
	// BcsProjectSuccessMsg 正常返回的消息
	BcsProjectSuccessMsg = "success"
	// BcsProjectParamErr 参数校验失败
	BcsProjectParamErr = errorCode.AdditionErrorCode + 400
	// BcsProjectParamErrMsg 参数校验失败消息
	BcsProjectParamErrMsg = "params error"
	// BcsInnerErr 内部服务异常
	BcsInnerErr = errorCode.AdditionErrorCode + 500
	// BcsInnerErrMsg 内部服务异常消息
	BcsInnerErrMsg = "inner error"
	// BcsProjectDbErr DB操作失败
	BcsProjectDbErr = errorCode.AdditionErrorCode + 501
	// BcsProjectDbErrMsg DB操作失败消息
	BcsProjectDbErrMsg = "db error"
)
