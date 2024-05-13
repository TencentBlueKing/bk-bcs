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

// Package bkmonitor xxx
package bkmonitor

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
)

var (
	createSpacePath = "/metadata_create_space/"
	listSpacesPath  = "/metadata_list_spaces/"
	timeout         = 10
)

// CreateSpaceResp create bkmonitor space response
type CreateSpaceResp struct {
	Code    int    `json:"code"`
	Result  bool   `json:"result"`
	Message string `json:"message"`
	Data    *Space `json:"data"`
}

// ListSpacesResp list bkmonitor spaces response
type ListSpacesResp struct {
	Code    int             `json:"code"`
	Result  bool            `json:"result"`
	Message string          `json:"message"`
	Data    *ListSpacesData `json:"data"`
}

// ListSpacesData list bkmonitor spaces data
type ListSpacesData struct {
	Count int      `json:"count"`
	List  []*Space `json:"list"`
}

// Space ITSM bkmonitor space
type Space struct {
	ID          int    `json:"id"`
	SpaceTypeID string `json:"space_type_id"`
	SpaceID     string `json:"space_id"`
	SpaceCode   string `json:"space_code"`
	SpaceName   string `json:"space_name"`
	DisplayName string `json:"display_name"`
	Status      string `json:"status"`
	TimeZone    string `json:"time_zone"`
	Language    string `json:"language"`
	IsBcsValid  bool   `json:"is_bcs_valid"`
	SpaceUID    string `json:"space_uid"`
}

// CreateSpace create bkmonitor space for bcs project
func CreateSpace(project *project.Project) error {
	bkmConf := config.GlobalConf.Bkmonitor
	// 使用网关访问
	reqURL := fmt.Sprintf("%s%s", bkmConf.GatewayHost, createSpacePath)
	req := gorequest.New().Post(reqURL)
	req.Data = map[string]interface{}{
		"space_name":    project.Name,
		"space_type_id": "bkci",
		"is_bcs_valid":  true,
		"space_id":      project.ProjectCode,
		"space_code":    project.ProjectID,
		"creator":       project.Creator,
	}
	// 请求API
	proxy := ""
	body, err := component.Request(*req, timeout, proxy, component.GetAuthHeader())
	if err != nil {
		logging.Error("request create bkmonitor space for project %s failed, %s", project.ProjectID, err.Error())
		return errorx.NewRequestBkMonitorErr(err.Error())
	}
	// 解析返回的body
	resp := &CreateSpaceResp{}
	if err := json.Unmarshal([]byte(body), resp); err != nil {
		logging.Error("parse bkmonitor body error, body: %v", body)
		return err
	}
	if resp.Code != 200 {
		logging.Error("request create bkmonitor space for project %s failed, msg: %s", project.ProjectID, resp.Message)
		return errors.New(resp.Message)
	}
	return nil
}

// ListSpaces list bkmonitor spaces for bcs
func ListSpaces() ([]*Space, error) {
	bkmConf := config.GlobalConf.Bkmonitor
	// 使用网关访问
	reqURL := fmt.Sprintf("%s%s", bkmConf.GatewayHost, listSpacesPath)
	spaces := make([]*Space, 0)
	var page, pageSize = 1, 1000
	for {
		req := gorequest.New().Get(reqURL)
		req.QueryData.Set("space_type_id", "bkci")
		req.QueryData.Set("page", strconv.Itoa(page))
		req.QueryData.Set("page_size", strconv.Itoa(pageSize))
		// 请求API
		proxy := ""
		body, err := component.Request(*req, timeout, proxy, component.GetAuthHeader())
		if err != nil {
			logging.Error("request list bkmonitor bcs spaces failed, %s", err.Error())
			return nil, errorx.NewRequestBkMonitorErr(err.Error())
		}
		// 解析返回的body
		resp := &ListSpacesResp{}
		if err := json.Unmarshal([]byte(body), resp); err != nil {
			logging.Error("parse bkmonitor body error, body: %v", body)
			return nil, err
		}
		if resp.Code != 200 {
			logging.Error("request list bkmonitor spaces failed, msg: %s", resp.Message)
			return nil, errors.New(resp.Message)
		}
		for _, space := range resp.Data.List {
			if space.IsBcsValid {
				spaces = append(spaces, space)
			}
		}
		if resp.Data.Count <= page*pageSize {
			break
		}
		page++
	}

	return spaces, nil
}
