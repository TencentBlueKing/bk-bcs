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
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/options"
	m "github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/models"
	rbacUtils "github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/rbac/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/storages/sqlstore"
	"k8s.io/client-go/kubernetes"
	"strings"
)

// init rbac data from init conf
func SyncRbacFromConf(rbacDatas []options.RbacData) {

	for _, rbacData := range rbacDatas {
		user := m.User{
			Name: rbacData.Username,
		}

		// Query if user exists
		userInDb := sqlstore.GetUserByCondition(&user)
		if userInDb == nil {
			blog.Warnf("user %s not exists, skip sync this user's rbac...", user.Name)
			continue
		}
		// Query if cluster exists
		cluster := sqlstore.GetCluster(rbacData.ClusterId)
		if cluster == nil {
			blog.Warnf("cluster %s not exists, skip sync this cluster's rbac for user %s ...", rbacData.ClusterId, rbacData.Username)
			continue
		}
		rolesList := rbacData.Roles
		syncConfRbacData(rbacData.Username, rbacData.ClusterId, rolesList)
	}
}

// sync rbac data to backend k8s cluster
func syncConfRbacData(user, clusterId string, rolesList []string) {
	username := strings.Replace(user, ":", ".", 1)
	kubeClient, err := rbacUtils.GetKubeClient(clusterId)
	if err != nil {
		blog.Errorf("failed to build kubeclient for cluster %s: %s", clusterId, err.Error())
		return
	}
	syncToCluster(username, clusterId, rolesList, kubeClient)
}

// sync rbac data to backend k8s cluster with a valid kubeclient
func syncToCluster(username, clusterId string, rolesList []string, kubeClient *kubernetes.Clientset) {
	rm := newRbacManager(clusterId, kubeClient)

	if err := rm.ensureRoles(rolesList); err != nil {
		blog.Errorf(err.Error())
		return
	}

	if err := rm.ensureClusterRoleBindings(username, rolesList); err != nil {
		blog.Errorf(err.Error())
		return
	}
}
