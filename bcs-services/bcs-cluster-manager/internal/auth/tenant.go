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

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"go-micro.dev/v4/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// TenantClientWhiteList tenant client white list
var TenantClientWhiteList = map[string][]string{}

// SkipMethod skip method tenant validation
func SkipMethod(ctx context.Context, req server.Request) bool {
	for _, v := range NoCheckTenantMethod {
		if v == req.Method() {
			return true
		}
	}
	return false
}

// SkipTenantValidation skip tenant validation
func SkipTenantValidation(ctx context.Context, req server.Request, client string) bool {
	if len(client) == 0 {
		return false
	}
	for _, v := range TenantClientWhiteList[client] {
		if strings.HasPrefix(v, "*") || v == req.Method() {
			return true
		}
	}
	return false
}

// CheckUserResourceTenantAttrFunc is the authorization function for go-micro
func CheckUserResourceTenantAttrFunc(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
		if !options.GetGlobalCMOptions().TenantConfig.EnableMultiTenantMode {
			return fn(ctx, req, rsp)
		}

		var (
			tenantId       = ""
			headerTenantId = GetHeaderTenantIdFromCtx(ctx)
			user           = GetAuthUserInfoFromCtx(ctx)
		)

		// exempt inner user
		if user.IsInner() {
			blog.Infof("CheckUserResourceTenantAttrFunc user[%s] inner client",
				user.GetUsername())
			return fn(ctx, req, rsp)
		}
		// skip method tenant validation
		if SkipMethod(ctx, req) {
			blog.Infof("CheckUserResourceTenantAttrFunc skip method[%s]", req.Method())
			return fn(ctx, req, rsp)
		}
		// exempt client
		if SkipTenantValidation(ctx, req, user.GetUsername()) {
			blog.Infof("CheckUserResourceTenantAttrFunc skip tenant[%s] validate", user.GetUsername())
			return fn(ctx, req, rsp)
		}

		// get tenant id
		if headerTenantId == "" {
			tenantId = user.GetTenantId()
		} else {
			if user.GetTenantId() != headerTenantId {
				tenantId = user.GetTenantId()
			} else {
				tenantId = headerTenantId
			}
		}

		// get resource tenant id
		resourceTenantId, err := GetResourceTenantId(ctx, req)
		if err != nil {
			blog.Errorf("CheckUserResourceTenantAttrFunc GetResourceTenantId failed, err: %s", err.Error())
			return err
		}
		blog.Infof("CheckUserResourceTenantAttrFunc headerTenantId[%s] userTenantId[%s] tenantId[%s] resourceTenantId[%s]",
			headerTenantId, user.GetTenantId(), tenantId, resourceTenantId)

		if tenantId != resourceTenantId {
			return fmt.Errorf("user[%s] tenant[%s] not match resource tenant[%s]",
				user.GetUsername(), tenantId, resourceTenantId)
		}

		return fn(ctx, req, rsp)
	}
}

// GetResourceTenantId get resource tenant id
func GetResourceTenantId(ctx context.Context, req server.Request) (string, error) {
	body := req.Body()
	b, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	blog.Infof("CheckUserResourceTenantAttrFunc mehtod[%s], body: %s", req.Method(), string(b))

	// parse resource id
	resource := &resourceID{}
	if err = json.Unmarshal(b, resource); err != nil {
		return "", err
	}

	// 通过资源获取租户ID
	return getTenantIdByResource(ctx, *resource)
}

// getTenantIdByResource get tenant id by resource
func getTenantIdByResource(ctx context.Context, resource resourceID) (string, error) {
	var (
		projectID = resource.ProjectID
	)

	if projectID == "" && resource.ClusterID != "" {
		cluster, err := store.GetStoreModel().GetCluster(context.TODO(), resource.ClusterID)
		if err != nil {
			return "", err
		}

		projectID = cluster.ProjectID
	}

	if projectID == "" && resource.ServerKey != "" {
		cluster, err := store.GetStoreModel().GetCluster(context.TODO(), resource.ServerKey)
		if err != nil {
			return "", err
		}

		projectID = cluster.ProjectID
	}

	if projectID == "" && resource.NodeGroupID != "" {
		group, err := store.GetStoreModel().GetNodeGroup(context.TODO(), resource.NodeGroupID)
		if err != nil {
			return "", err
		}

		projectID = group.ProjectID
	}

	if projectID == "" && resource.InnerIP != "" {
		node, err := store.GetStoreModel().GetNodeByIP(context.TODO(), resource.InnerIP)
		if err != nil {
			return "", err
		}
		cluster, err := store.GetStoreModel().GetCluster(context.TODO(), node.ClusterID)
		if err != nil {
			return "", err
		}

		projectID = cluster.ProjectID
	}

	if projectID == "" && resource.TaskID != "" {
		task, err := store.GetStoreModel().GetTask(context.TODO(), resource.TaskID)
		if err != nil {
			return "", err
		}
		projectID = task.ProjectID
	}

	if projectID == "" && resource.CloudID != "" && resource.AccountID != "" {
		cloud, err := store.GetStoreModel().GetCloudAccount(context.TODO(),
			resource.CloudID, resource.AccountID, false)
		if err != nil {
			return "", err
		}

		projectID = cloud.ProjectID
	}

	if projectID == "" {
		return "", fmt.Errorf("projectID is empty")
	}

	pro, err := project.GetProjectManagerClient().GetProjectInfo(context.TODO(), projectID, true)
	if err != nil {
		return "", err
	}

	return pro.TenantID, nil
}
