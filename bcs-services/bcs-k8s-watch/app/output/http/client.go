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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

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
	// NamespaceScopeURLFmt xxx
	// bcsstorage/v1/k8s/dynamic/namespace_resources/clusters/{clusterId}/namespaces/{namespace}/{resourceType}/{resourceName}
	NamespaceScopeURLFmt = "%s/bcsstorage/v1/k8s/dynamic/namespace_resources/clusters/%s/namespaces/%s/%s/%s"
	// HandlerGetNamespaceName xxx
	// handler namespace type name resource
	HandlerGetNamespaceName = "k8s_cluster_namespace_type_name"

	// ListNamespaceScopeURLFmt xxx
	// bcsstorage/v1/k8s/dynamic/namespace_resources/clusters/{clusterId}/namespaces/{namespace}/{resourceType}
	ListNamespaceScopeURLFmt = "%s/bcsstorage/v1/k8s/dynamic/namespace_resources/clusters/%s/namespaces/%s/%s"
	// HandlerListNamespaceName xxx
	// handler namespace type resource
	HandlerListNamespaceName = "k8s_cluster_namespace_type"

	// ClusterScopeURLFmt xxx
	// bcsstorage/v1/k8s/dynamic/cluster_resources/clusters/{clusterId}/{resourceType}/{resourceName}
	ClusterScopeURLFmt = "%s/bcsstorage/v1/k8s/dynamic/cluster_resources/clusters/%s/%s/%s"
	// HandlerGetClusterName xxx
	// handler cluster type resource
	HandlerGetClusterName = "k8s_cluster_type_name"

	// ListClusterScopeURLFmt xxx
	// bcsstorage/v1/k8s/dynamic/cluster_resources/clusters/{clusterId}/{resourceType}
	ListClusterScopeURLFmt = "%s/bcsstorage/v1/k8s/dynamic/cluster_resources/clusters/%s/%s"
	// HandlerListClusterName xxx
	// handler cluster resource
	HandlerListClusterName = "k8s_cluster_type"

	// EventScopeURLFmt xxx
	// event url
	EventScopeURLFmt = "%s/bcsstorage/v1/events"
	// HandlerEventName xxx
	// handler event name
	HandlerEventName = "events"

	// StorageRequestTimeoutSeconds xxx
	// request timeout
	StorageRequestTimeoutSeconds = 5

	// StorageRequestLimit is max entries of request.
	StorageRequestLimit = 500

	// NamespaceScopeWatchURLFmt xxx
	// bcsstorage/v1/k8s/watch/clusters/{clusterId}/namespaces/{namespace}/{resourceType}/{resourceName}
	NamespaceScopeWatchURLFmt = "%s/bcsstorage/v1/k8s/watch/clusters/%s/namespaces/%s/%s/%s"
	// HandlerWatchNamespaceName xxx
	// handler watch namespace name
	HandlerWatchNamespaceName = "k8s_watch_cluster_namespace_type_name"
)

// WatchKindSet xxx
var WatchKindSet = map[string]struct{}{
	"ExportService": {},
}

// StorageClient is http client for storage services.
type StorageClient struct {
	HTTPClientConfig *bcs.HTTPClientConfig
	ClusterID        string
	Namespace        string
	ResourceType     string
	ResourceName     string
}

// StorageResponse is response of storage services.
type StorageResponse struct {
	Result  bool        `json:"result"`
	Code    int64       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// StorageRequestBody is request body of storage services.
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
				NamespaceScopeURLFmt, client.HTTPClientConfig.URL, client.ClusterID, client.Namespace, client.ResourceType,
				client.ResourceName),
			HandlerGetNamespaceName
	}

	// cluster resource
	return fmt.Sprintf(ClusterScopeURLFmt, client.HTTPClientConfig.URL, client.ClusterID, client.ResourceType,
		client.ResourceName), HandlerGetClusterName
}

