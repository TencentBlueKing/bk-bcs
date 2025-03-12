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

// Package project xxx
package project

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/go-micro/plugins/v4/client/grpc"
	"github.com/go-micro/plugins/v4/registry/etcd"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/registry"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/requester"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/common"
)

var projectCli Client

// SetProjectClient set project client
func SetProjectClient(opts *ClientOptions) {
	cli := NewClient(opts)
	projectCli = cli
}

// GetProjectClient get project client
func GetProjectClient() Client {
	return projectCli
}

// Client client interface for request cluster manager
type Client interface {
	GetProject(context.Context, string) (*bcsproject.Project, error)
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

	// init project manager cli
	c := grpc.NewClient(
		client.Registry(etcd.NewRegistry(
			registry.Addrs(opts.EtcdEndpoints...),
			registry.TLSConfig(opts.EtcdTLS)),
		),
		grpc.AuthTLS(opts.ClientTLS),
	)
	cli := bcsproject.NewBCSProjectService(common.ModuleProjectManager, c)

	return &projectClient{
		debug:         false,
		opts:          opts,
		defaultHeader: header,
		projectSvc:    cli,
	}
}

type projectClient struct {
	debug         bool
	opts          *ClientOptions
	defaultHeader map[string]string
	projectSvc    bcsproject.BCSProjectService
}

func (c *projectClient) getMetadataCtx(ctx context.Context) context.Context {
	return metadata.NewContext(ctx, metadata.Metadata{
		common.BcsHeaderClientKey:   common.InnerModuleName,
		common.BcsHeaderUsernameKey: common.InnerModuleName,
	})
}
