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

package utils

const (
	// AttrResourceType is the resource type
	AttrResourceType = "resource_type"
)

// Attr is the namespace attr
type Attr struct {
	ID          string     `json:"id"`
	DisplayName string     `json:"display_name"`
	Values      []Instance `json:"values"`
}

// Instance is the instance
type Instance struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// GetAttrValues returns the attr values
func GetAttrValues() map[string]Attr {
	result := make(map[string]Attr, 0)
	attrs := make([]Instance, 0)
	for k, v := range resourceTypeMap {
		attrs = append(attrs, Instance{ID: k, DisplayName: v})
	}
	result[AttrResourceType] = Attr{DisplayName: "资源类型", Values: attrs}
	return result
}

// 列出常用的资源类型
var resourceTypeMap = map[string]string{
	"deployments":              "Deployment",
	"replicasets":              "ReplicaSet",
	"statefulsets":             "StatefulSet",
	"daemonsets":               "DaemonSet",
	"pods":                     "Pod",
	"gamedeployments":          "GameDeployment",
	"gamestatefulsets":         "GameStatefulSet",
	"hookruns":                 "HookRun",
	"hooktemplates":            "HookTemplate",
	"jobs":                     "Job",
	"cronjobs":                 "CronJob",
	"horizontalpodautoscalers": "HorizontalPodAutoscaler",
	"generalpodautoscalers":    "GeneralPodAutoscaler",
	"bklogconfigs":             "BkLogConfig",
	"events":                   "Event",
	"ingresses":                "Ingress",
	"podmonitors":              "PodMonitor",
	"servicemonitors":          "ServiceMonitor",
	"roles":                    "Role",
	"rolebindings":             "RoleBinding",
	"configmaps":               "ConfigMap",
	"secrets":                  "Secret",
	"persistentvolumeclaims":   "PersistentVolumeClaim",
	"other":                    "其他",
}

// GetResourceAttr returns the resource attr, k8s resource is like secrets, configmaps, pods
func GetResourceAttr(resource string) map[string]interface{} {
	_, ok := resourceTypeMap[resource]
	if !ok {
		// 没有列出的资源类型，统一归类为other
		return map[string]interface{}{AttrResourceType: "other"}
	}
	return map[string]interface{}{AttrResourceType: resource}
}
