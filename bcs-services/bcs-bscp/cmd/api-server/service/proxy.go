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

package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/iam/auth"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbas "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/auth-server"
	pbcs "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/grpcgw"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// proxy all server's mux proxy.
type proxy struct {
	cfgSvrMux           *runtime.ServeMux
	authSvrMux          http.Handler
	repo                *repoService
	state               serviced.State
	authorizer          auth.Authorizer
	cfgClient           pbcs.ConfigClient
	configImportService *configImport
}

// newProxy create new mux proxy.
func newProxy(dis serviced.Discover) (*proxy, error) {
	state, ok := dis.(serviced.State)
	if !ok {
		return nil, errors.New("discover convert state failed")
	}

	cfgSvrMux, err := newCfgServerMux(dis)
	if err != nil {
		return nil, err
	}

	authSvrMux, err := newAuthServerMux(dis)
	if err != nil {
		return nil, err
	}

	authorizer, err := auth.NewAuthorizer(dis, cc.ApiServer().Network.TLS)
	if err != nil {
		return nil, fmt.Errorf("new authorizer failed, err: %v", err)
	}

	repo, err := newRepoService(cc.ApiServer().Repo, authorizer)
	if err != nil {
		return nil, err
	}

	cfgClient, err := newCfgClient(dis)
	if err != nil {
		return nil, err
	}

	configImportService, err := newConfigImportService(cc.ApiServer().Repo, authorizer, cfgClient)
	if err != nil {
		return nil, err
	}

	p := &proxy{
		cfgSvrMux:           cfgSvrMux,
		repo:                repo,
		configImportService: configImportService,
		state:               state,
		authorizer:          authorizer,
		authSvrMux:          authSvrMux,
		cfgClient:           cfgClient,
	}

	p.initBizsOfTmplSpaces()

	return p, nil
}

func newAuthServerMux(dis serviced.Discover) (http.Handler, error) {
	opts, err := newGrpcDialOption(dis, cc.ApiServer().Network.TLS)
	if err != nil {
		return nil, err
	}

	// build conn.
	conn, err := grpc.Dial(serviced.GrpcServiceDiscoveryName(cc.AuthServerName), opts...)
	if err != nil {
		logs.Errorf("dial auth server failed, err: %v", err)
		return nil, err
	}

	// new grpc mux.
	mux := runtime.NewServeMux(grpcgw.MetadataOpt, grpcgw.JsonMarshalerOpt, grpcgw.BKErrorHandlerOpt,
		grpcgw.BSCPResponseOpt)

	// register client to mux.
	if err = pbas.RegisterAuthHandler(context.Background(), mux, conn); err != nil {
		logs.Errorf("register config server handler client failed, err: %v", err)
		return nil, err
	}

	return mux, nil
}

// newCfgServerMux new config server mux.
func newCfgServerMux(dis serviced.Discover) (*runtime.ServeMux, error) {
	opts, err := newGrpcDialOption(dis, cc.ApiServer().Network.TLS)
	if err != nil {
		return nil, err
	}

	// build conn.
	conn, err := grpc.Dial(serviced.GrpcServiceDiscoveryName(cc.ConfigServerName), opts...)
	if err != nil {
		logs.Errorf("dial config server failed, err: %v", err)
		return nil, err
	}

	// new grpc mux.
	mux := newGrpcMux()

	// register client to mux.
	if err = pbcs.RegisterConfigHandler(context.Background(), mux, conn); err != nil {
		logs.Errorf("register config server handler client failed, err: %v", err)
		return nil, err
	}

	return mux, nil
}

// newCfgClient new config client.
func newCfgClient(dis serviced.Discover) (pbcs.ConfigClient, error) {
	opts, err := newGrpcDialOption(dis, cc.ApiServer().Network.TLS)
	if err != nil {
		return nil, err
	}

	// build conn.
	conn, err := grpc.Dial(serviced.GrpcServiceDiscoveryName(cc.ConfigServerName), opts...)
	if err != nil {
		logs.Errorf("dial config server failed, err: %v", err)
		return nil, err
	}

	return pbcs.NewConfigClient(conn), nil
}

// newGrpcDialOption new grpc dial option by grpc balancer and tls.
func newGrpcDialOption(dis serviced.Discover, tls cc.TLSConfig) ([]grpc.DialOption, error) {
	opts := make([]grpc.DialOption, 0)

	// add dial load balancer.
	opts = append(opts, dis.LBRoundRobin())

	if !tls.Enable() {
		// dial without ssl
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		// dial with ssl.
		tlsC, err := tools.ClientTLSConfVerify(tls.InsecureSkipVerify, tls.CAFile, tls.CertFile, tls.KeyFile,
			tls.Password)
		if err != nil {
			return nil, fmt.Errorf("init grpc tls config failed, err: %v", err)
		}

		cred := credentials.NewTLS(tlsC)
		opts = append(opts, grpc.WithTransportCredentials(cred))
	}

	return opts, nil
}

// newGrpcMux new grpc mux that has some processing of built-in http request to grpc request.
func newGrpcMux() *runtime.ServeMux {
	return runtime.NewServeMux(grpcgw.MetadataOpt, grpcgw.JsonMarshalerOpt, grpcgw.BKErrorHandlerOpt,
		grpcgw.BSCPResponseOpt)
}
