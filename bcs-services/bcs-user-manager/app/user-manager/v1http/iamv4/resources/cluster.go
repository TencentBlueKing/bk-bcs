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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/component"
	iamv4 "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v1http/iamv4"
)

func listClusters(ctx context.Context, filter iamv4.Filter, page iamv4.Page) (int, []iamv4.Instance, error) {
	_, parentID := iamv4.ResolveParent(filter)
	if parentID == "" {
		return 0, nil, iamv4.InvalidRequest("parent is required")
	}
	clusters, err := component.GetClustersByProjectID(ctx, parentID)
	if err != nil {
		return 0, nil, err
	}
	items := make([]iamv4.Instance, 0, len(clusters))
	for _, r := range clusters {
		if r == nil {
			continue
		}
		display := iamv4.CombineNameID(r.ClusterName, r.ClusterID)
		if !iamv4.MatchKeyword(r.ClusterID, display, filter.Keyword) {
			continue
		}
		items = append(items, iamv4.Instance{
			ID:          r.ClusterID,
			DisplayName: display,
		})
	}
	total, results := iamv4.Paginate(items, page.Page, page.PageSize)
	return total, results, nil
}

func fetchClusters(filter iamv4.Filter) ([]iamv4.Instance, error) {
	clusterMap, err := component.GetClusterMap()
	if err != nil {
		return nil, err
	}
	results := make([]iamv4.Instance, 0, len(filter.IDs))
	for _, id := range filter.IDs {
		cls, ok := clusterMap[id]
		if !ok || cls == nil {
			continue
		}
		results = append(results, iamv4.Instance{
			ID:          cls.ClusterID,
			DisplayName: iamv4.CombineNameID(cls.ClusterName, cls.ClusterID),
			IAMPath:     iamv4.ProjectPath(cls.ProjectID),
			Approvers:   iamv4.UniqueApprovers(cls.Creator, cls.Updater),
		})
	}
	return results, nil
}
