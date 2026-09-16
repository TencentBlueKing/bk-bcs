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
	v4AuthBatchLimit  = 20
	v4AuthConcurrency = 8
)

// ErrIAMV4Unsupported V4 不提供 V3 分级管理员 / 资源创建者授权等管理接口
var ErrIAMV4Unsupported = errors.New("iam v4 does not support this management API")

var (
	_ PermAuthClient = (*iamClient)(nil)
	_ PermAuthClient = (*v4PermClient)(nil)
	_ PermClient     = (*v4PermClient)(nil)
)

func isIAMV4(version string) bool {
	return strings.EqualFold(strings.TrimSpace(version), VersionV4)
}

type v4PermClient struct {
	cli   *iamv4.Client
	opt   *Options
	cache *allowCache
}

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

func (c *v4PermClient) withTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultTimeOut)
}

func (c *v4PermClient) IsAllowedWithoutResource(actionID string, request PermissionRequest, cache bool) (bool, error) {
	return c.IsAllowedWithResource(actionID, request, nil, cache)
}

func (c *v4PermClient) IsAllowedWithResource(actionID string, request PermissionRequest, nodes []ResourceNode,
	cache bool) (bool, error) {
	if c == nil || c.cli == nil {
		return false, ErrServerNotInit
	}
	if err := validateAuthInput(actionID, request); err != nil {
		return false, err
	}
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

func (c *v4PermClient) MultiActionsAllowedWithoutResource(actions []string, request PermissionRequest) (
	map[string]bool, error) {
	return c.ResourceMultiActionsAllowed(actions, request, nil)
}

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

type actionAuthGroup struct {
	actions  []string
	resource *iamv4.AuthResource
}

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

func (c *v4PermClient) BatchResourceMultiActionsAllowed(actions []string, request PermissionRequest,
	nodes [][]ResourceNode) (map[string]map[string]bool, error) {
	if c == nil || c.cli == nil {
		return nil, ErrServerNotInit
	}
	if !request.validate() {
		return nil, fmt.Errorf("systemID/userName is required")
	}
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

type resourceAuthItem struct {
	key string
	res iamv4.AuthResource
}

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
	if !concurrent || len(chunks) == 1 {
		for _, chunk := range chunks {
			if err := run(chunk); err != nil {
				return err
			}
		}
		return nil
	}
	var mu sync.Mutex
	safeSet := set
	set = func(key string, allowed bool) {
		mu.Lock()
		defer mu.Unlock()
		safeSet(key, allowed)
	}
	return runAuthChunksConcurrent(ctx, chunks, run)
}

func runAuthChunksConcurrent[T any](ctx context.Context, jobs []T, run func(T) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	sem := make(chan struct{}, v4AuthConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var first error
	for _, job := range jobs {
		job := job
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
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
	if subtle.ConstantTimeCompare([]byte(token), []byte(user.BkToken)) != 1 {
		return fmt.Errorf("iam basic auth failed")
	}
	return nil
}

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

func (c *v4PermClient) CreateGradeManagers(context.Context, GradeManagerRequest) (uint64, error) {
	return 0, ErrIAMV4Unsupported
}

func (c *v4PermClient) CreateUserGroup(context.Context, uint64, CreateUserGroupRequest) ([]uint64, error) {
	return nil, ErrIAMV4Unsupported
}

func (c *v4PermClient) DeleteUserGroup(context.Context, uint64) error {
	return ErrIAMV4Unsupported
}

func (c *v4PermClient) AddUserGroupMembers(context.Context, uint64, AddGroupMemberRequest) error {
	return ErrIAMV4Unsupported
}

func (c *v4PermClient) DeleteUserGroupMembers(context.Context, uint64, DeleteGroupMemberRequest) error {
	return ErrIAMV4Unsupported
}

func (c *v4PermClient) CreateUserGroupPolicies(context.Context, uint64, AuthorizationScope) error {
	return ErrIAMV4Unsupported
}

func (c *v4PermClient) AuthResourceCreatorPerm(context.Context, ResourceCreator, []Ancestor) error {
	return ErrIAMV4Unsupported
}

func validateAuthInput(actionID string, request PermissionRequest) error {
	if strings.TrimSpace(actionID) == "" {
		return fmt.Errorf("action_id is required")
	}
	if !request.validate() {
		return fmt.Errorf("systemID/userName is required")
	}
	return nil
}

func nodesToV4Resource(nodes []ResourceNode) *iamv4.AuthResource {
	if len(nodes) == 0 {
		return nil
	}
	leaf := nodes[len(nodes)-1]
	res := iamv4.AuthResource{ID: leaf.RInstance, Attributes: map[string]interface{}{}}
	if p := resourceIAMPath(leaf); p != "" {
		res.Attributes[iamv4.IAMPathAttr] = p
	} else if len(nodes) > 1 {
		res.Attributes[iamv4.IAMPathAttr] = buildIAMPathFromAncestors(nodes[:len(nodes)-1])
	}
	for k, v := range leaf.Attr {
		res.Attributes[k] = v
	}
	if len(res.Attributes) == 0 {
		res.Attributes = nil
	}
	return &res
}

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

func resourceForRelatedType(nodes []ResourceNode, wantType string) *iamv4.AuthResource {
	if wantType == "" || len(nodes) == 0 {
		return nil
	}
	leaf := nodes[len(nodes)-1]
	if leaf.RType == wantType {
		return nodesToV4Resource(nodes)
	}
	chain := ancestorChain(nodes)
	for i := len(chain) - 1; i >= 0; i-- {
		if chain[i].RType == wantType {
			return nodesToV4Resource(chain[:i+1])
		}
	}
	return nil
}

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
		case 3:
			if parts[1] != "" && parts[2] != "" {
				out = append(out, ResourceNode{System: parts[0], RType: parts[1], RInstance: parts[2]})
			}
		}
	}
	return out
}

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

func chunkSlices[T any](in []T, n int) [][]T {
	if n <= 0 {
		n = v4AuthBatchLimit
	}
	out := make([][]T, 0, (len(in)+n-1)/n)
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

func cacheKey(actionID, user string, nodes []ResourceNode) string {
	return actionID + "|" + user + "|" + resourceIDFromNodes(nodes)
}

type allowCache struct {
	mu sync.Mutex
	m  map[string]cacheEntry
}

type cacheEntry struct {
	allowed bool
	expire  time.Time
}

func newAllowCache() *allowCache {
	return &allowCache{m: make(map[string]cacheEntry)}
}

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

func (c *allowCache) set(key string, allowed bool, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = cacheEntry{allowed: allowed, expire: time.Now().Add(ttl)}
}
