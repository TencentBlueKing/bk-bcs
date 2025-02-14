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

package core

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	scheduler_util "k8s.io/autoscaler/cluster-autoscaler/utils/scheduler"
	kubeclient "k8s.io/client-go/kubernetes"
	v1lister "k8s.io/client-go/listers/core/v1"
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	contextinternal "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/context"
	metricsinternal "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/metrics"
)

var _ Webhook = &WebScaler{}

// WebScaler implements Webhook via web
type WebScaler struct {
	url                 string
	token               string
	client              *http.Client
	configmapLister     v1lister.ConfigMapNamespaceLister
	maxBulkScaleUpCount int
	batchScaleUpCount   int
}

// NewWebScaler initializes a WebScaler
func NewWebScaler(kubeClient kubeclient.Interface, configNamespace, url, token string,
	maxBulkScaleUpCount, batchScaleUpCount int) Webhook {
	// nolint
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
	}
	stopChannel := make(chan struct{})
	lister := kubernetes.NewConfigMapListerForNamespace(kubeClient, stopChannel, configNamespace)
	return &WebScaler{url: url, token: token, client: client, configmapLister: lister.ConfigMaps(configNamespace),
		maxBulkScaleUpCount: maxBulkScaleUpCount, batchScaleUpCount: batchScaleUpCount}
}

// DoWebhook get responses from webhook, then execute scale based on responses
func (w *WebScaler) DoWebhook(context *contextinternal.Context,
	clusterStateRegistry *clusterstate.ClusterStateRegistry, sd *ScaleDown,
	nodes []*corev1.Node, pods []*corev1.Pod) errors.AutoscalerError {
	nodeNameToNodeInfo := scheduler_util.CreateNodeNameToInfoMap(pods, nodes)

	options, candidates, err := w.GetResponses(context, clusterStateRegistry,
		nodeNameToNodeInfo, nodes, sd)
	if err != nil {
		return errors.NewAutoscalerError(errors.InternalError,
			"failed to get response from web server: %v", err)
	}

	// // check limits for CPU and memory
	// checkErr := checkResourcesLimits(context, nodes, options, candidates)
	// if checkErr != nil {
	// 	return checkErr
	// }

	err = w.ExecuteScale(context, clusterStateRegistry, sd, nodes, options, candidates, nodeNameToNodeInfo)
	if err != nil {
		return errors.NewAutoscalerError(errors.CloudProviderError,
			"failed to execute scale from web server: %v", err)
	}
	return nil
}

// GetResponses returns the responses of webhook
func (w *WebScaler) GetResponses(context *contextinternal.Context,
	clusterStateRegistry *clusterstate.ClusterStateRegistry,
	nodeNameToNodeInfo map[string]*schedulerframework.NodeInfo,
	nodes []*corev1.Node, sd *ScaleDown) (ScaleUpOptions, ScaleDownCandidates, error) {
	// get node group's priority
	newPriorities, err := getPriority(w.configmapLister)
	if err != nil {
		context.LogRecorder.Eventf(corev1.EventTypeWarning, "PriorityConfigMapInvalid", err.Error())
		klog.Warning(err.Error())
		return nil, nil, err
	}

	// construct requests
	req, err := GenerateAutoscalerRequest(context.CloudProvider.NodeGroups(),
		clusterStateRegistry.GetUpcomingNodes(), newPriorities, sd.nodeDeletionTracker)
	if err != nil {
		return nil, nil, fmt.Errorf("Cannot generate autoscaler requests, err: %s", err.Error())
	}
	review := ClusterAutoscalerReview{
		Request:  req,
		Response: nil,
	}

	// post requests to the url
	b, err := json.Marshal(review)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"Cannot marshal review to bytes, err: %s", err.Error())
	}
	start := time.Now()
	result, err := postRequest(w.url, w.token, w.client, b)
	metricsinternal.UpdateWebhookExecDuration(start)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"Failed to post review to url: %s err: %s", w.url, err.Error())
	}
	var faResp ClusterAutoscalerReview
	err = json.Unmarshal(result, &faResp)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to unmarshal response to review, err: %s, response: %s",
			err.Error(), string(result))
	}

	// get response
	faResp.Request = req
	klog.Infof("Get webhook response from web: %+v", faResp.Response)
	options, candidates, err := HandleResponse(faResp, nodes, nodeNameToNodeInfo,
		sd, newPriorities, context.MaxNodeProvisionTime)
	if err != nil {
		return nil, nil, err
	}

	return options, candidates, nil
}

// ExecuteScale execute scale up and down based on webhook responses
func (w *WebScaler) ExecuteScale(context *contextinternal.Context,
	clusterStateRegistry *clusterstate.ClusterStateRegistry,
	sd *ScaleDown, nodes []*corev1.Node, options ScaleUpOptions,
	candidates ScaleDownCandidates,
	nodeNameToNodeInfo map[string]*schedulerframework.NodeInfo) error {
	err := ExecuteScaleUp(context, clusterStateRegistry, options, w.maxBulkScaleUpCount, w.batchScaleUpCount)
	if err != nil {
		return err
	}

	// if len(options) > 0 {
	// 	klog.Infof("Scaling up node groups now, skip scaling down progress")
	// 	return nil
	// }

	err = ExecuteScaleDown(context, sd, nodes, candidates, nodeNameToNodeInfo)
	if err != nil {
		return err
	}

	return nil
}

func postRequest(url, token string, client *http.Client, data []byte) ([]byte, error) {
	req, _ := http.NewRequest("POST", url, strings.NewReader(string(data)))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to finish this request: %v", err)
	}

	contentsBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %+v err: %v", resp, err)
	}
	if resp.StatusCode/100 > 2 {
		return nil, fmt.Errorf("failed to finish this request: %v, body: %v", resp.StatusCode, string(contentsBytes))
	}
	return contentsBytes, nil
}
