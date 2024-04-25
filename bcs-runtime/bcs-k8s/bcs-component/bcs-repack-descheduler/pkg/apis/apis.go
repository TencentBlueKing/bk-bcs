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
 *
 */

// Package apis xx
package apis

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/apis/tkex/v1alpha1"
)

const (
	// BcsGroupVersion defines the group version of bcs
	BcsGroupVersion = "tkex.tencent.com/v1alpha1"

	// ReplicaSetKind defines the kind of replicaset
	ReplicaSetKind = "ReplicaSet"
	// DeploymentKind defines the kind of deployment
	DeploymentKind = "Deployment"
	// StatefulSetKind defines the kind of statefulset
	StatefulSetKind = "StatefulSet"
	// GameDeploymentKind defines the kind of gamedeployment
	GameDeploymentKind = "GameDeployment"
	// GameStatefulSetKind defines the kind of gamestatefulset
	GameStatefulSetKind = "GameStatefulSet"
	// DaemonSetKind defines the kind of daemonset
	DaemonSetKind = "DaemonSet"

	// EvictionKind the kind of eviction
	EvictionKind = "Eviction"
	// EvictionSubresource the subresource of eviction
	EvictionSubresource = "pods/eviction"

	// PDBGroupBetaVersion v1beta1 version of pdb
	PDBGroupBetaVersion = "policy/v1beta1"
	// PDBGroupV1Version v1 version of pdb
	PDBGroupV1Version = "policy/v1"

	// ElectionID is the leader election id
	ElectionID = "deschedule.tkex.tencent.com"

	// GameDeploymentName the resource name of gamedeployment
	GameDeploymentName = "gamedeployments"
	// GameStatefulSetName the resource name of gamestatefulset
	GameStatefulSetName = "gamestatefulsets"

	sep = "/"

	// NodeNameLabel defines the nodeName label when send to calculator
	NodeNameLabel = "tkex.tencent.com/deschedule-node-name"
	// PodNameLabel defines the podName label when send to calculator
	PodNameLabel = "tkex.tencent.com/deschedule-pod-name"
	// PodNamespaceLabel defines the podNamespace label when send to calculator
	PodNamespaceLabel = "tkex.tencent.com/deschedule-pod-namespace"
	// WorkloadName defines the workloadName label when send to calculator
	WorkloadName = "tkex.tencent.com/workload-name"

	// NodeMasterLabel defines the master label
	NodeMasterLabel = "node-role.kubernetes.io/master"

	DefaultPolicyName       = "bcs-descheduler-policy"
	DefaultClusterNamespace = "bcs-system"
)

var (
	// SupportKind defines the support kind maps
	SupportKind = map[string]string{
		"deployment":      "deployment",
		"statefulset":     "statefulset",
		"gamedeployment":  "gamedeployment",
		"gamestatefulset": "gamestatefulset",
	}
	// NotAllowMigrateNamespace defines the namespace cannot be migrate
	NotAllowMigrateNamespace = map[string]string{
		"bcs-system":  "bcs-system",
		"kube-system": "kube-system",
		"bk-system":   "bk-system",
	}
)

var (
	// ControllerRetryTimeout is retry timeout for failed reconciling
	ControllerRetryTimeout = time.Duration(5) * time.Second
	// ControllerReSyncPeriod is the reSync period for controller
	ControllerReSyncPeriod = time.Duration(10) * time.Minute
	// DefaultQueryTimeout used to set the timeout that query k8s object
	DefaultQueryTimeout = time.Duration(10) * time.Second
	// InformerReSyncPeriod defines the pod informer reSync period
	InformerReSyncPeriod = time.Duration(30) * time.Minute
	// WaitInformerSyncTimeout wait CacheManager's all informers synced timeout
	WaitInformerSyncTimeout = time.Duration(300) * time.Second
)

// NamespacedName return the namespace link name
func NamespacedName(name types.NamespacedName) string {
	return name.Namespace + sep + name.Name
}

// NamespacedNameString return the namespace link name
func NamespacedNameString(namespace, name string) string {
	return namespace + sep + name
}

// NamespacedNamePolicy return the policy's namespace link name
func NamespacedNamePolicy(policy *v1alpha1.DeschedulePolicy) string {
	return policy.Namespace + sep + policy.Name
}

// PodName return the pod namespace link name
func PodName(namespace, podName string) string {
	return namespace + sep + podName
}

// PodNameSplit split the link name of pod to namespace and name
func PodNameSplit(typedName string) (namespace, podName string, err error) {
	s := strings.Split(typedName, sep)
	if len(s) != 2 {
		return namespace, podName, errors.Errorf("PodNameSplit '%s' length not 2 after aplit", typedName)
	}
	return s[0], s[1], nil
}

// NamespacedWorkloadKind link namespace + name + kind for workload
func NamespacedWorkloadKind(namespace, name, kind string) string {
	return strings.ToLower(namespace) + sep + strings.ToLower(name) + sep + strings.ToLower(kind)
}

// NamespacedWorkloadKindSplit split the workload link name to namespace/name/kind
func NamespacedWorkloadKindSplit(str string) (namespace string, name string, kind string, err error) {
	arr := strings.Split(str, sep)
	if len(arr) != 3 {
		return namespace, name, kind,
			errors.Errorf("NamespacedWorkloadKindSplit '%s' length not 3 after split", str)
	}
	return arr[0], arr[1], arr[2], nil
}

const (
	pdbErr = "disruption budget"
)

// IsPDBError check err is "Cannot evict pod as it would violate the pod's disruption budget."
func IsPDBError(err error) bool {
	return strings.Contains(err.Error(), pdbErr)
}
