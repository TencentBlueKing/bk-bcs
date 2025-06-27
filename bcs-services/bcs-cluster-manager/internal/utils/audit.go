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

package utils

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/audit"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	middleauth "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/server"
	grpcmeta "google.golang.org/grpc/metadata"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	raudit "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/audit"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

// resource 资源相关信息
type resource struct {
	BusinessID     string `json:"businessID" yaml:"businessID"`         // 业务ID
	ProjectID      string `json:"projectID" yaml:"projectID"`           // 项目ID
	ClusterID      string `json:"clusterID" yaml:"clusterID"`           // 集群ID
	NodeGroupID    string `json:"nodeGroupID" yaml:"nodeGroupID"`       // 节点组ID
	CloudID        string `json:"cloudID" yaml:"cloudID"`               // 云类型
	AccountName    string `json:"accountName" yaml:"accountName"`       // cloud账号名称
	AccountID      string `json:"accountID" yaml:"accountID"`           // cloud云账号ID
	AccountIDs     string `json:"accountIDs" yaml:"accountIDs"`         // cloud云账号ID
	NodeTemplateID string `json:"nodeTemplateID" yaml:"nodeTemplateID"` // 节点模板ID
	Name           string `json:"name" yaml:"name"`
	ClusterName    string `json:"clusterName" yaml:"clusterName"`
}

// resource to map
func (r resource) toMap() map[string]any {
	result := make(map[string]any, 0)
	if r.BusinessID != "" {
		result["BusinessID"] = r.BusinessID
	}
	if r.ProjectID != "" {
		result["ProjectID"] = r.ProjectID
	}
	if r.ClusterID != "" {
		result["ClusterID"] = r.ClusterID
	}
	if r.ClusterName != "" {
		result["ClusterName"] = r.ClusterName
	}
	if r.NodeGroupID != "" {
		result["NodeGroupID"] = r.NodeGroupID
	}
	if r.CloudID != "" {
		result["CloudID"] = r.CloudID
	}
	if r.AccountName != "" {
		result["AccountName"] = r.AccountName
	}
	if r.AccountID != "" {
		result["AccountID"] = r.AccountID
	}
	if r.AccountIDs != "" {
		result["AccountIDs"] = r.AccountIDs
	}
	if r.NodeTemplateID != "" {
		result["NodeTemplateID"] = r.NodeTemplateID
	}
	if r.Name != "" {
		result["Name"] = r.Name
	}

	return result
}

func getResourceID(req server.Request) resource {
	body := req.Body()
	b, _ := json.Marshal(body)

	resourceID := resource{}
	_ = json.Unmarshal(b, &resourceID)

	return resourceID
}

