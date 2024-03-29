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

// Package projectmanager define client for projectmanager
package projectmanager

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	pmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	cmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/micro/go-micro/v2/registry"
	"github.com/patrickmn/go-cache"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/discovery"
)

const (
	// ModuleProjectManager default discovery project manager module
	ModuleProjectManager = "project.bkbcs.tencent.com"
)

var (
	// errServerNotInit server not inited
	errServerNotInit = errors.New("ProjectManagerClient not inited")
)

// Options for init projectManager
type Options struct {
	Module          string
	Address         string
	EtcdRegistry    registry.Registry
	ClientTLSConfig *tls.Config
	AuthToken       string
	Username        string
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

// NewProjectManagerClient init project manager and start discovery module(projectmanager)
func NewProjectManagerClient(opts *Options) client.Client {
	ok := opts.validate()
	if !ok {
		return nil
	}

	pmClient := &projectManagerClient{
		opts: opts,
		// Create a cache with a default expiration time of 5 minutes, and which
		// purges expired items every 1 hour
		cache: cache.New(time.Minute*5, time.Minute*60),
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

// GetProjectManagerConnWithURL get conn with url
func (pm *projectManagerClient) GetProjectManagerConnWithURL() (*grpc.ClientConn, error) {
	header := map[string]string{
		"x-content-type": "application/grpc+proto",
		"Content-Type":   "application/grpc",
	}
	if len(pm.opts.AuthToken) != 0 {
		header["Authorization"] = fmt.Sprintf("Bearer %s", pm.opts.AuthToken)
		header["X-Project-Username"] = pm.opts.Username
	}
	md := metadata.New(header)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	if pm.opts.ClientTLSConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(pm.opts.ClientTLSConfig)))
	} else {
		opts = append(opts, grpc.WithInsecure())
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

// ProjectManagerClient client for projectmanager
type projectManagerClient struct {
	opts      *Options
	discovery discovery.Discovery
	cache     *cache.Cache
	ctx       context.Context
	cancel    context.CancelFunc
}

// GetCMDBClient returns the CMDBClient instance.
func (pm *projectManagerClient) GetCMDBClient() (client.CMDBClient, error) {
	// Implement me
	panic("implement me")
}

// GetClusterManagerConnWithURL returns a gRPC client connection with URL.
func (pm *projectManagerClient) GetClusterManagerConnWithURL() (*grpc.ClientConn, error) {
	// Implement me
	panic("implement me")
}

// GetClusterManagerClient returns the ClusterManagerClient instance.
func (pm *projectManagerClient) GetClusterManagerClient() (cmp.ClusterManagerClient, error) {
	// Implement me
	panic("implement me")
}

// GetClusterManagerConn returns a gRPC client connection.
func (pm *projectManagerClient) GetClusterManagerConn() (*grpc.ClientConn, error) {
	// Implement me
	panic("implement me")
}

// NewCMGrpcClientWithHeader returns a new ClusterManagerClientWithHeader instance.
func (pm *projectManagerClient) NewCMGrpcClientWithHeader(ctx context.Context,
	conn *grpc.ClientConn) *client.ClusterManagerClientWithHeader {
	// Implement me
	panic("implement me")
}

// GetStorageClient returns the Storage instance.
func (pm *projectManagerClient) GetStorageClient() (bcsapi.Storage, error) {
	// Implement me
	panic("implement me")
}

// GetDataManagerConnWithURL returns a gRPC client connection with URL.
func (pm *projectManagerClient) GetDataManagerConnWithURL() (*grpc.ClientConn, error) {
	// Implement me
	panic("implement me")
}

// GetDataManagerConn returns a gRPC client connection.
func (pm *projectManagerClient) GetDataManagerConn() (*grpc.ClientConn, error) {
	// Implement me
	panic("implement me")
}

// GetProjectManagerClient get pm client
func (pm *projectManagerClient) GetProjectManagerClient() (pmp.BCSProjectClient, error) {
	if pm == nil {
		return nil, errServerNotInit
	}

	// get bcs-cluster-manager server from etcd registry
	node, err := pm.discovery.GetRandomServiceInstance()
	if err != nil {
		blog.Errorf("module[%s] GetRandomServiceInstance failed: %v", pm.opts.Module, err)
		return nil, err
	}
	blog.V(4).Infof("get random project-manager instance [%s] from etcd registry successful", node.Address)

	cfg := client.Config{}
	// discovery hosts
	cfg.Hosts = []string{node.Address}
	cfg.TLSConfig = pm.opts.ClientTLSConfig
	projectCli := NewProjectManager(&cfg)

	if projectCli == nil {
		blog.Errorf("create project manager cli from config: %+v failed, please check discovery", cfg)
		return nil, fmt.Errorf("no available projectmanager client")
	}
	return projectCli, nil
}

// GetProjectManagerConn get conn
func (pm *projectManagerClient) GetProjectManagerConn() (*grpc.ClientConn, error) {
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
		opts = append(opts, grpc.WithInsecure())
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

// Stop stop projectManagerClient
func (pm *projectManagerClient) Stop() {
	if pm == nil {
		return
	}

	pm.discovery.Stop()
	pm.cancel()
}

// NewPMGrpcClientWithHeader new client with grpc header
func (pm *projectManagerClient) NewPMGrpcClientWithHeader(ctx context.Context,
	conn *grpc.ClientConn) *client.ProjectManagerClientWithHeader {
	header := make(map[string]string)
	if len(pm.opts.AuthToken) != 0 {
		header["Authorization"] = fmt.Sprintf("Bearer %s", pm.opts.AuthToken)
		header["X-Project-Username"] = pm.opts.Username
	}
	md := metadata.New(header)
	return &client.ProjectManagerClientWithHeader{
		Ctx: metadata.NewOutgoingContext(ctx, md),
		Cli: pmp.NewBCSProjectClient(conn),
	}
}

// NewProjectManager create ClusterManager SDK implementation
func NewProjectManager(config *client.Config) pmp.BCSProjectClient {
	rand.Seed(time.Now().UnixNano())
	if len(config.Hosts) == 0 {
		//! pay more attention for nil return
		return nil
	}
	// create grpc connection
	header := map[string]string{
		"x-content-type": "application/grpc+proto",
		"Content-Type":   "application/grpc",
	}
	if len(config.AuthToken) != 0 {
		header["Authorization"] = fmt.Sprintf("Bearer %s", config.AuthToken)
		header["X-Project-Username"] = config.Username
	}
	md := metadata.New(header)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	if config.TLSConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config.TLSConfig)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	var conn *grpc.ClientConn
	var err error
	maxTries := 3
	for i := 0; i < maxTries; i++ {
		selected := rand.Intn(1024) % len(config.Hosts) // nolint math/rand instead of crypto/rand
		addr := config.Hosts[selected]
		conn, err = grpc.Dial(addr, opts...)
		if err != nil {
			blog.Errorf("Create project manager grpc client with %s error: %s", addr, err.Error())
			continue
		}
		if conn != nil {
			break
		}
	}
	if conn == nil {
		blog.Errorf("create no project manager client after all instance tries")
		return nil
	}
	// init cluster manager client
	return pmp.NewBCSProjectClient(conn)
}

// NewProjectManagerGrpcGwClient return cluster manager grpc gateway client
func NewProjectManagerGrpcGwClient(opts *Options) (pmCli *client.ProjectManagerClientWithHeader, err error) {
	cli := NewProjectManagerClient(opts)
	if cli == nil {
		return nil, fmt.Errorf("init project manager client failed")
	}
	pmConn, err := cli.GetProjectManagerConn()
	if err != nil || pmConn == nil {
		return nil, fmt.Errorf("get project manager conn failed, err %s", err.Error())
	}

	pmCli = cli.NewPMGrpcClientWithHeader(context.Background(), pmConn)
	_, err = pmCli.Cli.ListProjects(pmCli.Ctx, &pmp.ListProjectsRequest{})
	if err != nil {
		return nil, fmt.Errorf("pmcli ping error %s", err.Error())
	}
	blog.Infof("init project manager client successfully")
	return pmCli, nil
}
