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

// Package thirdparty provides a client for interacting with bcs-thirdparty-service.
package thirdparty

import (
	"crypto/tls"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/go-micro/plugins/v4/client/grpc"
	"github.com/go-micro/plugins/v4/registry/etcd"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/registry"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/requester"
	third "github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/pkg/bcsapi/thirdparty-service"
)

var thirdpartyCli Client

// InitThirdpartyClient set thirdpartyCli client
func InitThirdpartyClient(opts *ClientOptions) {
	cli, err := NewClient(opts)
	if err != nil {
		blog.Errorf("failed to initialize thirdparty client: %v", err)
		return
	}
	thirdpartyCli = cli
}

// GetThirdpartyClient get thirdparty client
func GetThirdpartyClient() Client {
	return thirdpartyCli
}

// Client client interface
type Client interface {
	SendRtx(req *third.SendRtxRequest) error
	SendMail(req *third.SendMailRequest) error
}

// ClientOptions options for create client
type ClientOptions struct {
	ClientTLS     *tls.Config
	EtcdEndpoints []string
	EtcdTLS       *tls.Config
	requester.BaseOptions
}

// NewClient create client with options
func NewClient(opts *ClientOptions) (Client, error) {
	if opts.Sender == nil {
		opts.Sender = requester.NewRequester()
	}

	// init thirdparty manager cli
	c := grpc.NewClient(
		client.Registry(etcd.NewRegistry(
			registry.Addrs(opts.EtcdEndpoints...),
			registry.TLSConfig(opts.EtcdTLS)),
		),
		grpc.AuthTLS(opts.ClientTLS),
	)

	cli := third.NewBcsThirdpartyService(constant.ModuleThirdpartyServiceManager, c)
	return &thirdpartyClient{
		debug:         false,
		opts:          opts,
		thirdpartySvc: cli,
	}, nil
}

type thirdpartyClient struct {
	debug         bool
	opts          *ClientOptions
	thirdpartySvc third.BcsThirdpartyService
}
