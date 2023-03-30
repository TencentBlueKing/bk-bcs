/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"context"
	"net/http"
	"reflect"

	pbcs "bscp.io/pkg/protocol/config-server"
	"bscp.io/pkg/rest"
)

// Release client of release
type Release struct {
	client rest.ClientInterface
}

// NewReleaseClient get a new release client
func NewReleaseClient(client rest.ClientInterface) *Release {
	return &Release{
		client: client,
	}
}

// Create function to create release.
func (r *Release) Create(ctx context.Context, header http.Header, req *pbcs.CreateReleaseReq) (
	*pbcs.CreateReleaseResp, error) {

	resp := r.client.Post().
		WithContext(ctx).
		SubResourcef("/config/create/release/release/app_id/%d/biz_id/%d", req.AppId, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.CreateReleaseResp `json:"data"`
		Error *rest.ErrorPayload      `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}

// List to list release.
func (r *Release) List(ctx context.Context, header http.Header, req *pbcs.ListReleasesReq) (
	*pbcs.ListReleasesResp, error) {
	resp := r.client.Post().
		WithContext(ctx).
		SubResourcef("/config/list/release/release/app_id/%d/biz_id/%d", req.AppId, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.ListReleasesResp `json:"data"`
		Error *rest.ErrorPayload     `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}
