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

package app

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	glog "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/options"
)

// HealthChecker check k8s cluster healthz
type HealthChecker struct {
	k8sConfig *options.K8sConfig
}

func (h *HealthChecker) newRequest() (*gorequest.SuperAgent, error) {

	k8sConfig := h.k8sConfig
	u, err := url.Parse(k8sConfig.Master)
	if err != nil {
		panic(fmt.Errorf("invalid master url:%v", err))
	}

	request := gorequest.New()
	if u.Scheme == "https" {
		tlsConfig, err2 := ssl.ClientTslConfVerity(
			k8sConfig.TLS.CAFile,
			k8sConfig.TLS.CertFile,
			k8sConfig.TLS.KeyFile,
			"",
		)
		if err2 != nil {
			return nil, fmt.Errorf("init tls fail [clientConfig=%v, errors=%s]", tlsConfig, err2)
		}
		request = request.TLSClientConfig(tlsConfig)

	}
	return request, nil
}

func (h *HealthChecker) checkAPIServer(url string) bool {

	request, err := h.newRequest()
	if err != nil {
		glog.Errorf("healthCheck: create request fail! err=%s", err)
		return false
	}

	resp, bodyBytes, errs := request.Timeout(5 * time.Second).Get(url).EndBytes()
	if errs != nil {
		glog.Errorf("healthCheck: GET fail [url=%s, resp=%v, errors=%s]", url, resp, errs)
		return false
	}

	if string(bodyBytes) == "ok" {
		glog.Errorf("healthCheck: url=%s return not ok! err=%s", url, err)
		return true
	}
	return false
}

func (h *HealthChecker) healthz(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("%s/health", h.k8sConfig.Master)

	// current, only check healthz
	ok := h.checkAPIServer(url)
	if ok {
		fmt.Fprintf(w, "ok")
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - the APIServer is not available!"))
	}
}

// Run start http server
func (h *HealthChecker) Run(stopChan <-chan struct{}) {
	addr := "0.0.0.0:8000"
	http.HandleFunc("/healthz/", h.healthz)

	go func() {
		glog.Infof("Listening on http://%s\n", addr)

		if err := http.ListenAndServe(addr, nil); err != nil {
			glog.Fatal(err)
		}
	}()

	<-stopChan
	glog.Info("Health Check got exit signal, ready to exit!")
}
