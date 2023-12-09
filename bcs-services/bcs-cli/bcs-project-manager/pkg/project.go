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

package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
	"github.com/google/go-querystring/query"
	"github.com/pkg/errors"
)

const (
	listProjectsUrl       = "/bcsproject/v1/projects"
	getProjectUrl         = "/bcsproject/v1/projects/%s"
	updateProjectUrl      = "/bcsproject/v1/projects/%s"
	authorizedProjectsUrl = "/bcsproject/v1/authorized_projects"
)

type (
	// ListProjectsRequest list projects
	ListProjectsRequest struct {
		ProjectIDs  string `url:"projectIDs,omitempty"`
		ProjectCode string `url:"projectCode,omitempty"`
		SearchName  string `url:"searchName,omitempty"`
		Kind        string `url:"kind,omitempty"`
		Offset      int64  `url:"offset,omitempty"`
		Limit       int64  `url:"limit,omitempty"`
		All         bool   `url:"all,omitempty"`
	}

	// UpdateProjectRequest update project
	UpdateProjectRequest struct {
		BusinessID  string `json:"businessID"`
		CreateTime  string `json:"createTime"`
		Creator     string `json:"creator"`
		Description string `json:"description"`
		Kind        string `json:"kind"`
		Managers    string `json:"managers"`
		Name        string `json:"name"`
		ProjectCode string `json:"projectCode"`
		ProjectID   string `json:"projectID"`
		UpdateTime  string `json:"updateTime"`
	}
)

// ListProjects Get a list of projects based on a condition
func (p *ProjectManagerClient) ListProjects(in *ListProjectsRequest) (*bcsproject.ListProjectsResponse, error) {
	v, err := query.Values(in)
	if err != nil {
		return nil, fmt.Errorf("slice and Array values default to encoding as multiple URL values failed: %v", err)
	}
	bs, err := p.do(listProjectsUrl, http.MethodGet, v, nil)
	if err != nil {
		return nil, fmt.Errorf("list projects failed: %v", err)
	}
	resp := new(bcsproject.ListProjectsResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "list projects unmarshal failed with response '%s'", string(bs))
	}
	if resp != nil && resp.Code != 0 {
		return nil, fmt.Errorf("list projects response code not 0 but %d: %s", resp.Code, resp.Message)
	}
	return resp, nil
}

// GetProject Get project information based on project ID or CODE
func (p *ProjectManagerClient) GetProject(projectIDOrCode string) (*bcsproject.ProjectResponse, error) {
	bs, err := p.do(fmt.Sprintf(getProjectUrl, projectIDOrCode), http.MethodGet, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "get project with '%s' failed", projectIDOrCode)
	}
	resp := new(bcsproject.ProjectResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "get project unmarshal failed with response '%s'", string(bs))
	}
	if resp != nil && resp.Code != 0 {
		return nil, fmt.Errorf("get project response code not 0 but %d: %s", resp.Code, resp.Message)
	}
	return resp, nil
}

// UpdateProject edit project
func (p *ProjectManagerClient) UpdateProject(in *UpdateProjectRequest) (*bcsproject.ProjectResponse, error) {
	bs, err := p.do(fmt.Sprintf(updateProjectUrl, in.ProjectID), http.MethodPut, nil, in)
	if err != nil {
		return nil, fmt.Errorf("update project failed: %v", err)
	}
	resp := new(bcsproject.ProjectResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "update project unmarshal failed with response '%s'", string(bs))
	}
	if resp != nil && resp.Code != 0 {
		return nil, fmt.Errorf("update project response code not 0 but %d: %s", resp.Code, resp.Message)
	}
	return resp, nil
}

// ListAuthorizedProjects Query the list of items to which the user has permission
func (p *ProjectManagerClient) ListAuthorizedProjects() (*bcsproject.ListAuthorizedProjResp, error) {

	bs, err := p.do(authorizedProjectsUrl, http.MethodGet, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("list authorized projects failed: %v", err)
	}
	resp := new(bcsproject.ListAuthorizedProjResp)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "list authorized projects unmarshal failed with response '%s'", string(bs))
	}
	if resp != nil && resp.Code != 0 {
		return nil, fmt.Errorf("llist authorized projects response code not 0 but %d: %s", resp.Code, resp.Message)
	}
	return resp, nil
}
