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

	"bscp.io/cmd/sidecar/options"
	"bscp.io/cmd/sidecar/scheduler"
	"bscp.io/cmd/sidecar/service"
	"bscp.io/cmd/sidecar/stream"
	"bscp.io/cmd/sidecar/stream/types"
	"bscp.io/pkg/cc"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/metrics"
	"bscp.io/pkg/runtime/ctl"
	"bscp.io/pkg/runtime/ctl/cmd"
	"bscp.io/pkg/runtime/shutdown"
	sfs "bscp.io/pkg/sf-share"
	"bscp.io/pkg/tools"
)

// Run start the sidecar
func Run(opt *options.Option) error {
	sc := new(sidecar)
	if err := sc.prepare(opt); err != nil {
		return err
	}

	fp, err := sfs.GetFingerPrint()
	if err != nil {
		return fmt.Errorf("get sidecar fingerprint failed, err: %v", err)
	}
	sc.fingerPrint = fp

	logs.Infof("sidecar fingerprint: %s", sc.fingerPrint.Encode())

	upstream, err := stream.New(cc.Sidecar(), sc.fingerPrint)
	if err != nil {
		return fmt.Errorf("initialize upstream failed, err: %v", err)
	}
	sc.stream = upstream

	runtimeOpt, err := sc.initializeSidecar()
	if err != nil {
		return fmt.Errorf("initialize sidecar failed, err: %v", err)
	}
	sc.runtimeOption = runtimeOpt

	rs := map[string]interface{}{
		"setting":    cc.Sidecar(),
		"runtimeOpt": runtimeOpt,
	}
	if err := ctl.LoadCtl(cmd.WithLog(), cmd.WithQueryRuntimeSetting(rs), cmd.WithNotifyReconnect(
		upstream.NotifyReconnect)); err != nil {

		return fmt.Errorf("load control tool failed, err: %v", err)
	}

	opts := &scheduler.SchOptions{
		Settings:      cc.Sidecar(),
		RepositoryTLS: runtimeOpt.RepositoryTLS,
		AppReloads:    runtimeOpt.AppReloads,
		Stream:        upstream,
	}
	sch, err := scheduler.InitScheduler(opts)
	if err != nil {
		return fmt.Errorf("initialize scheduler failed, err: %v", err)
	}
	sc.sch = sch

	changeOpt := &types.OnChange{
		OnReleaseChange: sch.OnAppReleaseChange,
		CurrentRelease:  sch.CurrentRelease,
	}
	if err := sc.stream.StartWatch(changeOpt); err != nil {
		return fmt.Errorf("start scheduler watch failed, err: %v", err)
	}

	if err := sc.listenAndServe(); err != nil {
		return err
	}

	shutdown.RegisterFirstShutdown(sc.finalizer)
	shutdown.WaitShutdown(cc.Sidecar().Network.ShutdownTimeoutSec)
	return nil
}

type sidecar struct {
	serve         *http.Server
	service       *service.Service
	stream        stream.Interface
	sch           scheduler.Interface
	runtimeOption *sfs.SidecarRuntimeOption
	fingerPrint   sfs.FingerPrint
}

// prepare do prepare jobs before run sidecar.
func (s *sidecar) prepare(opt *options.Option) error {
	// load settings from config file.
	if err := cc.LoadSettings(opt.Sys); err != nil {
		return fmt.Errorf("load settings from config files failed, err: %v", err)
	}

	logs.InitLogger(cc.Sidecar().Log.Logs())

	metrics.InitMetrics(net.JoinHostPort(cc.Sidecar().Network.BindIP,
		strconv.Itoa(int(cc.Sidecar().Network.HttpPort))))

	svc, err := service.InitService()
	if err != nil {
		return fmt.Errorf("initial service failed, err: %v", err)
	}

	s.service = svc
	logs.Infof("load settings from config file success.")

	return nil
}

func (s *sidecar) finalizer() {
	list := make([]sfs.AppMeta, len(cc.Sidecar().AppSpec.Applications))
	for index, one := range cc.Sidecar().AppSpec.Applications {
		list[index] = sfs.AppMeta{
			AppID:     one.AppID,
			Namespace: one.Namespace,
			Uid:       one.Uid,
			Labels:    one.Labels,
		}
	}

	payload := &sfs.OfflinePayload{
		Applications: list,
	}

	if err := s.stream.FireEvent(payload); err != nil {
		logs.Errorf("fire offline event to upstream server failed, err: %v", err)
		return
	}

	logs.Infof("shutdown service success.")
}

// listenAndServe listen the http serve and set up the shutdown gracefully job.
func (s *sidecar) listenAndServe() error {
	network := cc.Sidecar().Network
	addr := net.JoinHostPort(network.BindIP, strconv.Itoa(int(network.HttpPort)))

	s.serve = &http.Server{
		Addr:    addr,
		Handler: s.service.Handler(),
	}

	go func() {
		notifier := shutdown.AddNotifier()
		select {
		case <-notifier.Signal:
			logs.Infof("start shutdown sidecar http server gracefully...")

			if err := s.serve.Close(); err != nil {
				logs.Errorf("close http server failed, err: %v", err)
			}

			notifier.Done()

			logs.Infof("shutdown sidecar http server success...")
		}
	}()

	if network.TLS.Enable() {
		tls := network.TLS
		tlsC, err := tools.ClientTLSConfVerify(tls.InsecureSkipVerify, tls.CAFile, tls.CertFile, tls.KeyFile,
			tls.Password)
		if err != nil {
			return fmt.Errorf("init tls config failed, err: %v", err)
		}

		s.serve.TLSConfig = tlsC
		go func() {
			if err := s.serve.ListenAndServeTLS("", ""); err != nil {
				logs.Errorf("https server listen and serve failed, err: %v", err)
				shutdown.SignalShutdownGracefully()
			}
		}()

	} else {
		go func() {
			if err := s.serve.ListenAndServe(); err != nil {
				logs.Errorf("http server listen and serve failed, err: %v", err)
				shutdown.SignalShutdownGracefully()
			}
		}()
	}

	logs.Infof("listen restful server at %s now.", addr)

	return nil
}
