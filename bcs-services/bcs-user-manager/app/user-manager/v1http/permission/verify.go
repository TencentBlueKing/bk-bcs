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

package permission

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/audit"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/namespace"
	authUtils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/component"
	blog "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/log"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/parser"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
)

var (
	// ErrServerNotInited err server not init
	ErrServerNotInited = errors.New("VerifyPermissionClient server not init")

	// ErrContextTimeout err context timeout
	ErrContextTimeout = errors.New("operation timeout")
)

const (
	// clusterScopedType
	clusterScopedType string = "cluster_scoped"
	// namespaceScopedType
	namespaceScopedType string = "namespace_scoped"
	// namespaceType
	namespaceType string = "namespace"

	// prefixClustersAPIK8S
	prefixClustersAPIK8S string = "/clusters"
	// prefixProjectsAPIK8S
	prefixProjectsAPIK8S string = "/projects"

	// defaultTimeout
	defaultTimeout = 2 * time.Second
)

// NewPermVerifyClient verify permission client
func NewPermVerifyClient(swi bool, iam iam.PermClient) *PermVerifyClient {
	return &PermVerifyClient{
		PermSwitch: swi,
		PermClient: iam,
	}
}

// PermVerifyClient permission client
type PermVerifyClient struct {
	PermSwitch bool
	PermClient iam.PermClient
}

// returnClusterType
func returnClusterType(resource ClusterResource) ClusterType {
	if resource.ProjectID == "" {
		return Single
	}

	return Shared
}

// VerifyClusterPermission verify cluster permission: single and shared cluster
func (cli *PermVerifyClient) VerifyClusterPermission(ctx context.Context, user *models.BcsUser, action string,
	resource ClusterResource) (bool, string) {
	if cli == nil {
		return false, ErrServerNotInited.Error()
	}

	switch resource.ClusterType {
	case K8s:
		// extract namespace
		requestInfo, err := getK8sRequestAPIInfo(action, resource.URL)
		if err != nil {
			blog.Log(ctx).Infof("VerifyClusterPermission getK8sRequestAPIInfo failed: %v", err.Error())
			return false, fmt.Sprintf("VerifyClusterPermission getK8sRequestAPIInfo failed: %v", err.Error())
		}
		blog.Log(ctx).Infof("PermVerifyClient VerifyClusterPermission getK8sRequestAPIInfo requestInfo %+v",
			requestInfo)

		// extract URL namespace
		if resource.Namespace == "" {
			resource.Namespace = requestInfo.Namespace
		}
		resource.ResourceType = requestInfo.Resource

		// retry verify permission
		allowed := false
		message := ""
		err = utils.RetryWithTimeout(func() error {
			blog.Log(ctx).Infof("try to verify %s permission", user.Name)
			var innerError error
			allowed, innerError = cli.verifyUserK8sClusterPermission(ctx, user, action, resource, requestInfo)
			return innerError
		}, utils.RetryAttempts(3), utils.RetryTimeout(defaultTimeout))
		if err != nil {
			message = err.Error()
		}
		blog.Log(ctx).Infof("user %s access to type: %s, resource: %v, action: %s, permission: %t",
			user.Name, resource.ClusterType, resource, action, allowed)

		return allowed, message
	case Mesos:
		allowed, message := verifyResourceReplica(user.ID, "cluster", resource.ClusterID, action)
		blog.Log(ctx).Infof("user %s access to type: %s, action: %s, permission: %t",
			user.Name, "cluster", action, allowed)

		return allowed, message
	default:
		return false, fmt.Sprintf("unsupported cluster_type %s", resource.ClusterType)
	}
}

