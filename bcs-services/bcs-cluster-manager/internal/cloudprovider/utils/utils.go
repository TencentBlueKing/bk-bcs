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

// Package utils for utils
package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/passcc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/user"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

const (
	// BCSNodeGroupTaintKey xxx
	BCSNodeGroupTaintKey = "bcs-cluster-manager"
	// BCSNodeGroupTaintValue xxx
	BCSNodeGroupTaintValue = "noSchedule"
	// BCSNodeGroupGkeTaintEffect xxx
	BCSNodeGroupGkeTaintEffect = "NO_EXECUTE"
	// BCSNodeGroupAzureTaintEffect xxx
	BCSNodeGroupAzureTaintEffect = "NoExecute"
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
func BuildBcsAgentToken(name string, isUser bool) (string, error) {
	var (
		token string
		err   error
	)

	token, err = user.GetUserManagerClient().GetUserToken(name)
	if err != nil {
		blog.Errorf("BuildBcsAgentToken GetUserToken[%s] failed: %v", name, err)
		return "", err
	}
	blog.Infof("BuildBcsAgentToken GetUserToken[%s] success", name)

	if token == "" {
		switch isUser {
		case true:
			token, err = user.GetUserManagerClient().CreateUserToken(user.CreateTokenReq{
				Username:   name,
				Expiration: -1,
			})
			if err != nil {
				blog.Errorf("BuildBcsAgentToken CreateUserToken[%s] failed: %v", name, err)
				return "", err
			}
			blog.Infof("BuildBcsAgentToken CreateUserToken[%s] success", name)
		case false:
			token, err = user.GetUserManagerClient().CreateClientToken(user.CreateClientTokenReq{
				ClientName: name,
				Expiration: -1,
			})
			if err != nil {
				blog.Errorf("BuildBcsAgentToken CreateClientToken[%s] failed: %v", name, err)
				return "", err
			}
			blog.Infof("BuildBcsAgentToken CreateClientToken[%s] success", name)
		}
	}

	// grant permission
	err = user.GetUserManagerClient().GrantUserPermission([]types.Permission{
		{
			UserName:     name,
			ResourceType: user.ResourceTypeClusterManager,
			Resource:     name,
			Role:         user.PermissionManagerRole,
		},
	})
	if err != nil {
		return "", err
	}

	return token, nil
}

// DeleteBcsAgentToken revoke token&permission when delete cluster
func DeleteBcsAgentToken(name string) error {
	var (
		token string
		err   error
	)

	// user-manager not enable
	if user.GetUserManagerClient() == nil {
		return nil
	}

	token, err = user.GetUserManagerClient().GetUserToken(name)
	if err != nil {
		return err
	}

	if token != "" {
		err = user.GetUserManagerClient().DeleteUserToken(token)
		if err != nil {
			return err
		}
	}

	// grant permission
	err = user.GetUserManagerClient().RevokeUserPermission([]types.Permission{
		{
			UserName:     name,
			ResourceType: user.ResourceTypeClusterManager,
			Resource:     name,
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
	if len(clusterID) > 0 {
		err := cloudprovider.GetStorageModel().DeleteClusterCredential(context.Background(), clusterID)
		if err != nil {
			blog.Errorf("DeleteClusterCredentialInfo[%s] failed: %v", clusterID, err)
			return err
		}
	}

	return nil
}

// GetCloudDefaultRuntimeVersion get cloud default k8sVersion runtimeInfo
func GetCloudDefaultRuntimeVersion(cloud *proto.Cloud, version string) (*proto.RunTimeInfo, error) {
	k8sVersion := version
	if k8sVersion == "" || !utils.StringInSlice(version, cloud.GetClusterManagement().GetAvailableVersion()) {
		return nil, fmt.Errorf("cloud[%s] not support version[%s]", cloud.CloudID, version)
	}

	return GetCloudRuntimeVersions(cloud, version)
}

// GetCloudRuntimeVersions get cloud runtime versions
func GetCloudRuntimeVersions(cloud *proto.Cloud, version string) (*proto.RunTimeInfo, error) {
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"cloudid":  cloud.CloudID,
		"version":  version,
		"moduleid": common.RuntimeFlag,
	})

	cloudModuleFlags, err := cloudprovider.GetStorageModel().ListCloudModuleFlag(context.Background(),
		cond, &storeopt.ListOption{})
	if err != nil {
		blog.Errorf("GetCloudRuntimes[%s:%s:%s] failed: %v", cloud.CloudID, version, common.RuntimeFlag, err)
		return nil, err
	}

	if len(cloudModuleFlags) == 0 {
		blog.Infof("GetCloudRuntimes[%s:%s:%s] runtime empty", cloud.CloudID, version, common.RuntimeFlag)
		return &proto.RunTimeInfo{
			ContainerRuntime: common.DefaultDockerRuntime.Runtime,
			RuntimeVersion:   common.DefaultDockerRuntime.Version,
		}, nil
	}

	var defaultRuntime = &proto.RunTimeInfo{
		ContainerRuntime: cloudModuleFlags[0].FlagName,
		RuntimeVersion:   cloudModuleFlags[0].DefaultValue,
	}
	for i := range cloudModuleFlags {
		if common.IsContainerdRuntime(cloudModuleFlags[i].FlagName) {
			defaultRuntime.ContainerRuntime = cloudModuleFlags[i].FlagName
			defaultRuntime.RuntimeVersion = cloudModuleFlags[i].DefaultValue
		}
	}

	return defaultRuntime, nil
}

// ObjToPrettyJson obj to json
func ObjToPrettyJson(obj interface{}) string {
	marshal, _ := json.MarshalIndent(obj, "", "    ")
	return bytes.NewBuffer(marshal).String()
}

// ObjToJson obj to json
func ObjToJson(obj interface{}) string {
	marshal, _ := json.Marshal(obj)
	return bytes.NewBuffer(marshal).String()
}

// AuthClusterResourceCreatorPerm auth resource cluster relative perms
func AuthClusterResourceCreatorPerm(ctx context.Context, clusterID, clusterName, user string) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	err := auth.IAMClient.AuthResourceCreatorPerm(ctx, iam.ResourceCreator{
		ResourceType: string(iam.SysCluster),
		ResourceID:   clusterID,
		ResourceName: clusterName,
		Creator:      user,
	}, nil)
	if err != nil {
		blog.Errorf("AuthClusterResourceCreatorPerm[%s] resource[%s:%s] failed: %v",
			taskID, clusterID, clusterName, user)
		return
	}

	blog.Infof("AuthClusterResourceCreatorPerm[%s] resource[%s:%s] successful",
		taskID, clusterID, clusterName, user)
}

// GetKubeletParas get kubelet paras from nodeTemplate
func GetKubeletParas(template *proto.NodeTemplate) map[string]string {
	if template == nil || template.ExtraArgs == nil || template.ExtraArgs[common.Kubelet] == "" {
		return common.DefaultNodeConfig
	}

	kubeletValues := strings.Split(template.ExtraArgs[common.Kubelet], ";")
	if len(kubeletValues) == 0 {
		return common.DefaultNodeConfig
	}

	var (
		exist = false
	)
	for i := range kubeletValues {
		kvs := strings.Split(kubeletValues[i], "=")
		if len(kvs) != 2 {
			continue
		}
		if kvs[0] == common.RootDir {
			exist = true
			break
		}
	}
	if exist {
		return template.ExtraArgs
	}

	kubeletValues = append(kubeletValues, fmt.Sprintf("%s=%s", common.RootDir, common.RootDirValue))
	return map[string]string{
		common.Kubelet: strings.Join(kubeletValues, ";"),
	}
}
