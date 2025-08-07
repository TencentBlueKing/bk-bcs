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

// Package clustermanager define client for clustermanager
package clustermanager

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	// pmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	cmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/patrickmn/go-cache"
	"go-micro.dev/v4/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/option"
	pmp "github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/types"
)

const (
	// ModuleClusterManager default discovery clustermanager module
	ModuleClusterManager = "clustermanager.bkbcs.tencent.com"
)

var (
	// errServerNotInit server not inited
	errServerNotInit = errors.New("ClusterManagerClient not inited")
)

// Options for init clusterManager
type Options struct {
	Module          string
	Address         string
	EtcdRegistry    registry.Registry
	ClientTLSConfig *tls.Config
	AuthToken       string
}

func (o *Options) validate() bool {
	if o == nil {
		return false
	}

	if o.Module == "" {
		o.Module = ModuleClusterManager
	}

	return true
}

// NewClusterManagerClient init cluster manager and start discovery module(clustermanager)
func NewClusterManagerClient(opts *Options) client.Client {
	ok := opts.validate()
	if !ok {
		return nil
	}

	cmClient := &clusterManagerClient{
		opts: opts,
		// Create a cache with a default expiration time of 5 minutes, and which
		// purges expired items every 1 hour
		cache: cache.New(time.Minute*5, time.Minute*60),
	}

	if opts.Address != "" {
		return cmClient
	}

	cmClient.ctx, cmClient.cancel = context.WithCancel(context.Background())
	cmClient.discovery = discovery.NewServiceDiscovery(opts.Module, opts.EtcdRegistry)
	err := cmClient.discovery.Start()
	if err != nil {
		blog.Errorf("start discovery client failed: %v", err)
		return nil
	}
	return cmClient
}

