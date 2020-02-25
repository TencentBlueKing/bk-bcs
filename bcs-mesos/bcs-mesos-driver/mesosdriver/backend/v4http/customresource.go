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

package v4http

import (
	"bk-bcs/bcs-common/common/blog"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	simplejson "github.com/bitly/go-simplejson"
	restful "github.com/emicklei/go-restful"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	//default kube custom resource definition url
	defaultCRDURL = "/apis/apiextensions.k8s.io/v1beta1/customresourcedefinitions"
	//default custom resource definition apiVersion we use
	defaultAPIVersion   = "apiextensions.k8s.io/v1beta1"
	defaultMesosVersion = "v4"
	defaultNamespaceURL = "api/v1/namespaces"
)

//kubeProxy proxy for custom resource
type kubeProxy struct {
	//kube config details
	config *rest.Config
	//client for namespace check
	client *http.Client
	//custom resource proxy
	crsProxy *httputil.ReverseProxy
	//custom resource definition proxy
	crdsProxy *httputil.ReverseProxy
}

func (proxy *kubeProxy) init() error {
	//create specified tranport from kube config
	httpRoundTripper, err := rest.TransportFor(proxy.config)
	if err != nil {
		blog.Errorf("bcs-mesos-driver initialize kube proxy transport failed, message: %s, config object: %v", err.Error(), proxy.config)
		return fmt.Errorf("bcs-mesos-driver create CustomResource transport failed")
	}
	proxy.client = &http.Client{
		Transport: httpRoundTripper,
	}
	proxy.crsProxy = &httputil.ReverseProxy{
		Director:  proxy.customResourceNamespaceValidate,
		Transport: httpRoundTripper,
	}
	proxy.crdsProxy = &httputil.ReverseProxy{
		Director:       proxy.apiVersionReqConvert,
		Transport:      httpRoundTripper,
		ModifyResponse: proxy.apiVersionResConvert,
	}
	blog.Infof("bcs-mesos-driver init local proxy for custom resource proxy successfully.")
	return nil
}

func (proxy *kubeProxy) isNamespaceActive(namespace string) bool {
	reqURL := fmt.Sprintf("%s/%s/%s", proxy.config.Host, defaultNamespaceURL, namespace)
	nsReq, _ := http.NewRequest(http.MethodGet, reqURL, nil)
	resp, err := proxy.client.Do(nsReq)
	if err != nil {
		blog.Errorf("check %s failed, %s", reqURL, err.Error())
		return false
	}
	if resp.StatusCode == http.StatusNotFound {
		blog.Infof("%s, namespace Not found", reqURL)
		return false
	}
	if resp.StatusCode == http.StatusOK {
		blog.Infof("%s, namespace %s is Found!", reqURL, namespace)
		return true
	}
	//others
	return false
}

func (proxy *kubeProxy) createNamespace(namespace string) error {
	reqURL := fmt.Sprintf("%s/%s", proxy.config.Host, defaultNamespaceURL)
	ns := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	nsBody, _ := json.Marshal(ns)
	nsReq, _ := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(nsBody))
	nsReq.Header.Add("Content-Type", "application/json")
	resp, err := proxy.client.Do(nsReq)
	if err != nil {
		blog.Errorf("create %s namespace failed. request body: %s", reqURL, string(nsBody))
		return err
	}
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusAccepted {
		blog.Infof("create namespace %s [%s] success.", reqURL, namespace)
		return nil
	}
	//others
	return fmt.Errorf("unknow reason for creation ns failed: %d", resp.StatusCode)
}

func (proxy *kubeProxy) customResourceNamespaceValidate(req *http.Request) {
	if req.Method != http.MethodPost {
		return
	}
	//validate namespace exist
	allBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		blog.Errorf("Reading custom resource Request for namespace validation failed, %s. URL: %s", err.Error(), req.URL.String())
		return
	}
	buffer := bytes.NewBuffer(allBytes)
	req.Body = ioutil.NopCloser(buffer)
	req.ContentLength = int64(buffer.Len())
	req.GetBody = func() (io.ReadCloser, error) {
		r := bytes.NewReader(allBytes)
		return ioutil.NopCloser(r), nil
	}
	jsonObj, err := simplejson.NewJson(allBytes)
	if err != nil {
		blog.Errorf("Custom Resource POST data is not expected json, %s. URL: %s", err.Error(), req.URL.String())
		return
	}
	meta := jsonObj.Get("metadata")
	namespace, _ := meta.Get("namespace").String()
	if len(namespace) == 0 {
		blog.Errorf("Custom Resource POST to %s lost Namespace. %s", req.URL.String(), err.Error())
		return
	}
	if proxy.isNamespaceActive(namespace) {
		return
	}
	blog.Infof("namespace %s do not exist, create first...", namespace)
	if err := proxy.createNamespace(namespace); err != nil {
		blog.Errorf("create Namespace for %s failed, %s", req.URL.String(), err.Error())
		return
	}
	blog.Infof("create Namespace for %s success.", req.URL.String())
}

