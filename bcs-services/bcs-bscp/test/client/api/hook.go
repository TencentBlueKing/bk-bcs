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

	pbcs "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
)

// Hook related interface.
type Hook struct {
	client rest.ClientInterface
}

// NewHookClient get a new hook client
func NewHookClient(client rest.ClientInterface) *Hook {
	return &Hook{
		client: client,
	}
}

// Create function to create hook.
func (c *Hook) Create(ctx context.Context, header http.Header, req *pbcs.CreateHookReq) (
	*pbcs.CreateHookResp, error) {

	resp := c.client.Post().
		WithContext(ctx).
		SubResourcef("/config/apps/%d/hooks", req.AppId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.CreateHookResp `json:"data"`
		Error *rest.ErrorPayload   `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}

// Update function to update hook.
func (c *Hook) Update(ctx context.Context, header http.Header, req *pbcs.UpdateHookReq) (
	*pbcs.UpdateHookResp, error) {

	resp := c.client.Put().
		WithContext(ctx).
		SubResourcef("/config/apps/%d/hooks/%d",
			req.AppId, req.HookId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.UpdateHookResp `json:"data"`
		Error *rest.ErrorPayload   `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}

// Delete function to delete hook.
func (c *Hook) Delete(ctx context.Context, header http.Header, req *pbcs.DeleteHookReq) (
	*pbcs.DeleteHookResp, error) {

	resp := c.client.Delete().
		WithContext(ctx).
		SubResourcef("/config/apps/%d/hooks/%d",
			req.AppId, req.HookId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.DeleteHookResp `json:"data"`
		Error *rest.ErrorPayload   `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}

// List to list hook.
func (c *Hook) List(ctx context.Context, header http.Header,
	req *pbcs.ListHooksReq) (*pbcs.ListHooksResp, error) {

	resp := c.client.Post().
		WithContext(ctx).
		SubResourcef("/config/apps/%d/hooks/list", req.AppId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.ListHooksResp `json:"data"`
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
