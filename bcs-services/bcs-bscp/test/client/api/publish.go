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

// Publish client of publish
type Publish struct {
	client rest.ClientInterface
}

// NewPublishClient get a new publish client
func NewPublishClient(client rest.ClientInterface) *Publish {
	return &Publish{
		client: client,
	}
}

// PublishWithStrategy function to publish with strategy.
func (p *Publish) PublishWithStrategy(ctx context.Context, header http.Header, req *pbcs.PublishReq) (
	*pbcs.PublishResp, error) {
	resp := p.client.Post().
		WithContext(ctx).
		SubResourcef("/config/update/strategy/publish/publish/strategy_id/%d/app_id/%d/biz_id/%d",
			req.ReleaseId, req.AppId, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.PublishResp  `json:"data"`
		Error *rest.ErrorPayload `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}

// FinishPublishWithStrategy function to finish publish with strategy.
func (p *Publish) FinishPublishWithStrategy(ctx context.Context, header http.Header,
	req *pbcs.FinishPublishReq) (*pbcs.FinishPublishResp, error) {

	resp := p.client.Put().
		WithContext(ctx).
		SubResourcef("/config/update/strategy/publish/finish/strategy_id/%d/app_id/%d/biz_id/%d",
			req.Id, req.AppId, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.FinishPublishResp `json:"data"`
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

// ListStrategyPublishHistory function to list strategy publish history.
func (p *Publish) ListStrategyPublishHistory(ctx context.Context, header http.Header,
	req *pbcs.ListPubStrategyHistoriesReq) (*pbcs.ListPubStrategyHistoriesResp, error) {

	resp := p.client.Post().
		WithContext(ctx).
		SubResourcef("/config/list/strategy/publish/history/app_id/%d/biz_id/%d", req.AppId, req.BizId).
		WithHeaders(header).
		Body(req).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	pbResp := &struct {
		Data  *pbcs.ListPubStrategyHistoriesResp `json:"data"`
		Error *rest.ErrorPayload                 `json:"error"`
	}{}
	if err := resp.Into(pbResp); err != nil {
		return nil, err
	}
	if !reflect.ValueOf(pbResp.Error).IsNil() {
		return nil, pbResp.Error
	}

	return pbResp.Data, nil
}