// nolint
var auditFuncMap = map[string]func(req server.Request, rsp interface{}) (audit.Resource, audit.Action){
	"ClusterManager.CreateCluster": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		// return the resource to be recorded.
		return audit.Resource{
			ResourceType: audit.ResourceTypeCluster, ResourceID: res.ClusterName, ResourceName: res.ClusterName,
			ResourceData: res.toMap(),
			ProjectCode:  res.ProjectID,
		}, audit.Action{ActionID: "cluster_create", ActivityType: audit.ActivityTypeCreate}
	},
	"ClusterManager.RetryCreateClusterTask": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		// return the resource to be recorded.
		return audit.Resource{
			ResourceType: audit.ResourceTypeCluster, ResourceID: res.ClusterID, ResourceName: res.ClusterID,
			ResourceData: res.toMap(),
			ProjectCode:  getProjectIDByCluster(res.ProjectID, res.ClusterID),
		}, audit.Action{ActionID: "retry_create_cluster", ActivityType: audit.ActivityTypeUpdate}
	},
	"ClusterManager.ImportCluster": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		// return the resource to be recorded.
		return audit.Resource{
			ResourceType: audit.ResourceTypeCluster, ResourceID: res.ClusterName, ResourceName: res.ClusterName,
			ResourceData: res.toMap(),
			ProjectCode:  res.ProjectID,
		}, audit.Action{ActionID: "cluster_create", ActivityType: audit.ActivityTypeCreate}
	},
	"ClusterManager.UpdateCluster": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		// return the resource to be recorded.
		return audit.Resource{
			ResourceType: audit.ResourceTypeCluster, ResourceID: res.ClusterID, ResourceName: res.ClusterID,
			ResourceData: res.toMap(),
			ProjectCode:  getProjectIDByCluster(res.ProjectID, res.ClusterID),
		}, audit.Action{ActionID: "cluster_manage", ActivityType: audit.ActivityTypeUpdate}
	},
	"ClusterManager.DeleteCluster": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		// return the resource to be recorded.
		return audit.Resource{
			ResourceType: audit.ResourceTypeCluster, ResourceID: res.ClusterID, ResourceName: res.ClusterID,
			ResourceData: res.toMap(),
			ProjectCode:  getProjectIDByCluster(res.ProjectID, res.ClusterID),
		}, audit.Action{ActionID: "cluster_delete", ActivityType: audit.ActivityTypeDelete}
	},
	"ClusterManager.AddNodesToCluster": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		// return the resource to be recorded.
		return audit.Resource{
			ResourceType: audit.ResourceTypeCluster, ResourceID: res.ClusterID, ResourceName: res.ClusterID,
			ResourceData: res.toMap(),
			ProjectCode:  getProjectIDByCluster(res.ProjectID, res.ClusterID),
		}, audit.Action{ActionID: "add_nodes_cluster", ActivityType: audit.ActivityTypeCreate}
	},
	"ClusterManager.DeleteNodesFromCluster": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		// return the resource to be recorded.
		return audit.Resource{
			ResourceType: audit.ResourceTypeCluster, ResourceID: res.ClusterID, ResourceName: res.ClusterID,
			ResourceData: res.toMap(),
			ProjectCode:  getProjectIDByCluster(res.ProjectID, res.ClusterID),
		}, audit.Action{ActionID: "delete_nodes_cluster", ActivityType: audit.ActivityTypeDelete}
	},
	"ClusterManager.BatchDeleteNodesFromCluster": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		// return the resource to be recorded.
		return audit.Resource{
			ResourceType: audit.ResourceTypeCluster, ResourceID: res.ClusterID, ResourceName: res.ClusterID,
			ResourceData: res.toMap(),
			ProjectCode:  getProjectIDByCluster(res.ProjectID, res.ClusterID),
		}, audit.Action{ActionID: "batch_delete_nodes_cluster", ActivityType: audit.ActivityTypeDelete}
	},
	"ClusterManager.CreateNodeGroup": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		// return the resource to be recorded.
		return audit.Resource{
			ResourceType: audit.ResourceTypeCluster, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
			ProjectCode:  getProjectIDByCluster(res.ProjectID, res.ClusterID),
		}, audit.Action{ActionID: "create_node_group", ActivityType: audit.ActivityTypeCreate}
	},
	"ClusterManager.UpdateNodeGroup": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		// return the resource to be recorded.
		return audit.Resource{
			ResourceType: audit.ResourceTypeCluster, ResourceID: res.NodeGroupID, ResourceName: res.NodeGroupID,
			ResourceData: res.toMap(),
			ProjectCode:  getProjectIDByNodeGroup(res.NodeGroupID),
		}, audit.Action{ActionID: "update_node_group", ActivityType: audit.ActivityTypeUpdate}
	},
	"ClusterManager.DeleteNodeGroup": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		// return the resource to be recorded.
		return audit.Resource{
			ResourceType: audit.ResourceTypeCluster, ResourceID: res.NodeGroupID, ResourceName: res.NodeGroupID,
			ResourceData: res.toMap(),
			ProjectCode:  getProjectIDByNodeGroup(res.NodeGroupID),
		}, audit.Action{ActionID: "delete_node_group", ActivityType: audit.ActivityTypeDelete}
	},
	"ClusterManager.RemoveNodesFromGroup": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		// return the resource to be recorded.
		return audit.Resource{
			ResourceType: audit.ResourceTypeCluster, ResourceID: res.NodeGroupID, ResourceName: res.NodeGroupID,
			ResourceData: res.toMap(),
			ProjectCode:  getProjectIDByNodeGroup(res.NodeGroupID),
		}, audit.Action{ActionID: "remove_nodes_group", ActivityType: audit.ActivityTypeDelete}
	},
	"ClusterManager.CleanNodesInGroup": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		// return the resource to be recorded.
		return audit.Resource{
			ResourceType: audit.ResourceTypeCluster, ResourceID: res.NodeGroupID, ResourceName: res.NodeGroupID,
			ResourceData: res.toMap(),
			ProjectCode:  getProjectIDByNodeGroup(res.NodeGroupID),
		}, audit.Action{ActionID: "clean_nodes_group", ActivityType: audit.ActivityTypeDelete}
	},
	"ClusterManager.CleanNodesInGroupV2": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		// return the resource to be recorded.
		return audit.Resource{
			ResourceType: audit.ResourceTypeCluster, ResourceID: res.NodeGroupID, ResourceName: res.NodeGroupID,
			ResourceData: res.toMap(),
			ProjectCode:  getProjectIDByNodeGroup(res.NodeGroupID),
		}, audit.Action{ActionID: "clean_nodes_group_v2", ActivityType: audit.ActivityTypeDelete}
	},
	"ClusterManager.EnableNodeGroupAutoScale": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		// return the resource to be recorded.
		return audit.Resource{
			ResourceType: audit.ResourceTypeCluster, ResourceID: res.NodeGroupID, ResourceName: res.NodeGroupID,
			ResourceData: res.toMap(),
			ProjectCode:  getProjectIDByNodeGroup(res.NodeGroupID),
		}, audit.Action{ActionID: "enable_node_group_auto_scale", ActivityType: audit.ActivityTypeUpdate}
	},
	"ClusterManager.DisableNodeGroupAutoScale": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		// return the resource to be recorded.
		return audit.Resource{
			ResourceType: audit.ResourceTypeCluster, ResourceID: res.NodeGroupID, ResourceName: res.NodeGroupID,
			ResourceData: res.toMap(),
			ProjectCode:  getProjectIDByNodeGroup(res.NodeGroupID),
		}, audit.Action{ActionID: "disable_node_group_auto_scale", ActivityType: audit.ActivityTypeUpdate}
	},
	"ClusterManager.CreateNodeTemplate": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		// return the resource to be recorded.
		return audit.Resource{
			ResourceType: audit.ResourceTypeCluster, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
			ProjectCode:  res.ProjectID,
		}, audit.Action{ActionID: "create_node_template", ActivityType: audit.ActivityTypeCreate}
	},
	"ClusterManager.UpdateNodeTemplate": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		// return the resource to be recorded.
		return audit.Resource{
			ResourceType: audit.ResourceTypeCluster, ResourceID: res.NodeTemplateID, ResourceName: res.NodeTemplateID,
			ResourceData: res.toMap(),
			ProjectCode:  res.ProjectID,
		}, audit.Action{ActionID: "update_node_template", ActivityType: audit.ActivityTypeUpdate}
	},
	"ClusterManager.DeleteNodeTemplate": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		// return the resource to be recorded.
		return audit.Resource{
			ResourceType: audit.ResourceTypeCluster, ResourceID: res.NodeTemplateID, ResourceName: res.NodeTemplateID,
			ResourceData: res.toMap(),
			ProjectCode:  res.ProjectID,
		}, audit.Action{ActionID: "delete_node_template", ActivityType: audit.ActivityTypeDelete}
	},
	"ClusterManager.CreateCloudAccount": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		// return the resource to be recorded.
		return audit.Resource{
			ResourceType: audit.ResourceTypeCluster, ResourceID: res.AccountName, ResourceName: res.AccountName,
			ResourceData: res.toMap(),
			ProjectCode:  res.ProjectID,
		}, audit.Action{ActionID: "cloud_account_create", ActivityType: audit.ActivityTypeCreate}
	},
	"ClusterManager.UpdateCloudAccount": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		// return the resource to be recorded.
		return audit.Resource{
			ResourceType: audit.ResourceTypeCluster, ResourceID: res.AccountID, ResourceName: res.AccountID,
			ResourceData: res.toMap(),
			ProjectCode:  getProjectIDByCloudAccount(res.ProjectID, res.CloudID, res.AccountID),
		}, audit.Action{ActionID: "cloud_account_manage", ActivityType: audit.ActivityTypeUpdate}
	},
	"ClusterManager.DeleteCloudAccount": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		// return the resource to be recorded.
		return audit.Resource{
			ResourceType: audit.ResourceTypeCluster, ResourceID: res.AccountID, ResourceName: res.AccountID,
			ResourceData: res.toMap(),
			ProjectCode:  getProjectIDByCloudAccount(res.ProjectID, res.CloudID, res.AccountID),
		}, audit.Action{ActionID: "cloud_account_manage", ActivityType: audit.ActivityTypeDelete}
	},
}

