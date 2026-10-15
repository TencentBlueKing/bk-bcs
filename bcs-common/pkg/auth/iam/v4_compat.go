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

package iam

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iamv4"
)

const (
	// v4AuthBatchLimit 单次 auth-by-actions / auth-by-resources 请求的最大条目数
	v4AuthBatchLimit = 20
	// v4AuthConcurrency 分批鉴权时的最大并发请求数
	v4AuthConcurrency = 8
)

// ErrIAMV4Unsupported V4 不提供 V3 分级管理员 / 资源创建者授权等管理接口
var ErrIAMV4Unsupported = errors.New("iam v4 does not support this management API")

// 编译期校验 V3 / V4 客户端均实现了统一的鉴权接口
var (
	_ PermAuthClient = (*iamClient)(nil)
	_ PermAuthClient = (*v4PermClient)(nil)
	_ PermClient     = (*v4PermClient)(nil)
)

// isIAMV4 判断配置的 IAM 版本是否为 V4（忽略大小写与首尾空白）
func isIAMV4(version string) bool {
	return strings.EqualFold(strings.TrimSpace(version), VersionV4)
}

// v4PermClient 基于 IAM V4 网关实现 PermClient，对上层保持与 V3 一致的调用方式；
// 不支持的 V3 管理类接口统一返回 ErrIAMV4Unsupported
type v4PermClient struct {
	cli   *iamv4.Client
	opt   *Options
	cache *allowCache
}

// newV4PermClient 根据通用 Options 创建 V4 鉴权客户端，并初始化本地鉴权结果缓存
func newV4PermClient(opt *Options) (*v4PermClient, error) {
	cli, err := iamv4.NewClient(&iamv4.Options{
		SystemID:    opt.SystemID,
		AppCode:     opt.AppCode,
		AppSecret:   opt.AppSecret,
		GateWayHost: opt.V4GateWayHost,
		TenantID:    opt.TenantId,
	})
	if err != nil {
		return nil, err
	}
	return &v4PermClient{cli: cli, opt: opt, cache: newAllowCache()}, nil
}

// withTimeout 返回带默认超时的上下文，用于单次对 IAM V4 的调用
func (c *v4PermClient) withTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultTimeOut)
}

// IsAllowedWithoutResource 校验用户对无关联资源操作的权限
func (c *v4PermClient) IsAllowedWithoutResource(actionID string, request PermissionRequest, cache bool) (bool, error) {
	return c.IsAllowedWithResource(actionID, request, nil, cache)
}

// IsAllowedWithResource 校验用户对单个资源的操作权限；cache 为 true 时优先读取本地缓存，
// 并在鉴权成功后写入缓存（TTL 为 defaultAllowTTL）
func (c *v4PermClient) IsAllowedWithResource(actionID string, request PermissionRequest, nodes []ResourceNode,
	cache bool) (bool, error) {
	if c == nil || c.cli == nil {
		return false, ErrServerNotInit
	}
	if err := validateAuthInput(actionID, request); err != nil {
		return false, err
	}
	// 缓存 key 包含用户与完整资源链，不同资源的结果互不覆盖
	key := cacheKey(actionID, request.UserName, nodes)
	if cache {
		if allowed, ok := c.cache.get(key); ok {
			return allowed, nil
		}
	}
	ctx, cancel := c.withTimeout()
	defer cancel()
	res, err := c.cli.DirectAuth(ctx, request.SystemID, iamv4.DirectAuthRequest{
		Subject:  iamv4.Subject{Type: iamv4.SubjectUser, ID: request.UserName},
		ActionID: actionID,
		Resource: nodesToV4Resource(nodes),
	})
	if err != nil {
		return false, err
	}
	if cache {
		c.cache.set(key, res.Allowed, defaultAllowTTL)
	}
	return res.Allowed, nil
}

