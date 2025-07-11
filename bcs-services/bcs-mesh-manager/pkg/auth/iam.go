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
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
	authutils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	"go-micro.dev/v4/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/cmd/mesh-manager/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/utils"
)

var (
	// IAMClient iam client
	IAMClient iam.PermClient
	// ClusterIamClient cluster iam client
	ClusterIamClient *cluster.BCSClusterPerm
	// NamespaceIamClient namespace iam client
	NamespaceIamClient *namespace.BCSNamespacePerm
	// ProjectIamClient project iam client
	ProjectIamClient *project.BCSProjectPerm
)

// InitPermClient new a perm client
func InitPermClient(iamClient iam.PermClient) {
	ClusterIamClient = cluster.NewBCSClusterPermClient(iamClient)
	NamespaceIamClient = namespace.NewBCSNamespacePermClient(iamClient)
	ProjectIamClient = project.NewBCSProjectPermClient(iamClient)
}

// CheckMeshPermissions 检查网格操作相关权限
func CheckMeshPermissions(username, projectID string, allClusters []string) map[string]bool {
	permissions := map[string]bool{
		common.MeshManagerUpdateIstio:    false,
		common.MeshManagerDeleteIstio:    false,
		common.MeshManagerGetIstioDetail: false,
	}

	if len(allClusters) == 0 {
		return permissions
	}

	permissions[common.MeshManagerUpdateIstio] = checkPermission(username,
		projectID, allClusters, namespace.CanUpdateNamespaceScopedResourceOperation)

	permissions[common.MeshManagerDeleteIstio] = checkPermission(username,
		projectID, allClusters, namespace.CanDeleteNamespaceScopedResourceOperation)

	permissions[common.MeshManagerGetIstioDetail] = checkPermission(username,
		projectID, allClusters, namespace.CanViewNamespaceScopedResourceOperation)

	return permissions
}

// checkPermission 通用的权限检查函数
func checkPermission(username, projectID string, allClusters []string, operation string) bool {

	credentialScope := options.CredentialScope{
		ProjectID: projectID,
		Namespace: common.IstioNamespace,
	}

	for _, clusterID := range allClusters {
		credentialScope.ClusterID = clusterID
		allow, _, _, err := CallIAM(username, operation, credentialScope)
		if err != nil {
			blog.Errorf("permission check failed for cluster %s: %v", clusterID, err)
			return false
		}
		if !allow {
			return false
		}
	}
	return true
}

// CheckUserPerm check user perm
func CheckUserPerm(ctx context.Context, req server.Request, username string) (bool, error) {
	action, ok := ActionPermissions[req.Method()]
	blog.Infof("CheckUserPerm called with username: %s, action: %s, ok: %v", username, action, ok)
	if !ok {
		return false, fmt.Errorf("operation %s is not authorized", req.Method())
	}
	return checkMeshUserPerm(ctx, req, username, action)
}

// checkMeshUserPerm 检查网格操作的用户权限
func checkMeshUserPerm(ctx context.Context, req server.Request, username, action string) (bool, error) {
	projectID := utils.GetProjectIDFromCtx(ctx)
	if projectID == "" {
		return false, fmt.Errorf("projectID not found in context")
	}

	// 根据请求类型获取集群信息并检查权限
	switch req.Method() {
	case common.MeshManagerInstallIstio:
		return checkPermissionForInstall(ctx, username, projectID, action)
	case common.MeshManagerUpdateIstio, common.MeshManagerDeleteIstio, common.MeshManagerGetIstioDetail:
		return checkPermissionForMeshOperation(ctx, username, projectID, action)
	case common.MeshManagerListIstio:
		return checkPermissionForList(username, projectID, action)
	default:
		return false, fmt.Errorf("unsupported method: %s", req.Method())
	}
}

// checkPermissionForInstall 检查安装权限
func checkPermissionForInstall(ctx context.Context, username, projectID, action string) (bool, error) {
	clusters, err := getClustersFromContext(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get clusters from context: %w", err)
	}
	return checkClustersPermission(username, projectID, action, clusters)
}

// checkPermissionForMeshOperation 检查网格操作权限（更新、删除、获取详情）
func checkPermissionForMeshOperation(ctx context.Context, username, projectID, action string) (bool, error) {
	clusters, err := getClustersFromMesh(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get clusters from mesh: %w", err)
	}
	return checkClustersPermission(username, projectID, action, clusters)
}

