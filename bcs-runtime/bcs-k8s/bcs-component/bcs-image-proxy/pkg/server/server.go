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

// Package server xxx
package server

import (
	"context"
	"crypto/tls"
	syserrors "errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/cleaner"

	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/internal/logctx"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/options"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/apiclient"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/bittorrent"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/recorder"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/server/ociscan"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/server/proxy"
	sserver "github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/server/server"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/server/staticwatcher"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/store"
)

// ImageProxyServer defines the server of image-proxy
type ImageProxyServer struct {
	op *options.ImageProxyOption

	globalCtx    context.Context
	globalCancel context.CancelFunc

	routerCustomAPI *mux.Router
	routerPprof     *mux.Router
	httpServer      *http.Server
	httpSServer     *http.Server
	scanHandler     *ociscan.ScanHandler
	customAPI       *sserver.CustomRegistry
	torrentHandler  *bittorrent.TorrentHandler
	staticWatcher   *staticwatcher.StaticFilesWatcher
}

// NewImageProxyServer create image-proxy server instance
func NewImageProxyServer() *ImageProxyServer {
	ctx, cancel := context.WithCancel(context.Background())
	s := &ImageProxyServer{
		op:              options.GlobalOptions(),
		routerCustomAPI: mux.NewRouter(),
		routerPprof:     mux.NewRouter(),
		scanHandler:     ociscan.NewScanHandler(),
		staticWatcher:   staticwatcher.NewStaticFileWatcher(),
		globalCtx:       ctx,
		globalCancel:    cancel,
	}
	return s
}

// Init the image proxy server
func (s *ImageProxyServer) Init() error {
	s.torrentHandler = bittorrent.NewTorrentHandler()
	if err := s.torrentHandler.Init(); err != nil {
		return err
	}
	if err := s.staticWatcher.Init(context.Background()); err != nil {
		return err
	}
	s.customAPI = sserver.NewCustomRegistry(s.torrentHandler, s.scanHandler)
	if err := s.scanHandler.Init(); err != nil {
		return errors.Wrapf(err, "failed to init oci scanhandler")
	}
	if err := cleaner.GlobalCleaner().Init(); err != nil {
		return err
	}
	s.routerPprof.Path("/debug/pprof/cmdline").HandlerFunc(pprof.Cmdline)
	s.routerPprof.Path("/debug/pprof/profile").HandlerFunc(pprof.Profile)
	s.routerPprof.Path("/debug/pprof/symbol").HandlerFunc(pprof.Symbol)
	s.routerPprof.Path("/debug/pprof/trace").HandlerFunc(pprof.Trace)
	s.routerPprof.PathPrefix("/debug/pprof/").HandlerFunc(pprof.Index)

	s.routerCustomAPI.Path(apiclient.CustomAPIGetManifest).HandlerFunc(
		s.customAPI.HTTPWrapper(s.customAPI.RegistryGetManifest))
	s.routerCustomAPI.Path(apiclient.CustomAPIGetLayerInfo).HandlerFunc(
		s.customAPI.HTTPWrapper(s.customAPI.RegistryGetLayerInfo))
	s.routerCustomAPI.Path(apiclient.CustomAPICheckStaticLayer).HandlerFunc(
		s.customAPI.HTTPWrapper(s.customAPI.CheckStaticLayer))
	s.routerCustomAPI.Path(apiclient.CustomAPICheckOCILayer).HandlerFunc(
		s.customAPI.HTTPWrapper(s.customAPI.CheckOCILayer))
	s.routerCustomAPI.Path(apiclient.CustomAPIDownloadLayerFromMaster).HandlerFunc(
		s.customAPI.HTTPWrapper(s.customAPI.DownloadLayerFromMaster))
	s.routerCustomAPI.Path(apiclient.CustomAPIDownloadLayerFromNode).HandlerFunc(
		s.customAPI.HTTPWrapper(s.customAPI.DownloadLayerFromNode))
	s.routerCustomAPI.Path(apiclient.CustomAPITransferLayerTCP).HandlerFunc(
		s.customAPI.HTTPWrapper(s.customAPI.TransferLayerTCP))
	s.routerCustomAPI.Path(apiclient.CustomAPIRecorder).HandlerFunc(
		s.customAPI.HTTPWrapper(s.customAPI.Recorder))
	s.routerCustomAPI.Path(apiclient.CustomAPITorrentStatus).HandlerFunc(
		s.customAPI.HTTPWrapper(s.customAPI.TorrentStatus))
	return nil
}

