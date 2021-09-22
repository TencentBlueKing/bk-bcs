/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-public-cluster-webhook/pkg/check"
	_ "github.com/Tencent/bk-bcs/bcs-k8s/bcs-public-cluster-webhook/pkg/metrics" // 注册prometheus指标
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-public-cluster-webhook/pkg/options"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// WebhookServer server for bcs webhook
type WebhookServer struct {
	Option    *options.ServerOption
	BlackList *check.BlackList
	Server    *http.Server
}

// NewWebhookServer new webhook server from options
func NewWebhookServer(opt *options.ServerOption) (*WebhookServer, error) {
	pair, err := tls.LoadX509KeyPair(opt.ServerCertFile, opt.ServerKeyFile)
	if err != nil {
		return nil, fmt.Errorf("load x509 key pair failed, err %s", err)
	}

	blackList, err := check.NewBlackList(opt.BlacklistConfig)
	if err != nil {
		return nil, fmt.Errorf("init blackList error, err %s", err)
	}

	whsvr := &WebhookServer{
		Option:    opt,
		BlackList: blackList,
		Server: &http.Server{
			Addr:      fmt.Sprintf("%s:%v", opt.Address, opt.Port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		},
	}

	return whsvr, nil
}

// Run run server
func (ws *WebhookServer) Run() {
	// define http server and server handler
	mux := http.NewServeMux()
	mux.HandleFunc("/bcs/webhook/validate/v1/k8s", ws.Validate)
	ws.Server.Handler = mux

	// start webhook server in new routine
	go func() {
		if err := ws.Server.ListenAndServeTLS("", ""); err != nil {
			blog.Errorf("Failed to listen and serve webhook server, err %s", err.Error())
		}
	}()

	blog.Infof("webhook server started")

	// save pid
	if err := common.SavePid(ws.Option.ProcessConfig); err != nil {
		blog.Errorf("fail to save pid, err:%s", err.Error())
	}

	// run prometheus server
	runPrometheusMetricsServer(ws.Option)

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	blog.Infof("Got OS shutdown signal, shutting down webhook server gracefully...")
	_ = ws.Server.Shutdown(context.Background())
}

func runPrometheusMetricsServer(opt *options.ServerOption) {
	blog.Infof("begin register prometheus metrics server: port(%d)", opt.MetricPort)

	// register prometheus server
	http.Handle("/metrics", promhttp.Handler())
	addr := opt.Address + ":" + strconv.Itoa(int(opt.MetricPort))
	go func() {
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			blog.Error("http.ListenAndServe error, %v", err)
		}
	}()

	blog.Infof("run prometheus server ok")
}
