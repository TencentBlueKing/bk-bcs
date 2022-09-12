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
	// NoError TODO
	NoError = 0
	// UnknownError TODO
	UnknownError = 300
	// UserError TODO
	UserError = 400
	// SysError TODO
	SysError = 500
	// NotFoundError TODO
	NotFoundError = 404

	// RecordNotFound TODO
	RecordNotFound = "记录不存在"
	// RecordNotFoundCode TODO
	RecordNotFoundCode = 404
	// JsonParseError TODO
	JsonParseError = "解析异常"
	// DBOperError TODO
	DBOperError = "DB操作异常"

	// TODO 禁用 APIError，该 ErrorCode 定义过于模糊，容易误用，考虑后续去除

	// ApiError TODO
	ApiError = "请求失败"
	// ApiErrorCode TODO
	ApiErrorCode = 40001
)
