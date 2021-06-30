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

package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-apiserver-proxy/cmd/config"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-apiserver-proxy/pkg/endpoint"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-apiserver-proxy/pkg/service"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-apiserver-proxy/pkg/utils/sets"
)

var (
	// ErrProxyManagerNotInited show ProxyManager not inited
	ErrProxyManagerNotInited = errors.New("ProxyManager not inited")
)

// NewProxyManager set ProxyManager by opts
func NewProxyManager(opts *config.ProxyAPIServerOptions) (*ProxyManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	pm := &ProxyManager{
		ctx:    ctx,
		cancel: cancel,
		stop:   make(chan error),
	}

	isValidate := opts.Validate()
	if !isValidate {
		errMsg := fmt.Errorf("validate ProxyApiServerOptions failed")
		return nil, errMsg
	}
	pm.options = opts

	return pm, nil
}

// ProxyManager struct proxy cluster master endpointIPs By LVS
type ProxyManager struct {
	options            *config.ProxyAPIServerOptions
	clusterEndpointsIP endpoint.ClusterEndpointsIP
	lvsProxy           service.LvsProxy
	httpServer         *http.Server

	// http server quit
	stop   chan error
	ctx    context.Context
	cancel context.CancelFunc
}

// Init init proxyManager
func (pm *ProxyManager) Init(options *config.ProxyAPIServerOptions) error {
	if pm == nil {
		return ErrProxyManagerNotInited
	}

	pm.initProxyOptions(options)

	err := pm.initLvsProxy()
	if err != nil {
		return err
	}

	err = pm.initClusterEndpointsClient()
	if err != nil {
		return err
	}

	err = pm.initHTTPServer()
	if err != nil {
		return err
	}

	err = pm.savePID()
	if err != nil {
		return err
	}

	err = pm.waitQuitHandler()
	if err != nil {
		return err
	}

	return nil
}

// Run run ProxyManager business
func (pm *ProxyManager) Run() error {
	if pm == nil {
		return ErrProxyManagerNotInited
	}

	coldStart := make(chan struct{}, 1)
	coldStart <- struct{}{}

	ticker := time.NewTicker(time.Duration(pm.options.SystemInterval.ManagerInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
		case <-coldStart:
		case <-pm.ctx.Done():
			blog.Infof("proxyManager run quit(pm.ctx.Done): %v", pm.ctx.Err())
			return nil
		}

		func() {
			err := pm.checkVirtualServerIsExist()
			if err != nil {
				blog.Errorf("checkVirtualServerAndCreateVsWhenNotExist failed: %v", err)
				return
			}

			adds, deletes, err := pm.getAddOrDeleteRealServers()
			if err != nil {
				blog.Errorf("getAddOrDeleteRealServers failed: %v", err)
				return
			}
			if len(adds) == 0 && len(deletes) == 0 {
				blog.Infof("cluster master endpointIPs equal lvs backend realServers, no need to sync")
				return
			}

			pm.syncLvsRealServers(adds, deletes)
		}()
	}
}

func (pm *ProxyManager) syncLvsRealServers(adds, deletes sets.String) error {
	if pm == nil {
		return ErrProxyManagerNotInited
	}

	blog.V(5).Infof("syncLvsRealServers, adds: [%v] deletes: [%v]", adds, deletes)

	if len(adds) > 0 {
		for s := range adds {
			err := pm.lvsProxy.CreateRealServer(s)
			if err != nil {
				blog.Errorf("syncLvsRealServers CreateRealServer[%s] failed: %v", s, err)
				continue
			}

			blog.Infof("syncLvsRealServers CreateRealServer[%s] successful", s)
		}
	}

	if len(deletes) > 0 {
		for s := range deletes {
			err := pm.lvsProxy.DeleteRealServer(s)
			if err != nil {
				blog.Errorf("syncLvsRealServers DeleteRealServer[%s] failed: %v", s, err)
				continue
			}

			blog.Infof("syncLvsRealServers DeleteRealServer[%s] successful", s)
		}
	}

	blog.V(5).Infof("syncLvsRealServers, adds: [%v] deletes: [%v] successful", adds, deletes)

	return nil
}

func (pm *ProxyManager) getAddOrDeleteRealServers() (sets.String, sets.String, error) {
	if pm == nil {
		return nil, nil, ErrProxyManagerNotInited
	}

	var (
		addServers, deleteServers sets.String
	)

	// get cluster master endpoint IPs
	clusterEndpoints, err := pm.clusterEndpointsIP.GetClusterEndpoints()
	if err != nil {
		return nil, nil, err
	}
	clusterRs := []string{}
	for _, ep := range clusterEndpoints {
		clusterRs = append(clusterRs, ep.String())
	}
	clusterRsMap := sets.NewString(clusterRs...)

	// get proxy lvs endpoint real server
	proxyRs, err := pm.lvsProxy.ListRealServer()
	if err != nil {
		return nil, nil, err
	}
	proxyRsMap := sets.NewString(proxyRs...)

	// diff get add & delete server
	addServers = clusterRsMap.Difference(proxyRsMap)
	deleteServers = proxyRsMap.Difference(clusterRsMap)

	return addServers, deleteServers, nil
}

