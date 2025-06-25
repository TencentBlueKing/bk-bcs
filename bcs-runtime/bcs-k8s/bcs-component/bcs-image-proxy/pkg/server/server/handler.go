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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/internal/lock"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/internal/logctx"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/options"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/apiclient"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/bittorrent"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/server/ociscan"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/server/proxy/registryauth"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/server/requester"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/utils"
)

// CustomRegistry defines the object of custom-registry. CustomRegistry defines the custom-api, used by the master-node
// to provide external services.
type CustomRegistry struct {
	op             *options.ImageProxyOption
	torrentHandler *bittorrent.TorrentHandler
	ociScanner     *ociscan.ScanHandler
	cacheStore     store.CacheStore

	authLock          lock.Interface
	authTokens        *ttlcache.Cache[string, string]
	contentLengths    *ttlcache.Cache[string, int64]
	manifestCacheLock lock.Interface
	imageManifests    *ttlcache.Cache[string, string]
	downloadLayerLock lock.Interface

	nodeDownloadLock  sync.Mutex
	nodeDownloadTasks map[string]int
}

func (s *CustomRegistry) buildAuthKey(host, repo string) string {
	return fmt.Sprintf("%s|%s", host, repo)
}

func (s *CustomRegistry) buildManifestKey(host, repo, tag string) string {
	return fmt.Sprintf("%s|%s|%s", host, repo, tag)
}

func (s *CustomRegistry) buildContentLengthKey(host, digest string) string {
	return fmt.Sprintf("%s|%s", host, digest)
}

func (s *CustomRegistry) getAuthToken(host, repo, originalToken string) string {
	authKey := s.buildAuthKey(host, repo)
	auth := s.authTokens.Get(authKey)
	var bearerToken = originalToken
	if auth != nil && !auth.IsExpired() && auth.Value() != "" {
		bearerToken = auth.Value()
	}
	return bearerToken
}

// NewCustomRegistry create custom registry instance
func NewCustomRegistry(torrentHandler *bittorrent.TorrentHandler,
	ociScanner *ociscan.ScanHandler) *CustomRegistry {
	return &CustomRegistry{
		op:                options.GlobalOptions(),
		torrentHandler:    torrentHandler,
		ociScanner:        ociScanner,
		cacheStore:        store.GlobalRedisStore(),
		authLock:          lock.NewLocalLock(),
		manifestCacheLock: lock.NewLocalLock(),
		downloadLayerLock: lock.NewLocalLock(),
		authTokens:        ttlcache.New[string, string](),
		contentLengths:    ttlcache.New[string, int64](),
		imageManifests:    ttlcache.New[string, string](),
	}
}

func (s *CustomRegistry) httpError(rw http.ResponseWriter, req *http.Request, errMsg string) {
	logctx.Errorf(req.Context(), "custom-server response error: %s", errMsg)
	http.Error(rw, errMsg, http.StatusBadRequest)
}

// HTTPWrapper defines the http-wrapper
func (s *CustomRegistry) HTTPWrapper(f func(writer http.ResponseWriter, req *http.Request) (interface{}, error)) func(
	http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, req *http.Request) {
		obj, err := f(writer, req)
		if err != nil {
			s.httpError(writer, req, fmt.Sprintf("request '%s' failed: %s", req.URL.Path, err.Error()))
			return
		}
		if obj == nil {
			return
		}
		switch obj.(type) {
		case []byte:
			_, _ = writer.Write(obj.([]byte))
		case string:
			_, _ = writer.Write([]byte(obj.(string)))
		default:
			var bs []byte
			bs, err = json.Marshal(obj)
			if err != nil {
				s.httpError(writer, req, fmt.Sprintf("request '%s' marshal response failed: %s",
					req.URL.Path, err.Error()))
			} else {
				_, _ = writer.Write(bs)
			}
		}
	}
}

