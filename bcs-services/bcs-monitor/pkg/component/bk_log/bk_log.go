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

// Package bklog log
package bklog

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
)

// ListLogCollectors list log collectors
func ListLogCollectors(ctx context.Context, clusterID, spaceUID string) ([]ListBCSCollectorRespData, error) {
	bcsPath := "list_bcs_collector"
	withoutBcsPath := "list_bcs_collector_without_rule"
	g, ctx := errgroup.WithContext(ctx)
	var result1 []ListBCSCollectorRespData
	var result2 []ListBCSCollectorRespData
	g.Go(func() error {
		var err error
		// list log collectors with bcsPath
		result1, err = ListLogCollectorsWithPath(ctx, clusterID, spaceUID, bcsPath)
		if err != nil {
			return err
		}
		return nil
	})
	g.Go(func() error {
		var err error
		// list log collectors without bcsPath
		result2, err = ListLogCollectorsWithPath(ctx, clusterID, spaceUID, withoutBcsPath)
		if err != nil {
			return err
		}
		for i := range result2 {
			result2[i].FromBKLog = true
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return append(result1, result2...), nil
}

// ListLogCollectorsWithPath list log collectors
func ListLogCollectorsWithPath(ctx context.Context, clusterID, spaceUID string,
	path string) ([]ListBCSCollectorRespData, error) {
	url := fmt.Sprintf("%s/%s", config.G.BKLog.APIServer, path)
	// generate bk api auth header, X-Bkapi-Authorization
	authInfo, err := component.GetBKAPIAuthorization("")
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
func CreateLogCollectors(ctx context.Context, req *CreateBCSCollectorReq) (*CreateBCSCollectorRespData, error) {
	url := fmt.Sprintf("%s/create_bcs_collector", config.G.BKLog.APIServer)
	// generate bk api auth header, X-Bkapi-Authorization
	authInfo, err := component.GetBKAPIAuthorization(req.Username)
	if err != nil {
		return nil, err
	}
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bkapi-Authorization", authInfo).
		SetBody(req).
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
func UpdateLogCollectors(ctx context.Context, ruleID int, req *UpdateBCSCollectorReq) (*UpdateBCSCollectoRespData,
	error) {
	url := fmt.Sprintf("%s/update_bcs_collector/%d", config.G.BKLog.APIServer, ruleID)
	// generate bk api auth header, X-Bkapi-Authorization
	authInfo, err := component.GetBKAPIAuthorization(req.Username)
	if err != nil {
		return nil, err
	}
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bkapi-Authorization", authInfo).
		SetBody(req).
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
	// generate bk api auth header, X-Bkapi-Authorization
	authInfo, err := component.GetBKAPIAuthorization("")
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

// RetryLogCollectors retry log collectors
func RetryLogCollectors(ctx context.Context, ruleID int, username string) error {
	url := fmt.Sprintf("%s/retry_bcs_collector/%d", config.G.BKLog.APIServer, ruleID)
	// generate bk api auth header, X-Bkapi-Authorization
	authInfo, err := component.GetBKAPIAuthorization(username)
	if err != nil {
		return err
	}
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bkapi-Authorization", authInfo).
		Post(url)

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
		return errors.Errorf("retry log collectors error, code: %d, message: %s, request_id: %s",
			result.GetCode(), result.Message, result.RequestID)
	}
	return nil
}

// StartLogCollectors start log collectors
func StartLogCollectors(ctx context.Context, ruleID int, username string) error {
	url := fmt.Sprintf("%s/start_bcs_collector/%d", config.G.BKLog.APIServer, ruleID)
	// generate bk api auth header, X-Bkapi-Authorization
	authInfo, err := component.GetBKAPIAuthorization(username)
	if err != nil {
		return err
	}
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bkapi-Authorization", authInfo).
		Post(url)

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
		return errors.Errorf("start log collectors error, code: %d, message: %s, request_id: %s",
			result.GetCode(), result.Message, result.RequestID)
	}
	return nil
}

// StopLogCollectors stop log collectors
func StopLogCollectors(ctx context.Context, ruleID int, username string) error {
	url := fmt.Sprintf("%s/stop_bcs_collector/%d", config.G.BKLog.APIServer, ruleID)
	// generate bk api auth header, X-Bkapi-Authorization
	authInfo, err := component.GetBKAPIAuthorization(username)
	if err != nil {
		return err
	}
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bkapi-Authorization", authInfo).
		Post(url)

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
		return errors.Errorf("stop log collectors error, code: %d, message: %s, request_id: %s",
			result.GetCode(), result.Message, result.RequestID)
	}
	return nil
}

// HasLog check indexSetID has log
func HasLog(ctx context.Context, indexSetID int) (bool, error) {
	url := fmt.Sprintf("%s/esquery_search", config.G.BKLog.APIServer)
	// generate bk api auth header, X-Bkapi-Authorization
	authInfo, err := component.GetBKAPIAuthorization("")
	if err != nil {
		return false, err
	}
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bkapi-Authorization", authInfo).
		SetBody(map[string]interface{}{"index_set_id": indexSetID}).
		Post(url)

	if err != nil {
		return false, err
	}

	if !resp.IsSuccess() {
		return false, errors.Errorf("http code %d != 200, body: %s", resp.StatusCode(), resp.Body())
	}

	result := &QueryLogResp{}
	err = json.Unmarshal(resp.Body(), result)
	if err != nil {
		return false, errors.Errorf("unmarshal resp body error, %s", err.Error())
	}

	if !result.IsSuccess() {
		return false, errors.Errorf("has log esquery_search error, code: %d, message: %s, request_id: %s",
			result.GetCode(), result.Message, result.RequestID)
	}
	return result.Data.Hits.Total > 0, nil
}

// GetStorageClusters get storage clusters
func GetStorageClusters(ctx context.Context, spaceUID string) ([]GetStorageClustersRespData, error) {
	url := fmt.Sprintf("%s/databus_storage/cluster_groups", config.G.BKLog.APIServer)
	// generate bk api auth header, X-Bkapi-Authorization
	authInfo, err := component.GetBKAPIAuthorization("")
	if err != nil {
		return nil, err
	}
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bkapi-Authorization", authInfo).
		SetQueryParam("space_uid", spaceUID).
		Get(url)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, errors.Errorf("http code %d != 200, body: %s", resp.StatusCode(), resp.Body())
	}

	result := &GetStorageClustersResp{}
	err = json.Unmarshal(resp.Body(), result)
	if err != nil {
		return nil, errors.Errorf("unmarshal resp body error, %s", err.Error())
	}

	if !result.IsSuccess() {
		return nil, errors.Errorf("has log esquery_search error, code: %d, message: %s, request_id: %s",
			result.GetCode(), result.Message, result.RequestID)
	}
	return result.Data, nil
}

// SwitchStorage switch storage
func SwitchStorage(ctx context.Context, spaceUID, bcsClusterID string, storageClusterID int, username string) error {
	url := fmt.Sprintf("%s/switch_bcs_collector_storage", config.G.BKLog.APIServer)
	// generate bk api auth header, X-Bkapi-Authorization
	authInfo, err := component.GetBKAPIAuthorization(username)
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"space_uid":          spaceUID,
		"bcs_cluster_id":     bcsClusterID,
		"storage_cluster_id": storageClusterID,
	}
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bkapi-Authorization", authInfo).
		SetBody(body).
		Post(url)

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
		return errors.Errorf("switch storage error, code: %d, message: %s, request_id: %s",
			result.GetCode(), result.Message, result.RequestID)
	}
	return nil
}

