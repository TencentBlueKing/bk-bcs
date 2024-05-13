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

package datainformer

import (
	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// Interface interface for datainformer
type Interface interface {
	// AddPodEventHandler add pod event handler to datainformer
	AddPodEventHandler(cache.ResourceEventHandler)
	// AddNamespaceEventHandler add namespace event handler to datainformer
	AddNamespaceEventHandler(cache.ResourceEventHandler)
	// AddNetworkpolicyEventHandler add network policy event handler to datainformer
	AddNetworkpolicyEventHandler(cache.ResourceEventHandler)
	// Run run data informer
	Run() error
	// Stop stop data informer
	Stop()
	// ListAllPods list all pods
	ListAllPods() ([]*corev1.Pod, error)
	// ListPodsByNamespace list pods in certain namespaces
	ListPodsByNamespace(ns string, labelsToMatch labels.Set) ([]*corev1.Pod, error)
	// ListNamespaces list all namespace
	ListNamespaces(labelsToMatch labels.Set) ([]*corev1.Namespace, error)
	// ListAllNetworkPolicy list all network policy
	ListAllNetworkPolicy() ([]*networking.NetworkPolicy, error)
}