// parseRequest parse the request body
func (s *CustomRegistry) parseRequest(r *http.Request, body interface{}) error {
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		return errors.Wrapf(err, "read request body failed")
	}
	defer r.Body.Close()
	if err = json.Unmarshal(bs, body); err != nil {
		return errors.Wrapf(err, "unmarshal rqeuest body failed")
	}
	return nil
}

// RegistryAuth convergence every proxy's auth request(to registry). Proxy will
// transfer auth request to master, master will cache the bearer-token to reduce
// the pressure on the upstream-registry.
func (s *CustomRegistry) authRegistry(ctx context.Context, originalHost string, authenticateHeader string) (string,
	error) {
	registry := s.op.FilterRegistryMappingByOriginal(originalHost)
	if registry == nil {
		return "", errors.Errorf("options not have registry config for original host '%s'", originalHost)
	}
	authReq, err := registryauth.ParseAuthRequest(authenticateHeader)
	if err != nil {
		return "", errors.Wrapf(err, "parse auth request failed")
	}

	authKey := s.buildAuthKey(originalHost, registryauth.ParseRepoFromScope(authReq.Scope))
	s.authLock.Lock(ctx, authKey)
	defer s.authLock.UnLock(ctx, authKey)
	auth := s.authTokens.Get(authKey)
	if auth != nil && !auth.IsExpired() {
		return auth.Value(), nil
	}

	authReq, err = registryauth.ParseAuthRequest(authenticateHeader)
	if err != nil {
		return "", errors.Wrapf(err, "parse auth request failed")
	}
	tokenResult, err := registryauth.HandleRegistryUnauthorized(ctx, authReq, registry)
	if err != nil {
		return "", errors.Wrapf(err, "handle registry unauthorized failed")
	}
	expire := tokenResult.ExpiresIn
	if expire == 0 {
		expire = 30
	}
	s.authTokens.Set(authKey, tokenResult.Token, time.Duration(expire)*time.Second)
	return tokenResult.Token, nil
}

// RegistryGetManifest convergence every proxy's get-manifest request(to registry).
// Master will convergence all same get-manifest requests within 5 seconds become 1.
// NOTE: api for master, only the master will be called this API
func (s *CustomRegistry) RegistryGetManifest(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	req := new(apiclient.GetManifestRequest)
	if err := s.parseRequest(r, req); err != nil {
		return nil, err
	}
	ctx := r.Context()

	manifestKey := s.buildManifestKey(req.OriginalHost, req.Repo, req.Tag)
	s.manifestCacheLock.Lock(ctx, manifestKey)
	defer s.manifestCacheLock.UnLock(ctx, manifestKey)
	manifest := s.imageManifests.Get(manifestKey)
	if manifest != nil && !manifest.IsExpired() {
		return manifest.Value(), nil
	}

	logctx.Infof(ctx, "getting manifest: {host: %s, repo: %s, tag: %s}", req.OriginalHost, req.Repo, req.Tag)
	var bearerToken = s.getAuthToken(req.OriginalHost, req.Repo, req.BearerToken)
	getManifestUrl := req.OriginalHost + req.ManifestUrl
	authedOnce := false
	for i := 0; i < 5; i++ {
		manifestReq := &utils.HTTPRequest{
			Url:    "https://" + getManifestUrl,
			Method: http.MethodGet,
			Header: map[string]string{
				"Authorization": "Bearer " + bearerToken,
			},
		}
		resp, respBody, err := utils.SendHTTPRequestReturnResponse(ctx, manifestReq)
		if err == nil {
			manifestValue := string(respBody)
			s.imageManifests.Set(manifestKey, manifestValue, 30*time.Second)
			return manifestValue, nil
		}
		if resp == nil || (resp.StatusCode != http.StatusUnauthorized &&
			resp.StatusCode != http.StatusTooManyRequests) {
			return "", err
		}

		switch resp.StatusCode {
		case http.StatusUnauthorized:
			if authedOnce {
				return "", errors.Errorf("get-manifest authed success but got 401 response again")
			}
			authenticateHeader := resp.Header.Get(apiclient.RegistryAuthenticateHeader)
			logctx.Warnf(ctx, "get-manifest got 401 unauthorized, authenticate header: %s", authenticateHeader)
			if authenticateHeader == "" {
				return "", errors.Errorf("get-manifest response not have authenticate header")
			}
			bearerToken, err = s.authRegistry(ctx, req.OriginalHost, authenticateHeader)
			if err != nil {
				return "", errors.Wrapf(err, "get-manifest auth to registry failed")
			}
			authedOnce = true
		case http.StatusTooManyRequests:
			logctx.Warnf(ctx, "get-manifest got '429 too many requests' status, will retry again")
			time.Sleep(2 * time.Second)
		}
	}
	return "", errors.Errorf("cannot get manifest after %d retries", 5)
}

