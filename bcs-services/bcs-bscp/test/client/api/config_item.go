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

// ConfigItem related interface.
type ConfigItem struct {
	client rest.ClientInterface
}

// NewConfigItemClient get a new config item client
func NewConfigItemClient(client rest.ClientInterface) *ConfigItem {
	return &ConfigItem{
		client: client,
	}
}

// Create function to create config item.
func (c *ConfigItem) Create(ctx context.Context, header http.Header, req *pbcs.CreateConfigItemReq) (
	*pbcs.CreateConfigItemResp, error) {

	resp := c.client.Post().
		WithContext(ctx).
		SubResourcef("/config/create/config_item/config_item/app_id/%d/biz_id/%d", req.AppId, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.CreateConfigItemResp `json:"data"`
		Error *rest.ErrorPayload         `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}

// Update function to update config item.
func (c *ConfigItem) Update(ctx context.Context, header http.Header, req *pbcs.UpdateConfigItemReq) (
	*pbcs.UpdateConfigItemResp, error) {

	resp := c.client.Put().
		WithContext(ctx).
		SubResourcef("/config/update/config_item/config_item/config_item_id/%d/app_id/%d/biz_id/%d",
			req.Id, req.AppId, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.UpdateConfigItemResp `json:"data"`
		Error *rest.ErrorPayload         `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}

// Delete function to delete config item.
func (c *ConfigItem) Delete(ctx context.Context, header http.Header, req *pbcs.DeleteConfigItemReq) (
	*pbcs.DeleteConfigItemResp, error) {

	resp := c.client.Delete().
		WithContext(ctx).
		SubResourcef("/config/delete/config_item/config_item/config_item_id/%d/app_id/%d/biz_id/%d",
			req.Id, req.AppId, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.DeleteConfigItemResp `json:"data"`
		Error *rest.ErrorPayload         `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}

// List to list config item.
func (c *ConfigItem) List(ctx context.Context, header http.Header,
	req *pbcs.ListConfigItemsReq) (*pbcs.ListConfigItemsResp, error) {

	resp := c.client.Post().
		WithContext(ctx).
		SubResourcef("/config/apps/%d/config_items", req.AppId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.ListConfigItemsResp `json:"data"`
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