// GetBody get body
func (client *StorageClient) GetBody(data interface{}) (interface{}, error) {
	if client.ResourceType != "Event" {
		body := StorageRequestBody{
			Data: data,
		}
		return body, nil
	}

	// not event
	// convert to unstructured object
	dataUnstructured, ok := data.(*unstructured.Unstructured)
	if !ok {
		glog.Errorf("Event Convert object to unstructured event fail! object is %v", data)
		return nil, fmt.Errorf("event report fail. covnvert fail")
	}

	// convert to corev1 object
	event := &v1.Event{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(dataUnstructured.UnstructuredContent(), event)
	if err != nil {
		glog.Errorf("Event Convert object to v1.Event fail! object is %v", dataUnstructured)
		return nil, fmt.Errorf("event report fail. covnvert fail")
	}

	eventTime := time.Time{}

	if !event.LastTimestamp.IsZero() {
		eventTime = event.LastTimestamp.Time
	} else if !event.FirstTimestamp.IsZero() {
		eventTime = event.FirstTimestamp.Time
	} else if !event.EventTime.IsZero() {
		eventTime = event.EventTime.Time
	} else if !event.CreationTimestamp.IsZero() {
		eventTime = event.CreationTimestamp.Time
	}

	return types.BcsStorageEventIf{
		ID:        "",
		Env:       "k8s",
		Kind:      types.EventKind(event.InvolvedObject.Kind),
		Level:     types.EventLevel(event.Type),
		Component: types.EventComponent(event.Source.Component),
		Type:      event.Reason,
		Describe:  event.Message,
		ClusterId: client.ClusterID,
		EventTime: eventTime.Unix(),
		ExtraInfo: types.EventExtraInfo{
			Namespace: event.InvolvedObject.Namespace,
			Name:      event.InvolvedObject.Name,
			Kind:      types.ExtraKind(event.InvolvedObject.Kind),
		},
		Data: data,
	}, nil

}

// NewRequest new request
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

// GET get
func (client *StorageClient) GET() (storageResp StorageResponse, err error) {
	start := time.Now()
	status := metrics.SucStatus
	// timeout
	url, handlerName := client.GetURL()

	request, err := client.NewRequest()
	if err != nil {
		status = metrics.ErrStatus
		metrics.ReportK8sWatchAPIMetrics(client.ClusterID, handlerName, client.ResourceType,
			http.MethodGet, status, start)
		return
	}
	resp, _, errs := request.
		Timeout(StorageRequestTimeoutSeconds*time.Second).
		Get(url).
		Retry(2, 2*time.Second, http.StatusBadRequest, http.StatusInternalServerError).
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

	metrics.ReportK8sWatchAPIMetrics(client.ClusterID, handlerName, client.ResourceType,
		http.MethodGet, status, start)
	return
}

// DELETE delete
func (client *StorageClient) DELETE() (storageResp StorageResponse, err error) {
	start := time.Now()
	status := metrics.SucStatus
	// timeout retry
	url, handlerName := client.GetURL()

	request, err := client.NewRequest()
	if err != nil {
		status = metrics.ErrStatus
		metrics.ReportK8sWatchAPIMetrics(client.ClusterID, handlerName, client.ResourceType,
			http.MethodDelete, status, start)
		return
	}
	resp, _, errs := request.
		Timeout(StorageRequestTimeoutSeconds*time.Second).
		Delete(url).
		Retry(3, 1*time.Second, http.StatusBadRequest, http.StatusInternalServerError).
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

	metrics.ReportK8sWatchAPIMetrics(client.ClusterID, handlerName, client.ResourceType,
		http.MethodDelete, status, start)
	return
}

// PUT put
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
		// todo 这里报错
		status = metrics.ErrStatus
		metrics.ReportK8sWatchAPIMetrics(client.ClusterID, handlerName, client.ResourceType,
			http.MethodPut, status, start)
		return
	}

	resp, _, errs := request.
		Timeout(StorageRequestTimeoutSeconds*time.Second).
		Put(url).
		Send(body).
		Retry(3, 1*time.Second, http.StatusBadRequest, http.StatusInternalServerError).
		EndStruct(&storageResp)

	if !storageResp.Result || errs != nil {
		debugBody, err := jsoniter.Marshal(body)
		if err != nil {
			glog.Errorf("method=PUT url=%s, body=%v, errors=%s, resp=%v, storageResp=%v", url, body, errs, resp, storageResp)
		} else {
			glog.Debug(fmt.Sprintf("method=PUT url=%s, body=%s, errors=%s, resp=%v, storageResp=%v", url, string(debugBody),
				errs, resp, storageResp))
		}
		status = metrics.ErrStatus
	}

	if errs != nil {
		glog.Errorf("PUT fail: [url=%s, data=%s, resp=%s, errors=%s]", url, body, resp, errs)
		err = errors.New("HTTP error")
		status = metrics.ErrStatus
	}

	metrics.ReportK8sWatchAPIMetrics(client.ClusterID, handlerName, client.ResourceType,
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
		metrics.ReportK8sWatchAPIMetrics(client.ClusterID, handlerName, client.ResourceType,
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

	if storageResp.Data != nil {
		d, ok := storageResp.Data.([]interface{})
		if !ok {
			status = metrics.ErrStatus
			err = fmt.Errorf("listResource interface conversion error! [url=%s, resp=%v, errs=%s]", url, resp, errs)
			return
		}
		data = d
	}
	return
}

// ListNamespaceResource list namespace resource
func (client *StorageClient) ListNamespaceResource() (data []interface{}, err error) {
	return client.ListNamespaceResourceWithLabelSelector("")
}

// ListNamespaceResourceWithLabelSelector list namespace resource with label selector
func (client *StorageClient) ListNamespaceResourceWithLabelSelector(labelSelector string) (
	[]interface{}, error) {
	const (
		handlerName = HandlerListNamespaceName
	)
	urlWithParams := ""
	if client.ResourceType == ResourceTypeEvent {
		url, _ := client.GetURL()
		now := time.Now()
		duration := time.Duration(1) * time.Hour // event will disappear after 1 hour
		urlWithParams = fmt.Sprintf(
			"%s?clusterId=%s&field=data.metadata.name&timeBegin=%d&timeEnd=%d&extraInfo.namespace=%s",
			url, client.ClusterID, now.Add(-duration).Unix(), now.Unix(), client.Namespace)
	} else {
		url := fmt.Sprintf(ListNamespaceScopeURLFmt,
			client.HTTPClientConfig.URL, client.ClusterID, client.Namespace, client.ResourceType)

		urlWithParams = fmt.Sprintf("%s?field=resourceName", url)
	}

	if len(labelSelector) != 0 {
		selectorMap, err := parseSelectors(labelSelector, "data.metadata.labels.")
		if err != nil {
			return nil, err
		}
		for labelKey, labelValue := range selectorMap {
			urlWithParams = fmt.Sprintf("%s&%s=%s", urlWithParams, labelKey, labelValue)
		}
	}

	offset := 0
	var data []interface{}
	for {
		urlWithLimit := ""
		if client.ResourceType == ResourceTypeEvent {
			urlWithLimit = fmt.Sprintf("%s&length=%d&offset=%d", urlWithParams, StorageRequestLimit, offset)
		} else {
			urlWithLimit = fmt.Sprintf("%s&limit=%d&offset=%d", urlWithParams, StorageRequestLimit, offset)
		}

		glog.V(2).Infof("sync call list namespace resource: %s", urlWithLimit)
		dataTmp, err := client.listResource(urlWithLimit, handlerName)
		if err != nil {
			glog.Errorf("list namespace resource fail: %v", err)
			return nil, err
		}
		data = append(data, dataTmp...)
		if len(dataTmp) == StorageRequestLimit {
			offset += StorageRequestLimit
			continue
		}
		break
	}

	if client.ResourceType == ResourceTypeEvent {
		for i := range data {
			data[i].(map[string]interface{})["resourceName"] =
				data[i].(map[string]interface{})["data"].(map[string]interface{})["metadata"].(map[string]interface{})["name"]
			delete(data[i].(map[string]interface{}), "data")
		}
	}

	return data, nil
}

// ListClusterResource list cluster resource
func (client *StorageClient) ListClusterResource() (data []interface{}, err error) {
	return client.ListClusterResourceWithLabelSelector("")
}

// ListClusterResourceWithLabelSelector list cluster resource with label selector
func (client *StorageClient) ListClusterResourceWithLabelSelector(labelSelector string) ([]interface{}, error) {
	const (
		handlerName = HandlerListClusterName
	)
	url := fmt.Sprintf(ListClusterScopeURLFmt,
		client.HTTPClientConfig.URL, client.ClusterID, client.ResourceType)

	urlWithParams := fmt.Sprintf("%s?field=resourceName", url)
	if len(labelSelector) != 0 {
		selectorMap, err := parseSelectors(labelSelector, "data.metadata.labels.")
		if err != nil {
			return nil, err
		}
		for labelKey, labelValue := range selectorMap {
			urlWithParams = fmt.Sprintf("%s&%s=%s", urlWithParams, labelKey, labelValue)
		}
	}

	glog.V(2).Infof("sync call list cluster resource: %s", urlWithParams)
	return client.listResource(urlWithParams, handlerName)
}