// RegistryGetLayerInfo get layer info from registry. Get the layer-info from original registry
func (s *CustomRegistry) RegistryGetLayerInfo(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "read request body failed")
	}
	defer r.Body.Close()
	req := new(apiclient.GetLayerInfoRequest)
	if err = json.Unmarshal(bs, req); err != nil {
		return nil, errors.Wrapf(err, "unmarshal rqeuest body failed")
	}
	repo, digest, isBlobGet := utils.IsBlobGet(req.LayerUrl)
	if !isBlobGet {
		return nil, errors.Wrapf(err, "request '%s' not blob url", req.LayerUrl)
	}
	ctx := logctx.SetLayerDigest(r.Context(), digest)
	var contentLength int64
	if contentLength, err = s.requestHeaderLayerContentLength(ctx, &downloadLayer{
		DownloadHost: req.OriginalHost,
		DownloadUrl:  req.LayerUrl,
		BearerToken:  s.getAuthToken(req.OriginalHost, repo, req.BearerToken),
		Repo:         repo,
		Digest:       digest,
	}); err != nil {
		return nil, errors.Wrapf(err, "request layer content-length failed")
	}
	return &apiclient.GetLayerInfoResponse{
		ContentLength: contentLength,
	}, nil
}

// DownloadLayerFromNode download layer from node. Provide the api to let slave-node to download layer from node.
func (s *CustomRegistry) DownloadLayerFromNode(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "read request body failed")
	}
	defer r.Body.Close()
	req := new(apiclient.CommonDownloadLayerRequest)
	if err = json.Unmarshal(bs, req); err != nil {
		return nil, errors.Wrapf(err, "unmarshal rqeuest body failed")
	}
	repo, digest, isBlobGet := utils.IsBlobGet(req.LayerUrl)
	if !isBlobGet {
		return nil, errors.Wrapf(err, "request '%s' not blob url", req.LayerUrl)
	}
	ctx := logctx.SetLayerDigest(r.Context(), digest)
	var contentLength int64
	resultPath := path.Join(s.op.TransferPath, utils.LayerFileName(digest))
	if contentLength, err = s.requestDownloadLayer(ctx, &downloadLayer{
		DownloadHost: req.OriginalHost,
		DownloadUrl:  req.LayerUrl,
		BearerToken:  req.BearerToken,
		Repo:         repo,
		Digest:       digest,
	}, resultPath); err != nil {
		return nil, errors.Wrapf(err, "download layer failed")
	}
	if err = s.cacheStore.SaveStaticLayer(ctx, digest, resultPath, true); err != nil {
		logctx.Errorf(ctx, "cace save static-layer '%s' failed: %s", resultPath, err.Error())
	}
	resp := &apiclient.CommonDownloadLayerResponse{
		Located:  s.op.Address,
		FilePath: resultPath,
		FileSize: contentLength,
	}
	if s.op.DisableTorrent || contentLength < apiclient.TwoHundredMB {
		return resp, nil
	}
	// should generate the torrent  if layer is too large
	var torrentBase64 string
	torrentBase64, err = s.torrentHandler.GenerateTorrent(ctx, digest, resultPath)
	if err != nil {
		logctx.Errorf(ctx, "generate torrent failed, just response located filepath: %s", err.Error())
	} else {
		resp.TorrentBase64 = torrentBase64
	}
	return resp, nil
}