//apiVersionReqConvert convert all request
//attention: all json key is
func (proxy *kubeProxy) apiVersionReqConvert(req *http.Request) {
	if req.Method == http.MethodGet || req.Method == http.MethodDelete {
		blog.V(3).Infof("bcs-mesos-driver skip %s convertion. Method: %s", req.URL.Path, req.Method)
		return
	}
	allBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		blog.Errorf("bcs-mesos-driver custom resource definition Request Convert failed, %s", err.Error())
		return
	}
	if len(allBytes) == 0 {
		blog.Errorf("bcs-mesos-driver get empty body when in POST or PUT request, URL: %s, Method: %s", req.URL.String(), req.Method)
		return
	}
	jsonObj, err := simplejson.NewJson(allBytes)
	if err != nil {
		blog.Errorf("bcs-mesos-driver decode json json failed, %s. details request body: %s", err.Error(), string(allBytes))
		return
	}
	blog.V(3).Infof("forwarding URL %s old body: %s", req.URL.String(), string(allBytes))
	jsonObj.Set("apiVersion", defaultAPIVersion)
	newBody, err := jsonObj.MarshalJSON()
	if err != nil {
		blog.Errorf("bcs-mesos-driver new custom resource definition Request json Marshal failed, %s. URL: %s", err.Error(), req.URL.Path)
		return
	}
	blog.V(3).Infof("forwarding URL %s new body: %s", req.URL.String(), string(newBody))
	buffer := bytes.NewBuffer(newBody)
	req.Body = ioutil.NopCloser(buffer)
	req.ContentLength = int64(buffer.Len())
	req.GetBody = func() (io.ReadCloser, error) {
		r := bytes.NewReader(newBody)
		return ioutil.NopCloser(r), nil
	}
	blog.Infof("bcs-mesos-driver custom resource definition [%s] Request convert success.", req.URL.Path)
}

func (proxy *kubeProxy) apiVersionResConvert(resp *http.Response) error {
	allBytes, _ := ioutil.ReadAll(resp.Body)
	if len(allBytes) == 0 {
		blog.Errorf("mesos-driver response for %s is Empty", resp.Request.URL.String())
		return nil
	}
	tmpStr := string(allBytes)
	newStr := strings.Replace(tmpStr, defaultAPIVersion, defaultMesosVersion, 1)
	buffer := bytes.NewBuffer([]byte(newStr))
	resp.Body = ioutil.NopCloser(buffer)
	resp.ContentLength = int64(buffer.Len())
	//setting header for contentLength
	resp.Header.Set("Content-Length", strconv.Itoa(buffer.Len()))
	blog.Infof("mesos-driver convert custom resource definition [%s] response success", resp.Request.URL.String())
	return nil
}

func (s *Scheduler) initKube() error {
	if s.config.KubeConfig == "" {
		//incluster configuration for mesos-driver container
		// blog.Infof("bcs-mesos-driver use in-cluster configuration...")
		// config, err := rest.InClusterConfig()
		// if err != nil {
		// 	blog.Errorf("bcs-mesos-driver create in-cluster configuraiont failed, %s", err.Error())
		// 	return nil, err
		// }
		// return config, nil
		blog.Infof("bcs-mesos-driver no kubeconfig detected, skip all compatible custom resource initialization")
		//compatible with history version, clean in-cluster code
		return nil
	}
	//outcluster deployment configuration for mesos-driver process
	blog.Infof("bcs-mesos-driver use process deployment with KubeConfig %s", s.config.KubeConfig)
	config, err := clientcmd.BuildConfigFromFlags("", s.config.KubeConfig)
	if err != nil {
		blog.Errorf("bcs-mesos-driver build configuration with kubeConfig %s failed, %s", s.config.KubeConfig, err.Error())
		return err
	}
	//create reversproxy for custom resource
	s.localProxy = &kubeProxy{
		config: config,
	}
	return s.localProxy.init()
}

//transparent forwarding for custom resource
// * block all watch request
// * redirect mesos driver url to api custom resource url
func (s *Scheduler) customResourceForwarding(req *restful.Request, resp *restful.Response) {
	if strings.Contains(req.Request.URL.Path, "watch") {
		blog.Warnf("bcs-mesos-driver custom resource do not support watch request. URL: %s", req.Request.URL.Path)
		resp.WriteErrorString(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}
	//change Path & Host
	rawRequest := req.Request
	original := rawRequest.URL.String()
	mesosURL := req.PathParameter("uri")
	kubeURL := filepath.Join("/apis", mesosURL)
	tmpURL, _ := url.Parse(s.localProxy.config.Host)
	rawRequest.URL.Scheme = tmpURL.Scheme
	rawRequest.URL.Host = tmpURL.Host
	rawRequest.URL.Path = kubeURL
	blog.Infof("bcs-mesos-driver custom resource forwarding from %s to %s", original, rawRequest.URL.String())
	s.localProxy.crsProxy.ServeHTTP(resp, rawRequest)
}

func (s *Scheduler) customResourceDefinitionForwarding(req *restful.Request, resp *restful.Response) {
	name := req.PathParameter("name")
	kubeURL := defaultCRDURL
	if len(name) != 0 {
		kubeURL = filepath.Join(defaultCRDURL, name)
	}
	rawRequest := req.Request
	original := rawRequest.URL.String()
	tmpURL, _ := url.Parse(s.localProxy.config.Host)
	rawRequest.URL.Scheme = tmpURL.Scheme
	rawRequest.URL.Host = tmpURL.Host
	rawRequest.URL.Path = kubeURL
	blog.Infof("bcs-mesos-driver CustomResourceDefinition forwarding from %s to %s", original, rawRequest.URL.String())
	s.localProxy.crdsProxy.ServeHTTP(resp, rawRequest)
}
