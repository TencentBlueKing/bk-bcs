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

package kubedriver

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-driver/kubedriver/custom"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-driver/kubedriver/options"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-driver/kubedriver/versions"
	"io/ioutil"
	"net/http"
	urllib "net/url"
	"strings"

	restful "github.com/emicklei/go-restful"
	"github.com/parnurzeal/gorequest"
)

const CodeRequestFailed = 4001
const CodeRequestSuccess = 0

type KubeSmartProxier struct {
	KubeMasterURL string
	TLSConfig     options.TLSConfig
	KubeURLPrefix string

	serverAPIPrefer KubeAPIPrefer
	serverVersion   KubeVersion
}

func NewKubeSmartProxier(kubeMasterURL string, TLSConfig options.TLSConfig) KubeSmartProxier {
	return KubeSmartProxier{
		KubeMasterURL: kubeMasterURL,
		TLSConfig:     TLSConfig,
		KubeURLPrefix: DefaultKubeURLPrefix,
	}
}

func (c *KubeSmartProxier) IfKubeNeedTls() bool {
	kubeURL, _ := urllib.Parse(c.KubeMasterURL)
	return kubeURL.Scheme == options.HTTPS
}

// RegisterToWS registers a smart Kube API Proxy client to a go-restful WebService
func (c *KubeSmartProxier) RegisterToWS(ws *restful.WebService) {
	methods := []string{"GET", "POST", "DELETE", "PUT", "PATCH"}
	for _, methodName := range methods {
		ws.Route(
			RouteByMethodName(ws, methodName, fmt.Sprintf("/%s/{subpath:*}", c.KubeURLPrefix)).To(c.GeneralAPIHandle))
	}
}

// RequestServerVersion request API server to get version
func (c *KubeSmartProxier) RequestServerVersion() (KubeVersion, error) {
	url := fmt.Sprintf("%s/version", strings.TrimSuffix(c.KubeMasterURL, "/"))
	result := KubeVersion{}

	// Use client certs if given
	goreq := gorequest.New()
	if c.IfKubeNeedTls() {
		tlsConfig, err := c.TLSConfig.ToConfigObj()
		if err != nil {
			return result, err
		}
		goreq = goreq.TLSClientConfig(tlsConfig)
	}

	resp, respBody, errs := goreq.Get(url).EndStruct(&result)
	if len(errs) > 0 {
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			err := errors.New("unauthorized request, maybe you give a wrong client certs")
			return result, err
		}

		err := errs[0]
		return result, err
	}
	if !result.IsValid() {
		err := fmt.Errorf("not a valid version response, body=%s", respBody)
		return result, err
	}

	c.serverVersion = result
	return result, nil
}

// RequestServerVersion request API server to get version
func (c *KubeSmartProxier) RequestAPIPrefer() error {
	url := fmt.Sprintf("%s/apis", strings.TrimSuffix(c.KubeMasterURL, "/"))
	var result KubeAPIPrefer

	// Use client certs if given
	goreq := gorequest.New()
	if c.IfKubeNeedTls() {
		tlsConfig, err := c.TLSConfig.ToConfigObj()
		if err != nil {
			return err
		}
		goreq = goreq.TLSClientConfig(tlsConfig)
	}

	_, _, errs := goreq.Get(url).EndStruct(&result)
	if len(errs) > 0 {
		return errs[0]
	}

	c.serverAPIPrefer = result
	return nil
}

//GeneralAPIHandle create
func (c *KubeSmartProxier) GeneralAPIHandle(request *restful.Request, response *restful.Response) {
	subPath := request.PathParameter("subpath")

	// Redirect to Custom API
	// Q: Why we do not add a new path to handler different custom api?
	// A: The original design of bcs-api & bcs-route don't care about specific uri, the routing task of uri pushing to k8s-apiServer.
	//    But now, adding new path for custom api could break the behavior of bcs-api & bcs-route, which we don't want to.
	//    So driver have to swallow the routing task itself.
	customApiRouter := custom.NewRouter()
	customHandler := customApiRouter.Route(subPath)

	if customHandler != nil {
		err := customHandler.Config(c.KubeMasterURL, c.TLSConfig)
		if err != nil {
			custom.CustomServerErrorResponse(response, err.Error())
			return
		}
		customHandler.Handler(request, response)
		return
	}

	// redirect to k8s apiServer
	c.ForwardToKubeAPI(request, response)
	return
}

// ForwardToKubeAPI forwards incoming request to kubernetes API Server
func (c *KubeSmartProxier) ForwardToKubeAPI(request *restful.Request, response *restful.Response) {
	subPath := request.PathParameter("subpath")
	rawRequest := request.Request
	body, err := ioutil.ReadAll(rawRequest.Body)
	if err != nil {
		custom.CustomServerErrorResponse(response, "error reading raw request body")
		return
	}

	// transfer subPath to FullUrl
	clientSetter := versions.ClientSetter{}
	version := strings.Join([]string{c.serverVersion.Major, c.serverVersion.Minor}, ".")
	err = clientSetter.GetClientSetUrl(subPath, version, c.serverAPIPrefer.Map())
	if err != nil {
		custom.CustomServerErrorResponse(response, err.Error())
		return
	}

	//if json contains apiVersion, then use it
	apiVersion := json.Get(body, "apiVersion").ToString()
	apiVersion = strings.TrimSpace(apiVersion)
	if apiVersion != "" {
		//clientSetter.ClientSet = /apis/extensions/v1beta1/
		clientSet := strings.Trim(clientSetter.ClientSet, "/")
		group := strings.Split(clientSet, "/")[0]
		clientSetter.ClientSet = fmt.Sprintf("%s/%s", group, apiVersion)
	}

	url := fmt.Sprintf("%s/%s/%s",
		strings.TrimSuffix(c.KubeMasterURL, "/"),
		strings.Trim(clientSetter.ClientSet, "/"),
		strings.TrimPrefix(subPath, "/"))

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
	if c.IfKubeNeedTls() {
		tlsConfig, err := c.TLSConfig.ToConfigObj()
		if err != nil {
			custom.CustomServerErrorResponse(response, err.Error())
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
		custom.CustomServerErrorResponse(response, message)
		return
	}

	respBody, _ := ioutil.ReadAll(resp.Body)
	for key := range resp.Header {
		if key == "Content-Length" {
			continue
		}
		response.AddHeader(key, resp.Header.Get(key))
	}

	var driverResp custom.APIResponse
	if resp.StatusCode >= 400 {
		blog.Errorf("request url %s resp code %d respBody %s", url, resp.StatusCode, string(respBody))
		driverResp = custom.APIResponse{
			Result:  false,
			Message: json.Get(respBody, "message").ToString(),
			Code:    CodeRequestFailed,
			Data:    nil,
		}
	} else {
		blog.V(3).Infof("request url %s resp success", url)
		driverResp = custom.APIResponse{
			Result:  true,
			Message: "success",
			Code:    CodeRequestSuccess,
			Data:    json.Get(respBody).GetInterface(),
		}
	}
	custom.CustomSimpleHTTPResponse(response, resp.StatusCode, driverResp)
	return
}