// DownloadLayerFromMaster master will download the layer from original registry
// NOTE: api for master, only the master will be called this API
func (s *CustomRegistry) DownloadLayerFromMaster(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "read request body failed")
	}
	defer r.Body.Close()
	req := new(apiclient.CommonDownloadLayerRequest)
	if err = json.Unmarshal(bs, req); err != nil {
		return nil, errors.Wrapf(err, "unmarshal rqeuest body failed")
	}
	repo, digest, isBlobGet := utils.IsBlobGet(req.LayerUrl)
	if !isBlobGet {
		return nil, errors.Wrapf(err, "request '%s' not blob url", req.LayerUrl)
	}
	ctx := logctx.SetLayerDigest(r.Context(), digest)
	logctx.Infof(ctx, "master handling download-layer request")
	resp, err := s.downloadLayerFromMaster(ctx, req, digest, repo)
	if err == nil {
		logctx.Infof(ctx, "master handle download-layer request success: %s", resp.String())
	}
	return resp, err
}

func (s *CustomRegistry) downloadLayerFromMaster(ctx context.Context, req *apiclient.CommonDownloadLayerRequest,
	digest, repo string) (*apiclient.CommonDownloadLayerResponse, error) {
	s.downloadLayerLock.Lock(ctx, digest)
	defer s.downloadLayerLock.UnLock(ctx, digest)
	bearerToken := s.getAuthToken(req.OriginalHost, repo, req.BearerToken)
	downloadReq := &downloadLayer{
		DownloadHost: req.OriginalHost,
		DownloadUrl:  req.LayerUrl,
		BearerToken:  bearerToken,
		Repo:         repo,
		Digest:       digest,
	}
	contentLength, err := s.requestHeaderLayerContentLength(ctx, downloadReq)
	if err != nil {
		return nil, errors.Wrapf(err, "request layer content-length failed")
	}
	var resp *apiclient.CommonDownloadLayerResponse
	if resp, err = s.checkLayerHasCached(ctx, digest, contentLength); err == nil {
		return resp, nil
	}
	logctx.Warnf(ctx, "check layer has cached failed: %s", err.Error())
	return s.distributeDownloadLayerTask(ctx, downloadReq, contentLength)
}

func (s *CustomRegistry) distributeDownloadLayerTask(ctx context.Context, req *downloadLayer, contentLength int64) (
	*apiclient.CommonDownloadLayerResponse, error) {
	// master should download directly if small layer
	if contentLength < apiclient.TwentyMB {
		var err error
		resultPath := path.Join(s.op.SmallFilePath, utils.LayerFileName(req.Digest))
		if contentLength, err = s.requestDownloadLayer(ctx, req, resultPath); err != nil {
			return nil, errors.Wrapf(err, "download small-layer from original registry '%s/%s' failed",
				req.DownloadHost, req.DownloadUrl)
		}
		return &apiclient.CommonDownloadLayerResponse{
			Located:  s.op.Address,
			FilePath: resultPath,
			FileSize: contentLength,
		}, nil
	}
	// master should distribute tasks to other nodes
	targetNode := s.distributeNode()
	defer s.releaseNode(targetNode)
	logctx.Infof(ctx, "distribute task to node '%s'", targetNode)
	magnetResp, err := requester.DownloadLayerFromNode(ctx, targetNode, &apiclient.CommonDownloadLayerRequest{
		OriginalHost: req.DownloadHost,
		LayerUrl:     req.DownloadUrl,
		BearerToken:  req.BearerToken,
	})
	if err != nil {
		return nil, err
	}
	logctx.Infof(ctx, "distribute task from node completed")
	return magnetResp, nil
}

