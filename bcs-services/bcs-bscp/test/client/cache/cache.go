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

// Package cache NOTES
package cache

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/cache-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// Client cache service client
type Client struct {
	client pbcs.CacheClient
}

// NewCacheClient get a new cache client
func NewCacheClient(host string) (*Client, error) {
	conn, err := grpc.Dial(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		client: pbcs.NewCacheClient(conn),
	}, nil
}

// GetAppMeta get application meta data
func (c *Client) GetAppMeta(ctx context.Context, req *pbcs.GetAppMetaReq) (*types.AppCacheMeta, error) {
	resp, err := c.client.GetAppMeta(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, errf.New(errf.Unknown, "response is nil")
	}

	respData := new(types.AppCacheMeta)
	err = json.Unmarshal([]byte(resp.JsonRaw), respData)
	if err != nil {
		return nil, err
	}

	return respData, nil
}
