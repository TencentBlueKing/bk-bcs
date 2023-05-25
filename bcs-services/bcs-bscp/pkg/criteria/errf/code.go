/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package errf

// NOTE: 错误码规则
// 40号段 + 5位错误码共7位
// 注意：
// - 特殊错误码, 4030403（未授权）, 内部保留

// common error code.
const (
	OK               int32 = 0
	PermissionDenied int32 = 4030403
)

// Note:
// this scope's error code ranges at [4000000, 4089999], and works for all the scenario
// except sidecar related scenario.
const (
	// Unknown is unknown error, it is always used when an
	// error is wrapped, but the error code is not parsed.
	Unknown int32 = 4000000
	// InvalidParameter means the request parameter  is invalid
	InvalidParameter int32 = 4000001
	// Aborted means the request is aborted because of some unexpected exceptions.
	Aborted int32 = 4000002
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

	// endOfThisScope is a flag to show this scope's error code's end.
	endOfThisScope int32 = 4089999
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
