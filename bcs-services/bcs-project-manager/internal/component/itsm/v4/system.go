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

// Package v4 xxx
package v4

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/bkuser"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
)

var (
	createSystemPath = "/api/v1/system/create"
	timeout          = 10
)

// CreateSystemReq xxx
type CreateSystemReq struct {
	// Name 系统名称
	Name string `json:"name"`
	// Code 系统标识
	Code uint32 `json:"code"`
	// Name 系统资源调用token，如果资源属于系统，需要在请求头中增加SYSTEM-TOKEN
	Token string `json:"token"`
	// Desc 系统描述
	Desc string `json:"desc"`
}

// CreateSystemResp xxx
type CreateSystemResp struct {
	component.CommonResp
	Data SystemData `json:"data"`
}

// SystemData system data
type SystemData struct {
	// Name 系统名称
	Name string `json:"name"`
	// Code 系统标识
	Code uint32 `json:"code"`
	// Desc 系统描述
	Desc string `json:"desc"`
}

// CreateSystem create itsmv4 system
func CreateSystem(ctx context.Context, data CreateSystemReq) (*SystemData, error) {
	itsmConf := config.GlobalConf.ITSM
	host := itsmConf.GatewayHost
	if itsmConf.External {
		host = itsmConf.Host
	}
	reqURL := fmt.Sprintf("%s%s", host, createSystemPath)
	req := gorequest.SuperAgent{
		Url:    reqURL,
		Method: "POST",
		Data: map[string]interface{}{
			"name": data.Name,
			"code": data.Code,
			"desc": data.Desc,
		},
	}

	// auth headers: ctx store tenant info
	headers, err := bkuser.GetAuthHeader(ctx)
	if err != nil {
		logging.Error("CreateSystem get auth header failed, %s", err.Error())
		return nil, errorx.NewRequestITSMErr(err.Error())
	}

	// 请求API
	body, err := component.Request(req, timeout, "", headers)
	if err != nil {
		logging.Error("request create itsm system %v failed, error: %s", data.Name, err.Error())
		return nil, errorx.NewRequestITSMErr(err.Error())
	}

	// 解析返回的body
	resp := &CreateSystemResp{}
	if err := json.Unmarshal([]byte(body), resp); err != nil {
		logging.Error("parse itsm body error, body: %v", body)
		return nil, err
	}
	if resp.Code != 0 {
		logging.Error("request create itsm system %v failed, msg: %s", data.Name, resp.Message)
		return nil, errors.New(resp.Message)
	}
	return &resp.Data, nil
}
