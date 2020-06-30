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

package custom

import (
	"bytes"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-driver/kubedriver/options"
	"io/ioutil"
	"net/http"
	urllib "net/url"
	"strings"

	restful "github.com/emicklei/go-restful"
)

type BcsClientAPIHandler struct {
	KubeMasterUrl string
	TLSConfig     options.TLSConfig
}

func (h *BcsClientAPIHandler) Handler(request *restful.Request, response *restful.Response) {
	subPath := request.PathParameter("subpath")
	targetPath := strings.Split(subPath, "bcsclient/")[1]

	rawRequest := request.Request
	body, err := ioutil.ReadAll(rawRequest.Body)
	if err != nil {
		CustomServerErrorResponse(response, "Reading raw request body failed")
		return
	}

	url := fmt.Sprintf("%s/%s",
		strings.TrimSuffix(h.KubeMasterUrl, "/"),
		strings.TrimPrefix(targetPath, "/"))

	// url param
	if rawRequest.URL.RawQuery != "" {
		url = fmt.Sprintf("%s?%s", url, rawRequest.URL.RawQuery)
	}
	proxyReq, _ := http.NewRequest(rawRequest.Method, url, bytes.NewReader(body))

	proxyReq.Header = make(http.Header)
	for key, value := range rawRequest.Header {
		proxyReq.Header[key] = value
	}

	// Respect client TLS config
	var httpClient *http.Client
	if h.IfKubeNeedTls() {
		tlsConfig, err := h.TLSConfig.ToConfigObj()
		if err != nil {
			CustomServerErrorResponse(response, err.Error())
			return
		}
		transport := &http.Transport{TLSClientConfig: tlsConfig}
		httpClient = &http.Client{Transport: transport}
	} else {
		httpClient = &http.Client{}
	}

	blog.V(3).Infof("forwarding request to %s, method=%s", url, rawRequest.Method)
	resp, err := httpClient.Do(proxyReq)
	if err != nil {
		message := fmt.Sprintf("error request kube server: %s", err)
		blog.Warn(message)
		CustomServerErrorResponse(response, message)
		return
	}

	respBody, _ := ioutil.ReadAll(resp.Body)
	for key := range resp.Header {
		response.AddHeader(key, resp.Header.Get(key))
	}
	if resp.Header.Get("Content-Type") == "" {
		response.AddHeader("Content-Type", "application/json")
	}
	response.Write(respBody)
}

func (h *BcsClientAPIHandler) Config(KubeMasterURL string, TLSConfig options.TLSConfig) error {
	h.KubeMasterUrl = KubeMasterURL
	h.TLSConfig = TLSConfig
	return nil
}

func (h *BcsClientAPIHandler) IfKubeNeedTls() bool {
	kubeURL, _ := urllib.Parse(h.KubeMasterUrl)
	return kubeURL.Scheme == options.HTTPS
}
