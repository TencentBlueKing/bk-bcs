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

// Package templatespacecollect 模板文件文件夹收藏
package templatespacecollect

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

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

// TemplateSpaceCollectAction provides the action to manager template space collect
// nolint
type TemplateSpaceCollectAction struct {
	ctx context.Context

	model store.ClusterResourcesModel
}

// NewTemplateSpaceCollectAction return a new TemplateSpaceCollectAction instance
func NewTemplateSpaceCollectAction(model store.ClusterResourcesModel) *TemplateSpaceCollectAction {
	return &TemplateSpaceCollectAction{
		model: model,
	}
}

func (t *TemplateSpaceCollectAction) checkAccess(ctx context.Context) error {
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
func (t *TemplateSpaceCollectAction) List(ctx context.Context) ([]map[string]interface{}, error) {
	if err := t.checkAccess(ctx); err != nil {
		return nil, err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 通过项目编码、用户名称检索
	templateSpaceCollect, err := t.model.ListTemplateSpaceCollect(ctx, "", p.Code, ctxkey.GetUsernameFromCtx(ctx))
	if err != nil {
		return nil, err
	}

	m := make([]map[string]interface{}, 0)
	for _, value := range templateSpaceCollect {
		m = append(m, value.ToMap())
	}
	return m, nil
}

// Create xxx
func (t *TemplateSpaceCollectAction) Create(
	ctx context.Context, req *clusterRes.CreateTemplateSpaceCollectReq) (string, error) {
	if err := t.checkAccess(ctx); err != nil {
		return "", err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return "", err
	}

	username := ctxkey.GetUsernameFromCtx(ctx)
	// 检测用户是否已经收藏
	templateSpaceCollects, err := t.model.ListTemplateSpaceCollect(ctx, req.TemplateSpaceID, p.Code, username)
	if err != nil {
		return "", err
	}

	if len(templateSpaceCollects) > 0 {
		return "", errorx.New(errcode.DuplicationNameErr, i18n.GetMsg(ctx, "该文件夹已被收藏"))
	}

	templateSpaceID, err := primitive.ObjectIDFromHex(req.TemplateSpaceID)
	if err != nil {
		return "", err
	}
	templateSpaceCollect := &entity.TemplateSpaceCollect{
		TemplateSpaceID: templateSpaceID,
		ProjectCode:     p.Code,
		Creator:         username,
	}
	id, err := t.model.CreateTemplateSpaceCollect(ctx, templateSpaceCollect)
	if err != nil {
		return "", err
	}
	return id, nil
}

// Delete xxx
func (t *TemplateSpaceCollectAction) Delete(ctx context.Context, id string) error {
	if err := t.checkAccess(ctx); err != nil {
		return err
	}

	p, err := project.FromContext(ctx)
	if err != nil {
		return err
	}
	// 通过id检索
	templateSpaceCollect, err := t.model.GetTemplateSpaceCollect(ctx, id)
	if err != nil {
		return err
	}

	// 检验更新 TemplateSpace 的权限
	if templateSpaceCollect.ProjectCode != p.Code {
		return errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "无权限访问"))
	}

	if err := t.model.DeleteTemplateSpaceCollect(ctx, id); err != nil {
		return err
	}
	return nil
}
