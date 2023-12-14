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

// StrategySet client of strategy set
type StrategySet struct {
	client rest.ClientInterface
}

// NewStrategySetClient get a new strategy set client
func NewStrategySetClient(client rest.ClientInterface) *StrategySet {
	return &StrategySet{
		client: client,
	}
}

// Create function to create strategy set.
func (ss *StrategySet) Create(ctx context.Context, header http.Header, req *pbcs.CreateStrategySetReq) (
	*pbcs.CreateStrategySetResp, error) {

	resp := ss.client.Post().
		WithContext(ctx).
		SubResourcef("/config/create/strategy_set/strategy_set/app_id/%d/biz_id/%d", req.AppId, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.CreateStrategySetResp `json:"data"`
		Error *rest.ErrorPayload          `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}

// Update function to update strategy set.
func (ss *StrategySet) Update(ctx context.Context, header http.Header, req *pbcs.UpdateStrategySetReq) (
	*pbcs.UpdateStrategySetResp, error) {

	resp := ss.client.Put().
		WithContext(ctx).
		SubResourcef("/config/update/strategy_set/strategy_set/strategy_set_id/%d/app_id/%d/biz_id/%d",
			req.Id, req.AppId, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.UpdateStrategySetResp `json:"data"`
		Error *rest.ErrorPayload          `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}

// Delete function to delete strategy set.
func (ss *StrategySet) Delete(ctx context.Context, header http.Header, req *pbcs.DeleteStrategySetReq) (
	*pbcs.DeleteStrategySetResp, error) {

	resp := ss.client.Delete().
		WithContext(ctx).
		SubResourcef("/config/delete/strategy_set/strategy_set/strategy_set_id/%d/app_id/%d/biz_id/%d",
			req.Id, req.AppId, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.DeleteStrategySetResp `json:"data"`
		Error *rest.ErrorPayload          `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}

// List to list strategy set.
func (ss *StrategySet) List(ctx context.Context, header http.Header, req *pbcs.ListStrategySetsReq) (
	*pbcs.ListStrategySetsResp, error) {

	resp := ss.client.Post().
		WithContext(ctx).
		SubResourcef("/config/list/strategy_set/strategy_set/app_id/%d/biz_id/%d", req.AppId, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.ListStrategySetsResp `json:"data"`
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
