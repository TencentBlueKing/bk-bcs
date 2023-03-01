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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/dal/repository"
	"bscp.io/pkg/iam/auth"
	"bscp.io/pkg/logs"
	pbas "bscp.io/pkg/protocol/auth-server"
	pbcs "bscp.io/pkg/protocol/config-server"
	"bscp.io/pkg/runtime/grpcgw"
	"bscp.io/pkg/runtime/handler"
	"bscp.io/pkg/serviced"
	"bscp.io/pkg/tools"
)

// proxy all server's mux proxy.
type proxy struct {
	cfgSvrMux    *runtime.ServeMux
	authSvrMux   http.Handler
	repoRevProxy repository.FileApiType
	state        serviced.State
	authorizer   auth.Authorizer
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

	repoProxy, err := newRepoProxy(authorizer)
	if err != nil {
		return nil, err
	}

	p := &proxy{
		cfgSvrMux:    cfgSvrMux,
		repoRevProxy: repoProxy,
		state:        state,
		authorizer:   authorizer,
		authSvrMux:   authSvrMux,
	}

	return p, nil
}

// handler return proxy handler.
func (p *proxy) handler() http.Handler {
	r := chi.NewRouter()
	r.Use(handler.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(handler.CORS)
	// r.Use(middleware.Timeout(60 * time.Second))

	r.HandleFunc("/healthz", p.Healthz)
	r.Mount("/", handler.RegisterCommonHandler())
	// 用户信息
	r.With(p.authorizer.UnifiedAuthentication).Get("/api/v1/auth/user/info", UserInfoHandler)
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Use(p.authorizer.UnifiedAuthentication)
		r.Mount("/", p.authSvrMux)
	})

	// 服务管理, 配置管理, 分组管理, 发布管理
	r.Route("/api/v1/config/", func(r chi.Router) {
		r.Use(p.authorizer.UnifiedAuthentication)
		r.Mount("/", p.cfgSvrMux)
	})

	// repo 上传 API
	r.Route("/api/v1/api/create/content/upload", func(r chi.Router) {
		r.Use(p.authorizer.UnifiedAuthentication)
		r.Put("/biz_id/{biz_id}/app_id/{app_id}", p.repoRevProxy.UploadFile)
	})

	// repo 下载 API
	r.Route("/api/v1/api/get/content/download", func(r chi.Router) {
		r.Use(p.authorizer.UnifiedAuthentication)
		r.Get("/biz_id/{biz_id}/app_id/{app_id}", p.repoRevProxy.DownloadFile)
	})

	return r
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
	mux := runtime.NewServeMux(grpcgw.MetadataOpt, grpcgw.BKJSONMarshalerOpt, grpcgw.BKErrorHandlerOpt)

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
	return runtime.NewServeMux(grpcgw.MetadataOpt, grpcgw.JsonMarshalerOpt)
}
