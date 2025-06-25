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

	corev1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	third "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/thirdparty"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
	trd "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/pkg/bcsapi/thirdparty-service"
	federationv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/pkg/kubeapi/federationquota/api/v1"
)

var (
	// CheckInSuanliStepName step name for check namespace quota in suanli
	CheckInSuanliStepName = fedsteps.StepNames{
		Alias: "check namespace quota in suanli",
		Name:  "CHECK_NAMESPACE_QUOTA_IN_SUANLI",
	}
)

// NewCheckInSuanliStep x
func NewCheckInSuanliStep() *CheckInSuanliStep {
	return &CheckInSuanliStep{}
}

// CheckInSuanliStep x
type CheckInSuanliStep struct{}

// Alias step name
func (s CheckInSuanliStep) Alias() string {
	return CheckInSuanliStepName.Alias
}

// GetName step name
func (s CheckInSuanliStep) GetName() string {
	return CheckInSuanliStepName.Name
}

// DoWork for worker exec task
func (s CheckInSuanliStep) DoWork(t *types.Task) error {
	blog.Infof("CheckInSuanliStep is running")

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

	nsStr, ok := t.GetCommonParams(fedsteps.SyncNamespaceQuotaKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.SyncNamespaceQuotaKey)
	}

	if nsStr == "" {
		blog.Errorf("getNamespaceQuota task params error, namespace: %s, hostClusterId: %s",
			nsName, hostClusterID)
		return fmt.Errorf("getNamespaceQuota task params error, namespace: %s, hostClusterId: %s",
			nsName, hostClusterID)
	}

	namespace := &corev1.Namespace{}
	nerr := json.Unmarshal([]byte(nsStr), &namespace)
	if nerr != nil {
		blog.Errorf("unmarshal namespace failed, namespace: %s, hostClusterId: %s, err: %s",
			nsName, hostClusterID, nerr.Error())
		return fmt.Errorf("unmarshal namespace failed, namespace: %s, hostClusterId: %s, err: %s",
			nsName, hostClusterID, nerr.Error())
	}
	if namespace == nil {
		return fmt.Errorf("getNamespaceQuota task params error, namespace: %s, hostClusterId: %s",
			nsName, hostClusterID)
	}

	multiClusterResourceQuotaList := make([]federationv1.MultiClusterResourceQuota, 0)
	quotaListStr, ok := t.GetCommonParams(fedsteps.NamespaceQuotaListKey)
	if ok {
		merr := json.Unmarshal([]byte(quotaListStr), &multiClusterResourceQuotaList)
		if merr != nil {
			blog.Errorf("unmarshal multiClusterResourceQuotaList failed, namespace: %s, hostClusterId: %s, err: %s",
				nsName, hostClusterID, merr.Error())
			return fmt.Errorf("unmarshal multiClusterResourceQuotaList failed, namespace: %s, hostClusterId: %s, err: %s",
				nsName, hostClusterID, merr.Error())
		}
	}

	// 查询thirdParty 查询 ns 是否已经存在
	resp, err := third.GetThirdpartyClient().GetKubeConfigForSuanli(nsName)
	if err != nil {
		blog.Errorf("getKubeConfigForSuanli failed namespace: %s, err: %s", nsName, err.Error())
		return fmt.Errorf("getKubeConfigForSuanli failed namespace: %s, err: %s", nsName, err.Error())
	}

	if resp == nil {
		blog.Infof("getKubeConfigForSuanli failed, namespace: %s, hostClusterId: %s", nsName, hostClusterID)
		return fmt.Errorf("getKubeConfigForSuanli failed, namespace: %s, hostClusterId: %s", nsName, hostClusterID)
	}

	// 当namespace未注册时，才去新增
	if strings.Contains(resp.Message, "namespace not register") {
		// 创建 suanli namespace quota
		cerr := createSlNamespaceQuota(hostClusterID, namespace, multiClusterResourceQuotaList)
		if cerr != nil {
			blog.Errorf("createSlNamespaceQuota failed, namespace: %s, err: %s", nsName, cerr.Error())
			return fmt.Errorf("createSlNamespaceQuota failed, namespace: %s, err: %s", nsName, cerr.Error())
		}
		blog.Infof("CheckInSuanliStep Success, taskId: %s, taskName: %s, namespace: %s, hostClusterId: %s",
			t.GetTaskID(), step.GetName(), nsName, hostClusterID)
		return nil
	} else if resp.Message != "SUCCESS" {
		blog.Errorf("getKubeConfigForSuanli failed, namespace: %s, hostClusterId: %s, err: %s",
			nsName, hostClusterID, resp.Message)
		return fmt.Errorf("getKubeConfigForSuanli failed, namespace: %s, hostClusterId: %s, err: %s",
			nsName, hostClusterID, resp.Message)
	}

	// 更新 suanli namespace quota
	uerr := updateSlNamespaceQuota(nsName, multiClusterResourceQuotaList)
	if uerr != nil {
		blog.Errorf("updateSlNamespaceQuota failed, namespace: %s, err: %s", nsName, uerr.Error())
		return fmt.Errorf("updateSlNamespaceQuota failed, namespace: %s, err: %s", nsName, uerr.Error())
	}

	blog.Infof("CheckInSuanliStep Success, taskId: %s, taskName: %s, namespace: %s, hostClusterId: %s",
		t.GetTaskID(), step.GetName(), nsName, hostClusterID)
	return nil
}

