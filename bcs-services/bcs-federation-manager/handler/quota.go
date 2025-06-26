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

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/store"
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

	if fedClusterNamespace == nil {
		blog.Errorf("federation namespace not found, clusterId: %s, namespace: %s", req.ClusterId, req.Namespace)
		return fmt.Errorf("federation namespace not found, "+
			"clusterId: %s, namespace: %s", req.ClusterId, req.Namespace)
	}

	// create quota
	err = f.clusterCli.CreateNamespaceQuota(fedCluster.HostClusterID, req)
	if err != nil {
		blog.Errorf("CreateNamespaceQuota failed req: %+v, err: %s", req, err.Error())
		return ErrReturn(resp, fmt.Sprintf("CreateNamespaceQuota failed, err: %s", err.Error()))
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
	if fedClusterNamespace == nil {
		blog.Errorf("federation namespace not found, clusterId: %s, namespace: %s", req.ClusterId, req.Namespace)
		return fmt.Errorf("federation namespace not found, "+
			"clusterId: %s, namespace: %s", req.ClusterId, req.Namespace)
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

	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	return nil
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

	// 查询是否已存在联邦集群命名空间
	fedClusterNamespace, err := f.clusterCli.GetNamespace(fedCluster.HostClusterID, req.Namespace)
	if err != nil {
		return fmt.Errorf("get federation namespace failed, "+
			"clusterId: %s, namespace: %s, err: %s", req.ClusterId, req.Namespace, err.Error())
	}
	if fedClusterNamespace == nil {
		blog.Errorf("federation namespace not found, clusterId: %s, namespace: %s", req.ClusterId, req.Namespace)
		return fmt.Errorf("federation namespace not found, "+
			"clusterId: %s, namespace: %s", req.ClusterId, req.Namespace)
	}

	err = f.clusterCli.DeleteNamespaceQuota(fedCluster.HostClusterID, req.Namespace, req.Name)
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("DeleteFederationClusterNamespaceQuota failed, err: %s", err.Error()))
	}

	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	return nil
}
