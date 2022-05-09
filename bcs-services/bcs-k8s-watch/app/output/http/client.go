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

package http

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/parnurzeal/gorequest"
	v1 "k8s.io/api/core/v1"

	glog "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/pkg/metrics"
)

// Client is http client for inner services.
type Client interface {
	// GetURL returns target url.
	GetURL()

	// GET is http restfull GET method.
	GET()

	// DELETE is http restfull DELETE method.
	DELETE()

	// PUT is http restfull PUT method.
	PUT()
}

const (
	// ResourceTypeEvent is resource type of event.
	ResourceTypeEvent = "Event"
	// bcsstorage/v1/k8s/dynamic/namespace_resources/clusters/{clusterId}/namespaces/{namespace}/{resourceType}/{resourceName}
	NamespaceScopeURLFmt = "%s/bcsstorage/v1/k8s/dynamic/namespace_resources/clusters/%s/namespaces/%s/%s/%s"
	// handler namespace type name resource
	HandlerGetNamespaceName = "k8s_cluster_namespace_type_name"

	// bcsstorage/v1/k8s/dynamic/namespace_resources/clusters/{clusterId}/namespaces/{namespace}/{resourceType}
	ListNamespaceScopeURLFmt = "%s/bcsstorage/v1/k8s/dynamic/namespace_resources/clusters/%s/namespaces/%s/%s"
	// handler namespace type resource
	HandlerListNamespaceName = "k8s_cluster_namespace_type"

	// bcsstorage/v1/k8s/dynamic/cluster_resources/clusters/{clusterId}/{resourceType}/{resourceName}
	ClusterScopeURLFmt = "%s/bcsstorage/v1/k8s/dynamic/cluster_resources/clusters/%s/%s/%s"
	// handler cluster type resource
	HandlerGetClusterName = "k8s_cluster_type_name"

	// bcsstorage/v1/k8s/dynamic/cluster_resources/clusters/{clusterId}/{resourceType}
	ListClusterScopeURLFmt = "%s/bcsstorage/v1/k8s/dynamic/cluster_resources/clusters/%s/%s"
	// handler cluster resource
	HandlerListClusterName = "k8s_cluster_type"

	// event url
	EventScopeURLFmt = "%s/bcsstorage/v1/events"
	// handler event name
	HandlerEventName = "events"

	// request timeout
	StorageRequestTimeoutSeconds = 5

	// bcsstorage/v1/k8s/watch/clusters/{clusterId}/namespaces/{namespace}/{resourceType}/{resourceName}
	NamespaceScopeWatchURLFmt = "%s/bcsstorage/v1/k8s/watch/clusters/%s/namespaces/%s/%s/%s"
	// handler watch namespace name
	HandlerWatchNamespaceName = "k8s_watch_cluster_namespace_type_name"
)

var WatchKindSet = map[string]struct{}{
	"ExportService": {},
}

type StorageClient struct {
	HTTPClientConfig *bcs.HTTPClientConfig
	ClusterID        string
	Namespace        string
	ResourceType     string
	ResourceName     string
}