// GetBcsCollectorStorage get bcs collector storage
func GetBcsCollectorStorage(ctx context.Context, spaceUID, clusterID string) (int, error) {
	url := fmt.Sprintf("%s/get_bcs_collector_storage", config.G.BKLog.APIServer)
	// generate bk api auth header, X-Bkapi-Authorization
	authInfo, err := component.GetBKAPIAuthorization("")
	if err != nil {
		return 0, err
	}
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bkapi-Authorization", authInfo).
		SetQueryParam("space_uid", spaceUID).
		SetQueryParam("bcs_cluster_id", clusterID).
		Get(url)

	if err != nil {
		return 0, err
	}

	if !resp.IsSuccess() {
		return 0, errors.Errorf("http code %d != 200, body: %s", resp.StatusCode(), resp.Body())
	}

	result := &GetBcsCollectorStorageResp{}
	err = json.Unmarshal(resp.Body(), result)
	if err != nil {
		return 0, errors.Errorf("unmarshal resp body error, %s", err.Error())
	}

	if !result.IsSuccess() {
		return 0, errors.Errorf("get_bcs_collector_storage error, code: %d, message: %s, request_id: %s",
			result.GetCode(), result.Message, result.RequestID)
	}
	data, err := result.Data.Int64()
	if err != nil {
		return 0, nil
	}
	return int(data), nil
}
