/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package project

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/component"
)

// ProjectClient ...
type ProjectClient struct {
	Host  string
	Token string
}

// Client ...
var Client *ProjectClient

// NewClient create project service client
func NewClient(c ProjectClient) *ProjectClient {
	client := &ProjectClient{Host: c.Host, Token: c.Token}
	Client = client
	return client
}

// ProjectData project service detail
type ProjectData struct {
	ProjectID   string `json:"projectID"`
	Kind        string `json:"kind"`
	BusinessID  string `json:"businessID"`
	ProjectCode string `json:"projectCode"`
	Name        string `json:"name"`
}

type ProjectResp struct {
	Code    int         `json:"code"`
	Data    ProjectData `json:"data"`
	Message string      `json:"message"`
}

var (
	getProjectPath = "/bcsapi/v4/bcsproject/v1/projects/%s"
	// 默认超时时间设置为20s
	defaultTimeout = 20
)

// GetProjectID get project id from project code
// TODO: projectID不会变动，可以添加下缓存
func GetProjectIDByCode(username string, projectCode string) (string, error) {
	p, err := Client.GetProjectDetail(username, projectCode)
	if err != nil {
		return "", err
	}
	return p.ProjectID, nil
}

// GetProjectDetail get project detail from project service
func (p *ProjectClient) GetProjectDetail(username string, projectCode string) (*ProjectData, error) {
	path := fmt.Sprintf(getProjectPath, projectCode)
	url := fmt.Sprintf("%s%s", p.Host, path)
	authorization := fmt.Sprintf("Bearer %s", p.Token)
	headers := map[string]string{"Content-Type": "application/json", "Authorization": authorization, "X-Project-Username": username}
	// 组装请求参数
	req := gorequest.SuperAgent{
		Url:    url,
		Method: "GET",
	}
	// 请求接口
	body, err := component.Request(req, defaultTimeout, "", headers)
	if err != nil {
		blog.Errorf("request project service error, project code: %s, err %s", projectCode, err.Error())
		return nil, common.ErrHelmManagerRequestComponentFailed.GenError()
	}
	var resp ProjectResp
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		blog.Errorf("parse project detail error, body: %v", body)
		return nil, err
	}
	if resp.Code != component.SuccessCode {
		blog.Errorf("parse project detail error, body: %v", body)
		return nil, errors.New(resp.Message)
	}
	return &resp.Data, nil
}