// verifyUserK8sClusterPermission
func (cli *PermVerifyClient) verifyUserK8sClusterPermission(ctx context.Context, user *models.BcsUser, action string,
	resource ClusterResource, requestInfo *parser.RequestInfo) (bool, error) {
	if cli == nil {
		return false, ErrServerNotInited
	}

	// check resourceType
	verifyType, err := cli.checkResourceType(resource, requestInfo)
	if err != nil {
		return false, err
	}
	blog.Log(ctx).Infof("PermVerifyClient verifyUserK8sClusterPermission verifyType[%s]", verifyType)

	switch verifyType {
	case clusterScopedType:
		// clusterScopedType nonResourceRequest will just skip
		if !requestInfo.IsResourceRequest {
			blog.Log(ctx).Infof("verifyUserClusterScopedPermission skip nonResourceRequest[%s]", requestInfo.Path)
			return true, nil
		}
		// verify cluster scope permission
		clusterScopedAllow, err := cli.verifyUserClusterScopedPermission(ctx, user.Name, action, resource)
		if err != nil {
			blog.Log(ctx).Errorf("verifyUserClusterScopedPermission failed: %v", err)
			return false, err
		}
		if !clusterScopedAllow {
			permission := cli.applyForPermission(action, verifyType)
			return clusterScopedAllow, fmt.Errorf("verifyUserClusterScopedPermission failed: %v, please apply for permission %s",
				err, permission)
		}

		return clusterScopedAllow, nil
	case namespaceType:
		// verify namespace permission
		namespaceAllow, err := cli.verifyUserNamespacePermission(ctx, user.Name, action, resource)
		if err != nil {
			blog.Log(ctx).Errorf("verifyUserNamespacePermission failed: %v", err)
			return false, err
		}
		if !namespaceAllow {
			permission := cli.applyForPermission(action, verifyType)
			return namespaceAllow, fmt.Errorf("verifyUserNamespacePermission failed: %v, please apply for permission %s",
				err, permission)
		}

		return namespaceAllow, nil
	case namespaceScopedType:
		var (
			namespaceScopedAllow bool
			err                  error
		)
		// client 类型的用户，首先使用 user-manager 的数据库鉴权，没有权限后再向权限中心鉴权
		if user.IsClient() {
			namespaceScopedAllow, err = cli.verifyClientNSScopedPermission(ctx, user, action, resource)
		} else {
			namespaceScopedAllow, err = cli.verifyUserNamespaceScopedPermission(ctx, user.Name, action, resource)
		}
		if err != nil {
			blog.Log(ctx).Errorf("verifyUserNamespaceScopedPermission failed: %s", err.Error())
			return false, err
		}
		if !namespaceScopedAllow {
			permission := cli.applyForPermission(action, verifyType)
			return namespaceScopedAllow, fmt.Errorf(
				"verifyUserNamespaceScopedPermission failed: %v, please apply for permission %s",
				err, permission)
		}

		return namespaceScopedAllow, nil
	default:
		return false, fmt.Errorf("unsupport verifyType[%s]", verifyType)
	}
}

// applyForPermission
func (cli *PermVerifyClient) applyForPermission(action string, verifyType string) string {
	permission := ""
	switch verifyType {
	// cluster scoped type
	case clusterScopedType:
		switch action {
		case http.MethodPost:
			permission = cluster.ClusterScopedCreate.String()
		case http.MethodDelete:
			permission = cluster.ClusterScopedDelete.String()
		case http.MethodPut, http.MethodPatch:
			permission = cluster.ClusterScopedUpdate.String()
		case http.MethodGet:
			permission = cluster.ClusterScopedView.String()
		}
	// namespace type
	case namespaceType:
		switch action {
		case http.MethodPost:
			permission = namespace.NameSpaceCreate.String()
		case http.MethodDelete:
			permission = namespace.NameSpaceDelete.String()
		case http.MethodPut, http.MethodPatch:
			permission = namespace.NameSpaceUpdate.String()
		case http.MethodGet:
			permission = namespace.NameSpaceView.String()
		}
	// namespace scoped type
	case namespaceScopedType:
		switch action {
		case http.MethodPost:
			permission = namespace.NameSpaceScopedCreate.String()
		case http.MethodDelete:
			permission = namespace.NameSpaceScopedDelete.String()
		case http.MethodPut, http.MethodPatch:
			permission = namespace.NameSpaceScopedUpdate.String()
		case http.MethodGet:
			permission = namespace.NameSpaceScopedView.String()
		}
	}

	return permission
}

