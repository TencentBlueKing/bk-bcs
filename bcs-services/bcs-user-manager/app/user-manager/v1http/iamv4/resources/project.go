/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
    10| * limitations under the License.
*/

package resources

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/component"
	blog "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/log"
	pkgutils "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/utils"
	iamv4 "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v1http/iamv4"
)

func listProjects(ctx context.Context, tenantID string, filter iamv4.Filter, page iamv4.Page) (int, []iamv4.Instance, error) {
	var params map[string]string
	if filter.Keyword != "" {
		params = map[string]string{"searchName": filter.Keyword}
	}
	result, err := component.QueryProjects(ctx, tenantID, page.PageSize, page.Page-1, params)
	if err != nil {
		return 0, nil, err
	}
	results := make([]iamv4.Instance, 0, len(result.Results))
	for _, r := range result.Results {
		results = append(results, iamv4.Instance{
			ID:          r.ProjectID,
			DisplayName: iamv4.CombineNameID(r.Name, r.GetProjectCode()),
		})
	}
	return result.Total, results, nil
}

func fetchProjects(ctx context.Context, filter iamv4.Filter) ([]iamv4.Instance, error) {
	results := make([]iamv4.Instance, 0, len(filter.IDs))
	for _, id := range filter.IDs {
		reqID := pkgutils.GetRequestIDFromContext(ctx)
		pctx := context.WithValue(ctx, pkgutils.ContextValueKeyRequestID, reqID)
		p, err := component.GetProject(pctx, id)
		if err != nil {
			blog.Log(ctx).Errorf("iamv4 get project failed")
			continue
		}
		results = append(results, iamv4.Instance{
			ID:          p.ProjectID,
			DisplayName: iamv4.CombineNameID(p.Name, p.GetProjectCode()),
			Approvers:   iamv4.SplitManagers(p.Managers),
		})
	}
	return results, nil
}
