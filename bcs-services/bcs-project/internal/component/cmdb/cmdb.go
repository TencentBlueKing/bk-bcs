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

package cmdb

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util/errorx"
)

var (
	defaultTimeout         = 10
	defaultSupplierAccount = "tencent"
	searchBizPath          = "/api/c/compapi/v2/cc/search_business/"
)

type cmdbResp struct {
	Code      int                    `json:"code"`
	Result    bool                   `json:"result"`
	Message   string                 `json:"message"`
	RequestID string                 `json:"request_id"`
	Data      map[string]interface{} `json:"data"`
}

// IsMaintainer 校验用户是否为指定业务的运维
func IsMaintainer(username string, bizID string) (bool, error) {
	resp, err := SearchBizByUserAndID(username, bizID)
	if err != nil {
		return false, err
	}
	if resp.Code != errorx.Success {
		return false, errorx.NewRequestCMDBErr(resp.Message)
	}
	// 判断是否存在当前用户为业务运维角色的业务
	// NOTE: count 为float64类型
	if resp.Data["count"].(float64) > 0 {
		return true, nil
	}
	return false, errorx.NewNoMaintainerRoleErr()
}

// SearchBizByUserAndID 通过用户和业务ID，查询业务
func SearchBizByUserAndID(username string, bizID string) (*cmdbResp, error) {
	cmdbConf := config.GlobalConf.CMDB
	reqUrl := fmt.Sprintf("%s%s", cmdbConf.Host, searchBizPath)
	// 获取超时时间
	timeout := getTimeout()
	headers := map[string]string{"Content-Type": "application/json"}
	bizIDInt, _ := strconv.Atoi(bizID)
	// 组装请求参数
	req := getReq(cmdbConf, reqUrl, username, bizIDInt)
	// 获取返回数据
	body, err := component.Request(req, timeout, cmdbConf.Proxy, headers)
	if err != nil {
		return nil, errorx.NewRequestCMDBErr(err)
	}
	// 解析返回的body
	var resp cmdbResp
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		logging.Error("parse search biz body error, body: %v", body)
		return nil, err
	}
	return &resp, nil
}

func getReq(c config.CMDBConfig, reqUrl string, username string, bizIDInt int) gorequest.SuperAgent {
	return gorequest.SuperAgent{
		Url:    reqUrl,
		Method: "POST",
		Data: map[string]interface{}{
			"condition": map[string]interface{}{
				"bk_biz_id":         bizIDInt,
				"bk_biz_maintainer": username,
			},
			"bk_supplier_account": getSupplierAccount(),
			"bk_app_code":         config.GlobalConf.App.Code,
			"bk_app_secret":       config.GlobalConf.App.Secret,
			"bk_username":         username,
		},
		Debug: c.Debug,
	}
}

func getTimeout() int {
	timeout := config.GlobalConf.CMDB.Timeout
	if timeout == 0 {
		return defaultTimeout
	}
	return timeout
}

// 获取开发商账号
func getSupplierAccount() string {
	supplierAccount := config.GlobalConf.CMDB.BKSupplierAccount
	if supplierAccount == "" {
		return defaultSupplierAccount
	}
	return supplierAccount
}
