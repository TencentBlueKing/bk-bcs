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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cloudaccount"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
	authutils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	"go-micro.dev/v4/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

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
	AccountID   string `json:"accountID,omitempty"`  // 云账号
	BusinessID  string `json:"businessID,omitempty"` // 业务Id
	ScopeType   string `json:"scopeType,omitempty"`  // 资源范围类型(biz/biz_set)
	ScopeId     string `json:"scopeId,omitempty"`    // 资源范围ID(业务ID/业务集ID)
}

func checkResourceID(resourceID *resourceID) error {
	if resourceID.ServerKey != "" && resourceID.ClusterID == "" {
		resourceID.ClusterID = resourceID.ServerKey
	}
	if resourceID.InnerIP != "" && resourceID.ClusterID == "" {
		node, err := store.GetStoreModel().GetNodeByIP(context.TODO(), resourceID.InnerIP)
		if err != nil {
			return err
		}
		resourceID.ClusterID = node.ClusterID
	}
	if resourceID.ClusterID != "" && resourceID.ProjectID == "" {
		cluster, err := store.GetStoreModel().GetCluster(context.TODO(), resourceID.ClusterID)
		if err != nil {
			return err
		}
		resourceID.ProjectID = cluster.ProjectID
	}
	if resourceID.NodeGroupID != "" && resourceID.ClusterID == "" {
		np, err := store.GetStoreModel().GetNodeGroup(context.TODO(), resourceID.NodeGroupID)
		if err != nil {
			return err
		}
		resourceID.ClusterID = np.ClusterID
		resourceID.ProjectID = np.ProjectID
	}
	if resourceID.TaskID != "" && resourceID.ClusterID == "" {
		task, err := store.GetStoreModel().GetTask(context.TODO(), resourceID.TaskID)
		if err != nil {
			return err
		}
		resourceID.ClusterID = task.ClusterID
		resourceID.ProjectID = task.ProjectID
	}
	if resourceID.CloudID != "" && resourceID.AccountID != "" && resourceID.ProjectID == "" {
		cloud, err := store.GetStoreModel().GetCloudAccount(context.TODO(),
			resourceID.CloudID, resourceID.AccountID, false)
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
	if err = json.Unmarshal(b, resourceID); err != nil {
		return false, err
	}

	action, ok := ActionPermissions[req.Method()]
	if !ok {
		return false, errors.New("operation has not authorized")
	}

	// check resourceID
	if err = checkResourceID(resourceID); err != nil {
		return false, fmt.Errorf("auth failed: err %s", err.Error())
	}

	allow, url, resources, err := callIAM(username, action, *resourceID)
	if err != nil {
		return false, err
	}

	blog.Infof("CheckUserPerm user[%s] allow[%v] url[%s] resources[%+v]", username, allow, url, resources)
	if !allow && url != "" && resources != nil {
		return false, &authutils.PermDeniedError{
			Perms: authutils.PermData{
				ApplyURL:   url,
				ActionList: resources,
			},
		}
	}

	return allow, nil
}

func callIAM(username, action string, resourceID resourceID) (bool, string, []authutils.ResourceAction, error) {
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
		allow, url, err := CloudAccountIamClient.CanManageCloudAccount(username, resourceID.ProjectID, resourceID.AccountID)
		return allow, url, nil, err
	case cloudaccount.CanCreateCloudAccountOperation:
		allow, url, err := CloudAccountIamClient.CanCreateCloudAccount(username, resourceID.ProjectID)
		return allow, url, nil, err
	case cloudaccount.CanUseCloudAccountOperation:
		// 存在account id 非必输的情况，跳过权限校验
		if resourceID.AccountID == "" {
			return true, "", nil, nil
		}
		allow, url, err := CloudAccountIamClient.CanUseCloudAccount(username, resourceID.ProjectID, resourceID.AccountID)
		return allow, url, nil, err
	case CanOperatorBiz:
		// 业务下操作人权限校验
		bizID := resourceID.BusinessID
		if bizID == "" && resourceID.ScopeType == string(common.Biz) {
			bizID = resourceID.ScopeId
		}
		// 其他情况则不进行权限校验
		if bizID == "" {
			return true, "", nil, nil
		}
		allow, err := checkUserBizPerm(username, bizID)
		return allow, "", nil, err
	default:
		return false, "", nil, errors.New("permission denied")
	}
}

const (
	// CanOperatorBiz 操作人可操作业务
	CanOperatorBiz = "CanOperatorBiz"
)

// 包被循环调用，只能使用SetCheckBizPerm设置
var checkUserBizPerm func(username string, businessID string) (bool, error)

// SetCheckBizPerm xxx
func SetCheckBizPerm(f func(username string, businessID string) (bool, error)) {
	checkUserBizPerm = f
}