func (s *CustomRegistry) distributeNode() string {
	s.nodeDownloadLock.Lock()
	defer s.nodeDownloadLock.Unlock()

	if s.nodeDownloadTasks == nil {
		s.nodeDownloadTasks = make(map[string]int)
	}
	eps := s.op.ExternalConfig.LeaderConfig.Endpoints
	epMap := make(map[string]struct{})
	for _, ep := range eps {
		epMap[ep] = struct{}{}
		_, ok := s.nodeDownloadTasks[ep]
		if !ok {
			s.nodeDownloadTasks[ep] = 0
		}
	}
	for k := range epMap {
		if _, ok := s.nodeDownloadTasks[k]; !ok {
			delete(s.nodeDownloadTasks, k)
		}
	}
	var result string
	ans := 100000
	for k, v := range s.nodeDownloadTasks {
		if ans > v {
			ans = v
			result = k
		}
	}
	s.nodeDownloadTasks[result]++
	return result
}

func (s *CustomRegistry) releaseNode(node string) {
	s.nodeDownloadLock.Lock()
	defer s.nodeDownloadLock.Unlock()
	if v, ok := s.nodeDownloadTasks[node]; ok {
		s.nodeDownloadTasks[node] = v - 1
	}
}

// checkLayerHasCached check layer whether have in-cache
func (s *CustomRegistry) checkLayerHasCached(ctx context.Context, digest string, contentLength int64) (
	*apiclient.CommonDownloadLayerResponse, error) {
	// 从缓存中查询是否有 static-layers，并且验证文件完整性
	staticLayers, err := s.cacheStore.QueryStaticLayer(ctx, digest)
	if err != nil {
		return nil, errors.Wrapf(err, "query static-layers from cache failed")
	}
	for _, layer := range staticLayers {
		logctx.Infof(ctx, "check static-layer '%s, %s' starting", layer.Located, layer.Data)
		var resp *apiclient.CheckStaticLayerResponse
		resp, err = requester.CheckStaticLayer(ctx, layer.Located,
			&apiclient.CheckStaticLayerRequest{
				Digest: digest, LayerPath: layer.Data, ExpectedContentLength: contentLength})
		if err != nil {
			logctx.Errorf(ctx, "check static-layer '%s, %s' failed: %s", layer.Located, layer.Data, err.Error())
			continue
		}
		return &apiclient.CommonDownloadLayerResponse{
			TorrentBase64: resp.TorrentBase64,
			Located:       resp.Located,
			FileSize:      resp.FileSize,
			FilePath:      resp.LayerPath,
		}, nil
	}

	// 从缓存汇总查询是否有 oci-layers，并且获取 oci-layers
	ociLayers, err := s.cacheStore.QueryOCILayer(ctx, digest)
	if err != nil {
		return nil, errors.Wrapf(err, "query oci-layers from cache failed")
	}
	for _, layer := range ociLayers {
		logctx.Infof(ctx, "check oci-layer '%s, %s' starting", layer.Located, layer.Data)
		var resp *apiclient.CheckOCILayerResponse
		resp, err = requester.CheckOCILayer(ctx, layer.Located,
			&apiclient.CheckOCILayerRequest{Digest: digest, OCIType: string(layer.Type)})
		if err != nil {
			logctx.Errorf(ctx, "check oci-layer '%s, %s' failed: %s", layer.Located, layer.Data, err.Error())
			continue
		}
		return &apiclient.CommonDownloadLayerResponse{
			Located:  resp.Located,
			FileSize: resp.FileSize,
			FilePath: resp.LayerPath,
		}, nil
	}
	return nil, errors.Errorf("not get cached layer")
}

type downloadLayer struct {
	DownloadHost string `json:"downloadHost"`
	DownloadUrl  string `json:"downloadUrl"`
	BearerToken  string `json:"bearerToken"`
	Repo         string `json:"repo"`
	Digest       string `json:"digest"`
}

