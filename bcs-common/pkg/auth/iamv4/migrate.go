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

package iamv4

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strconv"
	"text/template"

	"k8s.io/klog/v2"
)

const (
	// OpUpsertSystem 注册 / 更新系统
	OpUpsertSystem = "upsert_system"
	// OpUpsertResourceType 注册 / 更新资源类型
	OpUpsertResourceType = "upsert_resource_type"
	// OpUpsertAction 注册 / 更新操作
	OpUpsertAction = "upsert_action"
	// OpUpsertRole 注册 / 更新角色（含操作同步）
	OpUpsertRole = "upsert_role"

	listPageSize = 100
	maxListPages = 10000
)

var migrateFileRe = regexp.MustCompile(`^(\d+)_.+\.up\.json$`)

// MigrationFile 单份权限模型 migration
type MigrationFile struct {
	SystemID   string               `json:"system_id"`
	Enabled    bool                 `json:"enabled"`
	Operations []MigrationOperation `json:"operations"`
}

// MigrationOperation 单条模型变更
type MigrationOperation struct {
	Operation string          `json:"operation"`
	Data      json.RawMessage `json:"data"`
}

type migrateFile struct {
	version int
	name    string
}

// Migrate 读取 fs 中 `{version}_{name}.up.json`，按版本升序渲染 go 模板后幂等 upsert。
// V4 无官方 migrate API，本方法通过 retrieve/list → create/update 实现。
func (c *Client) Migrate(ctx context.Context, fsys fs.FS, templateVar interface{}) error {
	if c == nil || c.opt == nil {
		return ErrServerNotInit
	}
	if fsys == nil {
		return fmt.Errorf("migration fs is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	files, err := listMigrateFiles(fsys)
	if err != nil {
		return err
	}
	for _, file := range files {
		body, err := renderMigrateFile(fsys, file.name, templateVar)
		if err != nil {
			return fmt.Errorf("render %s: %w", file.name, err)
		}
		var mf MigrationFile
		if err := json.Unmarshal(body, &mf); err != nil {
			return fmt.Errorf("decode %s: %w", file.name, err)
		}
		if !mf.Enabled {
			klog.Infof("iamv4 migrate skip disabled file %s", file.name)
			continue
		}
		systemID := mf.SystemID
		if systemID == "" {
			systemID = c.opt.SystemID
		}
		klog.Infof("iamv4 migrate apply %s system=%s ops=%d", file.name, systemID, len(mf.Operations))
		for i, op := range mf.Operations {
			if err := c.applyOperation(ctx, systemID, op); err != nil {
				return fmt.Errorf("%s operations[%d] %s: %w", file.name, i, op.Operation, err)
			}
		}
	}
	return nil
}

func listMigrateFiles(fsys fs.FS) ([]migrateFile, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("read migration dir: %w", err)
	}
	files := make([]migrateFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		m := migrateFileRe.FindStringSubmatch(name)
		if m == nil {
			continue
		}
		ver, err := strconv.Atoi(m[1])
		if err != nil {
			return nil, fmt.Errorf("invalid migration version %s", name)
		}
		files = append(files, migrateFile{version: ver, name: name})
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].version != files[j].version {
			return files[i].version < files[j].version
		}
		return files[i].name < files[j].name
	})
	return files, nil
}

