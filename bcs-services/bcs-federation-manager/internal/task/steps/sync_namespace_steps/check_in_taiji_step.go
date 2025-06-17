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
	// CheckInTaijiStepName step name for check namespace quota in taiji
	CheckInTaijiStepName = fedsteps.StepNames{
		Alias: "check namespace quota in taiji",
		Name:  "CHECK_NAMESPACE_QUOTA_IN_TAIJI",
	}
)

// NewCheckInTaijiStep x
func NewCheckInTaijiStep() *CheckInTaijiStep {
	return &CheckInTaijiStep{}
}

// CheckInTaijiStep x
type CheckInTaijiStep struct{}

// Alias step name
func (s CheckInTaijiStep) Alias() string {
	return CheckInTaijiStepName.Alias
}

// GetName step name
func (s CheckInTaijiStep) GetName() string {
	return CheckInTaijiStepName.Name
}

// DoWork for worker exec task
func (s CheckInTaijiStep) DoWork(t *types.Task) error {
	blog.Infof("CheckInTaijiStep is running")

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
		blog.Errorf("get namespace quota task params error, namespace: %s, hostClusterId: %s",
			nsName, hostClusterID)
		return fmt.Errorf("get namespace quota task params error, namespace: %s, hostClusterId: %s",
			nsName, hostClusterID)
	}

	namespace := &corev1.Namespace{}
	nerr := json.Unmarshal([]byte(nsStr), &namespace)
	if nerr != nil {
		blog.Errorf("unmarshal namespace failed, namespace: %s, hostClusterId: %s, nsStr: %s, err: %s",
			nsName, hostClusterID, nsStr, nerr.Error())
		return fmt.Errorf("unmarshal namespace failed, namespace: %s, hostClusterId: %s, err: %s",
			nsName, hostClusterID, nerr.Error())
	}

	multiClusterResourceQuotaList := make([]federationv1.MultiClusterResourceQuota, 0)
	quotaListStr, ok := t.GetCommonParams(fedsteps.NamespaceQuotaListKey)
	if ok {
		merr := json.Unmarshal([]byte(quotaListStr), &multiClusterResourceQuotaList)
		if merr != nil {
			blog.Errorf("unmarshal namespace quota list failed, namespace: %s, hostClusterId: %s, nsStr: %s, err: %s",
				nsName, hostClusterID, nsStr, merr.Error())
			return fmt.Errorf("unmarshal namespace quota list failed, namespace: %s, hostClusterId: %s, err: %s",
				nsName, hostClusterID, merr.Error())
		}
	}

	// 查询thirdParty 查询 ns 是否已经存在
	resp, err := third.GetThirdpartyClient().GetKubeConfigForTaiji(nsName)
	if err != nil {
		blog.Errorf("getKubeConfigForTaiji failed namespace: %s, err: %s", nsName, err.Error())
		return fmt.Errorf("getKubeConfigForTaiji failed namespace: %s, err: %s", nsName, err.Error())
	}

	// 处理第三方服务返回的错误
	if resp != nil && resp.Error != nil {
		// 当namespace未注册时，才去新增
		if strings.Contains(resp.Error.Message, "namespace not register") {
			// 创建taiji namespace
			cerr := createTjNamespaceQuota(hostClusterID, namespace, multiClusterResourceQuotaList)
			if cerr != nil {
				blog.Errorf("createTjNamespaceQuota failed, namespace: %s, err: %s", nsName, cerr.Error())
				return fmt.Errorf("createTjNamespaceQuota failed, namespace: %s, err: %s", nsName, cerr.Error())
			}
			blog.Infof("CheckInTaijiStep Success, taskId: %s, taskName: %s, namespace: %s, hostClusterId: %s",
				t.GetTaskID(), step.GetName(), nsName, hostClusterID)
			return nil
		}
		// 其他错误类型直接返回，不执行更新逻辑
		blog.Errorf("getKubeConfigForTaiji returned unexpected error, namespace: %s, err: %s", nsName, resp.Error.Message)
		return fmt.Errorf("thirdparty service error: %s", resp.Error.Message)
	}

	// 更新taiji namespace
	uerr := updateTjNamespaceQuota(namespace, multiClusterResourceQuotaList)
	if uerr != nil {
		blog.Errorf("updateTjNamespaceQuota failed, namespace: %s, err: %s", nsName, uerr.Error())
		return uerr
	}

	blog.Infof("CheckInTaijiStep Success, taskId: %s, taskName: %s, namespace: %s, hostClusterId: %s",
		t.GetTaskID(), step.GetName(), nsName, hostClusterID)
	return nil
}