// requestHeaderLayerContentLength request the original registry to get the layer's content-length
func (s *CustomRegistry) requestHeaderLayerContentLength(ctx context.Context, req *downloadLayer) (
	int64, error) {
	key := s.buildContentLengthKey(req.DownloadHost, req.Digest)
	v := s.contentLengths.Get(key)
	if v != nil && !v.IsExpired() {
		return v.Value(), nil
	}
	resp, err := s.constructRequestLayer(ctx, req, http.MethodHead)
	if err != nil {
		return 0, errors.Wrapf(err, "http head-layer request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, errors.Errorf("http head-layer request failed with status %s", resp.Status)
	}
	layerSize := utils.FormatSize(resp.ContentLength)
	logctx.Infof(ctx, "request layer '%s' success, layer size: %s", req.Digest, layerSize)
	s.contentLengths.Set(key, resp.ContentLength, 10*time.Second)
	return resp.ContentLength, nil
}

var (
	retryTimes = 10
)

// constructRequestLayer construct get-layer request from original registry
func (s *CustomRegistry) constructRequestLayer(ctx context.Context, req *downloadLayer, method string) (
	*http.Response, error) {
	pullReq, err := http.NewRequest(method, "https://"+req.DownloadHost+req.DownloadUrl, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "create http request failed")
	}
	pullReq.Header.Set("Connection", "close")
	pullReq.Header.Set("Accept-Encoding", "gzip")
	pullReq.Header.Set("Authorization", "Bearer "+req.BearerToken)
	httpClient := &http.Client{
		Transport: s.op.HTTPProxyTransport(),
	}
	logctx.Infof(ctx, "do request: %s, %s", pullReq.Method, pullReq.URL.String())
	var resp *http.Response
	for i := 0; i < retryTimes; i++ {
		resp, err = httpClient.Do(pullReq)
		if err == nil {
			return resp, nil
		}
		logctx.Warnf(ctx, "do request '%s, %s' failed(retry=%d): %s", method, pullReq.URL.String(), i, err.Error())
		time.Sleep(time.Second)
	}
	return resp, err
}

// CheckStaticLayer check static layer whether have in other node
func (s *CustomRegistry) CheckStaticLayer(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "read request body failed")
	}
	defer r.Body.Close()
	req := new(apiclient.CheckStaticLayerRequest)
	if err = json.Unmarshal(bs, req); err != nil {
		return nil, errors.Wrapf(err, "unmarshal rqeuest body failed")
	}
	ctx := r.Context()
	ctx = logctx.SetLayerDigest(ctx, req.Digest)
	resp, err := s.checkStaticLayer(ctx, req)
	if err != nil {
		// should remove the cache from redis
		if delErr := s.cacheStore.DeleteStaticLayer(r.Context(), req.Digest); delErr != nil {
			logctx.Errorf(ctx, "delete static-layer from cache failed: %s", delErr.Error())
		}
		if delErr := s.cacheStore.DeleteTorrent(r.Context(), req.Digest); delErr != nil {
			logctx.Errorf(ctx, "delete torrent from cache failed: %s", delErr.Error())
		}
		return nil, err
	}
	return resp, nil
}

// checkStaticLayer check static-layer located on the corresponding node
func (s *CustomRegistry) checkStaticLayer(ctx context.Context, req *apiclient.CheckStaticLayerRequest) (
	*apiclient.CheckStaticLayerResponse, error) {
	fi, err := os.Stat(req.LayerPath)
	if err != nil {
		return nil, errors.Wrapf(err, "stat layer-file '%s' failed", req.LayerPath)
	}
	if fi.Size() != req.ExpectedContentLength {
		return nil, errors.Errorf("local file '%s' content-length '%d', not same as expcted '%d'",
			req.LayerPath, fi.Size(), req.ExpectedContentLength)
	}
	resp := &apiclient.CheckStaticLayerResponse{
		Located:   s.op.Address,
		LayerPath: req.LayerPath,
		FileSize:  fi.Size(),
	}
	if s.op.DisableTorrent || fi.Size() < s.op.TorrentThreshold {
		return resp, nil
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	torrentBase64, err := s.torrentHandler.GenerateTorrent(timeoutCtx, req.Digest, req.LayerPath)
	if err != nil {
		logctx.Errorf(ctx, "generate torrent for '%s' failed: %s", req.LayerPath, err.Error())
	} else {
		resp.TorrentBase64 = torrentBase64
	}
	return resp, nil
}

// CheckOCILayer check oci layer exist in all the k8s cluster
func (s *CustomRegistry) CheckOCILayer(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "read request body failed")
	}
	defer r.Body.Close()
	req := new(apiclient.CheckOCILayerRequest)
	if err = json.Unmarshal(bs, req); err != nil {
		return nil, errors.Wrapf(err, "unmarshal rqeuest body failed")
	}
	ctx := r.Context()
	ctx = logctx.SetLayerDigest(ctx, req.Digest)
	layerPath, err := s.ociScanner.GenerateLayer(ctx, req.OCIType, req.Digest)
	if err != nil {
		return nil, errors.Wrapf(err, "generate layer failed")
	}
	var fi os.FileInfo
	if fi, err = os.Stat(layerPath); err != nil {
		return nil, errors.Wrapf(err, "stat oc-layer '%s' failed", layerPath)
	}
	return &apiclient.CheckOCILayerResponse{
		Located:   s.op.Address,
		LayerPath: layerPath,
		FileSize:  fi.Size(),
	}, nil
}

