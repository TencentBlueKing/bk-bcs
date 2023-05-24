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

package permission

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/audit"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/namespace"
	authUtils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/cmanager"
	blog "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/log"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/parser"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/passcc"
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
	clusterScopedType   string = "cluster_scoped"
	namespaceScopedType string = "namespace_scoped"
	namespaceType       string = "namespace"

	prefixClustersAPIK8S string = "/clusters"
	prefixProjectsAPIK8S string = "/projects"

	defaultTimeout = 2 * time.Second
)

// NewPermVerifyClient verify permission client
func NewPermVerifyClient(swi bool, iam iam.PermClient,
	clusterCli *cmanager.ClusterManagerClient) *PermVerifyClient {
	return &PermVerifyClient{
		PermSwitch:    swi,
		PermClient:    iam,
		ClusterClient: clusterCli,
	}
}

// PermVerifyClient permission client
type PermVerifyClient struct {
	PermSwitch    bool
	PermClient    iam.PermClient
	ClusterClient *cmanager.ClusterManagerClient
}

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
			namespaceScopedAllow, err = cli.verifyClientNamespaceScopedPermission(ctx, user, action, resource)
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

func (cli *PermVerifyClient) applyForPermission(action string, verifyType string) string {
	permission := ""
	switch verifyType {
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
		return false, fmt.Errorf("invlid action[%s]", action)
	}

	if resource.ClusterID == "" || resource.Namespace == "" {
		return false, errors.New("resource clusterID or resource namespace is null")
	}
	projectID, err := cli.getProjectIDFromResource(ctx, resource)
	if err != nil {
		return false, fmt.Errorf("getProjectIDByClusterID[%s] failed", resource.ClusterID)
	}
	// if clusterType is shared, need to check namespace belong to project
	if clusterType == Shared {
		exist, err := cli.checkNamespaceInProjectCluster(ctx, projectID, resource.ClusterID, resource.Namespace)
		if err != nil {
			return false, err
		}
		if !exist {
			blog.Log(ctx).Infof("PermVerifyClient verifyUserNamespaceScopedPermission namespace[%s] not exist "+
				"project[%s] cluster[%s]", resource.Namespace, projectID, resource.ClusterID)
			return false, nil
		}
		blog.Log(ctx).Infof("PermVerifyClient verifyUserNamespaceScopedPermission namespace[%s] exist "+
			"project[%s] cluster[%s]", resource.Namespace, projectID, resource.ClusterID)
	}

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
			ProjectID: projectID,
			ClusterID: resource.ClusterID,
		},
	}
	blog.Log(ctx).Infof("PermVerifyClient verifyUserNamespaceScopedPermission user[%s] actionID[%s] resource[%+v]",
		user, actionID, rn1)
	start := time.Now()

	allow, err := cli.PermClient.IsAllowedWithResource(actionID, req, []iam.ResourceNode{rn1}, false)
	instanceData := map[string]interface{}{
		"ProjectID": projectID,
		"ClusterID": resource.ClusterID,
		"Namespace": resource.Namespace,
	}
	defer audit.AddEvent(actionID, string(rn1.RType), rn1.RInstance, user, allow, instanceData)
	blog.Log(ctx).Infof("PermVerifyClient verifyUserNamespaceScopedPermission taken %s", time.Since(start).String())
	if err != nil {
		blog.Log(ctx).Errorf("perm_client check namespaceScoped resource permission failed: %v", err)
		return false, err
	}
	return allow, nil
}

func (cli *PermVerifyClient) checkNamespaceInProjectCluster(ctx context.Context, projectID, clusterID string,
	namespace string) (bool, error) {
	// 获取共享集群所在项目的命名空间列表，检查是否属于该项目
	namespaceList, err := passcc.GetCCClient().GetProjectSharedNamespaces(projectID, clusterID)
	if err != nil {
		blog.Log(ctx).Errorf("checkNamespaceInProjectCluster failed: %v", err)
		return false, err
	}

	return utils.StringInSlice(namespace, namespaceList), nil
}

func (cli *PermVerifyClient) verifyUserNamespacePermission(ctx context.Context, user string, action string,
	resource ClusterResource) (bool, error) {

	var (
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

	if clusterType == Shared && actionID != namespace.NameSpaceView.String() {
		return false, fmt.Errorf("verifyUserNamespacePermission shared cluster[%s] not support %s permission %s",
			resource.ClusterID, namespaceScopedType, actionID)
	}

	projectID, err := cli.getProjectIDFromResource(ctx, resource)
	if err != nil {
		return false, err
	}

	// cal nameSpace perm ID
	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}

	var rn1 iam.ResourceNode
	if isClusterResourcePerm {
		rn1 = iam.ResourceNode{
			System:    iam.SystemIDBKBCS,
			RType:     string(cluster.SysCluster),
			RInstance: resource.ClusterID,
			Rp: namespace.NamespaceResourcePath{
				ProjectID:     projectID,
				IsClusterPerm: isClusterResourcePerm,
			},
		}
	} else {
		// if clusterType is shared, need to check namespace belong to project
		if clusterType == Shared {
			exist, err := cli.checkNamespaceInProjectCluster(ctx, projectID, resource.ClusterID, resource.Namespace)
			if err != nil {
				return false, err
			}
			if !exist {
				blog.Log(ctx).Infof("PermVerifyClient verifyUserNamespacePermission namespace[%s] not exist "+
					"project[%s] cluster[%s]", resource.Namespace, projectID, resource.ClusterID)
				return false, nil
			}
			blog.Log(ctx).Infof("PermVerifyClient verifyUserNamespacePermission namespace[%s] exist "+
				"project[%s] cluster[%s]", resource.Namespace, projectID, resource.ClusterID)
		}
		nameSpaceID := authUtils.CalcIAMNsID(resource.ClusterID, resource.Namespace)
		rn1 = iam.ResourceNode{
			System:    iam.SystemIDBKBCS,
			RType:     string(namespace.SysNamespace),
			RInstance: nameSpaceID,
			Rp: namespace.NamespaceResourcePath{
				ProjectID: projectID,
				ClusterID: resource.ClusterID,
			},
		}
	}

	blog.Log(ctx).Infof("PermVerifyClient verifyUserNamespacePermission user[%s] actionID[%s] resource[%+v]",
		user, actionID, rn1)
	start := time.Now()

	allow, err := cli.PermClient.IsAllowedWithResource(actionID, req, []iam.ResourceNode{rn1}, false)
	instanceData := map[string]interface{}{
		"ProjectID": projectID,
		"ClusterID": resource.ClusterID,
		"Namespace": resource.Namespace,
	}
	defer audit.AddEvent(actionID, string(rn1.RType), rn1.RInstance, user, allow, instanceData)
	blog.Log(ctx).Infof("PermVerifyClient verifyUserNamespacePermission taken %s", time.Since(start).String())
	if err != nil {
		blog.Log(ctx).Errorf("perm_client check namespace permission failed: %v", err)
		return false, err
	}

	return allow, nil
}

