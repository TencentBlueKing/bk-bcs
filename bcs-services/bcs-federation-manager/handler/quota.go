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

// Package handler quota service
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/store"
	fedtasks "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/tasks"
	trd "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/pkg/bcsapi/thirdparty-service"
	federationmgr "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/proto/bcs-federation-manager"
)

// GetFederationClusterNamespaceQuota get quota
func (f *FederationManager) GetFederationClusterNamespaceQuota(ctx context.Context,
	req *federationmgr.GetFederationClusterNamespaceQuotaRequest,
	resp *federationmgr.GetFederationClusterNamespaceQuotaResponse) error {

	blog.Infof("Received BcsFederationManager.GetFederationClusterNamespaceQuota request, req: %v", req)

	// validate
	if err := req.Validate(); err != nil {
		return ErrReturn(resp, fmt.Sprintf("validate GetFederationClusterNamespaceQuota request failed, err: %s",
			err.Error()))
	}

	// 根据联邦集群代理ID获取 联邦集群ID
	fedCluster, err := store.GetStoreModel().GetFederationCluster(context.Background(), req.ClusterId)
	if err != nil {
		return ErrReturn(resp,
			fmt.Sprintf("GetFederationClusterNamespaceQuota.GetFederationCluster failed, "+
				"clusterId: %s, err: %s", req.ClusterId, err.Error()))
	}

	// get quota
	mcResourceQuota, err := f.clusterCli.GetNamespaceQuota(fedCluster.HostClusterID, req.Namespace, req.Name)
	if err != nil {
		return ErrReturn(resp,
			fmt.Sprintf("GetNamespaceQuota failed, namespace: %s, quotaName: %s, err: %s",
				req.Namespace, req.Name, err.Error()))
	}

	if mcResourceQuota == nil {
		resp.Code = IntToUint32Ptr(common.BcsSuccess)
		resp.Message = common.BcsSuccessStr
		resp.Data = nil
		return nil
	}

	// build quota
	quotaData := new(federationmgr.Quota)
	quotaData.Name = mcResourceQuota.Name
	quotaData.Attributes = mcResourceQuota.Spec.TaskSelector
	quotaData.Annotations = mcResourceQuota.Annotations
	k8SResources := make([]*federationmgr.K8SResource, 0)
	if mcResourceQuota.Spec.TotalQuota.Hard != nil {
		// build k8s resource list
		for name, quantity := range mcResourceQuota.Spec.TotalQuota.Hard {
			res := &federationmgr.K8SResource{
				ResourceName:     string(name),
				ResourceQuantity: quantity.String(),
			}
			k8SResources = append(k8SResources, res)
		}
		quotaData.ResourceList = k8SResources
	}

	jsonData, err := json.MarshalIndent(mcResourceQuota, "", "  ")
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("GetFederationClusterNamespaceQuota.MarshalIndent failed, err: %s", err.Error()))
	}
	str := string(jsonData)
	quotaData.OriginK8SData = &str

	// success
	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	resp.Data = quotaData
	return nil
}