// addAudit 添加审计
func addAudit(ctx context.Context, req server.Request, rsp interface{}, startTime, endTime time.Time) {
	// get method audit func
	fn, ok := auditFuncMap[req.Method()]
	if !ok {
		return
	}

	res, act := fn(req, rsp)

	auditCtx := audit.RecorderContext{
		Username:  getUserFromCtx(ctx),
		SourceIP:  getSourceIPFromCtx(ctx),
		UserAgent: GetUserAgentFromCtx(ctx),
		RequestID: requestIDFromContext(ctx),
		StartTime: startTime,
		EndTime:   endTime,
	}
	resource := audit.Resource{
		ProjectCode:  res.ProjectCode,
		ResourceType: res.ResourceType,
		ResourceID:   res.ResourceID,
		ResourceName: res.ResourceName,
		ResourceData: res.ResourceData,
	}
	action := audit.Action{
		ActionID:     act.ActionID,
		ActivityType: act.ActivityType,
	}

	result := audit.ActionResult{
		Status: audit.ActivityStatusSuccess,
	}

	// get handle result
	v := reflect.ValueOf(rsp)
	codeField := v.Elem().FieldByName("Code")
	messageField := v.Elem().FieldByName("Message")
	if codeField.CanInterface() {
		code := int(codeField.Interface().(uint32))
		result.ResultCode = code
	}
	if messageField.CanInterface() {
		message := messageField.Interface().(string)
		result.ResultContent = message
	}
	if result.ResultCode != types.BcsErrClusterManagerSuccess {
		result.Status = audit.ActivityStatusFailed
	}
	_ = raudit.GetAuditClient().R().
		SetContext(auditCtx).SetResource(resource).SetAction(action).SetResult(result).Do()
}

