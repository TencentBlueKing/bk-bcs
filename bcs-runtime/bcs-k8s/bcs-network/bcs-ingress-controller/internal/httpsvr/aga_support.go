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

package httpsvr

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
)

// AgaHostRegion aga api can only be call in us-west-region
const AgaHostRegion = "us-west-2"

func (h *HttpServerClient) getPodRelatedAgaEntrance(request *restful.Request, response *restful.Response) {
	startTime := time.Now()
	mf := func(status string) {
		metrics.ReportAPIRequestMetric("aga_entrance", "GET", status, startTime)
	}
	if h.AgaSupporter == nil {
		_, _ = response.Write(CreateResponseData(errors.New("Not init, check if select aws as cloud provider"), "",
			nil))
		mf(strconv.Itoa(http.StatusInternalServerError))
		return
	}
	podName := request.QueryParameter("pod_name")
	podNamespace := request.QueryParameter("pod_namespace")
	if podName == "" || podNamespace == "" {
		_, _ = response.Write(CreateResponseData(errors.New("empty parameter: pod_name or pod_namespace"), "", nil))
		mf(strconv.Itoa(http.StatusInternalServerError))
		return
	}

	region := request.QueryParameter("region")
	if region == "" {
		blog.V(4).Infof("empty region parameter, use default region ")
		region = h.Ops.Region
	}
	agaHostRegion := request.QueryParameter("aga_host_region")
	if agaHostRegion == "" {
		blog.V(4).Infof("empty aga host region parameter, use default region, i.e. us-west-2")
		agaHostRegion = AgaHostRegion
	}

	blog.V(3).Infof("getPodRelatedAgaEntrance req[pod_name='%s', pod_namespace='%s', region='%s']", podName,
		podNamespace, region)

	pod := &v1.Pod{}
	// 通过API Server获取信息，避免获取Pod信息延迟（获取不到pod IP）
	if err := h.Mgr.GetAPIReader().Get(context.TODO(), types.NamespacedName{
		Namespace: podNamespace,
		Name:      podName,
	}, pod); err != nil {
		err = errors.Wrapf(err, "get pods '%s/%s' failed", podNamespace, podName)
		blog.Errorf(err.Error())
		_, _ = response.Write(CreateResponseData(err, "", nil))
		mf(strconv.Itoa(http.StatusInternalServerError))
		return
	}

	if pod.Status.PodIP == "" {
		err := errors.Errorf("empty pod ip [%s/%s]", podNamespace, podName)
		blog.Warnf(err.Error())
		_, _ = response.Write(CreateResponseData(err, "", nil))
		return
	}
	// get node from pod's spec
	node := &v1.Node{}
	if err := h.Mgr.GetClient().Get(context.TODO(), types.NamespacedName{
		Name: pod.Spec.NodeName,
	}, node); err != nil {
		err = errors.Wrapf(err, "get node '%s' failed", pod.Spec.NodeName)
		blog.Errorf(err.Error())
		_, _ = response.Write(CreateResponseData(err, "", nil))
		mf(strconv.Itoa(http.StatusInternalServerError))
		return
	}

	resp, err := h.AgaSupporter.ListCustomRoutingByDefinition(agaHostRegion, region, getNodeInstanceID(node),
		pod.Status.PodIP)
	if err != nil {
		_, _ = response.Write(CreateResponseData(err, "", nil))
		mf(strconv.Itoa(http.StatusInternalServerError))
		return
	}

	_, _ = response.Write(CreateResponseData(nil, "", resp))
	mf(strconv.Itoa(http.StatusOK))
}

// aws中， node.Spec.ProviderID的格式为aws:///<availability_zone>/<instance_id>
// 如， aws:///us-west-2a/i-0abcdef1234567890
func getNodeInstanceID(node *v1.Node) string {
	providerID := node.Spec.ProviderID
	parts := strings.Split(providerID, "/")
	instanceID := parts[len(parts)-1]
	return instanceID
}
