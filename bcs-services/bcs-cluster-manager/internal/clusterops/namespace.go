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

package clusterops

import (
	"context"
	"encoding/json"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// NamespaceInfo detailed info
type NamespaceInfo struct {
	Name        string
	Labels      map[string]string
	Annotations map[string]string
}

// CreateNamespace create namespace
func (ko *K8SOperator) CreateNamespace(ctx context.Context, clusterID string, info NamespaceInfo) error {
	if ko == nil {
		return ErrServerNotInit
	}
	clientInterface, err := ko.GetClusterClient(clusterID)
	if err != nil {
		blog.Errorf("CreateNamespace[%s] GetClusterClient failed: %v", clusterID, err)
		return err
	}

	_, err = clientInterface.CoreV1().Namespaces().Get(ctx, info.Name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		blog.Errorf("CreateNamespace[%s:%s] getNamespace failed: %v", clusterID, info.Name, err)
		return err
	}

	if err == nil {
		blog.Errorf("CreateNamespace[%s:%s] getNamespace success", clusterID, info.Name)
		return nil
	}

	newNs := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: info.Name,
		},
	}
	if len(info.Labels) > 0 {
		newNs.Labels = info.Labels
	}
	if len(info.Annotations) > 0 {
		newNs.Annotations = info.Annotations
	}

	_, err = clientInterface.CoreV1().Namespaces().Create(ctx, newNs, metav1.CreateOptions{})
	if err != nil {
		blog.Errorf("CreateNamespace[%s:%s] createNamespace failed: %v", clusterID, info.Name, err)
		return err
	}
	blog.Infof("CreateNamespace[%s:%s] createNamespace success", clusterID, info.Name)

	return nil
}

// DeleteNamespace delete namespace
func (ko *K8SOperator) DeleteNamespace(ctx context.Context, clusterID, name string) error {
	if ko == nil {
		return ErrServerNotInit
	}
	clientInterface, err := ko.GetClusterClient(clusterID)
	if err != nil {
		blog.Errorf("DeleteNamespace[%s] GetClusterClient failed: %v", clusterID, err)
		return err
	}

	_, err = clientInterface.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		blog.Errorf("DeleteNamespace[%s:%s] getNamespace failed: %v", clusterID, name, err)
		return err
	}

	if errors.IsNotFound(err) {
		blog.Infof("DeleteNamespace[%s:%s] notfound", clusterID, name)
		return nil
	}

	err = clientInterface.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		blog.Errorf("DeleteNamespace[%s:%s] failed: %v", clusterID, name, err)
		return err
	}
	blog.Infof("DeleteNamespace[%s:%s] success", clusterID, name)

	return nil
}

// ResourceQuotaInfo resource quota info
type ResourceQuotaInfo struct {
	Name          string
	Namespace     string
	CpuRequests   string
	CpuLimits     string
	MemRequests   string
	MemLimits     string
	ServiceLimits string
}

// CreateResourceQuota create namespace resource quota
func (ko *K8SOperator) CreateResourceQuota(ctx context.Context, clusterID string, info ResourceQuotaInfo) error {
	if ko == nil {
		return ErrServerNotInit
	}
	clientInterface, err := ko.GetClusterClient(clusterID)
	if err != nil {
		blog.Errorf("CreateResourceQuota[%s] GetClusterClient failed: %v", clusterID, err)
		return err
	}

	if info.Namespace == "" {
		info.Namespace = info.Name
	}

	_, err = clientInterface.CoreV1().ResourceQuotas(info.Name).Get(ctx, info.Name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		blog.Errorf("CreateResourceQuota[%s:%s] getResourceQuota failed: %v", clusterID, info.Name, err)
		return err
	}

	if err == nil {
		blog.Errorf("CreateResourceQuota[%s:%s] getResourceQuota success", clusterID, info.Name)
		return nil
	}

	nsResourceQuota := &apiv1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name: info.Name,
		},
		Spec: apiv1.ResourceQuotaSpec{
			Hard: map[apiv1.ResourceName]resource.Quantity{
				apiv1.ResourceRequestsCPU:    resource.MustParse(info.CpuRequests),
				apiv1.ResourceLimitsCPU:      resource.MustParse(info.CpuLimits),
				apiv1.ResourceRequestsMemory: resource.MustParse(info.MemRequests),
				apiv1.ResourceLimitsMemory:   resource.MustParse(info.MemLimits),
				apiv1.ResourceServices: func() resource.Quantity {
					q, pErr := resource.ParseQuantity(info.ServiceLimits)
					if pErr != nil {
						blog.Errorf("CreateResourceQuota[%s:%s] invalid ServiceLimits format '%s': %v, using default value 0",
							clusterID, info.Name, info.ServiceLimits, pErr)
						return resource.Quantity{}
					}
					return q
				}(),
			},
		},
	}

	_, err = clientInterface.CoreV1().ResourceQuotas(info.Name).Create(ctx, nsResourceQuota, metav1.CreateOptions{})
	if err != nil {
		blog.Errorf("CreateResourceQuota[%s:%s] createResourceQuota failed: %v", clusterID, info.Name, err)
		return err
	}
	blog.Infof("CreateResourceQuota[%s:%s] createResourceQuota success", clusterID, info.Name)

	return nil
}

