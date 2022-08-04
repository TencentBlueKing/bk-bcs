/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	errors_util "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	scheduler_util "k8s.io/autoscaler/cluster-autoscaler/utils/scheduler"
	"k8s.io/klog"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	contextinternal "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/context"
)

var client = http.Client{
	Timeout: 5 * time.Second,
}

var _ Webhook = &WebScaler{}

// WebScaler impletements Webhook via web
type WebScaler struct {
	url string
}

// NewWebScaler initilizes a WebScaler
func NewWebScaler(url string) Webhook {
	return &WebScaler{url: url}
}

// DoWebhook get responses from webhook, then execute scale based on responses
func (w *WebScaler) DoWebhook(context *contextinternal.Context,
	clusterStateRegistry *clusterstate.ClusterStateRegistry, sd *ScaleDown,
	nodes []*corev1.Node, pods []*corev1.Pod) errors.AutoscalerError {
	nodeNameToNodeInfo := scheduler_util.CreateNodeNameToInfoMap(pods, nodes)

	options, candidates, err := w.GetResponses(context, clusterStateRegistry,
		nodeNameToNodeInfo, nodes, sd)
	if err != nil {
		return errors.NewAutoscalerError(errors.ApiCallError,
			"failed to get response from web server: %v", err)
	}
	err = w.ExecuteScale(context, clusterStateRegistry, sd, nodes, options, candidates, nodeNameToNodeInfo)
	if err != nil {
		return errors.NewAutoscalerError(errors.ApiCallError,
			"failed to execute scale from web server: %v", err)
	}
	return nil
}

// GetResponses returns the responses of webhook
func (w *WebScaler) GetResponses(context *contextinternal.Context,
	clusterStateRegistry *clusterstate.ClusterStateRegistry,
	nodeNameToNodeInfo map[string]*schedulernodeinfo.NodeInfo,
	nodes []*corev1.Node, sd *ScaleDown) (ScaleUpOptions, ScaleDownCandidates, error) {

	// construct requests
	req, err := GenerateAutoscalerRequest(context.CloudProvider.NodeGroups(), clusterStateRegistry.GetUpcomingNodes())
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
	res, err := client.Post(w.url, "application/json", strings.NewReader(string(b)))
	if err != nil {
		return nil, nil, fmt.Errorf(
			"Failed to post review to url: %s err: %s", w.url, err.Error())
	}
	defer func() {
		if cerr := res.Body.Close(); cerr != nil {
			if err != nil {
				err = errors_util.Wrap(err, cerr.Error())
			} else {
				err = cerr
			}
		}
	}()
	result, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to read body of response, err: %s", err.Error())
	}
	var faResp ClusterAutoscalerReview
	err = json.Unmarshal(result, &faResp)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to unmarshal response to review, err: %s", err.Error())
	}

	// get response
	faResp.Request = req
	klog.Infof("Get webhook response from web: %+v", faResp.Response)
	options, candidates, err := HandleResponse(faResp, nodes, nodeNameToNodeInfo, sd)
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
	nodeNameToNodeInfo map[string]*schedulernodeinfo.NodeInfo) error {
	err := ExecuteScaleUp(context, clusterStateRegistry, options)
	if err != nil {
		return err
	}

	if len(options) > 0 {
		klog.Infof("Scaling up node groups now, skip scaling down progress")
		return nil
	}

	err = ExecuteScaleDown(context, sd, nodes, candidates, nodeNameToNodeInfo)
	if err != nil {
		return err
	}

	return nil
}
