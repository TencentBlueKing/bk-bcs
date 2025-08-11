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

// Package proxy xxx
package proxy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/internal/lock"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/internal/logctx"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/options"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/apiclient"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/bittorrent"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/recorder"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/server/proxy/transport"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/server/requester"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/utils"
)

var (
	createLock sync.Mutex
	proxies    = make(map[string]*upstreamProxy)
)

// UpstreamProxyInterface defines the interface of upstream
type UpstreamProxyInterface interface {
	ServeHTTP(requestURI string, rw http.ResponseWriter, req *http.Request)
}

type upstreamProxy struct {
	layerLock lock.Interface

	op            *options.ImageProxyOption
	proxyHost     string
	proxyType     options.ProxyType
	proxyRegistry *options.RegistryMapping

	reverseProxy   *httputil.ReverseProxy
	torrentHandler *bittorrent.TorrentHandler
	cacheStore     store.CacheStore

	event *recorder.EventRecorder
}

// NewUpstreamProxy create the upstream proxy instance
func NewUpstreamProxy(proxyHost string, proxyType options.ProxyType,
	torrentHandler *bittorrent.TorrentHandler) UpstreamProxyInterface {
	createLock.Lock()
	defer createLock.Unlock()

	op := options.GlobalOptions()
	p, ok := proxies[proxyHost]
	if ok {
		p.proxyRegistry = op.FilterRegistryMapping(proxyHost, proxyType)
		return p
	}
	proxyRegistry := op.FilterRegistryMapping(proxyHost, proxyType)
	if proxyRegistry == nil {
		return nil
	}
	p = &upstreamProxy{
		op:             op,
		layerLock:      lock.NewLocalLock(),
		proxyHost:      proxyHost,
		proxyType:      proxyType,
		proxyRegistry:  op.FilterRegistryMapping(proxyHost, proxyType),
		torrentHandler: torrentHandler,
		cacheStore:     store.GlobalRedisStore(),
		event:          recorder.GlobalRecorder(),
	}
	p.initReverseProxy()
	proxies[proxyHost] = p
	return p
}

// initReverseProxy will reverse the request to original registry host
func (p *upstreamProxy) initReverseProxy() {
	p.reverseProxy = &httputil.ReverseProxy{
		Director: func(request *http.Request) {},
		ErrorHandler: func(writer http.ResponseWriter, req *http.Request, err error) {
			logctx.Errorf(req.Context(), "reverse proxy '%s' failed: %s. header: %+v", req.URL.String(),
				err.Error(), req.Header)
			p.event.SendObjEvent(req.Context(), recorder.Warning, fmt.Sprintf("Reverse request '%s' "+
				"failed: %s", req.URL.String(), err.Error()))
		},
		Transport: transport.DefaultProxyTransport(p.proxyRegistry),
		ModifyResponse: func(resp *http.Response) error {
			req := resp.Request
			logctx.Infof(req.Context(), "reverse proxy to '%s' response code '%d'", req.URL.String(),
				resp.StatusCode)
			if resp.StatusCode >= 400 {
				p.event.SendObjEvent(req.Context(), recorder.Warning, fmt.Sprintf("Reverse request '%s' resp "+
					"code '%d' >= 400 , header: %+v", req.URL.String(), resp.StatusCode, req.Header))
			} else {
				p.event.SendObjEvent(req.Context(), recorder.Warning, fmt.Sprintf("Reverse request '%s' resp "+
					"code '%d' success", req.URL.String(), resp.StatusCode))
			}
			return nil
		},
	}
}

func (p *upstreamProxy) httpError(ctx context.Context, rw http.ResponseWriter, errMsg string, code int) {
	logctx.Errorf(ctx, "upstream-proxy response error: %s", errMsg)
	http.Error(rw, errMsg, http.StatusBadRequest)
}

