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

// Package api wraps Alibaba Cloud SDK clients for cloud provider operations.
package api

import (
	"errors"
	"regexp"
	"sync"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	sts "github.com/alibabacloud-go/sts-20150401/v2/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

const (
	connectTimeout = 5000
	readTimeout    = 10000
)

// Bound credential input without depending on the length of current AccessKeys.
var accessKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9+/=_-]{1,128}$`)

// Client is the shared entry point for Alibaba Cloud API operations.
// Service SDK clients are initialized on first use and reused by this client.
type Client struct {
	config    openapi.Config
	stsOnce   sync.Once
	stsClient *sts.Client
	stsErr    error
}

// NewClient validates credentials and captures shared SDK settings without making API calls.
// Region is optional for STS; regional operations must require it when implemented.
func NewClient(opt *cloudprovider.CommonOption) (*Client, error) {
	if opt == nil || opt.Account == nil {
		return nil, errors.New("alicloud account is required")
	}
	if !accessKeyPattern.MatchString(opt.Account.SecretID) || !accessKeyPattern.MatchString(opt.Account.SecretKey) {
		return nil, errors.New("alicloud AccessKeyId and AccessKeySecret must contain 1-128 valid characters")
	}
	config := openapi.Config{
		AccessKeyId:     tea.String(opt.Account.SecretID),
		AccessKeySecret: tea.String(opt.Account.SecretKey),
		Protocol:        tea.String("https"),
		ConnectTimeout:  tea.Int(connectTimeout),
		ReadTimeout:     tea.Int(readTimeout),
	}
	if opt.Region != "" {
		config.RegionId = tea.String(opt.Region)
	}
	return &Client{config: config}, nil
}

// clientConfig gives each service its own endpoint and configuration copy.
func (c *Client) clientConfig(endpoint string) *openapi.Config {
	config := c.config
	config.Endpoint = tea.String(endpoint)
	return &config
}
