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

// Package errf provides bscp common error.
package errf

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

// NOTE: 错误码规则
// 40号段 + 5位错误码共7位
// 注意：
// - 特殊错误码, 4030403（未授权）, 内部保留

// common error code.
const (
	OK               int32 = 0
	PermissionDenied int32 = 4030403
)

// Note: 之前的错误码后续将统一改成蓝鲸规范错误码和bscp专用错误码，新代码使用新错误码规范
// this scope's error code ranges at [4000000, 4089999], and works for all the scenario
// except sidecar related scenario.
const (
	/*
		// Unknown is unknown error, it is always used when an
		// error is wrapped, but the error code is not parsed.
		Unknown int32 = 4000000
	*/

	// InvalidParameter means the request parameter  is invalid
	InvalidParameter int32 = 4000001

	/*
		// Aborted means the request is aborted because of some unexpected exceptions.
		//Aborted int32 = 4000002
	*/

	// DBOpFailed means read or write db failed
	DBOpFailed int32 = 4000003
	// RecordNotFound means resource not exist.
	RecordNotFound int32 = 4000005
	// RelatedResNotExist means attachment resource is not exist.
	RelatedResNotExist int32 = 4000006
	// DoAuthorizeFailed try to do user's operate authorize, but got an error,
	// so we do not know if the user has the permission or not.
	DoAuthorizeFailed int32 = 4000007
	// TooManyRequest means the incoming request have already exceeded the max limit.
	// and the incoming request is rejected.
	TooManyRequest int32 = 4000008
	// UnHealth means service health check failed, current service has problem.
	UnHealth int32 = 4000009
	// ErrGroupAlreadyPublished means the group has already been published in specified app.
	ErrGroupAlreadyPublished int32 = 4000010
)

// Note: Sidecar related error code scope, ranges at [4090000, 409999]
// all the error code should be prefixed with 'Side' lateral.
const (
	// SideInvalidMeta means the requested metadata from sidecar is invalid, which can
	// be one of the scenario as follows:
	// 1. requested biz is not exist.
	// 2. requested app is not exist.
	SideInvalidMeta int32 = 4090000
)

// 蓝鲸错误码规范，粗粒度的错误分类
const (
	// InvalidArgument 参数不符合参数格式
	InvalidArgument int32 = 10000
	// InvalidRequest 参数符合参数格式，但参数不符合业务规则
	InvalidRequest int32 = 10001
	// OutOfRange 客户端指定了无效范围
	OutOfRange int32 = 10002
	// FailedPrecondition 请求无法在当前系统状态下执行，例如删除非空目录
	FailedPrecondition int32 = 10003
	// Unauthenticated 未提供身份认证凭证
	Unauthenticated int32 = 10004
	// IamNoPermission 权限中心没有相关权限(有协议要求)
	IamNoPermission int32 = 10005
	// NoPermission 没有相关权限(非权限中心)
	NoPermission int32 = 10006
	// NotFound 资源不存在
	NotFound int32 = 10007
	// AlreadyExists 客户端尝试创建的资源已存在
	AlreadyExists int32 = 10008
	// Aborted 并发冲突，例如读取/修改/写入冲突
	Aborted int32 = 10009
	// RatelimitExceed 超过频率限制
	RatelimitExceed int32 = 10010
	// ResourceExhausted 资源配额不足
	ResourceExhausted int32 = 10011
	// Internal 出现内部服务器错误
	Internal int32 = 10012
	// Unknown 出现未知的服务器错误
	Unknown int32 = 10013
	// NotImplemented API方法未通过服务器实现
	NotImplemented int32 = 10014
)

// bscp专用错误码，细粒度具体场景的错误码
const (
	// AppNotExists means the app is not exist.
	AppNotExists int32 = 11000
)

