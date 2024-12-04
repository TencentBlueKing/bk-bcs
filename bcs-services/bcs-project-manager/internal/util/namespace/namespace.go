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

// Package namespace xxx
package namespace

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/constant"
)

// FilterNamespaces filter shared namespace
func FilterNamespaces(namespaceList *corev1.NamespaceList, shared bool, projectCode string) []corev1.Namespace {
	nsList := []corev1.Namespace{}
	for _, ns := range namespaceList.Items {
		if shared && ns.Annotations[constant.AnnotationKeyProjectCode] != projectCode {
			continue
		}
		nsList = append(nsList, ns)
	}
	return nsList
}

// FilterTableNamespaces filter shared table namespace
func FilterTableNamespaces(namespaceList *metav1.Table, projectCode string) (*metav1.Table, error) {
	nsList := []metav1.TableRow{}
	for _, ns := range namespaceList.Rows {
		pom := metav1.PartialObjectMetadata{}
		if err := json.Unmarshal(ns.Object.Raw, &pom); err != nil {
			return namespaceList, err
		}
		if pom.Annotations == nil || pom.Annotations[constant.AnnotationKeyProjectCode] != projectCode {
			continue
		}
		nsList = append(nsList, ns)
	}
	namespaceList.Rows = nsList
	return namespaceList, nil
}

// FilterOutVcluster filter out vcluster namespaces
func FilterOutVcluster(namespaces []corev1.Namespace) []corev1.Namespace {
	nsList := []corev1.Namespace{}
	for _, ns := range namespaces {
		// annotation exists means it is a vcluster namespace, do not show it in shared cluster view
		if _, exists := ns.Annotations[constant.AnnotationKeyVcluster]; exists {
			continue
		}
		nsList = append(nsList, ns)
	}
	return nsList
}
