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

// Package auth xxx
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"go-micro.dev/v4/server"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/tasks"
	clusterauth "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	namespaceauth "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/namespace"
	projectauth "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
	authutils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
)

// ClientPermissions client 类型用户拥有的权限，clientID -> actions
var ClientPermissions = map[string][]string{}

// EnableAuth enable auth
var EnableAuth = true

// SkipHandler skip handler
func SkipHandler(ctx context.Context, req server.Request) bool {
	// if not enable auth, skip
	if !EnableAuth {
		return true
	}
	// skip auth for some method
	for _, v := range NoAuthMethod {
		if v == req.Method() {
			return true
		}
	}
	return false
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

type resourceID struct {
	ProjectId           string `json:"projectId,omitempty"`
	ClusterId           string `json:"clusterId,omitempty"`
	SubClusterProjectId string `json:"subclusterProjectId,omitempty"`
	SubClusterId        string `json:"subclusterId,omitempty"`
	TaskId              string `json:"taskId,omitempty"`
	Namespace           string `json:"namespace,omitempty"`
}

func checkResourceID(ctx context.Context, resourceID *resourceID) error {
	// fed cluster or host cluster
	if resourceID.ClusterId != "" && resourceID.ProjectId == "" {
		cls, err := cluster.GetClusterClient().GetCluster(ctx, resourceID.ClusterId)
		if err != nil {
			return err
		}
		resourceID.ProjectId = cls.ProjectID
	}

	// sub cluster
	if resourceID.SubClusterId != "" && resourceID.SubClusterProjectId == "" {
		cls, err := cluster.GetClusterClient().GetCluster(ctx, resourceID.SubClusterId)
		if err != nil {
			return err
		}
		resourceID.SubClusterProjectId = cls.ProjectID
	}

	// taskid
	if resourceID.TaskId != "" && resourceID.ClusterId == "" {
		err := checkTaskResourceID(ctx, resourceID)
		if err != nil {
			return err
		}
	}

	return nil
}

func checkTaskResourceID(ctx context.Context, resourceID *resourceID) error {
	t, err := task.GetTaskManagerClient().GetTaskWithID(ctx, resourceID.TaskId)
	if err != nil {
		return err
	}

	switch t.TaskType {
	case tasks.InstallFederationTaskName.Type:
		resourceID.ClusterId = t.TaskIndex
		cls, err := cluster.GetClusterClient().GetCluster(ctx, resourceID.ClusterId)
		if err != nil {
			return err
		}
		resourceID.ProjectId = cls.ProjectID
		return nil
	case tasks.RegisterSubclusterTaskName.Type, tasks.RemoveSubclusterTaskName.Type:
		ids := strings.Split(t.TaskIndex, "/")
		if len(ids) != 2 {
			return fmt.Errorf("invalid taskindex for task: %s", resourceID.TaskId)
		}

		resourceID.ClusterId, resourceID.SubClusterId = ids[0], ids[1]
		// fed cluster or host cluster
		cls, err := cluster.GetClusterClient().GetCluster(ctx, resourceID.ClusterId)
		if err != nil {
			return err
		}
		resourceID.ProjectId = cls.ProjectID

		// subcluster
		subCls, err := cluster.GetClusterClient().GetCluster(ctx, resourceID.SubClusterId)
		if err != nil {
			return err
		}
		resourceID.SubClusterProjectId = subCls.ProjectID
		return nil
	case tasks.HandleNamespaceQuotaTaskName.Type:
		ids := strings.Split(t.TaskIndex, "/")
		if len(ids) != 2 {
			return fmt.Errorf("invalid taskindex for task: %s", resourceID.TaskId)
		}
		resourceID.ClusterId, resourceID.Namespace = ids[0], ids[1]

		// fed cluster or host cluster
		cls, err := cluster.GetClusterClient().GetCluster(ctx, resourceID.ClusterId)
		if err != nil {
			return err
		}
		resourceID.ProjectId = cls.ProjectID
		return nil
	default:
		return fmt.Errorf("invalid tasktype for task: %s", t.TaskType)
	}

}

// CheckUserPerm check user perm
func CheckUserPerm(ctx context.Context, req server.Request, username string) (bool, error) {
	if len(username) == 0 {
		return false, errors.New("username is empty")
	}
	body := req.Body()
	b, err := json.Marshal(body)
	if err != nil {
		return false, err
	}

	resourceId := &resourceID{}
	if err = json.Unmarshal(b, resourceId); err != nil {
		return false, err
	}

	action, ok := ActionPermissions[req.Method()]
	if !ok {
		return false, errors.New("operation has not authorized")
	}

	// check resourceID
	if err = checkResourceID(ctx, resourceId); err != nil {
		return false, fmt.Errorf("auth failed: err %s", err.Error())
	}

	allow, url, resources, err := callIAM(username, action, *resourceId)
	if err != nil {
		return false, fmt.Errorf("auth failed, resourceId: %v: err %s", resources, err.Error())
	}

	blog.Infof("CheckUserPerm method[%s], user[%s] allow[%v] url[%s] resources[%+v]", req.Method(), username, allow, url, resources)
	if !allow && url != "" && resources != nil {
		return false, &authutils.PermDeniedError{
			Perms: authutils.PermData{
				ApplyURL:   url,
				ActionList: resources,
			},
		}
	}

	// if sub clusterId is not null, call iam with sub clusterId
	subAllow := true
	if resourceId.SubClusterId != "" {
		iResourceID := &resourceID{
			ClusterId: resourceId.SubClusterId,
			ProjectId: resourceId.SubClusterProjectId,
		}
		iAllow, iUrl, iResources, iErr := callIAM(username, action, *iResourceID)
		if iErr != nil {
			return false, fmt.Errorf("auth failed, resourceId: %v: err %s", iResources, iErr.Error())
		}
		blog.Infof("CheckUserPerm user[%s] allow[%v] url[%s] resources[%+v]", username, iAllow, iUrl, iResources)
		if !subAllow && iUrl != "" && iResources != nil {
			blog.Infof("CheckUserPerm failed user[%s] subAllow[%v] url[%s] resources[%+v]", username, iAllow, iUrl, iResources)
			return false, &authutils.PermDeniedError{
				Perms: authutils.PermData{
					ApplyURL:   iUrl,
					ActionList: iResources,
				},
			}
		}
		subAllow = iAllow
	}

	return allow && subAllow, nil
}

func callIAM(username, action string, resourceID resourceID) (bool, string, []authutils.ResourceAction, error) {
	var isSharedCluster bool
	if resourceID.ClusterId != "" {
		cls, err := cluster.GetClusterClient().GetCluster(context.TODO(), resourceID.ClusterId)
		if err != nil {
			blog.Infof("call iam failed: get cluster %s error: %s", resourceID.ClusterId, err.Error())
			return false, "", nil, err
		}
		// if cluster is shared, projectID should be the projectID of shared cluster instead of the projectID of federation cluster
		isSharedCluster = cls.GetIsShared() && cls.GetProjectID() != resourceID.ProjectId
	}

	switch action {
	case projectauth.CanViewProjectOperation:
		return ProjectIamClient.CanViewProject(username, resourceID.ProjectId)
	case clusterauth.CanViewClusterOperation:
		return ClusterIamClient.CanViewCluster(username, resourceID.ProjectId, resourceID.ClusterId)
	case clusterauth.CanManageClusterOperation:
		return ClusterIamClient.CanManageCluster(username, resourceID.ProjectId, resourceID.ClusterId)
	case namespaceauth.CanViewNamespaceOperation:
		return NamespaceIamClient.CanViewNamespace(username, resourceID.ProjectId, resourceID.ClusterId, resourceID.Namespace, isSharedCluster)
	case namespaceauth.CanListNamespaceOperation:
		return NamespaceIamClient.CanListNamespace(username, resourceID.ProjectId, resourceID.ClusterId, isSharedCluster)
	default:
		return false, "", nil, fmt.Errorf("action %s not support", action)
	}
}
