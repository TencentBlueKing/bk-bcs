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
	// CheckInHunbuStepName step name for check namespace quota in hunbu
	CheckInHunbuStepName = fedsteps.StepNames{
		Alias: "check namespace quota in hunbu",
		Name:  "CHECK_NAMESPACE_QUOTA_IN_HUNBU",
	}
)

// NewCheckInHunbuStep x
func NewCheckInHunbuStep() *CheckInHunbuStep {
	return &CheckInHunbuStep{}
}

// CheckInHunbuStep x
type CheckInHunbuStep struct{}

// Alias step name
func (s CheckInHunbuStep) Alias() string {
	return CheckInHunbuStepName.Alias
}

// GetName step name
func (s CheckInHunbuStep) GetName() string {
	return CheckInHunbuStepName.Name
}

// DoWork for worker exec task
func (s CheckInHunbuStep) DoWork(t *types.Task) error {
	blog.Infof("CheckInHunbuStep is running")

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

	if nsName == "" || subClusterID == "" {
		return fmt.Errorf("get namespace quota task params error, namespace: %s, subClusterID: %s",
			nsName, subClusterID)
	}

	managedClusterLabelsStr, ok := t.GetCommonParams(fedsteps.ManagedClusterLabelsKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.ManagedClusterLabelsKey)
	}

	managedClusterLabelsMap := make(map[string]string)
	if err := json.Unmarshal([]byte(managedClusterLabelsStr), &managedClusterLabelsMap); err != nil {
		blog.Errorf("unmarshal managed cluster labels %s error, %s", managedClusterLabelsStr, err.Error())
		return fmt.Errorf("unmarshal managed cluster labels %s error, %s", managedClusterLabelsStr, err.Error())
	}

	// 判断是否存在ns
	namespace, err := cluster.GetClusterClient().GetNamespace(subClusterID, nsName)
	if err != nil && !errors.IsNotFound(err) {
		blog.Errorf("GetNamespace(%s, %s) failed! err:%s", nsName, subClusterID, err.Error())
		return fmt.Errorf("GetNamespace(%s, %s) failed! err:%s", nsName, subClusterID, err.Error())
	}

	// 当namespace未注册时，才去新增
	if namespace == nil {
		cerr := createHbNamespaceQuota(nsName, subClusterID, managedClusterLabelsMap)
		if cerr != nil {
			blog.Errorf("createHbNamespaceQuota failed, namespace: %s, err: %s", nsName, cerr.Error())
			return fmt.Errorf("createHbNamespaceQuota failed, namespace: %s, err: %s", nsName, cerr.Error())
		}
		blog.Infof("CheckInHunbuStep Success, taskId: %s, taskName: %s, namespace: %s, subClusterID: %s",
			t.GetTaskID(), step.GetName(), nsName, subClusterID)
		return nil
	}

	uerr := updateHbNamespaceQuota(subClusterID, namespace, managedClusterLabelsMap)
	if uerr != nil {
		blog.Errorf("updateHbNamespaceQuota failed, namespace: %s, err: %s", nsName, uerr.Error())
		return fmt.Errorf("updateHbNamespaceQuota failed, namespace: %s, err: %s", nsName, uerr.Error())
	}

	blog.Infof("CheckInHunbuStep Success, taskId: %s, taskName: %s, namespace: %s, subClusterID: %s",
		t.GetTaskID(), step.GetName(), nsName, subClusterID)
	return nil
}

func createHbNamespaceQuota(nsName, subClusterID string, labelsMap map[string]string) error {
	blog.Infof("createHbNamespaceQuota, namespace: %s, labels: %v", nsName, labelsMap)

	annotations := buildHbReq(labelsMap)
	if annotations == nil {
		blog.Errorf("buildHbCreateReq failed, subClusterId: %s, namespace: %s, labelsMap: %v",
			subClusterID, nsName, labelsMap)
		return nil
	}

	cerr := cluster.GetClusterClient().CreateClusterNamespace(nsName, subClusterID, annotations)
	if cerr != nil {
		blog.Errorf("createHbNamespaceQuota failed, namespace: %s, err: %s", nsName, cerr.Error())
		return cerr
	}

	blog.Infof("createHbNamespaceQuota namespace %s Success", nsName)
	return nil
}

func buildHbReq(labelsMap map[string]string) map[string]string {

	hbNsAnnotations := make(map[string]string)
	if labelsMap[cluster.LabelsMixerClusterKey] == cluster.ValueIsTrue {
		hbNsAnnotations[cluster.AnnotationMixerClusterMixerNamespaceKey] = cluster.ValueIsTrue
		// 判断是否有混部集群TKE网络方案
		if labelsMap[cluster.LabelsMixerClusterTkeNetworksKey] != "" {
			hbNsAnnotations[cluster.AnnotationMixerClusterNetworksKey] = labelsMap[cluster.LabelsMixerClusterTkeNetworksKey]
		}
		// 是否有优先级
		if labelsMap[cluster.LabelsMixerClusterPriorityKey] == cluster.ValueIsTrue {
			hbNsAnnotations[cluster.AnnotationMixerClusterPreemptionPolicyKey] = cluster.MixerClusterPreemptionPolicyValue
			hbNsAnnotations[cluster.AnnotationMixerClusterPreemptionClassKey] = cluster.MixerClusterPreemptionClassValue
			hbNsAnnotations[cluster.AnnotationMixerClusterPreemptionValueKey] = cluster.MixerClusterPreemptionValue
		}
	}

	return hbNsAnnotations
}

func updateHbNamespaceQuota(subClusterId string, namespace *corev1.Namespace, labels map[string]string) error {

	blog.Infof("updateHbNamespaceQuota, namespace: %s, labels: %v", namespace.Name, labels)
	annotations := namespace.Annotations
	reqAnnotations := buildHbReq(labels)
	if reqAnnotations == nil {
		blog.Errorf("buildHbUpdateReq failed, namespace: %s, labels: %v", namespace.Name, labels)
		return nil
	}

	for k, v := range reqAnnotations {
		annotations[k] = v
	}

	namespace.Annotations = annotations
	cerr := cluster.GetClusterClient().UpdateNamespace(subClusterId, namespace)
	if cerr != nil {
		blog.Errorf("updateHbNamespaceQuota failed, namespace: %s, err: %s", namespace.Name, cerr.Error())
		return cerr
	}

	blog.Infof("updateHbNamespaceQuota namespace %s Success", namespace.Name)
	return nil
}

// BuildStep build step
func (s CheckInHunbuStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	// stepName/s.GetName() 用于标识这个step
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
