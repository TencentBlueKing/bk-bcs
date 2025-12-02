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

// Package templateconfig templateconfig operate
package templateconfig

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/bcs"
)

// DeleteTemplateConfigReq delete template config request
type DeleteTemplateConfigReq struct {
	TemplateConfigID string `json:"cloudID" in:"path=templateConfigID"`
	BusinessID       string `json:"businessID" in:"query=businessID"`
	ProjectID        string `json:"projectID" in:"query=projectID"`
}

// ListTemplateConfigReq list template config request
type ListTemplateConfigReq struct {
	BusinessID string `json:"businessID" in:"query=businessID"`
	ProjectID  string `json:"projectID" in:"query=projectID"`
	ClusterID  string `json:"clusterID" in:"query=clusterID"`
	Provider   string `json:"provider" in:"query=provider"`
	ConfigType string `json:"configType" in:"query=configType"`
}

// CreateTemplateConfig 创建TemplateConfig
// @Summary 创建TemplateConfig
// @Tags    Logs
// @Produce json
// @Success 200 {array} k8sclient.Container
// @Router  /templateConfig [post]
func CreateTemplateConfig(c context.Context, req *bcs.CreateTemplateConfigReq) (*bool, error) {
	result, err := bcs.CreateTemplateConfig(req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteTemplateConfig 删除TemplateConfig
// @Summary 删除TemplateConfig
// @Tags    Logs
// @Produce json
// @Success 200 {array} k8sclient.Container
// @Router  /templateConfig/{templateConfigID} [delete]
func DeleteTemplateConfig(c context.Context, req *DeleteTemplateConfigReq) (*bool, error) {
	result, err := bcs.DeleteTemplateConfig(req.TemplateConfigID, req.BusinessID, req.ProjectID)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// ListTemplateConfig 获取TemplateConfig列表
// @Summary 获取TemplateConfig列表
// @Tags    Logs
// @Produce json
// @Success 200 {array} k8sclient.Container
// @Router  /templateConfig/{templateConfigID} [get]
func ListTemplateConfig(c context.Context, req *ListTemplateConfigReq) (*[]*bcs.TemplateConfigInfo, error) {
	result, err := bcs.ListTemplateConfig(req.BusinessID, req.ProjectID, req.ClusterID, req.Provider, req.ConfigType)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
