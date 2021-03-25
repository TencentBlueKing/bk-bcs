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

package etcd

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/parnurzeal/gorequest"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	netservicetypes "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/netservice"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/cluster"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/service"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/types"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/util"
)

const (
	// defaultSyncInterval is default sync interval.
	defaultSyncInterval = 60 * time.Second

	// defaultNetServiceTimeout is default netservice timeout.
	defaultNetServiceTimeout = 20 * time.Second

	// defaultHTTPRetryerCount is default http request retry count.
	defaultHTTPRetryerCount = 2

	// defaultHTTPRetryerTime is default http request retry time.
	defaultHTTPRetryerTime = 3 * time.Second
)

func reportIPPoolStaticMetrics(clusterID, action, status string) {
	util.ReportSyncTotal(clusterID, cluster.DataTypeIPPoolStatic, action, status)
}

func reportIPPoolStaticDetailMetrics(clusterID, action, status string) {
	util.ReportSyncTotal(clusterID, cluster.DataTypeIPPoolStaticDetail, action, status)
}

// NetServiceWatcher watchs resources in netservice, and sync to storage.
type NetServiceWatcher struct {
	clusterID  string
	report     cluster.Reporter
	netservice *service.InnerService
}

// NewNetServiceWatcher return a new netservice watcher.
func NewNetServiceWatcher(clusterID string, reporter cluster.Reporter, netservice *service.InnerService) *NetServiceWatcher {
	return &NetServiceWatcher{
		clusterID:  clusterID,
		report:     reporter,
		netservice: netservice,
	}
}

func (w *NetServiceWatcher) httpClient(httpConfig *service.HTTPClientConfig) (*gorequest.SuperAgent, error) {
	request := gorequest.New().Set("Accept", "application/json").Set("BCS-ClusterID", w.clusterID)

	if httpConfig.Scheme == "https" {
		tlsConfig, err := ssl.ClientTslConfVerity(httpConfig.CAFile, httpConfig.CertFile,
			httpConfig.KeyFile, httpConfig.Password)

		if err != nil {
			return nil, fmt.Errorf("init tls fail [clientConfig=%v, errors=%s]", tlsConfig, err)
		}
		request = request.TLSClientConfig(tlsConfig)
	}

	return request, nil
}

func (w *NetServiceWatcher) queryIPResource() (*netservicetypes.NetResponse, error) {
	targets := w.netservice.Servers()
	serversCount := len(targets)

	if serversCount == 0 {
		return nil, errors.New("netservice server list is empty, there is no available services now")
	}

	var httpClientConfig *service.HTTPClientConfig
	if serversCount == 1 {
		httpClientConfig = targets[0]
	} else {
		index := rand.Intn(serversCount)
		httpClientConfig = targets[index]
	}

	request, err := w.httpClient(httpClientConfig)
	if err != nil {
		return nil, fmt.Errorf("can't create netservice client, %+v, %+v", httpClientConfig, err)
	}

	url := fmt.Sprintf("%s/v1/pool/%s", httpClientConfig.URL, w.clusterID)
	response := &netservicetypes.NetResponse{}

	if _, _, err := request.
		Timeout(defaultNetServiceTimeout).
		Get(url).
		Retry(defaultHTTPRetryerCount, defaultHTTPRetryerTime, http.StatusBadRequest, http.StatusInternalServerError).
		EndStruct(response); err != nil {
		return nil, fmt.Errorf("request to netservice, get ip resource failed, %+v", err)
	}

	if response.Code != 0 {
		return nil, fmt.Errorf("request to netservice, get ip resource failed, code[%d], message[%s]",
			response.Code, response.Message)
	}
	return response, nil
}

func (w *NetServiceWatcher) queryIPResourceDetail() (*netservicetypes.NetResponse, error) {
	targets := w.netservice.Servers()
	serversCount := len(targets)

	if serversCount == 0 {
		return nil, errors.New("netservice server list is empty, there is no available services now")
	}

	var httpClientConfig *service.HTTPClientConfig
	if serversCount == 1 {
		httpClientConfig = targets[0]
	} else {
		index := rand.Intn(serversCount)
		httpClientConfig = targets[index]
	}

	request, err := w.httpClient(httpClientConfig)
	if err != nil {
		return nil, fmt.Errorf("can't create netservice client, %+v, %+v", httpClientConfig, err)
	}

	url := fmt.Sprintf("%s/v1/pool/%s?info=detail", httpClientConfig.URL, w.clusterID)
	response := &netservicetypes.NetResponse{}

	if _, _, err := request.
		Timeout(defaultNetServiceTimeout).
		Get(url).
		Retry(defaultHTTPRetryerCount, defaultHTTPRetryerTime, http.StatusBadRequest, http.StatusInternalServerError).
		EndStruct(response); err != nil {
		return nil, fmt.Errorf("request to netservice, get ip resource detail failed, %+v", err)
	}

	if response.Code != 0 {
		return nil, fmt.Errorf("request to netservice, get ip resource detail failed, code[%d], message[%s]",
			response.Code, response.Message)
	}
	return response, nil
}

// SyncIPResource syncs target ip resources to storages.
func (w *NetServiceWatcher) SyncIPResource() {
	// query resource from netservice.
	resource, err := w.queryIPResource()
	if err != nil {
		blog.Warnf("sync netservice ip resource, query from netservice failed, %+v", err)
		return
	}

	// only sync ip pool static information.
	if resource.Type != netservicetypes.ResponseType_PSTATIC {
		blog.Warnf("sync netservice ip resource, query from netservice, invalid response type[%+v]", resource.Type)
		return
	}

	// sync ip resource.
	data := &types.BcsSyncData{
		DataType: cluster.DataTypeIPPoolStatic,
		Action:   types.ActionUpdate,
		Item:     resource.Data,
	}

	if err := w.report.ReportData(data); err != nil {
		reportIPPoolStaticMetrics(w.clusterID, types.ActionUpdate, cluster.SyncFailure)
	} else {
		reportIPPoolStaticMetrics(w.clusterID, types.ActionUpdate, cluster.SyncSuccess)
	}
}

// SyncIPResourceDetail sync target ip resource detail to storages.
func (w *NetServiceWatcher) SyncIPResourceDetail() {
	// query resource detail from netservice.
	resource, err := w.queryIPResourceDetail()
	if err != nil {
		blog.Warnf("sync netservice ip resource detail, query from netservice failed, %+v", err)
		return
	}

	// only sync ip pool detail information.
	if resource.Type != netservicetypes.ResponseType_POOL {
		blog.Warnf("sync netservice ip resource detail, query from netservice, invalid response type[%+v]", resource.Type)
		return
	}

	// sync ip resource detail.
	data := &types.BcsSyncData{
		DataType: cluster.DataTypeIPPoolStaticDetail,
		Action:   types.ActionUpdate,
		Item:     resource.Data,
	}

	if err := w.report.ReportData(data); err != nil {
		reportIPPoolStaticDetailMetrics(w.clusterID, types.ActionUpdate, cluster.SyncFailure)
	} else {
		reportIPPoolStaticDetailMetrics(w.clusterID, types.ActionUpdate, cluster.SyncSuccess)
	}
}

// Run starts the netservice watcher.
func (w *NetServiceWatcher) Run(stopCh <-chan struct{}) {
	// sync ip resource.
	go wait.NonSlidingUntil(w.SyncIPResource, defaultSyncInterval, stopCh)

	// sync ip resource detail.
	go wait.NonSlidingUntil(w.SyncIPResourceDetail, defaultSyncInterval, stopCh)

	// TODO: add more resource-sync logics here.
}