// checkResourceType
func (cli *PermVerifyClient) checkResourceType(resource ClusterResource, req *parser.RequestInfo) (string, error) {
	if resource.ClusterID == "" {
		return "", errors.New("cluster_resource clusterID not null")
	}

	if req.Resource == "namespaces" {
		return namespaceType, nil
	}

	if req.Namespace != "" && req.Resource != "namespaces" {
		return namespaceScopedType, nil
	}

	return clusterScopedType, nil
}

// verifyUserNamespaceScopedPermission
func (cli *PermVerifyClient) verifyUserNamespaceScopedPermission(ctx context.Context, user string, action string,
	resource ClusterResource) (bool, error) {

	actionID := ""
	// get cluster type
	clusterType := returnClusterType(resource)

	// not namespace permission
	switch action {
	case http.MethodPost:
		actionID = namespace.NameSpaceScopedCreate.String()
	case http.MethodDelete:
		actionID = namespace.NameSpaceScopedDelete.String()
	case http.MethodPut, http.MethodPatch:
		actionID = namespace.NameSpaceScopedUpdate.String()
	case http.MethodGet:
		actionID = namespace.NameSpaceScopedView.String()
	default:
		return false, fmt.Errorf("invalid action[%s]", action)
	}

	// get project id
	if resource.ClusterID == "" || resource.Namespace == "" {
		return false, errors.New("resource clusterID or resource namespace is null")
	}
	project, err := cli.getProjectFromResource(ctx, resource)
	if err != nil {
		return false, fmt.Errorf("getProjectIDByClusterID[%s] failed", resource.ClusterID)
	}
	// if clusterType is shared, need to check namespace belong to project
	if clusterType == Shared {
		var exist bool
		exist, err = cli.checkNamespaceInProjectCluster(ctx, project.ProjectCode, resource.ClusterID, resource.Namespace)
		if err != nil {
			return false, err
		}
		if !exist {
			blog.Log(ctx).Infof("PermVerifyClient verifyUserNamespaceScopedPermission namespace[%s] not exist "+
				"project[%s] cluster[%s]", resource.Namespace, project.ProjectCode, resource.ClusterID) // nolint goconst
			return false, nil
		}
		blog.Log(ctx).Infof("PermVerifyClient verifyUserNamespaceScopedPermission namespace[%s] exist "+
			"project[%s] cluster[%s]", resource.Namespace, project.ProjectCode, resource.ClusterID)
	}

	// generate namespace id
	nameSpaceID := authUtils.CalcIAMNsID(resource.ClusterID, resource.Namespace)

	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}
	rn1 := iam.ResourceNode{
		System:    iam.SystemIDBKBCS,
		RType:     string(namespace.SysNamespace),
		RInstance: nameSpaceID,
		Rp: namespace.NamespaceScopedResourcePath{
			ProjectID: project.ProjectID,
			ClusterID: resource.ClusterID,
		},
		Attr: utils.GetResourceAttr(resource.ResourceType),
	}
	blog.Log(ctx).Infof("PermVerifyClient verifyUserNamespaceScopedPermission user[%s] actionID[%s] resource[%+v]",
		user, actionID, rn1)
	start := time.Now()

	// check permission
	allow, err := cli.PermClient.IsAllowedWithResource(actionID, req, []iam.ResourceNode{rn1}, true)
	blog.Log(ctx).Infof("PermVerifyClient verifyUserNamespaceScopedPermission taken %s", time.Since(start).String())
	if err != nil {
		blog.Log(ctx).Errorf("perm_client check namespaceScoped resource permission failed: %v", err)
		return false, err
	}

	go addAudit(ctx, user, project.ProjectCode, action, actionID, resource)
	return allow, nil
}

