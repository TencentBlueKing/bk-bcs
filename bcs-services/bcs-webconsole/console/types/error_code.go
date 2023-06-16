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
 *
 */

package types

const (
	// NoError ..
	NoError = 0
	// UnknownError ..
	UnknownError = 300
	// UserError ..
	UserError = 400
	// SysError ..
	SysError = 500
	// NotFoundError ..
	NotFoundError = 404

	// RecordNotFound ..
	RecordNotFound = "记录不存在"
	// RecordNotFoundCode ..
	RecordNotFoundCode = 404
	// JsonParseError ..
	JsonParseError = "解析异常"
	// DBOperError ..
	DBOperError = "DB操作异常"

	// Note 禁用 APIError，该 ErrorCode 定义过于模糊，容易误用，考虑后续去除

	// ApiError ..
	ApiError = "请求失败"
	// ApiErrorCode ..
	ApiErrorCode = 40001
)