// ListFederationClusterNamespaceQuota list federation cluster ns quotas list
func (f *FederationManager) ListFederationClusterNamespaceQuota(ctx context.Context,
	req *federationmgr.ListFederationClusterNamespaceQuotaRequest,
	resp *federationmgr.ListFederationClusterNamespaceQuotaResponse) error {

	blog.Infof("Received BcsFederationManager.ListFederationClusterNamespaceQuota request, req: %v", req)

	// validate
	if err := req.Validate(); err != nil {
		return ErrReturn(resp, fmt.Sprintf("validate ListFederationClusterNamespaceQuota request failed, err: %s",
			err.Error()))
	}

	// 根据联邦集群代理ID获取 联邦集群ID
	fedCluster, err := store.GetStoreModel().GetFederationCluster(context.Background(), req.ClusterId)
	if err != nil {
		return ErrReturn(resp,
			fmt.Sprintf("ListFederationClusterNamespaceQuota.GetFederationCluster failed, "+
				"clusterId: %s, err: %s", req.ClusterId, err.Error()))
	}

	mcResourceQuotas, err := f.clusterCli.ListNamespaceQuota(fedCluster.HostClusterID, req.Namespace)
	if err != nil {
		return ErrReturn(resp,
			fmt.Sprintf("ListNamespaceQuota failed, namespace: %s, err: %s", req.Namespace, err.Error()))
	}

	// build quotas
	quotas := make([]*federationmgr.Quota, 0)
	for _, mcResourceQuota := range mcResourceQuotas.Items {
		quotaData := new(federationmgr.Quota)
		quotaData.Name = mcResourceQuota.Name
		quotaData.Attributes = mcResourceQuota.Spec.TaskSelector
		quotaData.Annotations = mcResourceQuota.Annotations
		k8SResources := make([]*federationmgr.K8SResource, 0)
		if mcResourceQuota.Spec.TotalQuota.Hard != nil {
			// build k8s resource list
			for name, quantity := range mcResourceQuota.Spec.TotalQuota.Hard {
				res := &federationmgr.K8SResource{
					ResourceName:     string(name),
					ResourceQuantity: quantity.String(),
				}
				k8SResources = append(k8SResources, res)
			}
			quotaData.ResourceList = k8SResources
		}

		jsonData, err := json.MarshalIndent(mcResourceQuota, "", "  ")
		if err != nil {
			return ErrReturn(resp,
				fmt.Sprintf("ListNamespaceQuota.MarshalIndent failed,err: %s", err.Error()))
		}
		str := string(jsonData)
		quotaData.OriginK8SData = &str

		quotas = append(quotas, quotaData)
	}

	// success
	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	resp.Data = quotas
	return nil
}

// CreateFederationClusterNamespaceQuota create federation quota and sub quota
// NOCC:tosa/fn_length(设计如此)
func (f *FederationManager) CreateFederationClusterNamespaceQuota(ctx context.Context,
	req *federationmgr.CreateFederationClusterNamespaceQuotaRequest,
	resp *federationmgr.CreateFederationClusterNamespaceQuotaResponse) error {

	blog.Infof("Received BcsFederationManager.CreateFederationClusterNamespaceQuota request, req: %v", req)

	// validate
	if err := req.Validate(); err != nil {
		return ErrReturn(resp, fmt.Sprintf("validate CreateFederationClusterNamespaceQuota request failed, err: %s", err.Error()))
	}

	// 根据联邦集群代理ID获取 联邦集群ID
	fedCluster, err := store.GetStoreModel().GetFederationCluster(context.Background(), req.ClusterId)
	if err != nil {
		return ErrReturn(resp,
			fmt.Sprintf("CreateFederationClusterNamespaceQuota.GetFederationCluster "+
				"get federation cluster from federationmanager failed, "+
				"clusterId: %s, err: %s", req.ClusterId, err.Error()))
	}

	// 查询是否已存在联邦集群命名空间
	fedClusterNamespace, err := f.clusterCli.GetNamespace(fedCluster.HostClusterID, req.Namespace)
	if err != nil {
		return fmt.Errorf("get federation namespace failed, "+
			"clusterId: %s, namespace: %s, err: %s", req.ClusterId, req.Namespace, err.Error())
	}

	// create quota
	err = f.clusterCli.CreateNamespaceQuota(fedCluster.HostClusterID, req)
	if err != nil {
		blog.Errorf("CreateNamespaceQuota failed req: %+v, err: %s", req, err.Error())
		return ErrReturn(resp, fmt.Sprintf("CreateNamespaceQuota failed, err: %s", err.Error()))
	}

	// build subQuotaInfos
	err = f.taskUpdateQuotas(req, fedCluster, fedClusterNamespace)
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("CreateNamespaceQuota build task failed, err: %s", err.Error()))
	}

	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr

	return nil
}

