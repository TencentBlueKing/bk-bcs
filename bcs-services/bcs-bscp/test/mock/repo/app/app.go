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
	"os"
	"strconv"
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/shutdown"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/mock/repo/options"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/mock/repo/service"
)

type repoMock struct {
	serve    *http.Server
	service  *service.Service
	settings *Setting
}

// Run start the api server
func Run(opt *options.Option) error {
	rp := new(repoMock)

	if err := rp.prepare(opt); err != nil {
		return err
	}

	if err := rp.listenAndServe(); err != nil {
		return err
	}

	shutdown.RegisterFirstShutdown(rp.finalizer)
	shutdown.WaitShutdown(20)
	return nil
}

func (rp *repoMock) prepare(opt *options.Option) error {
	settings, err := LoadSettings(opt.Sys)
	if err != nil {
		return err
	}
	rp.settings = settings

	logs.InitLogger(rp.settings.Log.Logs())

	logs.Infof("load settings from config file success.")

	// create workspace.
	if err := os.MkdirAll(rp.settings.Workspace.RootDirectory, os.ModePerm); err != nil {
		return fmt.Errorf("mkdir workspace directory failed, err: %v", err)
	}

	return nil
}

func (rp *repoMock) listenAndServe() error {

	svc, err := service.NewService(rp.settings.Workspace.RootDirectory)
	if err != nil {
		return err
	}

	rp.service = svc

	root := http.NewServeMux()
	root.HandleFunc("/", rp.service.Handler().ServeHTTP)

	network := rp.settings.Network
	addr := net.JoinHostPort(rp.settings.Network.BindIP, strconv.FormatUint(uint64(rp.settings.Network.Port), 10))

	rp.serve = &http.Server{
		Addr:    addr,
		Handler: root,
	}

	go func() {
		notifier := shutdown.AddNotifier()
		select {
		case <-notifier.Signal:
			defer notifier.Done()

			logs.Infof("start shutdown http server gracefully...")

			ctx, cancel := context.WithTimeout(context.TODO(), 20*time.Second)
			defer cancel()
			if err := rp.serve.Shutdown(ctx); err != nil {
				logs.Errorf("shutdown http server failed, err: %v", err)
				return
			}

			logs.Infof("shutdown http server success...")
		}
	}()

	if network.TLS.Enable() {
		tls := network.TLS
		tlsC, err := tools.ClientTLSConfVerify(tls.InsecureSkipVerify, tls.CAFile, tls.CertFile, tls.KeyFile,
			tls.Password)
		if err != nil {
			return fmt.Errorf("init tls config failed, err: %v", err)
		}

		rp.serve.TLSConfig = tlsC

		go func() {
			if err := rp.serve.ListenAndServeTLS("", ""); err != nil {
				logs.Errorf("https server listen and serve failed, err: %v", err)
				shutdown.SignalShutdownGracefully()
			}
		}()
	} else {
		go func() {
			if err := rp.serve.ListenAndServe(); err != nil {
				logs.Errorf("http server listen and serve failed, err: %v", err)
				shutdown.SignalShutdownGracefully()
			}
		}()
	}
	logs.Infof("api server listen and serve success.")

	return nil
}

func (rp *repoMock) finalizer() {
	return
}
