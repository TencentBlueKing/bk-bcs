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

// Package envmanage environment manage
package envmanage

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/component/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam"
	projectAuth "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm/resource/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// EnvManageAction provides the action to manager envManage
// nolint
type EnvManageAction struct {
	ctx context.Context

	model store.ClusterResourcesModel
}

// NewEnvManageAction return a new EnvManageAction instance
func NewEnvManageAction(model store.ClusterResourcesModel) *EnvManageAction {
	return &EnvManageAction{
		model: model,
	}
}

func (e *EnvManageAction) checkAccess(ctx context.Context) error {
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
		TenantID:  ctxkey.GetTenantIDFromCtx(ctx),
	}
	if allow, err := iam.NewProjectPerm().CanView(permCtx); err != nil {
		return err
	} else if !allow {
		return errorx.New(errcode.NoIAMPerm, i18n.GetMsg(ctx, "无项目查看权限"))
	}
	return nil
}

// Get xxx
func (e *EnvManageAction) Get(ctx context.Context, id, projectCode string) (map[string]interface{}, error) {
	if err := e.checkAccess(ctx); err != nil {
		return nil, err
	}

	envManage, err := e.model.GetEnvManage(ctx, id)
	if err != nil {
		return nil, err
	}

	// 只能查看当前项目环境管理
	if envManage.ProjectCode != projectCode {
		return nil, errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "无权限访问"))
	}

	return envManage.ToMap(), nil
}

// List xxx
func (e *EnvManageAction) List(ctx context.Context) ([]map[string]interface{}, error) {
	if err := e.checkAccess(ctx); err != nil {
		return nil, err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 通过项目编码检索
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode: p.Code,
	})

	envManages, err := e.model.ListEnvManages(ctx, cond)
	if err != nil {
		return nil, err
	}

	m := make([]map[string]interface{}, 0)
	for _, v := range envManages {
		m = append(m, v.ToMap())
	}
	return m, nil
}

// Create xxx
func (e *EnvManageAction) Create(ctx context.Context, req *clusterRes.CreateEnvManageReq) (string, error) {
	if err := e.checkAccess(ctx); err != nil {
		return "", err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return "", err
	}

	// 检测同名
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode: p.Code,
		entity.FieldKeyEnv:         req.GetEnv(),
	})
	envManages, err := e.model.ListEnvManages(ctx, cond)
	if err != nil {
		return "", err
	}

	if len(envManages) > 0 {
		return "", errorx.New(errcode.DuplicationNameErr, i18n.GetMsg(ctx, "环境名称已存在"))
	}

	envManage := &entity.EnvManage{
		Env:               req.GetEnv(),
		ProjectCode:       p.Code,
		ClusterNamespaces: protoClusterNamespacesToEntity(req.GetClusterNamespaces()),
	}
	id, err := e.model.CreateEnvManage(ctx, envManage)
	if err != nil {
		return "", err
	}
	return id, nil
}

// Update xxx
func (e *EnvManageAction) Update(ctx context.Context, req *clusterRes.UpdateEnvManageReq) error {
	if err := e.checkAccess(ctx); err != nil {
		return err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return err
	}

	envManage, err := e.model.GetEnvManage(ctx, req.GetId())
	if err != nil {
		return err
	}

	// 检验更新 envManage 的权限
	if envManage.ProjectCode != p.Code {
		return errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "无权限访问"))
	}

	// 检测环境名称是否重复
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyEnv:         req.GetEnv(),
		entity.FieldKeyProjectCode: p.Code,
	})
	envManages, err := e.model.ListEnvManages(ctx, cond)
	if err != nil {
		return err
	}

	// 存在同一个projectCode的环境名称则不能更新
	if len(envManages) > 0 && envManages[0].ID.Hex() != req.GetId() {
		return errorx.New(errcode.DuplicationNameErr, i18n.GetMsg(ctx, "环境名称已存在"))
	}

	updateEnvManage := entity.M{
		"clusterNamespaces": protoClusterNamespacesToEntity(req.GetClusterNamespaces()),
		"env":               req.GetEnv(),
	}
	if err := e.model.UpdateEnvManage(ctx, req.GetId(), updateEnvManage); err != nil {
		return err
	}
	return nil
}

// Rename xxx
func (e *EnvManageAction) Rename(ctx context.Context, req *clusterRes.RenameEnvManageReq) error {
	if err := e.checkAccess(ctx); err != nil {
		return err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return err
	}

	envManage, err := e.model.GetEnvManage(ctx, req.GetId())
	if err != nil {
		return err
	}

	// 检验更新 envManage 的权限
	if envManage.ProjectCode != p.Code {
		return errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "无权限访问"))
	}

	// 检测环境名称是否重复
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyEnv:         req.GetEnv(),
		entity.FieldKeyProjectCode: p.Code,
	})
	envManages, err := e.model.ListEnvManages(ctx, cond)
	if err != nil {
		return err
	}

	// 存在同一个projectCode的环境名称则不能更新
	if len(envManages) > 0 && envManages[0].ID.Hex() != req.GetId() {
		return errorx.New(errcode.DuplicationNameErr, i18n.GetMsg(ctx, "环境名称已存在"))
	}

	updateEnvManage := entity.M{
		"env": req.Env,
	}
	if err = e.model.UpdateEnvManage(ctx, req.GetId(), updateEnvManage); err != nil {
		return err
	}
	return nil
}

// Delete xxx
func (e *EnvManageAction) Delete(ctx context.Context, id string) error {
	if err := e.checkAccess(ctx); err != nil {
		return err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return err
	}

	envManage, err := e.model.GetEnvManage(ctx, id)
	if err != nil {
		return err
	}

	// 检验更新 envManage 的权限
	if envManage.ProjectCode != p.Code {
		return errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "无权限访问"))
	}

	if err = e.model.DeleteEnvManage(ctx, id); err != nil {
		return err
	}
	return nil
}

// protoClusterNamespacesToEntity 转换关联命名空间
func protoClusterNamespacesToEntity(ns []*clusterRes.ClusterNamespaces) []entity.ClusterNamespaces {
	result := make([]entity.ClusterNamespaces, 0)
	for _, v := range ns {
		result = append(result, entity.ClusterNamespaces{
			ClusterID:  v.ClusterID,
			Namespaces: v.Namespaces})
	}
	return result
}
