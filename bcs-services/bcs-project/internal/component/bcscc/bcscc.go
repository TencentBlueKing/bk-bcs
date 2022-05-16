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

// NOTE: 在项目完全切换的空窗期，需要向 BCS CC 写入项目数据，防止出现数据不一致情况；待稳定后，删除下面功能

package bcscc

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/logging"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util/errorx"
)

var (
	createProjectPath = "/projects/"
	updateProjectPath = "/projects/%s/"
	timeout           = 10
)

type projectResp struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// CreateProject request bcs cc api, create a project record
func CreateProject(p *pm.Project) error {
	bcsCCConf := config.GlobalConf.BCSCC
	reqUrl := fmt.Sprintf("%s%s", bcsCCConf.Host, createProjectPath)
	data := constructProjectData(p)
	data["app_code"] = config.GlobalConf.App.Code
	data["app_secret"] = config.GlobalConf.App.Secret
	req := gorequest.SuperAgent{
		Url:    reqUrl,
		Method: "POST",
		Data:   data,
	}
	// 获取返回
	return requestAndParse(req)
}

// UpdateProject request bcs cc api, update a project record
func UpdateProject(p *pm.Project) error {
	bcsCCConf := config.GlobalConf.BCSCC
	realPath := fmt.Sprintf(updateProjectPath, p.ProjectID)
	reqUrl := fmt.Sprintf("%s%s", bcsCCConf.Host, realPath)
	data := constructProjectData(p)
	data["app_code"] = config.GlobalConf.App.Code
	data["app_secret"] = config.GlobalConf.App.Secret
	req := gorequest.SuperAgent{
		Url:    reqUrl,
		Method: "PUT",
		Data:   data,
	}
	return requestAndParse(req)
}

// 组装数据
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

func requestAndParse(req gorequest.SuperAgent) error {
	// 获取返回数据
	headers := map[string]string{"Content-Type": "application/json"}
	body, err := component.Request(req, timeout, "", headers)
	if err != nil {
		logging.Error("request bcs cc error, data: %v, err: %v", req.Data, err)
		return errorx.NewRequestBCSCCErr(err)
	}
	// 解析返回
	var resp projectResp
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		logging.Error("parse resp error, body: %v", body)
		return err
	}

	if resp.Code != 0 {
		logging.Error("request project api error, message: %s", resp.Message)
	}
	return nil
}