// ServeHTTP handle the request of upstream. Requests are divided into three categories: Auth/GetManifest/DownloadLayer.
// The function will handle the three requests.
func (p *upstreamProxy) ServeHTTP(requestURI string, rw http.ResponseWriter, req *http.Request) {
	originalHost := p.proxyRegistry.OriginalHost
	ctx := logctx.SetProxyHost(req.Context(), originalHost)

	fullPath := fmt.Sprintf("https://%s%s", originalHost, requestURI)
	newURL, err := url.Parse(fullPath)
	if err != nil {
		p.httpError(ctx, rw, fmt.Sprintf("build new full path '%s' failed: %s", fullPath, err.Error()),
			http.StatusBadRequest)
		return
	}
	req.URL = newURL
	req.Host = originalHost

	manifestRepo, manifestTag, isManifest := utils.IsManifestGet(req)
	blobRepo, digest, isBlob := utils.IsBlobGet(req.URL.Path)
	switch {
	case isManifest:
		ctx = p.event.SetManifestRequest(ctx, originalHost, manifestRepo, manifestTag)
		if err = p.handleManifestGetRequest(ctx, req, rw); err == nil {
			return
		}
		// get manifest from master failed, so we need reverse the request to original registry
		logctx.Warnf(ctx, "handle get-manifest request failed and will reverse: %s", err.Error())
	case isBlob:
		ctx = p.event.SetLayerRequest(ctx, originalHost, blobRepo, digest)
		var canReverse bool
		canReverse, err = p.handleBlobGetRequest(ctx, req, rw)
		if err == nil {
			p.event.SendObjEvent(ctx, recorder.Normal, fmt.Sprintf("Download layer '%s' success", digest))
			return
		}
		logctx.Errorf(ctx, "handle get-blob request failed: %s", err.Error())
		p.event.SendObjEvent(ctx, recorder.Warning, fmt.Sprintf("Download layer '%s' failed: %s "+
			"(canReverse=%v)", digest, err.Error(), canReverse))
		if !canReverse {
			logctx.Errorf(ctx, "handle get-blob cannot reverse, directly return error")
			return
		}
	default:
		logctx.Infof(ctx, "handling request not manifest and blob")
	}
	// even though we intercept the request and something goes wrong,
	// we can still go back to the source registry
	if isManifest || isBlob {
		p.event.SendObjEvent(ctx, recorder.Normal, fmt.Sprintf("Start reverse request: %s", newURL))
	}
	req = req.WithContext(ctx)
	p.reverseProxy.ServeHTTP(rw, req)
}

// handleManifestGetRequest handle the get-manifest request. It will request the manifest from master-node
func (p *upstreamProxy) handleManifestGetRequest(ctx context.Context, req *http.Request, rw http.ResponseWriter) error {
	repo, tag, isManifest := utils.IsManifestGet(req)
	if !isManifest {
		return nil
	}

	p.event.SendObjEvent(ctx, recorder.Normal, fmt.Sprintf("Starting get manifest with tag '%s'", tag))
	logctx.Infof(ctx, "handling get-manifest request")
	getManifestReq := &apiclient.GetManifestRequest{
		OriginalHost: req.Host,
		ManifestUrl:  req.URL.RequestURI(),
		Repo:         repo,
		Tag:          tag,
		BearerToken:  getBearerToken(req),
	}
	master, manifest, err := requester.GetManifest(ctx, getManifestReq)
	if err != nil {
		p.event.SendObjEvent(ctx, recorder.Warning,
			fmt.Sprintf("Get manifest with tag '%s' from master '%s' failed: %s", tag, master, err.Error()))
		return err
	}
	p.event.SendObjEvent(ctx, recorder.Normal,
		fmt.Sprintf("Get manifest with tag '%s' from master '%s' success", tag, master))
	logctx.Infof(ctx, "get manifest from master success")
	rw.Header().Add("Content-Type", "application/json")
	_, _ = rw.Write([]byte(manifest))
	return nil
}

