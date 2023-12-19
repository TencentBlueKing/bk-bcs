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

package bklogin

import "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/thirdparty/esb/types"

// nolint
const (
	codeOK         = 0       // 成功
	codeNotLogin   = 1302100 // 用户认证失败，即用户登录态无效
	codeNotHasPerm = 1302403 // 用户认证成功，但用户无应用访问权限
)

// IsLoginResult .
type IsLoginResult struct {
	BKUsername string `json:"bk_username"`
}

// IsLoginResp is bklogin isLogin response.
type IsLoginResp struct {
	types.BaseResponse
	IsLoginResult `json:"data"`
}