// requestDownloadLayer request the original registry to download layer
func (s *CustomRegistry) requestDownloadLayer(ctx context.Context, req *downloadLayer, destPath string) (int64, error) {
	resp, err := s.constructRequestLayer(ctx, req, http.MethodGet)
	if err != nil {
		return 0, errors.Wrapf(err, "http download-layer request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var bs []byte
		bs, err = io.ReadAll(resp.Body)
		if err != nil {
			return 0, errors.Wrapf(err, "read response body failed")
		}
		return 0, errors.Errorf("http response status not 200 but %d: %s", resp.StatusCode, string(bs))
	}
	contentLength := resp.ContentLength
	layerSize := utils.FormatSize(contentLength)

	layerFullPath := path.Join(s.op.StoragePath, utils.LayerFileName(req.Digest))
	_ = os.RemoveAll(layerFullPath)
	layer, err := os.Create(layerFullPath)
	if err != nil {
		return 0, errors.Wrapf(err, "handle download_layer create file '%s' failed",
			layerFullPath)
	}
	defer layer.Close()

	progressCh := make(chan struct{})
	go func() {
		tick := time.NewTicker(5 * time.Second)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				var fi os.FileInfo
				if fi, err = layer.Stat(); err != nil {
					logctx.Infof(ctx, "downloading layer from original registry '%s' got stats failed: %s",
						layerFullPath, err.Error())
				} else {
					percent := float64(fi.Size()) / float64(resp.ContentLength) * 100
					downloadSize := utils.FormatSize(fi.Size())
					logctx.Infof(ctx, "downloading layer from original registry(%.2f%%): %s/%s",
						percent, downloadSize, layerSize)
				}
			case <-progressCh:
				return
			}
		}
	}()
	defer close(progressCh)
	if _, err = io.Copy(layer, resp.Body); err != nil {
		_ = os.RemoveAll(layer.Name())
		return 0, errors.Wrapf(err, "handle download_layer io copy failed")
	}
	logctx.Infof(ctx, "download layer '%s' successfully", layerFullPath)
	if err = os.Rename(layerFullPath, destPath); err != nil {
		return 0, errors.Wrapf(err, "renamse '%s' to '%s' failed", layerFullPath, destPath)
	}
	return contentLength, nil
}

// TransferLayerTCP transfer layer with tcp
func (s *CustomRegistry) TransferLayerTCP(rw http.ResponseWriter, r *http.Request) (interface{}, error) {
	requestFile := r.URL.Query().Get("file")
	if requestFile == "" {
		return nil, errors.Errorf("quyer param 'file' cannot empty")
	}
	if _, err := os.Stat(requestFile); err != nil {
		return nil, errors.Wrapf(err, "query file '%s' stat failed", requestFile)
	}
	http.ServeFile(rw, r, requestFile)
	return nil, nil
}