// GetClusterManagerConnWithURL get conn with url
func (cm *clusterManagerClient) GetClusterManagerConnWithURL() (*grpc.ClientConn, error) {
	header := map[string]string{
		"x-content-type": "application/grpc+proto",
		"Content-Type":   "application/grpc",
	}
	if len(cm.opts.AuthToken) != 0 {
		header["Authorization"] = fmt.Sprintf("Bearer %s", cm.opts.AuthToken)
	}
	md := metadata.New(header)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	if cm.opts.ClientTLSConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(cm.opts.ClientTLSConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(cm.opts.Address, opts...)
	if err != nil {
		blog.Errorf("Create cluster manager grpc client with %s error: %s", cm.opts.Address, err.Error())
		return nil, err
	}

	if conn == nil {
		blog.Errorf("create no cluster manager client after all instance tries")
		return nil, fmt.Errorf("conn is nil")
	}
	return conn, nil
}

// ClusterManagerClient client for clustermanager
type clusterManagerClient struct {
	opts      *Options
	discovery discovery.Discovery
	cache     *cache.Cache
	ctx       context.Context
	cancel    context.CancelFunc
}

// GetCMDBClient returns a CMDB client instance.
func (cm *clusterManagerClient) GetCMDBClient() (client.CMDBClient, error) {
	// implement me
	panic("implement me")
}

// GetProjectManagerConnWithURL returns a gRPC client connection with URL.
func (cm *clusterManagerClient) GetProjectManagerConnWithURL() (*grpc.ClientConn, error) {
	// implement me
	panic("implement me")
}

// GetProjectManagerClient returns a project manager client instance.
func (cm *clusterManagerClient) GetProjectManagerClient() (pmp.BCSProjectClient, error) {
	// implement me
	panic("implement me")
}

// GetProjectManagerConn returns a gRPC client connection for project manager.
func (cm *clusterManagerClient) GetProjectManagerConn() (*grpc.ClientConn, error) {
	// implement me
	panic("implement me")
}

// NewPMGrpcClientWithHeader creates a new project manager gRPC client with header.
func (cm *clusterManagerClient) NewPMGrpcClientWithHeader(ctx context.Context,
	conn *grpc.ClientConn) *client.ProjectManagerClientWithHeader {
	// implement me
	panic("implement me")
}

// GetStorageClient returns a storage client instance.
func (cm *clusterManagerClient) GetStorageClient() (bcsapi.Storage, error) {
	// implement me
	panic("implement me")
}

// GetDataManagerConnWithURL returns a gRPC client connection with URL for data manager.
func (cm *clusterManagerClient) GetDataManagerConnWithURL() (*grpc.ClientConn, error) {
	// implement me
	panic("implement me")
}

// GetDataManagerConn returns a gRPC client connection for data manager.
func (cm *clusterManagerClient) GetDataManagerConn() (*grpc.ClientConn, error) {
	// implement me
	panic("implement me")
}

// GetClusterManagerClient get cm client
func (cm *clusterManagerClient) GetClusterManagerClient() (cmp.ClusterManagerClient, error) {
	if cm == nil {
		return nil, errServerNotInit
	}

	// get bcs-cluster-manager server from etcd registry
	node, err := cm.discovery.GetRandomServiceInstance()
	if err != nil {
		blog.Errorf("module[%s] GetRandomServiceInstance failed: %v", cm.opts.Module, err)
		return nil, err
	}
	blog.V(4).Infof("get random cluster-manager instance [%s] from etcd registry successful", node.Address)

	cfg := client.Config{}
	// discovery hosts
	cfg.Hosts = []string{node.Address}
	cfg.TLSConfig = cm.opts.ClientTLSConfig
	clusterCli := NewClusterManager(&cfg)

	if clusterCli == nil {
		blog.Errorf("create cluster manager cli from config: %+v failed, please check discovery", cfg)
		return nil, fmt.Errorf("no available clustermanager client")
	}
	return clusterCli, nil
}

// GetClusterManagerConn get conn
func (cm *clusterManagerClient) GetClusterManagerConn() (*grpc.ClientConn, error) {
	if cm == nil {
		return nil, errServerNotInit
	}

	if cm.opts.Address != "" {
		return cm.GetClusterManagerConnWithURL()
	}

	// get bcs-cluster-manager server from etcd registry
	node, err := cm.discovery.GetRandomServiceInstance()
	if err != nil {
		blog.Errorf("module[%s] GetRandomServiceInstance failed: %v", cm.opts.Module, err)
		return nil, err
	}
	blog.V(4).Infof("get random cluster-manager instance [%s] from etcd registry successful", node.Address)

	header := map[string]string{
		"x-content-type": "application/grpc+proto",
		"Content-Type":   "application/grpc",
	}
	md := metadata.New(header)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	if cm.opts.ClientTLSConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(cm.opts.ClientTLSConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	var conn *grpc.ClientConn
	conn, err = grpc.Dial(node.Address, opts...)
	if err != nil {
		blog.Errorf("Create cluster manager grpc client with %s error: %s", node.Address, err.Error())
		return nil, err
	}

	if conn == nil {
		blog.Errorf("create no cluster manager client after all instance tries")
		return nil, fmt.Errorf("conn is nil")
	}
	return conn, nil
}

// Stop stop clusterManagerClient
func (cm *clusterManagerClient) Stop() {
	if cm == nil {
		return
	}

	cm.discovery.Stop()
	cm.cancel()
}

// NewCMGrpcClientWithHeader new client with grpc header
func (cm *clusterManagerClient) NewCMGrpcClientWithHeader(ctx context.Context,
	conn *grpc.ClientConn) *client.ClusterManagerClientWithHeader {
	header := make(map[string]string)
	if len(cm.opts.AuthToken) != 0 {
		header["Authorization"] = fmt.Sprintf("Bearer %s", cm.opts.AuthToken)
	}
	md := metadata.New(header)
	return &client.ClusterManagerClientWithHeader{
		Ctx: metadata.NewOutgoingContext(ctx, md),
		Cli: cmp.NewClusterManagerClient(conn),
	}
}

// NewClusterManager create ClusterManager SDK implementation
func NewClusterManager(config *client.Config) cmp.ClusterManagerClient {
	rand.Seed(time.Now().UnixNano()) // nolint
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
	}
	md := metadata.New(header)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	if config.TLSConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config.TLSConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	var conn *grpc.ClientConn
	var err error
	maxTries := 3
	for i := 0; i < maxTries; i++ {
		selected := rand.Intn(1024) % len(config.Hosts) // nolint
		addr := config.Hosts[selected]
		conn, err = grpc.Dial(addr, opts...)
		if err != nil {
			blog.Errorf("Create clsuter manager grpc client with %s error: %s", addr, err.Error())
			continue
		}
		if conn != nil {
			break
		}
	}
	if conn == nil {
		blog.Errorf("create no cluster manager client after all instance tries")
		return nil
	}
	// init cluster manager client
	return cmp.NewClusterManagerClient(conn)
}

// NewClusterManagerGrpcGwClient return cluster manager grpc gateway client
func NewClusterManagerGrpcGwClient(opts *Options) (cmCli *client.ClusterManagerClientWithHeader, err error) {
	cli := NewClusterManagerClient(opts)
	if cli == nil {
		return nil, fmt.Errorf("init cluster manager client failed")
	}
	cmConn, err := cli.GetClusterManagerConn()
	if err != nil || cmConn == nil {
		return nil, fmt.Errorf("get cluster manager conn failed, err %s", err.Error())
	}

	cmCli = cli.NewCMGrpcClientWithHeader(context.Background(), cmConn)
	_, err = cmCli.Cli.ListCluster(cmCli.Ctx, &cmp.ListClusterReq{})
	if err != nil {
		return nil, fmt.Errorf("cmcli ping error %s", err.Error())
	}
	blog.Infof("init cluster manager client successfully")
	return cmCli, nil
}

// GetClusterManagerGrpcGwClient get cluster manager client
func GetClusterManagerGrpcGwClient() (cmCli *client.ClusterManagerClientWithHeader, err error) {
	tlsConfig, err := option.InitTClientTlsConfig()
	if err != nil {
		return nil, err
	}

	opts := &Options{
		Module:          ModuleClusterManager,
		Address:         option.GetGlobalConfig().Bcsapi.GrpcAddr,
		EtcdRegistry:    nil,
		ClientTLSConfig: tlsConfig,
		AuthToken:       option.GetGlobalConfig().Bcsapi.BearerToken,
	}
	cmCli, err = NewClusterManagerGrpcGwClient(opts)
	if err != nil {
		return nil, err
	}
	return cmCli, nil
}
