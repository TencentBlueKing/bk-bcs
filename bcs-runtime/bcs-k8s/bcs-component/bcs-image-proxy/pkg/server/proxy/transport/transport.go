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

// Package transport xxx
package transport

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/internal/logctx"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/options"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/apiclient"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/server/proxy/registryauth"
)

var (
	defaultRetry = 5
)

// ProxyTransport defines the proxy transport
type ProxyTransport struct {
	proxyRegistry *options.RegistryMapping
}

// DefaultProxyTransport return the default proxy tansport
func DefaultProxyTransport(proxyRegistry *options.RegistryMapping) *ProxyTransport {
	return &ProxyTransport{
		proxyRegistry: proxyRegistry,
	}
}

// RoundTrip round-trip the request with retries
func (r *ProxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	var resp *http.Response
	var err error
	op := options.GlobalOptions()
	tp := op.HTTPProxyTransport()
	for i := 0; i < defaultRetry; i++ {
		resp, err = tp.RoundTrip(req)
		if err != nil {
			logctx.Warnf(ctx, "transport do request '%s' got failed(retry=%d): %s", req.URL.String(),
				i, err.Error())
			time.Sleep(time.Second)
			continue
		}
		if resp.StatusCode < http.StatusInternalServerError &&
			resp.StatusCode != http.StatusTemporaryRedirect &&
			resp.StatusCode != http.StatusTooManyRequests &&
			resp.StatusCode != http.StatusUnauthorized {
			return resp, nil
		}
		if i != 0 {
			logctx.Warnf(ctx, "transport do request got status code: %d, retry=%d", resp.StatusCode, i)
		}
		// no need retry or handle un-auth for this request
		if req.RequestURI == "/v1/" || req.RequestURI == "/v2/" {
			return resp, nil
		}
		switch resp.StatusCode {
		case http.StatusTemporaryRedirect:
			location := resp.Header.Get("Location")
			logctx.Infof(ctx, "Location: %s", location)
			var locationResp *http.Response
			locationResp, err = http.Get(location)
			if err != nil {
				return resp, errors.Wrapf(err, "get location '%s' failed", location)
			}
			return locationResp, nil
		case http.StatusUnauthorized:
			var authReq *registryauth.AuthRequest
			authReq, err = registryauth.ParseAuthRequest(resp.Header.Get(apiclient.RegistryAuthenticateHeader))
			if err != nil {
				logctx.Warnf(ctx, "pase auth request failed: %s", err.Error())
				return resp, nil
			}
			var tokenResult *registryauth.AuthToken
			tokenResult, err = registryauth.HandleRegistryUnauthorized(ctx, authReq, r.proxyRegistry)
			if err != nil {
				logctx.Warnf(ctx, "handle unauthorized request failed: %s", err.Error())
				return resp, nil
			}
			logctx.Infof(ctx, "transport handle unauthorized request return bearer token: %s", tokenResult.Token)
			// set auth token and to retry again
			req.Header.Set("Authorization", "Bearer "+tokenResult.Token)
		}
		// handle response status code 429, sleeping random time
		rand.New(rand.NewSource(time.Now().UnixNano()))
		time.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond)
	}
	return resp, errors.Wrapf(err, "handle request failed after %d retries", defaultRetry)
}
