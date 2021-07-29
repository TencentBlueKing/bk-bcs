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

package predicate

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/pkg/actions"
	v1 "github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/pkg/ipscheduler/v1"
	v2 "github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/pkg/ipscheduler/v2"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/pkg/metrics"
	"time"

	"github.com/emicklei/go-restful"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
)

const (
	PredicatePrefix = "predicate"
)

func init() {
	actions.RegisterAction(actions.Action{Verb: "POST", Path: actions.BcsCustomSchedulerPrefix + "ipscheduler/" + "{version}/" + PredicatePrefix,
		Params: nil, Handler: handleIpSchedulerPredicate})
}

func handleIpSchedulerPredicate(req *restful.Request, resp *restful.Response) {
	ipSchedulerVersion := req.PathParameter("version")

	var extenderArgs schedulerapi.ExtenderArgs
	var extenderFilterResult *schedulerapi.ExtenderFilterResult
	var metricsArgs = &metrics.RecordConfig{
		Version: ipSchedulerVersion,
		Handler: "handleIpSchedulerPredicate",
		Method:  "POST",
		Status:  "",
		Started: time.Now(),
	}

	defer func() {
		metrics.ReportK8sCustomSchedulerAPIMetrics(metricsArgs.Version, metricsArgs.Handler, metricsArgs.Method,
			metricsArgs.Status, metricsArgs.Started)
	}()

	err := req.ReadEntity(&extenderArgs)
	if err != nil {
		blog.Infof("error when read request: %s", err.Error())
		extenderFilterResult = &schedulerapi.ExtenderFilterResult{
			Nodes:       nil,
			FailedNodes: nil,
			Error:       err.Error(),
		}

		metricsArgs.Status = metrics.ErrStatus
		resp.WriteEntity(extenderFilterResult)
		return
	}

	metricsArgs.Status = metrics.SucStatus
	if ipSchedulerVersion == actions.IpSchedulerV1 {
		extenderFilterResult, err = v1.HandleIpSchedulerPredicate(extenderArgs)
	} else if ipSchedulerVersion == actions.IpSchedulerV2 {
		extenderFilterResult, err = v2.HandleIpSchedulerPredicate(extenderArgs)
	} else {
		extenderFilterResult = &schedulerapi.ExtenderFilterResult{
			Nodes:       nil,
			FailedNodes: nil,
			Error:       "invalid IpScheduler version",
		}
	}
	if err != nil {
		extenderFilterResult = &schedulerapi.ExtenderFilterResult{
			Nodes:       nil,
			FailedNodes: nil,
			Error:       err.Error(),
		}
		metricsArgs.Status = metrics.ErrStatus
	}

	resp.WriteEntity(extenderFilterResult)
	return
}
