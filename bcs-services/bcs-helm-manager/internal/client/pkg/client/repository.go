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

package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/client/pkg"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

const (
	urlRepository    = "/projects/%s/repos"
	urlGetRepository = "/projects/%s/repos/%s"
)

// Repository return a pkg.RepositoryClient instance
func (c *Client) Repository() pkg.RepositoryClient {
	return &repository{Client: c}
}

type repository struct {
	*Client
}

// Create repository
func (rp *repository) Create(ctx context.Context, req *helmmanager.CreateRepositoryReq) error {
	if req == nil {
		return fmt.Errorf("create repository request is empty")
	}

	name := req.GetName()
	if name == "" {
		return fmt.Errorf("repository name can not be empty")
	}
	projectCode := req.GetProjectCode()
	if projectCode == "" {
		return fmt.Errorf("repository projectCode can not be empty")
	}

	data, _ := json.Marshal(req)

	resp, err := rp.post(ctx, urlPrefix+fmt.Sprintf(urlRepository, projectCode), nil, data)
	if err != nil {
		return err
	}

	var r helmmanager.CreateRepositoryResp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return err
	}

	if r.GetCode() != resultCodeSuccess {
		return fmt.Errorf("create repository get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return nil
}

// Update repository
func (rp *repository) Update(ctx context.Context, req *helmmanager.UpdateRepositoryReq) error {
	if req == nil {
		return fmt.Errorf("update repository request is empty")
	}

	name := req.GetName()
	if name == "" {
		return fmt.Errorf("repository name can not be empty")
	}
	projectCode := req.GetProjectCode()
	if projectCode == "" {
		return fmt.Errorf("repository projectCode can not be empty")
	}

	data, _ := json.Marshal(req)

	resp, err := rp.put(ctx, urlPrefix+fmt.Sprintf(urlRepository, projectCode)+"/"+name, nil, data)
	if err != nil {
		return err
	}

	var r helmmanager.UpdateRepositoryResp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return err
	}

	if r.GetCode() != resultCodeSuccess {
		return fmt.Errorf("update repository get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return nil
}

// Delete repository
func (rp *repository) Delete(ctx context.Context, req *helmmanager.DeleteRepositoryReq) error {
	if req == nil {
		return fmt.Errorf("delete repository request is empty")
	}

	name := req.GetName()
	if name == "" {
		return fmt.Errorf("repository name can not be empty")
	}
	projectCode := req.GetProjectCode()
	if projectCode == "" {
		return fmt.Errorf("repository projectCode can not be empty")
	}

	data, _ := json.Marshal(req)

	resp, err := rp.delete(ctx, urlPrefix+fmt.Sprintf(urlGetRepository, projectCode, name), nil, data)
	if err != nil {
		return err
	}

	var r helmmanager.DeleteRepositoryResp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return err
	}

	if r.GetCode() != resultCodeSuccess {
		return fmt.Errorf("delete repository get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return nil
}

// List repository
func (rp *repository) List(ctx context.Context, req *helmmanager.ListRepositoryReq) (
	[]*helmmanager.Repository, error) {
	if req == nil {
		return nil, fmt.Errorf("list repository request is empty")
	}

	projectCode := req.GetProjectCode()
	if projectCode == "" {
		return nil, fmt.Errorf("repository project can not be empty")
	}

	resp, err := rp.get(
		ctx,
		urlPrefix+fmt.Sprintf(urlRepository, projectCode),
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	var r helmmanager.ListRepositoryResp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return nil, err
	}

	if r.GetCode() != resultCodeSuccess {
		return nil, fmt.Errorf("list repository get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return r.Data, nil
}

// Get repository
func (rp *repository) Get(ctx context.Context, req *helmmanager.GetRepositoryReq) (
	*helmmanager.Repository, error) {
	if req == nil {
		return nil, fmt.Errorf("get repository request is empty")
	}

	projectCode := req.GetProjectCode()
	if projectCode == "" {
		return nil, fmt.Errorf("repository project can not be empty")
	}
	repositoryName := req.GetName()
	if repositoryName == "" {
		return nil, fmt.Errorf("repository name can not be empty")
	}
	resp, err := rp.get(
		ctx,
		urlPrefix+fmt.Sprintf(urlGetRepository, projectCode, repositoryName),
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	var r helmmanager.GetRepositoryResp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return nil, err
	}

	if r.GetCode() != resultCodeSuccess {
		return nil, fmt.Errorf("get repository get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return r.Data, nil
}
