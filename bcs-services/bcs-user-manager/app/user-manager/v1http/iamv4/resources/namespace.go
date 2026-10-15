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

package resources

import (
	"context"

	authUtils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/component"
	blog "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/log"
	iamv4 "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v1http/iamv4"
)

func listNamespaces(ctx context.Context, filter iamv4.Filter, page iamv4.Page) (int, []iamv4.Instance, error) {
	projectID, clusterID, err := resolveNamespaceParent(ctx, filter)
	if err != nil {
		return 0, nil, err
	}
	project, err := component.GetProjectWithCache(ctx, projectID)
	if err != nil {
		return 0, nil, err
	}
	nss, err := component.GetClusterNamespaces(ctx, project.ProjectCode, clusterID)
	if err != nil {
		return 0, nil, err
	}
	items := make([]iamv4.Instance, 0, len(nss))
	for _, r := range nss {
		if !iamv4.MatchKeyword(r.Name, r.Name, filter.Keyword) {
			continue
		}
		items = append(items, iamv4.Instance{
			ID:          authUtils.CalcIAMNsID(clusterID, r.Name),
			DisplayName: r.Name,
		})
	}
	total, results := iamv4.Paginate(items, page.Page, page.PageSize)
	return total, results, nil
}

func fetchNamespaces(ctx context.Context, filter iamv4.Filter) ([]iamv4.Instance, error) {
	results := make([]iamv4.Instance, 0, len(filter.IDs))
	for _, nsID := range filter.IDs {
		clusterID, err := iamv4.ParseNSID(nsID)
		if err != nil {
			blog.Log(ctx).Errorf("iamv4 invalid namespace id")
			continue
		}
		ns, err := component.GetCachedNamespace(ctx, clusterID, nsID)
		if err != nil {
			blog.Log(ctx).Errorf("iamv4 get namespace failed")
			continue
		}
		cls, err := component.GetClusterByClusterID(ctx, clusterID)
		if err != nil {
			blog.Log(ctx).Errorf("iamv4 get cluster for namespace failed")
			continue
		}
		results = append(results, iamv4.Instance{
			ID:          nsID,
			DisplayName: ns.Name,
			IAMPath:     iamv4.NamespacePath(cls.ProjectID, clusterID),
			Approvers:   iamv4.UniqueApprovers(ns.Managers...),
		})
	}
	return results, nil
}

func resolveNamespaceParent(ctx context.Context, filter iamv4.Filter) (projectID, clusterID string, err error) {
	projectID = iamv4.AncestorID(filter, iamv4.Project)
	parentType, parentID := iamv4.ResolveParent(filter)
	if parentID == "" {
		return "", "", iamv4.InvalidRequest("parent is required")
	}
	if parentType != "" && parentType != iamv4.Cluster {
		return "", "", iamv4.InvalidRequest("parent is required")
	}
	clusterID = parentID
	if projectID == "" {
		cls, e := component.GetClusterByClusterID(ctx, clusterID)
		if e != nil {
			return "", "", e
		}
		projectID = cls.ProjectID
	}
	if projectID == "" {
		return "", "", iamv4.InvalidRequest("parent is required")
	}
	return projectID, clusterID, nil
}
