/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

// Package app NOTES
package app

import (
	"fmt"
	"net"
	"net/http"
	"strconv"

	"bscp.io/cmd/api-server/options"
	"bscp.io/cmd/api-server/service"
	"bscp.io/pkg/cc"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/metrics"
	"bscp.io/pkg/runtime/shutdown"
	"bscp.io/pkg/serviced"
	"bscp.io/pkg/tools"
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
	addr := net.JoinHostPort(network.BindIP, strconv.Itoa(int(network.HttpPort)))

	handler, err := as.service.Handler()
	if err != nil {
		return err
	}
	as.serve = &http.Server{Addr: addr, Handler: handler}

	go func() {
		notifier := shutdown.AddNotifier()
		select {
		case <-notifier.Signal:
			logs.Infof("start shutdown api server http server gracefully...")

			as.serve.Close()
			notifier.Done()

			logs.Infof("shutdown api server http server success...")
		}
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
			if err := as.serve.ListenAndServeTLS("", ""); err != nil {
				logs.Errorf("https server listen and serve failed, err: %v", err)
				shutdown.SignalShutdownGracefully()
			}
		}()
	} else {
		go func() {
			if err := as.serve.ListenAndServe(); err != nil {
				logs.Errorf("http server listen and serve failed, err: %v", err)
				shutdown.SignalShutdownGracefully()
			}
		}()
	}
	logs.Infof("api server listen and serve success. addr=%s", addr)

	return nil
}

func (as *apiServer) finalizer() {
	return
}