// checkNamespaceInProjectCluster
func (cli *PermVerifyClient) checkNamespaceInProjectCluster(ctx context.Context, projectCode, clusterID string,
	namespace string) (bool, error) {
	// 获取共享集群所在项目的命名空间列表，检查是否属于该项目
	namespaceList, err := component.GetCachedClusterNamespaces(ctx, projectCode, clusterID)
	if err != nil {
		blog.Log(ctx).Errorf("checkNamespaceInProjectCluster failed: %v", err)
		return false, err
	}

	for _, v := range namespaceList {
		if v.Name == namespace {
			return true, nil
		}
	}
	return false, nil
}

// verifyUserNamespacePermission
func (cli *PermVerifyClient) verifyUserNamespacePermission(ctx context.Context, user string, action string,
	resource ClusterResource) (bool, error) {

	var (
		// nolint
		actionID              = ""
		isClusterResourcePerm = false
		clusterType           ClusterType
	)
	// get cluster type
	clusterType = returnClusterType(resource)

	// namespace permission
	switch action {
	case http.MethodPost:
		actionID = namespace.NameSpaceCreate.String()
		isClusterResourcePerm = true
	case http.MethodDelete:
		actionID = namespace.NameSpaceDelete.String()
	case http.MethodPut, http.MethodPatch:
		actionID = namespace.NameSpaceUpdate.String()
	case http.MethodGet:
		actionID = namespace.NameSpaceView.String()
		if resource.Namespace == "" {
			actionID = namespace.NameSpaceList.String()
			isClusterResourcePerm = true
		}
	default:
		return false, fmt.Errorf("invalid action[%s]", action)
	}

	// get project id
	if clusterType == Shared && (actionID != namespace.NameSpaceView.String() &&
		actionID != namespace.NameSpaceList.String()) {
		return false, fmt.Errorf("verifyUserNamespacePermission shared cluster[%s] not support %s permission %s",
			resource.ClusterID, namespaceScopedType, actionID)
	}

	project, err := cli.getProjectFromResource(ctx, resource)
	if err != nil {
		return false, err
	}

	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}

	// get resource
	rn1, err := cli.getResource(ctx, isClusterResourcePerm, clusterType, resource, project)
	if err != nil {
		return false, err
	}
	if rn1 == nil {
		return false, nil
	}
	blog.Log(ctx).Infof("PermVerifyClient verifyUserNamespacePermission user[%s] actionID[%s] resource[%+v]",
		user, actionID, rn1)
	start := time.Now()

	allow, err := cli.PermClient.IsAllowedWithResource(actionID, req, []iam.ResourceNode{*rn1}, true)
	blog.Log(ctx).Infof("PermVerifyClient verifyUserNamespacePermission taken %s", time.Since(start).String())
	if err != nil {
		blog.Log(ctx).Errorf("perm_client check namespace permission failed: %v", err)
		return false, err
	}

	go addAudit(ctx, user, project.ProjectCode, action, actionID, resource)
	return allow, nil
}

// getResource get resource
func (cli *PermVerifyClient) getResource(ctx context.Context, isClusterResourcePerm bool, clusterType ClusterType,
	resource ClusterResource, project *component.Project) (*iam.ResourceNode, error) {
	// cal nameSpace perm ID

	var rn1 *iam.ResourceNode
	if isClusterResourcePerm {
		rn1 = &iam.ResourceNode{
			System:    iam.SystemIDBKBCS,
			RType:     string(cluster.SysCluster),
			RInstance: resource.ClusterID,
			Rp: namespace.NamespaceResourcePath{
				ProjectID:     project.ProjectID,
				IsClusterPerm: isClusterResourcePerm,
			},
		}
	} else {
		// if clusterType is shared, need to check namespace belong to project
		if clusterType == Shared {
			exist, err := cli.checkNamespaceInProjectCluster(ctx, project.ProjectCode, resource.ClusterID, resource.Namespace)
			if err != nil {
				return nil, err
			}
			if !exist {
				blog.Log(ctx).Infof("PermVerifyClient verifyUserNamespacePermission namespace[%s] not exist "+
					"project[%s] cluster[%s]", resource.Namespace, project.ProjectCode, resource.ClusterID)
				return nil, nil
			}
			blog.Log(ctx).Infof("PermVerifyClient verifyUserNamespacePermission namespace[%s] exist "+
				"project[%s] cluster[%s]", resource.Namespace, project.ProjectCode, resource.ClusterID)
		}
		// generate namespace id
		nameSpaceID := authUtils.CalcIAMNsID(resource.ClusterID, resource.Namespace)
		rn1 = &iam.ResourceNode{
			System:    iam.SystemIDBKBCS,
			RType:     string(namespace.SysNamespace),
			RInstance: nameSpaceID,
			Rp: namespace.NamespaceResourcePath{
				ProjectID: project.ProjectID,
				ClusterID: resource.ClusterID,
			},
		}
	}
	return rn1, nil
}

