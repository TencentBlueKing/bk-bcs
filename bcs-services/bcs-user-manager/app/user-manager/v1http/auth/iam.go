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
	"encoding/json"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cloudaccount"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/templateset"
	authutil "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	iamsdk "github.com/TencentBlueKing/iam-go-sdk"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
)

// PermCtx perm context
type PermCtx struct {
	ResourceType string      `json:"resource_type"`
	ProjectID    string      `json:"project_id"`
	ClusterID    string      `json:"cluster_id"`
	Namespace    string      `json:"name"`
	TemplateID   json.Number `json:"template_id"`
	AccountID    string      `json:"account_id"`
}

// GetResourceNodeFromPermCtx 根据 resource type 拼装 iam.ResourceNode
func GetResourceNodeFromPermCtx(permCtx *PermCtx) iam.ResourceNode {
	node := iam.ResourceNode{System: config.GetGlobalConfig().IAMConfig.SystemID, RType: permCtx.ResourceType}
	switch permCtx.ResourceType {
	case string(project.SysProject):
		node.RInstance = permCtx.ProjectID
		node.Rp = project.ProjectResourcePath{}
	case string(cluster.SysCluster):
		node.RInstance = permCtx.ClusterID
		node.Rp = cluster.ClusterResourcePath{ProjectID: permCtx.ProjectID}
	case string(namespace.SysNamespace):
		node.RInstance = authutil.CalcIAMNsID(permCtx.ClusterID, permCtx.Namespace)
		node.Rp = namespace.NamespaceResourcePath{ProjectID: permCtx.ProjectID, ClusterID: permCtx.ClusterID}
	case string(templateset.SysTemplateSet):
		node.RInstance = permCtx.TemplateID.String()
		node.Rp = templateset.TemplateSetResourcePath{ProjectID: permCtx.ProjectID}
	case string(cloudaccount.SysCloudAccount):
		node.RInstance = permCtx.AccountID
		node.Rp = cloudaccount.AccountResourcePath{ProjectID: permCtx.ProjectID}
	}
	return node
}

// GetApplicationsFromPermCtx 根据 resource type 拼装 iam.ApplicationAction
func GetApplicationsFromPermCtx(permCtx *PermCtx, actionsID string) []iam.ApplicationAction {
	apps := make([]iam.ApplicationAction, 0)
	if permCtx == nil {
		return []iam.ApplicationAction{{ActionID: actionsID,
			RelatedResources: []iamsdk.ApplicationRelatedResourceType{}}}
	}
	switch permCtx.ResourceType {
	case string(project.SysProject):
		apps = project.BuildProjectSameInstanceApplication(false, []string{actionsID}, []string{permCtx.ProjectID})
	case string(cluster.SysCluster):
		apps = cluster.BuildClusterSameInstanceApplication(false, []string{actionsID}, []cluster.ProjectClusterData{
			{
				Project: permCtx.ProjectID,
				Cluster: permCtx.ClusterID,
			},
		})
	case string(namespace.SysNamespace):
		apps = append(apps, namespace.BuildNamespaceApplicationInstance(namespace.NamespaceApplicationAction{
			ActionID: actionsID,
			Data: []namespace.ProjectNamespaceData{{
				Project:   permCtx.ProjectID,
				Cluster:   permCtx.ClusterID,
				Namespace: authutil.CalcIAMNsID(permCtx.ClusterID, permCtx.Namespace),
			}},
		}))
	case string(templateset.SysTemplateSet):
		instances := make([][]iam.Instance, 0)
		instances = append(instances, []iam.Instance{
			{
				ResourceType: string(project.SysProject),
				ResourceID:   permCtx.ProjectID,
			},
			{
				ResourceType: string(templateset.SysTemplateSet),
				ResourceID:   permCtx.TemplateID.String(),
			},
		})
		rr := make([]iamsdk.ApplicationRelatedResourceType, 0)
		rr = append(rr, authutil.BuildRelatedSystemResource(iam.SystemIDBKBCS, permCtx.ResourceType, instances))
		apps = []iam.ApplicationAction{
			{
				ActionID:         actionsID,
				RelatedResources: rr,
			},
		}
	case string(cloudaccount.SysCloudAccount):
		apps = cloudaccount.BuildAccountSameInstanceApplication(false, []string{actionsID},
			[]cloudaccount.ProjectAccountData{
				{
					Project: permCtx.ProjectID,
					Account: permCtx.AccountID,
				},
			})
	default:
		apps = []iam.ApplicationAction{{ActionID: actionsID,
			RelatedResources: []iamsdk.ApplicationRelatedResourceType{}}}
	}
	return apps
}

