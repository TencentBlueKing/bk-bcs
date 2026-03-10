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

package component

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/utils"
	apputils "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
)

// ListTenantResp list tenant response
type ListTenantResp struct {
	Data []BkTenant `json:"data"`
}

// BkTenant bk tenant
type BkTenant struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// BKUserAPIServer bk user api server
var BKUserAPIServer = os.Getenv("BKUSER_API_SERVER")

// ListTenant list bk tenant
func ListTenant(ctx context.Context) ([]BkTenant, error) {
	url := fmt.Sprintf("%s/%s", BKUserAPIServer, "api/v3/open/tenants/")
	// generate bk api auth header, X-Bkapi-Authorization
	authInfo, err := GetBKAPIAuthorization("")
	if err != nil {
		return nil, err
	}
	resp, err := GetClient().R().
		SetContext(ctx).
		SetHeader(apputils.HeaderTenantID, "system").
		SetHeader("X-Bkapi-Authorization", authInfo).
		SetHeaders(utils.GetLaneIDByCtx(ctx)).
		Get(url)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, errors.Errorf("http code %d != 200, body: %s", resp.StatusCode(), resp.Body())
	}

	result := &ListTenantResp{}
	err = json.Unmarshal(resp.Body(), result)
	if err != nil {
		return nil, errors.Errorf("unmarshal resp body error, %s", err.Error())
	}

	return result.Data, nil
}
