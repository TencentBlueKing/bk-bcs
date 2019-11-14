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

package filter

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"bk-bcs/bcs-common/common"
	"bk-bcs/bcs-common/common/blog"
	bhttp "bk-bcs/bcs-common/common/http"
	"bk-bcs/bcs-common/common/http/httpclient"
	commtypes "bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-common/common/zkclient"
	"bk-bcs/bcs-mesos/bcs-mesos-driver/mesosdriver/backend"

	"github.com/emicklei/go-restful"
)

const (
	RegexUrlApplication = ".*/namespaces/[^/]+/applications$"
	RegexUrlDeployment  = ".*/namespaces/[^/]+/deployments$"
)

type AdmissionWebhookFilter struct {
	sync.RWMutex

	scheduler backend.Scheduler
	//request scheduler http client
	schedClient *httpclient.HttpClient

	//key = Operation_Kind
	admissionHooks map[string]*commtypes.AdmissionWebhookConfiguration
	//mesos cluster zk client
	zkClient *zkclient.ZkClient
	//zk servers
	zkServers []string
}

func NewAdmissionWebhookFilter(scheduler backend.Scheduler, zkServers []string) RequestFilterFunction {
	hookFilter := &AdmissionWebhookFilter{
		scheduler:   scheduler,
		schedClient: scheduler.GetHttpClient(),
		zkServers:   zkServers,
	}
	hookFilter.zkClient = zkclient.NewZkClient(zkServers)

	err := hookFilter.zkClient.ConnectEx(time.Second * 5)
	if err != nil {
		blog.Errorf("AdmissionWebhookFilter connect zk %s error %v", hookFilter.zkServers, err.Error())
		os.Exit(1)
	}

	go hookFilter.start()

	return hookFilter
}

func (hook *AdmissionWebhookFilter) Execute(req *restful.Request) (int, error) {
	/*	//if http method not POST&PUT return
		if req.Request.Method!=http.MethodPost&&req.Request.Method!=http.MethodPut {
			return nil, 0
		}
		appMatch,_ := regexp.MatchString(RegexUrlApplication,req.Request.RequestURI)
		depMatch,_ := regexp.MatchString(RegexUrlDeployment,req.Request.RequestURI)
		if !appMatch&&!depMatch {
			return nil,0
		}*/

	body, err := ioutil.ReadAll(req.Request.Body)
	if err != nil || len(body) == 0 {
		return 0, nil
	}

	var meta *commtypes.TypeMeta
	err = json.Unmarshal(body, &meta)
	if err != nil {
		blog.V(3).Infof("AdmissionWebhookFilter handler url %s method %s Unmarshal data error %s, and return",
			req.Request.RequestURI, req.Request.Method, err.Error())
		req.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		return 0, nil
	}

	var operation string
	switch req.Request.Method {
	case http.MethodPost:
		operation = commtypes.AdmissionOperationCreate
	case http.MethodPut:
		operation = commtypes.AdmissionOperationUpdate
	default:
		operation = commtypes.AdmissionOperationUnknown
		blog.V(3).Infof("AdmissionWebhookFilter handler url %s method %s is invalid, and return",
			req.Request.RequestURI, req.Request.Method)
		req.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		return 0, nil
	}

	uuid := strings.ToUpper(fmt.Sprintf("%s_%s", operation, meta.Kind))
	admissionHook, ok := hook.admissionHooks[uuid]
	if !ok {
		blog.V(3).Infof("AdmissionWebhookFilter handler url %s method %s not match webhook, and return",
			req.Request.RequestURI, req.Request.Method)
		req.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		return 0, nil
	}

	blog.Infof("AdmissionWebhookFilter handler url %s method %s match webhook, and execute webhook",
		req.Request.RequestURI, req.Request.Method)
	newBody := body
	for _, webhook := range admissionHook.AdmissionWebhooks {
		hookBody, err := hook.requestAdmissionWebhook(webhook, newBody)
		if err != nil {
			blog.Errorf("admissionwebhook %s request webhoook %s error %s", uuid, webhook.Name, err.Error())
			if webhook.FailurePolicy == commtypes.WebhookFailurePolicyFail {
				blog.Errorf("AdmissionWebhookFilter handler url %s method %s failed, and policy fail return",
					req.Request.RequestURI, req.Request.Method)
				return common.BcsErrMesosDriverHttpFilterFailed, fmt.Errorf("request webhoook %s error %s", webhook.Name, err.Error())
			}
			blog.Infof("AdmissionWebhookFilter handler url %s method %s failed, and policy ignore continue",
				req.Request.RequestURI, req.Request.Method)
			continue
		}

		newBody = hookBody
	}

	req.Request.Body = ioutil.NopCloser(bytes.NewBuffer(newBody))
	return 0, nil
}

