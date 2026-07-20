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
	"k8s.io/apimachinery/pkg/api/errors"

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

	blog.Infof("CheckInNormalStep is running, taskId: %s, hostClusterID: %s, subClusterID: %s, namespace: %s",
		t.GetTaskID(), hostClusterID, subClusterID, nsName)

	if nsName == "" || hostClusterID == "" {
		return fmt.Errorf("CheckInNormalStep task params error, namespace: %s, hostClusterID: %s",
			nsName, hostClusterID)
	}

	nsStr, ok := t.GetCommonParams(fedsteps.SyncNamespaceQuotaKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.SyncNamespaceQuotaKey)
	}

	if nsStr == "" {
		return fmt.Errorf("CheckInNormalStep syncNamespaceQuota param is empty, namespace: %s, hostClusterID: %s",
			nsName, hostClusterID)
	}

	fedNamespace := &corev1.Namespace{}
	nerr := json.Unmarshal([]byte(nsStr), &fedNamespace)
	if nerr != nil {
		blog.Errorf("CheckInNormalStep unmarshal namespace failed, namespace: %s, hostClusterID: %s, err: %s",
			nsName, hostClusterID, nerr.Error())
		return fmt.Errorf("unmarshal namespace failed, namespace: %s, hostClusterID: %s, err: %s",
			nsName, hostClusterID, nerr.Error())
	}

	subClusterNamespace, err := cluster.GetClusterClient().GetNamespace(subClusterID, nsName)
	if err != nil && !errors.IsNotFound(err) {
		blog.Errorf("CheckInNormalStep GetNamespace failed, subClusterID: %s, namespace: %s, err: %s",
			subClusterID, nsName, err.Error())
		return fmt.Errorf("GetNamespace(%s, %s) failed, err: %s", subClusterID, nsName, err.Error())
	}

	if subClusterNamespace == nil {
		cerr := createNormalNamespaceQuota(nsName, subClusterID, fedNamespace)
		if cerr != nil {
			blog.Errorf("CheckInNormalStep createNormalNamespaceQuota failed, subClusterID: %s, namespace: %s, err: %s",
				subClusterID, nsName, cerr.Error())
			return fmt.Errorf("createNormalNamespaceQuota failed, subClusterID: %s, namespace: %s, err: %s",
				subClusterID, nsName, cerr.Error())
		}
		blog.Infof("CheckInNormalStep create success, taskId: %s, stepName: %s, subClusterID: %s, namespace: %s",
			t.GetTaskID(), step.GetName(), subClusterID, nsName)
		return nil
	}

	uerr := updateNormalNamespaceQuota(subClusterID, subClusterNamespace, fedNamespace)
	if uerr != nil {
		blog.Errorf("CheckInNormalStep updateNormalNamespaceQuota failed, subClusterID: %s, namespace: %s, err: %s",
			subClusterID, nsName, uerr.Error())
		return fmt.Errorf("updateNormalNamespaceQuota failed, subClusterID: %s, namespace: %s, err: %s",
			subClusterID, nsName, uerr.Error())
	}

	blog.Infof("CheckInNormalStep update success, taskId: %s, stepName: %s, subClusterID: %s, namespace: %s",
		t.GetTaskID(), step.GetName(), subClusterID, nsName)
	return nil
}

// buildNormalSubClusterAnnotations builds annotations for sub-cluster namespace based on host namespace annotations.
// Priority: obs-cmdb-business-id > bill-projectcode > projectcode
func buildNormalSubClusterAnnotations(hostAnnotations map[string]string) map[string]string {
	annotations := make(map[string]string)

	if val, ok := hostAnnotations[cluster.FedNamespaceObsCmdbBusinessIdKey]; ok && val != "" {
		annotations[cluster.SubClusterBusinessLevel2IdKey] = val
		return annotations
	}

	if val, ok := hostAnnotations[cluster.FedNamespaceBillProjectCodeKey]; ok && val != "" {
		annotations[cluster.FedNamespaceProjectCodeKey] = val
		return annotations
	}

	if val, ok := hostAnnotations[cluster.FedNamespaceProjectCodeKey]; ok && val != "" {
		annotations[cluster.FedNamespaceProjectCodeKey] = val
		return annotations
	}

	return annotations
}

func createNormalNamespaceQuota(nsName, subClusterID string, fedNamespace *corev1.Namespace) error {
	annotations := buildNormalSubClusterAnnotations(fedNamespace.Annotations)
	blog.Infof("createNormalNamespaceQuota, subClusterID: %s, namespace: %s, annotations: %v",
		subClusterID, nsName, annotations)

	cerr := cluster.GetClusterClient().CreateClusterNamespace(subClusterID, nsName, annotations)
	if cerr != nil {
		blog.Errorf("createNormalNamespaceQuota failed, subClusterID: %s, namespace: %s, err: %s",
			subClusterID, nsName, cerr.Error())
		return cerr
	}

	blog.Infof("createNormalNamespaceQuota success, subClusterID: %s, namespace: %s", subClusterID, nsName)
	return nil
}

func updateNormalNamespaceQuota(subClusterId string, subClusterNamespace *corev1.Namespace,
	fedNamespace *corev1.Namespace) error {
	blog.Infof("updateNormalNamespaceQuota, subClusterID: %s, namespace: %s", subClusterId, subClusterNamespace.Name)

	if subClusterNamespace.Annotations == nil {
		subClusterNamespace.Annotations = make(map[string]string)
	}
	newAnnotations := buildNormalSubClusterAnnotations(fedNamespace.Annotations)
	for k, v := range newAnnotations {
		subClusterNamespace.Annotations[k] = v
	}

	cerr := cluster.GetClusterClient().UpdateNamespace(subClusterId, subClusterNamespace)
	if cerr != nil {
		blog.Errorf("updateNormalNamespaceQuota failed, subClusterID: %s, namespace: %s, err: %s",
			subClusterId, subClusterNamespace.Name, cerr.Error())
		return cerr
	}

	blog.Infof("updateNormalNamespaceQuota success, subClusterID: %s, namespace: %s",
		subClusterId, subClusterNamespace.Name)
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
