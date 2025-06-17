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

	corev1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
)

var (
	// CheckInNormalStepName step name for check namespace quota in normal
	CheckInNormalStepName = fedsteps.StepNames{
		Alias: "check namespace quota in normal",
		Name:  "CHECK_NAMESPACE_QUOTA_IN_NORMAL",
	}
)

// NewCheckInNormalStep x
func NewCheckInNormalStep() *CheckInNormalStep {
	return &CheckInNormalStep{}
}

// CheckInNormalStep x
type CheckInNormalStep struct{}

// Alias step name
func (s CheckInNormalStep) Alias() string {
	return CheckInNormalStepName.Alias
}

// GetName step name
func (s CheckInNormalStep) GetName() string {
	return CheckInNormalStepName.Name
}

// DoWork for worker exec task
func (s CheckInNormalStep) DoWork(t *types.Task) error {
	blog.Infof("CheckInNormalStep is running")

	step, exist := t.GetStep(s.GetName())
	if !exist {
		return fmt.Errorf("task[%s] not exist step[%s]", t.TaskID, s.GetName())
	}

	nsName, ok := t.GetCommonParams(fedsteps.NamespaceKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.NamespaceKey)
	}

	subClusterID, ok := t.GetCommonParams(fedsteps.SubClusterIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.SubClusterIdKey)
	}

	hostClusterID, ok := t.GetCommonParams(fedsteps.HostClusterIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.HostClusterIdKey)
	}

	if nsName == "" || hostClusterID == "" {
		return fmt.Errorf("getFedNamespaceQuota task params error, fedNamespace: %s, hostClusterID: %s",
			nsName, hostClusterID)
	}

	nsStr, ok := t.GetCommonParams(fedsteps.SyncNamespaceQuotaKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.SyncNamespaceQuotaKey)
	}

	if nsStr == "" {
		blog.Errorf("getFedNamespaceQuota task params error, fedNamespace: %s, hostClusterId: %s",
			nsName, hostClusterID)
		return fmt.Errorf("getFedNamespaceQuota task params error, fedNamespace: %s, hostClusterId: %s",
			nsName, hostClusterID)
	}

	fedNamespace := &corev1.Namespace{}
	nerr := json.Unmarshal([]byte(nsStr), &fedNamespace)
	if nerr != nil {
		blog.Errorf("unmarshal namespace failed, fedNamespace: %s, hostClusterId: %s, err: %s",
			nsName, hostClusterID, nerr.Error())
		return fmt.Errorf("unmarshal namespace failed, fedNamespace: %s, hostClusterId: %s", nsName, hostClusterID)
	}

	// 判断是否存在ns
	subClusterNamespace, err := cluster.GetClusterClient().GetNamespace(subClusterID, nsName)
	if err != nil {
		blog.Errorf("GetNamespace(%s, %s) failed! err:%s", nsName, subClusterID, err.Error())
		return fmt.Errorf("GetNamespace(%s, %s) failed! err: %s", nsName, subClusterID, err.Error())
	}

	// 当subClusterNamespace未注册时，才去新增
	if subClusterNamespace == nil {
		cerr := createNormalNamespaceQuota(nsName, subClusterID, fedNamespace)
		if cerr != nil {
			blog.Errorf("createNormalNamespaceQuota failed, fedNamespace: %s, err: %s", nsName, cerr.Error())
			return fmt.Errorf("createNormalNamespaceQuota failed, fedNamespace: %s, err: %s", nsName, cerr.Error())
		}
		blog.Infof("CheckInNormalStep Success, taskId: %s, taskName: %s, subClusterNamespace: %s, subClusterID: %s",
			t.GetTaskID(), step.GetName(), nsName, subClusterID)
		return nil
	}

	uerr := updateNormalNamespaceQuota(subClusterID, subClusterNamespace, fedNamespace)
	if uerr != nil {
		blog.Errorf("updateNormalNamespaceQuota failed, subClusterNamespace: %s, err: %s", nsName, uerr.Error())
		return fmt.Errorf("updateNormalNamespaceQuota failed, subClusterNamespace: %s, err: %s", nsName, uerr.Error())
	}

	blog.Infof("CheckInNormalStep Success, taskId: %s, taskName: %s, subClusterNamespace: %s, subClusterID: %s",
		t.GetTaskID(), step.GetName(), nsName, subClusterID)
	return nil
}

func createNormalNamespaceQuota(nsName, subClusterID string, fedNamespace *corev1.Namespace) error {
	blog.Infof("createNormalNamespaceQuota, subClusterID: %s, namespace: %s", fedNamespace.Name)

	cerr := cluster.GetClusterClient().CreateClusterNamespace(nsName, subClusterID, fedNamespace.Annotations)
	if cerr != nil {
		blog.Errorf("createNormalNamespaceQuota failed, namespace: %s, err: %s", nsName, cerr.Error())
		return cerr
	}

	blog.Infof("createNormalNamespaceQuota namespace %s Success", nsName)
	return nil
}

func updateNormalNamespaceQuota(subClusterId string, subClusterNamespace *corev1.Namespace, fedNamespace *corev1.Namespace) error {
	blog.Infof("updateNormalNamespaceQuota, namespace: %s", subClusterNamespace.Name)

	subClusterNamespace.Annotations = fedNamespace.Annotations
	cerr := cluster.GetClusterClient().UpdateNamespace(subClusterId, subClusterNamespace)
	if cerr != nil {
		blog.Errorf("updateNormalNamespaceQuota failed, namespace: %s, err: %s", subClusterNamespace.Name, cerr.Error())
		return cerr
	}

	blog.Infof("updateNormalNamespaceQuota namespace %s Success", subClusterNamespace.Name)
	return nil
}

// BuildStep build step
func (s CheckInNormalStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	// stepName/s.GetName() 用于标识这个step
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
