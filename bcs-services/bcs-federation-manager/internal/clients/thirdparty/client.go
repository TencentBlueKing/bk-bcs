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

	"github.com/go-micro/plugins/v4/client/grpc"
	"github.com/go-micro/plugins/v4/registry/etcd"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/registry"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/requester"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/common"
	third "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/pkg/bcsapi/thirdparty-service"
)

const (
	// ResultSuccessKey SUCCESS
	ResultSuccessKey = "SUCCESS"
)

var thirdpartyCli Client

// InitThirdpartyClient set thirdpartyCli client
func InitThirdpartyClient(opts *ClientOptions) {
	cli := NewClient(opts)
	thirdpartyCli = cli
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
func NewClient(opts *ClientOptions) Client {
	header := make(map[string]string)
	header[common.HeaderAuthorizationKey] = fmt.Sprintf("Bearer %s", opts.Token)
	header[common.BcsHeaderClientKey] = common.InnerModuleName
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

	cli := third.NewBcsThirdpartyService(common.ModuleThirdpartyServiceManager, c)
	return &thirdpartyClient{
		debug:         false,
		opts:          opts,
		defaultHeader: header,
		thirdpartySvc: cli,
	}
}

type thirdpartyClient struct {
	debug         bool
	opts          *ClientOptions
	defaultHeader map[string]string
	thirdpartySvc third.BcsThirdpartyService
}

func (t *thirdpartyClient) getMetadataCtx(ctx context.Context) context.Context {
	return metadata.NewContext(ctx, metadata.Metadata{
		common.BcsHeaderClientKey:   common.InnerModuleName,
		common.BcsHeaderUsernameKey: common.InnerModuleName,
	})
}
