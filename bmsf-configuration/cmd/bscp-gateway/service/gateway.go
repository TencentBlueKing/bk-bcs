/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
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
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/coreos/etcd/clientv3"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"bk-bscp/internal/framework"
	pb "bk-bscp/internal/protocol/accessserver"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/grpclb"
	"bk-bscp/pkg/logger"
)

// Gateway is bscp gRPC gateway.
type Gateway struct {
	// settings for server.
	setting framework.Setting

	// configs handler.
	viper *viper.Viper

	// gRPC gateway server mux.
	gwmux *runtime.ServeMux

	// http server mux.
	mux *http.ServeMux

	// access server gRPC connection/client.
	accSvrConn *grpclb.GRPCConn
	accSvrCli  pb.AccessClient
}

// NewGateway creates a new access server instance.
func NewGateway() *Gateway {
	return &Gateway{}
}

// Init initialize the settings.
func (gw *Gateway) Init(setting framework.Setting) {
	gw.setting = setting
}

// initialize config and check base content.
func (gw *Gateway) initConfig() {
	cfg := config{}
	viper, err := cfg.init(gw.setting.Configfile)
	if err != nil {
		log.Fatal(err)
	}
	gw.viper = viper
}

// initialize logger.
func (gw *Gateway) initLogger() {
	logger.InitLogger(logger.LogConfig{
		LogDir:          gw.viper.GetString("logger.directory"),
		LogMaxSize:      gw.viper.GetUint64("logger.maxsize"),
		LogMaxNum:       gw.viper.GetInt("logger.maxnum"),
		ToStdErr:        gw.viper.GetBool("logger.stderr"),
		AlsoToStdErr:    gw.viper.GetBool("logger.alsoStderr"),
		Verbosity:       gw.viper.GetInt32("logger.level"),
		StdErrThreshold: gw.viper.GetString("logger.stderrThreshold"),
		VModule:         gw.viper.GetString("logger.vmodule"),
		TraceLocation:   gw.viper.GetString("traceLocation"),
	})
	logger.Info("logger init success dir[%s] level[%d].",
		gw.viper.GetString("logger.directory"), gw.viper.GetInt32("logger.level"))

	logger.Info("dump configs: gateway[%+v, %+v, %+v, %+v] accessserver[%+v, %+v] etcdCluster[%+v, %+v] TLS[%+v, %+v, %+v, %+v]",
		gw.viper.Get("gateway.endpoint.ip"), gw.viper.Get("gateway.endpoint.port"), gw.viper.Get("gateway.api.dir"), gw.viper.Get("gateway.api.open"),
		gw.viper.Get("accessserver.servicename"), gw.viper.Get("accessserver.calltimeout"), gw.viper.Get("etcdCluster.endpoints"), gw.viper.Get("etcdCluster.dialtimeout"),
		gw.viper.Get("gateway.tls.certPassword"), gw.viper.Get("gateway.tls.cafile"), gw.viper.Get("gateway.tls.certfile"), gw.viper.Get("gateway.tls.keyfile"))
}

// create a mux for gateway server.
func (gw *Gateway) initGWMux() {
	opt := runtime.WithMarshalerOption(runtime.MIMEWildcard,
		&runtime.JSONPb{EnumsAsInts: true, EmitDefaults: true, OrigName: true})
	gw.gwmux = runtime.NewServeMux(opt)
}

// create a mux for http server..
func (gw *Gateway) initSvrMux() {
	gw.mux = http.NewServeMux()

	// handle gateway.
	gw.mux.Handle("/", gw.gwmux)

	// handle swagger files.
	if gw.viper.GetBool("gateway.api.open") {
		gw.mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, path.Join(gw.viper.GetString("gateway.api.dir"), strings.TrimPrefix(r.URL.Path, "/swagger/")))
		})
	}
}