// handleBlobGetRequest handle the download-layer request. It will request the master-node get the layer info, and
// then download the layer. Perhaps download by Torrent or TCP
func (p *upstreamProxy) handleBlobGetRequest(ctx context.Context, req *http.Request,
	rw http.ResponseWriter) (bool, error) {
	_, digest, isBlob := utils.IsBlobGet(req.URL.Path)
	if !isBlob {
		return false, nil
	}

	p.event.SendObjEvent(ctx, recorder.Normal, fmt.Sprintf("Received download layer '%s' request", digest))
	p.layerLock.Lock(ctx, digest)
	defer p.layerLock.UnLock(ctx, digest)
	logctx.Infof(ctx, "handling get-blob request")
	// download layer if local existed
	if p.downloadLayerFromLocal(ctx, rw) {
		return true, nil
	}

	// get layer from master
	logctx.Infof(ctx, "download layer from master")
	layerReq := &apiclient.CommonDownloadLayerRequest{
		OriginalHost: req.Host,
		LayerUrl:     req.URL.RequestURI(),
		BearerToken:  getBearerToken(req),
	}
	layerResp, master, err := requester.DownloadLayerFromMaster(ctx, layerReq, digest)
	if err != nil {
		p.event.SendObjEvent(ctx, recorder.Warning, fmt.Sprintf("Get layer-info from master '%s' failed: %s",
			master, err.Error()))
		return true, errors.Wrapf(err, "download layer from master failed")
	}
	logctx.Infof(ctx, "get layer-info from master success, located: %s, filePath: %s, size: %s",
		layerResp.Located, layerResp.FilePath, humanize.Bytes(uint64(layerResp.FileSize)))
	haveTorrent := "no-torrent"
	if layerResp.TorrentBase64 != "" {
		haveTorrent = "(too long)"
	}
	p.event.SendObjEvent(ctx, recorder.Normal, fmt.Sprintf("Get layer-info from master '%s' success, "+
		"located: %s, filePath: %s, size: %s, torrent: %s", master, layerResp.Located, layerResp.FilePath,
		humanize.Bytes(uint64(layerResp.FileSize)), haveTorrent))
	// Should download layer from local again, maybe already have it on local
	// Because when we download the layer from the master, the master may assign the task of downloading the
	// layer to us. When we get the layer information, the layer may have been downloaded to the current node.
	if p.downloadLayerFromLocal(ctx, rw) {
		return true, nil
	}

	var canReverse bool
	if canReverse, err = p.handleLayerDownload(ctx, rw, layerResp, digest); err != nil {
		return canReverse, errors.Wrapf(err, "handle download layer failed")
	}
	return true, nil
}

// downloadLayerFromLocal download layer from local, if local have the layer
func (p *upstreamProxy) downloadLayerFromLocal(ctx context.Context, rw http.ResponseWriter) bool {
	digest := logctx.GetLayerDigest(ctx)
	layerFileInfo, layerPath := p.checkLocalLayer(digest)
	if layerFileInfo == nil {
		return false
	}
	fi, err := os.Open(layerPath)
	if err != nil {
		logctx.Warnf(ctx, "open local transfer layer '%s' failed: %s", layerPath, err.Error())
		return false
	}
	defer fi.Close()
	logctx.Infof(ctx, "download layer from local starting")
	p.event.SendObjEvent(ctx, recorder.Normal, fmt.Sprintf("Download layer from local '%s'", layerPath))
	if _, err = io.Copy(rw, fi); err != nil {
		logctx.Errorf(ctx, "io copy layer failed: %s", err.Error())
		p.event.SendObjEvent(ctx, recorder.Warning, fmt.
			Sprintf("Copy local '%s' failed: %s", layerPath, err.Error()))
		return false
	}
	logctx.Infof(ctx, "download layer from local success. Content-Length: %d", layerFileInfo.Size())
	p.event.SendObjEvent(ctx, recorder.Normal, fmt.Sprintf("Download layer from local '%s' success",
		layerPath))
	return true
}

func (p *upstreamProxy) checkLocalLayer(digest string) (os.FileInfo, string) {
	layerName := utils.LayerFileName(digest)
	localLayer := path.Join(p.op.TransferPath, layerName)
	fi, err := os.Stat(localLayer)
	if err == nil {
		return fi, localLayer
	}
	localLayer = path.Join(p.op.SmallFilePath, layerName)
	fi, err = os.Stat(localLayer)
	if err == nil {
		return fi, localLayer
	}
	localLayer = path.Join(p.op.OCIPath, layerName)
	fi, err = os.Stat(localLayer)
	if err == nil {
		return fi, localLayer
	}
	return nil, ""
}