// verifyClientNSScopedPermission
func (cli *PermVerifyClient) verifyClientNSScopedPermission(ctx context.Context, user *models.BcsUser, action string,
	resource ClusterResource) (bool, error) {
	blog.Log(ctx).Infof("verifyClientNamespaceScopedPermission for user %s, type %s, resource %s, action %s",
		user.Name, NamespaceScoped, resource.Namespace, action)
	nsAllow, _ := verifyResourceReplica(user.ID, NamespaceScoped, resource.Namespace, action)
	if !nsAllow {
		return cli.verifyUserNamespaceScopedPermission(ctx, user.Name, action, resource)
	}
	return nsAllow, nil
}

// getProjectFromResource
func (cli *PermVerifyClient) getProjectFromResource(
	ctx context.Context, resource ClusterResource) (*component.Project, error) {
	if resource.ClusterID == "" {
		return nil, fmt.Errorf("resource clusterID is null")
	}

	var (
		projectID string
	)
	if resource.ProjectID != "" {
		projectID = resource.ProjectID
	} else {
		cls, err := component.GetClusterByClusterID(ctx, resource.ClusterID)
		if err != nil {
			return nil, fmt.Errorf("GetClusterByClusterID[%s] failed", resource.ClusterID)
		}
		projectID = cls.ProjectID
		blog.Log(ctx).Infof("got projectID %s from cluster %s", projectID, resource.ClusterID)
	}
	project, err := component.GetProjectWithCache(ctx, projectID)
	if err != nil {
		blog.Log(ctx).Errorf("get project with projectID failed: %v", err)
		return nil, err
	}

	return project, nil
}

// verifyUserClusterScopedPermission
func (cli *PermVerifyClient) verifyUserClusterScopedPermission(ctx context.Context, user string, action string,
	resource ClusterResource) (bool, error) {
	actionID := ""
	clusterType := returnClusterType(resource)
	if clusterType == Shared {
		return false, fmt.Errorf("shared cluster[%s] not support %s permission", resource.ClusterID, clusterScopedType)
	}

	// cluster scoped permission
	switch action {
	case http.MethodPost:
		actionID = cluster.ClusterScopedCreate.String()
	case http.MethodPut, http.MethodPatch:
		actionID = cluster.ClusterScopedUpdate.String()
	case http.MethodGet:
		actionID = cluster.ClusterScopedView.String()
	case http.MethodDelete:
		actionID = cluster.ClusterScopedDelete.String()
	default:
		return false, fmt.Errorf("invalid action[%s]", action)
	}

	project, err := cli.getProjectFromResource(ctx, resource)
	if err != nil {
		return false, err
	}
	blog.Log(ctx).Infof("verifyUserClusterScopedPermission getProjectIDByClusterID[%s]: %s", resource.ClusterID,
		project.ProjectCode)

	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}

	rn1 := iam.ResourceNode{
		System:    iam.SystemIDBKBCS,
		RType:     string(cluster.SysCluster),
		RInstance: resource.ClusterID,
		Rp: cluster.ClusterScopedResourcePath{
			ProjectID: project.ProjectID,
		},
	}

	// check permission
	start := time.Now()
	blog.Log(ctx).Infof("PermVerifyClient verifyUserClusterScopedPermission user[%s] actionID[%s] resource[%+v]",
		user, actionID, rn1)
	allow, err := cli.PermClient.IsAllowedWithResource(actionID, req, []iam.ResourceNode{rn1}, true)
	blog.Log(ctx).Infof("PermVerifyClient verifyUserClusterScopedPermission taken %s", time.Since(start).String())
	if err != nil {
		blog.Log(ctx).Errorf("perm_client check cluster permission failed: %v", err)
		return false, err
	}

	go addAudit(ctx, user, project.ProjectCode, action, actionID, resource)
	return allow, nil
}