// Run the image proxy server
func (s *ImageProxyServer) Run() error {
	defer func() {
		if err := store.GlobalRedisStore().CleanHostCache(context.Background()); err != nil {
			blog.Errorf("clean host cache failed, err %s", err.Error())
		}
	}()
	fs := []func(errCh chan error){s.runHTTPServer, s.runHTTPSServer, s.runRecorder, s.runCleaner,
		s.runOCITickReporter, s.runTorrentTickReporter, s.runStaticFilesWatcher}
	errCh := make(chan error, len(fs))
	for i := range fs {
		go fs[i](errCh)
	}
	// for-loop wait every goroutine normal finish
	for i := 0; i < len(fs); i++ {
		e := <-errCh
		// we should return error if e not nil, perhaps some goroutines are
		// exited with error. So we need exit the server
		if e != nil {
			return errors.Wrapf(e, "run server failed")
		}
	}
	return nil
}

func (s *ImageProxyServer) runHTTPServer(errCh chan error) {
	defer blog.Warnf("http server exit")
	serverAddr := fmt.Sprintf("0.0.0.0:%d", s.op.HTTPPort)
	s.httpServer = &http.Server{
		Addr:    serverAddr,
		Handler: s,
	}
	if err := s.httpServer.ListenAndServe(); err != nil && !syserrors.Is(err, http.ErrServerClosed) {
		errCh <- err
		blog.Errorf("failed to start http server: %s", err.Error())
		return
	}
	errCh <- nil
}

func (s *ImageProxyServer) runHTTPSServer(errCh chan error) {
	defer blog.Warnf("http(s) server exit")
	serverAddr := fmt.Sprintf("0.0.0.0:%d", s.op.HTTPSPort)
	tlsCerts := make([]tls.Certificate, 0)
	defaultCert := s.op.ExternalConfig.BuiltInCerts[options.LocalhostCert]
	if defaultCert == nil {
		errCh <- fmt.Errorf("not have default 'localhost' tls cert")
		return
	}
	defaultKeyPair, err := tls.X509KeyPair([]byte(defaultCert.Cert), []byte(defaultCert.Key))
	if err != nil {
		errCh <- fmt.Errorf("generate tls cert for default failed: %s", err.Error())
		return
	}
	tlsCerts = append(tlsCerts, defaultKeyPair)
	for _, mp := range s.op.ExternalConfig.RegistryMappings {
		if mp.ProxyCert == "" || mp.ProxyKey == "" {
			continue
		}
		var kp tls.Certificate
		kp, err = tls.X509KeyPair([]byte(mp.ProxyCert), []byte(mp.ProxyKey))
		if err != nil {
			errCh <- fmt.Errorf("generate tls cert for '%s' failed: %s", mp.ProxyHost, err.Error())
			return
		}
		tlsCerts = append(tlsCerts, kp)
	}
	s.httpSServer = &http.Server{
		Addr:    serverAddr,
		Handler: s,
		TLSConfig: &tls.Config{
			Certificates: tlsCerts,
		},
	}
	if err = s.httpSServer.ListenAndServeTLS("", ""); err != nil && !syserrors.Is(err,
		http.ErrServerClosed) {
		errCh <- err
		blog.Errorf("failed to start http(s) server: %s", err.Error())
		return
	}
	errCh <- nil
}

func (s *ImageProxyServer) runRecorder(errCh chan error) {
	defer blog.Warnf("audit recorder exit")
	if err := recorder.GlobalRecorder().Run(s.globalCtx); err != nil {
		blog.Errorf("audit recorder exit with err: %s", err.Error())
		errCh <- err
		return
	}
	errCh <- nil
}

