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

// Package bkuser bk user related types
package bkuser

// AuthInfo auth app
type AuthInfo struct {
	// BkAppCode bk app code
	BkAppCode string `json:"bk_app_code"`
	// BkAppSecret bk app secret
	BkAppSecret string `json:"bk_app_secret"`
}

// Options for client
type Options struct {
	AppCode   string
	AppSecret string
	Server    string
	Debug     bool
}

// LookupVirtualUserRsp resp xxx
type LookupVirtualUserRsp struct {
	Data []VirtualUserData `json:"data"`
}

// VirtualUserData virtual user data
type VirtualUserData struct {
	BkUsername  string `json:"bk_username"`
	LoginName   string `json:"login_name"`
	DisplayName string `json:"display_name"`
}