// checkPermissionForList 检查列表权限（只需要项目查看权限）
func checkPermissionForList(username, projectID, action string) (bool, error) {
	allow, _, _, err := CallIAM(username, action, options.CredentialScope{
		ProjectID: projectID,
	})
	if err != nil {
		blog.Errorf("permission check failed for list operation: %v", err)
		return false, fmt.Errorf("permission check failed for list operation: %w", err)
	}
	return allow, nil
}

// getClustersFromContext 从 context 获取集群列表
func getClustersFromContext(ctx context.Context) ([]string, error) {
	primaryClusters := utils.GetPrimaryClustersFromCtx(ctx)
	remoteClusters := utils.GetRemoteClustersFromCtx(ctx)

	if len(primaryClusters) == 0 {
		return nil, fmt.Errorf("primaryClusters not found in context")
	}

	return utils.MergeSlices(primaryClusters, remoteClusters), nil
}

// getClustersFromMesh 从网格信息获取集群列表
func getClustersFromMesh(ctx context.Context) ([]string, error) {
	meshID := utils.GetMeshIDFromCtx(ctx)
	if meshID == "" {
		return nil, fmt.Errorf("meshID not found in context")
	}

	mesh, err := getMeshByID(ctx, meshID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mesh by ID %s: %w", meshID, err)
	}

	return utils.MergeSlices(mesh.PrimaryClusters, mesh.RemoteClusters), nil
}

// checkClustersPermission 检查集群权限的公共逻辑
func checkClustersPermission(username, projectID, action string, clusters []string) (bool, error) {
	for _, clusterID := range clusters {
		allow, _, _, err := CallIAM(username, action, options.CredentialScope{
			ProjectID: projectID,
			ClusterID: clusterID,
			Namespace: common.IstioNamespace,
		})
		if err != nil {
			blog.Errorf("permission check failed for cluster %s: %v", clusterID, err)
			return false, fmt.Errorf("permission check failed for cluster %s: %w", clusterID, err)
		}
		if !allow {
			return false, nil
		}
	}
	return true, nil
}

var meshModel store.MeshManagerModel

// SetMeshModel 设置网格模型，用于权限检查
func SetMeshModel(model store.MeshManagerModel) {
	meshModel = model
}

// getMeshByID 根据网格ID获取网格信息
func getMeshByID(ctx context.Context, meshID string) (*entity.MeshIstio, error) {
	if meshModel == nil {
		return nil, fmt.Errorf("mesh model not initialized")
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyMeshID: meshID,
	})

	mesh, err := meshModel.Get(ctx, cond)
	if err != nil {
		return nil, err
	}

	if mesh == nil {
		return nil, fmt.Errorf("mesh not found: %s", meshID)
	}

	return mesh, nil
}

// CallIAM call iam
func CallIAM(username, action string, resourceID options.CredentialScope) (bool, string,
	[]authutils.ResourceAction, error) {
	// 根据操作类型返回不同的权限结果
	switch action {
	case project.CanViewProjectOperation:
		return ProjectIamClient.CanViewProject(username, resourceID.ProjectID)
	case namespace.CanCreateNamespaceScopedResourceOperation:
		return NamespaceIamClient.CanCreateNamespaceScopedResource(username, resourceID.ProjectID,
			resourceID.ClusterID, resourceID.Namespace)
	case namespace.CanUpdateNamespaceScopedResourceOperation:
		return NamespaceIamClient.CanUpdateNamespaceScopedResource(username, resourceID.ProjectID,
			resourceID.ClusterID, resourceID.Namespace)
	case namespace.CanDeleteNamespaceScopedResourceOperation:
		return NamespaceIamClient.CanDeleteNamespaceScopedResource(username, resourceID.ProjectID,
			resourceID.ClusterID, resourceID.Namespace)
	case namespace.CanViewNamespaceScopedResourceOperation:
		return NamespaceIamClient.CanViewNamespaceScopedResource(username, resourceID.ProjectID,
			resourceID.ClusterID, resourceID.Namespace)
	default:
		blog.Infof("Denying operation: %s", action)
		return false, "", nil, nil
	}
}
