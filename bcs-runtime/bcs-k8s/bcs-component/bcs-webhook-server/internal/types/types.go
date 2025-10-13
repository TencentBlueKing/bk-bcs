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

// Package types xxx
package types

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// NamespaceBcs bcs system namespace
	NamespaceBcs = "bcs-system"
	// BcsWebhookAnnotationInjectKey inject flag for object in ignored namespaces
	BcsWebhookAnnotationInjectKey = "webhook.inject.bkbcs.tencent.com"
	// PatchOperationAdd patch add operation
	PatchOperationAdd = "add"
	// PatchOperationReplace patch replace operation
	PatchOperationReplace = "replace"
	// PatchOperationRemove patch remove operation
	PatchOperationRemove = "remove"
)

// IgnoredNamespaces namespaces to ignore inject
var IgnoredNamespaces = []string{
	metav1.NamespaceSystem,
	metav1.NamespacePublic,
	NamespaceBcs,
}

// IsIgnoredNamespace see if ns should be ignored
func IsIgnoredNamespace(ns string) bool {
	for _, ins := range IgnoredNamespaces {
		if ns == ins {
			return true
		}
	}
	return false
}

// PatchOperation struct for k8s webhook patch
type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}
