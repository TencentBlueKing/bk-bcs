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
	"strings"
	"sync"
	"time"

	"bk-bcs/bcs-common/common"
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/http/httpclient"
	commtypes "bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-mesos/bcs-mesos-driver/mesosdriver/backend"
	"bk-bcs/bcs-mesos/pkg/apis/bkbcs/v2"
	"bk-bcs/bcs-mesos/pkg/client/informers"
	"bk-bcs/bcs-mesos/pkg/client/internalclientset"
	v2lister "bk-bcs/bcs-mesos/pkg/client/lister/bkbcs/v2"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type AdmissionWebhookFilter struct {
	sync.RWMutex

	scheduler backend.Scheduler
	//request scheduler http client
	schedClient *httpclient.HttpClient

	//key = Operation_Kind
	admissionHooks map[string]*commtypes.AdmissionWebhookConfiguration

	//admissionwebhook cache.SharedIndexInformer
	adInformer cache.SharedIndexInformer
	adLister   v2lister.AdmissionWebhookConfigurationLister
}

func NewAdmissionWebhookFilter(scheduler backend.Scheduler, kubeconfig string) (RequestFilterFunction, error) {
	blog.Info("AdmissionWebhookFilter initialize...")
	hookFilter := &AdmissionWebhookFilter{
		scheduler:      scheduler,
		schedClient:    scheduler.GetHttpClient(),
		admissionHooks: make(map[string]*commtypes.AdmissionWebhookConfiguration),
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		blog.Errorf("AdmissionWebhookFilter build kubeconfig %s error %s", kubeconfig, err.Error())
		return nil, err
	}
	blog.Infof("AdmissionWebhookFilter build kubeconfig %s success", kubeconfig)

	bkbcsClientset, err := internalclientset.NewForConfig(restConfig)
	if err != nil {
		blog.Errorf("AdmissionWebhookFilter build clientset error %s", err.Error())
		return nil, err
	}

	stopCh := make(chan struct{})
	factory := informers.NewSharedInformerFactory(bkbcsClientset, time.Minute)
	hookFilter.adInformer = factory.Bkbcs().V2().AdmissionWebhookConfigurations().Informer()
	hookFilter.adLister = factory.Bkbcs().V2().AdmissionWebhookConfigurations().Lister()
	blog.Infof("AdmissionWebhookFilter SharedInformerFactory start...")
	factory.Start(stopCh)
	// Wait for all caches to sync.
	factory.WaitForCacheSync(stopCh)
	hookFilter.syncAdmissionWebhooks()

	hookFilter.adInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    hookFilter.addNodeToCache,
			UpdateFunc: hookFilter.updateNodeInCache,
			DeleteFunc: hookFilter.deleteNodeFromCache,
		},
	)
	blog.Infof("AdmissionWebhookFilter sync data to cache done")

	return hookFilter, nil
}

func (hook *AdmissionWebhookFilter) addNodeToCache(obj interface{}) {
	ad, ok := obj.(*v2.AdmissionWebhookConfiguration)
	if !ok {
		blog.Errorf("cannot convert to *v2.Application: %v", obj)
		return
	}
	hook.addAdmissionWebhook(&ad.Spec.AdmissionWebhookConfiguration)
}

func (hook *AdmissionWebhookFilter) updateNodeInCache(oldObj, newObj interface{}) {
	ad, ok := newObj.(*v2.AdmissionWebhookConfiguration)
	if !ok {
		blog.Errorf("cannot convert to *v2.Application: %v", newObj)
		return
	}
	hook.addAdmissionWebhook(&ad.Spec.AdmissionWebhookConfiguration)
}

func (hook *AdmissionWebhookFilter) deleteNodeFromCache(obj interface{}) {
	ad, ok := obj.(*v2.AdmissionWebhookConfiguration)
	if !ok {
		blog.Errorf("cannot convert to *v2.Application: %v", obj)
		return
	}
	hook.delAdmissionWebhook(ad.Name)
}

type Meta struct {
	commtypes.TypeMeta   `json:",inline"`
	commtypes.ObjectMeta `json:"metadata"`
}