// BatchResourceIsAllowed 批量校验同一操作对多个资源的权限，返回 资源 key -> 是否有权限；
// 资源数超过单批上限时按批并发请求
func (c *v4PermClient) BatchResourceIsAllowed(actionID string, request PermissionRequest,
	nodes [][]ResourceNode) (map[string]bool, error) {
	if c == nil || c.cli == nil {
		return nil, ErrServerNotInit
	}
	if err := validateAuthInput(actionID, request); err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(nodes))
	if len(nodes) == 0 {
		return out, nil
	}
	items := make([]resourceAuthItem, 0, len(nodes))
	for _, ns := range nodes {
		key := resourceIDFromNodes(ns)
		out[key] = false
		// 空资源链没有叶子节点，退化为以 key 作为资源 ID 鉴权
		res := nodesToV4Resource(ns)
		if res == nil {
			res = &iamv4.AuthResource{ID: key}
		}
		items = append(items, resourceAuthItem{key: key, res: *res})
	}
	ctx, cancel := c.withTimeout()
	defer cancel()
	err := c.authByResources(ctx, request.SystemID,
		iamv4.Subject{Type: iamv4.SubjectUser, ID: request.UserName},
		actionID, items, len(nodes) > v4AuthBatchLimit,
		func(key string, allowed bool) { out[key] = allowed })
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MultiActionsAllowedWithoutResource 批量校验多个无关联资源操作的权限
func (c *v4PermClient) MultiActionsAllowedWithoutResource(actions []string, request PermissionRequest) (
	map[string]bool, error) {
	return c.ResourceMultiActionsAllowed(actions, request, nil)
}

