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

// Package itsm xxx
package itsm

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
)

var (
	createCatalogPath = "/v2/itsm/create_service_catalog/"
	getCatalogsPath   = "/v2/itsm/get_service_catalogs/"
)

// CreateCatalogReq 创建服务目录请求
type CreateCatalogReq struct {
	// ProjectKey 项目id, 默认为 "0"
	ProjectKey string `json:"project_key"`
	// ParentID 父目录ID
	ParentID uint32 `json:"parent__id"`
	// Name 目录名称
	Name string `json:"name"`
	// Desc 目录描述
	Desc string `json:"desc"`
}

// CreateCatalogResp 创建服务目录返回
type CreateCatalogResp struct {
	component.CommonResp
	Data CreateCatalogData `json:"data"`
}

// CreateCatalogData 创建服务目录返回数据
type CreateCatalogData struct {
	ID uint32 `json:"id"`
}

// CreateCatalog 创建服务目录，返回目录ID
func CreateCatalog(data CreateCatalogReq) (uint32, error) {
	itsmConf := config.GlobalConf.ITSM
	// 使用网关访问
	reqURL := fmt.Sprintf("%s%s", itsmConf.GatewayHost, createCatalogPath)
	req := gorequest.SuperAgent{
		Url:    reqURL,
		Method: "POST",
		Data: map[string]interface{}{
			"project_key": data.ProjectKey,
			"parent__id":  data.ParentID,
			"name":        data.Name,
			"desc":        data.Desc,
		},
	}
	// 请求API
	body, err := component.Request(req, timeout, "", component.GetAuthHeader())
	if err != nil {
		logging.Error("request create itsm catalog %v failed, error: %s", data.Name, err.Error())
		return 0, errorx.NewRequestITSMErr(err.Error())
	}
	// 解析返回的body
	resp := &CreateCatalogResp{}
	if err := json.Unmarshal([]byte(body), resp); err != nil {
		logging.Error("parse itsm body error, body: %v", body)
		return 0, err
	}
	if resp.Code != 0 {
		logging.Error("request create itsm catalog %v failed, msg: %s", data.Name, resp.Message)
		return 0, errors.New(resp.Message)
	}
	return resp.Data.ID, nil
}

// Catalog ITSM 服务目录
type Catalog struct {
	// ID 目录ID
	ID uint32 `json:"id"`
	// Name 目录名称
	Name string `json:"name"`
	// Desc 目录描述
	Desc string `json:"desc"`
	// Leven 目录层级
	Leven int `json:"leven"`
	// Key 目录key
	Key string `json:"key"`
	// Children 子目录
	Children []Catalog `json:"children"`
}

// ListCatalogsResp 获取服务目录列表返回
type ListCatalogsResp struct {
	component.CommonResp
	Data []Catalog `json:"data"`
}

// ListCatalogs 获取服务目录列表
func ListCatalogs() ([]Catalog, error) {
	itsmConf := config.GlobalConf.ITSM
	// 使用网关访问
	reqURL := fmt.Sprintf("%s%s?project_key=0", itsmConf.GatewayHost, getCatalogsPath)
	req := gorequest.SuperAgent{
		Url:    reqURL,
		Method: "GET",
	}
	// 请求API
	body, err := component.Request(req, timeout, "", component.GetAuthHeader())
	if err != nil {
		logging.Error("request get itsm catalogs failed, error: %s", err.Error())
		return nil, errorx.NewRequestITSMErr(err.Error())
	}
	// 解析返回的body
	resp := &ListCatalogsResp{}
	if err := json.Unmarshal([]byte(body), resp); err != nil {
		logging.Error("parse itsm body error, body: %v", body)
		return nil, err
	}
	if resp.Code != 0 {
		logging.Error("request get itsm catalogs failed, msg: %s", resp.Message)
		return nil, errors.New(resp.Message)
	}
	return resp.Data, nil
}
