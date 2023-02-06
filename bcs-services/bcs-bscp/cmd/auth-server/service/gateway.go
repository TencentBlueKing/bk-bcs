/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/iam/sys"
	"bscp.io/pkg/logs"
	pbas "bscp.io/pkg/protocol/auth-server"
	"bscp.io/pkg/runtime/grpcgw"
	"bscp.io/pkg/serviced"
	"bscp.io/pkg/tools"
)

// gateway auth server's grpc-gateway.
type gateway struct {
	iamSys *sys.Sys
	mux    *runtime.ServeMux
	state  serviced.State
}

// newGateway create new auth server's grpc-gateway.
func newGateway(st serviced.State, iamSys *sys.Sys) (*gateway, error) {
	mux, err := newAuthServerMux()
	if err != nil {
		return nil, err
	}

	g := &gateway{
		state:  st,
		mux:    mux,
		iamSys: iamSys,
	}

	return g, nil
}

// handler return gateway handler.
func (g *gateway) handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", g.mux)
	return g.setFilter(mux)
}

// newAuthServerMux new auth server mux.
func newAuthServerMux() (*runtime.ServeMux, error) {
	opts := make([]grpc.DialOption, 0)

	network := cc.AuthServer().Network
	tls := network.TLS
	if !tls.Enable() {
		// dial without ssl
		opts = append(opts, grpc.WithInsecure())
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

	// build conn.
	addr := net.JoinHostPort(network.BindIP, strconv.Itoa(int(network.RpcPort)))
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		logs.Errorf("dial auth server failed, err: %v", err)
		return nil, err
	}

	// new grpc mux.
	mux := newGrpcMux()

	// register client to mux.
	if err = pbas.RegisterAuthHandler(context.Background(), mux, conn); err != nil {
		logs.Errorf("register auth server handler client failed, err: %v", err)
		return nil, err
	}

	return mux, nil
}

// newGrpcMux new grpc mux that has some processing of built-in http request to grpc request.
func newGrpcMux() *runtime.ServeMux {
	return runtime.NewServeMux(grpcgw.MetadataOpt, grpcgw.MarshalerOpt, grpcgw.ErrorHandlerOpt)
}