func (s *ImageProxyServer) runCleaner(errCh chan error) {
	if s.op.CleanConfig.Cron == "" {
		return
	}
	defer blog.Warnf("auto-cleaner exit")
	cleaner.GlobalCleaner().Run(s.globalCtx)
	errCh <- nil
}

func (s *ImageProxyServer) runOCITickReporter(errCh chan error) {
	defer blog.Warnf("oci tick reporter exit")
	s.scanHandler.TickerReport(s.globalCtx)
	errCh <- nil
}

func (s *ImageProxyServer) runTorrentTickReporter(errCh chan error) {
	defer blog.Warnf("torrent tick reporter exit")
	s.torrentHandler.TickReport(s.globalCtx)
	errCh <- nil
}

func (s *ImageProxyServer) runStaticFilesWatcher(errCh chan error) {
	defer blog.Warnf("static-files watcher exit")
	if err := s.staticWatcher.Watch(s.globalCtx); err != nil {
		blog.Errorf("static-files watcher exit with err: %s", err.Error())
		errCh <- err
		return
	}
	errCh <- nil
}

// Shutdown shutdown the image proxy server
func (s *ImageProxyServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.globalCancel()
	s.httpServer.Shutdown(ctx)
	s.httpSServer.Shutdown(ctx)
}

const (
	// LocalHost defines the localhost
	LocalHost = "localhost"
	// LocalHostAddr defines the localhost address
	LocalHostAddr = "127.0.0.1"
)

var (
	proxyHostRegex = regexp.MustCompile(`^/v[1-2]/([^/]+)/`)
)

func (s *ImageProxyServer) httpError(ctx context.Context, rw http.ResponseWriter, errMsg string, code int) {
	logctx.Errorf(ctx, "image-proxy server response error: %s", errMsg)
	http.Error(rw, errMsg, http.StatusBadRequest)
}

func (s *ImageProxyServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	req = apiclient.SetContext(req)
	ctx := req.Context()
	logctx.Infof(ctx, "received request: %s, %s%s", req.Method, req.Host, req.URL.String())
	switch {
	case strings.HasPrefix(req.RequestURI, "/debug/"):
		s.routerPprof.ServeHTTP(rw, req)
		return
	case strings.HasPrefix(req.RequestURI, "/custom_api/"):
		s.routerCustomAPI.ServeHTTP(rw, req)
		return
	}

	hosts := strings.Split(req.Host, ":")
	if len(hosts) != 2 {
		s.httpError(ctx, rw, fmt.Sprintf("invalid host: %s", req.Host), http.StatusBadRequest)
		return
	}
	var proxyHost string
	var proxyType options.ProxyType
	var requestURI = req.RequestURI
	switch hosts[0] {
	// 如果传递过来的 Host 是本地地址，则认为用户使用的是 RegistryMirror 模式
	case LocalHost, LocalHostAddr:
		proxyType = options.RegistryMirror

		queryNS := req.URL.Query().Get("ns")
		if queryNS != "" {
			// for containerd
			proxyHost = queryNS
		} else {
			match := proxyHostRegex.FindStringSubmatch(req.RequestURI)
			// 如果 match[1] 未从 options 中查找到对应配置，则认为其 proxyHost 就是空的
			if len(match) > 1 && s.op.FilterRegistryMapping(match[1], proxyType) != nil {
				proxyHost = match[1]
				requestURI = strings.Replace(req.RequestURI, "/"+proxyHost+"/", "/", 1)
			}
		}
	// 传递过来的 Host 是个域名地址，则认为用户使用的是域名代理模式
	default:
		proxyType = options.DomainProxy
		proxyHost = hosts[0]
	}
	// logctx.Infof(ctx, "parse request, proxyType: %s, proxyHost: %s", string(proxyType), proxyHost)

	upstreamProxy := proxy.NewUpstreamProxy(proxyHost, proxyType, s.torrentHandler)
	if upstreamProxy == nil {
		s.httpError(ctx, rw, fmt.Sprintf("no handler for proxy host '%s'", proxyHost), http.StatusBadRequest)
		return
	}
	upstreamProxy.ServeHTTP(requestURI, rw, req)
}