// create access server gRPC client.
func (gw *Gateway) initAccessServerClient() {
	var etcdCfg clientv3.Config

	caFile := gw.viper.GetString("etcdCluster.tls.cafile")
	certFile := gw.viper.GetString("etcdCluster.tls.certfile")
	keyFile := gw.viper.GetString("etcdCluster.tls.keyfile")
	certPassword := gw.viper.GetString("etcdCluster.tls.certPassword")

	if len(caFile) != 0 || len(certFile) != 0 || len(keyFile) != 0 {
		tlsConf, err := ssl.ClientTslConfVerity(caFile, certFile, keyFile, certPassword)
		if err != nil {
			logger.Fatalf("load etcd tls files failed, %+v", err)
		}
		etcdCfg = clientv3.Config{
			Endpoints:   gw.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: gw.viper.GetDuration("etcdCluster.dialtimeout"),
			TLS:         tlsConf,
		}

	} else {
		etcdCfg = clientv3.Config{
			Endpoints:   gw.viper.GetStringSlice("etcdCluster.endpoints"),
			DialTimeout: gw.viper.GetDuration("etcdCluster.dialtimeout"),
		}
	}

	ctx := &grpclb.Context{
		Target:     gw.viper.GetString("accessserver.servicename"),
		EtcdConfig: etcdCfg,
	}

	// gRPC dial options, with insecure and timeout.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(gw.viper.GetDuration("accessserver.calltimeout")),
	}

	// build gRPC client of access server.
	conn, err := grpclb.NewGRPCConn(ctx, opts...)
	if err != nil {
		logger.Fatal("can't create access server gRPC client, %+v", err)
	}
	gw.accSvrConn = conn
	gw.accSvrCli = pb.NewAccessClient(conn.Conn())
	logger.Info("create access server gRPC client success.")
}

// initMods initialize the server modules.
func (gw *Gateway) initMods() {
	// initialize access server gRPC client.
	gw.initAccessServerClient()

	// initialize gateway mux.
	gw.initGWMux()

	// initialize server mux.
	gw.initSvrMux()
}

// Run runs access server
func (gw *Gateway) Run() {
	// initialize config.
	gw.initConfig()

	// initialize logger.
	gw.initLogger()
	defer gw.Stop()

	// initialize server modules.
	gw.initMods()

	// register access server handler.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := pb.RegisterAccessHandlerClient(ctx, gw.gwmux, gw.accSvrCli); err != nil {
		logger.Fatal("gateway register access server handler from endpoint, %+v", err)
	}

	// gateway service listen and serve.
	if !(gw.viper.IsSet("gateway.endpoint.ip") && gw.viper.IsSet("gateway.endpoint.port")) &&
		!(gw.viper.IsSet("gateway.insecureEndpoint.ip") && gw.viper.IsSet("gateway.insecureEndpoint.port")) {
		logger.Fatal("no available http server here, check local config")
	}

	// secure http server.
	if gw.viper.IsSet("gateway.endpoint.ip") && gw.viper.IsSet("gateway.endpoint.port") {
		gwEndpoint := common.Endpoint(gw.viper.GetString("gateway.endpoint.ip"), gw.viper.GetInt("gateway.endpoint.port"))
		httpServer := &http.Server{Addr: gwEndpoint, Handler: gw.mux}

		// http server with TLS must verify client in gateway.
		tlsConf, err := ssl.ServerTslConfVerityClient(
			gw.viper.GetString("gateway.tls.cafile"), gw.viper.GetString("gateway.tls.certfile"),
			gw.viper.GetString("gateway.tls.keyfile"), gw.viper.GetString("gateway.tls.certPassword"))
		if err != nil {
			logger.Fatal("gateway setup https server failed, %+v", err)
		}
		httpServer.TLSConfig = tlsConf

		// listen and serve with TLS.
		go func() {
			logger.Info("Gateway with TLS running now.")
			if err := httpServer.ListenAndServeTLS("", ""); err != nil {
				logger.Fatal("https server listen and serve, %+v", err)
			}
		}()
	}

	// insecure http server.
	if gw.viper.IsSet("gateway.insecureEndpoint.ip") && gw.viper.IsSet("gateway.insecureEndpoint.port") {
		gwInsecureEndpoint := common.Endpoint(gw.viper.GetString("gateway.insecureEndpoint.ip"), gw.viper.GetInt("gateway.insecureEndpoint.port"))
		httpInsecureServer := &http.Server{Addr: gwInsecureEndpoint, Handler: gw.mux}

		// listen and serve without TLS.
		go func() {
			logger.Info("Gateway without TLS running now.")
			if err := httpInsecureServer.ListenAndServe(); err != nil {
				logger.Fatal("http server listen and serve, %+v", err)
			}
		}()
	}

	// hanging here.
	select {}
}

// Stop stops the gateway.
func (gw *Gateway) Stop() {
	// close access server gRPC connection when server exit.
	if gw.accSvrConn != nil {
		gw.accSvrConn.Close()
	}

	// close logger.
	logger.CloseLogs()
}
