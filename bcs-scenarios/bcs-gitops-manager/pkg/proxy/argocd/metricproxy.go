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

package argocd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
)

// MetricPlugin defines the metric plugin to proxy all the metrics
type MetricPlugin struct {
	*mux.Router
	middleware mw.MiddlewareInterface

	monitorClient *monitoring.Clientset
	k8sClient     *kubernetes.Clientset
}

// Init will init the metric proxy
func (plugin *MetricPlugin) Init() error {
	plugin.Path("/{namespace}/{servicemonitor}").Methods("GET").
		Handler(plugin.middleware.HttpWrapper(plugin.metric))
	if err := plugin.inClusterClient(); err != nil {
		return errors.Wrapf(err, "init metric proxy plugin failed")
	}
	blog.Infof("metric proxy plugin init successfully")
	return nil
}

func (plugin *MetricPlugin) inClusterClient() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return errors.Wrapf(err, "get k8s incluster config failed")
	}
	plugin.monitorClient, err = monitoring.NewForConfig(config)
	if err != nil {
		return errors.Wrapf(err, "create prometheus client failed")
	}
	plugin.k8sClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrapf(err, "create k8s lient failed")
	}
	return nil
}

func (plugin *MetricPlugin) metric(r *http.Request) (*http.Request, *mw.HttpResponse) {
	namespace, smName, resp := plugin.parseParam(r.Context(), r)
	if resp != nil {
		return r, resp
	}
	serviceMonitor, err := plugin.monitorClient.MonitoringV1().ServiceMonitors(namespace).
		Get(r.Context(), smName, metav1.GetOptions{})
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Wrapf(err, "get service monitor '%s/%s' failed", namespace, smName))
	}
	metricPortPath := make(map[string][]string)
	for _, ep := range serviceMonitor.Spec.Endpoints {
		v, ok := metricPortPath[ep.Port]
		if ok {
			metricPortPath[ep.Port] = append(v, ep.Path)
		} else {
			metricPortPath[ep.Port] = []string{ep.Path}
		}
	}

	labelSelector := labels.SelectorFromSet(serviceMonitor.Spec.Selector.MatchLabels)
	endpoints, err := plugin.k8sClient.CoreV1().Endpoints(namespace).List(r.Context(), metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Wrapf(err, "get endpoints by label '%s' failed", labelSelector.String()))
	}

	result := make([]string, 0)
	for _, ep := range endpoints.Items {
		epMetrics := plugin.buildEndpointsMetrics(smName, metricPortPath, &ep)
		result = append(result, epMetrics...)
	}
	return r, mw.ReturnDirectResponse(strings.Join(result, "\n"))
}

func (plugin *MetricPlugin) buildEndpointsMetrics(smName string, metricPortPath map[string][]string,
	ep *corev1.Endpoints) []string {
	result := make([]string, 0)
	for _, subset := range ep.Subsets {
		portPaths := make(map[int32][]string)
		for _, port := range subset.Ports {
			paths, ok := metricPortPath[port.Name]
			if !ok {
				continue
			}
			portPaths[port.Port] = paths
		}
		subsetMetrics := plugin.buildSubsetsMetricResult(smName, &subset, portPaths)
		result = append(result, subsetMetrics...)
	}
	return result
}

func (plugin *MetricPlugin) buildSubsetsMetricResult(smName string, subset *corev1.EndpointSubset,
	portPaths map[int32][]string) []string {
	subsetMetrics := make([]string, 0)
	for _, addr := range subset.Addresses {
		for port, paths := range portPaths {
			podMetrics := make([]string, 0)
			for _, path := range paths {
				url := fmt.Sprintf("http://%s:%d%s", addr.IP, port, path)
				bs, err := plugin.getMetric(url)
				if err != nil {
					blog.Warnf("Metric[%s] get metric failed for pod '%s': %s", smName, addr.TargetRef.Name)
					continue
				}
				metricRewrite := plugin.rewriteMetric(bs, port, path, &addr)
				podMetrics = append(podMetrics, metricRewrite...)
				blog.Infof("Metric[%s] get metric for pod '%s' with url '%s' success", smName,
					addr.TargetRef.Name, url)
			}
			subsetMetrics = append(subsetMetrics, podMetrics...)
		}
	}
	return subsetMetrics
}

func (plugin *MetricPlugin) getMetric(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "get metric '%s' failed", url)
	}
	defer resp.Body.Close()

	var bs []byte
	bs, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "get metric '%s' read resp body failed", url)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("get metric '%s' resp code not 200 but %d: %s",
			url, resp.StatusCode, string(bs))
	}
	return bs, nil
}

const (
	labelPodName    = "podname"
	labelPodIP      = "podip"
	labelMetricPath = "metricpath"
	labelMetricPort = "metricport"
)

// rewriteMetric 拼接 metric 增加固定标签
func (plugin *MetricPlugin) rewriteMetric(bs []byte, port int32, path string, addr *corev1.EndpointAddress) []string {
	metrics := string(bs)
	arr := strings.Split(metrics, "\n")
	result := make([]string, 0, len(arr))
	for i := range arr {
		if strings.HasPrefix(arr[i], "#") {
			continue
		}
		raw := arr[i]
		if !strings.Contains(raw, "}") {
			metricArr := strings.Split(raw, " ")
			if len(metricArr) != 2 {
				continue
			}
			prefix := metricArr[0] + fmt.Sprintf(`{%s="%s",%s="%s",%s="%s",%s="%d} `,
				labelPodName, addr.TargetRef.Name,
				labelPodIP, addr.IP,
				labelMetricPath, path,
				labelMetricPort, port,
			)
			result = append(result, prefix+metricArr[1])
		} else {
			metricArr := strings.Split(arr[i], "} ")
			if len(metricArr) != 2 {
				continue
			}
			prefix := metricArr[0] + fmt.Sprintf(`,%s="%s",%s="%s",%s="%s",%s="%d"} `,
				labelPodName, addr.TargetRef.Name,
				labelPodIP, addr.IP,
				labelMetricPath, path,
				labelMetricPort, port,
			)
			result = append(result, prefix+metricArr[1])
		}
	}
	return result
}

func (plugin *MetricPlugin) matchPortPath(metricPortPath map[string][]string,
	ports []corev1.EndpointPort) map[int32][]string {
	result := make(map[int32][]string)
	for _, port := range ports {
		paths, ok := metricPortPath[port.Name]
		if !ok {
			continue
		}
		result[port.Port] = paths
	}
	return result
}

func (plugin *MetricPlugin) parseParam(ctx context.Context, r *http.Request) (string, string, *mw.HttpResponse) {
	var namespace, smName string
	user := mw.User(ctx)
	if user.ClientID != proxy.AdminClientUser && user.ClientID != proxy.AdminGitOpsUser {
		return namespace, smName, mw.ReturnErrorResponse(http.StatusUnauthorized, errors.Errorf("not authorized"))
	}
	namespace = mux.Vars(r)["namespace"]
	if namespace == "" {
		return namespace, smName,
			mw.ReturnErrorResponse(http.StatusBadRequest, errors.Errorf("namespace cannot be empty"))
	}
	smName = mux.Vars(r)["servicemonitor"]
	if smName == "" {
		return namespace, smName,
			mw.ReturnErrorResponse(http.StatusBadRequest, errors.Errorf("service monitor cannot be empty"))
	}
	return namespace, smName, nil
}
