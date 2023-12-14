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

// Package client NOTES
package client

import (
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/test/client/api"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/test/client/cache"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/test/client/feed"
)

// Client for suite test.
type Client struct {
	ApiClient   *api.Client
	CacheClient *cache.Client
	FeedClient  *feed.Client
}

// Config suite test config
type Config struct {
	ApiHost   string
	CacheHost string
	FeedHost  string
}

// NewClient get a new suite client
func NewClient(cfg Config) (*Client, error) {

	// the suite test client don't need TLS certificate, so give a nil
	apiCli, err := api.NewApiClient(cfg.ApiHost, nil)
	if err != nil {
		return nil, err
	}

	cacheCli, err := cache.NewCacheClient(cfg.CacheHost)
	if err != nil {
		return nil, err
	}

	feedCli, err := feed.NewFeedClient(cfg.FeedHost, nil)
	if err != nil {
		return nil, err
	}

	return &Client{
		ApiClient:   apiCli,
		CacheClient: cacheCli,
		FeedClient:  feedCli,
	}, nil
}