// UpdateFederationClusterNamespaceQuota update federation quota and sub quota
// NOCC:tosa/fn_length(设计如此)
func (f *FederationManager) UpdateFederationClusterNamespaceQuota(ctx context.Context,
	req *federationmgr.UpdateFederationClusterNamespaceQuotaRequest,
	resp *federationmgr.UpdateFederationClusterNamespaceQuotaResponse) error {

	blog.Infof("Received BcsFederationManager.UpdateNamespaceQuota request, req: %v", req)
	if err := req.Validate(); err != nil {
		return ErrReturn(resp, fmt.Sprintf("validate UpdateNamespaceQuota request failed, err: %s", err.Error()))
	}

	// 根据联邦集群代理ID获取 联邦集群ID
	fedCluster, err := store.GetStoreModel().GetFederationCluster(context.Background(), req.ClusterId)
	if err != nil {
		return ErrReturn(resp,
			fmt.Sprintf("UpdateNamespaceQuota.GetFederationCluster failed, clusterId: %s, err: %s",
				req.ClusterId, err.Error()))
	}

	// 查询是否已存在联邦集群命名空间
	fedClusterNamespace, err := f.clusterCli.GetNamespace(fedCluster.HostClusterID, req.Namespace)
	if err != nil {
		return fmt.Errorf("get federation namespace failed, "+
			"clusterId: %s, namespace: %s, err: %s", req.ClusterId, req.Namespace, err.Error())
	}
	mcResourceQuota, err := f.clusterCli.GetMultiClusterResourceQuota(fedCluster.HostClusterID, req.Namespace, req.Name)
	if err != nil {
		blog.Errorf(
			"UpdateNamespaceQuota.GetMultiClusterResourceQuota failed clusterId: %s, namespace: %s, quotaName: %s, "+
				"err: %s", fedCluster.HostClusterID, req.Namespace, req.Name, err.Error())
		return ErrReturn(resp, fmt.Sprintf("GetMultiClusterResourceQuota failed, err: %s", err.Error()))
	}
	if mcResourceQuota.Spec.TaskSelector == nil {
		mcResourceQuota.Spec.TaskSelector = req.Quota.Attributes
	} else {
		for key, val := range req.Quota.Attributes {
			mcResourceQuota.Spec.TaskSelector[key] = val
		}
	}
	if mcResourceQuota.Annotations == nil {
		mcResourceQuota.Annotations = req.Quota.Annotations
	} else {
		for key, val := range req.Quota.Annotations {
			mcResourceQuota.Annotations[key] = val
		}
	}
	for _, res := range req.Quota.ResourceList {
		resName := v1.ResourceName(res.ResourceName)
		quantity, iErr := resource.ParseQuantity(res.ResourceQuantity)
		if iErr != nil {
			return iErr
		}
		mcResourceQuota.Spec.TotalQuota.Hard[resName] = quantity
	}

	// update quota
	err = f.clusterCli.UpdateNamespaceQuota(fedCluster.HostClusterID, req.Namespace, mcResourceQuota)
	if err != nil {
		blog.Errorf(
			"UpdateNamespaceQuota failed clusterId: %s, namespace: %s, mcResourceQuota: %v, err: %s",
			fedCluster.HostClusterID, req.Namespace, mcResourceQuota, err.Error())
		return ErrReturn(resp, fmt.Sprintf("UpdateNamespaceQuota failed, err: %s", err.Error()))
	}

	upload := &federationmgr.CreateFederationClusterNamespaceQuotaRequest{
		ClusterId: req.ClusterId,
		Namespace: req.Namespace,
		QuotaList: []*federationmgr.Quota{req.Quota},
		Operator:  req.Operator,
	}
	err = f.taskUpdateQuotas(upload, fedCluster, fedClusterNamespace)
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("UpdateNamespaceQuota build task failed, err: %s", err.Error()))
	}

	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	return nil
}

