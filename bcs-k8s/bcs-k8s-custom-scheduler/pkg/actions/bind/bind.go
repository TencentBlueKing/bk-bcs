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

package bind

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/pkg/actions"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/pkg/ipscheduler/v1"
	"github.com/emicklei/go-restful"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
)

const (
	BindPrefix = "bind"
)

func init() {
	actions.RegisterAction(actions.Action{Verb: "POST", Path: actions.BcsCustomSchedulerPrefix + "ipscheduler/" + "{version}/" + BindPrefix,
		Params: nil, Handler: handleIpSchedulerBind})
}

func handleIpSchedulerBind(req *restful.Request, resp *restful.Response) {

	var extenderBindingArgs schedulerapi.ExtenderBindingArgs
	var extenderBindingResult *schedulerapi.ExtenderBindingResult

	err := req.ReadEntity(&extenderBindingArgs)
	if err != nil {
		blog.Errorf("error when read request: %s", err.Error())
		extenderBindingResult = &schedulerapi.ExtenderBindingResult{
			Error: err.Error(),
		}

		resp.WriteEntity(extenderBindingResult)
		return
	}

	ipSchedulerVersion := req.PathParameter("version")
	if ipSchedulerVersion == actions.IpSchedulerV1 {
		err = v1.HandleIpSchedulerBinding(extenderBindingArgs)
	} else {
		err = fmt.Errorf("invalid IpScheduler version")
	}
	if err != nil {
		blog.Errorf("error handling extender binding: %s", err.Error())
		extenderBindingResult = &schedulerapi.ExtenderBindingResult{
			Error: err.Error(),
		}

		resp.WriteEntity(extenderBindingResult)
		return
	}

	extenderBindingResult = &schedulerapi.ExtenderBindingResult{
		Error: "",
	}

	blog.Info("binding finished")
	resp.WriteEntity(extenderBindingResult)
	return
}