// getK8sRequestAPIInfo
func getK8sRequestAPIInfo(method, url string) (*parser.RequestInfo, error) {
	resolver := parser.NewRequestInfoResolver()
	dstURL := transToAPIServerURL(url)
	if dstURL == "" {
		return nil, fmt.Errorf("url format err: %s", url)
	}

	req, err := http.NewRequest(method, dstURL, nil)
	if err != nil {
		return nil, err
	}

	requestInfo, err := resolver.NewRequestInfo(req)
	if err != nil {
		return nil, err
	}

	return requestInfo, nil
}

// transToAPIServerURL xxx
// trans bcs clusters API to k8s API url
func transToAPIServerURL(url string) string {
	if strings.HasPrefix(url, prefixClustersAPIK8S) {
		urlStrs := strings.Split(url, "/")
		if len(urlStrs) <= 3 {
			return ""
		}

		return "/" + strings.Join(urlStrs[3:], "/")
	}

	if strings.HasPrefix(url, prefixProjectsAPIK8S) {
		urlStrs := strings.Split(url, "/")
		if len(urlStrs) <= 5 {
			return ""
		}

		return "/" + strings.Join(urlStrs[5:], "/")
	}

	return ""
}

func addAudit(pCtx context.Context, user, projectCode, method, actionID string, res ClusterResource) {
	requestID := utils.GetRequestIDFromContext(pCtx)
	ctx := context.WithValue(context.Background(), utils.ContextValueKeyRequestID, requestID)

	var activityType audit.ActivityType
	switch method {
	case http.MethodPost:
		activityType = audit.ActivityTypeCreate
	case http.MethodPut:
		activityType = audit.ActivityTypeUpdate
	case http.MethodDelete:
		activityType = audit.ActivityTypeDelete
	default:
		// Get 类型不需要审计
		return
	}

	auditCtx := audit.RecorderContext{
		Username:  user,
		RequestID: utils.GetRequestIDFromContext(ctx),
		StartTime: time.Now(),
		EndTime:   time.Now(),
	}
	resource := audit.Resource{
		ProjectCode:  projectCode,
		ResourceType: audit.ResourceTypeK8SResource,
		ResourceID:   res.URL,
		ResourceName: res.URL,
		ResourceData: map[string]interface{}{
			"ProjectCode": projectCode,
			"ClusterID":   res.ClusterID,
			"Namespace":   res.Namespace,
			"URL":         res.URL,
			"Method":      method,
		},
	}

	action := audit.Action{
		ActionID:     actionID,
		ActivityType: activityType,
	}

	result := audit.ActionResult{
		Status: audit.ActivityStatusSuccess,
		ResultContent: fmt.Sprintf("user %s access to %s in cluster %s, method: %s", auditCtx.Username,
			resource.ResourceID, res.ClusterID, method),
	}

	// audit
	_ = component.GetAuditClient().R().DisableActivity().
		SetContext(auditCtx).SetResource(resource).SetAction(action).SetResult(result).Do()

	// activity
	err := sqlstore.CreateActivity([]*models.Activity{
		{
			ProjectCode:  resource.ProjectCode,
			ResourceType: string(resource.ResourceType),
			ResourceName: resource.ResourceName,
			ResourceID:   resource.ResourceID,
			ActivityType: string(action.ActivityType),
			Status:       models.GetStatus(string(result.Status)),
			Username:     auditCtx.Username,
			Description:  result.ResultContent,
		},
	})
	if err != nil {
		blog.Log(ctx).Errorf("create activity failed: %v", err.Error())
		return
	}
}
