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
	"errors"
	"fmt"
	"net/http"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/iam/auth"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/metrics"
	pbcs "bscp.io/pkg/protocol/config-server"
	"bscp.io/pkg/runtime/grpcgw"
	"bscp.io/pkg/serviced"
	"bscp.io/pkg/tools"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// proxy all server's mux proxy.
type proxy struct {
	cfgSvrMux    *runtime.ServeMux
	repoRevProxy *repoProxy
	state        serviced.State
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

	authorizer, err := auth.NewAuthorizer(dis, cc.ApiServer().Network.TLS)
	if err != nil {
		return nil, fmt.Errorf("new authorizer failed, err: %v", err)
	}

	repoProxy, err := newRepoProxy(authorizer)
	if err != nil {
		return nil, err
	}

	p := &proxy{
		cfgSvrMux:    cfgSvrMux,
		repoRevProxy: repoProxy,
		state:        state,
	}

	return p, nil
}

// handler return proxy handler.
func (p *proxy) handler() http.Handler {
	root := http.NewServeMux()

	// new http mux and set handle.
	apiMux := http.NewServeMux()
	apiMux.Handle("/", p.router())
	root.Handle("/api/v1/", p.setupFilters(apiMux))

	root.HandleFunc("/debug/", http.DefaultServeMux.ServeHTTP)
	root.HandleFunc("/metrics", metrics.Handler().ServeHTTP)
	root.HandleFunc("/healthz", p.Healthz)

	return root
}

// router return proxy router.
func (p *proxy) router() *mux.Router {
	router := mux.NewRouter()

	// register config server interfaces.
	router.PathPrefix("/api/v1/config/").Handler(p.cfgSvrMux)

	// repo http proxy. put: upload
	router.HandleFunc("/api/v1/api/create/content/upload/biz_id/{biz_id}/app_id/{app_id}",
		func(w http.ResponseWriter, req *http.Request) {
			p.repoRevProxy.ServeHTTP(w, req)
		}).Methods(http.MethodPut)

	// repo http proxy. get: download
	router.HandleFunc("/api/v1/api/get/content/download/biz_id/{biz_id}/app_id/{app_id}",
		func(w http.ResponseWriter, req *http.Request) {
			p.repoRevProxy.ServeHTTP(w, req)
		}).Methods(http.MethodGet)

	return router
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

// newGrpcDialOption new grpc dial option by grpc balancer and tls.
func newGrpcDialOption(dis serviced.Discover, tls cc.TLSConfig) ([]grpc.DialOption, error) {
	opts := make([]grpc.DialOption, 0)

	// add dial load balancer.
	opts = append(opts, dis.LBRoundRobin())

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

	return opts, nil
}

// newGrpcMux new grpc mux that has some processing of built-in http request to grpc request.
func newGrpcMux() *runtime.ServeMux {
	return runtime.NewServeMux(grpcgw.MetadataOpt, grpcgw.MarshalerOpt)
}
