/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package view 资源视图管理
package view

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam"
	projectAuth "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm/resource/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/utils"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// SystemViews 系统预置视图
var SystemViews = []string{"默认视图", "Default view"}

// ViewAction provides the action to manager view
type ViewAction struct {
	ctx context.Context

	model store.ClusterResourcesModel
}

// NewViewAction return a new ViewAction instance
func NewViewAction(model store.ClusterResourcesModel) *ViewAction {
	return &ViewAction{
		model: model,
	}
}

func (v *ViewAction) checkAccess(ctx context.Context) error {
	if config.G.Auth.Disabled {
		return nil
	}
	project, err := project.FromContext(ctx)
	if err != nil {
		return err
	}
	// 权限控制为项目查看
	permCtx := &projectAuth.PermCtx{
		Username:  ctx.Value(ctxkey.UsernameKey).(string),
		ProjectID: project.ID,
	}
	if allow, err := iam.NewProjectPerm().CanView(permCtx); err != nil {
		return err
	} else if !allow {
		return errorx.New(errcode.NoIAMPerm, i18n.GetMsg(ctx, "无项目查看权限"))
	}
	return nil
}

// List xxx
func (v *ViewAction) List(ctx context.Context) ([]map[string]interface{}, error) {
	if err := v.checkAccess(ctx); err != nil {
		return nil, err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// get query cond
	// 获取公共视图
	condPublic := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode: p.Code,
		entity.FieldKeyScope:       entity.ViewScopePublic,
	})
	// 获取用户自己创建的视图
	condPrivate := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode: p.Code,
		entity.FieldKeyScope:       entity.ViewScopePrivate,
		entity.FieldKeyCreateBy:    ctxkey.GetUsernameFromCtx(ctx),
	})
	cond := operator.NewBranchCondition(operator.Or, condPublic, condPrivate)
	_, views, err := v.model.ListViews(ctx, cond, &utils.ListOption{Sort: map[string]int{"name": 1}})
	if err != nil {
		return nil, err
	}

	m := make([]map[string]interface{}, 0)
	for _, v := range views {
		m = append(m, v.ToMap())
	}
	return m, nil
}

// Get xxx
func (v *ViewAction) Get(ctx context.Context, id, projectCode string) (map[string]interface{}, error) {
	if err := v.checkAccess(ctx); err != nil {
		return nil, err
	}

	view, err := v.model.GetView(ctx, id)
	if err != nil {
		return nil, err
	}

	// 只能查看当前项目视图，或者自己创建的视图，或者公共视图
	if view.ProjectCode != projectCode || (view.Scope == entity.ViewScopePrivate &&
		view.CreateBy != ctxkey.GetUsernameFromCtx(ctx)) {
		return nil, errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "无权限访问"))
	}

	return view.ToMap(), nil
}

// Create xxx
func (v *ViewAction) Create(ctx context.Context, req *clusterRes.CreateViewConfigReq) (string, error) {
	if err := v.checkAccess(ctx); err != nil {
		return "", err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return "", err
	}

	// 检测同名
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode: p.Code,
	})
	_, views, err := v.model.ListViews(ctx, cond, &utils.ListOption{})
	if err != nil {
		return "", err
	}

	// 另存为的时候，自动命名，如: aaa 会被自动命名为 aaa copy，aaa copy 再次另存为时，会被命名为 aaa copy 2
	name := strings.TrimRight(strings.TrimLeft(req.GetName(), " "), " ")
	if req.GetSaveAs() {
		name = getViewCopyName(name, views)
	} else {
		// 非另存为的情况，检测是否重名
		var names []string
		names = append(names, SystemViews...)
		for _, v := range views {
			// 在自己可见的视图中，检测是否重名
			if v.Scope == entity.ViewScopePrivate && v.CreateBy != ctxkey.GetUsernameFromCtx(ctx) {
				continue
			}
			names = append(names, v.Name)
		}
		if slice.StringInSlice(name, names) {
			return "", errorx.New(errcode.DuplicationNameErr, i18n.GetMsg(ctx, "视图名称重复"))
		}
	}

	view := &entity.View{
		Name:        name,
		ProjectCode: p.Code,
		ClusterID:   req.GetClusterID(),
		Namespace:   req.GetNamespace(),
		Filter:      protoFilterToFilter(req.GetFilter()),
		Scope:       entity.ViewScopePrivate,
		CreateBy:    ctxkey.GetUsernameFromCtx(ctx),
	}
	id, err := v.model.CreateView(ctx, view)
	if err != nil {
		return "", err
	}
	return id, nil
}

