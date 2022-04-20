/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmanager

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	bcsCm "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/discovery"
	"github.com/micro/go-micro/v2/registry"
	"github.com/patrickmn/go-cache"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
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
func NewClusterManagerClient(opts *Options) *ClusterManagerClient {
	ok := opts.validate()
	if !ok {
		return nil
	}

	cmClient := &ClusterManagerClient{
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
func (cm *ClusterManagerClient) GetClusterManagerConnWithURL() (*grpc.ClientConn, error) {
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
		opts = append(opts, grpc.WithInsecure())
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
type ClusterManagerClient struct {
	opts      *Options
	discovery discovery.Discovery
	cache     *cache.Cache
	ctx       context.Context
	cancel    context.CancelFunc
}

// GetClusterManagerClient get cm client
func (cm *ClusterManagerClient) GetClusterManagerClient() (bcsCm.ClusterManagerClient, error) {
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

	cfg := bcsapi.Config{}
	// discovery hosts
	cfg.Hosts = []string{node.Address}
	cfg.TLSConfig = cm.opts.ClientTLSConfig
	clusterCli := bcsapi.NewClusterManager(&cfg)

	if clusterCli == nil {
		blog.Errorf("create cluster manager cli from config: %+v failed, please check discovery", cfg)
		return nil, fmt.Errorf("no available clustermanager client")
	}
	return clusterCli, nil
}

// GetClusterManagerConn get conn
func (cm *ClusterManagerClient) GetClusterManagerConn() (*grpc.ClientConn, error) {
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
		opts = append(opts, grpc.WithInsecure())
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
func (cm *ClusterManagerClient) Stop() {
	if cm == nil {
		return
	}

	cm.discovery.Stop()
	cm.cancel()
}

// ClusterManagerClientWithHeader client for cluster manager
type ClusterManagerClientWithHeader struct {
	Cli bcsCm.ClusterManagerClient
	Ctx context.Context
}

// NewGrpcClientWithHeader new client with grpc header
func (cm *ClusterManagerClient) NewGrpcClientWithHeader(ctx context.Context,
	conn *grpc.ClientConn) *ClusterManagerClientWithHeader {
	header := make(map[string]string)
	if len(cm.opts.AuthToken) != 0 {
		header["Authorization"] = fmt.Sprintf("Bearer %s", cm.opts.AuthToken)
	}
	md := metadata.New(header)
	return &ClusterManagerClientWithHeader{
		Ctx: metadata.NewOutgoingContext(ctx, md),
		Cli: bcsCm.NewClusterManagerClient(conn),
	}
}
