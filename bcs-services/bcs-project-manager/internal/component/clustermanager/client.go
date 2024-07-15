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

// Package clustermanager xxx
package clustermanager

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/cache"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/constant"
	common "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
)

var (
	// ErrNotInited err server not init
	ErrNotInited = errors.New("server not init")

	// ClusterStatusRunning cluster status running
	ClusterStatusRunning = "RUNNING"
	// CacheKeyClusterPrefix cluster Prefix
	CacheKeyClusterPrefix = "CLUSTER_%s"
)

type authentication struct {
}

func (a *authentication) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{middleware.InnerClientHeaderKey: constant.ServiceName}, nil
}

func (a *authentication) RequireTransportSecurity() bool {
	return true
}

// ClsManClient xxx
type ClsManClient struct {
	tlsConfig *tls.Config
	disc      *discovery.ModuleDiscovery
}

// ClusterClient xxx
var ClusterClient *ClsManClient

// SetClusterManagerClient set cluster manager client config
func SetClusterManagerClient(tlsConfig *tls.Config, disc *discovery.ModuleDiscovery) {
	ClusterClient = &ClsManClient{
		tlsConfig: tlsConfig,
		disc:      disc,
	}
}

// GetClusterManagerClient get cm client by discovery
func GetClusterManagerClient() (ClusterManagerClient, func(), error) {
	if ClusterClient == nil {
		return nil, nil, ErrNotInited
	}

	if ClusterClient.disc == nil {
		return nil, nil, fmt.Errorf("resourceManager module not enable dsicovery")
	}

	nodeServer, err := ClusterClient.disc.GetRandomServiceNode()
	if err != nil {
		return nil, nil, err
	}

	blog.Infof("ResManClient get node[%s] from disc", nodeServer.Address)
	conf := &Config{
		Hosts:     []string{nodeServer.Address},
		TLSConfig: ClusterClient.tlsConfig,
	}
	cli, closeCon := NewClusterManager(conf)

	return cli, closeCon, nil
}

// Config for connect cm server
type Config struct {
	Hosts     []string
	AuthToken string
	TLSConfig *tls.Config
}

// NewClusterManager create ResourceManager SDK implementation
func NewClusterManager(config *Config) (ClusterManagerClient, func()) {
	// NOCC: gosec/crypto(没有特殊的安全需求)
	//nolint:gosec
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if len(config.Hosts) == 0 {
		// ! pay more attention for nil return
		return nil, nil
	}
	// create grpc connection
	header := map[string]string{
		"x-content-type":                "application/grpc+proto",
		"Content-Type":                  "application/grpc",
		middleware.InnerClientHeaderKey: common.ServiceDomain,
	}
	if len(config.AuthToken) != 0 {
		header["Authorization"] = fmt.Sprintf("Bearer %s", config.AuthToken)
	}
	md := metadata.New(header)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	opts = append(opts, grpc.WithPerRPCCredentials(&authentication{}))
	if config.TLSConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config.TLSConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	var conn *grpc.ClientConn
	var err error
	maxTries := 3
	for i := 0; i < maxTries; i++ {
		selected := r.Intn(1024) % len(config.Hosts)
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

	// init cluster manager client
	// nolint
	return NewClusterManagerClient(conn), func() { conn.Close() }
}

// GetCluster get cluster by clusterID
func GetCluster(clusterID string) (*Cluster, error) {
	// 1. if hit, get from cache
	c := cache.GetCache()
	if cluster, exists := c.Get(fmt.Sprintf(CacheKeyClusterPrefix, clusterID)); exists {
		return cluster.(*Cluster), nil
	}
	cli, closeCon, err := GetClusterManagerClient()
	if err != nil {
		logging.Error("get cluster manager client failed, err: %s", err.Error())
		return nil, err
	}
	defer closeCon()
	req := &GetClusterReq{
		ClusterID: clusterID,
	}
	resp, err := cli.GetCluster(context.Background(), req)
	if err != nil {
		logging.Error("get cluster from cluster manager failed, err: %s", err.Error())
		return nil, err
	}
	if resp.GetCode() != 0 {
		logging.Error("get cluster from cluster manager failed, msg: %s", resp.GetMessage())
		return nil, errors.New(resp.GetMessage())
	}
	_ = c.Add(fmt.Sprintf(CacheKeyClusterPrefix, clusterID), resp.GetData(), 5*time.Minute)
	return resp.GetData(), nil
}

// ListClusters list clusters by projectID
func ListClusters(projectID string) ([]*Cluster, error) {
	cli, closeCon, err := GetClusterManagerClient()
	if err != nil {
		logging.Error("get cluster manager client failed, err: %s", err.Error())
		return nil, err
	}
	defer closeCon()
	req := &ListClusterReq{
		ProjectID: projectID,
		Status:    ClusterStatusRunning,
	}
	resp, err := cli.ListCluster(context.Background(), req)
	if err != nil {
		logging.Error("list clusters from cluster manager failed, err: %s", err.Error())
		return nil, err
	}
	if resp.GetCode() != 0 {
		logging.Error("list clusters from cluster manager failed, msg: %s", resp.GetMessage())
		return nil, errors.New(resp.GetMessage())
	}
	return resp.GetData(), nil
}