// ResourceMultiActionsAllowed 校验同一资源上多个操作的权限，返回 操作 ID -> 是否有权限
func (c *v4PermClient) ResourceMultiActionsAllowed(actions []string, request PermissionRequest,
	nodes []ResourceNode) (map[string]bool, error) {
	if c == nil || c.cli == nil {
		return nil, ErrServerNotInit
	}
	if !request.validate() {
		return nil, fmt.Errorf("systemID/userName is required")
	}
	out := make(map[string]bool, len(actions))
	if len(actions) == 0 {
		return out, nil
	}
	for _, id := range actions {
		if strings.TrimSpace(id) == "" {
			return nil, fmt.Errorf("action_id is required")
		}
		out[id] = false
	}
	ctx, cancel := c.withTimeout()
	defer cancel()
	subject := iamv4.Subject{Type: iamv4.SubjectUser, ID: request.UserName}
	// V4 auth-by-actions 要求一批操作必须关联同一资源类型（或全部无资源）。
	// 按关联类型拆组，避免 project_view + namespace_list 等混合调用触发 INVALID_ARGUMENT。
	for _, g := range groupActionsByResource(actions, nodes) {
		if err := c.authActionsGroup(ctx, request.SystemID, subject, g.actions, g.resource, out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// actionAuthGroup 关联同一资源（或均无资源）的一组操作，可在一次 auth-by-actions 中鉴权
type actionAuthGroup struct {
	actions  []string
	resource *iamv4.AuthResource
}

// authActionsGroup 按批调用 auth-by-actions 鉴权一组操作，结果写入 out；
// 某批批量鉴权失败时降级为逐个操作鉴权
func (c *v4PermClient) authActionsGroup(ctx context.Context, systemID string, subject iamv4.Subject,
	actions []string, resource *iamv4.AuthResource, out map[string]bool) error {
	for _, chunk := range chunkSlices(actions, v4AuthBatchLimit) {
		results, err := c.cli.DirectAuthByActions(ctx, systemID, iamv4.AuthByActionsRequest{
			Subject:   subject,
			ActionIDs: chunk,
			Resource:  resource,
		})
		if err != nil {
			if _, ferr := c.authActionsFallback(ctx, systemID, subject, chunk, resource, out); ferr != nil {
				return ferr
			}
			continue
		}
		for _, r := range results {
			out[r.ActionID] = r.Allowed
		}
	}
	return nil
}

// authActionsFallback 逐个操作调用 DirectAuth 鉴权，作为批量接口失败时的降级方案
func (c *v4PermClient) authActionsFallback(ctx context.Context, systemID string, subject iamv4.Subject,
	actions []string, resource *iamv4.AuthResource, out map[string]bool) (map[string]bool, error) {
	for _, actionID := range actions {
		res, err := c.cli.DirectAuth(ctx, systemID, iamv4.DirectAuthRequest{
			Subject:  subject,
			ActionID: actionID,
			Resource: resource,
		})
		if err != nil {
			return nil, err
		}
		out[actionID] = res.Allowed
	}
	return out, nil
}

// BatchResourceMultiActionsAllowed 批量校验多个资源上多个操作的权限，返回 资源 key -> 操作 ID -> 是否有权限；
// 每个操作按其关联的资源类型对资源去重后分批鉴权，无关联资源的操作只鉴权一次；
// 结果预先全部填充为 false，未返回结果的资源/操作视为无权限；
// 多个资源链可能指向同一关联资源（如同项目下的多个命名空间），去重后只鉴权一次，结果回填到所有资源链；
// 部分实现不回填 resource_id，此时按请求顺序对齐结果；
// 仅当单个操作的资源超过一批时才并发请求，并发回写结果时加锁
func (c *v4PermClient) BatchResourceMultiActionsAllowed(actions []string, request PermissionRequest,
	nodes [][]ResourceNode) (map[string]map[string]bool, error) {
	if c == nil || c.cli == nil {
		return nil, ErrServerNotInit
	}
	if !request.validate() {
		return nil, fmt.Errorf("systemID/userName is required")
	}
	out, initErr := newBatchResourceActionsResult(actions, nodes)
	if initErr != nil {
		return nil, initErr
	}
	if len(nodes) == 0 || len(actions) == 0 {
		return out, nil
	}

	ctx, cancel := c.withTimeout()
	defer cancel()
	subject := iamv4.Subject{Type: iamv4.SubjectUser, ID: request.UserName}

	type actionJob struct {
		actionID  string
		resources []iamv4.AuthResource
		idToKeys  map[string][]string
	}
	jobs := make([]actionJob, 0)
	maxUnique := 0
	for _, actionID := range actions {
		uniq, idToKeys, noRes := collectUniqueActionResources(actionID, nodes)
		if noRes {
			res, err := c.cli.DirectAuth(ctx, request.SystemID, iamv4.DirectAuthRequest{
				Subject:  subject,
				ActionID: actionID,
			})
			if err != nil {
				return nil, err
			}
			for _, perms := range out {
				perms[actionID] = res.Allowed
			}
			continue
		}
		if len(uniq) > maxUnique {
			maxUnique = len(uniq)
		}
		for _, chunk := range chunkSlices(uniq, v4AuthBatchLimit) {
			jobs = append(jobs, actionJob{actionID: actionID, resources: chunk, idToKeys: idToKeys})
		}
	}

	var mu sync.Mutex
	apply := func(job actionJob, results []iamv4.ResourceAuthResult) {
		byID := make(map[string]bool, len(results))
		for i, r := range results {
			byID[r.ResourceID] = r.Allowed
			if r.ResourceID == "" && i < len(job.resources) {
				byID[job.resources[i].ID] = r.Allowed
			}
		}
		mu.Lock()
		defer mu.Unlock()
		for id, allowed := range byID {
			for _, key := range job.idToKeys[id] {
				if out[key] != nil {
					out[key][job.actionID] = allowed
				}
			}
		}
	}
	run := func(job actionJob) error {
		results, err := c.cli.DirectAuthByResources(ctx, request.SystemID, iamv4.AuthByResourcesRequest{
			Subject:   subject,
			ActionID:  job.actionID,
			Resources: job.resources,
		})
		if err != nil {
			return err
		}
		apply(job, results)
		return nil
	}
	if maxUnique > v4AuthBatchLimit && len(jobs) > 1 {
		return out, runAuthChunksConcurrent(ctx, jobs, run)
	}
	for _, job := range jobs {
		if err := run(job); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// newBatchResourceActionsResult 按资源 key 初始化所有 action 的鉴权结果为 false
func newBatchResourceActionsResult(actions []string, nodes [][]ResourceNode) (map[string]map[string]bool, error) {
	out := make(map[string]map[string]bool, len(nodes))
	for _, ns := range nodes {
		key := resourceIDFromNodes(ns)
		out[key] = make(map[string]bool, len(actions))
		for _, actionID := range actions {
			if strings.TrimSpace(actionID) == "" {
				return nil, fmt.Errorf("action_id is required")
			}
			out[key][actionID] = false
		}
	}
	return out, nil
}

// collectUniqueActionResources 计算操作在各资源链上实际关联的资源并去重，同时返回 资源 ID -> 资源 key 列表 的映射；
// 操作已注册且无关联资源类型时第三个返回值为 true
func collectUniqueActionResources(actionID string, nodes [][]ResourceNode) (
	[]iamv4.AuthResource, map[string][]string, bool) {
	want, known := iamv4.LookupActionResourceType(actionID)
	if known && want == "" {
		return nil, nil, true
	}
	order := make([]string, 0)
	uniq := make(map[string]iamv4.AuthResource)
	idToKeys := make(map[string][]string)
	for _, ns := range nodes {
		key := resourceIDFromNodes(ns)
		var res *iamv4.AuthResource
		if known {
			res = resourceForRelatedType(ns, want)
		} else {
			res = nodesToV4Resource(ns)
		}
		if res == nil || res.ID == "" {
			continue
		}
		// 保持资源首次出现的顺序，便于结果与请求对齐
		if _, ok := uniq[res.ID]; !ok {
			uniq[res.ID] = *res
			order = append(order, res.ID)
		}
		idToKeys[res.ID] = append(idToKeys[res.ID], key)
	}
	resources := make([]iamv4.AuthResource, 0, len(order))
	for _, id := range order {
		resources = append(resources, uniq[id])
	}
	return resources, idToKeys, false
}

// resourceAuthItem 待鉴权资源及其在返回结果中的 key
type resourceAuthItem struct {
	key string
	res iamv4.AuthResource
}

// authByResources 按批调用 auth-by-resources 鉴权同一操作对多个资源的权限，通过 set 回写结果；
// concurrent 为 true 且存在多批时并发执行
func (c *v4PermClient) authByResources(ctx context.Context, systemID string, subject iamv4.Subject,
	actionID string, items []resourceAuthItem, concurrent bool, set func(key string, allowed bool)) error {
	if len(items) == 0 {
		return nil
	}
	chunks := chunkSlices(items, v4AuthBatchLimit)
	run := func(chunk []resourceAuthItem) error {
		resources := make([]iamv4.AuthResource, 0, len(chunk))
		for _, it := range chunk {
			resources = append(resources, it.res)
		}
		results, err := c.cli.DirectAuthByResources(ctx, systemID, iamv4.AuthByResourcesRequest{
			Subject:   subject,
			ActionID:  actionID,
			Resources: resources,
		})
		if err != nil {
			return err
		}
		byID := make(map[string]bool, len(results))
		for i, r := range results {
			byID[r.ResourceID] = r.Allowed
			// 部分实现不回填 resource_id，按请求顺序对齐结果
			if r.ResourceID == "" && i < len(resources) {
				byID[resources[i].ID] = r.Allowed
			}
		}
		for _, it := range chunk {
			if allowed, ok := byID[it.res.ID]; ok {
				set(it.key, allowed)
			}
		}
		return nil
	}
	// 串行执行时 set 无需加锁
	if !concurrent || len(chunks) == 1 {
		for _, chunk := range chunks {
			if err := run(chunk); err != nil {
				return err
			}
		}
		return nil
	}
	// 并发执行时包装 set，保证调用方回调的并发安全
	var mu sync.Mutex
	safeSet := set
	set = func(key string, allowed bool) {
		mu.Lock()
		defer mu.Unlock()
		safeSet(key, allowed)
	}
	return runAuthChunksConcurrent(ctx, chunks, run)
}

// runAuthChunksConcurrent 以 v4AuthConcurrency 为上限并发执行 jobs，
// 任一任务失败即取消其余任务并返回首个错误
func runAuthChunksConcurrent[T any](ctx context.Context, jobs []T, run func(T) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	sem := make(chan struct{}, v4AuthConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var first error
	for _, job := range jobs {
		// 拷贝循环变量，避免 goroutine 共享同一变量
		job := job
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			// 已有任务失败或上游取消时不再发起新请求
			case <-ctx.Done():
				return
			}
			if err := run(job); err != nil {
				mu.Lock()
				if first == nil {
					first = err
					cancel()
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return first
}

// GetToken 获取 IAM V4 为本系统签发的回调鉴权 token
func (c *v4PermClient) GetToken() (string, error) {
	if c == nil || c.cli == nil {
		return "", ErrServerNotInit
	}
	ctx, cancel := c.withTimeout()
	defer cancel()
	data, err := c.cli.RetrieveSystemAuthToken(ctx, c.opt.SystemID)
	if err != nil {
		return "", err
	}
	if data == nil || data.AuthToken == "" {
		return "", fmt.Errorf("iam v4 auth_token is empty")
	}
	return data.AuthToken, nil
}

// IsBasicAuthAllowed 校验 IAM 回调请求携带的 basic auth token 是否与系统 token 一致
func (c *v4PermClient) IsBasicAuthAllowed(user BkUser) error {
	if c == nil || c.cli == nil {
		return ErrServerNotInit
	}
	if strings.TrimSpace(user.BkUserName) == "" || user.BkToken == "" {
		return fmt.Errorf("username/token is required")
	}
	token, err := c.GetToken()
	if err != nil {
		return err
	}
	// 常量时间比较，避免通过响应耗时推测 token
	if subtle.ConstantTimeCompare([]byte(token), []byte(user.BkToken)) != 1 {
		return fmt.Errorf("iam basic auth failed")
	}
	return nil
}

// GetApplyURL 生成权限申请链接；生成失败时返回 IamAppURL 与错误
func (c *v4PermClient) GetApplyURL(request ApplicationRequest, relatedResources []ApplicationAction,
	_ BkUser) (string, error) {
	if c == nil || c.cli == nil {
		return "", ErrServerNotInit
	}
	systemID := request.SystemID
	if systemID == "" {
		systemID = c.opt.SystemID
	}
	perms := toV4ApplyPermissions(relatedResources)
	if len(perms) == 0 {
		return "", fmt.Errorf("permissions is required")
	}
	ctx, cancel := c.withTimeout()
	defer cancel()
	data, err := c.cli.GeneratePermApplyURL(ctx, iamv4.GenerateApplyURLRequest{
		SystemID:    systemID,
		Permissions: perms,
	})
	if err != nil {
		return IamAppURL, err
	}
	return data.URL, nil
}

// CreateGradeManagers V4 不支持分级管理员，固定返回 ErrIAMV4Unsupported
func (c *v4PermClient) CreateGradeManagers(context.Context, GradeManagerRequest) (uint64, error) {
	return 0, ErrIAMV4Unsupported
}

// CreateUserGroup V4 不支持用户组管理，固定返回 ErrIAMV4Unsupported
func (c *v4PermClient) CreateUserGroup(context.Context, uint64, CreateUserGroupRequest) ([]uint64, error) {
	return nil, ErrIAMV4Unsupported
}

// DeleteUserGroup V4 不支持用户组管理，固定返回 ErrIAMV4Unsupported
func (c *v4PermClient) DeleteUserGroup(context.Context, uint64) error {
	return ErrIAMV4Unsupported
}

// AddUserGroupMembers V4 不支持用户组管理，固定返回 ErrIAMV4Unsupported
func (c *v4PermClient) AddUserGroupMembers(context.Context, uint64, AddGroupMemberRequest) error {
	return ErrIAMV4Unsupported
}

// DeleteUserGroupMembers V4 不支持用户组管理，固定返回 ErrIAMV4Unsupported
func (c *v4PermClient) DeleteUserGroupMembers(context.Context, uint64, DeleteGroupMemberRequest) error {
	return ErrIAMV4Unsupported
}

// CreateUserGroupPolicies V4 不支持用户组授权，固定返回 ErrIAMV4Unsupported
func (c *v4PermClient) CreateUserGroupPolicies(context.Context, uint64, AuthorizationScope) error {
	return ErrIAMV4Unsupported
}

// AuthResourceCreatorPerm V4 不支持资源创建者授权，固定返回 ErrIAMV4Unsupported
func (c *v4PermClient) AuthResourceCreatorPerm(context.Context, ResourceCreator, []Ancestor) error {
	return ErrIAMV4Unsupported
}

// validateAuthInput 校验单操作鉴权请求的必填参数
func validateAuthInput(actionID string, request PermissionRequest) error {
	if strings.TrimSpace(actionID) == "" {
		return fmt.Errorf("action_id is required")
	}
	if !request.validate() {
		return fmt.Errorf("systemID/userName is required")
	}
	return nil
}

// nodesToV4Resource 将资源链转换为 V4 鉴权资源：以叶子节点为资源实例，
// 优先使用叶子自带的 _bk_iam_path_，否则由祖先节点拼接；叶子属性一并透传
func nodesToV4Resource(nodes []ResourceNode) *iamv4.AuthResource {
	if len(nodes) == 0 {
		return nil
	}
	leaf := nodes[len(nodes)-1]
	res := iamv4.AuthResource{ID: leaf.RInstance, Attributes: map[string]interface{}{}}
	// 叶子自带路径时优先使用，否则由祖先节点拼接
	if p := resourceIAMPath(leaf); p != "" {
		res.Attributes[iamv4.IAMPathAttr] = p
	} else if len(nodes) > 1 {
		res.Attributes[iamv4.IAMPathAttr] = buildIAMPathFromAncestors(nodes[:len(nodes)-1])
	}
	for k, v := range leaf.Attr {
		res.Attributes[k] = v
	}
	// 无属性时置空，请求体中省略 attributes 字段
	if len(res.Attributes) == 0 {
		res.Attributes = nil
	}
	return &res
}

// resourceIAMPath 读取节点的 IAM 路径，依次尝试 ResourcePath 构建与 Attr 中的 _bk_iam_path_
func resourceIAMPath(node ResourceNode) string {
	if node.Rp != nil {
		if p := node.Rp.BuildIAMPath(); p != "" {
			return p
		}
	}
	if node.Attr != nil {
		if v, ok := node.Attr[string(BkIAMPath)].(string); ok {
			return v
		}
	}
	return ""
}

// buildIAMPathFromAncestors 由祖先节点拼接 IAM 路径，格式为 /type,id/type,id/
func buildIAMPathFromAncestors(nodes []ResourceNode) string {
	var b strings.Builder
	for _, n := range nodes {
		if n.RType == "" || n.RInstance == "" {
			continue
		}
		b.WriteByte('/')
		b.WriteString(n.RType)
		b.WriteByte(',')
		b.WriteString(n.RInstance)
	}
	if b.Len() == 0 {
		return ""
	}
	b.WriteByte('/')
	return b.String()
}

// groupActionsByResource 按操作关联的资源类型与资源 ID 对操作分组，保持首次出现顺序；
// 资源链中找不到关联类型的操作直接跳过（结果保持 false）
func groupActionsByResource(actions []string, nodes []ResourceNode) []actionAuthGroup {
	type groupKey struct {
		typ string
		id  string
	}
	type bucket struct {
		actions  []string
		resource *iamv4.AuthResource
	}
	buckets := make(map[groupKey]*bucket)
	order := make([]groupKey, 0)
	for _, id := range actions {
		// 未注册的操作无法判断关联类型，按完整资源链鉴权并单独成组
		want, known := iamv4.LookupActionResourceType(id)
		var res *iamv4.AuthResource
		if known {
			if want != "" {
				res = resourceForRelatedType(nodes, want)
				if res == nil {
					continue
				}
			}
		} else {
			res = nodesToV4Resource(nodes)
		}
		key := groupKey{typ: want, id: ""}
		if res != nil {
			key.id = res.ID
		}
		if !known {
			key.typ = "_unknown"
		}
		if buckets[key] == nil {
			buckets[key] = &bucket{resource: res}
			order = append(order, key)
		}
		buckets[key].actions = append(buckets[key].actions, id)
	}
	out := make([]actionAuthGroup, 0, len(order))
	for _, k := range order {
		b := buckets[k]
		out = append(out, actionAuthGroup{actions: b.actions, resource: b.resource})
	}
	return out
}

// resourceForRelatedType 从资源链（含 IAM 路径解析出的祖先）中截取到指定类型的资源
func resourceForRelatedType(nodes []ResourceNode, wantType string) *iamv4.AuthResource {
	if wantType == "" || len(nodes) == 0 {
		return nil
	}
	// 叶子即为目标类型时直接使用完整资源链
	leaf := nodes[len(nodes)-1]
	if leaf.RType == wantType {
		return nodesToV4Resource(nodes)
	}
	chain := ancestorChain(nodes)
	// 从叶子向根查找，取离叶子最近的目标类型节点
	for i := len(chain) - 1; i >= 0; i-- {
		if chain[i].RType == wantType {
			return nodesToV4Resource(chain[:i+1])
		}
	}
	return nil
}

// ancestorChain 返回完整资源链；仅有叶子节点时从其 IAM 路径解析出祖先节点
func ancestorChain(nodes []ResourceNode) []ResourceNode {
	if len(nodes) == 0 {
		return nil
	}
	if len(nodes) > 1 {
		return nodes
	}
	leaf := nodes[0]
	parsed := parseIAMPath(resourceIAMPath(leaf))
	if len(parsed) == 0 {
		return nodes
	}
	return append(parsed, leaf)
}

// parseIAMPath 解析 /type,id/ 或 /system,type,id/ 格式的 IAM 路径为资源节点列表
func parseIAMPath(p string) []ResourceNode {
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}
	out := make([]ResourceNode, 0)
	for _, seg := range strings.Split(p, "/") {
		if seg == "" {
			continue
		}
		parts := strings.Split(seg, ",")
		switch len(parts) {
		case 2:
			if parts[0] != "" && parts[1] != "" {
				out = append(out, ResourceNode{RType: parts[0], RInstance: parts[1]})
			}
		// 带系统前缀的段：system,type,id
		case 3:
			if parts[1] != "" && parts[2] != "" {
				out = append(out, ResourceNode{System: parts[0], RType: parts[1], RInstance: parts[2]})
			}
		}
	}
	return out
}

// resourceIDFromNodes 生成资源链在批量鉴权结果中的 key：单节点为实例 ID，多节点为 type:id 以 / 连接
func resourceIDFromNodes(nodes []ResourceNode) string {
	if len(nodes) == 0 {
		return ""
	}
	if len(nodes) == 1 {
		return nodes[0].RInstance
	}
	parts := make([]string, 0, len(nodes))
	for _, n := range nodes {
		parts = append(parts, n.RType+":"+n.RInstance)
	}
	return strings.Join(parts, "/")
}

// toV4ApplyPermissions 将 V3 风格的申请操作转换为 V4 权限申请参数，实例链的最后一个节点为资源、其余为祖先
func toV4ApplyPermissions(actions []ApplicationAction) []iamv4.ApplyPermission {
	out := make([]iamv4.ApplyPermission, 0, len(actions))
	for _, a := range actions {
		perm := iamv4.ApplyPermission{ActionID: a.ActionID}
		for _, rt := range a.RelatedResources {
			for _, inst := range rt.Instances {
				if len(inst) == 0 {
					continue
				}
				leaf := inst[len(inst)-1]
				res := iamv4.ApplyResource{ID: leaf.ID, Type: leaf.Type}
				if res.Type == "" {
					res.Type = rt.Type
				}
				if len(inst) > 1 {
					res.Ancestors = make([]iamv4.ApplyAncestor, 0, len(inst)-1)
					for _, n := range inst[:len(inst)-1] {
						res.Ancestors = append(res.Ancestors, iamv4.ApplyAncestor{ID: n.ID, Type: n.Type})
					}
				}
				perm.Resources = append(perm.Resources, res)
			}
		}
		out = append(out, perm)
	}
	return out
}

// chunkSlices 将切片按 n 个一组切分，n 非正数时使用 v4AuthBatchLimit
func chunkSlices[T any](in []T, n int) [][]T {
	if n <= 0 {
		n = v4AuthBatchLimit
	}
	out := make([][]T, 0, (len(in)+n-1)/n)
	// 子切片共享底层数组，调用方不应修改返回结果
	for len(in) > 0 {
		m := n
		if m > len(in) {
			m = len(in)
		}
		out = append(out, in[:m])
		in = in[m:]
	}
	return out
}

// cacheKey 生成鉴权结果缓存 key：操作 ID|用户名|资源 key
func cacheKey(actionID, user string, nodes []ResourceNode) string {
	return actionID + "|" + user + "|" + resourceIDFromNodes(nodes)
}

// allowCache 带过期时间的鉴权结果本地缓存，并发安全
type allowCache struct {
	mu sync.Mutex
	m  map[string]cacheEntry
}

// cacheEntry 单条鉴权结果及其过期时间
type cacheEntry struct {
	allowed bool
	expire  time.Time
}

// newAllowCache 创建空的鉴权结果缓存
func newAllowCache() *allowCache {
	return &allowCache{m: make(map[string]cacheEntry)}
}

// get 读取未过期的鉴权结果，第二个返回值表示是否命中；过期条目会被顺带清理
func (c *allowCache) get(key string) (bool, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.m[key]
	if !ok || time.Now().After(e.expire) {
		if ok {
			delete(c.m, key)
		}
		return false, false
	}
	return e.allowed, true
}

// set 写入鉴权结果并设置过期时间
func (c *allowCache) set(key string, allowed bool, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = cacheEntry{allowed: allowed, expire: time.Now().Add(ttl)}
}
