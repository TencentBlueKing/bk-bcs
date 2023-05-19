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
 *
 */

package bklog

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
)

// ListLogCollectors list log collectors
func ListLogCollectors(ctx context.Context, clusterID, spaceUID string) ([]ListBCSCollectorRespData, error) {
	url := fmt.Sprintf("%s/list_bcs_collector", config.G.BKLog.APIServer)
	authInfo, err := component.GetBKAPIAuthorization()
	if err != nil {
		return nil, err
	}
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bkapi-Authorization", authInfo).
		SetQueryParam("bcs_cluster_id", clusterID).
		SetQueryParam("space_uid", spaceUID).
		Get(url)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, errors.Errorf("http code %d != 200, body: %s", resp.StatusCode(), resp.Body())
	}

	result := &ListBCSCollectorResp{}
	err = json.Unmarshal(resp.Body(), result)
	if err != nil {
		return nil, errors.Errorf("unmarshal resp body error, %s", err.Error())
	}

	if !result.IsSuccess() {
		return nil, errors.Errorf("list log collectors error, code: %d, message: %s, request_id: %s",
			result.GetCode(), result.Message, result.RequestID)
	}
	return result.Data, nil
}

// CreateLogCollectors create log collectors
func CreateLogCollectors(ctx context.Context, lc *LogCollector) (*CreateBCSCollectorRespData, error) {
	url := fmt.Sprintf("%s/create_bcs_collector", config.G.BKLog.APIServer)
	authInfo, err := component.GetBKAPIAuthorization()
	if err != nil {
		return nil, err
	}
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bkapi-Authorization", authInfo).
		SetBody(lc).
		Post(url)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, errors.Errorf("http code %d != 200, body: %s", resp.StatusCode(), resp.Body())
	}

	result := &CreateBCSCollectorResp{}
	err = json.Unmarshal(resp.Body(), result)
	if err != nil {
		return nil, errors.Errorf("unmarshal resp body error, %s", err.Error())
	}

	if !result.IsSuccess() {
		return nil, errors.Errorf("create log collectors error, code: %d, message: %s, request_id: %s",
			result.GetCode(), result.Message, result.RequestID)
	}
	return &result.Data, nil
}

// UpdateLogCollectors update log collectors
func UpdateLogCollectors(ctx context.Context, ruleID int, lc *LogCollector) (*UpdateBCSCollectoRespData, error) {
	url := fmt.Sprintf("%s/update_bcs_collector/%d", config.G.BKLog.APIServer, ruleID)
	authInfo, err := component.GetBKAPIAuthorization()
	if err != nil {
		return nil, err
	}
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bkapi-Authorization", authInfo).
		SetBody(lc).
		Post(url)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, errors.Errorf("http code %d != 200, body: %s", resp.StatusCode(), resp.Body())
	}

	result := &UpdateBCSCollectorResp{}
	err = json.Unmarshal(resp.Body(), result)
	if err != nil {
		return nil, errors.Errorf("unmarshal resp body error, %s", err.Error())
	}

	if !result.IsSuccess() {
		return nil, errors.Errorf("update log collectors error, code: %d, message: %s, request_id: %s",
			result.GetCode(), result.Message, result.RequestID)
	}
	return &result.Data, nil
}

// DeleteLogCollectors delete log collectors
func DeleteLogCollectors(ctx context.Context, ruleID int) error {
	url := fmt.Sprintf("%s/delete_bcs_collector/%d", config.G.BKLog.APIServer, ruleID)
	authInfo, err := component.GetBKAPIAuthorization()
	if err != nil {
		return err
	}
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bkapi-Authorization", authInfo).
		Delete(url)

	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return errors.Errorf("http code %d != 200, body: %s", resp.StatusCode(), resp.Body())
	}

	result := &BaseResp{}
	err = json.Unmarshal(resp.Body(), result)
	if err != nil {
		return errors.Errorf("unmarshal resp body error, %s", err.Error())
	}

	if !result.IsSuccess() {
		return errors.Errorf("delete log collectors error, code: %d, message: %s, request_id: %s",
			result.GetCode(), result.Message, result.RequestID)
	}
	return nil
}
