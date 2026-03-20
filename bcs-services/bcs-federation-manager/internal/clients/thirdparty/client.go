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

// Package thirdparty xxx
package thirdparty

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/requester"
	third "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/pkg/bcsapi/thirdparty-service"
)

const (
	// ResultSuccessKey SUCCESS
	ResultSuccessKey = "SUCCESS"
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

// Client client interface for request cluster manager
type Client interface {
	CreateModule(moduleName string) (*third.CreateModuleResponse, error)

	UpdateQuotaInfoForTaiji(req *third.UpdateQuotaInfoForTaijiRequest) error
	CreateNamespaceForTaijiV3(req *third.CreateNamespaceForTaijiV3Request) error
	GetKubeConfigForTaiji(namespace string) (*third.GetKubeConfigForTaijiResponse, error)

	CreateNamespaceForSuanli(req *third.CreateNamespaceForSuanliRequest) error
	UpdateQuotaInfoForSuanli(req *third.UpdateNamespaceForSuanliRequest) error
	GetKubeConfigForSuanli(namespace string) (*third.GetKubeConfigForSuanliResponse, error)
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
	header := map[string]string{
		"x-content-type": "application/grpc+proto",
		"Content-Type":   "application/grpc",
	}
	if len(opts.Token) != 0 {
		header["Authorization"] = fmt.Sprintf("Bearer %s", opts.Token)
	}
	md := metadata.New(header)
	var grpcOpts []grpc.DialOption
	grpcOpts = append(grpcOpts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	if opts.ClientTLS != nil {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(credentials.NewTLS(opts.ClientTLS)))
	} else {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	var conn *grpc.ClientConn
	// 解析 URL
	parsedURL, err := url.Parse(opts.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("error parsing URL: %s", err.Error())
	}
	conn, err = grpc.NewClient(parsedURL.Host, grpcOpts...)
	if err != nil {
		return nil, fmt.Errorf("dial bcs service failed:%s", err.Error())
	}

	if conn == nil {
		return nil, fmt.Errorf("conn is nil")
	}
	return &thirdpartyClient{
		debug:         false,
		opts:          opts,
		defaultHeader: header,
		thirdpartySvc: third.NewBcsThirdpartyServiceClient(conn),
		conn:          conn,
	}, nil
}

type thirdpartyClient struct {
	debug         bool
	opts          *ClientOptions
	defaultHeader map[string]string
	thirdpartySvc third.BcsThirdpartyServiceClient
	conn          *grpc.ClientConn
}

func (c *thirdpartyClient) getMetadataCtx(ctx context.Context) context.Context {
	header := map[string]string{
		"x-content-type": "application/grpc+proto",
		"Content-Type":   "application/grpc",
	}
	header["Authorization"] = fmt.Sprintf("Bearer %s", c.opts.Token)
	md := metadata.New(header)
	return metadata.NewOutgoingContext(ctx, md)
}
