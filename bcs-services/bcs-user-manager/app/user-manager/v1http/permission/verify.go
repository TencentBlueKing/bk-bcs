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
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/cmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/parser"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/namespace"
)

var (
	// ErrServerNotInited err server not init
	ErrServerNotInited = errors.New("VerifyPermissionClient server not init")
)

const (
	clusterScopedType   string = "cluster_scoped"
	namespaceScopedType string = "namespace_scoped"
	namespaceType       string = "namespace"

	prefixAPIK8S string = "/clusters"
)

// NewPermVerifyClient verify permission client
func NewPermVerifyClient(swi bool, iam iam.PermClient, clusterCli *cmanager.ClusterManagerClient) *PermVerifyClient {
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

// VerifyClusterPermission verify cluster permission
func (cli *PermVerifyClient) VerifyClusterPermission(user UserInfo, action string, resource ClusterResource) (bool, string) {
	if cli == nil {
		return false, ErrServerNotInited.Error()
	}

	switch resource.ClusterType {
	case K8s:
		// extract namespace
		requestInfo, err := getK8sRequestAPIInfo(action, resource.URL)
		if err != nil {
			blog.Infof("VerifyClusterPermission getK8sRequestAPIInfo failed: %v", err.Error())
			return false, fmt.Sprintf("VerifyClusterPermission getK8sRequestAPIInfo failed: %v", err.Error())
		}
		blog.V(4).Infof("PermVerifyClient VerifyClusterPermission getK8sRequestAPIInfo requestInfo %+v", requestInfo)

		// extract URL namespace
		if resource.Namespace == "" {
			resource.Namespace = requestInfo.Namespace
		}

		message := ""
		allowed, err := cli.verifyUserK8sClusterPermission(user.UserName, action, resource, requestInfo)
		blog.Infof("user %s access to type: %s, resource: %v, action: %s, permission: %t",
			user.UserName, resource.ClusterType, resource, action, allowed)

		if err != nil {
			message = err.Error()
		}

		// build operation log
		err = buildK8sOperationLog(user.UserName, resource, requestInfo, allowed)
		if err != nil {
			blog.Errorf("VerifyClusterPermission buildK8sOperationLog failed: %v", err.Error())
		}

		return allowed, message
	case Mesos:
		allowed, message := verifyResourceReplica(user.UserID, "cluster", resource.ClusterID, action)
		blog.Infof("user %s access to type: %s, action: %s, permission: %t",
			user.UserName, "cluster", action, allowed)

		return allowed, message
	default:
		return false, fmt.Sprintf("unsupported cluster_type %s", resource.ClusterType)
	}
}

func (cli *PermVerifyClient) verifyUserK8sClusterPermission(user, action string,
	resource ClusterResource, requestInfo *parser.RequestInfo) (bool, error) {
	if cli == nil {
		return false, ErrServerNotInited
	}

	// check resourceType
	verifyType, err := cli.checkResourceType(resource, requestInfo)
	if err != nil {
		return false, err
	}
	blog.V(4).Infof("PermVerifyClient verifyUserK8sClusterPermission verifyType[%s]", verifyType)

	switch verifyType {
	case clusterScopedType:
		clusterScopedAllow, err := cli.verifyUserClusterScopedPermission(user, action, resource)
		if err != nil {
			blog.Errorf("verifyUserClusterScopedPermission failed: %v", err)
			return false, err
		}
		if !clusterScopedAllow {
			permission := cli.applyForPermission(action, verifyType)
			return clusterScopedAllow, fmt.Errorf("verifyUserClusterScopedPermission failed: %v, please apply for permission %s",
				err, permission)
		}

		return clusterScopedAllow, nil
	case namespaceType:
		namespaceAllow, err := cli.verifyUserNamespacePermission(user, action, resource)
		if err != nil {
			blog.Errorf("verifyUserNamespacePermission failed: %v", err)
			return false, err
		}
		if !namespaceAllow {
			permission := cli.applyForPermission(action, verifyType)
			return namespaceAllow, fmt.Errorf("verifyUserNamespacePermission failed: %v, please apply for permission %s",
				err, permission)
		}

		return namespaceAllow, nil
	case namespaceScopedType:
		namespaceScopedAllow, err := cli.verifyUserNamespaceScopedPermission(user, action, resource)
		if err != nil {
			blog.Errorf("verifyUserNamespaceScopedPermission failed: %v", err)
			return false, err
		}
		if !namespaceScopedAllow {
			permission := cli.applyForPermission(action, verifyType)
			return namespaceScopedAllow, fmt.Errorf("verifyUserNamespaceScopedPermission failed: %v, please apply for permission %s",
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

func (cli *PermVerifyClient) verifyUserNamespaceScopedPermission(user string, action string, resource ClusterResource) (bool, error) {
	actionID := ""

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
	projectID, err := cli.getProjectIDByClusterID(resource.ClusterID)
	if err != nil {
		return false, fmt.Errorf("getProjectIDByClusterID[%s] failed", resource.ClusterID)
	}
	nameSpaceID, _ := utils.CalIAMNamespaceID(resource.ClusterID, resource.Namespace)

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
	blog.Infof("PermVerifyClient verifyUserNamespaceScopedPermission user[%s] actionID[%s] resource[%+v]",
		user, actionID, rn1)

	allow, err := cli.PermClient.IsAllowedWithResource(actionID, req, []iam.ResourceNode{rn1}, false)
	if err != nil {
		blog.Errorf("perm_client check namespaceScoped resource permission failed: %v", err)
		return false, err
	}
	return allow, nil
}

func (cli *PermVerifyClient) verifyUserNamespacePermission(user string, action string, resource ClusterResource) (bool, error) {

	var (
		actionID              = ""
		isClusterResourcePerm = false
	)

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

	if resource.ClusterID == "" {
		return false, fmt.Errorf("resource clusterID or namespace is null")
	}
	projectID, err := cli.getProjectIDByClusterID(resource.ClusterID)
	if err != nil {
		return false, fmt.Errorf("getProjectIDByClusterID[%s] failed", resource.ClusterID)
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
		nameSpaceID, _ := utils.CalIAMNamespaceID(resource.ClusterID, resource.Namespace)
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

	blog.Infof("PermVerifyClient verifyUserNamespacePermission user[%s] actionID[%s] resource[%+v]",
		user, actionID, rn1)

	allow, err := cli.PermClient.IsAllowedWithResource(actionID, req, []iam.ResourceNode{rn1}, false)
	if err != nil {
		blog.Errorf("perm_client check namespace permission failed: %v", err)
		return false, err
	}

	return allow, nil
}

func (cli *PermVerifyClient) getProjectIDByClusterID(clusterID string) (string, error) {
	projectID, err := cli.ClusterClient.GetProjectIDByClusterID(clusterID)
	if err != nil {
		blog.Infof("PermVerifyClient getProjectIDByClusterID[%s] failed: %v", clusterID, err)
		return "", err
	}

	return projectID, nil
}

func (cli *PermVerifyClient) verifyUserClusterScopedPermission(user string, action string, resource ClusterResource) (bool, error) {
	actionID := ""

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

	if resource.ClusterID == "" {
		return false, fmt.Errorf("resource clusterID is null")
	}
	projectID, err := cli.getProjectIDByClusterID(resource.ClusterID)
	if err != nil {
		return false, fmt.Errorf("getProjectIDByClusterID[%s] failed", resource.ClusterID)
	}
	blog.Infof("verifyUserClusterScopedPermission getProjectIDByClusterID[%s]: %s", resource.ClusterID, projectID)

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

	blog.Infof("PermVerifyClient verifyUserClusterScopedPermission user[%s] actionID[%s] resource[%+v]",
		user, actionID, rn1)
	allow, err := cli.PermClient.IsAllowedWithResource(actionID, req, []iam.ResourceNode{rn1}, false)
	if err != nil {
		blog.Errorf("perm_client check cluster permission failed: %v", err)
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

// trans bcs clusters API to k8s API url
func transToAPIServerURL(url string) string {
	if len(url) == 0 || !strings.HasPrefix(url, prefixAPIK8S) {
		return ""
	}

	urlStrs := strings.Split(url, "/")
	if len(urlStrs) <= 3 {
		return ""
	}

	return "/" + strings.Join(urlStrs[3:], "/")
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
		ClusterType: req.ResourceType,
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