var (
	// BscpCodeMap bscp错误码->错误字符串映射
	BscpCodeMap = map[int32]string{
		// 兼容grpc错误码
		int32(codes.Canceled):           "CANCELED",
		int32(codes.Unknown):            "UNKNOWN",
		int32(codes.InvalidArgument):    "INVALID_ARGUMENT",
		int32(codes.DeadlineExceeded):   "DEADLINE_EXCEEDED",
		int32(codes.NotFound):           "NOT_FOUND",
		int32(codes.AlreadyExists):      "ALREADY_EXISTS",
		int32(codes.PermissionDenied):   "PERMISSION_DENIED",
		int32(codes.ResourceExhausted):  "RESOURCE_EXHAUSTED",
		int32(codes.FailedPrecondition): "FAILED_PRECONDITION",
		int32(codes.Aborted):            "ABORTED",
		int32(codes.OutOfRange):         "OUT_OF_RANGE",
		int32(codes.Unimplemented):      "UNIMPLEMENTED",
		int32(codes.Internal):           "INTERNAL",
		int32(codes.Unavailable):        "UNAVAILABLE",
		int32(codes.DataLoss):           "DATA_LOSS",
		int32(codes.Unauthenticated):    "UNAUTHENTICATED",

		// 蓝鲸规范错误码
		InvalidArgument:    "INVALID_ARGUMENT",
		InvalidRequest:     "INVALID_REQUEST",
		OutOfRange:         "OUT_OF_RANGE",
		FailedPrecondition: "FAILED_PRECONDITION",
		Unauthenticated:    "UNAUTHENTICATED",
		IamNoPermission:    "IAM_NO_PERMISSION",
		NoPermission:       "NO_PERMISSION",
		NotFound:           "NOT_FOUND",
		AlreadyExists:      "ALREADY_EXISTS",
		Aborted:            "ABORTED",
		RatelimitExceed:    "RATELIMIT_EXCEED",
		ResourceExhausted:  "RESOURCE_EXHAUSTED",
		Internal:           "INTERNAL",
		Unknown:            "UNKNOWN",
		NotImplemented:     "NOT_IMPLEMENTED",

		// bscp专用错误码
		AppNotExists: "APP_NOT_EXISTS",
	}

	// BscpStatusMap bscp错误码->状态映射
	BscpStatusMap = map[int32]int{
		// 兼容grpc错误码
		int32(codes.Canceled):           http.StatusBadRequest,
		int32(codes.Unknown):            http.StatusBadRequest,
		int32(codes.InvalidArgument):    http.StatusBadRequest,
		int32(codes.DeadlineExceeded):   http.StatusBadRequest,
		int32(codes.NotFound):           http.StatusNotFound,
		int32(codes.AlreadyExists):      http.StatusBadRequest,
		int32(codes.PermissionDenied):   http.StatusForbidden,
		int32(codes.ResourceExhausted):  http.StatusBadRequest,
		int32(codes.FailedPrecondition): http.StatusBadRequest,
		int32(codes.Aborted):            http.StatusBadRequest,
		int32(codes.OutOfRange):         http.StatusBadRequest,
		int32(codes.Unimplemented):      http.StatusBadRequest,
		int32(codes.Internal):           http.StatusBadRequest,
		int32(codes.Unavailable):        http.StatusBadRequest,
		int32(codes.DataLoss):           http.StatusBadRequest,
		int32(codes.Unauthenticated):    http.StatusUnauthorized,

		// 蓝鲸规范错误码
		InvalidArgument:    http.StatusBadRequest,
		InvalidRequest:     http.StatusBadRequest,
		OutOfRange:         http.StatusBadRequest,
		FailedPrecondition: http.StatusBadRequest,
		Unauthenticated:    http.StatusUnauthorized,
		IamNoPermission:    http.StatusForbidden,
		NoPermission:       http.StatusForbidden,
		NotFound:           http.StatusNotFound,
		AlreadyExists:      http.StatusConflict,
		Aborted:            http.StatusConflict,
		RatelimitExceed:    http.StatusTooManyRequests,
		ResourceExhausted:  http.StatusTooManyRequests,
		Internal:           http.StatusInternalServerError,
		Unknown:            http.StatusInternalServerError,
		NotImplemented:     http.StatusNotImplemented,

		// bscp专用错误码
		AppNotExists: http.StatusNotFound,
	}
)
