/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
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

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/criteria/uuid"
	"bscp.io/pkg/dal/cache"
	"bscp.io/pkg/rest"
	"bscp.io/pkg/thirdparty/esb/types"
)

// Client is an esb client to request cmdb.
type Client interface {
	// SearchBusiness 通用查询
	SearchBusiness(ctx context.Context, params *SearchBizParams) (*SearchBizResp, error)
	// ListAllBusiness 读取全部业务列表
	ListAllBusiness(ctx context.Context) (*SearchBizResult, error)
	// GeBusinessbyID
	GeBusinessbyID(ctx context.Context, bizID string) (*Biz, error)
}

// NewClient initialize a new cmdb client
func NewClient(client rest.ClientInterface, config *cc.Esb) Client {
	return &cmdb{
		client: client,
		config: config,
	}
}

// cmdb is an esb client to request cmdb.
type cmdb struct {
	config *cc.Esb
	// http client instance
	client rest.ClientInterface
}

// SearchBusiness 通用查询
func (c *cmdb) SearchBusiness(ctx context.Context, params *SearchBizParams) (*SearchBizResp, error) {
	resp := new(SearchBizResp)

	req := &esbSearchBizParams{
		CommParams:      types.GetCommParams(c.config),
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
func (c *cmdb) GeBusinessbyID(ctx context.Context, bizID string) (*Biz, error) {
	key := fmt.Sprintf("cmdb.GeBusinessbyID:%s", bizID)
	if cacheResult, ok := cache.LocalCache.Slot.Get(key); ok {
		return cacheResult.(*Biz), nil
	}

	params := &SearchBizParams{
		Fields: []string{},
		Condition: map[string]string{
			"bk_biz_id": bizID,
		},
	}
	resp, err := c.SearchBusiness(ctx, params)
	if err != nil {
		return nil, err
	}
	if len(resp.Info) == 0 {
		return nil, fmt.Errorf("biz %s not found", bizID)
	}

	cache.LocalCache.Slot.Set(key, &resp.Info[0], time.Hour*24)

	return &resp.Info[0], nil
}
