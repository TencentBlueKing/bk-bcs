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

// Package cmdb NOTES
package cmdb

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bluele/gcache"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/uuid"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
)

// Client is an esb client to request cmdb.
type Client interface {
	// SearchBusiness 通用查询
	SearchBusiness(ctx context.Context, params *SearchBizParams) (*SearchBizResp, error)
	// ListAllBusiness 读取全部业务列表
	ListAllBusiness(ctx context.Context) (*SearchBizResult, error)
	// GeBusinessbyID
	GeBusinessbyID(ctx context.Context, bizID uint32) (*Biz, error)
}

// NewClient initialize a new cmdb client
func NewClient(client rest.ClientInterface) Client {
	return &cmdb{
		client: client,
		cache:  gcache.New(1000).Expiration(time.Hour * 24).EvictType(gcache.TYPE_LRU).Build(),
	}
}

// cmdb is an esb client to request cmdb.
type cmdb struct {
	cache gcache.Cache
	// http client instance
	client rest.ClientInterface
}

// SearchBusiness 通用查询
func (c *cmdb) SearchBusiness(ctx context.Context, params *SearchBizParams) (*SearchBizResp, error) {
	resp := new(SearchBizResp)

	req := &esbSearchBizParams{
		SearchBizParams: params,
	}

	h := http.Header{}
	h.Set(constant.RidKey, uuid.UUID())

	err := c.client.Post().
		SubResourcef("/cc/search_business/").
		WithContext(ctx).
		WithHeaders(h).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if !resp.Result || resp.Code != 0 {
		return nil, fmt.Errorf("search business failed, code: %d, msg: %s, rid: %s", resp.Code, resp.Message, resp.Rid)
	}

	return resp, nil
}

// ListAllBusiness 读取全部业务列表
func (c *cmdb) ListAllBusiness(ctx context.Context) (*SearchBizResult, error) {
	params := &SearchBizParams{}
	resp, err := c.SearchBusiness(ctx, params)
	if err != nil {
		return nil, err
	}

	return &resp.SearchBizResult, nil
}

// GeBusinessbyID 读取单个biz
func (c *cmdb) GeBusinessbyID(ctx context.Context, bizID uint32) (*Biz, error) {
	if cacheResult, err := c.cache.Get(bizID); err == nil {
		return cacheResult.(*Biz), nil
	}

	params := &SearchBizParams{
		Page: BasePage{Limit: 1},
		BizPropertyFilter: &QueryFilter{
			Rule: CombinedRule{
				Condition: ConditionAnd,
				Rules: []Rule{
					AtomRule{
						Field:    BizIDField,
						Operator: OperatorEqual,
						Value:    bizID,
					}},
			}},
	}
	resp, err := c.SearchBusiness(ctx, params)
	if err != nil {
		return nil, err
	}
	if len(resp.Info) == 0 {
		return nil, fmt.Errorf("biz %d not found", bizID)
	}

	c.cache.Set(bizID, &resp.Info[0]) //nolint:errcheck

	return &resp.Info[0], nil
}
