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

	hostClusterID, _ := t.GetCommonParams(fedsteps.HostClusterIdKey)

	blog.Infof("CheckInHunbuStep is running, taskId: %s, hostClusterID: %s, subClusterID: %s, namespace: %s",
		t.GetTaskID(), hostClusterID, subClusterID, nsName)

	if nsName == "" || subClusterID == "" {
		return fmt.Errorf("CheckInHunbuStep task params error, namespace: %s, subClusterID: %s",
			nsName, subClusterID)
	}

	managedClusterLabelsStr, ok := t.GetCommonParams(fedsteps.ManagedClusterLabelsKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.ManagedClusterLabelsKey)
	}

	managedClusterLabelsMap := make(map[string]string)
	if err := json.Unmarshal([]byte(managedClusterLabelsStr), &managedClusterLabelsMap); err != nil {
		blog.Errorf("CheckInHunbuStep unmarshal managed cluster labels failed, subClusterID: %s, err: %s",
			subClusterID, err.Error())
		return fmt.Errorf("unmarshal managed cluster labels error: %s", err.Error())
	}

	hostAnnotations := make(map[string]string)
	if hostAnnotationsStr, exists := t.GetCommonParams(fedsteps.HostNamespaceAnnotationsKey); exists &&
		hostAnnotationsStr != "" {
		if err := json.Unmarshal([]byte(hostAnnotationsStr), &hostAnnotations); err != nil {
			blog.Errorf("CheckInHunbuStep unmarshal host namespace annotations failed, subClusterID: %s, "+
				"namespace: %s, err: %s", subClusterID, nsName, err.Error())
			return fmt.Errorf("unmarshal host namespace annotations error: %s", err.Error())
		}
	}

	namespace, err := cluster.GetClusterClient().GetNamespace(subClusterID, nsName)
	if err != nil && !errors.IsNotFound(err) {
		blog.Errorf("CheckInHunbuStep GetNamespace failed, subClusterID: %s, namespace: %s, err: %s",
			subClusterID, nsName, err.Error())
		return fmt.Errorf("GetNamespace(%s, %s) failed, err: %s", subClusterID, nsName, err.Error())
	}

	if namespace == nil {
		cerr := createHbNamespaceQuota(nsName, subClusterID, managedClusterLabelsMap, hostAnnotations)
		if cerr != nil {
			blog.Errorf("CheckInHunbuStep createHbNamespaceQuota failed, subClusterID: %s, namespace: %s, err: %s",
				subClusterID, nsName, cerr.Error())
			return fmt.Errorf("createHbNamespaceQuota failed, subClusterID: %s, namespace: %s, err: %s",
				subClusterID, nsName, cerr.Error())
		}
		blog.Infof("CheckInHunbuStep create success, taskId: %s, stepName: %s, subClusterID: %s, namespace: %s",
			t.GetTaskID(), step.GetName(), subClusterID, nsName)
		return nil
	}

	uerr := updateHbNamespaceQuota(subClusterID, namespace, managedClusterLabelsMap, hostAnnotations)
	if uerr != nil {
		blog.Errorf("CheckInHunbuStep updateHbNamespaceQuota failed, subClusterID: %s, namespace: %s, err: %s",
			subClusterID, nsName, uerr.Error())
		return fmt.Errorf("updateHbNamespaceQuota failed, subClusterID: %s, namespace: %s, err: %s",
			subClusterID, nsName, uerr.Error())
	}

	blog.Infof("CheckInHunbuStep update success, taskId: %s, stepName: %s, subClusterID: %s, namespace: %s",
		t.GetTaskID(), step.GetName(), subClusterID, nsName)
	return nil
}

func createHbNamespaceQuota(nsName, subClusterID string, labelsMap, hostAnnotations map[string]string) error {
	annotations := buildHbReq(labelsMap)

	projectAnnotations := buildNormalSubClusterAnnotations(hostAnnotations)
	for k, v := range projectAnnotations {
		annotations[k] = v
	}

	blog.Infof("createHbNamespaceQuota, subClusterID: %s, namespace: %s, annotations: %v",
		subClusterID, nsName, annotations)

	cerr := cluster.GetClusterClient().CreateClusterNamespace(subClusterID, nsName, annotations)
	if cerr != nil {
		blog.Errorf("createHbNamespaceQuota failed, subClusterID: %s, namespace: %s, err: %s",
			subClusterID, nsName, cerr.Error())
		return cerr
	}

	blog.Infof("createHbNamespaceQuota success, subClusterID: %s, namespace: %s", subClusterID, nsName)
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

func updateHbNamespaceQuota(subClusterId string, namespace *corev1.Namespace,
	labels, hostAnnotations map[string]string) error {
	blog.Infof("updateHbNamespaceQuota, subClusterID: %s, namespace: %s", subClusterId, namespace.Name)

	if namespace.Annotations == nil {
		namespace.Annotations = make(map[string]string)
	}

	reqAnnotations := buildHbReq(labels)
	for k, v := range reqAnnotations {
		namespace.Annotations[k] = v
	}

	projectAnnotations := buildNormalSubClusterAnnotations(hostAnnotations)
	for k, v := range projectAnnotations {
		namespace.Annotations[k] = v
	}

	cerr := cluster.GetClusterClient().UpdateNamespace(subClusterId, namespace)
	if cerr != nil {
		blog.Errorf("updateHbNamespaceQuota failed, subClusterID: %s, namespace: %s, err: %s",
			subClusterId, namespace.Name, cerr.Error())
		return cerr
	}

	blog.Infof("updateHbNamespaceQuota success, subClusterID: %s, namespace: %s", subClusterId, namespace.Name)
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
