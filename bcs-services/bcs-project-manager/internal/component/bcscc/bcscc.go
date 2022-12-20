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

// NOTE: 在项目和命名空间完全切换的空窗期，需要向 BCS CC 写入项目数据，防止出现数据不一致情况；待稳定后，删除下面功能

// Package bcscc xxx
package bcscc

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
)

var (
	createProjectPath   = "/projects/"
	updateProjectPath   = "/projects/%s/"
	createNamespacePath = "/projects/%s/clusters/%s/namespaces/"
	listNamespacesPath  = "/projects/%s/clusters/%s/namespaces/"
	deleteNamespacePath = "/projects/%s/clusters/%s/namespaces/%d"
	timeout             = 10
)

type commonResp struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

type listNamespacesResp struct {
	Code      int               `json:"code"`
	Message   string            `json:"message"`
	RequestID string            `json:"request_id"`
	Data      listNamespaceData `json:"data"`
}

type listNamespaceData struct {
	Count   int64           `json:"count"`
	Results []NamespaceData `json:"results"`
}

// NamespaceData paas-cc namespace entity
type NamespaceData struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ClusterID   string `json:"cluster_id"`
	ProjectID   string `json:"project_id"`
	Creator     string `json:"creator"`
	Description string `json:"description"`
}

// CreateProject request bcs cc api, create a project record
func CreateProject(p *pm.Project) error {
	bcsCCConf := config.GlobalConf.BCSCC
	if !bcsCCConf.Enable {
		return nil
	}
	reqURL := fmt.Sprintf("%s%s", bcsCCConf.Host, createProjectPath)
	data := constructProjectData(p)
	data["app_code"] = config.GlobalConf.App.Code
	data["app_secret"] = config.GlobalConf.App.Secret
	req := gorequest.SuperAgent{
		Url:    reqURL,
		Method: "POST",
		Data:   data,
	}
	// 获取返回
	return requestCommonAndParse(req)
}

// UpdateProject request bcs cc api, update a project record
func UpdateProject(p *pm.Project) error {
	bcsCCConf := config.GlobalConf.BCSCC
	if !bcsCCConf.Enable {
		return nil
	}
	realPath := fmt.Sprintf(updateProjectPath, p.ProjectID)
	reqURL := fmt.Sprintf("%s%s", bcsCCConf.Host, realPath)
	data := constructProjectData(p)
	data["app_code"] = config.GlobalConf.App.Code
	data["app_secret"] = config.GlobalConf.App.Secret
	req := gorequest.SuperAgent{
		Url:    reqURL,
		Method: "PUT",
		Data:   data,
	}
	return requestCommonAndParse(req)
}

// CreateNamespace request bcs cc api, create a namespace record
func CreateNamespace(projectCode, clusterID, name, creator string) error {
	bcsCCConf := config.GlobalConf.BCSCC
	if !bcsCCConf.Enable {
		return nil
	}
	model := store.GetModel()
	p, err := model.GetProject(context.Background(), projectCode)
	if err != nil {
		logging.Error("get project by code %s failed, err: %s", projectCode, err.Error())
		return err
	}
	realPath := fmt.Sprintf(createNamespacePath, p.ProjectID, clusterID)
	logging.Info("request url: %s, creator: %s", realPath, creator)
	reqURL := fmt.Sprintf("%s%s", bcsCCConf.Host, realPath)
	data := map[string]interface{}{
		"name":             name,
		"creator":          creator,
		"env_type":         "prod",
		"has_image_secret": false,
	}
	data["app_code"] = config.GlobalConf.App.Code
	data["app_secret"] = config.GlobalConf.App.Secret
	req := gorequest.SuperAgent{
		Url:    reqURL,
		Method: "POST",
		Data:   data,
	}
	logging.Info("req data:%v", req)
	return requestCommonAndParse(req)
}

// ListNamespaces request bcs cc api, list namespace records by projectID and clusterID
func ListNamespaces(projectCode, clusterID string) (*listNamespaceData, error) {
	bcsCCConf := config.GlobalConf.BCSCC
	model := store.GetModel()
	p, err := model.GetProject(context.Background(), projectCode)
	if err != nil {
		logging.Error("get project by code %s failed, err: %s", projectCode, err.Error())
		return nil, err
	}
	realPath := fmt.Sprintf(listNamespacesPath, p.ProjectID, clusterID)
	reqURL := fmt.Sprintf("%s%s", bcsCCConf.Host, realPath)
	reqURL = reqURL + fmt.Sprintf("?app_code=%s", config.GlobalConf.App.Code)
	reqURL = reqURL + fmt.Sprintf("&app_secret=%s", config.GlobalConf.App.Secret)
	reqURL = reqURL + "&desire_all_data=1"
	req := gorequest.SuperAgent{
		Url:    reqURL,
		Method: "GET",
	}
	return requestListNamespacesAndParse(req)
}

