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

package api

import (
	"context"
	"net/http"
	"reflect"

	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
)

// App application client
type App struct {
	client rest.ClientInterface
}

// NewAppClient get a new app client
func NewAppClient(client rest.ClientInterface) *App {
	return &App{
		client: client,
	}
}

// Create function to create application.
func (a *App) Create(ctx context.Context, header http.Header, req *pbcs.CreateAppReq) (*pbcs.CreateAppResp, error) {
	resp := a.client.Post().
		WithContext(ctx).
		SubResourcef("/config/create/app/app/biz_id/%d", req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.CreateAppResp `json:"data"`
		Error *rest.ErrorPayload  `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}

// Update function to update application.
func (a *App) Update(ctx context.Context, header http.Header, req *pbcs.UpdateAppReq) (*pbcs.UpdateAppResp, error) {
	resp := a.client.Put().
		WithContext(ctx).
		SubResourcef("/config/update/app/app/app_id/%d/biz_id/%d", req.Id, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.UpdateAppResp `json:"data"`
		Error *rest.ErrorPayload  `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}

// Delete function to delete application.
func (a *App) Delete(ctx context.Context, header http.Header, req *pbcs.DeleteAppReq) (*pbcs.DeleteAppResp, error) {
	resp := a.client.Delete().
		WithContext(ctx).
		SubResourcef("/config/delete/app/app/app_id/%d/biz_id/%d", req.Id, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.DeleteAppResp `json:"data"`
		Error *rest.ErrorPayload  `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}

// List to list application.
func (a *App) List(ctx context.Context, header http.Header, req *pbcs.ListAppsReq) (*pbcs.ListAppsResp, error) {
	resp := a.client.Post().
		WithContext(ctx).
		SubResourcef("/config/list/app/app/biz_id/%d", req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.ListAppsResp `json:"data"`
		Error *rest.ErrorPayload `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	return pbResp.Data, pbResp.Error
}
