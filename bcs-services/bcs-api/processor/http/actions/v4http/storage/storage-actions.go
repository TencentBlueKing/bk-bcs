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

package storage

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/emicklei/go-restful"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/metric"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/processor/http/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/regdiscv"
)

const (
	BCSAPIPrefix       = "/bcsapi/v4/storage/"
	BCSStoragePrefixV1 = "/bcsstorage/v1/"
)

var (
	// FlushInterval specifies the flush interval
	// to flush to the client while copying the
	// response body.
	flushImmediately time.Duration = -1
)

func init() {
	actions.RegisterAction(actions.Action{Verb: "POST", Path: "/bcsapi/v4/storage/{uri:*}", Params: nil, Handler: storageProxyActions})
	actions.RegisterAction(actions.Action{Verb: "PUT", Path: "/bcsapi/v4/storage/{uri:*}", Params: nil, Handler: storageProxyActions})
	actions.RegisterAction(actions.Action{Verb: "GET", Path: "/bcsapi/v4/storage/{uri:*}", Params: nil, Handler: storageProxyActions})
	actions.RegisterAction(actions.Action{Verb: "DELETE", Path: "/bcsapi/v4/storage/{uri:*}", Params: nil, Handler: storageProxyActions})
}

// defaultStorageTransport is default storage transport instance.
var defaultStorageTransport = &http.Transport{
	// base proxy, could be covered by director.
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,

	// TLSHandshakeTimeout specifies the maximum amount of time waiting to
	// wait for a TLS handshake. Zero means no timeout.
	TLSHandshakeTimeout: 10 * time.Second,

	// MaxConnsPerHost optionally limits the total number of
	// connections per host, including connections in the dialing,
	// active, and idle states. On limit violation, dials will block.
	// Zero means no limit.
	// MaxConnsPerHost: 100,

	// MaxIdleConnsPerHost, if non-zero, controls the maximum idle
	// (keep-alive) connections to keep per-host.
	MaxIdleConnsPerHost: 25,

	// IdleConnTimeout is the maximum amount of time an idle
	// (keep-alive) connection will remain idle before closing itself.
	// Zero means no limit.
	// IdleConnTimeout: 10 * time.Minute,
}

// storageDirector directe http request to target storage service.
func storageDirector(req *http.Request) {
	rd, err := regdiscv.GetRDiscover()
	if err != nil {
		blog.Error("storage director, can't get discovery handler, %+v", err)
		return
	}
	serv, err := rd.GetModuleServers(types.BCS_MODULE_STORAGE)
	if err != nil {
		blog.Error("storage director, can't get target server module[%s] from RD, %+v", types.BCS_MODULE_STORAGE, err)
		return
	}
	ser, ok := serv.(*types.BcsStorageInfo)
	if !ok {
		blog.Errorf("storage director, can't parse storage info from RD, %+v", serv)
		return
	}

	// directe to new storage URL.
	req.URL.Scheme = ser.Scheme
	req.URL.Host = fmt.Sprintf("%s:%d", ser.IP, ser.Port)
	req.URL.Path = BCSStoragePrefixV1 + strings.Replace(req.URL.Path, BCSAPIPrefix, "", 1)

	if strings.ToLower(ser.Scheme) == "https" &&
		defaultStorageTransport.TLSClientConfig == nil {

		cliTls, err := rd.GetClientTls()
		if err != nil {
			blog.Errorf("storage director, can't get storage client TLS configs from RD, %+v", err)
			return
		}
		// common TLS configs, used by default storage transport.
		defaultStorageTransport.TLSClientConfig = cliTls
	}
}

// storageModifyResponse modifies storage response.
func storageModifyResponse(resp *http.Response) error {
	// do nothing.
	return nil
}

// defaultStorageProxy is default storage reverse proxy.
var defaultStorageProxy = &httputil.ReverseProxy{
	FlushInterval:  flushImmediately,
	Director:       storageDirector,
	Transport:      defaultStorageTransport,
	ModifyResponse: storageModifyResponse,
}

// storageProxyActions is actions handler for storage proxy.
func storageProxyActions(req *restful.Request, resp *restful.Response) {
	start := time.Now()
	blog.V(3).Infof("request to storage, client[%s] req[%+v], method[%s]",
		req.Request.RemoteAddr, req.Request.URL.Path, req.Request.Method)

	// proxy to storage.
	defaultStorageProxy.ServeHTTP(resp, req.Request)

	metric.RequestCount.WithLabelValues("storage", req.Request.Method).Inc()
	metric.RequestLatency.WithLabelValues("storage", req.Request.Method).Observe(time.Since(start).Seconds())
}
