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

// Package util 提供 Istio 的实用函数
package util

import (
	"strings"

	"istio.io/istio/pkg/config/resource"
	"istio.io/istio/pkg/kube/inject"
)

// IsSystemNamespace returns true for system namespaces
func IsSystemNamespace(ns resource.Namespace) bool {
	return inject.IgnoredNamespaces.Contains(ns.String())
}

// IsIstioControlPlane returns true for resources that are part of the Istio control plane
func IsIstioControlPlane(r *resource.Instance) bool {
	if _, ok := r.Metadata.Labels["istio"]; ok {
		return true
	}
	if r.Metadata.Labels["release"] == "istio" {
		return true
	}
	return false
}

// IsMatched check if the term can be matched in a slice of string
func IsMatched(slice []string, term string) bool {
	for _, val := range slice {
		matched := strings.Contains(term, val)
		return matched
	}
	return false
}

// GetInjectorConfigMapName 获取注入配置映射名称
func GetInjectorConfigMapName(revision string) string {
	name := InjectionConfigMap
	if revision == "" || revision == "default" {
		return name
	}
	return name + "-" + revision
}
