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

// Strategy client of strategy.
type Strategy struct {
	client rest.ClientInterface
}

// NewStrategyClient get a new strategy client
func NewStrategyClient(client rest.ClientInterface) *Strategy {
	return &Strategy{
		client: client,
	}
}

// Create function to create strategy.
func (s *Strategy) Create(ctx context.Context, header http.Header, req *pbcs.CreateStrategyReq) (
	*pbcs.CreateStrategyResp, error) {

	resp := s.client.Post().
		WithContext(ctx).
		SubResourcef("/config/create/strategy/strategy/strategy_set_id/%d/app_id/%d/biz_id/%d",
			req.StrategySetId, req.AppId, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.CreateStrategyResp `json:"data"`
		Error *rest.ErrorPayload       `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}

// Update function to update strategy.
func (s *Strategy) Update(ctx context.Context, header http.Header, req *pbcs.UpdateStrategyReq) (
	*pbcs.UpdateStrategyResp, error) {

	resp := s.client.Put().
		WithContext(ctx).
		SubResourcef("/config/update/strategy/strategy/strategy_id/%d/app_id/%d/biz_id/%d",
			req.Id, req.AppId, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.UpdateStrategyResp `json:"data"`
		Error *rest.ErrorPayload       `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}

// Delete function to delete strategy.
func (s *Strategy) Delete(ctx context.Context, header http.Header, req *pbcs.DeleteStrategyReq) (
	*pbcs.DeleteStrategyResp, error) {

	resp := s.client.Delete().
		WithContext(ctx).
		SubResourcef("/config/delete/strategy/strategy/strategy_id/%d/app_id/%d/biz_id/%d",
			req.Id, req.AppId, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.DeleteStrategyResp `json:"data"`
		Error *rest.ErrorPayload       `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}

// List to list strategy.
func (s *Strategy) List(ctx context.Context, header http.Header, req *pbcs.ListStrategiesReq) (
	*pbcs.ListStrategiesResp, error) {

	resp := s.client.Post().
		WithContext(ctx).
		SubResourcef("/config/list/strategy/strategy/app_id/%d/biz_id/%d", req.AppId, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.ListStrategiesResp `json:"data"`
		Error *rest.ErrorPayload       `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}