func (pm *ProxyManager) initProxyOptions(options *config.ProxyAPIServerOptions) {
	if pm == nil {
		blog.Errorf("server failed:%v", ErrProxyManagerNotInited)
	}

	pm.options = options
}

func (pm *ProxyManager) checkVirtualServerIsExist() error {
	if pm == nil {
		return ErrProxyManagerNotInited
	}

	available := pm.lvsProxy.IsVirtualServerAvailable(pm.options.ProxyLvs.VirtualAddress)
	if !available {
		err := pm.lvsProxy.CreateVirtualServer(pm.options.ProxyLvs.VirtualAddress)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pm *ProxyManager) initLvsProxy() error {
	if pm == nil {
		return ErrProxyManagerNotInited
	}

	lvsProxy := service.NewLvsProxy()
	pm.lvsProxy = lvsProxy

	// exist lvs
	available := lvsProxy.IsVirtualServerAvailable(pm.options.ProxyLvs.VirtualAddress)
	if available {
		blog.Infof("VirtualServerAvailable %s is available", pm.options.ProxyLvs.VirtualAddress)

		rsServers, err := lvsProxy.ListRealServer()
		if err != nil {
			blog.Infof("VirtualServerAvailable ListRealServer failed: %v", err)
			return err
		}

		for i := range rsServers {
			err = lvsProxy.CreateRealServer(rsServers[i])
			if err != nil {
				blog.Errorf("lvsProxy CreateRealServer[%s] failed: %v", rsServers[i], err)
			}
		}

		return nil
	}

	err := lvsProxy.CreateVirtualServer(pm.options.ProxyLvs.VirtualAddress)
	if err != nil {
		return err
	}

	return nil
}

func (pm *ProxyManager) initClusterEndpointsClient() error {
	if pm == nil {
		return ErrProxyManagerNotInited
	}

	opts := []endpoint.EndpointsClientOption{}
	if pm.options.HealthCheck.HealthScheme != "" && pm.options.HealthCheck.HealthPath != "" {
		opts = append(opts, endpoint.WithHealthConfig(endpoint.EndpointsHealthOptions{
			Scheme: pm.options.HealthCheck.HealthScheme,
			Path:   pm.options.HealthCheck.HealthPath,
		}))
	}

	opts = append(opts, endpoint.WithK8sConfig(endpoint.K8sConfig{
		Mater:      pm.options.K8sConfig.Master,
		KubeConfig: pm.options.K8sConfig.KubeConfig,
	}))

	if pm.options.SystemInterval.EndpointInterval > 0 {
		opts = append(opts, endpoint.WithInterval(time.Second*time.Duration(pm.options.SystemInterval.EndpointInterval)))
	}

	endpointClient, err := endpoint.NewEndpointsClient(opts...)
	if err != nil {
		return err
	}

	pm.clusterEndpointsIP = endpointClient
	go pm.clusterEndpointsIP.SyncClusterEndpoints(pm.ctx)

	return nil
}

// init prometheus metrics handler
func (pm *ProxyManager) initMetrics(router *mux.Router) {
	blog.Infof("init metrics handler")
	router.Handle("/metrics", promhttp.Handler())
}

// init pprof handler
func (pm *ProxyManager) initPProf(router *mux.Router) {
	if pm == nil {
		return
	}

	if !pm.options.DebugMode {
		blog.Infof("pprof debugMode is off")
		return
	}

	blog.Infof("pprof debugMode is on")

	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

// init extra http server(metrics, serverSwagger, pprof)
func (pm *ProxyManager) initHTTPServer() error {
	if pm == nil {
		return ErrProxyManagerNotInited
	}

	router := mux.NewRouter()
	pm.initMetrics(router)
	pm.initPProf(router)

	mux := http.NewServeMux()
	mux.Handle("/", router)

	httpAddress := pm.options.ServiceConfig.Address + ":" + strconv.Itoa(int(pm.options.Port))
	pm.httpServer = &http.Server{
		Addr:    httpAddress,
		Handler: mux,
	}

	go func() {
		var err error
		blog.Infof("initHttpServer address: %s", httpAddress)

		err = pm.httpServer.ListenAndServe()
		if err != nil {
			blog.Errorf("initHttpServer failed: %v", err)
			pm.stop <- err
		}
	}()

	return nil
}

func (pm *ProxyManager) waitQuitHandler() error {
	if pm == nil {
		return ErrProxyManagerNotInited
	}

	quitSignal := make(chan os.Signal, 10)
	signal.Notify(quitSignal, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)

	go func() {
		select {
		case e := <-quitSignal:
			blog.Infof("reveice interrupt signal: %s", e.String())
			pm.close()
		case <-pm.stop:
			blog.Infof("http server quit")
			pm.close()
		}
	}()

	return nil
}

// close proxyManager
func (pm *ProxyManager) close() {
	if pm == nil {
		return
	}

	pm.clusterEndpointsIP.Stop()
	pm.lvsProxy.DeleteVirtualServer(pm.options.ProxyLvs.VirtualAddress)
	pm.cancel()
}

func (pm *ProxyManager) savePID() error {
	if pm == nil {
		return ErrProxyManagerNotInited
	}

	err := common.SavePid(pm.options.ProcessConfig)
	if err != nil {
		blog.Errorf("proxyManager save pid failed: %v", err)
	}

	return nil
}