func updateTjNamespaceQuota(namespace *corev1.Namespace, quotaList []federationv1.MultiClusterResourceQuota) error {

	blog.Infof("updateTjNamespaceQuota, namespace: %s, quotaListLen: %v", namespace.Name, len(quotaList))
	annotations := namespace.Annotations
	req, err := buildTjUpdateReq(namespace.Name, annotations[cluster.AnnotationSubClusterForTaiji], quotaList)
	if err != nil {
		blog.Errorf("buildTjCreateReq failed, namespace: %s, err: %s", namespace.Name, err.Error())
		return err
	}

	terr := third.GetThirdpartyClient().UpdateQuotaInfoForTaiji(req)
	if terr != nil {
		blog.Errorf("updateTjNamespaceQuota failed, namespace: %s, err: %s", namespace.Name, terr.Error())
		return terr
	}

	blog.Infof("updateTjNamespaceQuota namespace %s Success", namespace.Name)
	return nil
}

func buildTjUpdateReq(nsName, location string, mcResourceQuotas []federationv1.MultiClusterResourceQuota) (
	*trd.UpdateQuotaInfoForTaijiRequest, error) {

	quotas := make([]*trd.NamespaceQuotaForTaiji, 0)
	for _, mcResourceQuota := range mcResourceQuotas {
		if val, ok := mcResourceQuota.Annotations[cluster.AnnotationSubClusterForTaiji]; ok {
			location = val
			quotaResources := make(map[string]string)
			if mcResourceQuota.Spec.TotalQuota.Hard != nil {
				for name, quantity := range mcResourceQuota.Spec.TotalQuota.Hard {
					quotaResources[string(name)] = quantity.String()
				}
			}
			// 转换为taiji参数 GPUName
			tjAttributes := make(map[string]string)
			for k, v := range mcResourceQuota.Spec.TaskSelector {
				if k == cluster.TaskGpuTypeKey {
					tjAttributes[cluster.TaijiGPUNameKey] = v
					continue
				}
				tjAttributes[k] = v
			}
			quotaInfo := &trd.NamespaceQuotaForTaiji{
				Name:              mcResourceQuota.Name,
				SubQuotaLabels:    tjAttributes,
				SubQuotaResources: quotaResources,
				Location:          val,
			}
			quotas = append(quotas, quotaInfo)
		}
	}

	// 只有quota变化才需要更新
	if len(quotas) == 0 || (len(quotas) > 0 && location == "") {
		blog.Errorf("buildUpdateReq failed, namespace: %s, err: location is empty", nsName)
		return nil, fmt.Errorf("buildUpdateReq failed, namespace: %s, err: location is empty", nsName)
	}

	return &trd.UpdateQuotaInfoForTaijiRequest{
		Namespace:     nsName,
		SubQuotaInfos: quotas,
		Location:      location,
		Operator:      "admin",
	}, nil
}

func createTjNamespaceQuota(hostClusterId string, namespace *corev1.Namespace,
	quotaList []federationv1.MultiClusterResourceQuota) error {

	blog.Infof("createTjNamespaceQuota, namespace: %s, quotaListLen: %v", namespace.Name, len(quotaList))
	req, err := buildTjCreateReq(hostClusterId, namespace, quotaList)
	if err != nil {
		blog.Errorf("buildTjCreateReq failed, namespace: %s, err: %s", namespace.Name, err.Error())
		return err
	}

	cerr := third.GetThirdpartyClient().CreateNamespaceForTaijiV3(req)
	if cerr != nil {
		blog.Errorf("createTaijiNamespaceV3 failed, namespace: %s, err: %s", namespace.Name, cerr.Error())
		return cerr
	}

	blog.Infof("createTjNamespaceQuota namespace %s Success", namespace.Name)
	return nil
}

