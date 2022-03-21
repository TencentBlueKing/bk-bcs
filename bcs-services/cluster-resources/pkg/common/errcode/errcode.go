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

package errcode

const (
	// None 没有错误
	None = 0
	// Default 默认错误码
	Default = 1
	// ValidateErr 参数校验失败
	ValidateErr = 2
	// Unsupported 功能未支持
	Unsupported = 3
	// NoPerm 无权限
	NoPerm = 4
)
