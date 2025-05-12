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

// Package helmmanager xxx
package helmmanager

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/discovery"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/header"
	headerpkg "github.com/Tencent/bk-bcs/bcs-common/pkg/header"
)

var (
	clientConfig *bcsapi.ClientConfig
)

// SetClientConfig set helm manager client config
// disc nil 表示使用k8s 内置的service 进行服务访问
func SetClientConfig(tlsConfig *tls.Config, disc *discovery.ModuleDiscovery) {
	clientConfig = &bcsapi.ClientConfig{
		TLSConfig: tlsConfig,
		Discovery: disc,
	}
}

// GetClient get cm client by discovery
func GetClient(innerClientName string) (*HelmClientWrapper, func(), error) {
	if clientConfig == nil {
		return nil, nil, bcsapi.ErrNotInited
	}
	var addr string
	if discovery.UseServiceDiscovery() {
		addr = fmt.Sprintf("%s:%d", discovery.HelmManagerServiceName, discovery.ServiceGrpcPort)
	} else {
		if clientConfig.Discovery == nil {
			return nil, nil, fmt.Errorf("helm manager module not enable discovery")
		}

		nodeServer, err := clientConfig.Discovery.GetRandomServiceNode()
		if err != nil {
			return nil, nil, err
		}
		addr = nodeServer.Address
	}
	klog.Infof("get helm manager client with address: %s", addr)
	conf := &bcsapi.Config{
		Hosts:           []string{addr},
		TLSConfig:       clientConfig.TLSConfig,
		InnerClientName: innerClientName,
	}
	cli, closeCon, err := NewHelmClientWrapper(conf)
	if err != nil {
		return nil, nil, err
	}

	return cli, closeCon, nil
}

// HelmClientWrapper helm manager client wrapper
type HelmClientWrapper struct {
	helmManagerClient   HelmManagerClient
	ClusterAddonsClient ClusterAddonsClient
}

// newHelmConn creates a grpc connection for helm manager clients
func newHelmConn(config *bcsapi.Config) (*grpc.ClientConn, func(), error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano())) // nolint
	if len(config.Hosts) == 0 {
		return nil, nil, fmt.Errorf("no hosts provided")
	}
	// create grpc connection
	header := map[string]string{
		"x-content-type":            "application/grpc+proto",
		"Content-Type":              "application/grpc",
		header.InnerClientHeaderKey: config.InnerClientName,
	}
	if len(config.AuthToken) != 0 {
		header["Authorization"] = fmt.Sprintf("Bearer %s", config.AuthToken)
	}
	for k, v := range config.Header {
		header[k] = v
	}
	md := metadata.New(header)
	auth := &bcsapi.Authentication{InnerClientName: config.InnerClientName}
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	if config.TLSConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config.TLSConfig)))
	} else {
		opts = append(opts, grpc.WithInsecure()) // nolint
		auth.Insecure = true
	}
	opts = append(opts, grpc.WithPerRPCCredentials(auth))
	if config.AuthToken != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(bcsapi.NewTokenAuth(config.AuthToken)))
	}
	opts = append(opts, grpc.WithUnaryInterceptor(headerpkg.LaneHeaderInterceptor()))

	var conn *grpc.ClientConn
	var err error
	maxTries := 3
	for i := 0; i < maxTries; i++ {
		selected := r.Intn(1024) % len(config.Hosts)
		addr := config.Hosts[selected]
		conn, err = grpc.Dial(addr, opts...)
		if err != nil {
			klog.Errorf("Create helm manager grpc client with %s error: %s", addr, err.Error())
			continue
		}
		if conn != nil {
			break
		}
	}
	if conn == nil {
		klog.Errorf("create no helm manager client after all instance tries")
		return nil, nil, fmt.Errorf("failed to create grpc connection")
	}
	return conn, func() { _ = conn.Close() }, nil
}

// NewHelmClientWrapper create HelmManager SDK implementation
func NewHelmClientWrapper(config *bcsapi.Config) (*HelmClientWrapper, func(), error) {
	conn, closeFn, err := newHelmConn(config)
	if err != nil {
		return nil, nil, err
	}
	return &HelmClientWrapper{
		helmManagerClient:   NewHelmManagerClient(conn),
		ClusterAddonsClient: NewClusterAddonsClient(conn),
	}, closeFn, nil
}

// NewHelmClient create HelmManager SDK implementation
func NewHelmClient(config *bcsapi.Config) (HelmManagerClient, func(), error) {
	conn, closeFn, err := newHelmConn(config)
	if err != nil {
		return nil, nil, err
	}
	return NewHelmManagerClient(conn), closeFn, nil
}

// NewHelmAddonsClient create HelmManager addons SDK implementation
func NewHelmAddonsClient(config *bcsapi.Config) (ClusterAddonsClient, func(), error) {
	conn, closeFn, err := newHelmConn(config)
	if err != nil {
		return nil, nil, err
	}
	return NewClusterAddonsClient(conn), closeFn, nil
}
