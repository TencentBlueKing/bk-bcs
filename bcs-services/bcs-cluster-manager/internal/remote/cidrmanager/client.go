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

// Package cidrmanager xxxx
package cidrmanager

import (
	"crypto/tls"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

var (
	// ErrNotInited err server not init
	ErrNotInited = errors.New("server not init")
)

// Options for cidr-manager
type Options struct {
	// Enable enable discovery
	Enable bool
	// Module module name
	Module string
	// other configInfo
	TLSConfig *tls.Config
}

// cidrClient  global cidr client
var cidrClient *CidrManageClient

// SetCidrClient set global cidr client
func SetCidrClient(opts *Options, disc *discovery.ModuleDiscovery) {
	cidrClient = &CidrManageClient{
		opts: opts,
		disc: disc,
	}
}

// GetCidrClient get cidr client
func GetCidrClient() *CidrManageClient {
	return cidrClient
}

// CidrManageClient client
type CidrManageClient struct {
	opts *Options
	disc *discovery.ModuleDiscovery
}

// GetCidrManagerClient get cidrManager client
func (rm *CidrManageClient) GetCidrManagerClient() (CidrManagerClient, func(), error) {
	if rm == nil {
		return nil, nil, ErrNotInited
	}

	if rm.disc == nil {
		return nil, nil, fmt.Errorf("resourceManager module not enable dsicovery")
	}

	nodeServer, err := rm.disc.GetRandomServiceNode()
	if err != nil {
		return nil, nil, err
	}
	endpoints := utils.GetServerEndpointsFromRegistryNode(nodeServer)

	blog.Infof("ResManClient get node[%s] from disc", nodeServer.Address)
	conf := &Config{
		Hosts:     endpoints,
		TLSConfig: rm.opts.TLSConfig,
	}

	cli, closeCon := NewCidrManager(conf)

	return cli, closeCon, nil
}

// Config xxx
type Config struct {
	Hosts     []string
	AuthToken string
	TLSConfig *tls.Config
}

// NewCidrManager create CidrManager SDK implementation
func NewCidrManager(config *Config) (CidrManagerClient, func()) {
	rand.Seed(time.Now().UnixNano()) // nolint
	if len(config.Hosts) == 0 {
		//! pay more attention for nil return
		return nil, nil
	}
	// create grpc connection
	header := map[string]string{
		"x-content-type": "application/grpc+proto",
		"Content-Type":   "application/grpc",
	}
	if len(config.AuthToken) != 0 {
		header["Authorization"] = fmt.Sprintf("Bearer %s", config.AuthToken)
	}
	md := metadata.New(header)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	if config.TLSConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config.TLSConfig)))
	} else {
		// nolint
		opts = append(opts, grpc.WithInsecure())
	}
	var conn *grpc.ClientConn
	var err error
	maxTries := 3
	for i := 0; i < maxTries; i++ {
		selected := rand.Intn(1024) % len(config.Hosts) // nolint
		addr := config.Hosts[selected]
		conn, err = grpc.Dial(addr, opts...)
		if err != nil {
			blog.Errorf("Create resource manager grpc client with %s error: %s", addr, err.Error())
			continue
		}
		if conn != nil {
			break
		}
	}
	if conn == nil {
		blog.Errorf("create no resource manager client after all instance tries")
		return nil, nil
	}

	// init cidr manager client
	return NewCidrManagerClient(conn), func() { conn.Close() } // nolint
}
