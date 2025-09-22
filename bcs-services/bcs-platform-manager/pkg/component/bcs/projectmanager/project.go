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

// Package projectmanager xxx
package projectmanager

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
)

// GetProject get project from project manager
func GetProject(ctx context.Context, projectIDOrCode string) (*bcsproject.Project, error) {
	cli, close, err := bcsproject.GetClient(config.ServiceDomain)
	if err != nil {
		return nil, err
	}

	defer Close(close)

	p, err := cli.Project.GetProject(ctx, &bcsproject.GetProjectRequest{ProjectIDOrCode: projectIDOrCode})
	if err != nil {
		return nil, fmt.Errorf("GetProject error: %s", err)
	}

	if p.Code != 0 || p.Data == nil {
		return nil, fmt.Errorf("GetProject error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return p.Data, nil
}

// ListAllProject list all project from project manager
func ListAllProject(ctx context.Context) ([]*bcsproject.Project, error) {
	cli, close, err := bcsproject.GetClient(config.ServiceDomain)
	if err != nil {
		return nil, err
	}

	defer Close(close)

	p, err := cli.Project.ListProjects(ctx, &bcsproject.ListProjectsRequest{All: true})
	if err != nil {
		return nil, fmt.Errorf("ListProject error: %s", err)
	}

	if p.Code != 0 || p.Data == nil {
		return nil, fmt.Errorf("ListProject error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return p.Data.GetResults(), nil
}

// ListProject list project from project manager
func ListProject(ctx context.Context, req *bcsproject.ListProjectsRequest) ([]*bcsproject.Project, error) {
	cli, close, err := bcsproject.GetClient(config.ServiceDomain)
	if err != nil {
		return nil, fmt.Errorf("GetClient error: %s", err)
	}

	defer Close(close)

	p, err := cli.Project.ListProjects(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ListProject error: %s", err)
	}

	if p.Code != 0 || p.Data == nil {
		return nil, fmt.Errorf("ListProject error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return p.Data.GetResults(), nil
}

// UpdateProject update project from project manager
func UpdateProject(ctx context.Context, req *bcsproject.UpdateProjectRequest) (bool, error) {
	cli, close, err := bcsproject.GetClient(config.ServiceDomain)
	if err != nil {
		return false, err
	}

	defer Close(close)

	p, err := cli.Project.UpdateProject(ctx, req)
	if err != nil {
		return false, fmt.Errorf("UpdateProject error: %s", err)
	}

	if p.Code != 0 {
		return false, fmt.Errorf("UpdateProject error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return true, nil
}
