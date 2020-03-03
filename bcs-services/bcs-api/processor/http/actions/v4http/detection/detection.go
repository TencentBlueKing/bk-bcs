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

package detection

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-services/bcs-api/metric"
	"bk-bcs/bcs-services/bcs-api/processor/http/actions"
	"bk-bcs/bcs-services/bcs-api/regdiscv"

	"github.com/emicklei/go-restful"
)

const (
	BCSAPIPrefix         = "/bcsapi/v4/detection/"
	BCSDetectionPrefixV4 = "/detection/v4/"
)

var (
	// FlushInterval specifies the flush interval
	// to flush to the client while copying the
	// response body.
	flushImmediately time.Duration = -1
)

func init() {
	actions.RegisterAction(actions.Action{Verb: "POST", Path: "/bcsapi/v4/detection/{uri:*}", Params: nil, Handler: detectionProxyActions})
	actions.RegisterAction(actions.Action{Verb: "PUT", Path: "/bcsapi/v4/detection/{uri:*}", Params: nil, Handler: detectionProxyActions})
	actions.RegisterAction(actions.Action{Verb: "GET", Path: "/bcsapi/v4/detection/{uri:*}", Params: nil, Handler: detectionProxyActions})
	actions.RegisterAction(actions.Action{Verb: "DELETE", Path: "/bcsapi/v4/detection/{uri:*}", Params: nil, Handler: detectionProxyActions})
}

// defaultdetectionTransport is default detection transport instance.
var defaultdetectionTransport = &http.Transport{
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

// detectionDirector directe http request to target detection service.
func detectionDirector(req *http.Request) {
	rd, err := regdiscv.GetRDiscover()
	if err != nil {
		blog.Error("detection director, can't get discovery handler, %+v", err)
		return
	}
	serv, err := rd.GetModuleServers(types.BCS_MODULE_NETWORKDETECTION)
	if err != nil {
		blog.Error("detection director, can't get target server module[%s] from RD, %+v", types.BCS_MODULE_NETWORKDETECTION, err)
		return
	}
	ser, ok := serv.(*types.NetworkDetectionServInfo)
	if !ok {
		blog.Errorf("detection director, can't parse detection info from RD, %+v", serv)
		return
	}

	// directe to new detection URL.
	req.URL.Scheme = ser.Scheme
	req.URL.Host = fmt.Sprintf("%s:%d", ser.IP, ser.Port)
	req.URL.Path = BCSDetectionPrefixV4 + strings.Replace(req.URL.Path, BCSAPIPrefix, "", 1)

	if strings.ToLower(ser.Scheme) == "https" &&
		defaultdetectionTransport.TLSClientConfig == nil {

		cliTls, err := rd.GetClientTls()
		if err != nil {
			blog.Errorf("detection director, can't get detection client TLS configs from RD, %+v", err)
			return
		}
		// common TLS configs, used by default detection transport.
		defaultdetectionTransport.TLSClientConfig = cliTls
	}
}

// detectionModifyResponse modifies detection response.
func detectionModifyResponse(resp *http.Response) error {
	// do nothing.
	return nil
}

// defaultdetectionProxy is default detection reverse proxy.
var defaultdetectionProxy = &httputil.ReverseProxy{
	FlushInterval:  flushImmediately,
	Director:       detectionDirector,
	Transport:      defaultdetectionTransport,
	ModifyResponse: detectionModifyResponse,
}

// detectionProxyActions is actions handler for detection proxy.
func detectionProxyActions(req *restful.Request, resp *restful.Response) {
	start := time.Now()
	blog.V(3).Infof("request to detection, client[%s] req[%+v], method[%s]",
		req.Request.RemoteAddr, req.Request.URL.Path, req.Request.Method)

	// proxy to detection.
	defaultdetectionProxy.ServeHTTP(resp, req.Request)

	metric.RequestCount.WithLabelValues("detection", req.Request.Method).Inc()
	metric.RequestLatency.WithLabelValues("detection", req.Request.Method).Observe(time.Since(start).Seconds())
}