// GetUserAgentFromCtx 通过 ctx 获取 userAgent
func GetUserAgentFromCtx(ctx context.Context) string {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return ""
	}
	forwarded, _ := md.Get(common.UserAgentHeaderKey)
	return forwarded
}

// getSourceIPFromCtx 通过 ctx 获取 sourceIP
func getSourceIPFromCtx(ctx context.Context) string {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return ""
	}
	forwarded, _ := md.Get(common.ForwardedForHeaderKey)
	return forwarded
}

func requestIDFromContext(ctx context.Context) string {
	meta, ok := grpcmeta.FromIncomingContext(ctx)
	if !ok {
		blog.Warnf("get grpc metadata from context failed")
		return ""
	}
	requestIDStrs := meta.Get("X-Request-Id")
	if len(requestIDStrs) == 0 {
		return ""
	}
	return requestIDStrs[0]
}

// 通过cloudID, accountID 获取 projectID
func getProjectIDByCloudAccount(projectID, cloudID, accountID string) string {
	if projectID != "" {
		return projectID
	}
	account, err := store.GetStoreModel().GetCloudAccount(context.TODO(), cloudID, accountID, false)
	if err != nil || errors.Is(err, drivers.ErrTableRecordNotFound) {
		blog.Errorf("Get CloudAccount %s:%s in checking failed, err %s", cloudID, accountID, err.Error())
		return ""
	}
	return account.ProjectID
}

// 通过clusterID 获取 projectID
func getProjectIDByCluster(projectID, clusterID string) string {
	if projectID != "" {
		return projectID
	}
	cluster, err := store.GetStoreModel().GetCluster(context.TODO(), clusterID)
	if err != nil || errors.Is(err, drivers.ErrTableRecordNotFound) {
		blog.Errorf("get cluster %s failed, %s", clusterID, err.Error())
		return ""
	}
	return cluster.ProjectID
}

// 通过nodeGroupID 获取 projectID
func getProjectIDByNodeGroup(nodeGroupID string) string {
	np, err := store.GetStoreModel().GetNodeGroup(context.TODO(), nodeGroupID)
	if err != nil || errors.Is(err, drivers.ErrTableRecordNotFound) {
		blog.Errorf("get NodeGroup %s failed, %s", nodeGroupID, err.Error())
		return ""
	}
	return np.ProjectID
}

// getUserFromCtx 通过 ctx 获取当前用户
func getUserFromCtx(ctx context.Context) string {
	authUser, _ := middleauth.GetUserFromContext(ctx)
	return authUser.GetUsername()
}
