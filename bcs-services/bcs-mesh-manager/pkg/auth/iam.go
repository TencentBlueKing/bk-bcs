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

// GetMeshOpPerm 获取服务网格页面操作权限状态
func GetMeshOpPerm(username, projectID string, allClusters []string) map[string]bool {
	permissions := map[string]bool{
		common.MeshManagerUpdateIstio:    false,
		common.MeshManagerDeleteIstio:    false,
		common.MeshManagerGetIstioDetail: false,
	}

	if len(allClusters) == 0 {
		return permissions
	}

	allow, _, _, err := CheckClustersPerm(username,
		projectID, namespace.CanUpdateNamespaceScopedResourceOperation, allClusters)
	permissions[common.MeshManagerUpdateIstio] = err == nil && allow

	allow, _, _, err = CheckClustersPerm(username,
		projectID, namespace.CanDeleteNamespaceScopedResourceOperation, allClusters)
	permissions[common.MeshManagerDeleteIstio] = err == nil && allow

	allow, _, _, err = CheckClustersPerm(username,
		projectID, namespace.CanViewNamespaceScopedResourceOperation, allClusters)
	permissions[common.MeshManagerGetIstioDetail] = err == nil && allow

	return permissions
}

// CheckUserPerm check user perm
func CheckUserPerm(ctx context.Context, req server.Request, username string) (bool, error) {
	action, ok := ActionPermissions[req.Method()]
	blog.Infof("CheckUserPerm called with username: %s, action: %s, ok: %v", username, action, ok)
	if !ok {
		return false, fmt.Errorf("operation %s is not authorized", req.Method())
	}
	projectID := utils.GetProjectIDFromCtx(ctx)
	if projectID == "" {
		return false, fmt.Errorf("projectID not found in context")
	}

	// 根据请求类型获取集群信息并检查权限
	switch req.Method() {
	case common.MeshManagerInstallIstio,
		common.MeshManagerUpdateIstio,
		common.MeshManagerDeleteIstio,
		common.MeshManagerGetIstioDetail:
		clusters, err := getClustersFromRequest(ctx, req)
		if err != nil {
			return false, fmt.Errorf("failed to get clusters from request: %w", err)
		}
		return checkPermForMeshOp(username, projectID, action, clusters)
	case common.MeshManagerListIstio, common.MeshManagerGetClusterInfo:
		return checkPermForList(username, projectID, action)
	default:
		return false, fmt.Errorf("unsupported method: %s", req.Method())
	}
}

// getClustersFromRequest 根据请求类型获取集群列表
func getClustersFromRequest(ctx context.Context, req server.Request) ([]string, error) {
	switch req.Method() {
	case common.MeshManagerInstallIstio:
		primaryClusters, remoteClusters, err := getClusters(req)
		if err != nil {
			return nil, fmt.Errorf("failed to get clusters from request: %w", err)
		}
		if len(primaryClusters) == 0 {
			return nil, fmt.Errorf("primaryClusters is empty")
		}
		return utils.MergeSlices(primaryClusters, remoteClusters), nil

	case common.MeshManagerUpdateIstio, common.MeshManagerDeleteIstio, common.MeshManagerGetIstioDetail:
		// 网格操作：通过meshID查询数据库获取集群列表
		meshID, err := getMeshID(req)
		if err != nil {
			return nil, fmt.Errorf("failed to get meshID from request: %w", err)
		}
		if meshID == "" {
			return nil, fmt.Errorf("meshID is empty for method %s", req.Method())
		}
		mesh, err := getMeshByID(ctx, meshID)
		if err != nil {
			return nil, fmt.Errorf("failed to get mesh by ID %s: %w", meshID, err)
		}
		remoteClusters := make([]string, 0, len(mesh.RemoteClusters))
		for _, cluster := range mesh.RemoteClusters {
			remoteClusters = append(remoteClusters, cluster.ClusterID)
		}
		return utils.MergeSlices(mesh.PrimaryClusters, remoteClusters), nil

	default:
		return nil, fmt.Errorf("unsupported method for cluster extraction: %s", req.Method())
	}
}

// checkPermForMeshOp 检查网格操作权限
func checkPermForMeshOp(username, projectID, action string, clusters []string) (bool, error) {
	allow, actionList, resourceActionList, err := CheckClustersPerm(username, projectID, action, clusters)
	if err != nil {
		return false, fmt.Errorf("failed to check clusters permission: %w", err)
	}
	if allow {
		return true, nil
	}
	applyURL, err := BuildApplyURL(actionList)
	if err != nil {
		return false, fmt.Errorf("failed to build apply URL: %w", err)
	}
	return false, &authutils.PermDeniedError{
		Perms: authutils.PermData{
			ApplyURL:   applyURL,
			ActionList: resourceActionList,
		},
	}
}

// checkPermissionForList 检查列表权限（只需要项目查看权限）
func checkPermForList(username, projectID, action string) (bool, error) {
	allow, url, resourceActions, err := CallIAM(username, action, options.CredentialScope{
		ProjectID: projectID,
	})
	if err != nil {
		blog.Errorf("permission check failed for list operation: %v", err)
		return false, fmt.Errorf("permission check failed for list operation: %w", err)
	}
	if !allow && url != "" {
		return false, &authutils.PermDeniedError{
			Perms: authutils.PermData{
				ApplyURL:   url,
				ActionList: resourceActions,
			},
		}
	}
	return allow, nil
}

