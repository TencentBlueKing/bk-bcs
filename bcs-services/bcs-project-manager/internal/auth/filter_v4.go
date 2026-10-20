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
	"fmt"
	"strings"
	"sync"
	"time"

	bcsIAM "github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	iamv4 "github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iamv4"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/page"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
)

const (
	// unrestrictedResourceID IAM V4 用资源 ID "*" 表示对该资源类型全部实例的无限制授权
	unrestrictedResourceID = "*"
	v4AuthBatchSize        = 20
	v4SubjectPageSize      = 100
	maxSubjectPages        = 100
	v4AuthConcurrency      = 4
	v4ListTimeout          = 60 * time.Second
)

// projectViewRoles 角色模型里包含 project_view 的角色。
// 未覆盖的新角色不会走无限制短路，仍会落到逐项目鉴权。
var projectViewRoles = []string{
	iamv4.RoleReadOnly.String(),
	iamv4.RoleDeveloper.String(),
	iamv4.RoleOperator.String(),
}

// projectIDLister 拉取待鉴权的项目 ID。测试可替换。
var projectIDLister = listTenantProjectIDs

// listAuthorizedProjectIDsV4 用 IAM V4 计算用户可查看的项目。
// V4 开放接口没有 V3 PolicyQuery。无限制授权只认授权主体为当前用户、资源 ID 为 "*"；
// 其余授权对本地项目做 project_view 批量鉴权。
func listAuthorizedProjectIDsV4(ctx context.Context, username, tenantID string) ([]string, bool, error) {
	if strings.TrimSpace(username) == "" {
		return nil, false, fmt.Errorf("username is required")
	}
	cli, err := newProjectIAMV4Client(tenantID)
	if err != nil {
		return nil, false, err
	}
	ctx, cancel := withIAMTimeout(ctx)
	defer cancel()

	unrestricted, err := hasUnrestrictedProjectView(ctx, cli, username)
	if err != nil {
		return nil, false, err
	}
	if unrestricted {
		return nil, true, nil
	}

	candidates, err := projectIDLister(ctx, tenantID)
	if err != nil {
		return nil, false, fmt.Errorf("list projects for iam v4 auth: %w", err)
	}
	allowed, err := filterProjectView(ctx, cli, username, candidates)
	if err != nil {
		return nil, false, err
	}
	return allowed, false, nil
}

func newProjectIAMV4Client(tenantID string) (*iamv4.Client, error) {
	if config.GlobalConf == nil {
		return nil, fmt.Errorf("iam config is not initialized")
	}
	host := strings.TrimSpace(config.GlobalConf.IAM.V4GatewayHost)
	if host == "" {
		return nil, fmt.Errorf("iam v4 gateway host is required")
	}
	return iamv4.NewClient(&iamv4.Options{
		SystemID:    bcsIAM.SystemIDBKBCS,
		AppCode:     config.GlobalConf.App.Code,
		AppSecret:   config.GlobalConf.App.Secret,
		GateWayHost: host,
		TenantID:    tenantID,
	})
}

func withIAMTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, v4ListTimeout)
}

func hasUnrestrictedProjectView(ctx context.Context, cli *iamv4.Client, username string) (bool, error) {
	for _, roleID := range projectViewRoles {
		ok, err := roleGrantsAllProjects(ctx, cli, roleID, username)
		if err != nil {
			return false, fmt.Errorf("query unrestricted project view: %w", err)
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

func roleGrantsAllProjects(ctx context.Context, cli *iamv4.Client, roleID, username string) (bool, error) {
	for pageNo := 1; pageNo <= maxSubjectPages; pageNo++ {
		data, err := cli.ListAuthorizationSubject(ctx, bcsIAM.SystemIDBKBCS, iamv4.ListAuthorizationSubjectRequest{
			RoleID:                roleID,
			RelatedResourceTypeID: iamv4.SysProject.String(),
			Resource: &iamv4.ResourceRef{
				Type: iamv4.SysProject.String(),
				ID:   unrestrictedResourceID,
			},
			Page:     pageNo,
			PageSize: v4SubjectPageSize,
		})
		if err != nil {
			return false, err
		}
		if data == nil || len(data.Results) == 0 {
			return false, nil
		}
		for _, item := range data.Results {
			if item.Subject.Type == iamv4.SubjectUser && item.Subject.ID == username {
				return true, nil
			}
		}
		if len(data.Results) < v4SubjectPageSize || pageNo*v4SubjectPageSize >= data.Count {
			return false, nil
		}
	}
	return false, nil
}

func filterProjectView(ctx context.Context, cli *iamv4.Client, username string, ids []string) ([]string, error) {
	ids = compactIDs(ids)
	if len(ids) == 0 {
		return nil, nil
	}
	allowed, err := authProjectView(ctx, cli, username, ids)
	if err != nil {
		return nil, fmt.Errorf("auth project view: %w", err)
	}
	out := make([]string, 0, len(allowed))
	for _, id := range ids {
		if allowed[id] {
			out = append(out, id)
		}
	}
	return out, nil
}

func authProjectView(ctx context.Context, cli *iamv4.Client, username string, ids []string) (map[string]bool, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	out := make(map[string]bool, len(ids))
	var mu sync.Mutex
	var first error
	sem := make(chan struct{}, v4AuthConcurrency)
	var wg sync.WaitGroup
	for _, chunk := range chunkStrings(ids, v4AuthBatchSize) {
		chunk := chunk
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				mu.Lock()
				if first == nil {
					first = ctx.Err()
				}
				mu.Unlock()
				return
			}
			resources := make([]iamv4.AuthResource, len(chunk))
			for i, id := range chunk {
				resources[i] = iamv4.AuthResource{ID: id}
			}
			results, err := cli.DirectAuthByResources(ctx, bcsIAM.SystemIDBKBCS, iamv4.AuthByResourcesRequest{
				Subject:   iamv4.Subject{Type: iamv4.SubjectUser, ID: username},
				ActionID:  ProjectView,
				Resources: resources,
			})
			if err != nil {
				mu.Lock()
				if first == nil {
					first = err
					cancel()
				}
				mu.Unlock()
				return
			}
			mu.Lock()
			defer mu.Unlock()
			for i, result := range results {
				id := result.ResourceID
				if id == "" && i < len(chunk) {
					id = chunk[i]
				}
				if result.Allowed && id != "" {
					out[id] = true
				}
			}
		}()
	}
	wg.Wait()
	if first != nil {
		return nil, first
	}
	return out, nil
}

func listTenantProjectIDs(ctx context.Context, tenantID string) ([]string, error) {
	model := store.GetModel()
	if model == nil {
		return nil, fmt.Errorf("project model is not initialized")
	}
	var cond *operator.Condition
	if tenantID != "" {
		cond = operator.NewLeafCondition(operator.Eq, operator.M{"tenantID": tenantID})
	}
	projects, _, err := model.ListProjects(ctx, cond, &page.Pagination{All: true})
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(projects))
	for _, item := range projects {
		if item.ProjectID != "" {
			ids = append(ids, item.ProjectID)
		}
	}
	return ids, nil
}

func compactIDs(ids []string) []string {
	out := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func chunkStrings(in []string, n int) [][]string {
	if n <= 0 {
		n = v4AuthBatchSize
	}
	out := make([][]string, 0, (len(in)+n-1)/n)
	for len(in) > 0 {
		size := n
		if size > len(in) {
			size = len(in)
		}
		out = append(out, in[:size])
		in = in[size:]
	}
	return out
}