func (hook *AdmissionWebhookFilter) Execute(req *restful.Request) (int, error) {
	body, err := ioutil.ReadAll(req.Request.Body)
	if err != nil || len(body) == 0 {
		return 0, nil
	}

	var meta *Meta
	err = json.Unmarshal(body, &meta)
	if err != nil {
		blog.V(3).Infof("AdmissionWebhookFilter handler url %s method %s Unmarshal data error %s, and return",
			req.Request.RequestURI, req.Request.Method, err.Error())
		req.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		return 0, nil
	}

	var operation commtypes.AdmissionOperation
	switch req.Request.Method {
	case http.MethodPost:
		operation = commtypes.AdmissionOperationCreate
	case http.MethodPut:
		operation = commtypes.AdmissionOperationUpdate
	default:
		blog.V(3).Infof("AdmissionWebhookFilter handler url %s method %s is invalid, and return",
			req.Request.RequestURI, req.Request.Method)
		req.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		return 0, nil
	}

	var metaKind commtypes.AdmissionResourcesKind
	switch meta.Kind {
	case commtypes.BcsDataType_APP:
		metaKind = commtypes.AdmissionResourcesApplication
	case commtypes.BcsDataType_DEPLOYMENT:
		metaKind = commtypes.AdmissionResourcesDeployment
	default:
		blog.V(3).Infof("AdmissionWebhookFilter handler url %s metaKind %s is invalid, and return",
			req.Request.RequestURI, meta.Kind)
		req.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		return 0, nil
	}

	matchWebhooks := make([]*commtypes.AdmissionWebhookConfiguration, 0)
	hook.RLock()
	for _, admissionHook := range hook.admissionHooks {
		ref := admissionHook.ResourcesRef
		if operation != ref.Operation || metaKind != ref.Kind {
			blog.V(3).Infof("AdmissionWebhook %s don't match %s(%s:%s)", admissionHook.Name, meta.Kind,
				meta.NameSpace, meta.Name)
			continue
		}
		matchWebhooks = append(matchWebhooks, admissionHook)
	}
	hook.RUnlock()

	newBody := body
	for _, admissionHook := range matchWebhooks {
		blog.Infof("AdmissionWebhook %s match %s(%s:%s), and execute webhook",
			admissionHook.Name, meta.Kind, meta.NameSpace, meta.Name)
		for _, webhook := range admissionHook.AdmissionWebhooks {
			if webhook.NamespaceSelector != nil && !webhook.NamespaceSelector.CheckSelector(meta.NameSpace) {
				blog.Infof("admissionwebhook %s webhook %s NamespaceSelector namespace(%s) not match, then continue",
					admissionHook.Name, webhook.Name, meta.NameSpace)
				continue
			}

			hookBody, err := hook.requestAdmissionWebhook(webhook, newBody)
			if err != nil {
				blog.Errorf("admissionwebhook %s request webhoook %s error %s", admissionHook.Name, webhook.Name, err.Error())
				if webhook.FailurePolicy == commtypes.WebhookFailurePolicyFail {
					blog.Errorf("AdmissionWebhookFilter handler url %s method %s failed, and policy fail return",
						req.Request.RequestURI, req.Request.Method)
					return common.BcsErrMesosDriverHttpFilterFailed, fmt.Errorf("request webhoook %s error %s", webhook.Name, err.Error())
				}
				blog.Infof("AdmissionWebhookFilter handler url %s method %s failed, and policy ignore continue",
					req.Request.RequestURI, req.Request.Method)
				continue
			}
			blog.V(3).Infof("admissionwebhook %s webhook %s handle object(%s:%s) success",
				admissionHook.Name, webhook.Name, meta.NameSpace, meta.Name)

			newBody = hookBody
		}
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

func (hook *AdmissionWebhookFilter) syncAdmissionWebhooks() {
	admissions, err := hook.adLister.List(labels.Everything())
	if err != nil {
		blog.Errorf("AdmissionWebhookFilter fetch all admission webhooks error %s", err.Error())
		return
	}

	for _, admission := range admissions {
		ad := admission.Spec.AdmissionWebhookConfiguration
		hook.addAdmissionWebhook(&ad)
	}
}

func (hook *AdmissionWebhookFilter) addAdmissionWebhook(ad *commtypes.AdmissionWebhookConfiguration) {
	hook.Lock()
	defer hook.Unlock()

	if len(ad.AdmissionWebhooks) == 0 {
		blog.Errorf("AdmissionWebhookConfiguration %s have no AdmissionWebhooks, and return", ad.Name)
		return
	}
	//get webhook servers info
	for _, webhook := range ad.AdmissionWebhooks {
		var server string
		if webhook.ClientConfig.Url != "" {
			server = webhook.ClientConfig.Url
		} else {
			if webhook.ClientConfig.Port <= 0 {
				webhook.ClientConfig.Port = 443
			}
			if webhook.ClientConfig.Path == "" {
				webhook.ClientConfig.Path = "/"
			}
			server = fmt.Sprintf("https://%s.%s:%d%s", webhook.ClientConfig.Name, webhook.ClientConfig.Namespace,
				webhook.ClientConfig.Port, webhook.ClientConfig.Path)
		}

		blog.Infof("AdmissionWebhookFilter add AdmissionWebhook %s Name %s server %s", ad.Name, webhook.Name, server)
		webhook.WebhookServers = []string{server}
	}

	hook.admissionHooks[ad.Name] = ad
}

func (hook *AdmissionWebhookFilter) delAdmissionWebhook(name string) {
	hook.Lock()
	delete(hook.admissionHooks, name)
	hook.Unlock()
}
