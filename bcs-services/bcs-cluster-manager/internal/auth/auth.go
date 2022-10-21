/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cloudaccount"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
	"github.com/micro/go-micro/v2/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
)

// NoAuthMethod 不需要用户身份认证的方法
var NoAuthMethod = []string{
	"ClusterManager.CheckCloudKubeConfig",
	"ClusterManager.ListCommonCluster",
	"ClusterManager.CheckNodeInCluster",
	"ClusterManager.GetCloud",
	"ClusterManager.ListCloud",
	"ClusterManager.ListCloudVPC",
	"ClusterManager.ListCloudRegions",
	"ClusterManager.GetVPCCidr",
	"ClusterManager.ListCloudAccountToPerm",
	"ClusterManager.GetCloudRegions",
	"ClusterManager.GetCloudRegionZones",
	"ClusterManager.ListCloudRegionCluster",
	"ClusterManager.ListCloudSubnets",
	"ClusterManager.ListCloudSecurityGroups",
	"ClusterManager.ListCloudInstanceTypes",
	"ClusterManager.ListCloudOsImage",
	"ClusterManager.ListOperationLogs",
	"ClusterManager.ListResourceSchema",
	"ClusterManager.QueryPermByActionID",
	"ClusterManager.Health",
}

// ClientPermissions client 类型用户拥有的权限，clientID -> actions
var ClientPermissions = map[string][]string{}

// SkipHandler skip handler
func SkipHandler(ctx context.Context, req server.Request) bool {
	// disable auth
	if !enableAuth() {
		return true
	}
	for _, v := range NoAuthMethod {
		if v == req.Method() {
			return true
		}
	}
	return false
}

func enableAuth() bool {
	op := options.GetGlobalCMOptions()
	return op.Auth.Enable
}

// SkipClient skip client
func SkipClient(ctx context.Context, req server.Request, client string) bool {
	if len(client) == 0 {
		return false
	}
	for _, v := range ClientPermissions[client] {
		if strings.HasPrefix(v, "*") || v == req.Method() {
			return true
		}
	}
	return false
}

// 资源 ID
type resourceID struct {
	ProjectID   string `json:"projectID,omitempty"`
	ClusterID   string `json:"clusterID,omitempty"`
	NodeGroupID string `json:"nodeGroupID,omitempty"`
	TaskID      string `json:"taskID,omitempty"`
	ServerKey   string `json:"serverKey,omitempty"` // same as clusterID
	InnerIP     string `json:"innerIP,omitempty"`   // 节点表示
	CloudID     string `json:"cloudID,omitempty"`
	AccountID   string `json:"accountID,omitempty"` // 云账号
}

func checkResourceID(resourceID *resourceID) error {
	if resourceID.ServerKey != "" && resourceID.ClusterID == "" {
		resourceID.ClusterID = resourceID.ServerKey
	}
	if resourceID.InnerIP != "" && resourceID.ClusterID == "" {
		node, err := cloudprovider.GetStorageModel().GetNodeByIP(context.TODO(), resourceID.InnerIP)
		if err != nil {
			return err
		}
		resourceID.ClusterID = node.ClusterID
	}
	if resourceID.ClusterID != "" && resourceID.ProjectID == "" {
		cluster, err := cloudprovider.GetStorageModel().GetCluster(context.TODO(), resourceID.ClusterID)
		if err != nil {
			return err
		}
		resourceID.ProjectID = cluster.ProjectID
	}
	if resourceID.NodeGroupID != "" && resourceID.ClusterID == "" {
		np, err := cloudprovider.GetStorageModel().GetNodeGroup(context.TODO(), resourceID.NodeGroupID)
		if err != nil {
			return err
		}
		resourceID.ClusterID = np.ClusterID
		resourceID.ProjectID = np.ProjectID
	}
	if resourceID.TaskID != "" && resourceID.ClusterID == "" {
		task, err := cloudprovider.GetStorageModel().GetTask(context.TODO(), resourceID.TaskID)
		if err != nil {
			return err
		}
		resourceID.ClusterID = task.ClusterID
		resourceID.ProjectID = task.ProjectID
	}
	if resourceID.CloudID != "" && resourceID.AccountID != "" && resourceID.ProjectID == "" {
		cloud, err := cloudprovider.GetStorageModel().GetCloudAccount(context.TODO(), resourceID.CloudID, resourceID.AccountID)
		if err != nil {
			return err
		}
		resourceID.ProjectID = cloud.ProjectID
	}
	return nil
}

// CheckUserPerm check user perm
func CheckUserPerm(ctx context.Context, req server.Request, username string) (bool, error) {
	blog.Infof("CheckUserPerm: method/%s, username: %s", req.Method(), username)

	if len(username) == 0 {
		return false, errors.New("username is empty")
	}
	body := req.Body()
	b, err := json.Marshal(body)
	if err != nil {
		return false, err
	}

	resourceID := &resourceID{}
	if err := json.Unmarshal(b, resourceID); err != nil {
		return false, err
	}

	action, ok := ActionPermissions[req.Method()]
	if !ok {
		return false, errors.New("operation has not authorized")
	}

	// check resourceID
	if err := checkResourceID(resourceID); err != nil {
		return false, fmt.Errorf("auth failed: err %s", err.Error())
	}

	allow, _, err := callIAM(username, action, *resourceID)
	if err != nil {
		return false, err
	}
	return allow, nil
}

func callIAM(username, action string, resourceID resourceID) (bool, string, error) {
	// related actions
	switch action {
	case cluster.CanCreateClusterOperation:
		return ClusterIamClient.CanCreateCluster(username, resourceID.ProjectID)
	case cluster.CanManageClusterOperation:
		return ClusterIamClient.CanManageCluster(username, resourceID.ProjectID, resourceID.ClusterID)
	case cluster.CanViewClusterOperation:
		return ClusterIamClient.CanViewCluster(username, resourceID.ProjectID, resourceID.ClusterID)
	case cluster.CanDeleteClusterOperation:
		return ClusterIamClient.CanDeleteCluster(username, resourceID.ProjectID, resourceID.ClusterID)
	case project.CanCreateProjectOperation:
		return ProjectIamClient.CanCreateProject(username)
	case project.CanEditProjectOperation:
		return ProjectIamClient.CanEditProject(username, resourceID.ProjectID)
	case project.CanViewProjectOperation:
		return ProjectIamClient.CanViewProject(username, resourceID.ProjectID)
	case project.CanDeleteProjectOperation:
		return ProjectIamClient.CanDeleteProject(username, resourceID.ProjectID)
	case cloudaccount.CanManageCloudAccountOperation:
		return CloudAccountIamClient.CanManageCloudAccount(username, resourceID.ProjectID, resourceID.AccountID)
	case cloudaccount.CanUseCloudAccountOperation:
		return CloudAccountIamClient.CanUseCloudAccount(username, resourceID.ProjectID, resourceID.AccountID)
	default:
		return false, "", errors.New("permission denied")
	}
}
