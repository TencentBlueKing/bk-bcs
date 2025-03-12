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

// Package user xxx
package user

import (
	"crypto/tls"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/requester"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/common"
)

var userCli Client

// Client user manager client interface
type Client interface {
	// GetUserToken get user token
	GetUserToken(username string) (string, error)
}

// ClientOptions options for create client
type ClientOptions struct {
	ClientTLS *tls.Config
	requester.BaseOptions
}

// SetUserClient set user manager client
func SetUserClient(opt *ClientOptions) {
	cli := NewClient(opt)
	userCli = cli
}

// GetUserClient get user manager client
func GetUserClient() Client {
	return userCli
}

// NewClient create client with options
func NewClient(opts *ClientOptions) Client {
	header := make(map[string]string)
	header[common.HeaderAuthorizationKey] = fmt.Sprintf("Bearer %s", opts.Token)
	header[common.BcsHeaderClientKey] = common.InnerModuleName
	if opts.Sender == nil {
		opts.Sender = requester.NewRequester()
	}

	return &userClient{
		opt:           opts,
		defaultHeader: header,
	}
}

type userClient struct {
	opt           *ClientOptions
	defaultHeader map[string]string
}
