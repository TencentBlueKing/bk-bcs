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

// Package ccv3 xxx
package ccv3

const (
	searchBusinessApi = "/api/c/compapi/v2/cc/search_business/"
)

// Interface defines the interface to bkcc
type Interface interface {
	SearchBusiness(bkBizIds []int64) ([]CCBusiness, error)
}

type appInfo struct {
	AppCode   string `json:"bk_app_code"`   // app code for api from bk center
	AppSecret string `json:"bk_app_secret"` // app secret for api from bk center
	Operator  string `json:"bk_username"`   // user name
}

// CCBusiness business for cc
type CCBusiness struct {
	BkBizId      int64  `json:"bk_biz_id"`
	BkBizName    string `json:"bk_biz_name"`
	BkMaintainer string `json:"bk_biz_maintainer"`
}