func (p *upstreamProxy) handleLayerDownload(ctx context.Context, rw http.ResponseWriter,
	resp *apiclient.CommonDownloadLayerResponse, digest string) (bool, error) {
	// download layer from target directly with tcp
	if resp.TorrentBase64 == "" {
		p.event.SendObjEvent(ctx, recorder.Normal, fmt.Sprintf("Download-by-tcp from '%s' "+
			"with file '%s' started", resp.Located, resp.FilePath))
		if err := p.downloadByTCP(ctx, rw, resp.Located, 0, resp.FilePath); err != nil {
			p.event.SendObjEvent(ctx, recorder.Warning, fmt.Sprintf("Download-by-tcp from '%s "+
				"with file '%s' failed: %s", resp.Located, resp.FilePath, err.Error()))
			return true, errors.Wrapf(err, "download by tcp failed")
		}
		p.event.SendObjEvent(ctx, recorder.Normal, fmt.Sprintf("Download-by-tcp from '%s' "+
			"with file '%s' success", resp.Located, resp.FilePath))
		return true, nil
	}

	logctx.Infof(ctx, "download layer with torrent is starting")
	p.event.SendObjEvent(ctx, recorder.Normal, fmt.Sprintf("Download-by-torrent '%s' from '%s' started",
		digest, resp.Located))
	remedy, transmitted, err := p.torrentHandler.DownloadTorrent(ctx, rw, resp.Located, digest, resp.TorrentBase64)
	if err == nil {
		logctx.Infof(ctx, "layer download-by-torrent rewrite to http.writer success")
		p.event.SendObjEvent(ctx, recorder.Normal, fmt.Sprintf("Download-by-torrent '%s' from '%s' success",
			digest, resp.Located))
		return true, nil
	}
	p.event.SendObjEvent(ctx, recorder.Warning, fmt.Sprintf("Download-by-torrent '%s' from '%s' failed: %s",
		digest, resp.Located, err.Error()))
	if !remedy {
		return false, err
	}
	logctx.Warnf(ctx, "downlaod layer with torrent failed and will download-by-tcp: %s", err.Error())
	p.event.SendObjEvent(ctx, recorder.Normal, fmt.Sprintf("Download-by-tcp from '%s' "+
		"with file '%s' started (because torrent download failed)", resp.Located, resp.FilePath))
	if err = p.downloadByTCP(ctx, rw, resp.Located, transmitted, resp.FilePath); err != nil {
		canReverse := transmitted == 0
		p.event.SendObjEvent(ctx, recorder.Warning, fmt.Sprintf("Download-by-tcp from '%s "+
			"with file '%s' failed: %s", resp.Located, resp.FilePath, err.Error()))
		return canReverse, errors.Wrapf(err, "download by tcp failed")
	}
	p.event.SendObjEvent(ctx, recorder.Normal, fmt.Sprintf("Download-by-tcp from '%s' "+
		"with file '%s' success", resp.Located, resp.FilePath))
	return true, nil
}

func (p *upstreamProxy) downloadByTCP(ctx context.Context, rw http.ResponseWriter, target string, startPos int64,
	filePath string) error {
	// NOCC:Server Side Request Forgery(只是代码封装，所有 URL都是可信的)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s:%d%s", target,
		p.op.HTTPPort, apiclient.CustomAPITransferLayerTCP), nil)
	if err != nil {
		return errors.Wrapf(err, "create http.request failed")
	}
	query := req.URL.Query()
	query.Set("file", filePath)
	req.URL.RawQuery = query.Encode()
	logctx.Infof(ctx, "download layer from target '%s' with tcp starting", target)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "download layer from target '%s' with tcp failed", target)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("download layer from target '%s' with tcp resp code not 200 but %d",
			target, resp.StatusCode)
	}
	if startPos != 0 {
		if _, err = io.CopyN(io.Discard, resp.Body, startPos); err != nil {
			return errors.Wrapf(err, "download-by-tcp give up '%d' bytes failed", startPos)
		}
	}
	//buf := make([]byte, 1024*1024)
	//if _, err = io.CopyBuffer(rw, resp.Body, buf); err != nil {
	//	return errors.Wrapf(err, "download-by-tcp io.copy failed")
	//}
	if _, err = io.Copy(rw, resp.Body); err != nil {
		return errors.Wrapf(err, "download-by-tcp io.copy failed")
	}
	logctx.Infof(ctx, "layer download-by-tcp rewrite to http.writer success")
	return nil
}

func getBearerToken(req *http.Request) string {
	auth := req.Header.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(auth, "Bearer ")
}
