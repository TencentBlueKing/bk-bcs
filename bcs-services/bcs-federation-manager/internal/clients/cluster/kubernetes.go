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

// Package cluster xxx
package cluster

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	federationv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/pkg/kubeapi/federationquota/api/v1"
	federationmgr "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/proto/bcs-federation-manager"
)

// CreateNamespace create namespace in target cluster
func (h *clusterClient) CreateNamespace(clusterID, namespace string) error {
	client, err := h.getKubeClientByClusterId(clusterID)
	if err != nil {
		return err
	}

	// create namespace
	_, err = client.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"name": namespace,
			},
		},
	}, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

// GetLoadbalancerIp get loadbalancer ip
func (h *clusterClient) GetLoadbalancerIp(opt *ResourceGetOptions) (string, error) {
	client, err := h.getKubeClientByClusterId(opt.ClusterId)
	if err != nil {
		return "", err
	}

	svc, err := client.CoreV1().Services(opt.Namespace).Get(context.TODO(), opt.ResourceName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return "", fmt.Errorf("service %s not found in namespace %s", opt.ResourceName, opt.Namespace)
		}
		return "", err
	}

	if len(svc.Status.LoadBalancer.Ingress) == 0 {
		return "", fmt.Errorf("service %s in namespace %s has no external ip", opt.ResourceName, opt.Namespace)
	}

	return svc.Status.LoadBalancer.Ingress[0].IP, nil
}

// CreateSecret create secret, if exists will return k8s AlreadyExists error
func (h *clusterClient) CreateSecret(secret *corev1.Secret, opt *ResourceCreateOptions) error {
	client, err := h.getKubeClientByClusterId(opt.ClusterId)
	if err != nil {
		return err
	}

	_, err = client.CoreV1().Secrets(opt.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	// if IsAlreadyExists return err too, caller should handle error
	return err
}

// ListSecrets list secrets
func (h *clusterClient) ListSecrets(opt *ResourceGetOptions) ([]corev1.Secret, error) {
	client, err := h.getKubeClientByClusterId(opt.ClusterId)
	if err != nil {
		return nil, err
	}

	secretList, err := client.CoreV1().Secrets(opt.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return secretList.Items, nil
}

// GetBootstrapSecret get register token secret
func (h *clusterClient) GetBootstrapSecret(opt *ResourceGetOptions) (*corev1.Secret, error) {
	client, err := h.getKubeClientByClusterId(opt.ClusterId)
	if err != nil {
		return nil, err
	}

	secretList, err := client.CoreV1().Secrets(opt.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	if len(secretList.Items) == 0 {
		return nil, fmt.Errorf("secret not found in namespace %s", opt.Namespace)
	}

	for _, secret := range secretList.Items {
		// check secret prefix
		if strings.HasPrefix(secret.Name, "bootstrap-token-") &&
			strings.Contains(string(secret.Data["auth-extra-groups"]), "register-cluster-token") {
			return &secret, nil
		}
	}

	return nil, fmt.Errorf("boots strap secret not found in cluster/namespace: %s/%s", opt.ClusterId, opt.Namespace)
}

// CreateClusterNamespace create namespace
func (h *clusterClient) CreateClusterNamespace(clusterId, namespace string,
	annotations map[string]string) error {

	kubeClient, err := h.getKubeClientByClusterId(clusterId)
	if err != nil {
		blog.Errorf("CreateClusterNamespace failed clusterId: %s; err: %+v", clusterId, err)
		return err
	}

	ns := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        namespace,
			Annotations: annotations,
		},
	}

	// 创建联邦集群ns
	_, err = kubeClient.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
	if err != nil {
		blog.Errorf("CreateClusterNamespace failed ns: %+v; err: %s", ns, err.Error())
		return err
	}

	return nil
}

// GetNamespace get namespace by clusterId nsName
func (h *clusterClient) GetNamespace(clusterID, nsName string) (*corev1.Namespace, error) {

	client, err := h.getKubeClientByClusterId(clusterID)
	if err != nil {
		return nil, err
	}

	ns, err := client.CoreV1().Namespaces().Get(context.TODO(), nsName, metav1.GetOptions{})
	if err != nil {
		blog.Errorf("GetNamespace getNamespace clusterID: %s; namespace: %s; failed: %+v",
			clusterID, nsName, err)
		return nil, err
	}

	return ns, nil
}

// UpdateNamespace update namespace
func (h *clusterClient) UpdateNamespace(clusterId string, ns *corev1.Namespace) error {

	client, err := h.getKubeClientByClusterId(clusterId)
	if err != nil {
		return err
	}

	// 更新 ns
	_, err = client.CoreV1().Namespaces().Update(context.TODO(), ns, metav1.UpdateOptions{})
	if err != nil {
		blog.Errorf("UpdateNamespace updateNamespace ns: %+v; failed: %+v", ns, err)
		return err
	}

	return nil
}

// DeleteNamespace delete namespace by clusterId nsName
func (h *clusterClient) DeleteNamespace(clusterID, nsName string) error {

	client, err := h.getKubeClientByClusterId(clusterID)
	if err != nil {
		return err
	}

	err = client.CoreV1().Namespaces().Delete(context.TODO(), nsName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		blog.Errorf(
			"DeleteNamespace deleteNamespace clusterID: %s; namespace: %s; failed: %+v",
			clusterID, nsName, err)
		return err
	}

	return nil
}

// ListNamespace get namespaceList by clusterId
func (h *clusterClient) ListNamespace(clusterID string) ([]corev1.Namespace, error) {

	client, err := h.getKubeClientByClusterId(clusterID)
	if err != nil {
		return nil, err
	}

	namespaces, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		blog.Errorf("ListNamespace getNamespaces clusterID: %s; failed: %+v", clusterID, err)
		return nil, err
	}

	if len(namespaces.Items) == 0 {
		return nil, nil
	}

	return namespaces.Items, nil

}