func (cli *PermVerifyClient) verifyClientNamespaceScopedPermission(ctx context.Context, user *models.BcsUser, action string,
	resource ClusterResource) (bool, error) {
	blog.Log(ctx).Infof("verifyClientNamespaceScopedPermission for user %s, type %s, resource %s, action %s",
		user.Name, NamespaceScoped, resource.Namespace, action)
	nsAllow, _ := verifyResourceReplica(user.ID, NamespaceScoped, resource.Namespace, action)
	if !nsAllow {
		return cli.verifyUserNamespaceScopedPermission(ctx, user.Name, action, resource)
	}
	return nsAllow, nil
}

func (cli *PermVerifyClient) getProjectIDFromResource(ctx context.Context, resource ClusterResource) (string, error) {
	if resource.ClusterID == "" {
		return "", fmt.Errorf("resource clusterID is null")
	}

	var (
		projectID string
		err       error
	)
	if resource.ProjectID != "" {
		projectID = resource.ProjectID
	} else {
		projectID, err = cli.ClusterClient.GetProjectIDByClusterID(resource.ClusterID)
		if err != nil {
			return "", fmt.Errorf("getProjectIDByClusterID[%s] failed", resource.ClusterID)
		}
		blog.Log(ctx).Infof("get projectID %s from cluster %s", resource.ClusterID, projectID)
	}

	return projectID, nil
}

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

	projectID, err := cli.getProjectIDFromResource(ctx, resource)
	if err != nil {
		return false, err
	}
	blog.Log(ctx).Infof("verifyUserClusterScopedPermission getProjectIDByClusterID[%s]: %s", resource.ClusterID,
		projectID)

	req := iam.PermissionRequest{
		SystemID: iam.SystemIDBKBCS,
		UserName: user,
	}

	rn1 := iam.ResourceNode{
		System:    iam.SystemIDBKBCS,
		RType:     string(cluster.SysCluster),
		RInstance: resource.ClusterID,
		Rp: cluster.ClusterScopedResourcePath{
			ProjectID: projectID,
		},
	}

	start := time.Now()
	blog.Log(ctx).Infof("PermVerifyClient verifyUserClusterScopedPermission user[%s] actionID[%s] resource[%+v]",
		user, actionID, rn1)
	allow, err := cli.PermClient.IsAllowedWithResource(actionID, req, []iam.ResourceNode{rn1}, false)
	instanceData := map[string]interface{}{
		"ProjectID": projectID,
		"ClusterID": resource.ClusterID,
		"Namespace": resource.Namespace,
	}
	defer audit.AddEvent(actionID, string(rn1.RType), rn1.RInstance, user, allow, instanceData)
	blog.Log(ctx).Infof("PermVerifyClient verifyUserClusterScopedPermission taken %s", time.Since(start).String())
	if err != nil {
		blog.Log(ctx).Errorf("perm_client check cluster permission failed: %v", err)
		return false, err
	}

	return allow, nil
}

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

func buildK8sOperationLog(user string, resource ClusterResource, info *parser.RequestInfo, allow bool) error {
	const (
		MessageTemplate = "user[%s] perm[%t] verb[%s] APIPrefix[%s] GVK[%s-%s-%s] namespace[%s] resource[%s] subresource[%s]"
	)

	log := &models.BcsOperationLog{
		ClusterType: resource.ClusterType.String(),
		ClusterID:   resource.ClusterID,
		Path:        resource.URL,
		Message: fmt.Sprintf(MessageTemplate, user, allow, info.Verb, info.APIPrefix, info.APIGroup, info.APIVersion,
			info.Resource, info.Namespace, info.Resource, info.Subresource),
		OpUser:    user,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := sqlstore.CreateOperationLog(log)
	if err != nil {
		return err
	}

	return nil
}

func buildAdminOperationLog(user string, req VerifyPermissionReq) error {
	const (
		MessageTemplate = "管理员用户[%s]操作[%s]资源类型[%s]资源[%s]allow[%v]"
	)

	log := &models.BcsOperationLog{
		ClusterType: req.ResourceType.String(),
		ClusterID:   req.Resource,
		Path:        req.RequestURL,
		Message:     fmt.Sprintf(MessageTemplate, user, req.Action, req.ResourceType, req.Resource, true),
		OpUser:      user,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := sqlstore.CreateOperationLog(log)
	if err != nil {
		return err
	}

	return nil
}