// GetApplyURL get apply url
func GetApplyURL(applications []iam.ApplicationAction, tenantID string) (string, error) {
	url, err := config.GloablIAMClient(tenantID).GetApplyURL(iam.ApplicationRequest{
		SystemID: config.GetGlobalConfig().IAMConfig.SystemID}, applications, iam.BkUser{
		BkUserName: iam.SystemUser,
	})
	if err != nil {
		return iam.IamAppURL, err
	}
	return url, nil
}

// GetResourceTypeFromAction get resource type from action
// NOCC:CCN_threshold(工具误报:),golint/fnsize(设计如此:)
func GetResourceTypeFromAction(action string) string { // nolint
	switch action {
	case project.ProjectCreate.String():
		return ""
	case project.ProjectView.String():
		return string(project.SysProject)
	case project.ProjectEdit.String():
		return string(project.SysProject)
	case project.ProjectDelete.String():
		return string(project.SysProject)
	case cluster.ClusterCreate.String():
		return string(project.SysProject)
	case cluster.ClusterView.String():
		return string(cluster.SysCluster)
	case cluster.ClusterManage.String():
		return string(cluster.SysCluster)
	case cluster.ClusterDelete.String():
		return string(cluster.SysCluster)
	case cluster.ClusterUse.String():
		return string(cluster.SysCluster)
	case namespace.NameSpaceCreate.String():
		return string(cluster.SysCluster)
	case namespace.NameSpaceView.String():
		return string(namespace.SysNamespace)
	case namespace.NameSpaceUpdate.String():
		return string(namespace.SysNamespace)
	case namespace.NameSpaceDelete.String():
		return string(namespace.SysNamespace)
	case namespace.NameSpaceList.String():
		return string(cluster.SysCluster)
	case cluster.ClusterScopedCreate.String():
		return string(cluster.SysCluster)
	case cluster.ClusterScopedView.String():
		return string(cluster.SysCluster)
	case cluster.ClusterScopedUpdate.String():
		return string(cluster.SysCluster)
	case cluster.ClusterScopedDelete.String():
		return string(cluster.SysCluster)
	case namespace.NameSpaceScopedCreate.String():
		return string(namespace.SysNamespace)
	case namespace.NameSpaceScopedView.String():
		return string(namespace.SysNamespace)
	case namespace.NameSpaceScopedUpdate.String():
		return string(namespace.SysNamespace)
	case namespace.NameSpaceScopedDelete.String():
		return string(namespace.SysNamespace)
	case templateset.TemplateSetCreate.String():
		return string(project.SysProject)
	case templateset.TemplateSetView.String():
		return string(templateset.SysTemplateSet)
	case templateset.TemplateSetCopy.String():
		return string(templateset.SysTemplateSet)
	case templateset.TemplateSetUpdate.String():
		return string(templateset.SysTemplateSet)
	case templateset.TemplateSetDelete.String():
		return string(templateset.SysTemplateSet)
	case templateset.TemplateSetInstantiate.String():
		return string(templateset.SysTemplateSet)
	case cloudaccount.AccountCreate.String():
		return string(project.SysProject)
	case cloudaccount.AccountManage.String():
		return string(cloudaccount.SysCloudAccount)
	case cloudaccount.AccountUse.String():
		return string(cloudaccount.SysCloudAccount)
	default:
		return ""
	}
}
