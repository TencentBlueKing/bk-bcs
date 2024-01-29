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

// Instance client of instance
type Instance struct {
	client rest.ClientInterface
}

// NewInstanceClient get a new instance client
func NewInstanceClient(client rest.ClientInterface) *Instance {
	return &Instance{
		client: client,
	}
}

// Publish to publish instance.
func (ins *Instance) Publish(ctx context.Context, header http.Header, req *pbcs.PublishInstanceReq) (
	*pbcs.PublishInstanceResp, error) {
	resp := ins.client.Post().
		WithContext(ctx).
		SubResourcef("/config/create/instance/publish/app_id/%d/biz_id/%d", req.AppId, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.PublishInstanceResp `json:"data"`
		Error *rest.ErrorPayload        `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}

// Delete to delete publish instance.
func (ins *Instance) Delete(ctx context.Context, header http.Header, req *pbcs.DeletePublishedInstanceReq) (
	*pbcs.DeletePublishedInstanceResp, error) {

	resp := ins.client.Delete().
		WithContext(ctx).
		SubResourcef("/config/delete/instance/publish/id/%d/app_id/%d/biz_id/%d", req.Id, req.AppId, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.DeletePublishedInstanceResp `json:"data"`
		Error *rest.ErrorPayload                `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}

// List to list publish instance.
func (ins *Instance) List(ctx context.Context, header http.Header, req *pbcs.ListPublishedInstanceReq) (
	*pbcs.ListPublishedInstanceResp, error) {

	resp := ins.client.Post().
		WithContext(ctx).
		SubResourcef("/config/list/instance/publish/biz_id/%d", req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.ListPublishedInstanceResp `json:"data"`
		Error *rest.ErrorPayload              `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}