func renderMigrateFile(fsys fs.FS, name string, templateVar interface{}) ([]byte, error) {
	raw, err := fs.ReadFile(fsys, path.Clean(name))
	if err != nil {
		return nil, err
	}
	if templateVar == nil {
		return raw, nil
	}
	tpl, err := template.New(name).Option("missingkey=error").Parse(string(raw))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, templateVar); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *Client) applyOperation(ctx context.Context, systemID string, op MigrationOperation) error {
	switch op.Operation {
	case OpUpsertSystem:
		var data CreateSystemRequest
		if err := json.Unmarshal(op.Data, &data); err != nil {
			return err
		}
		return c.upsertSystem(ctx, data)
	case OpUpsertResourceType:
		var data ResourceType
		if err := json.Unmarshal(op.Data, &data); err != nil {
			return err
		}
		return c.upsertResourceType(ctx, systemID, data)
	case OpUpsertAction:
		var data Action
		if err := json.Unmarshal(op.Data, &data); err != nil {
			return err
		}
		return c.upsertAction(ctx, systemID, data)
	case OpUpsertRole:
		var data Role
		if err := json.Unmarshal(op.Data, &data); err != nil {
			return err
		}
		return c.upsertRole(ctx, systemID, data)
	default:
		return fmt.Errorf("unsupported operation %q", op.Operation)
	}
}

func (c *Client) upsertSystem(ctx context.Context, req CreateSystemRequest) error {
	if err := requireModelID("id", req.ID); err != nil {
		return err
	}
	exist, err := c.RetrieveSystem(ctx, req.ID)
	if err != nil && !isNotFound(err) {
		return err
	}
	if exist == nil || isNotFound(err) {
		if _, err := c.CreateSystem(ctx, req); err != nil {
			if isAlreadyExists(err) {
				return c.UpdateSystem(ctx, req.ID, systemToUpdate(req))
			}
			return err
		}
		klog.Infof("iamv4 migrate created system %s", req.ID)
		return nil
	}
	if err := c.UpdateSystem(ctx, req.ID, systemToUpdate(req)); err != nil {
		return err
	}
	klog.Infof("iamv4 migrate updated system %s", req.ID)
	return nil
}

func systemToUpdate(req CreateSystemRequest) UpdateSystemRequest {
	return UpdateSystemRequest{
		Name:        req.Name,
		Description: req.Description,
		Managers:    req.Managers,
		Clients:     req.Clients,
		CallbackURL: req.CallbackURL,
	}
}

func (c *Client) upsertResourceType(ctx context.Context, systemID string, item ResourceType) error {
	if err := requireModelID("resource_type.id", item.ID); err != nil {
		return err
	}
	exist, err := c.findResourceType(ctx, systemID, item.ID)
	if err != nil {
		return err
	}
	if exist == nil {
		if _, err := c.BatchCreateResourceType(ctx, systemID, []ResourceType{item}); err != nil {
			if isAlreadyExists(err) {
				return c.UpdateResourceType(ctx, systemID, item.ID, UpdateResourceTypeRequest{
					Name: item.Name, Ancestors: item.Ancestors,
				})
			}
			return err
		}
		klog.Infof("iamv4 migrate created resource_type %s", item.ID)
		return nil
	}
	if err := c.UpdateResourceType(ctx, systemID, item.ID, UpdateResourceTypeRequest{
		Name: item.Name, Ancestors: item.Ancestors,
	}); err != nil {
		return err
	}
	klog.Infof("iamv4 migrate updated resource_type %s", item.ID)
	return nil
}

func (c *Client) upsertAction(ctx context.Context, systemID string, item Action) error {
	if err := requireModelID("action.id", item.ID); err != nil {
		return err
	}
	exist, err := c.findAction(ctx, systemID, item.ID)
	if err != nil {
		return err
	}
	if exist == nil {
		if _, err := c.BatchCreateAction(ctx, systemID, []Action{item}); err != nil {
			if isAlreadyExists(err) {
				return c.UpdateAction(ctx, systemID, item.ID, UpdateActionRequest{Name: item.Name})
			}
			return err
		}
		klog.Infof("iamv4 migrate created action %s", item.ID)
		return nil
	}
	if item.Name == "" || item.Name == exist.Name {
		klog.Infof("iamv4 migrate action %s unchanged", item.ID)
		return nil
	}
	if err := c.UpdateAction(ctx, systemID, item.ID, UpdateActionRequest{Name: item.Name}); err != nil {
		return err
	}
	klog.Infof("iamv4 migrate updated action %s", item.ID)
	return nil
}

