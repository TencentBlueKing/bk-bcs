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

package route

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/iam"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/sessions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
)

// projectCodeAnnoKey 共享集群中命名空间所属项目的注解
const projectCodeAnnoKey = "io.tencent.bcs.projectcode"

// PermissionRequired 权限控制，必须都为真才可以
func PermissionRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		authCtx := MustGetAuthContext(c)

		// 校验项目，集群信息的正确性
		if err := ValidateProjectCluster(c, authCtx); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, types.APIResponse{
				Code:      types.ApiErrorCode,
				Message:   err.Error(),
				RequestID: authCtx.RequestId,
			})
			return
		}

		c.Set("auth_context", authCtx)

		// 管理员不校验权限, 包含管理员凭证
		if config.G.IsManager(authCtx.Username, authCtx.ClusterId) {
			c.Next()
			return
		}

		if err := initContextWithIAMProject(c, authCtx); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, types.APIResponse{
				Code:      types.ApiErrorCode,
				Message:   err.Error(),
				RequestID: authCtx.RequestId,
			})
			return
		}

		c.Next()
	}
}

// ProjectViewRequired 项目查看权限, 用于项目维度的接口
// 只校验当前项目, 不按集群过滤, 共享集群由 cluster-manager 随项目集群一并返回
func ProjectViewRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		authCtx := MustGetAuthContext(c)

		if err := validateProject(c, authCtx); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, types.APIResponse{
				Code:      types.ApiErrorCode,
				Message:   err.Error(),
				RequestID: authCtx.RequestId,
			})
			return
		}

		c.Set("auth_context", authCtx)

		if err := checkProjectView(c, authCtx); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, types.APIResponse{
				Code:      types.ApiErrorCode,
				Message:   err.Error(),
				RequestID: authCtx.RequestId,
			})
			return
		}

		c.Next()
	}
}

// checkProjectView 校验用户是否有项目查看权限
func checkProjectView(c *gin.Context, authCtx *AuthContext) error {
	if config.G.IsManager(authCtx.Username, "") || config.G.Auth.Disable {
		return nil
	}

	// portal session 等非用户态请求没有用户名
	if authCtx.Username == "" {
		return errors.New(i18n.T(c, "没有权限"))
	}

	// 未绑定集群, 只校验项目查看权限
	return checkIAMPerm(c, authCtx, "")
}

// SessionRequired session 权限校验
func SessionRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		authCtx := MustGetAuthContext(c)

		podCtx, err := validateSession(c, authCtx)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, types.APIResponse{
				Code:      types.ApiErrorCode,
				Message:   err.Error(),
				RequestID: authCtx.RequestId,
			})
			return
		}

		authCtx.BindSession = podCtx
		c.Set("auth_context", authCtx)

		c.Next()
	}
}

// 校验用户session权限
func validateSession(c *gin.Context, authCtx *AuthContext) (*types.PodContext, error) {
	sessionId := GetSessionId(c)
	if sessionId == "" {
		return nil, errors.New("session_id is required")
	}

	podCtx, err := sessions.NewStore().OpenAPIScope().Get(c.Request.Context(), sessionId)
	if err != nil {
		return nil, errors.New("session已经过期或不合法")
	}

	if config.G.IsManager(authCtx.Username, podCtx.ClusterId) {
		return podCtx, nil
	}

	if podCtx.HasPerm(authCtx.Username) {
		return podCtx, nil
	}

	return nil, errors.New("用户无权限登入此session")
}

// validateProject 校验项目信息的正确性
func validateProject(c *gin.Context, authCtx *AuthContext) error {
	projectId := GetProjectIdOrCode(c)
	if projectId == "" {
		return errors.New("project_id or code is required")
	}

	project, err := bcs.GetProject(c.Request.Context(), config.G.BCS, projectId)
	if err != nil {
		return errors.Wrap(err, i18n.T(c, "项目不正确"))
	}

	authCtx.BindProject = project
	authCtx.ProjectId = project.ProjectId
	authCtx.ProjectCode = project.Code
	authCtx.TenantId = project.TenantID

	return nil
}

// ValidateProjectCluster xxx
func ValidateProjectCluster(c *gin.Context, authCtx *AuthContext) error {
	if GetProjectIdOrCode(c) == "" {
		return errors.New("project_id or code is required")
	}

	clusterId := GetClusterId(c)
	if clusterId == "" {
		return errors.New("clusterId required")
	}

	if err := validateProject(c, authCtx); err != nil {
		return err
	}

	cluster, err := bcs.GetCluster(c.Request.Context(), authCtx.ProjectId, clusterId)
	if err != nil {
		return errors.Wrap(err, i18n.T(c, "项目或者集群Id不正确"))
	}

	authCtx.BindCluster = cluster
	authCtx.ClusterId = cluster.ClusterId

	return nil
}

// initContextWithIAMProject Dev环境, 可以设置环境变量
func initContextWithIAMProject(c *gin.Context, authCtx *AuthContext) error {
	if config.G.Auth.Disable {
		return nil
	}

	return checkIAMPerm(c, authCtx, GetNamespace(c))
}

