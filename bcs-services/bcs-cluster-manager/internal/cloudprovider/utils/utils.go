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

package utils

import (
	"context"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/passcc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/user"
)

// SyncClusterInfoToPassCC sync clusterInfo to pass-cc
func SyncClusterInfoToPassCC(taskID string, cluster *proto.Cluster) {
	err := passcc.GetCCClient().CreatePassCCCluster(cluster)
	if err != nil {
		blog.Errorf("UpdateCreateClusterDBInfoTask[%s] syncClusterInfoToPassCC CreatePassCCCluster[%s] failed: %v",
			taskID, cluster.ClusterID, err)
	} else {
		blog.Infof("UpdateCreateClusterDBInfoTask[%s] syncClusterInfoToPassCC CreatePassCCCluster[%s] successful",
			taskID, cluster.ClusterID)
	}

	err = passcc.GetCCClient().CreatePassCCClusterSnapshoot(cluster)
	if err != nil {
		blog.Errorf("UpdateCreateClusterDBInfoTask[%s] syncClusterInfoToPassCC CreatePassCCClusterSnapshoot[%s] failed: %v",
			taskID, cluster.ClusterID, err)
	} else {
		blog.Infof("UpdateCreateClusterDBInfoTask[%s] syncClusterInfoToPassCC CreatePassCCClusterSnapshoot[%s] successful",
			taskID, cluster.ClusterID)
	}
}

// SyncDeletePassCCCluster sync delete pass-cc cluster
func SyncDeletePassCCCluster(taskID string, cluster *proto.Cluster) {
	err := passcc.GetCCClient().DeletePassCCCluster(cluster.ProjectID, cluster.ClusterID)
	if err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: DeletePassCCCluster[%s] failed: %v", taskID, cluster.ClusterID, err)
	} else {
		blog.Infof("CleanClusterDBInfoTask[%s]: DeletePassCCCluster[%s] successful", taskID, cluster.ClusterID)
	}
}

// BuildBcsAgentToken create cluster
func BuildBcsAgentToken(cluster *proto.Cluster) (string, error) {
	var (
		token string
		err   error
	)

	token, err = user.GetUserManagerClient().GetUserToken(cluster.ClusterID)
	if err != nil {
		return "", err
	}

	if token == "" {
		token, err = user.GetUserManagerClient().CreateUserToken(user.CreateTokenReq{
			Username:   cluster.ClusterID,
			Expiration: -1,
		})
		if err != nil {
			return "", err
		}
	}

	// grant permission
	err = user.GetUserManagerClient().GrantUserPermission([]types.Permission{
		types.Permission{
			UserName:     cluster.ClusterID,
			ResourceType: user.ResourceTypeClusterManager,
			Resource:     cluster.ClusterID,
			Role:         user.PermissionManagerRole,
		},
	})
	if err != nil {
		return "", err
	}

	return token, nil
}

// DeleteBcsAgentToken revoke token&permission when delete cluster
func DeleteBcsAgentToken(cluster *proto.Cluster) error {
	var (
		token string
		err   error
	)

	// user-manager not enable
	if user.GetUserManagerClient() == nil {
		return nil
	}

	token, err = user.GetUserManagerClient().GetUserToken(cluster.ClusterID)
	if err != nil {
		return  err
	}

	if token != "" {
		err = user.GetUserManagerClient().DeleteUserToken(token)
		if err != nil {
			return err
		}
	}

	// grant permission
	err = user.GetUserManagerClient().RevokeUserPermission([]types.Permission{
		types.Permission{
			UserName:     cluster.ClusterID,
			ResourceType: user.ResourceTypeClusterManager,
			Resource:     cluster.ClusterID,
			Role:         user.PermissionManagerRole,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

// DeleteClusterCredentialInfo delete cluster credential info
func DeleteClusterCredentialInfo(clusterID string) error {
	err := cloudprovider.GetStorageModel().DeleteClusterCredential(context.Background(), clusterID)
	if err != nil{
		blog.Errorf("DeleteClusterCredentialInfo[%s] failed: %v", clusterID, err)
		return err
	}

	return nil
}

func getResourceType(env string) string {
	if env == "prod" {
		return auth.ClusterProd
	}

	return auth.ClusterTest
}
