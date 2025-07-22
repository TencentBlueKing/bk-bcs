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

// Package mesos xxx
package mesos

import (
	"crypto/tls"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bhttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	restful "github.com/emicklei/go-restful/v3"
	"github.com/ghodss/yaml"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
)

const (
	// mediaHeader key for http media content type
	medieTypeHeader = "Content-Type"
	// mediaTypeApplicationJSON json payload for http body
	mediaTypeApplicationJSON = "application/json" // nolint
	// mediaTypeApplicationYaml yaml payload for http body
	mediaTypeApplicationYaml = "application/x-yaml"
)

// TunnelHandler mesos tunnel request handler
type TunnelHandler struct {
	clientTLS *tls.Config

	tunnelDispatcher *WsTunnelDispatcher
}

// NewTunnelHandler create mesos tunnel handler
func NewTunnelHandler(clientTLS *tls.Config, tunnelDispatcher *WsTunnelDispatcher) *TunnelHandler {
	return &TunnelHandler{
		clientTLS:        clientTLS,
		tunnelDispatcher: tunnelDispatcher,
	}
}

func (th *TunnelHandler) request2mesosapi(req *restful.Request, uri, method string) (string, error) {
	start := time.Now()

	data, err := io.ReadAll(req.Request.Body)
	if err != nil {
		metrics.ReportAPIRequestMetric("mesos_tunnel_request", method, metrics.LibCallStatusErr, start)
		blog.Error("handler url %s read request body failed, error: %s", uri, err.Error())
		err1 := bhttp.InternalError(common.BcsErrCommHttpReadBodyFail, common.BcsErrCommHttpReadBodyFailStr)
		return err1.Error(), nil
	}
	// check application media type
	if mediaTypeApplicationYaml == req.Request.Header.Get(medieTypeHeader) {
		data, err = yamlTOJSON(data)
		if err != nil {
			blog.Errorf("bcs-user-manager handle url %s yaml to json failed, %s", uri, err.Error())
			mediaErr := bhttp.InternalError(common.BcsErrApiMediaTypeError, common.BcsErrApiMediaTypeErrorStr)
			return mediaErr.Error(), nil
		}
		blog.V(3).Infof("bcs-user-manager handle url %s converting yaml to json successfully", uri)
	}

	cluster := req.Request.Header.Get("BCS-ClusterID")
	if cluster == "" {
		metrics.ReportAPIRequestMetric("mesos_tunnel_request", method, metrics.LibCallStatusErr, start)
		blog.Error("handler url %s read header BCS-ClusterID is empty", uri)
		err1 := bhttp.InternalError(common.BcsErrCommHttpParametersFailed, "http header BCS-ClusterID can't be empty")
		return err1.Error(), nil
	}

	httpcli := httpclient.NewHttpClient()
	httpcli.SetHeader(medieTypeHeader, "application/json")
	httpcli.SetHeader("Accept", "application/json")

	// 先从websocket dialer缓存中查找websocket链
	serverAddr, tp, found := th.tunnelDispatcher.LookupWsTransport(cluster)
	if found {
		url := fmt.Sprintf("%s%s", serverAddr, uri)
		if strings.HasPrefix(serverAddr, "https") {
			if th.clientTLS == nil {
				blog.Errorf("client tls is empty")
			}
			tp.TLSClientConfig = th.clientTLS
		}
		httpcli.SetTransPort(tp)

		blog.Info(url)
		reply, err := httpcli.Request(url, method, req.Request.Header, data)
		if err != nil {
			metrics.ReportAPIRequestMetric("mesos_tunnel_request", method, metrics.LibCallStatusErr, start)
			blog.Error("request url %s error %s", url, err.Error())
			err1 := bhttp.InternalError(common.BcsErrApiRequestMesosApiFail, common.BcsErrApiRequestMesosApiFailStr)
			return err1.Error(), nil
		}

		metrics.ReportAPIRequestMetric("mesos_tunnel_request", method, metrics.LibCallStatusOK, start)
		return string(reply), err
	}

	err1 := bhttp.InternalError(common.BcsErrApiGetMesosApiFail, fmt.Sprintf("mesos cluster %s not found", cluster))
	return err1.Error(), nil
}

// HandlePostActions handle post action
func (th *TunnelHandler) HandlePostActions(req *restful.Request, resp *restful.Response) {
	blog.V(3).Infof("client %s request %s", req.Request.RemoteAddr, req.Request.URL.Path)

	url := req.Request.URL.Path

	if req.Request.URL.RawQuery != "" {
		url = fmt.Sprintf("%s?%s", url, req.Request.URL.RawQuery)
	}

	data, _ := th.request2mesosapi(req, url, "POST")
	_, _ = resp.Write([]byte(data))
}

// HandleGetActions handle get action
func (th *TunnelHandler) HandleGetActions(req *restful.Request, resp *restful.Response) {
	blog.V(3).Infof("client %s request %s", req.Request.RemoteAddr, req.Request.URL.Path)
	url := req.Request.URL.Path

	if req.Request.URL.RawQuery != "" {
		url = fmt.Sprintf("%s?%s", url, req.Request.URL.RawQuery)
	}

	data, _ := th.request2mesosapi(req, url, "GET")
	_, _ = resp.Write([]byte(data))
}

// HandleDeleteActions handle delete action
func (th *TunnelHandler) HandleDeleteActions(req *restful.Request, resp *restful.Response) {
	blog.V(3).Infof("client %s request %s", req.Request.RemoteAddr, req.Request.URL.Path)
	url := req.Request.URL.Path

	if req.Request.URL.RawQuery != "" {
		url = fmt.Sprintf("%s?%s", url, req.Request.URL.RawQuery)
	}

	data, _ := th.request2mesosapi(req, url, "DELETE")
	_, _ = resp.Write([]byte(data))
}

// HandlePutActions handle put action
func (th *TunnelHandler) HandlePutActions(req *restful.Request, resp *restful.Response) {
	blog.V(3).Infof("client %s request %s", req.Request.RemoteAddr, req.Request.URL.Path)
	url := req.Request.URL.Path

	if req.Request.URL.RawQuery != "" {
		url = fmt.Sprintf("%s?%s", url, req.Request.URL.RawQuery)
	}

	data, _ := th.request2mesosapi(req, url, "PUT")
	_, _ = resp.Write([]byte(data))
}

// yamlTOJSON check if mesos request body is yaml,
// then convert yaml to json
func yamlTOJSON(rawData []byte) ([]byte, error) {
	if len(rawData) == 0 {
		return nil, nil
	}
	return yaml.YAMLToJSON(rawData)
}
