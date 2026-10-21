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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/bkuser"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/tenant"
)

const (
	migrateItsm = "/data/bcs/bcs-project-manager/itsmv4.json"
	migratePath = "/api/v1/system/migrate/"

	fileName = "itsmv4.json"
)

// MigrateResp resp
type MigrateResp struct {
	Result  bool        `json:"result"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    MigrateData `json:"data"`
}

// MigrateData data
type MigrateData struct {
	Message string `json:"message"`
}

// MigrateSystem migrate system workflow
func MigrateSystem(ctx context.Context, content []byte) error {
	tenantId := tenant.GetTenantIdFromContext(ctx)

	itsmConf := config.GlobalConf.ITSM
	host := itsmConf.GatewayHost
	reqURL := fmt.Sprintf("%s%s", host, migratePath)

	// auth headers: ctx store tenant info
	headers, err := bkuser.GetAuthHeader(ctx)
	if err != nil {
		logging.Error("MigrateSystem get auth header failed, %s", err.Error())
		return errorx.NewRequestITSMErr(err.Error())
	}

	// 创建 Resty 客户端
	request := resty.New().R()
	request.SetDebug(true)
	request.SetHeaders(headers)
	request.SetMultipartFormData(map[string]string{
		"tenant_id": tenantId,
	})
	request.SetMultipartField("file", fileName, "text/plain", bytes.NewReader(content))
	resp, err := request.Post(reqURL)

	if err != nil {
		return err
	}

	if resp.RawResponse.StatusCode < 200 || resp.RawResponse.StatusCode >= 300 {
		return fmt.Errorf("MigrateSystem api failed return statusCode: %d", resp.StatusCode)
	}

	// 解析返回的body
	migrateData := &MigrateResp{}
	if err := json.Unmarshal(resp.Body(), migrateData); err != nil {
		logging.Error("parse itsm body error, body: %v", string(resp.Body()))
		return err
	}
	if !migrateData.Result {
		logging.Error("request migrate itsm system %v failed, msg: %s", migrateData.Code, migrateData.Message)
		return errors.New(migrateData.Message)
	}

	return nil
}