// GetNamespaceQuota get quota by clusterId nsName quotaName
func (h *clusterClient) GetNamespaceQuota(clusterID, nsName, quotaName string) (
	*federationv1.MultiClusterResourceQuota, error) {
	dynamicClient, err := h.getDynamicClientByClusterId(clusterID)
	if err != nil {
		return nil, err
	}

	gvr := schema.GroupVersionResource{
		Group:    "federation.bkbcs.tencent.com",
		Version:  "v1",
		Resource: "multiclusterresourcequotas",
	}

	result, err := dynamicClient.Resource(gvr).Namespace(nsName).Get(context.TODO(), quotaName, metav1.GetOptions{})
	if err != nil {
		blog.Errorf("GetNamespaceQuota get result failed, gvr: %+v, "+
			"nsName: %s, quotaName: %s, err: %s", gvr, nsName, quotaName, err.Error())
		return nil, err
	}

	// 将资源转换为结构体
	mcResourceQuota := &federationv1.MultiClusterResourceQuota{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(result.UnstructuredContent(), mcResourceQuota)
	if err != nil {
		blog.Errorf("GetNamespaceQuota FromUnstructured result failed, result: %+v, "+
			"err: %s", result, err.Error())
		return nil, err
	}

	return mcResourceQuota, nil
}

// ListNamespaceQuota get quotas by clusterId nsName
func (h *clusterClient) ListNamespaceQuota(clusterID, nsName string) (
	*federationv1.MultiClusterResourceQuotaList, error) {

	dynamicClient, err := h.getDynamicClientByClusterId(clusterID)
	if err != nil {
		return nil, err
	}

	gvr := schema.GroupVersionResource{
		Group:    "federation.bkbcs.tencent.com",
		Version:  "v1",
		Resource: "multiclusterresourcequotas",
	}

	results, err := dynamicClient.Resource(gvr).Namespace(nsName).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		blog.Errorf("ListNamespaceQuota get results failed, gvr: %+v, "+
			"nsName: %s, err: %s", gvr, nsName, err.Error())
		return nil, err
	}

	// 将资源转换为结构体
	mcResourceQuotas := &federationv1.MultiClusterResourceQuotaList{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(results.UnstructuredContent(), mcResourceQuotas)
	if err != nil {
		blog.Errorf("ListNamespaceQuota FromUnstructured results failed, result: %+v, "+
			"err: %s", results, err.Error())
		return nil, err
	}

	return mcResourceQuotas, nil
}

// DeleteNamespaceQuota delete quota by clusterId nsName quotaName
func (h *clusterClient) DeleteNamespaceQuota(clusterID, nsName, quotaName string) error {

	dynamicClient, err := h.getDynamicClientByClusterId(clusterID)
	if err != nil {
		return err
	}

	gvr := schema.GroupVersionResource{
		Group:    "federation.bkbcs.tencent.com",
		Version:  "v1",
		Resource: "multiclusterresourcequotas",
	}

	err = dynamicClient.Resource(gvr).Namespace(nsName).Delete(context.TODO(), quotaName, metav1.DeleteOptions{})
	if err != nil {
		blog.Errorf("DeleteNamespaceQuota delete quota failed, gvr: %+v, "+
			"nsName: %s, err: %s", gvr, nsName, err.Error())
		return err
	}

	return nil
}

// GetMultiClusterResourceQuota get multiClusterResourceQuota
func (h *clusterClient) GetMultiClusterResourceQuota(clusterId, namespace, quotaName string) (
	*federationv1.MultiClusterResourceQuota, error) {

	dynamicClient, err := h.getDynamicClientByClusterId(clusterId)
	if err != nil {
		return nil, err
	}

	gvr := schema.GroupVersionResource{
		Group:    "federation.bkbcs.tencent.com",
		Version:  "v1",
		Resource: "multiclusterresourcequotas",
	}

	result, err := dynamicClient.Resource(gvr).Namespace(namespace).Get(context.TODO(),
		quotaName, metav1.GetOptions{})
	if err != nil {
		blog.Errorf("GetMultiClusterResourceQuota get result failed, gvr: %+v, "+
			"nsName: %s, err: %s", gvr, namespace, err.Error())
		return nil, err
	}

	// 将资源转换为结构体
	mcResourceQuota := &federationv1.MultiClusterResourceQuota{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(result.UnstructuredContent(), mcResourceQuota)
	if err != nil {
		blog.Errorf("GetMultiClusterResourceQuota FromUnstructured failed, result: %+v, "+
			"err: %s", result, err.Error())
		return nil, err
	}

	return mcResourceQuota, nil
}

// UpdateNamespaceQuota update quota
func (h *clusterClient) UpdateNamespaceQuota(clusterId, namespace string,
	mcResourceQuota *federationv1.MultiClusterResourceQuota) error {
	dynamicClient, err := h.getDynamicClientByClusterId(clusterId)
	if err != nil {
		return err
	}

	gvr := schema.GroupVersionResource{
		Group:    "federation.bkbcs.tencent.com",
		Version:  "v1",
		Resource: "multiclusterresourcequotas",
	}

	toUnstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(mcResourceQuota)
	if err != nil {
		blog.Errorf("UpdateNamespaceQuota ToUnstructured failed, mcResourceQuota: %+v, "+
			"err: %s", mcResourceQuota, err.Error())
		return err
	}

	obj := &unstructured.Unstructured{Object: toUnstructured}
	_, err = dynamicClient.Resource(gvr).Namespace(namespace).Update(context.TODO(), obj, metav1.UpdateOptions{})
	if err != nil {
		blog.Errorf("UpdateNamespaceQuota update quota failed, gvr: %+v, "+
			"namespace: %s, obj: %+v,err: %s", mcResourceQuota, namespace, obj, err.Error())
		return err
	}

	return nil
}

// CreateMultiClusterResourceQuota 创建 MultiClusterResourceQuota 资源
func (h *clusterClient) CreateMultiClusterResourceQuota(clusterId, namespace string,
	mc *federationv1.MultiClusterResourceQuota) error {
	dynamicClient, err := h.getDynamicClientByClusterId(clusterId)
	if err != nil {
		blog.Errorf("CreateMultiClusterResourceQuota failed clusterId: %s, err: %s", clusterId, err.Error())
		return err
	}
	gvr := schema.GroupVersionResource{
		Group:    "federation.bkbcs.tencent.com",
		Version:  "v1",
		Resource: "multiclusterresourcequotas",
	}

	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(mc)
	if err != nil {
		blog.Errorf("CreateMultiClusterResourceQuota.ToUnstructured failed clusterId: %+v, err: %s", mc, err.Error())
		return errors.NewInternalError(err)
	}

	unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}
	// 创建 MultiClusterResourceQuota 资源
	result, err := dynamicClient.Resource(gvr).Namespace(namespace).Create(context.TODO(),
		unstructuredObj, metav1.CreateOptions{})
	if err != nil {
		blog.Errorf("CreateMultiClusterResourceQuota.Create failed "+
			"namespace: %s, gvr: %+v, unstructuredObj: %+v, err: %s", namespace,
			gvr, unstructuredObj, err.Error())
		return err
	}

	retMultiClusterResourceQuota := &federationv1.MultiClusterResourceQuota{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(result.UnstructuredContent(), retMultiClusterResourceQuota)
	if err != nil {
		blog.Errorf("CreateMultiClusterResourceQuota.FromUnstructured failed "+
			"result: %+v, err: %s", result, err.Error())
		return err
	}

	blog.Infof("retMultiClusterResourceQuota: %+v", retMultiClusterResourceQuota)
	return nil
}

// CreateNamespaceQuota create namespace quota
func (h *clusterClient) CreateNamespaceQuota(federationId string,
	req *federationmgr.CreateFederationClusterNamespaceQuotaRequest) error {

	for _, quota := range req.QuotaList {
		hardList := corev1.ResourceList{}
		for _, k8SResource := range quota.ResourceList {
			resourceName := corev1.ResourceName(k8SResource.ResourceName)
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
				Namespace:         req.Namespace,
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

		// 创建 MultiClusterResourceQuota 资源
		err := h.CreateMultiClusterResourceQuota(federationId, req.Namespace, mcResourceQuota)
		if err != nil {
			return err
		}
	}

	return nil
}