// DeleteNamespace request bcs cc api, delete namespace record
func DeleteNamespace(projectCode, clusterID, name string) error {
	bcsCCConf := config.GlobalConf.BCSCC
	if !bcsCCConf.Enable {
		return nil
	}
	model := store.GetModel()
	p, err := model.GetProject(context.Background(), projectCode)
	if err != nil {
		logging.Error("get project by code %s failed, err: %s", projectCode, err.Error())
		return err
	}
	// get id from paascc
	nsList, err := ListNamespaces(projectCode, clusterID)
	if err != nil {
		return err
	}
	var id int
	for _, namespace := range nsList.Results {
		if namespace.Name == name {
			id = namespace.ID
			break
		}
	}
	if id == 0 {
		logging.Error("namespace %s/%s/%s not exists in paas-cc", projectCode, clusterID, name)
		return fmt.Errorf("namespace %s/%s/%s not exists in paas-cc", projectCode, clusterID, name)
	}
	realPath := fmt.Sprintf(deleteNamespacePath, p.ProjectID, clusterID, id)
	reqURL := fmt.Sprintf("%s%s", bcsCCConf.Host, realPath)
	reqURL = reqURL + fmt.Sprintf("?app_code=%s", config.GlobalConf.App.Code)
	reqURL = reqURL + fmt.Sprintf("&app_secret=%s", config.GlobalConf.App.Secret)
	req := gorequest.SuperAgent{
		Url:    reqURL,
		Method: "DELETE",
	}
	return requestCommonAndParse(req)
}

// constructProjectData 组装数据
func constructProjectData(p *pm.Project) map[string]interface{} {
	// biz id to int
	bizIDInt, _ := strconv.Atoi(p.BusinessID)
	// default is 0
	bcsCCKind := 0
	// 1: k8s, 2: mesos
	if p.Kind == "k8s" {
		bcsCCKind = 1
	} else if p.Kind == "mesos" {
		bcsCCKind = 2
	}
	// bg id
	bgID, _ := strconv.Atoi(p.BGID)
	// dept id
	deptID, _ := strconv.Atoi(p.DeptID)
	// center id
	centerID, _ := strconv.Atoi(p.CenterID)
	return map[string]interface{}{
		"project_id":   p.ProjectID,
		"project_name": p.Name,
		"english_name": p.ProjectCode,
		"project_type": p.ProjectType,
		"use_bk":       p.UseBKRes,
		"bg_id":        bgID,
		"bg_name":      p.BGName,
		"dept_id":      deptID,
		"dept_name":    p.DeptName,
		"center_id":    centerID,
		"center_name":  p.CenterName,
		"cc_app_id":    bizIDInt,
		"creator":      p.Creator,
		"updator":      p.Updater,
		"description":  p.Description,
		"kind":         bcsCCKind,
		"deploy_type":  []uint32{p.DeployType},
		"is_secrecy":   p.IsSecret,
		"is_offlined":  p.IsOffline,
	}
}

func requestCommonAndParse(req gorequest.SuperAgent) error {
	// 获取返回数据
	headers := map[string]string{"Content-Type": "application/json"}
	body, err := component.Request(req, timeout, "", headers)
	if err != nil {
		logging.Error("request paas-cc error, data: %v, err: %v", req.Data, err)
		return errorx.NewRequestBCSCCErr(err)
	}
	// 解析返回
	var resp commonResp
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		logging.Error("parse resp error, body: %v", body)
		return err
	}

	if resp.Code != 0 {
		logging.Error("request paas-cc api error, code: %d, message: %s", resp.Code, resp.Message)
		return fmt.Errorf(resp.Message)
	}
	return nil
}

func requestListNamespacesAndParse(req gorequest.SuperAgent) (*listNamespaceData, error) {
	// 获取返回数据
	headers := map[string]string{"Content-Type": "application/json"}
	body, err := component.Request(req, timeout, "", headers)
	if err != nil {
		logging.Error("request paas-cc error, data: %v, err: %v", req.Data, err)
		return nil, errorx.NewRequestBCSCCErr(err)
	}
	// 解析返回
	var resp listNamespacesResp
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		logging.Error("parse resp error, body: %v", body)
		return nil, err
	}

	if resp.Code != 0 {
		logging.Error("request paas-cc api error, code: %d, message: %s", resp.Code, resp.Message)
		return nil, errorx.NewRequestBCSCCErr(err.Error())
	}
	return &resp.Data, nil
}
