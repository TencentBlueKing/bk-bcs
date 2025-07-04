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
	"fmt"
	"net/url"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

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
	ClientTLS *tls.Config
	Endpoint  string
	AuthToken string
}

// NewClient create client with options
func NewClient(opts *ClientOptions) (Client, error) {
	header := map[string]string{
		"x-content-type": "application/grpc+proto",
		"content-type":   "application/grpc",
	}
	if len(opts.AuthToken) != 0 {
		header["authorization"] = fmt.Sprintf("Bearer %s", opts.AuthToken)
	}
	md := metadata.New(header)

	var grpcOpts []grpc.DialOption
	grpcOpts = append(grpcOpts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	if opts.ClientTLS != nil {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(credentials.NewTLS(opts.ClientTLS)))
	} else {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	address := opts.Endpoint
	if u, err := url.Parse(opts.Endpoint); err == nil && u.Host != "" {
		address = u.Host
	}

	conn, err := grpc.Dial(address, grpcOpts...)
	if err != nil {
		blog.Errorf("failed to dial thirdparty service: %v", err)
		return nil, fmt.Errorf("failed to dial thirdparty service: %w", err)
	}

	cli := third.NewBcsThirdpartyServiceClient(conn)
	return &thirdpartyClient{
		opts:          opts,
		defaultHeader: header,
		thirdpartySvc: cli,
		conn:          conn,
	}, nil
}

type thirdpartyClient struct {
	opts          *ClientOptions
	defaultHeader map[string]string
	thirdpartySvc third.BcsThirdpartyServiceClient
	conn          *grpc.ClientConn
}

// Close closes the gRPC connection associated with the thirdpartyClient.
func (c *thirdpartyClient) Close() error {
	return c.conn.Close()
}
