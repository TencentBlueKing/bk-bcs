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

// CheckIstioResourceExists 检查集群中是否存在 Istio 关联资源
// 返回是否存在以及详细的资源信息
func CheckIstioResourceExists(ctx context.Context, clusterID string) (bool, []string, error) {
	client, err := GetDynamicClient(clusterID)
	if err != nil {
		return false, nil, fmt.Errorf("get dynamic client failed: %v", err)
	}
	discoveryClient, err := GetDiscoveryClient(clusterID)
	if err != nil {
		return false, nil, fmt.Errorf("get discovery client failed: %v", err)
	}

	apiResourceList, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		return false, nil, fmt.Errorf("get server preferred resources failed: %v", err)
	}

	var resourceDetails []string

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
				blog.Errorf("list istio resource failed: %s/%s, err: %v", group, res.Name, err)
				return false, nil, fmt.Errorf("list istio resource failed: %s/%s", group, res.Name)
			}
			if len(list.Items) > 0 {
				for _, item := range list.Items {
					// 资源名称
					name := item.GetName()
					// 资源命名空间
					namespace := item.GetNamespace()
					if namespace != "" {
						resourceDetails = append(resourceDetails, fmt.Sprintf("%s/%s (%s/%s)", namespace, name, group, res.Name))
					} else {
						resourceDetails = append(resourceDetails, fmt.Sprintf("%s (%s/%s, cluster-scoped)", name, group, res.Name))
					}
				}
			}
		}
	}

	return len(resourceDetails) > 0, resourceDetails, nil
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
	case "Secret":
		return schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "secrets",
		}, nil
	case "Gateway":
		return schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "gateways",
		}, nil
	case "VirtualService":
		return schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "virtualservices",
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

// DeployTelemetry 部署Telemetry资源用于链路追踪
func DeployTelemetry(ctx context.Context, clusterID []string, randomSamplingPercnt int) error {
	for _, cluster := range clusterID {
		if err := DeployResourceByYAML(
			ctx,
			cluster,
			common.GetTelemetryYAML(randomSamplingPercnt),
			common.TelemetryKind,
			common.TelemetryName,
		); err != nil {
			blog.Errorf("deploy Telemetry failed for cluster %s, err: %v", cluster, err)
			return err
		}
	}
	return nil
}

// DeployServiceMonitor 部署ServiceMonitor资源用于监控
func DeployServiceMonitor(ctx context.Context, clusterID []string) error {
	for _, cluster := range clusterID {
		if err := DeployResourceByYAML(
			ctx,
			cluster,
			common.GetServiceMonitorYAML(common.ServiceMonitorName),
			common.ServiceMonitorKind,
			common.ServiceMonitorName,
		); err != nil {
			blog.Errorf("deploy ServiceMonitor failed for cluster %s, err: %v", cluster, err)
			return err
		}
	}
	return nil
}

// DeployPodMonitor 部署PodMonitor资源用于监控
func DeployPodMonitor(ctx context.Context, clusterID []string) error {
	for _, cluster := range clusterID {
		if err := DeployResourceByYAML(
			ctx,
			cluster,
			common.GetPodMonitorYAML(common.PodMonitorName),
			common.PodMonitorKind,
			common.PodMonitorName,
		); err != nil {
			blog.Errorf("deploy PodMonitor failed for cluster %s, err: %v", cluster, err)
			return err
		}
	}
	return nil
}

// DeleteTelemetry 删除Telemetry资源用于链路追踪
func DeleteTelemetry(ctx context.Context, clusterID []string) error {
	for _, cluster := range clusterID {
		if err := DeleteResource(
			ctx,
			cluster,
			common.TelemetryKind,
			common.TelemetryName,
		); err != nil {
			blog.Errorf("delete Telemetry failed for cluster %s, err: %v", cluster, err)
			return err
		}
	}
	return nil
}

// DeleteServiceMonitor 删除ServiceMonitor资源用于监控
func DeleteServiceMonitor(ctx context.Context, clusterID []string) error {
	for _, cluster := range clusterID {
		if err := DeleteResource(ctx, cluster, common.ServiceMonitorKind, common.ServiceMonitorName); err != nil {
			blog.Errorf("delete ServiceMonitor failed for cluster %s, err: %v", cluster, err)
			return err
		}
	}
	return nil
}

// DeletePodMonitor 删除PodMonitor资源用于监控
func DeletePodMonitor(ctx context.Context, clusterID []string) error {
	for _, cluster := range clusterID {
		if err := DeleteResource(ctx, cluster, common.PodMonitorKind, common.PodMonitorName); err != nil {
			blog.Errorf("delete PodMonitor failed for cluster %s, err: %v", cluster, err)
			return err
		}
	}
	return nil
}

// Gateway 东西向网关
type Gateway struct {
	ClusterID string
	YAML      string
	Kind      string
	Name      string
}

// DeployGateway 部署东西向网关
func DeployGateway(ctx context.Context, gateway *Gateway) error {
	return DeployResourceByYAML(ctx, gateway.ClusterID, gateway.YAML, common.GatewayKind, common.GatewayName)
}
