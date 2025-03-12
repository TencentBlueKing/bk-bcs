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

// Package steps include all steps for federation manager
package steps

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/avast/retry-go"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	third "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/thirdparty"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
	trd "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/pkg/bcsapi/thirdparty-service"
)

var (
	// HandleTaijiNamespaceStepName step name for create cluster
	HandleTaijiNamespaceStepName = fedsteps.StepNames{
		Alias: "handle taiji namespace",
		Name:  "HANDLE_TAIJI_NAMESPACE",
	}
)

// NewHandleTaijiNamespaceStep x
func NewHandleTaijiNamespaceStep() *HandleTaijiNamespaceStep {
	return &HandleTaijiNamespaceStep{}
}

// HandleTaijiNamespaceStep x
type HandleTaijiNamespaceStep struct{}

// Alias step name
func (s HandleTaijiNamespaceStep) Alias() string {
	return HandleTaijiNamespaceStepName.Alias
}

// GetName step name
func (s HandleTaijiNamespaceStep) GetName() string {
	return HandleTaijiNamespaceStepName.Name
}

// DoWork for worker exec task
func (s HandleTaijiNamespaceStep) DoWork(t *types.Task) error {
	step, exist := t.GetStep(s.GetName())
	if !exist {
		return fmt.Errorf("task[%s] not exist step[%s]", t.TaskID, s.GetName())
	}

	hostClusterId, ok := step.GetParam(fedsteps.HostClusterIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.HostClusterIdKey)
	}

	opt, ok := step.GetParam(fedsteps.ParameterKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, "taiji parameter")
	}

	handleType, ok := step.GetParam(fedsteps.HandleTypeKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, "taiji handleType")
	}

	blog.Infof("taiji task handleType: %s  opt: %s", handleType, opt)

	switch handleType {
	case fedsteps.CreateKey:
		var reqList []*trd.CreateNamespaceForTaijiV3Request
		blog.Infof("create taiji namespace opt: %s ", opt)
		err := json.Unmarshal([]byte(opt), &reqList)
		if err != nil {
			blog.Errorf("task[%s] create taiji namespace.Unmarshal failed "+
				"body: %s, err: %s", t.TaskID, opt, err.Error())
			return err
		}

		taijiV3Request := reqList[0]
		// rpc 调用 taiji 创建 ns 接口
		bkBizId := taijiV3Request.BkBizId
		bkModuleId := taijiV3Request.BkModuleId
		// 判断 bkBizId是否为空
		if bkBizId == "" && bkModuleId == "" {
			// 请求taiji createModule api
			result, err := third.GetThirdpartyClient().CreateModule(taijiV3Request.Namespace)
			if err != nil {
				blog.Errorf("task[%s] create taiji namespace.CreateModule failed "+
					"bkModuleName: %s, err: %s", t.TaskID, taijiV3Request.Namespace, err.Error())
				return err
			}

			bkBizId = strconv.FormatInt(result.Data.BkBizId, 10)
			bkModuleId = strconv.FormatInt(result.Data.BkModuleId, 10)
			blog.Infof("bkBizId: %s, bkModuleId: %s,result: %+v", bkBizId, bkModuleId, result)

			// 查询ns
			fedNamespace, err := cluster.GetClusterClient().GetNamespace(hostClusterId, taijiV3Request.Namespace)
			if err != nil {
				return err
			}

			// 将bkBizId和bkModuleId存到联邦集群ns annotations里面
			fedNamespace.Annotations[cluster.FedNamespaceBkBizId] = bkBizId
			fedNamespace.Annotations[cluster.FedNamespaceBkModuleId] = bkModuleId
			fedNamespace.Annotations[cluster.NamespaceUpdateTimestamp] = time.Now().Format(time.RFC3339)
			err = cluster.GetClusterClient().UpdateNamespace(hostClusterId, fedNamespace)
			if err != nil {
				return err
			}
		} else if bkBizId == "" || bkModuleId == "" {
			return fmt.Errorf("task create taiji namespace failed, bkBizId,bkModuleId is empty")
		}

		if err := retry.Do(func() error {
			blog.Info("CreateNamespaceForTaijiV3 is running")
			err = third.GetThirdpartyClient().CreateNamespaceForTaijiV3(&trd.CreateNamespaceForTaijiV3Request{
				Namespace:         taijiV3Request.Namespace,
				BkBizId:           bkBizId,
				BkModuleId:        bkModuleId,
				SubQuotaInfos:     taijiV3Request.SubQuotaInfos,
				Location:          taijiV3Request.Location,
				IsPrivateResource: taijiV3Request.IsPrivateResource,
				ScheduleAlgorithm: taijiV3Request.ScheduleAlgorithm,
			})

			if err != nil {
				blog.Errorf("task[%s] create taiji namespace.CreateNamespaceForTaijiV3 failed, location: %s, "+
					"subQuotaInfos: %v, bkModuleId: %s, bkBizId: %s, namespace: %s, err: %s",
					t.TaskID, taijiV3Request.Location, taijiV3Request.SubQuotaInfos, bkModuleId, bkBizId, taijiV3Request.Namespace, err.Error())
				return err
			}

			blog.Info("CreateNamespaceForTaijiV3 is over")
			return nil
		}, retry.Attempts(fedsteps.DefaultAttemptTimes), retry.Delay(fedsteps.DefaultRetryDelay*time.Minute),
			retry.DelayType(retry.BackOffDelay), retry.MaxDelay(fedsteps.DefaultMaxDelay*time.Minute)); err != nil {
			return err
		}
	case fedsteps.UpdateKey:
		var reqList []*trd.UpdateQuotaInfoForTaijiRequest
		err := json.Unmarshal([]byte(opt), &reqList)
		if err != nil {
			blog.Errorf("task[%s] update taiji quota.Unmarshal failed "+
				"body: %s, err: %s", t.TaskID, opt, err.Error())
			return err
		}

		blog.Infof("task[%s] update taiji namespace running opt: %s", t.TaskID, opt)
		// waiting for request bcs-thirdparty-service taiji update ns quota api
		if err := retry.Do(func() error {
			for _, taijiRequest := range reqList {
				err = third.GetThirdpartyClient().UpdateQuotaInfoForTaiji(taijiRequest)
				if err != nil {
					blog.Errorf("task[%s] update taiji namespace.UpdateQuotaInfoForTaiji failed "+
						"taijiRequest: %+v, err: %s", t.TaskID, taijiRequest, err.Error())
					return err
				}
			}

			return nil
		}, retry.Attempts(fedsteps.DefaultAttemptTimes), retry.Delay(fedsteps.DefaultRetryDelay*time.Minute),
			retry.DelayType(retry.BackOffDelay), retry.MaxDelay(fedsteps.DefaultMaxDelay*time.Minute)); err != nil {
			return err
		}
	}

	blog.Infof("taskId: %s, taskType: %s, taskName: %s result: %v\n", t.GetTaskID(), t.GetTaskType(),
		step.GetName(), fedsteps.Success)
	return nil
}

// BuildStep build step
func (s HandleTaijiNamespaceStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	// stepName/s.GetName() 用于标识这个step
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
