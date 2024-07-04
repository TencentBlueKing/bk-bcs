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

package permitcheck

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	cm "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapiv4/clustermanager"
	iamcluster "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/cluster"
	iamnamespace "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/namespace"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

// CheckBCSClusterPermission check cluster permission
func (c *checker) CheckBCSClusterPermission(ctx context.Context, user, clusterID string,
	action iam.ActionID) (int, error) {
	clusterResp, err := c.option.ClusterManagerClient.GetCluster(
		metadata.NewOutgoingContext(ctx,
			metadata.New(map[string]string{"Authorization": fmt.Sprintf("Bearer %s", c.option.APIGatewayToken)}),
		), &cm.GetClusterReq{ClusterID: clusterID})
	if err != nil {
		return http.StatusInternalServerError, errors.Wrapf(err, "get cluster '%s' failed", clusterID)
	}
	if clusterResp.Code != 0 {
		return http.StatusBadRequest, errors.Errorf("get cluster '%s' code not 0: %s",
			clusterID, clusterResp.Message)
	}
	projectID := clusterResp.Data.ProjectID
	var permit bool
	switch action {
	case iamcluster.ClusterScopedCreate:
		permit, _, _, err = c.clusterPermission.CanCreateClusterScopedResource(user, projectID, clusterID)
	case iamcluster.ClusterScopedView:
		permit, _, _, err = c.clusterPermission.CanViewClusterScopedResource(user, projectID, clusterID)
	case iamcluster.ClusterScopedUpdate:
		permit, _, _, err = c.clusterPermission.CanUpdateClusterScopedResource(user, projectID, clusterID)
	case iamcluster.ClusterScopedDelete:
		permit, _, _, err = c.clusterPermission.CanDeleteClusterScopedResource(user, projectID, clusterID)
	default:
		permit = false
		err = errors.Errorf("unknown iam action '%s'", action)
	}
	if err != nil {
		return http.StatusInternalServerError, errors.Wrapf(err, "auth center failed")
	}
	if !permit {
		return http.StatusForbidden, errors.Errorf("cluster '%s' for user '%s' with %s is forbidden",
			clusterID, user, action)
	}
	return http.StatusOK, nil
}

// CheckBCSPermissions check the permission of bcs
func (c *checker) CheckBCSPermissions(req *http.Request) (bool, int, error) {
	userID := req.URL.Query().Get("user")
	if userID == "" {
		return false, http.StatusBadRequest, fmt.Errorf("query param 'user' cannot be empty")
	}
	iamAction := req.URL.Query().Get("action")
	if iamAction == "" {
		return false, http.StatusBadRequest, fmt.Errorf("query param 'action' cannot be empty")
	}
	var statusCode int
	var permit bool
	var err error
	switch iam.ActionID(iamAction) {
	case iam.ProjectView, iam.ProjectEdit, iam.ProjectDelete:
		statusCode, permit, err = c.checkBCSPermissionProject(req, iam.ActionID(iamAction), userID)
	case iam.ClusterCreate, iam.ClusterView, iam.ClusterManage, iam.ClusterDelete:
		statusCode, permit, err = c.checkBCSPermissionCluster(req, iam.ActionID(iamAction), userID)
	case iamnamespace.NameSpaceCreate, iamnamespace.NameSpaceView, iamnamespace.NameSpaceUpdate,
		iamnamespace.NameSpaceDelete, iamnamespace.NameSpaceList:
		statusCode, permit, err = c.checkBCSPermissionNamespace(req, iam.ActionID(iamAction), userID)
	case iamnamespace.NameSpaceScopedCreate, iamnamespace.NameSpaceScopedView,
		iamnamespace.NameSpaceScopedUpdate, iamnamespace.NameSpaceScopedDelete:
		statusCode, permit, err = c.checkBCSPermissionNamespaceScoped(req, iam.ActionID(iamAction), userID)
	default:
		return false, http.StatusBadRequest, fmt.Errorf("no handler for action '%s'", iamAction)
	}
	return permit, statusCode, err
}

// checkBCSPermissionProject check bcs project permission
func (c *checker) checkBCSPermissionProject(r *http.Request, action iam.ActionID,
	user string) (int, bool, error) {
	projectID := r.URL.Query().Get("projectID")
	if projectID == "" {
		return http.StatusBadRequest, false, fmt.Errorf("query 'projectID' cannot be empty")
	}
	var permit bool
	var err error
	switch action {
	case iam.ProjectView:
		permit, _, _, err = c.projectPermission.CanViewProject(user, projectID)
	case iam.ProjectEdit:
		permit, _, _, err = c.projectPermission.CanEditProject(user, projectID)
	case iam.ProjectDelete:
		permit, _, _, err = c.projectPermission.CanDeleteProject(user, projectID)
	}
	if err != nil {
		return http.StatusInternalServerError, false, errors.Wrapf(err, "check permission failed")
	}
	return http.StatusOK, permit, nil
}