// CheckClustersPerm 检查多个集群下istio-system命名空间的权限
func CheckClustersPerm(username, projectID, action string, clusters []string) (
	bool, []iam.ApplicationAction, []authutils.ResourceAction, error) {
	// 权限申请构建列表 - 用于生成权限申请URL
	actionList := make([]iam.ApplicationAction, 0)
	// 缺失权限记录列表
	resourceActionList := make([]authutils.ResourceAction, 0)
	// 仅所有资源权限都校验通过时，才返回 true
	allowFlag := true
	// 标记是否申请过项目级别权限，避免重复申请
	projectBuilt := false

	for _, clusterID := range clusters {
		allow, _, resourceActions, err := CallIAM(username, action, options.CredentialScope{
			ProjectID: projectID,
			ClusterID: clusterID,
			Namespace: common.IstioNamespace,
		})
		if err != nil {
			blog.Errorf("permission check failed for cluster %s: %v", clusterID, err)
			return false,
				actionList,
				resourceActionList,
				fmt.Errorf("permission check failed for cluster %s: %w", clusterID, err)
		}
		if !allow {
			allowFlag = false
			// 收集权限检查失败的资源权限信息
			resourceActionList = append(resourceActionList, resourceActions...)
			if !projectBuilt {
				// 仅构建一次项目级别的权限申请信息
				actionList = append(actionList, buildProjectApplication(projectID))
				projectBuilt = true
			}
			// 计算 IAM 命名空间 ID
			namespaceID := authutils.CalcIAMNsID(clusterID, common.IstioNamespace)
			actionList = append(actionList, buildApplication(projectID, clusterID, namespaceID, action)...)
		}
	}
	if allowFlag {
		return true, actionList, resourceActionList, nil
	}
	return false, actionList, resourceActionList, nil
}

// BuildApplyURL 生成权限申请URL
func BuildApplyURL(actionList []iam.ApplicationAction) (string, error) {
	url, err := NamespaceIamClient.GenerateIAMApplicationURL(iam.SystemIDBKBCS, actionList)
	if err != nil {
		return "", fmt.Errorf("failed to generate IAM application URL: %w", err)
	}
	return url, nil
}

// buildProjectApplication 构建项目级别的权限申请
func buildProjectApplication(projectID string) iam.ApplicationAction {
	return project.BuildProjectApplicationInstance(project.ProjectApplicationAction{
		ActionID: project.ProjectView.String(),
		Data:     []string{projectID},
	})
}

// buildApplication 构建集群和命名空间级别的权限申请
func buildApplication(projectID, clusterID, namespaceID, action string) []iam.ApplicationAction {
	// 集群查看权限申请
	clusterApp := cluster.BuildClusterApplicationInstance(cluster.ClusterApplicationAction{
		ActionID: cluster.ClusterView.String(),
		Data:     []cluster.ProjectClusterData{{Project: projectID, Cluster: clusterID}},
	})

	// 命名空间查看权限申请
	nsApp := namespace.BuildNamespaceApplicationInstance(namespace.NamespaceApplicationAction{
		ActionID: namespace.NameSpaceView.String(),
		Data:     []namespace.ProjectNamespaceData{{Project: projectID, Cluster: clusterID, Namespace: namespaceID}},
	})

	// 根据具体操作类型确定命名空间域权限申请类型
	var namespaceScopedActionID string
	switch action {
	case namespace.CanCreateNamespaceScopedResourceOperation:
		namespaceScopedActionID = namespace.NameSpaceScopedCreate.String()
	case namespace.CanUpdateNamespaceScopedResourceOperation:
		namespaceScopedActionID = namespace.NameSpaceScopedUpdate.String()
	case namespace.CanDeleteNamespaceScopedResourceOperation:
		namespaceScopedActionID = namespace.NameSpaceScopedDelete.String()
	case namespace.CanViewNamespaceScopedResourceOperation:
		namespaceScopedActionID = namespace.NameSpaceScopedView.String()
	default:
		// 默认使用查看权限
		namespaceScopedActionID = namespace.NameSpaceScopedView.String()
	}

	// 命名空间域资源操作权限申请
	nssApp := namespace.BuildNSScopedAppInstance(namespace.NamespaceScopedApplicationAction{
		ActionID: namespaceScopedActionID,
		Data:     []namespace.ProjectNamespaceData{{Project: projectID, Cluster: clusterID, Namespace: namespaceID}},
	})

	return []iam.ApplicationAction{clusterApp, nsApp, nssApp}
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

// getMeshIDFromRequest 从请求中获取网格ID
func getMeshID(req server.Request) (string, error) {
	body := req.Body()
	b, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	type meshIDStruct struct {
		MeshID string `json:"meshID,omitempty"`
	}
	var m meshIDStruct
	if err := json.Unmarshal(b, &m); err != nil {
		return "", err
	}
	if m.MeshID != "" {
		return m.MeshID, nil
	}
	return "", fmt.Errorf("meshID not found in request for method %s", req.Method())
}

// getClusters 从安装请求中获取集群信息
func getClusters(req server.Request) ([]string, []string, error) {
	body := req.Body()
	b, err := json.Marshal(body)
	if err != nil {
		return nil, nil, err
	}

	type RemoteCluster struct {
		ClusterID string `json:"clusterID,omitempty"`
		Region    string `json:"region,omitempty"`
		JoinTime  int64  `json:"joinTime,omitempty"`
	}

	type clustersStruct struct {
		PrimaryClusters []string         `json:"primaryClusters,omitempty"`
		RemoteClusters  []*RemoteCluster `json:"remoteClusters,omitempty"`
	}

	var c clustersStruct
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, nil, err
	}

	var remoteClusters []string
	for _, cluster := range c.RemoteClusters {
		if cluster != nil && cluster.ClusterID != "" {
			remoteClusters = append(remoteClusters, cluster.ClusterID)
		}
	}

	return c.PrimaryClusters, remoteClusters, nil
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
