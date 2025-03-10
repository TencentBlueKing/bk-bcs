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

// Package server xxx
package server

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/ipv6server"
	"github.com/Tencent/bk-bcs/bcs-common/common/tcp/listener"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/header"
	grpccli "github.com/go-micro/plugins/v4/client/grpc"
	"github.com/go-micro/plugins/v4/registry/etcd"
	grpcsvr "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"google.golang.org/grpc"
	grpccred "google.golang.org/grpc/credentials"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/app/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/constants"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/handler"
	pb "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/proto"
)

// MicroServer v2 server
type MicroServer struct {
	op *options.StorageOptions

	// server
	grpcServer micro.Service
	httpServer *ipv6server.IPv6Server

	ctx        context.Context
	cancelFunc context.CancelFunc

	serverTLSConfig *tls.Config
	clientTLSConfig *tls.Config
	etcdTLSConfig   *tls.Config
}

// NewMicroServer 创建 MicroServer
func NewMicroServer(ctx context.Context, cancelFunc context.CancelFunc) *MicroServer {
	return &MicroServer{
		ctx:        ctx,
		cancelFunc: cancelFunc,
	}
}

// stop close server
func (m *MicroServer) stop() {
	m.cancelFunc()
}

// initHTTPServer 初始化httpServer
func (m *MicroServer) initHTTPServer() error {
	grpcDialOpts := make([]grpc.DialOption, 0)

	gMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(header.CustomHeaderMatcher),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{}),
	)

	if m.clientTLSConfig == nil || m.serverTLSConfig == nil {
		// nolint
		grpcDialOpts = append(grpcDialOpts, grpc.WithInsecure())
	} else {
		grpcDialOpts = append(grpcDialOpts, grpc.WithTransportCredentials(grpccred.NewTLS(m.clientTLSConfig)))
	}

	if err := pb.RegisterStorageGwFromEndpoint(m.ctx, gMux, net.JoinHostPort(m.op.Address,
		strconv.FormatUint(m.op.GRPCPort, 10)), grpcDialOpts); err != nil {
		return errors.Wrapf(err, "register http gateway failed")
	}

	m.httpServer = ipv6server.NewTlsIPv6Server([]string{m.op.Address, m.op.IPv6Address},
		strconv.FormatUint(m.op.HttpPort, 10), "tcp", m.serverTLSConfig, gMux,
	)

	return nil
}

// initGrpcServer 初始化grpcServer
func (m *MicroServer) initGrpcServer() (err error) {
	port := strconv.FormatUint(m.op.GRPCPort, 10)

	ipv4 := m.op.Address
	ipv6 := m.op.IPv6Address

	metadata := make(map[string]string)
	metadata[constants.MicroMetaKeyHTTPPort] = port

	// 适配单栈环境（ipv6注册地址不能是本地回环地址）
	if v := net.ParseIP(ipv6); v != nil && !v.IsLoopback() {
		metadata[types.IPV6] = net.JoinHostPort(ipv6, port)
	}

	globalRegistry := etcd.NewRegistry(
		registry.Addrs(strings.Split(m.op.Etcd.Address, ",")...),
		registry.TLSConfig(m.etcdTLSConfig),
	)

	// 创建双栈监听
	dualStackListener := listener.NewDualStackListener()
	if err = dualStackListener.AddListener(ipv4, port); err != nil { // 添加IPv4地址监听
		return errors.Wrapf(err, "add IPv4 address failed")
	}
	if err = dualStackListener.AddListener(ipv6, port); err != nil { // 添加IPv6地址监听
		return errors.Wrapf(err, "add IPv6 address failed")
	}

	// 创建go-micro服务
	m.grpcServer = micro.NewService(
		micro.Server(grpcsvr.NewServer(
			grpcsvr.AuthTLS(m.serverTLSConfig),
			// 注入双栈监听
			grpcsvr.Listener(dualStackListener),
		)),
		micro.Client(
			grpccli.NewClient(
				grpccli.AuthTLS(m.clientTLSConfig),
			),
		),
		micro.Name(constants.ServerV4Name),
		micro.Version(version.BcsVersion),
		micro.Metadata(metadata),
		micro.Address(net.JoinHostPort(ipv4, port)),
		micro.Registry(globalRegistry),
		micro.Context(m.ctx),
		micro.RegisterTTL(30*time.Second),
		micro.RegisterInterval(25*time.Second),
	)

	if err = pb.RegisterStorageHandler(m.grpcServer.Server(), handler.New()); err != nil {
		return errors.Wrapf(err, "go-micro server registration failed")
	}

	return nil
}

// runHttpServer 运行httpServer
func (m *MicroServer) runHttpServer() (err error) {
	blog.Infof("run v2 http server")
	if m.serverTLSConfig != nil {
		err = m.httpServer.ListenAndServeTLS("", "")
	} else {
		err = m.httpServer.ListenAndServe()
	}
	return errors.Wrapf(err, "run http server failed")
}

// runHttpServer 运行grpcServer
func (m *MicroServer) runGrpcSever() error {
	blog.Infof("run v2 grpc server")
	return m.grpcServer.Run()
}

// Init 初始化 MicroServer
func (m *MicroServer) Init(op *options.StorageOptions) (err error) {
	m.op = op
	if err = m.initGrpcServer(); err != nil {
		return errors.Wrapf(err, "init grpc server failed")
	}
	if err = m.initHTTPServer(); err != nil {
		return errors.Wrapf(err, "init http server failed")
	}
	return nil
}

// Run 运行MicroServer
func (m *MicroServer) Run() error {
	stopChan := make(chan error, 1)

	defer m.stop()

	go func() {
		stopChan <- m.runGrpcSever()
	}()

	go func() {
		stopChan <- m.runHttpServer()
	}()

	select {
	case err := <-stopChan:
		return err
	case <-m.ctx.Done():
		return nil
	}
}

// SetEtcdTLSConfig set etcd tls config
func (m *MicroServer) SetEtcdTLSConfig(c *tls.Config) {
	m.etcdTLSConfig = c
}

// SetServerTLSConfig set server tls config
func (m *MicroServer) SetServerTLSConfig(c *tls.Config) {
	m.serverTLSConfig = c
}

// SetClientTLSConfig set client tls config
func (m *MicroServer) SetClientTLSConfig(c *tls.Config) {
	m.clientTLSConfig = c
}