// Update xxx
func (v *ViewAction) Update(ctx context.Context, req *clusterRes.UpdateViewConfigReq) error {
	if err := v.checkAccess(ctx); err != nil {
		return err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return err
	}

	view, err := v.model.GetView(ctx, req.GetId())
	if err != nil {
		return err
	}

	// 检验更新 view 的权限
	if view.ProjectCode != p.Code || (view.Scope == entity.ViewScopePrivate &&
		view.CreateBy != ctxkey.GetUsernameFromCtx(ctx)) {
		return errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "无权限访问"))
	}

	// 检测同名
	name := strings.TrimRight(strings.TrimLeft(req.GetName(), " "), " ")
	objectID, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return err
	}
	// 在自己可见的视图中，检测是否重名
	condPublic := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode: p.Code,
		entity.FieldKeyScope:       entity.ViewScopePublic,
		entity.FieldKeyName:        name,
	})
	condPrivate := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode: p.Code,
		entity.FieldKeyScope:       entity.ViewScopePrivate,
		entity.FieldKeyCreateBy:    ctxkey.GetUsernameFromCtx(ctx),
		entity.FieldKeyName:        name,
	})
	condName := operator.NewBranchCondition(operator.Or, condPublic, condPrivate)
	condID := operator.NewLeafCondition(operator.Ne, operator.M{
		entity.FieldKeyObjectID: objectID,
	})
	cond := operator.NewBranchCondition(operator.And, condName, condID)
	count, _, err := v.model.ListViews(ctx, cond, &utils.ListOption{})
	if err != nil {
		return err
	}

	if count > 0 || slice.StringInSlice(name, SystemViews) {
		return errorx.New(errcode.DuplicationNameErr, i18n.GetMsg(ctx, "视图名称重复"))
	}

	updateView := entity.M{
		"name":      name,
		"clusterID": req.GetClusterID(),
		"namespace": req.GetNamespace(),
		"filter":    protoFilterToFilter(req.GetFilter()),
	}
	if err := v.model.UpdateView(ctx, req.GetId(), updateView); err != nil {
		return err
	}
	return nil
}

// Rename xxx
func (v *ViewAction) Rename(ctx context.Context, req *clusterRes.RenameViewConfigReq) error {
	if err := v.checkAccess(ctx); err != nil {
		return err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return err
	}

	view, err := v.model.GetView(ctx, req.GetId())
	if err != nil {
		return err
	}

	// 检验更新 view 的权限
	if view.ProjectCode != p.Code || (view.Scope == entity.ViewScopePrivate &&
		view.CreateBy != ctxkey.GetUsernameFromCtx(ctx)) {
		return errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "无权限访问"))
	}

	// 检测同名
	name := strings.TrimRight(strings.TrimLeft(req.GetName(), " "), " ")
	objectID, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return err
	}
	// 在自己可见的视图中，检测是否重名
	condPublic := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode: p.Code,
		entity.FieldKeyScope:       entity.ViewScopePublic,
		entity.FieldKeyName:        name,
	})
	condPrivate := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode: p.Code,
		entity.FieldKeyScope:       entity.ViewScopePrivate,
		entity.FieldKeyCreateBy:    ctxkey.GetUsernameFromCtx(ctx),
		entity.FieldKeyName:        name,
	})
	condName := operator.NewBranchCondition(operator.Or, condPublic, condPrivate)
	condID := operator.NewLeafCondition(operator.Ne, operator.M{
		entity.FieldKeyObjectID: objectID,
	})
	cond := operator.NewBranchCondition(operator.And, condName, condID)
	count, _, err := v.model.ListViews(ctx, cond, &utils.ListOption{})
	if err != nil {
		return err
	}

	if count > 0 || slice.StringInSlice(name, SystemViews) {
		return errorx.New(errcode.DuplicationNameErr, i18n.GetMsg(ctx, "视图名称重复"))
	}

	updateView := entity.M{
		"name": name,
	}
	if err := v.model.UpdateView(ctx, req.GetId(), updateView); err != nil {
		return err
	}
	return nil
}

// Delete xxx
func (v *ViewAction) Delete(ctx context.Context, id string) error {
	if err := v.checkAccess(ctx); err != nil {
		return err
	}

	view, err := v.model.GetView(ctx, id)
	if err != nil {
		return err
	}

	// 检验该 view 是否是该用户创建的
	if view.CreateBy != ctxkey.GetUsernameFromCtx(ctx) {
		return errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "无权限访问"))
	}

	if err := v.model.DeleteView(ctx, id); err != nil {
		return err
	}
	return nil
}

func protoFilterToFilter(filter *clusterRes.ViewFilter) *entity.ViewFilter {
	if filter == nil {
		return nil
	}
	return &entity.ViewFilter{
		Name:          filter.GetName(),
		Creator:       filter.GetCreator(),
		LabelSelector: filter.GetLabelSelector(),
	}
}

func getViewCopyName(name string, views []*entity.View) string {
	names := make([]string, 0, len(views))
	for _, v := range views {
		names = append(names, v.Name)
	}
	return getNewFileName(name, names)
}

func getNewFileName(name string, files []string) string {
	sort.Strings(files)
	maxCopy := 0
	baseName := name
	if strings.Contains(name, "copy") {
		baseName = strings.TrimSpace(strings.Split(name, "copy")[0])
	}
	for _, file := range files {
		if strings.HasPrefix(file, baseName) {
			parts := strings.Split(file, " ")
			if len(parts) > 2 {
				copyNum, err := strconv.Atoi(parts[len(parts)-1])
				if err == nil && copyNum > maxCopy {
					maxCopy = copyNum
				}
			} else if len(parts) == 2 && parts[1] == "copy" {
				maxCopy = 1
			}
		}
	}

	if maxCopy == 0 {
		return fmt.Sprintf("%s copy", baseName)
	}
	return fmt.Sprintf("%s copy %d", baseName, maxCopy+1)
}
