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

// Package app NOTES
package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/tcp/listener"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/api-server/options"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/api-server/service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/components/bknotice"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/metrics"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/shutdown"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// Run start the api server
func Run(opt *options.Option) error {
	as := new(apiServer)
	if err := as.prepare(opt); err != nil {
		return err
	}

	if err := as.listenAndServe(); err != nil {
		return err
	}
	shutdown.RegisterFirstShutdown(as.finalizer)
	shutdown.WaitShutdown(20)
	return nil
}

type apiServer struct {
	serve    *http.Server
	service  *service.Service
	discover serviced.Discover
}

// prepare do prepare jobs before run api discover.
func (as *apiServer) prepare(opt *options.Option) error {
	// load settings from config file.
	if err := cc.LoadSettings(opt.Sys); err != nil {
		return fmt.Errorf("load settings from config files failed, err: %v", err)
	}

	logs.InitLogger(cc.ApiServer().Log.Logs())

	logs.Infof("load settings from config file success.")

	// init metrics
	metrics.InitMetrics(net.JoinHostPort(cc.ApiServer().Network.BindIP,
		strconv.Itoa(int(cc.ApiServer().Network.HttpPort))))
	metrics.Register().MustRegister(metrics.BSCPServerHandledTotal)

	etcdOpt, err := cc.ApiServer().Service.Etcd.ToConfig()
	if err != nil {
		return fmt.Errorf("get etcd config failed, err: %v", err)
	}

	// new discovery client.
	dis, err := serviced.NewDiscovery(etcdOpt)
	if err != nil {
		return fmt.Errorf("new discovery faield, err: %v", err)
	}

	as.discover = dis
	logs.Infof("create discovery success.")

	// register system to bknotice service
	if cc.ApiServer().BKNotice.Enable {
		if err := bknotice.RegisterSystem(context.TODO()); err != nil {
			logs.Errorf("register system to bknotice failed, err: %v", err)
		}
	}

	return nil
}

// listenAndServe listen the http serve and set up the shutdown gracefully job.
func (as *apiServer) listenAndServe() error {
	svc, err := service.NewService(as.discover)
	if err != nil {
		return fmt.Errorf("initialize service failed, err: %v", err)
	}

	as.service = svc

	network := cc.ApiServer().Network
	addr := tools.GetListenAddr(network.BindIP, int(network.HttpPort))
	ipv6Addr := tools.GetListenAddr(network.BindIPv6, int(network.HttpPort))
	dualStackListener := listener.NewDualStackListener()
	if e := dualStackListener.AddListenerWithAddr(addr); e != nil {
		return e
	}

	if network.BindIPv6 != "" && network.BindIPv6 != network.BindIP {
		if e := dualStackListener.AddListenerWithAddr(ipv6Addr); e != nil {
			return e
		}
		logs.Infof("api serve dualStackListener with ipv6: %s", ipv6Addr)
	}

	handler, err := as.service.Handler()
	if err != nil {
		return err
	}
	as.serve = &http.Server{Addr: addr, Handler: handler}

	go func() {
		notifier := shutdown.AddNotifier()
		<-notifier.Signal
		logs.Infof("start shutdown api server http server gracefully...")

		_ = as.serve.Close()
		notifier.Done()

		logs.Infof("shutdown api server http server success...")
	}()

	if network.TLS.Enable() {
		tls := network.TLS
		tlsC, err := tools.ClientTLSConfVerify(tls.InsecureSkipVerify, tls.CAFile, tls.CertFile, tls.KeyFile,
			tls.Password)
		if err != nil {
			return fmt.Errorf("init tls config failed, err: %v", err)
		}

		as.serve.TLSConfig = tlsC

		go func() {
			if err := as.serve.ServeTLS(dualStackListener, "", ""); err != nil {
				logs.Errorf("https server listen and serve failed, err: %v", err)
				shutdown.SignalShutdownGracefully()
			}
		}()
	} else {
		go func() {
			if err := as.serve.Serve(dualStackListener); err != nil {
				logs.Errorf("http server listen and serve failed, err: %v", err)
				shutdown.SignalShutdownGracefully()
			}
		}()
	}
	logs.Infof("api server listen and serve success. addr=%s", addr)

	return nil
}

func (as *apiServer) finalizer() {
	// for structural consistency
}
