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

// Package types NOTES
package types

import "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"

// CommParams defines esb request common parameter
type CommParams struct {
	AppCode   string `json:"bk_app_code"`
	AppSecret string `json:"bk_app_secret"`
	UserName  string `json:"bk_username"`
}

// GetCommParams generate esb request common parameter from esb config and request user
func GetCommParams(config *cc.Esb) *CommParams {
	return &CommParams{
		AppCode:   config.AppCode,
		AppSecret: config.AppSecret,
		UserName:  config.User,
	}
}

// BaseResponse is esb http base response.
type BaseResponse struct {
	Result  bool   `json:"result"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Rid     string `json:"request_id"`
}