func (c *Client) upsertRole(ctx context.Context, systemID string, item Role) error {
	if err := requireModelID("role.id", item.ID); err != nil {
		return err
	}
	if len(item.Actions) == 0 {
		return fmt.Errorf("role %s actions is required", item.ID)
	}
	exist, err := c.findRole(ctx, systemID, item.ID)
	if err != nil {
		return err
	}
	if exist == nil {
		if _, err := c.BatchCreateRole(ctx, systemID, []Role{item}); err != nil {
			if isAlreadyExists(err) {
				exist, err = c.findRole(ctx, systemID, item.ID)
				if err != nil {
					return err
				}
				if exist == nil {
					return fmt.Errorf("role %s already exists but not found", item.ID)
				}
			} else {
				return err
			}
		} else {
			klog.Infof("iamv4 migrate created role %s", item.ID)
			return nil
		}
	}
	if err := c.UpdateRole(ctx, systemID, item.ID, UpdateRoleRequest{
		Name: item.Name, Description: item.Description,
	}); err != nil {
		return err
	}
	if err := c.syncRoleActions(ctx, systemID, item.ID, exist.Actions, item.Actions); err != nil {
		return err
	}
	klog.Infof("iamv4 migrate updated role %s", item.ID)
	return nil
}

func (c *Client) syncRoleActions(ctx context.Context, systemID, roleID string, current, wanted []RoleAction) error {
	cur := make(map[string]RoleAction, len(current))
	for _, a := range current {
		cur[a.ID] = a
	}
	want := make(map[string]RoleAction, len(wanted))
	for _, a := range wanted {
		want[a.ID] = a
	}
	var toAdd []RoleAction
	var toDelete []string
	for id, a := range want {
		if exist, ok := cur[id]; !ok || exist.ResourceTypeID != a.ResourceTypeID {
			if ok {
				toDelete = append(toDelete, id)
			}
			toAdd = append(toAdd, a)
		}
	}
	for id := range cur {
		if _, ok := want[id]; !ok {
			toDelete = append(toDelete, id)
		}
	}
	if len(toDelete) > 0 {
		if err := c.BatchDeleteRoleAction(ctx, systemID, roleID, toDelete); err != nil && !isNotFound(err) {
			return err
		}
	}
	if len(toAdd) > 0 {
		if _, err := c.BatchCreateRoleAction(ctx, systemID, roleID, toAdd); err != nil && !isAlreadyExists(err) {
			return err
		}
	}
	return nil
}

func (c *Client) findResourceType(ctx context.Context, systemID, id string) (*ResourceType, error) {
	items, err := listAll(func(page PageQuery) (*PageResult[ResourceType], error) {
		return c.ListResourceType(ctx, systemID, page)
	})
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].ID == id {
			return &items[i], nil
		}
	}
	return nil, nil
}

func (c *Client) findAction(ctx context.Context, systemID, id string) (*Action, error) {
	items, err := listAll(func(page PageQuery) (*PageResult[Action], error) {
		return c.ListAction(ctx, systemID, page)
	})
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].ID == id {
			return &items[i], nil
		}
	}
	return nil, nil
}

func (c *Client) findRole(ctx context.Context, systemID, id string) (*Role, error) {
	items, err := listAll(func(page PageQuery) (*PageResult[Role], error) {
		return c.ListRole(ctx, systemID, page)
	})
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].ID == id {
			return &items[i], nil
		}
	}
	return nil, nil
}

func listAll[T any](fetch func(PageQuery) (*PageResult[T], error)) ([]T, error) {
	var all []T
	for page := 1; page <= maxListPages; page++ {
		res, err := fetch(PageQuery{Page: page, PageSize: listPageSize})
		if err != nil {
			return nil, err
		}
		if res == nil {
			break
		}
		all = append(all, res.Results...)
		if len(res.Results) < listPageSize {
			break
		}
		if res.Count > 0 && len(all) >= res.Count {
			break
		}
	}
	return all, nil
}