func checkIAMPerm(c *gin.Context, authCtx *AuthContext, namespace string) error {
	allow, err := iam.IsAllowedWithResource(c.Request.Context(), authCtx.ProjectId, authCtx.ClusterId, namespace,
		authCtx.Username, authCtx.TenantId)
	if err != nil {
		return err
	}
	if !allow {
		return errors.New(i18n.T(c, "没有权限"))
	}

	return nil
}

// ValidateContainerNamespace 直连容器模式下, 按容器实际所在的命名空间校验用户权限
// PermissionRequired 校验的是请求参数中的 namespace, container_id 模式下与容器实际所在的命名空间可能不一致
func ValidateContainerNamespace(c *gin.Context, authCtx *AuthContext, namespace string) error {
	if namespace == "" {
		return errors.New(i18n.T(c, "没有权限"))
	}

	if config.G.IsManager(authCtx.Username, authCtx.ClusterId) || config.G.Auth.Disable {
		return nil
	}

	if err := checkSharedClusterNamespace(c, authCtx, namespace); err != nil {
		return err
	}

	return checkIAMPerm(c, authCtx, namespace)
}

// checkSharedClusterNamespace 共享集群非管理端项目, 只能访问属于本项目的命名空间
func checkSharedClusterNamespace(c *gin.Context, authCtx *AuthContext, namespace string) error {
	if !authCtx.BindCluster.IsShared || authCtx.BindCluster.ProjectId == authCtx.ProjectId {
		return nil
	}

	k8sClient, err := k8sclient.GetK8SClientByClusterId(authCtx.ClusterId)
	if err != nil {
		return err
	}
	ns, err := k8sClient.CoreV1().Namespaces().Get(c.Request.Context(), namespace, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if ns.Annotations[projectCodeAnnoKey] != authCtx.ProjectCode {
		return errors.New(i18n.T(c, "命名空间 %s 在该共享集群中不属于当前项目", namespace))
	}

	return nil
}

// CredentialRequired xxx
func CredentialRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		authCtx := MustGetAuthContext(c)

		if err := ValidateProjectCluster(c, authCtx); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, types.APIResponse{
				Code:      types.ApiErrorCode,
				Message:   err.Error(),
				RequestID: authCtx.RequestId,
			})
			return
		}

		c.Set("auth_context", authCtx)

		bkAppCode := authCtx.BKAppCode()

		if bkAppCode == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, types.APIResponse{
				Code:      types.ApiErrorCode,
				Message:   "not valid bk apigw request",
				RequestID: authCtx.RequestId,
			})
			return
		}

		// namespace 模式
		namespace := getNamespace(getRequestBody(c.Request))

		// 校验项目 and 集群 and 命名空间权限，
		// 其中共享集群命名空间校验，只支持 namespace/pod_name/container_name 模式，container_id 模式在共享集群检索效率较低，暂不支持
		if err := checkAppCredScope(authCtx, bkAppCode, namespace); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, types.APIResponse{
				Code:      types.ApiErrorCode,
				Message:   err.Error(),
				RequestID: authCtx.RequestId,
			})
			return
		}

		c.Next()
	}
}

// ValidateCredContainerNamespace 直连容器模式下, 按容器实际所在的命名空间校验 App 凭证
// CredentialRequired 校验的是请求体中的 namespace, container_id 模式下与容器实际所在的命名空间可能不一致
func ValidateCredContainerNamespace(authCtx *AuthContext, namespace string) error {
	if namespace == "" {
		return errors.New("namespace is required")
	}
	return checkAppCredScope(authCtx, authCtx.BKAppCode(), namespace)
}

func checkAppCredScope(authCtx *AuthContext, bkAppCode, namespace string) error {
	scopeValues := map[config.ScopeType]string{
		config.ScopeProjectCode: authCtx.ProjectCode,
		config.ScopeClusterId:   authCtx.ClusterId,
		config.ScopeNamespace:   namespace,
	}
	if config.G.ValidateCred(config.CredentialAppCode, bkAppCode, scopeValues) {
		return nil
	}

	return fmt.Errorf(
		"app %s have no permission, %s, %s, namespace<%s>", bkAppCode, authCtx.BindProject, authCtx.BindCluster, namespace)
}

// ManagersRequired 校验用户是否为manager
func ManagersRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authCtx := MustGetAuthContext(c)
		user := authCtx.Username
		managers := config.G.Base.Managers
		for _, manager := range managers {
			if user == manager {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, types.APIResponse{
			Code:      types.ApiErrorCode,
			Message:   fmt.Sprintf("%s is not managers, has no permission", user),
			RequestID: authCtx.RequestId,
		})
	}
}

// 获取请求体
func getRequestBody(r *http.Request) []byte {
	// 读取请求体
	body, _ := io.ReadAll(r.Body)
	// 恢复请求体
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	return body
}

// 其他内容忽略
type namespaceOnly struct {
	Namespace string `json:"namespace"`
}

// 获取body里面的namespace
func getNamespace(b []byte) string {
	namespace := namespaceOnly{}
	_ = json.Unmarshal(b, &namespace)
	return namespace.Namespace
}
