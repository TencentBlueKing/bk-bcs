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

// Package bcsproject xxx
package bcsproject

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bcsProject "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	"github.com/patrickmn/go-cache"
	"go-micro.dev/v4/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/discovery"
)

const (
	// ModuleProjectManager default discovery projectmanager module
	ModuleProjectManager = "project.bkbcs.tencent.com"
)

var (
	// errServerNotInit server not inited
	errServerNotInit = errors.New("project Manager Client not inited")
)

// BcsProjectManagerClient interface for bcs project manager client
type BcsProjectManagerClient interface { // nolint
	GetBcsProjectManagerConn() (*grpc.ClientConn, error)
	NewGrpcClientWithHeader(ctx context.Context, conn *grpc.ClientConn) *BcsProjectClientWithHeader
}

// BcsProjectClientWithHeader client for bcs project
type BcsProjectClientWithHeader struct { // nolint
	Cli bcsProject.BCSProjectClient
	Ctx context.Context
}

// Options for init bcs project manager
type Options struct {
	Module          string
	Address         string
	EtcdRegistry    registry.Registry
	ClientTLSConfig *tls.Config
	AuthToken       string
	UserName        string
}

func (o *Options) validate() bool {
	if o == nil {
		return false
	}

	if o.Module == "" {
		o.Module = ModuleProjectManager
	}

	return true
}

// bcsProjectManagerClient client for bcs project manager
type bcsProjectManagerClient struct {
	opts      *Options
	discovery discovery.Discovery
	cache     *cache.Cache // nolint
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewBcsProjectManagerClient init bcs project manager and start discovery module(projectmanager)
func NewBcsProjectManagerClient(opts *Options) BcsProjectManagerClient {
	ok := opts.validate()
	if !ok {
		return nil
	}

	pmClient := &bcsProjectManagerClient{
		opts: opts,
	}

	if opts.Address != "" {
		return pmClient
	}

	pmClient.ctx, pmClient.cancel = context.WithCancel(context.Background())
	pmClient.discovery = discovery.NewServiceDiscovery(opts.Module, opts.EtcdRegistry)
	err := pmClient.discovery.Start()
	if err != nil {
		blog.Errorf("start discovery client failed: %v", err)
		return nil
	}
	return pmClient
}

// NewGrpcClientWithHeader new client with grpc header
func (pm *bcsProjectManagerClient) NewGrpcClientWithHeader(ctx context.Context,
	conn *grpc.ClientConn) *BcsProjectClientWithHeader {
	header := make(map[string]string)
	if len(pm.opts.Address) != 0 && len(pm.opts.AuthToken) != 0 {
		header["Authorization"] = fmt.Sprintf("Bearer %s", pm.opts.AuthToken)
		header["X-Project-Username"] = pm.opts.UserName
	}
	header["X-Bcs-Client"] = "bcs-data-manager"
	md := metadata.New(header)
	return &BcsProjectClientWithHeader{
		Ctx: metadata.NewOutgoingContext(ctx, md),
		Cli: bcsProject.NewBCSProjectClient(conn),
	}
}

// GetBcsProjectManagerConn get conn
func (pm *bcsProjectManagerClient) GetBcsProjectManagerConn() (*grpc.ClientConn, error) {
	if pm == nil {
		return nil, errServerNotInit
	}

	if pm.opts.Address != "" {
		return pm.GetProjectManagerConnWithURL()
	}

	// get bcs-project-manager server from etcd registry
	node, err := pm.discovery.GetRandomServiceInstance()
	if err != nil {
		blog.Errorf("module[%s] GetRandomServiceInstance failed: %v", pm.opts.Module, err)
		return nil, err
	}
	blog.V(4).Infof("get random project-manager instance [%s] from etcd registry successful", node.Address)

	header := map[string]string{
		"x-content-type": "application/grpc+proto",
		"Content-Type":   "application/grpc",
	}
	md := metadata.New(header)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	if pm.opts.ClientTLSConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(pm.opts.ClientTLSConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	var conn *grpc.ClientConn
	conn, err = grpc.Dial(node.Address, opts...)
	if err != nil {
		blog.Errorf("Create project manager grpc client with %s error: %s", node.Address, err.Error())
		return nil, err
	}

	if conn == nil {
		blog.Errorf("create no project manager client after all instance tries")
		return nil, fmt.Errorf("conn is nil")
	}
	return conn, nil
}

// GetProjectManagerConnWithURL get conn with url
func (pm *bcsProjectManagerClient) GetProjectManagerConnWithURL() (*grpc.ClientConn, error) {
	header := map[string]string{
		"x-content-type": "application/grpc+proto",
		"Content-Type":   "application/grpc",
	}
	if len(pm.opts.AuthToken) != 0 {
		header["Authorization"] = fmt.Sprintf("Bearer %s", pm.opts.AuthToken)
	}
	md := metadata.New(header)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	if pm.opts.ClientTLSConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(pm.opts.ClientTLSConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(pm.opts.Address, opts...)
	if err != nil {
		blog.Errorf("Create project manager grpc client with %s error: %s", pm.opts.Address, err.Error())
		return nil, err
	}

	if conn == nil {
		blog.Errorf("create no project manager client after all instance tries")
		return nil, fmt.Errorf("conn is nil")
	}
	return conn, nil
}
