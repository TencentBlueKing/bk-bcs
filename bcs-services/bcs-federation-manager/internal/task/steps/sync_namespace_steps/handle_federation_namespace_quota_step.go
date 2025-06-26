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
	"time"

	"github.com/avast/retry-go"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
	federationv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/pkg/kubeapi/federationquota/api/v1"
	federationmgr "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/proto/bcs-federation-manager"
)

var (
	// HandleFederationNamespaceQuotaStepName step name for create cluster
	HandleFederationNamespaceQuotaStepName = fedsteps.StepNames{
		Alias: "handle federation namespace quota",
		Name:  "HANDLE_FEDERATION_NAMESPACE_QUOTA",
	}
)

// NewHandleFederationNamespaceQuotaStep new step
// NOCC:tosa/fn_length(设计如此)
func NewHandleFederationNamespaceQuotaStep() *HandleFederationNamespaceQuotaStep {
	return &HandleFederationNamespaceQuotaStep{}
}

// HandleFederationNamespaceQuotaStep x
type HandleFederationNamespaceQuotaStep struct{}

// Alias step name
func (s HandleFederationNamespaceQuotaStep) Alias() string {
	return HandleFederationNamespaceQuotaStepName.Alias
}

// GetName step name
func (s HandleFederationNamespaceQuotaStep) GetName() string {
	return HandleFederationNamespaceQuotaStepName.Name
}

// DoWork for worker exec task
func (s HandleFederationNamespaceQuotaStep) DoWork(t *types.Task) error {

	step, exist := t.GetStep(s.GetName())
	if !exist {
		return fmt.Errorf("task[%s] not exist step[%s]", t.TaskID, s.GetName())
	}

	hostClusterId, ok := step.GetParam(fedsteps.HostClusterIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.HostClusterIdKey)
	}

	namespace, ok := step.GetParam(fedsteps.NamespaceKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.NamespaceKey)
	}

	opt, ok := step.GetParam(fedsteps.ParameterKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, "quota parameter")
	}

	handleType, ok := step.GetParam(fedsteps.HandleTypeKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, "quota handleType")
	}

	blog.Infof("quota task hostClusterId: %s; namespace: %s; handleType: %s; opt: %s", hostClusterId, namespace, handleType, opt)

	switch handleType {
	case fedsteps.CreateKey:
		var quotaList []*federationmgr.Quota
		err := json.Unmarshal([]byte(opt), &quotaList)
		if err != nil {
			blog.Errorf(
				"task[%s] handle create quota.Unmarshal failed "+
					"body: %s, "+
					"err: %s", t.TaskID, opt, err.Error())
			return err
		}

		if err := retry.Do(func() error {
			for _, quota := range quotaList {
				hardList := v1.ResourceList{}
				for _, k8SResource := range quota.ResourceList {
					resourceName := v1.ResourceName(k8SResource.ResourceName)
					resourceQuantity := resource.MustParse(k8SResource.ResourceQuantity)
					hardList[resourceName] = resourceQuantity
				}

				mcResourceQuota := &federationv1.MultiClusterResourceQuota{
					TypeMeta: metav1.TypeMeta{
						Kind:       "MultiClusterResourceQuota",
						APIVersion: "federation.bkbcs.tencent.com/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:              quota.Name,
						Namespace:         namespace,
						CreationTimestamp: metav1.Now(),
						Annotations:       quota.Annotations,
					},
					Spec: federationv1.MultiClusterResourceQuotaSpec{
						TotalQuota: federationv1.MultiClusterResourceQuotaTotalQuotaSpec{
							Hard: hardList,
						},
						TaskSelector: quota.Attributes,
					},
				}

				bytes, err := json.Marshal(mcResourceQuota)
				if err != nil {
					return err
				}

				blog.Infof(
					"DoWork.CreateFedClusterResourceQuota "+
						"创建 MultiClusterResourceQuota namespace: %s, hostClusterId: %s, mcResourceQuota: %s",
					namespace, hostClusterId, string(bytes))

				err = cluster.GetClusterClient().CreateMultiClusterResourceQuota(hostClusterId, namespace, mcResourceQuota)
				if err != nil {
					blog.Errorf(
						"task[%s] create quota.CreateMultiClusterResourceQuota failed "+
							"hostClusterId: %s, "+
							"namespace: %s, "+
							"mcResourceQuota: %+v, "+
							"err: %s", t.TaskID, hostClusterId, namespace, mcResourceQuota, err.Error())
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
func (s HandleFederationNamespaceQuotaStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	// stepName/s.GetName() 用于标识这个step
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
