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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
)

var (
	// GetNamespaceQuotaStepName step name for get namespace and quota
	GetNamespaceQuotaStepName = fedsteps.StepNames{
		Alias: "get namespace and quota",
		Name:  "GET_NAMESPACE_AND_QUOTA",
	}
)

// NewGetNamespaceQuotaStep x
func NewGetNamespaceQuotaStep() *GetNamespaceQuotaStep {
	return &GetNamespaceQuotaStep{}
}

// GetNamespaceQuotaStep x
type GetNamespaceQuotaStep struct{}

// Alias step name
func (s GetNamespaceQuotaStep) Alias() string {
	return GetNamespaceQuotaStepName.Alias
}

// GetName step name
func (s GetNamespaceQuotaStep) GetName() string {
	return GetNamespaceQuotaStepName.Name
}

// DoWork for worker exec task
func (s GetNamespaceQuotaStep) DoWork(t *types.Task) error {
	blog.Infof("getNamespaceQuotaStep is running")

	step, exist := t.GetStep(s.GetName())
	if !exist {
		return fmt.Errorf("task[%s] not exist step[%s]", t.TaskID, s.GetName())
	}

	nsName, ok := t.GetCommonParams(fedsteps.NamespaceKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.NamespaceKey)
	}

	hostClusterID, ok := t.GetCommonParams(fedsteps.HostClusterIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.HostClusterIdKey)
	}

	if nsName == "" || hostClusterID == "" {
		return fmt.Errorf("get namespace quota task params error, namespace: %s, hostClusterId: %s",
			nsName, hostClusterID)
	}
	// get cluster namespace
	namespace, err := cluster.GetClusterClient().GetNamespace(hostClusterID, nsName)
	if err != nil {
		blog.Errorf("getNamespaceQuotaStep taskId: %s, namespace: %s, hostClusterId: %s, err: %s",
			t.GetTaskID(), nsName, hostClusterID, err.Error())
		return fmt.Errorf("getNamespaceQuotaStep taskId: %s, namespace: %s, hostClusterId: %s, err: %s",
			t.GetTaskID(), nsName, hostClusterID, err.Error())
	}

	if namespace == nil {
		blog.Errorf("getNamespaceQuotaStep taskId: %s, namespace: %s, hostClusterId: %s, err: %s",
			t.GetTaskID(), nsName, hostClusterID, "namespace is nil")
		return fmt.Errorf("getNamespaceQuotaStep taskId: %s, namespace: %s, hostClusterId: %s, err: %s",
			t.GetTaskID(), nsName, hostClusterID, "namespace is nil")
	}

	nsBytes, nerr := json.Marshal(namespace)
	if nerr != nil {
		blog.Errorf("getNamespaceQuotaStep taskId: %s, namespace: %s, hostClusterId: %s, err: %s",
			t.GetTaskID(), nsName, hostClusterID, nerr.Error())
		return fmt.Errorf("getNamespaceQuotaStep taskId: %s, namespace: %s, hostClusterId: %s, err: %s",
			t.GetTaskID(), nsName, hostClusterID, nerr.Error())
	}

	t.AddCommonParams(fedsteps.SyncNamespaceQuotaKey, string(nsBytes))

	// list namespace quotas
	multiClusterResourceQuotaList, err := cluster.GetClusterClient().ListNamespaceQuota(hostClusterID, nsName)
	if err != nil {
		blog.Errorf("getNamespaceQuotaStep taskId: %s, namespace: %s, hostClusterId: %s, err: %s",
			t.GetTaskID(), nsName, hostClusterID, err.Error())
		return fmt.Errorf("getNamespaceQuotaStep taskId: %s, namespace: %s, hostClusterId: %s, err: %s",
			t.GetTaskID(), nsName, hostClusterID, err.Error())
	}

	if len(multiClusterResourceQuotaList.Items) > 0 {
		quotaListBytes, qerr := json.Marshal(multiClusterResourceQuotaList.Items)
		if qerr != nil {
			blog.Errorf("getNamespaceQuotaStep taskId: %s, namespace: %s, hostClusterId: %s, err: %s",
				t.GetTaskID(), nsName, hostClusterID, qerr.Error())
			return fmt.Errorf("getNamespaceQuotaStep taskId: %s, namespace: %s, hostClusterId: %s, err: %s",
				t.GetTaskID(), nsName, hostClusterID, qerr.Error())
		}
		t.AddCommonParams(fedsteps.NamespaceQuotaListKey, string(quotaListBytes))
	}

	blog.Infof("getNamespaceQuotaStep Success, taskId: %s, taskName: %s, namespace: %s, hostClusterId: %s",
		t.GetTaskID(), step.GetName(), nsName, hostClusterID)
	return nil
}

// BuildStep build step
func (s GetNamespaceQuotaStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	// stepName/s.GetName() 用于标识这个step
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
