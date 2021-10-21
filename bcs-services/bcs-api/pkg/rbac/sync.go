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

package rbac

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	rbacUtils "github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/rbac/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/storages/sqlstore"
	"k8s.io/client-go/kubernetes"
)

const (
	resourceTypeCluster   = "cluster"
	resourceTypeNamespace = "namespace"
	allCluster            = "*"
	allNamespace          = "*"
	policyFromPattern     = "pattern_policy"
	policyFromCommon      = "common"
	//operationAdd          = "add"
	//operationDelete       = "delete"
)

// syncAuthRbacData sync rbac data to clusters
func syncAuthRbacData(rbacData *AuthRbacData) error {

	if rbacData.ResourceType == resourceTypeCluster {
		return extractClusterLevelData(rbacData)
	} else if rbacData.ResourceType == resourceTypeNamespace {
		return extractNamespaceLevelData(rbacData)
	}
	return fmt.Errorf("invalid resource type: %s", rbacData.ResourceType)
}

// extractClusterLevelData extract cluster level rbac data, and sync cluster level rbac data to clusters
func extractClusterLevelData(rbacData *AuthRbacData) error {
	username := fmt.Sprintf("%s.%s", rbacData.Principal.PrincipalType, rbacData.Principal.PrincipalId)
	clusterIdFromAuth := rbacData.ResourceInstance.Cluster

	// 如果是任意集群的权限，则对 bcs-apiserver 上托管的每个k8s集群都同步数据
	if rbacData.ResourceInstance.Cluster == allCluster {
		clusters := sqlstore.GetAllCluster()
		if len(clusters) == 0 {
			return nil
		}

		for _, cluster := range clusters {
			err := syncClusterLevelData(username, rbacData.Action, rbacData.Operation, cluster.ID, clusterRoleBindingTypeFromAny)
			if err != nil {
				return err
			}
		}
		return nil
	}

	cluster := sqlstore.GetClusterByFuzzyClusterId(clusterIdFromAuth)
	if cluster == nil {
		return fmt.Errorf("cluster not exist in bcs-apiserver, cluster: %s", clusterIdFromAuth)
	}

	// 如果该权限是用户直接申请特定集群的权限，则对之后的 clusterrolebinding 打上 clusterRoleBindingTypeFromCommon 的 label
	if rbacData.PolicyFrom == policyFromCommon {
		return syncClusterLevelData(username, rbacData.Action, rbacData.Operation, cluster.ID, clusterRoleBindingTypeFromCommon)
	} else if rbacData.PolicyFrom == policyFromPattern {
		// 如果该权限是从用户申请任意集群的权限而来，则带上 clusterRoleBindingTypeFromAny 的 label
		return syncClusterLevelData(username, rbacData.Action, rbacData.Operation, cluster.ID, clusterRoleBindingTypeFromAny)
	}
	return fmt.Errorf("invalid policyfrom: %s", rbacData.PolicyFrom)
}

// syncClusterLevelData sync cluster level rbac data to the cluster
func syncClusterLevelData(username, action, operation, clusterId, bindingType string) error {
	kubeClient, err := rbacUtils.GetKubeClient(clusterId)
	if err != nil {
		return fmt.Errorf("failed to build kubeclient for cluster %s: %s", clusterId, err.Error())
	}

	if operation == "add" {
		return addClusterLevelRbac(username, action, clusterId, bindingType, kubeClient)
	} else if operation == "delete" {
		return deleteClusterLevelRbac(username, action, clusterId, bindingType, kubeClient)
	}
	return fmt.Errorf("invalid operabion: %s", operation)
}

// addClusterLevelRbac add rbac to cluster
func addClusterLevelRbac(username, action, clusterId, bindingType string, kubeClient *kubernetes.Clientset) error {
	rm := newRbacManager(clusterId, kubeClient)
	if err := rm.ensureRole(action); err != nil {
		return err
	}

	if err := rm.ensureAddClusterRoleBinding(username, action, bindingType); err != nil {
		return err
	}
	return nil
}

// deleteClusterLevelRbac delete rbac data from cluster
func deleteClusterLevelRbac(username, action, clusterId, bindingType string, kubeClient *kubernetes.Clientset) error {
	rm := newRbacManager(clusterId, kubeClient)
	return rm.ensureDeleteClusterRoleBinding(username, action, bindingType)
}

// extractNamespaceLevelData extract namespace level rbac data, and sync namespace level rbac data to clusters
func extractNamespaceLevelData(rbacData *AuthRbacData) error {
	username := fmt.Sprintf("%s.%s", rbacData.Principal.PrincipalType, rbacData.Principal.PrincipalId)
	clusterIdFromAuth := rbacData.ResourceInstance.Cluster

	cluster := sqlstore.GetClusterByFuzzyClusterId(clusterIdFromAuth)
	if cluster == nil {
		return fmt.Errorf("cluster not exist in bcs-apiserver, cluster: %s", clusterIdFromAuth)
	}

	// 如果是任意 namespace 的权限，则创建 clusterrolebinding
	if rbacData.ResourceInstance.Namespace == allNamespace {
		return syncClusterLevelData(username, rbacData.Action, rbacData.Operation, cluster.ID, clusterRoleBindingTypeFromNamespace)
	}

	// 如果 type 是 pattern_policy ， 说明已经在这个集群中为任意 namespace 的权限创建 clusterrolebinding， 跳过
	if rbacData.PolicyFrom == policyFromPattern {
		blog.Infof("sync namespace level rbac from pattern_policy, skipping. cluster: %s, namespace: %s, user: %s, action: %s", clusterIdFromAuth, rbacData.ResourceInstance.Namespace, username, rbacData.Action)
		return nil
	} else if rbacData.PolicyFrom == policyFromCommon {
		return syncNamespaceLevelData(username, rbacData.Action, rbacData.Operation, cluster.ID, rbacData.ResourceInstance.Namespace)
	}

	return fmt.Errorf("invalid policyfrom: %s", rbacData.PolicyFrom)
}

// syncNamespaceLevelData sync namespace leve rbac data to cluster
func syncNamespaceLevelData(username, action, operation, clusterId, namespace string) error {
	kubeClient, err := rbacUtils.GetKubeClient(clusterId)
	if err != nil {
		return fmt.Errorf("failed to build kubeclient for cluster %s: %s", clusterId, err.Error())
	}

	if operation == "add" {
		return addNamespaceLevelRbac(username, action, clusterId, namespace, kubeClient)
	} else if operation == "delete" {
		return deleteNamespaceLevelRbac(username, action, clusterId, namespace, kubeClient)
	}
	return fmt.Errorf("invalid operabion: %s", operation)
}

// addNamespaceLevelRbac add namespace level rbac data to cluster namespace
func addNamespaceLevelRbac(username, action, clusterId, namespace string, kubeClient *kubernetes.Clientset) error {
	rm := newRbacManager(clusterId, kubeClient)
	if err := rm.ensureRole(action); err != nil {
		return err
	}

	if err := rm.ensureAddRoleBinding(username, action, namespace); err != nil {
		return err
	}
	return nil
}

// deleteNamespaceLevelRbac delete namespace level rbac data to cluster namespace
func deleteNamespaceLevelRbac(username, action, clusterId, namespace string, kubeClient *kubernetes.Clientset) error {
	rm := newRbacManager(clusterId, kubeClient)
	return rm.ensureDeleteRoleBinding(username, action, namespace)
}