type StorageResponse struct {
	Result  bool        `json:"result"`
	Code    int64       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type StorageRequestBody struct {
	Data interface{} `json:"data"`
}

// GetURL get url
func (client *StorageClient) GetURL() (string, string) {
	// Event
	if client.ResourceType == "Event" {
		return fmt.Sprintf(EventScopeURLFmt, client.HTTPClientConfig.URL), HandlerEventName
	}

	if _, ok := WatchKindSet[client.ResourceType]; ok {
		return fmt.Sprintf(
			NamespaceScopeWatchURLFmt,
			client.HTTPClientConfig.URL,
			client.ClusterID,
			client.Namespace,
			strings.ToLower(client.ResourceType),
			client.ResourceName), HandlerWatchNamespaceName
	}

	// namespace resource
	if client.Namespace != "" {
		return fmt.Sprintf(
				NamespaceScopeURLFmt, client.HTTPClientConfig.URL, client.ClusterID, client.Namespace, client.ResourceType, client.ResourceName),
			HandlerGetNamespaceName
	}

	// cluster resource
	return fmt.Sprintf(ClusterScopeURLFmt, client.HTTPClientConfig.URL, client.ClusterID, client.ResourceType, client.ResourceName), HandlerGetClusterName
}

func (client *StorageClient) GetBody(data interface{}) (interface{}, error) {
	if client.ResourceType != "Event" {
		body := StorageRequestBody{
			Data: data,
		}
		return body, nil
	}

	// not event
	event, ok := data.(*v1.Event)
	if !ok {
		glog.Errorf("Event Convert object to v1.Event fail! object is %v", data)
		return nil, errors.New("event report fail. covnvert fail")
	}

	return types.BcsStorageEventIf{
		Env:       "k8s",
		Kind:      types.EventKind(event.InvolvedObject.Kind),
		Level:     types.EventLevel(event.Type),
		Component: types.EventComponent(event.Source.Component),
		Type:      event.Reason,
		Describe:  event.Message,
		ClusterId: client.ClusterID,
		EventTime: event.LastTimestamp.Unix(),
		ExtraInfo: types.EventExtraInfo{
			Namespace: event.InvolvedObject.Namespace,
			Name:      event.InvolvedObject.Name,
		},
		Data: data,
	}, nil

}

func (client *StorageClient) NewRequest() (*gorequest.SuperAgent, error) {
	request := gorequest.New()

	httpConfig := client.HTTPClientConfig

	// handler tls
	if httpConfig.Scheme == "https" {
		tlsConfig, err2 := ssl.ClientTslConfVerity(
			httpConfig.CAFile,
			httpConfig.CertFile,
			httpConfig.KeyFile,
			httpConfig.Password)
		if err2 != nil {
			return nil, fmt.Errorf("init tls fail [clientConfig=%v, errors=%s]", tlsConfig, err2)
		}
		request = request.TLSClientConfig(tlsConfig)
	}
	return request, nil
}

func (client *StorageClient) GET() (storageResp StorageResponse, err error) {
	start := time.Now()
	status := metrics.SucStatus
	// timeout
	url, handlerName := client.GetURL()

	request, err := client.NewRequest()
	if err != nil {
		status = metrics.ErrStatus
		metrics.ReportK8sWatchAPIMetrics(client.ClusterID, handlerName, client.Namespace, client.ResourceType,
			http.MethodGet, status, start)
		return
	}
	resp, _, errs := request.
		Timeout(StorageRequestTimeoutSeconds * time.Second).
		Get(url).
		EndStruct(&storageResp)

	if !storageResp.Result {
		glog.Debug(fmt.Sprintf("method=GET url=%s, resp=%v", url, storageResp))
		status = metrics.ErrStatus
	}

	if errs != nil {
		glog.Errorf("GET fail: [url=%s, resp=%v, errors=%s]", url, resp, errs)
		err = errors.New("HTTP error")
		status = metrics.ErrStatus
	}

	metrics.ReportK8sWatchAPIMetrics(client.ClusterID, handlerName, client.Namespace, client.ResourceType,
		http.MethodGet, status, start)
	return
}

func (client *StorageClient) DELETE() (storageResp StorageResponse, err error) {
	start := time.Now()
	status := metrics.SucStatus
	// timeout retry
	url, handlerName := client.GetURL()

	request, err := client.NewRequest()
	if err != nil {
		status = metrics.ErrStatus
		metrics.ReportK8sWatchAPIMetrics(client.ClusterID, handlerName, client.Namespace, client.ResourceType,
			http.MethodDelete, status, start)
		return
	}
	resp, _, errs := request.
		Timeout(StorageRequestTimeoutSeconds*time.Second).
		Delete(url).
		Retry(2, 1*time.Second, http.StatusBadRequest, http.StatusInternalServerError).
		EndStruct(&storageResp)

	if !storageResp.Result {
		glog.Debug(fmt.Sprintf("method=DELETE, url is %s, all response: %v", url, storageResp))
		status = metrics.ErrStatus
	}

	if errs != nil {
		glog.Errorf("DELETE fail: [url=%s, resp=%v, errors=%s]", url, resp, errs)
		err = errors.New("HTTP error")
		status = metrics.ErrStatus
	}

	metrics.ReportK8sWatchAPIMetrics(client.ClusterID, handlerName, client.Namespace, client.ResourceType,
		http.MethodDelete, status, start)
	return
}

func (client *StorageClient) PUT(data interface{}) (storageResp StorageResponse, err error) {
	start := time.Now()
	status := metrics.SucStatus
	url, handlerName := client.GetURL()

	request, err := client.NewRequest()
	if err != nil {
		return
	}

	body, err := client.GetBody(data)
	if err != nil {
		status = metrics.ErrStatus
		metrics.ReportK8sWatchAPIMetrics(client.ClusterID, handlerName, client.Namespace, client.ResourceType,
			http.MethodPut, status, start)
		return
	}

	resp, _, errs := request.
		Timeout(StorageRequestTimeoutSeconds*time.Second).
		Put(url).
		Send(body).
		Retry(2, 1*time.Second, http.StatusBadRequest, http.StatusInternalServerError).
		EndStruct(&storageResp)

	if !storageResp.Result || errs != nil {
		debugBody, err := jsoniter.Marshal(body)
		if err != nil {
			glog.Errorf("method=PUT url=%s, body=%v, errors=%s, resp=%v, storageResp=%v", url, body, errs, resp, storageResp)
		} else {
			glog.Debug(fmt.Sprintf("method=PUT url=%s, body=%s, errors=%s, resp=%v, storageResp=%v", url, string(debugBody), errs, resp, storageResp))
		}
		status = metrics.ErrStatus
	}

	if errs != nil {
		glog.Errorf("PUT fail: [url=%s, data=%s, resp=%s, errors=%s]", url, body, resp, errs)
		err = errors.New("HTTP error")
		status = metrics.ErrStatus
	}

	metrics.ReportK8sWatchAPIMetrics(client.ClusterID, handlerName, client.Namespace, client.ResourceType,
		http.MethodPut, status, start)
	return
}

func (client *StorageClient) listResource(url string, handlerName string) (data []interface{}, err error) {
	var (
		start       = time.Now()
		storageResp = StorageResponse{}
		status      = metrics.SucStatus
	)

	defer func() {
		metrics.ReportK8sWatchAPIMetrics(client.ClusterID, handlerName, client.Namespace, client.ResourceType,
			http.MethodGet, status, start)
	}()

	request, err := client.NewRequest()
	if err != nil {
		status = metrics.ErrStatus
		return
	}

	resp, _, errs := request.
		Timeout(StorageRequestTimeoutSeconds * time.Second).
		Get(url).
		EndStruct(&storageResp)

	if !storageResp.Result {
		status = metrics.ErrStatus
		err = fmt.Errorf("listResource result=false [url=%s, resp=%v, storageResp=%v]", url, resp, storageResp)
		return
	}

	if errs != nil {
		status = metrics.ErrStatus
		err = fmt.Errorf("listResource do GET fail! [url=%s, resp=%v, errs=%s]", url, resp, errs)
		return
	}

	data = storageResp.Data.([]interface{})
	return
}

func (client *StorageClient) ListNamespaceResource() (data []interface{}, err error) {
	const (
		handlerName = HandlerListNamespaceName
	)
	url := fmt.Sprintf(ListNamespaceScopeURLFmt,
		client.HTTPClientConfig.URL, client.ClusterID, client.Namespace, client.ResourceType)

	urlWithParams := fmt.Sprintf("%s?field=resourceName", url)

	if client.ResourceType == ResourceTypeEvent {
		url, _ = client.GetURL()
		now := time.Now()
		duration := time.Duration(1) * time.Hour // event will disappear after 1 hour
		urlWithParams = fmt.Sprintf(
			"%s?clusterId=%s&field=data.metadata.name&timeBegin=%d&timeEnd=%d&extra={\"namespace\":\"%s\"}",
			url, client.ClusterID, now.Add(-duration).Unix(), now.Unix(), client.Namespace)
	}

	glog.V(2).Infof("sync call list namespace resource: %s", urlWithParams)

	data, err = client.listResource(urlWithParams, handlerName)
	if client.ResourceType == ResourceTypeEvent {
		for i := range data {
			data[i].(map[string]interface{})["resourceName"] = data[i].(map[string]interface{})["data"].(map[string]interface{})["metadata"].(map[string]interface{})["name"]
			delete(data[i].(map[string]interface{}), "data")
		}
	}
	return
}

func (client *StorageClient) ListClusterResource() (data []interface{}, err error) {
	const (
		handlerName = HandlerListClusterName
	)
	url := fmt.Sprintf(ListClusterScopeURLFmt,
		client.HTTPClientConfig.URL, client.ClusterID, client.ResourceType)

	urlWithParams := fmt.Sprintf("%s?field=resourceName", url)

	glog.V(2).Infof("sync call list cluster resource: %s", urlWithParams)
	data, err = client.listResource(urlWithParams, handlerName)
	return
}
