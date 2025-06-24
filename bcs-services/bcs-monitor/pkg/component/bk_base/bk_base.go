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

// Package bkbase bkbase
package bkbase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
)

const (
	// BKBaseBCSNamespace bk base bcs namespace
	BKBaseBCSNamespace = "bk_bcs"
)

// GenDataIDName generate data id name
func GenDataIDName(clusterID string) string {
	return fmt.Sprintf("bkbcs_audit_dataid_%s", strings.ToLower(clusterID))
}

// GenDatabusName generate databus name
func GenDatabusName(clusterID string) string {
	return fmt.Sprintf("bkbcs_audit_databus_%s", strings.ToLower(clusterID))
}

// ApplyDataID apply data id
func ApplyDataID(ctx context.Context, name string, bizID, dataID int, topic string) error {
	url := fmt.Sprintf("%s/v4/apply", config.G.BKBase.APIServer)
	// generate bk api auth header, X-Bkapi-Authorization
	authInfo, err := component.GetBKAPIAuthorization("")
	if err != nil {
		return err
	}
	body := fmt.Sprintf(CreateDataIDBody, BKBaseBCSNamespace, name, name, name, bizID, dataID, BKBaseBCSNamespace,
		config.G.BKBase.AuditChannelName, topic)
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

	if !result.Result {
		return errors.New(result.Message)
	}
	return nil
}

// ApplyDatabus apply databus
func ApplyDatabus(ctx context.Context, name string, dataIDName string) error {
	url := fmt.Sprintf("%s/v4/apply", config.G.BKBase.APIServer)
	// generate bk api auth header, X-Bkapi-Authorization
	authInfo, err := component.GetBKAPIAuthorization("")
	if err != nil {
		return err
	}
	body := fmt.Sprintf(CreateDatabusBody, BKBaseBCSNamespace, name, BKBaseBCSNamespace, dataIDName, BKBaseBCSNamespace,
		config.G.BKBase.AuditChannelBindingName)
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

	if !result.Result {
		return errors.New(result.Message)
	}
	return nil
}
