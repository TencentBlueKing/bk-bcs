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

// Package helm xxx
package helm

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/go-micro/plugins/v4/client/grpc"
	"github.com/go-micro/plugins/v4/registry/etcd"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/registry"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/helmmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/requester"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/types"
)

var helmCli Client

// SetHelmClient set helm client
func SetHelmClient(opts *ClientOptions) {
	cli := NewClient(opts)
	helmCli = cli
}

// GetHelmClient get helm client
func GetHelmClient() Client {
	return helmCli
}

// Client client interface for request cluster manager
type Client interface {
	// General Functions for helm manager
	IsInstalled(opt *HelmOptions) (bool, error)
	InstallRelease(opt *HelmOptions, helmValues ...string) error
	UninstallRelease(opt *HelmOptions) error

	// install federation modules
	GetFederationCharts() *types.FederationCharts
	IsInstalledForFederation(opt *ReleaseBaseOptions) (bool, error)
	InstallClusternetHub(opt *ReleaseBaseOptions) error
	InstallClusternetScheduler(opt *ReleaseBaseOptions) error
	InstallClusternetController(opt *ReleaseBaseOptions) error
	InstallUnifiedApiserver(opt *BcsUnifiedApiserverOptions) error
	InstallClusternetAgent(opt *BcsClusternetAgentOptions) error
	InstallEstimatorAgent(opt *BcsEstimatorAgentOptions) error

	// uninstall federation modules
	UninstallClusternetHub(opt *ReleaseBaseOptions) error
	UninstallClusternetScheduler(opt *ReleaseBaseOptions) error
	UninstallClusternetController(opt *ReleaseBaseOptions) error
	UninstallUnifiedApiserver(opt *BcsUnifiedApiserverOptions) error
	UninstallClusternetAgent(opt *BcsClusternetAgentOptions) error
	UninstallEstimatorAgent(opt *BcsEstimatorAgentOptions) error
}

// ClientOptions options for create client
type ClientOptions struct {
	ClientTLS     *tls.Config
	EtcdEndpoints []string
	EtcdTLS       *tls.Config
	Charts        *types.FederationCharts
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

	// init helm manager cli
	c := grpc.NewClient(
		client.Registry(etcd.NewRegistry(
			registry.Addrs(opts.EtcdEndpoints...),
			registry.TLSConfig(opts.EtcdTLS)),
		),
		grpc.AuthTLS(opts.ClientTLS),
	)
	cli := helmmanager.NewHelmManagerService(common.ModuleHelmManager, c)

	return &helmClient{
		debug:         false,
		opts:          opts,
		defaultHeader: header,
		helmSvc:       cli,
	}
}

type helmClient struct {
	debug         bool
	opts          *ClientOptions
	defaultHeader map[string]string
	helmSvc       helmmanager.HelmManagerService
}

func (c *helmClient) getMetadataCtx(ctx context.Context) context.Context {
	return metadata.NewContext(ctx, metadata.Metadata{
		common.BcsHeaderClientKey:   common.InnerModuleName,
		common.BcsHeaderUsernameKey: common.InnerModuleName,
	})
}
