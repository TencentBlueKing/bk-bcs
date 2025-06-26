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

package k8s

import (
	"context"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
)

// CheckIstioResourceExists 检查集群中是否存在任意 Istio 关联资源
// 存在则返回true，否则返回false
func CheckIstioResourceExists(ctx context.Context, clusterID string) (bool, error) {
	client, err := GetDynamicClient(clusterID)
	if err != nil {
		return false, fmt.Errorf("get dynamic client failed: %v", err)
	}
	discoveryClient, err := GetDiscoveryClient(clusterID)
	if err != nil {
		return false, fmt.Errorf("get discovery client failed: %v", err)
	}

	apiResourceList, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		return false, fmt.Errorf("get server preferred resources failed: %v", err)
	}
	// 编译 apiResourceList,如果是Istio资源，则查询资源是否存在
	for _, apiList := range apiResourceList {
		groupVersion := strings.Split(apiList.GroupVersion, "/")
		if len(groupVersion) != 2 {
			continue
		}
		group, version := groupVersion[0], groupVersion[1]
		if !IsIstioGroup(group) {
			continue
		}
		// 遍历apiList.APIResources，查询资源是否存在
		for _, res := range apiList.APIResources {
			gvr := schema.GroupVersionResource{
				Group:    group,
				Version:  version,
				Resource: res.Name,
			}
			list, err := client.Resource(gvr).List(ctx, metav1.ListOptions{})
			if err != nil {
				blog.Errorf("check istio resource exists, clusterID: %s, group: %s, version: %s, resource: %s, err: %v",
					clusterID, group, version, res.Name, err)
				return false, fmt.Errorf("list istio resource failed: %v", err)
			}
			if len(list.Items) > 0 {
				return true, nil
			}
		}
	}
	return false, nil
}

// DeployResourceByYAML 通过yaml文件部署kubernetes资源
func DeployResourceByYAML(ctx context.Context, clusterID, resourceYAML, kind, name string) error {
	// 获取dynamic client
	dynamicClient, err := GetDynamicClient(clusterID)
	if err != nil {
		return fmt.Errorf("get dynamic client failed: %v", err)
	}

	// 解析YAML到unstructured对象
	var obj unstructured.Unstructured
	if unmarshalErr := yaml.Unmarshal([]byte(resourceYAML), &obj); unmarshalErr != nil {
		return fmt.Errorf("unmarshal %s yaml failed: %v", kind, unmarshalErr)
	}

	// 根据资源类型设置GVR
	gvr, err := getGVR(kind)
	if err != nil {
		return err
	}

	// 尝试获取已存在的资源
	namespace := obj.GetNamespace()
	if namespace == "" {
		namespace = common.IstioNamespace
	}

	// 尝试创建资源，如果已存在则更新
	_, err = dynamicClient.Resource(gvr).Namespace(namespace).Create(ctx, &obj, metav1.CreateOptions{})
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("create %s failed: %v", kind, err)
		}
		// 资源已存在，忽略
		blog.Warnf("%s %s already exists, skip creation in cluster %s", kind, name, clusterID)
	}
	blog.Infof("%s %s created successfully in cluster %s", kind, name, clusterID)

	return nil
}

// DeleteResource 通过名称删除kubernetes资源
func DeleteResource(ctx context.Context, clusterID, kind, name string) error {
	// 获取dynamic client
	dynamicClient, err := GetDynamicClient(clusterID)
	if err != nil {
		return fmt.Errorf("get dynamic client failed: %v", err)
	}

	// 根据资源类型设置GVR
	gvr, err := getGVR(kind)
	if err != nil {
		return err
	}

	// 删除资源
	err = dynamicClient.Resource(gvr).Namespace(common.IstioNamespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			blog.Infof("%s %s not found in cluster %s, skip deletion", kind, name, clusterID)
			return nil
		}
		return fmt.Errorf("delete %s %s failed: %v", kind, name, err)
	}

	blog.Infof("%s %s deleted successfully from cluster %s", kind, name, clusterID)
	return nil
}

// getGVR 根据资源类型获取GroupVersionResource
func getGVR(kind string) (schema.GroupVersionResource, error) {
	switch kind {
	case "PodMonitor":
		return schema.GroupVersionResource{
			Group:    "monitoring.coreos.com",
			Version:  "v1",
			Resource: "podmonitors",
		}, nil
	case "ServiceMonitor":
		return schema.GroupVersionResource{
			Group:    "monitoring.coreos.com",
			Version:  "v1",
			Resource: "servicemonitors",
		}, nil
	case "Telemetry":
		return schema.GroupVersionResource{
			Group:    "telemetry.istio.io",
			Version:  "v1alpha1",
			Resource: "telemetries",
		}, nil
	default:
		return schema.GroupVersionResource{}, fmt.Errorf("unsupported resource kind: %s", kind)
	}
}

// DeleteIstioCrd 删除istio crd
func DeleteIstioCrd(ctx context.Context, clusterID string) error {
	blog.Infof("deleting Istio CRDs for cluster %s", clusterID)

	// 获取dynamic client
	dynamicClient, err := GetDynamicClient(clusterID)
	if err != nil {
		return fmt.Errorf("get dynamic client failed: %v", err)
	}

	// 定义需要删除的 Istio CRD 列表
	istioCRDs := []string{
		"authorizationpolicies.security.istio.io",
		"destinationrules.networking.istio.io",
		"envoyfilters.networking.istio.io",
		"gateways.networking.istio.io",
		"istiooperators.install.istio.io",
		"peerauthentications.security.istio.io",
		"proxyconfigs.networking.istio.io",
		"requestauthentications.security.istio.io",
		"serviceentries.networking.istio.io",
		"sidecars.networking.istio.io",
		"telemetries.telemetry.istio.io",
		"virtualservices.networking.istio.io",
		"wasmplugins.extensions.istio.io",
		"workloadentries.networking.istio.io",
		"workloadgroups.networking.istio.io",
	}

	// CRD 的 GVR
	crdGVR := schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}

	// 删除每个 CRD
	for _, crdName := range istioCRDs {
		err := dynamicClient.Resource(crdGVR).Delete(ctx, crdName, metav1.DeleteOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				blog.Infof("CRD %s not found in cluster %s, skip deletion", crdName, clusterID)
				continue
			}
			blog.Errorf("delete CRD %s failed in cluster %s, err: %v", crdName, clusterID, err)
			return fmt.Errorf("delete CRD %s failed: %v", crdName, err)
		}
		blog.Infof("CRD %s deleted successfully from cluster %s", crdName, clusterID)
	}

	blog.Infof("Istio CRDs cleanup completed for cluster %s", clusterID)
	return nil
}