// checkBCSPermissionCluster check bcs permission for cluster
func (c *checker) checkBCSPermissionCluster(r *http.Request, action iam.ActionID,
	user string) (int, bool, error) {
	projectID := r.URL.Query().Get("projectID")
	if projectID == "" {
		return http.StatusBadRequest, false, fmt.Errorf("query 'projectID' cannot be empty")
	}
	var permit bool
	var err error
	switch action {
	case iam.ClusterCreate:
		permit, _, _, err = c.clusterPermission.CanCreateCluster(user, projectID)
	case iam.ClusterView:
		clusterID := r.URL.Query().Get("clusterID")
		if clusterID == "" {
			return http.StatusBadRequest, false, fmt.Errorf("query 'clusterID' cannot be empty")
		}
		permit, _, _, err = c.clusterPermission.CanViewCluster(user, projectID, clusterID)
	case iam.ClusterManage:
		clusterID := r.URL.Query().Get("clusterID")
		if clusterID == "" {
			return http.StatusBadRequest, false, fmt.Errorf("query 'clusterID' cannot be empty")
		}
		permit, _, _, err = c.clusterPermission.CanManageCluster(user, projectID, clusterID)
	case iam.ClusterDelete:
		clusterID := r.URL.Query().Get("clusterID")
		if clusterID == "" {
			return http.StatusBadRequest, false, fmt.Errorf("query 'clusterID' cannot be empty")
		}
		permit, _, _, err = c.clusterPermission.CanDeleteCluster(user, projectID, clusterID)
	}
	if err != nil {
		return http.StatusInternalServerError, false, errors.Wrapf(err, "check permission failed")
	}
	return http.StatusOK, permit, nil
}

// checkBCSPermissionNamespace check bcs permission for namespace
func (c *checker) checkBCSPermissionNamespace(r *http.Request, action iam.ActionID,
	user string) (int, bool, error) {
	projectID := r.URL.Query().Get("projectID")
	if projectID == "" {
		return http.StatusBadRequest, false, fmt.Errorf("query 'projectID' cannot be empty")
	}
	clusterID := r.URL.Query().Get("clusterID")
	if clusterID == "" {
		return http.StatusBadRequest, false, fmt.Errorf("query 'clusterID' cannot be empty")
	}
	var permit bool
	var err error
	switch action {
	case iamnamespace.NameSpaceCreate:
		permit, _, _, err = c.namespacePermission.CanCreateNamespace(user, projectID, clusterID, false)
	case iamnamespace.NameSpaceView:
		ns := r.URL.Query().Get("namespace")
		if ns == "" {
			return http.StatusBadRequest, false, fmt.Errorf("query 'clusterID' cannot be empty")
		}
		permit, _, _, err = c.namespacePermission.CanViewNamespace(user, projectID, clusterID, ns, false)
	case iamnamespace.NameSpaceUpdate:
		ns := r.URL.Query().Get("namespace")
		if ns == "" {
			return http.StatusBadRequest, false, fmt.Errorf("query 'clusterID' cannot be empty")
		}
		permit, _, _, err = c.namespacePermission.CanUpdateNamespace(user, projectID, clusterID, ns, false)
	case iamnamespace.NameSpaceDelete:
		ns := r.URL.Query().Get("namespace")
		if ns == "" {
			return http.StatusBadRequest, false, fmt.Errorf("query 'clusterID' cannot be empty")
		}
		permit, _, _, err = c.namespacePermission.CanDeleteNamespace(user, projectID, clusterID, ns, false)
	case iamnamespace.NameSpaceList:
		permit, _, _, err = c.namespacePermission.CanListNamespace(user, projectID, clusterID, false)
	}
	if err != nil {
		return http.StatusInternalServerError, false, errors.Wrapf(err, "check permission failed")
	}
	return http.StatusOK, permit, nil
}

// checkBCSPermissionNamespaceScoped check bcs permission namespace scoped
func (c *checker) checkBCSPermissionNamespaceScoped(r *http.Request, action iam.ActionID,
	user string) (int, bool, error) {
	projectID := r.URL.Query().Get("projectID")
	if projectID == "" {
		return http.StatusBadRequest, false, fmt.Errorf("query 'projectID' cannot be empty")
	}
	clusterID := r.URL.Query().Get("clusterID")
	if clusterID == "" {
		return http.StatusBadRequest, false, fmt.Errorf("query 'clusterID' cannot be empty")
	}
	ns := r.URL.Query().Get("namespace")
	if ns == "" {
		return http.StatusBadRequest, false, fmt.Errorf("query 'clusterID' cannot be empty")
	}
	var permit bool
	var err error
	switch action {
	case iamnamespace.NameSpaceScopedCreate:
		permit, _, _, err = c.namespacePermission.CanCreateNamespaceScopedResource(user, projectID, clusterID, ns)
	case iamnamespace.NameSpaceScopedView:
		permit, _, _, err = c.namespacePermission.CanViewNamespaceScopedResource(user, projectID, clusterID, ns)
	case iamnamespace.NameSpaceScopedUpdate:
		permit, _, _, err = c.namespacePermission.CanUpdateNamespaceScopedResource(user, projectID, clusterID, ns)
	case iamnamespace.NameSpaceScopedDelete:
		permit, _, _, err = c.namespacePermission.CanDeleteNamespaceScopedResource(user, projectID, clusterID, ns)
	}
	if err != nil {
		return http.StatusInternalServerError, false, errors.Wrapf(err, "check permission failed")
	}
	return http.StatusOK, permit, nil
}