// DeleteResourceQuota delete namespace resource quota
func (ko *K8SOperator) DeleteResourceQuota(ctx context.Context, clusterID, namespace, name string) error {
	if ko == nil {
		return ErrServerNotInit
	}
	clientInterface, err := ko.GetClusterClient(clusterID)
	if err != nil {
		blog.Errorf("DeleteResourceQuota[%s] GetClusterClient failed: %v", clusterID, err)
		return err
	}

	_, err = clientInterface.CoreV1().ResourceQuotas(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		blog.Errorf("DeleteResourceQuota[%s:%s] getNamespaceResourceQuota failed: %v", clusterID, name, err)
		return err
	}

	if errors.IsNotFound(err) {
		blog.Infof("DeleteResourceQuota[%s:%s] notfound", clusterID, name)
		return nil
	}

	err = clientInterface.CoreV1().ResourceQuotas(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		blog.Errorf("DeleteResourceQuota[%s:%s] failed: %v", clusterID, name, err)
		return err
	}
	blog.Infof("DeleteResourceQuota[%s:%s] success", clusterID, name)

	return nil
}

// UpdateResourceQuota update namespace resource quota
func (ko *K8SOperator) UpdateResourceQuota(ctx context.Context, clusterID string, info ResourceQuotaInfo) error {
	if ko == nil {
		return ErrServerNotInit
	}
	clientInterface, err := ko.GetClusterClient(clusterID)
	if err != nil {
		blog.Errorf("UpdateResourceQuota[%s] GetClusterClient failed: %v", clusterID, err)
		return err
	}
	if info.Namespace == "" {
		info.Namespace = info.Name
	}

	quota, err := clientInterface.CoreV1().ResourceQuotas(info.Namespace).Get(ctx, info.Name, metav1.GetOptions{})
	if err != nil {
		blog.Errorf("UpdateResourceQuota[%s:%s] getNamespaceResourceQuota failed: %v", clusterID, info.Name, err)
		return err
	}
	oldData, err := json.Marshal(quota)
	if err != nil {
		return err
	}
	quota.Spec.Hard[apiv1.ResourceRequestsCPU] = resource.MustParse(info.CpuRequests)
	quota.Spec.Hard[apiv1.ResourceLimitsCPU] = resource.MustParse(info.CpuLimits)
	quota.Spec.Hard[apiv1.ResourceRequestsMemory] = resource.MustParse(info.MemRequests)
	quota.Spec.Hard[apiv1.ResourceLimitsMemory] = resource.MustParse(info.MemLimits)
	if info.ServiceLimits != "" {
		quota.Spec.Hard[apiv1.ResourceServices] = func() resource.Quantity {
			q, pErr := resource.ParseQuantity(info.ServiceLimits)
			if pErr != nil {
				blog.Errorf("UpdateResourceQuota[%s:%s] invalid ServiceLimits format '%s': %v, using default value 0",
					clusterID, info.Name, info.ServiceLimits, pErr)
				return resource.Quantity{}
			}
			return q
		}()
	}

	newData, err := json.Marshal(quota)
	if err != nil {
		return err
	}

	patchBytes, patchErr := strategicpatch.CreateTwoWayMergePatch(oldData, newData, quota)
	if patchErr == nil {
		patchOptions := metav1.PatchOptions{}
		_, err = clientInterface.CoreV1().ResourceQuotas(info.Namespace).Patch(ctx, info.Name,
			types.StrategicMergePatchType, patchBytes, patchOptions)
	} else {
		updateOptions := metav1.UpdateOptions{}
		_, err = clientInterface.CoreV1().ResourceQuotas(info.Namespace).Update(ctx, quota, updateOptions)
	}
	if err != nil {
		blog.Errorf("UpdateResourceQuota CreateTwoWayMergePatch[%s] failed: %v", info.Name, err)
	}

	blog.Infof("UpdateResourceQuota[%s:%s] success", clusterID, info.Name)

	return nil
}