func (f *FederationManager) taskUpdateQuotas(req *federationmgr.CreateFederationClusterNamespaceQuotaRequest,
	fedCluster *store.FederationCluster, namespace *v1.Namespace) error {

	blog.Infof("taskUpdateQuotas request, req: %v, namespace: %v", req, namespace)
	// 获取入参的子集群范围
	subClusterIds := make([]string, 0)
	if clusterRangeStr, ok := namespace.Annotations[cluster.FedNamespaceClusterRangeKey]; ok {
		if len(clusterRangeStr) != 0 {
			lower := strings.Split(clusterRangeStr, ",")
			for _, sc := range lower {
				subClusterId := strings.ToUpper(sc)
				subClusterIds = append(subClusterIds, subClusterId)
			}
		}
	}

	forTaijiList, forSuanliList, err := f.initTaskUpdateQuotaParams(req, fedCluster, subClusterIds)
	if err != nil {
		return err
	}
	reqMap := make(map[string]string)
	if len(forTaijiList) > 0 {
		forTaijiListBytes, err := json.Marshal(forTaijiList)
		if err != nil {
			return fmt.Errorf("taskUpdateQuotas.Marshal failed,err: %s", err.Error())
		}
		reqMap[cluster.SubClusterForTaiji] = string(forTaijiListBytes)
	}

	if len(forSuanliList) > 0 {
		forSuanliListBytes, err := json.Marshal(forSuanliList)
		if err != nil {
			return fmt.Errorf("taskUpdateQuotas.Marshal failed,err: %s", err.Error())
		}
		reqMap[cluster.SubClusterForSuanli] = string(forSuanliListBytes)
	}

	if len(reqMap) == 0 {
		// 更新annotations中的状态
		if namespace.Annotations[cluster.HostClusterNamespaceStatus] != cluster.NamespaceSuccess {
			namespace.Annotations[cluster.HostClusterNamespaceStatus] = cluster.NamespaceSuccess
			err := f.clusterCli.UpdateNamespace(fedCluster.HostClusterID, namespace)
			if err != nil {
				blog.Errorf("taskUpdateQuotas update namespace failed, HostClusterID %s, req %+v",
					fedCluster.HostClusterID, req)
				return err
			}
		}
		blog.Errorf("taskUpdateQuotas taskParams is empty, HostClusterID %s, req %+v",
			fedCluster.HostClusterID, req)
	} else {
		reqParameterBytes, err := json.Marshal(reqMap)
		if err != nil {
			return fmt.Errorf("taskUpdateQuotas.Marshal failed,err: %s", err.Error())
		}
		blog.Infof("taskUpdateQuotas reqParameterBytes is %s", string(reqParameterBytes))

		// 更新quota
		t, err := fedtasks.NewHandleNamespaceQuotaTask(
			&fedtasks.HandleNamespaceQuotaOptions{
				FedClusterId:  fedCluster.FederationClusterID,
				HandleType:    cluster.UpdateKey,
				HostClusterId: fedCluster.HostClusterID,
				Namespace:     req.Namespace,
				Parameter:     string(reqParameterBytes),
			}).BuildTask(req.Operator)
		if err != nil {
			blog.Errorf(
				"taskUpdateQuotas build task failed clusterId: %s, namespace: %s, body: %s, err: %s",
				fedCluster.HostClusterID, req.Namespace, string(reqParameterBytes), err.Error())
			return fmt.Errorf("taskUpdateQuotas failed, err: %s", err.Error())
		}

		if err = f.taskmanager.Dispatch(t); err != nil {
			return fmt.Errorf("taskUpdateQuotas build task failed, err: %s", err.Error())
		}
	}

	return nil
}

