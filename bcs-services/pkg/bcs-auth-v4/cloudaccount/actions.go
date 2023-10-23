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

// Package cloudaccount xxx
package cloudaccount

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
)

// cloudAccount resource actions
const (
	// AccountManage xxx
	AccountManage iam.ActionID = "cloud_account_manage"
	// AccountUse xxx
	AccountUse iam.ActionID = "cloud_account_use"
	// AccountCreate xxx
	AccountCreate iam.ActionID = "cloud_account_create"
)

// ActionIDNameMap map ActionID to name
var ActionIDNameMap = map[iam.ActionID]string{
	AccountManage: "云账号管理",
	AccountUse:    "云账号使用",
	AccountCreate: "云账号创建",
}