func buildTjCreateReq(hostClusterId string, namespace *corev1.Namespace,
	mcResourceQuotas []federationv1.MultiClusterResourceQuota) (*trd.CreateNamespaceForTaijiV3Request, error) {

	annotations := namespace.Annotations
	location := annotations[cluster.AnnotationSubClusterForTaiji]
	isPrivateResource := false
	if annotations[cluster.AnnotationIsPrivateResourceKey] == "true" {
		isPrivateResource = true
	}
	scheduleAlgorithm := annotations[cluster.AnnotationScheduleAlgorithmKey]
	if scheduleAlgorithm == "" {
		scheduleAlgorithm = cluster.AnnotationScheduleAlgorithmValue
	}
	quotas := make([]*trd.NamespaceQuotaForTaiji, 0)
	for _, mcResourceQuota := range mcResourceQuotas {
		if val, ok := mcResourceQuota.Annotations[cluster.AnnotationSubClusterForTaiji]; ok {
			location = val
			quotaResources := make(map[string]string)
			if mcResourceQuota.Spec.TotalQuota.Hard != nil {
				for name, quantity := range mcResourceQuota.Spec.TotalQuota.Hard {
					quotaResources[string(name)] = quantity.String()
				}
			}
			// 转换为taiji参数 GPUName
			tjAttributes := make(map[string]string)
			for k, v := range mcResourceQuota.Spec.TaskSelector {
				if k == cluster.TaskGpuTypeKey {
					tjAttributes[cluster.TaijiGPUNameKey] = v
					continue
				}
				tjAttributes[k] = v
			}
			quotaInfo := &trd.NamespaceQuotaForTaiji{
				Name:              mcResourceQuota.Name,
				SubQuotaLabels:    tjAttributes,
				SubQuotaResources: quotaResources,
				Location:          val,
			}
			quotas = append(quotas, quotaInfo)
		}
	}
	// 判断location是否为空
	if len(quotas) > 0 && location == "" {
		blog.Errorf("buildTjCreateReq failed, namespace: %s, err: location is empty", namespace.Name)
		return nil, fmt.Errorf("buildTjCreateReq failed, namespace: %s, err: location is empty", namespace.Name)
	}

	var bkBizId, bkModuleId = annotations[cluster.FedNamespaceBkBizId], annotations[cluster.FedNamespaceBkModuleId]
	if bkModuleId == "" || bkBizId == "" {
		// 请求taiji createModule api
		result, err := third.GetThirdpartyClient().CreateModule(namespace.Name)
		if err != nil {
			blog.Errorf("buildTjCreateReq failed, namespace: %s, err: %s", namespace.Name, err.Error())
			return nil, err
		}
		bkBizId = fmt.Sprintf("%d", result.Data.BkBizId)
		bkModuleId = fmt.Sprintf("%d", result.Data.BkModuleId)
		annotations[cluster.FedNamespaceBkBizId] = bkBizId
		annotations[cluster.FedNamespaceBkModuleId] = bkModuleId
		annotations[cluster.NamespaceUpdateTimestamp] = time.Now().Format(time.RFC3339)
		namespace.Annotations = annotations
		err = cluster.GetClusterClient().UpdateNamespace(hostClusterId, namespace)
		if err != nil {
			return nil, err
		}
	}

	return &trd.CreateNamespaceForTaijiV3Request{
		Location:          location,
		Namespace:         namespace.Name,
		Creator:           "admin",
		ScheduleAlgorithm: &scheduleAlgorithm,
		SubQuotaInfos:     quotas,
		IsPrivateResource: &isPrivateResource,
		BkBizId:           bkBizId,
		BkModuleId:        bkModuleId,
	}, nil
}

// BuildStep build step
func (s CheckInTaijiStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	// stepName/s.GetName() 用于标识这个step
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
