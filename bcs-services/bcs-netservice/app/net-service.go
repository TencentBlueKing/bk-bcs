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
 *
 */

package app

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/netservice"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/storage/etcd"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/storage/zookeeper"
)

// Run entry point for bcs-netservice
func Run(cfg *Config) error {
	var st storage.Storage
	var err error
	blog.Infof("store type is selected as '%s'", cfg.Store)
	switch cfg.Store {
	case "zookeeper":
		// check storage data
		if cfg.BCSZk == "" {
			return errors.Errorf("parameter backend host lost")
		}
		// start storage
		st, err = zookeeper.NewStorage(cfg.BCSZk)
	case "etcd":
		if cfg.Registry.Endpoints == "" {
			return errors.Errorf("registry param endpoints cannot be empty")
		}
		hosts := strings.Split(cfg.Registry.Endpoints, ",")
		blog.Infof("init etcd tls config for '%v' started", hosts)
		if cfg.Registry.CA != "" && cfg.Registry.Key != "" && cfg.Registry.Cert != "" {
			tlsConfig, err := ssl.ClientTslConfVerity(cfg.Registry.CA, cfg.Registry.Cert,
				cfg.Registry.Key, static.ServerCertPwd)
			if err != nil {
				return errors.Wrapf(err, "load server tls config failed")
			}
			blog.Infof("init etcd tls config success")
			cfg.Registry.TLSConfig = tlsConfig
		}
		st, err = etcd.NewStorage(hosts, cfg.Registry.TLSConfig)
	default:
		return errors.Errorf("unknown store type '%s'", cfg.Store)
	}
	if err != nil {
		return errors.Wrapf(err, "create storage failed")
	}
	blog.Infof("create storage success")

	// create netservice
	netSvr := netservice.NewNetService(cfg.Address, int(cfg.Port), int(cfg.MetricPort), st)
	if netSvr == nil {
		return fmt.Errorf("create net server failed")
	}
	blog.Infof("create storage with store '%s' success", cfg.Store)

	// pid
	if err := common.SavePid(cfg.ProcessConfig); err != nil {
		blog.Errorf("fail to save pid: err: %s", err.Error())
	}

	// start signal handler
	interrupt := make(chan os.Signal, 10)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	// start http(s) service
	httpSrv := api.NewHTTPService(cfg.Address, int(cfg.Port))
	go handleSysSignal(interrupt, httpSrv, st)

	api.RegisterPoolHandler(httpSrv, netSvr)
	api.RegisterHostHandler(httpSrv, netSvr)
	api.RegisterAllocator(httpSrv, netSvr)
	api.RegisterResourceHandler(httpSrv, netSvr)
	api.RegisterIPInstanceHandler(httpSrv, netSvr)

	// prometheus metrics
	metricsAddr := cfg.Address + ":" + strconv.Itoa(int(cfg.MetricPort))
	api.RegisterMetrics(metricsAddr)

	// netservice http server
	if cfg.ServerKeyFile == "" {
		return httpSrv.ListenAndServe()
	}
	return httpSrv.ListenAndServeTLS(cfg.CAFile, cfg.ServerKeyFile, cfg.ServerCertFile, static.ServerCertPwd)
}

func handleSysSignal(signalChan <-chan os.Signal, htp *api.HTTPService, st storage.Storage) {
	<-signalChan
	blog.Info("Get signal from system. bcs-netservice was killed, ready to Stop\n")
	st.Stop()
	htp.Stop(5)
}
