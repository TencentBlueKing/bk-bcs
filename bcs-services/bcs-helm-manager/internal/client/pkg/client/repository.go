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
	"fmt"
	"net/url"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/client/pkg"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

const (
	urlRepository     = "/helmmanager/v1/repository/%s/%s"
	urlRepositoryList = "/helmmanager/v1/repository/%s"
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

	req.Operator = common.GetStringP(rp.conf.Operator)
	name := req.GetName()
	if name == "" {
		return fmt.Errorf("repository name can not be empty")
	}
	projectID := req.GetProjectID()
	if projectID == "" {
		return fmt.Errorf("repository projectID can not be empty")
	}

	var data []byte
	_ = codec.EncJson(req, &data)

	resp, err := rp.post(ctx, urlPrefix+fmt.Sprintf(urlRepository, projectID, name), nil, data)
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

	req.Operator = common.GetStringP(rp.conf.Operator)
	name := req.GetName()
	if name == "" {
		return fmt.Errorf("repository name can not be empty")
	}
	projectID := req.GetProjectID()
	if projectID == "" {
		return fmt.Errorf("repository projectID can not be empty")
	}

	var data []byte
	_ = codec.EncJson(req, &data)

	resp, err := rp.put(ctx, urlPrefix+fmt.Sprintf(urlRepository, projectID, name), nil, data)
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

	req.Operator = common.GetStringP(rp.conf.Operator)
	name := req.GetName()
	if name == "" {
		return fmt.Errorf("repository name can not be empty")
	}
	projectID := req.GetProjectID()
	if projectID == "" {
		return fmt.Errorf("repository projectID can not be empty")
	}

	var data []byte
	_ = codec.EncJson(req, &data)

	resp, err := rp.delete(ctx, urlPrefix+fmt.Sprintf(urlRepository, projectID, name), nil, data)
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
	*helmmanager.RepositoryListData, error) {
	if req == nil {
		return nil, fmt.Errorf("list repository request is empty")
	}

	projectID := req.GetProjectID()
	if projectID == "" {
		return nil, fmt.Errorf("repository project can not be empty")
	}

	resp, err := rp.get(
		ctx,
		urlPrefix+fmt.Sprintf(urlRepositoryList, projectID)+"?"+rp.listRepositoryQuery(req).Encode(),
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

func (rp *repository) listRepositoryQuery(req *helmmanager.ListRepositoryReq) url.Values {
	query := url.Values{}
	if req.Page != nil {
		query.Set("page", strconv.FormatInt(int64(req.GetPage()), 10))
	}
	if req.Size != nil {
		query.Set("size", strconv.FormatInt(int64(req.GetSize()), 10))
	}
	if req.ProjectID != nil {
		query.Set("projectID", req.GetProjectID())
	}
	if req.Name != nil {
		query.Set("name", req.GetName())
	}
	if req.Type != nil {
		query.Set("type", req.GetType())
	}
	query.Set("sort", "createTime")
	return query
}