func (f *FederationManager) initTaskUpdateQuotaParams(req *federationmgr.CreateFederationClusterNamespaceQuotaRequest,
	fedCluster *store.FederationCluster, subClusterIds []string) ([]*trd.UpdateQuotaInfoForTaijiRequest,
	[]*trd.UpdateNamespaceForSuanliRequest, error) {

	forTaijiList := make([]*trd.UpdateQuotaInfoForTaijiRequest, 0)
	forSuanliList := make([]*trd.UpdateNamespaceForSuanliRequest, 0)
	for _, subClusterId := range subClusterIds {
		// 获取managedCluster 对象
		managedCluster, err := f.clusterCli.GetManagedCluster(fedCluster.HostClusterID, subClusterId)
		if err != nil {
			blog.Errorf("initTaskUpdateQuotaParams failed, subClusterId: %s, err: %s", subClusterId, err.Error())
			return nil, nil, err
		}
		// 判断managedCluster对象的labels的标识
		switch managedCluster.Labels[cluster.ManagedClusterTypeLabel] {
		case cluster.SubClusterForTaiji:
			for _, quota := range req.QuotaList {
				if val, ok := quota.Annotations[cluster.AnnotationSubClusterForTaiji]; ok {
					quotaResources := make(map[string]string)
					for _, k8SResource := range quota.ResourceList {
						quotaResources[k8SResource.ResourceName] = k8SResource.ResourceQuantity
					}
					// 转换为taiji参数 GPUName
					tjAttributes := make(map[string]string)
					for k, v := range quota.Attributes {
						if k == cluster.TaskGpuTypeKey {
							tjAttributes[cluster.TaijiGPUNameKey] = v
							continue
						}
						tjAttributes[k] = v
					}
					parameter := &trd.UpdateQuotaInfoForTaijiRequest{
						Namespace: req.Namespace,
						SubQuotaInfos: []*trd.NamespaceQuotaForTaiji{{
							Name:              quota.Name,
							SubQuotaLabels:    tjAttributes,
							SubQuotaResources: quotaResources,
							Location:          val,
						}},
						Location: val,
						Operator: req.Operator,
					}

					forTaijiList = append(forTaijiList, parameter)
				}
			}
		case cluster.SubClusterForSuanli:
			for _, quota := range req.QuotaList {
				if val, ok := quota.Annotations[cluster.AnnotationKeyInstalledPlatform]; ok {
					if val != cluster.SubClusterForSuanli {
						continue
					}
					quotaResources := make(map[string]string)
					for _, k8SResource := range quota.ResourceList {
						quotaResources[k8SResource.ResourceName] = k8SResource.ResourceQuantity
					}

					slAttributes := make(map[string]string)
					for k, v := range quota.Attributes {
						slAttributes[k] = v
					}

					parameter := &trd.UpdateNamespaceForSuanliRequest{
						Namespace: req.Namespace,
						SubQuotaInfos: []*trd.NamespaceQuotaForSuanli{{
							Name:              quota.Name,
							SubQuotaLabels:    slAttributes,
							SubQuotaResources: quotaResources,
						}},
						Operator: req.Operator,
					}

					forSuanliList = append(forSuanliList, parameter)
				}
			}
		}
	}
	return forTaijiList, forSuanliList, nil
}

// DeleteFederationClusterNamespaceQuota delete federation quota and sub quota
// NOCC:tosa/fn_length(设计如此)
func (f *FederationManager) DeleteFederationClusterNamespaceQuota(ctx context.Context,
	req *federationmgr.DeleteFederationClusterNamespaceQuotaRequest,
	resp *federationmgr.DeleteFederationClusterNamespaceQuotaResponse) error {

	blog.Infof("Received BcsFederationManager.DeleteFederationClusterNamespaceQuota request, req: %v", req)

	// validate
	if err := req.Validate(); err != nil {
		return ErrReturn(resp, fmt.Sprintf("validate DeleteFederationClusterNamespaceQuota request failed, err: %s",
			err.Error()))
	}

	// 根据联邦集群代理ID获取 联邦集群ID
	fedCluster, err := store.GetStoreModel().GetFederationCluster(context.Background(), req.ClusterId)
	if err != nil {
		return ErrReturn(resp,
			fmt.Sprintf("DeleteFederationClusterNamespaceQuota.GetFederationCluster failed, clusterId: %s, "+
				"err: %s", req.ClusterId, err.Error()))
	}

	err = f.clusterCli.DeleteNamespaceQuota(fedCluster.HostClusterID, req.Namespace, req.Name)
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("DeleteFederationClusterNamespaceQuota failed, err: %s", err.Error()))
	}

	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	return nil
}