func updateSlNamespaceQuota(nsName string, quotaList []federationv1.MultiClusterResourceQuota) error {
	blog.Infof("updateSlNamespaceQuota, namespace: %s, quotaListLen: %v", nsName, len(quotaList))

	req := buildSlUpdateReq(nsName, quotaList)
	terr := third.GetThirdpartyClient().UpdateQuotaInfoForSuanli(req)
	if terr != nil {
		blog.Errorf("updateSlNamespaceQuota failed, namespace: %s, err: %s", nsName, terr.Error())
		return terr
	}

	blog.Infof("updateSlNamespaceQuota namespace %s Success", nsName)
	return nil
}

func buildSlUpdateReq(nsName string, quotaList []federationv1.MultiClusterResourceQuota) *trd.UpdateNamespaceForSuanliRequest {

	// build quotas
	quotas := make([]*trd.NamespaceQuotaForSuanli, 0)
	for _, mcResourceQuota := range quotaList {
		if val, ok := mcResourceQuota.Annotations[cluster.AnnotationKeyInstalledPlatform]; ok {
			if val != cluster.SubClusterForSuanli {
				continue
			}
			quotaResources := make(map[string]string)
			if mcResourceQuota.Spec.TotalQuota.Hard != nil {
				for name, quantity := range mcResourceQuota.Spec.TotalQuota.Hard {
					quotaResources[string(name)] = quantity.String()
				}
			}

			quotaInfo := &trd.NamespaceQuotaForSuanli{
				Name:              mcResourceQuota.Name,
				SubQuotaLabels:    mcResourceQuota.Spec.TaskSelector,
				SubQuotaResources: quotaResources,
			}

			quotas = append(quotas, quotaInfo)
		}
	}

	return &trd.UpdateNamespaceForSuanliRequest{
		Namespace:     nsName,
		Operator:      "admin",
		SubQuotaInfos: quotas,
	}
}

func createSlNamespaceQuota(hostClusterID string, namespace *corev1.Namespace,
	quotaList []federationv1.MultiClusterResourceQuota) error {

	req, err := buildSlCreateReq(hostClusterID, namespace, quotaList)
	if err != nil {
		blog.Errorf("buildSlCreateReq failed, namespace: %s, err: %s", namespace.Name, err.Error())
		return err
	}

	terr := third.GetThirdpartyClient().CreateNamespaceForSuanli(req)
	if terr != nil {
		blog.Errorf("CreateNamespaceForSuanli failed, namespace: %s, err: %s", namespace.Name, terr.Error())
		return terr
	}

	blog.Infof("createSlNamespaceQuota namespace %s Success", namespace.Name)
	return nil

}

func buildSlCreateReq(hostClusterId string, namespace *corev1.Namespace,
	quotaList []federationv1.MultiClusterResourceQuota) (*trd.CreateNamespaceForSuanliRequest, error) {

	// build quotas
	quotas := make([]*trd.NamespaceQuotaForSuanli, 0)
	for _, mcResourceQuota := range quotaList {
		if val, ok := mcResourceQuota.Annotations[cluster.AnnotationKeyInstalledPlatform]; ok {
			if val != cluster.SubClusterForSuanli {
				continue
			}
			quotaResources := make(map[string]string)
			if mcResourceQuota.Spec.TotalQuota.Hard != nil {
				for name, quantity := range mcResourceQuota.Spec.TotalQuota.Hard {
					quotaResources[string(name)] = quantity.String()
				}
			}

			quotaInfo := &trd.NamespaceQuotaForSuanli{
				Name:              mcResourceQuota.Name,
				SubQuotaLabels:    mcResourceQuota.Spec.TaskSelector,
				SubQuotaResources: quotaResources,
			}

			quotas = append(quotas, quotaInfo)
		}
	}

	annotations := namespace.Annotations
	var bkBizId, bkModuleId = annotations[cluster.FedNamespaceBkBizId], annotations[cluster.FedNamespaceBkModuleId]
	if bkModuleId == "" || bkBizId == "" {
		// 请求taiji createModule api
		result, terr := third.GetThirdpartyClient().CreateModule(namespace.Name)
		if terr != nil {
			blog.Errorf("buildTjCreateReq failed, namespace: %s, err: %s", namespace.Name, terr.Error())
			return nil, terr
		}
		bkBizId = fmt.Sprintf("%d", result.Data.BkBizId)
		bkModuleId = fmt.Sprintf("%d", result.Data.BkModuleId)
		annotations[cluster.FedNamespaceBkBizId] = bkBizId
		annotations[cluster.FedNamespaceBkModuleId] = bkModuleId
		annotations[cluster.NamespaceUpdateTimestamp] = time.Now().Format(time.RFC3339)
		namespace.Annotations = annotations
		cerr := cluster.GetClusterClient().UpdateNamespace(hostClusterId, namespace)
		if cerr != nil {
			blog.Errorf("updateNamespace failed, namespace: %s, err: %s", namespace.Name, cerr.Error())
			return nil, cerr
		}
	}

	return &trd.CreateNamespaceForSuanliRequest{
		Namespace:     namespace.Name,
		Creator:       "admin",
		SubQuotaInfos: quotas,
		BkBizId:       bkBizId,
		BkModuleId:    bkModuleId,
	}, nil
}

// BuildStep build step
func (s CheckInSuanliStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	// stepName/s.GetName() 用于标识这个step
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
