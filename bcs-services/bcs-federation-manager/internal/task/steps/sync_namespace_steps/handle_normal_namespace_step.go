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
	"strings"
	"time"

	"github.com/avast/retry-go"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
	tktyps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/types"
)

var (
	// HandleNormalNamespaceStepName step name for create cluster
	HandleNormalNamespaceStepName = fedsteps.StepNames{
		Alias: "handle normal namespace",
		Name:  "HANDLE_NORMAL_NAMESPACE",
	}
)

// NewHandleNormalNamespaceStep x
func NewHandleNormalNamespaceStep() *HandleNormalNamespaceStep {
	return &HandleNormalNamespaceStep{}
}

// HandleNormalNamespaceStep x
type HandleNormalNamespaceStep struct{}

// Alias step name
func (s HandleNormalNamespaceStep) Alias() string {
	return HandleNormalNamespaceStepName.Alias
}

// GetName step name
func (s HandleNormalNamespaceStep) GetName() string {
	return HandleNormalNamespaceStepName.Name
}

// DoWork for worker exec task
func (s HandleNormalNamespaceStep) DoWork(t *types.Task) error {
	step, exist := t.GetStep(s.GetName())
	if !exist {
		return fmt.Errorf("task[%s] not exist step[%s]", t.TaskID, s.GetName())
	}

	namespace, ok := step.GetParam(fedsteps.NamespaceKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.NamespaceKey)
	}

	parameter, ok := step.GetParam(fedsteps.ParameterKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.ParameterKey)
	}

	handleType, ok := step.GetParam(fedsteps.HandleTypeKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.HandleTypeKey)
	}

	blog.Infof("normal task namespace: %s; handleType: %s; opt: %s", namespace, handleType, parameter)

	if err := retry.Do(func() error {
		switch handleType {
		case fedsteps.CreateKey:
			forNormalList := make([]*tktyps.HandleNormalNamespace, 0)
			err := json.Unmarshal([]byte(parameter), &forNormalList)
			if err != nil {
				blog.Errorf(
					"task[%s] handle create HandleNormal namespace.Unmarshal failed "+
						"body: %s, err: %s", t.TaskID, parameter, err.Error())
				return err
			}

			blog.Infof("task[%s] create normal namespace running", t.TaskID)
			for _, nmReq := range forNormalList {
				err := cluster.GetClusterClient().CreateClusterNamespace(
					nmReq.SubClusterId, namespace, nmReq.Annotations)
				if err != nil {
					blog.Errorf(
						"task.CreateClusterNamespace 创建 ns failed, "+
							"subClusterId: %s, namespace: %s, annotations: %+v, err: %s",
						nmReq.SubClusterId, namespace, nmReq.Annotations, err.Error())
					return err
				}
			}

		case fedsteps.UpdateKey:

			forNormalList := make([]*tktyps.HandleNormalNamespace, 0)
			err := json.Unmarshal([]byte(parameter), &forNormalList)
			if err != nil {
				blog.Errorf(
					"task[%s] update normal namespace.Unmarshal failed "+
						"body: %s, err: %s", t.TaskID, parameter, err.Error())
				return err
			}

			blog.Infof("task[%s] update normal namespace running", t.TaskID)
			for _, nmReq := range forNormalList {
				normalNamespace, err := cluster.GetClusterClient().GetNamespace(
					nmReq.SubClusterId, namespace)
				if err != nil {
					blog.Errorf(
						"task.GetNamespace failed, subClusterId: %s, namespace: %s, err: %s",
						nmReq.SubClusterId, namespace, err.Error())
					return err
				}

				for k, v := range nmReq.Annotations {
					normalNamespace.Annotations[k] = v
				}

				err = cluster.GetClusterClient().UpdateNamespace(nmReq.SubClusterId, normalNamespace)
				if err != nil {
					blog.Errorf(
						"task.UpdateNamespace failed, subClusterId: %s, normalNamespace: %+v, err: %s",
						nmReq.SubClusterId, normalNamespace, err.Error())
					return err
				}
			}
		case fedsteps.DeleteKey:
			blog.Infof("task[%s] delete normal namespace running parameter: %s", t.TaskID, parameter)
			clusterIds := strings.Split(parameter, ",")
			for _, clusterId := range clusterIds {
				err := cluster.GetClusterClient().DeleteNamespace(clusterId, namespace)
				if err != nil {
					blog.Errorf(
						"task[%s] handle delete normal namespace failed "+
							"body: %s, err: %s", t.TaskID, parameter, err.Error())
					return err
				}
			}
		}

		return nil
	}, retry.Attempts(fedsteps.DefaultAttemptTimes), retry.Delay(fedsteps.DefaultRetryDelay*time.Minute),
		retry.DelayType(retry.BackOffDelay), retry.MaxDelay(fedsteps.DefaultMaxDelay*time.Minute)); err != nil {
		return err
	}

	blog.Infof("normal namespace taskId: %s, taskType: %s, taskName: %s result: %v\n", t.GetTaskID(), t.GetTaskType(),
		step.GetName(), fedsteps.Success)
	return nil
}

// BuildStep build step
func (s HandleNormalNamespaceStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	// stepName/s.GetName() 用于标识这个step
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