func (hook *AdmissionWebhookFilter) requestAdmissionWebhook(webhook *commtypes.AdmissionWebhook, body []byte) ([]byte, error) {
	if len(webhook.WebhookServers) == 0 {
		return nil, fmt.Errorf("webhook %s not found servers", webhook.Name)
	}

	pemCert, err := base64.StdEncoding.DecodeString(webhook.ClientConfig.CaBundle)
	if err != nil {
		return nil, err
	}

	hookClient := hook.initWebhookClient(pemCert)
	resp, err := hookClient.Post(webhook.WebhookServers[0], "application/json", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	by, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(string(by))
	}

	return by, nil
}

func (hook *AdmissionWebhookFilter) initWebhookClient(pemCerts []byte) *http.Client {
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(pemCerts)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: pool},
	}
	client := &http.Client{Transport: tr}
	return client
}

func (hook *AdmissionWebhookFilter) start() {
	//loop sync all admission webhooks
	go hook.loopSyncAdmissionWebhooks()
}

func (hook *AdmissionWebhookFilter) loopSyncAdmissionWebhooks() {
	time.Sleep(time.Second * 3)
	hook.syncAdmissionWebhooks()

	ticker := time.NewTicker(time.Second * 60)

	for {
		select {
		case <-ticker.C:
			hook.syncAdmissionWebhooks()
		}
	}
}

func (hook *AdmissionWebhookFilter) syncAdmissionWebhooks() {
	admissions, err := hook.fetchAllAdmissionWebhooks()
	if err != nil {
		blog.Errorf("AdmissionWebhookFilter fetch all admission webhooks error %s", err.Error())
		return
	}

	admissionHooks := make(map[string]*commtypes.AdmissionWebhookConfiguration)
	for _, ad := range admissions {
		key := strings.ToUpper(fmt.Sprintf("%s_%s", ad.ResourcesRef.Operation, ad.ResourcesRef.Kind))
		blog.V(3).Infof("AdmissionWebhookFilter add AdmissionWebhook(%s:%s) key %s", ad.NameSpace, ad.Name, key)
		//get webhook servers info
		for _, webhook := range ad.AdmissionWebhooks {
			servers, err := hook.fetchWebhookServers(webhook.ClientConfig.Namespace, webhook.ClientConfig.Name)
			if err != nil {
				blog.Errorf("AdmissionWebhookConfiguration(%s:%s) webhook %s fetch server error %s",
					ad.NameSpace, ad.Name, webhook.Name, err.Error())
				continue
			}

			webhook.WebhookServers = servers
		}

		admissionHooks[key] = ad
	}

	hook.Lock()
	hook.admissionHooks = admissionHooks
	hook.Unlock()
}

func (hook *AdmissionWebhookFilter) fetchAllAdmissionWebhooks() ([]*commtypes.AdmissionWebhookConfiguration, error) {
	blog.V(3).Info("AdmissionWebhookFilter fetch all admissionwebhooks")

	var err error
	if hook.scheduler.GetHost() == "" {
		err = fmt.Errorf("no scheduler is connected by driver")
		return nil, err
	}

	url := hook.scheduler.GetHost() + "/v1/admissionwebhooks"
	reply, err := hook.schedClient.GET(url, nil, nil)
	if err != nil {
		return nil, err
	}

	var resp *bhttp.APIRespone
	err = json.Unmarshal(reply, &resp)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal data %s to bhttp.APIRespone error %s", string(reply), err.Error())
	}

	if resp.Code != common.BcsSuccess {
		return nil, fmt.Errorf("request url %s resp code %d message %s", url, resp.Code, resp.Message)
	}

	var admissions []*commtypes.AdmissionWebhookConfiguration
	by, _ := json.Marshal(resp.Data)
	err = json.Unmarshal(by, &admissions)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal data %s to commtypes.AdmissionWebhookConfiguration error %s",
			string(by), err.Error())
	}

	return admissions, nil
}

func (hook *AdmissionWebhookFilter) fetchWebhookServers(ns, name string) ([]string, error) {
	key := fmt.Sprintf("/blueking/endpoint/%s/%s", ns, name)
	data, err := hook.zkClient.Get(key)
	if err != nil {
		return nil, fmt.Errorf("AdmissionWebhookFilter get zk %s error %s", key, err.Error())
	}

	var endpoints *commtypes.BcsEndpoint
	err = json.Unmarshal([]byte(data), &endpoints)
	if err != nil {
		return nil, err
	}

	servers := make([]string, 0)
	for _, end := range endpoints.Endpoints {
		if len(end.Ports) == 0 {
			blog.Errorf("webhook(%s:%s) Endpoints Ports is empty", ns, name)
			continue
		}

		port := end.Ports[0]
		switch end.NetworkMode {
		case "BRIDGE":
			server := fmt.Sprintf("https://%s.%s:%d", endpoints.Name, endpoints.NameSpace, port.HostPort)
			servers = append(servers, server)
		default:
			server := fmt.Sprintf("https://%s.%s:%d", endpoints.Name, endpoints.NameSpace, port.ContainerPort)
			servers = append(servers, server)
		}

		break
	}

	return servers, nil
}
