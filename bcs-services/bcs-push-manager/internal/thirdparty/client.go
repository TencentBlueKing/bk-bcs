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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/discovery"
	headerpkg "github.com/Tencent/bk-bcs/bcs-common/pkg/header"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	third "github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/pkg/bcsapi/thirdparty-service"
)

const (
	innerClientName = "bcs-push-manager"
)

var thirdpartyCli Client

// InitThirdpartyClient set thirdpartyCli client
func InitThirdpartyClient(opts *ClientOptions) error {
	cli, err := NewClient(opts)
	if err != nil {
		blog.Errorf("failed to initialize thirdparty client: %v", err)
		return err
	}
	thirdpartyCli = cli
	return nil
}

// GetThirdpartyClient get thirdparty client
func GetThirdpartyClient() Client {
	return thirdpartyCli
}

// CloseThirdpartyClient closes the thirdparty client connection
func CloseThirdpartyClient() error {
	if thirdpartyCli != nil {
		return thirdpartyCli.Close()
	}
	return nil
}

// Client client interface
type Client interface {
	SendRtx(req *third.SendRtxRequest) error
	SendMail(req *third.SendMailRequest) error
	SendMsg(req *third.SendMsgRequest) error
	Close() error
}

// ClientOptions options for create client
type ClientOptions struct {
	ClientTLS *tls.Config
	Discovery *discovery.ModuleDiscovery
}

// NewClient create client with options
func NewClient(opts *ClientOptions) (Client, error) {
	var addr string

	if discovery.UseServiceDiscovery() {
		addr = fmt.Sprintf("%s:%d", discovery.ThirdpartyServiceName, discovery.ServiceGrpcPort)
	} else {
		// etcd 服务发现
		if opts.Discovery == nil {
			return nil, fmt.Errorf("thirdparty service module not enable discovery")
		}
		nodeServer, err := opts.Discovery.GetRandomServiceNode()
		if err != nil {
			return nil, fmt.Errorf("failed to get random service node: %v", err)
		}
		addr = nodeServer.Address
	}

	header := map[string]string{
		"x-content-type": "application/grpc+proto",
		"Content-Type":   "application/grpc",
	}
	md := metadata.New(header)
	var grpcOpts []grpc.DialOption
	grpcOpts = append(grpcOpts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	auth := &bcsapi.Authentication{InnerClientName: innerClientName}
	grpcOpts = append(grpcOpts, grpc.WithUnaryInterceptor(headerpkg.LaneHeaderInterceptor()))

	if opts.ClientTLS != nil {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(credentials.NewTLS(opts.ClientTLS)))
	} else {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		auth.Insecure = true
	}
	grpcOpts = append(grpcOpts, grpc.WithPerRPCCredentials(auth))

	conn, err := grpc.Dial(addr, grpcOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial grpc server %s: %v", addr, err)
	}

	return &thirdpartyClient{
		thirdpartySvc: third.NewBcsThirdpartyServiceClient(conn),
		conn:          conn,
	}, nil
}

type thirdpartyClient struct {
	thirdpartySvc third.BcsThirdpartyServiceClient
	conn          *grpc.ClientConn
}

// Close closes the grpc connection
func (t *thirdpartyClient) Close() error {
	if t.conn != nil {
		if err := t.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}
