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

package service

import (
	"context"
	"fmt"
	"path"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbcommit "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/commit"
	pbci "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/config-item"
	pbcontent "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/content"
	pbrci "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/released-ci"
	pbtv "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/template-variable"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
)

// CreateConfigItem create config item with option
func (s *Service) CreateConfigItem(ctx context.Context, req *pbcs.CreateConfigItemReq) (
	*pbcs.CreateConfigItemResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Update, ResourceID: req.AppId}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}
	// 1. validate if file content uploaded.
	metadata, err := s.client.provider.Metadata(grpcKit, req.Sign)
	if err != nil {
		logs.Errorf("validate file content uploaded failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	// 2. insert config item, content and commit to db.
	cciReq := &pbds.CreateConfigItemReq{
		ConfigItemAttachment: &pbci.ConfigItemAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
		ConfigItemSpec: &pbci.ConfigItemSpec{
			Name:     req.Name,
			Path:     req.Path,
			FileType: req.FileType,
			FileMode: req.FileMode,
			Memo:     req.Memo,
			Permission: &pbci.FilePermission{
				User:      req.User,
				UserGroup: req.UserGroup,
				Privilege: req.Privilege,
			},
		},
		ContentSpec: &pbcontent.ContentSpec{
			Signature: req.Sign,
			Md5:       metadata.Md5,
			ByteSize:  req.ByteSize,
		},
	}
	cciResp, err := s.client.DS.CreateConfigItem(grpcKit.RpcCtx(), cciReq)
	if err != nil {
		logs.Errorf("create config item failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.CreateConfigItemResp{
		Id: cciResp.Id,
	}

	return resp, nil
}

// BatchUpsertConfigItems batch upsert config items with option
func (s *Service) BatchUpsertConfigItems(ctx context.Context, req *pbcs.BatchUpsertConfigItemsReq) (
	*pbcs.BatchUpsertConfigItemsResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.BatchUpsertConfigItemsResp)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Update, ResourceID: req.AppId}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	items := make([]*pbds.BatchUpsertConfigItemsReq_ConfigItem, 0, len(req.Items))
	for _, item := range req.Items {
		// validate if file content uploaded.
		metadata, err := s.client.provider.Metadata(grpcKit, item.Sign)
		if err != nil {
			logs.Errorf("validate file content uploaded failed, err: %v, rid: %s", err, grpcKit.Rid)
			return nil, err
		}
		vars := make([]*pbtv.TemplateVariableSpec, 0, len(item.Variables))
		for _, v := range item.GetVariables() {
			vars = append(vars, &pbtv.TemplateVariableSpec{
				Name:       v.Name,
				Type:       v.Type,
				DefaultVal: v.DefaultVal,
				Memo:       v.Memo,
			})
		}

		items = append(items, &pbds.BatchUpsertConfigItemsReq_ConfigItem{
			ConfigItemAttachment: &pbci.ConfigItemAttachment{
				BizId: req.BizId,
				AppId: req.AppId,
			},
			ConfigItemSpec: &pbci.ConfigItemSpec{
				Name:     item.Name,
				Path:     item.Path,
				FileType: item.FileType,
				FileMode: item.FileMode,
				Memo:     item.Memo,
				Permission: &pbci.FilePermission{
					User:      item.User,
					UserGroup: item.UserGroup,
					Privilege: item.Privilege,
				},
			},
			ContentSpec: &pbcontent.ContentSpec{
				Signature: item.Sign,
				ByteSize:  item.ByteSize,
				Md5:       metadata.Md5,
			},
			Variables: vars,
		})
	}

	buReq := &pbds.BatchUpsertConfigItemsReq{
		BizId:      req.BizId,
		AppId:      req.AppId,
		Items:      items,
		ReplaceAll: req.ReplaceAll,
	}
	batchUpsertConfigResp, e := s.client.DS.BatchUpsertConfigItems(grpcKit.RpcCtx(), buReq)
	if e != nil {
		logs.Errorf("batch upsert config item failed, err: %v, rid: %s", e, grpcKit.Rid)
		return nil, e
	}
	resp.Ids = batchUpsertConfigResp.Ids
	return resp, nil
}

// UpdateConfigItem update config item with option
// nolint: funlen
func (s *Service) UpdateConfigItem(ctx context.Context, req *pbcs.UpdateConfigItemReq) (
	*pbcs.UpdateConfigItemResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.UpdateConfigItemResp)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Update, ResourceID: req.AppId}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	// 1. update config_item
	r := &pbds.UpdateConfigItemReq{
		Id: req.Id,
		Attachment: &pbci.ConfigItemAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
		Spec: &pbci.ConfigItemSpec{
			Name:     req.Name,
			Path:     req.Path,
			FileType: req.FileType,
			FileMode: req.FileMode,
			Memo:     req.Memo,
			Permission: &pbci.FilePermission{
				User:      req.User,
				UserGroup: req.UserGroup,
				Privilege: req.Privilege,
			},
		},
	}
	_, err = s.client.DS.UpdateConfigItem(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("update config item failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// 2. check if content sign changed,if changed,create content and commit
	// 2.1. get latest commit and compare content sign
	glcReq := &pbds.GetLatestCommitReq{
		BizId:        req.BizId,
		AppId:        req.AppId,
		ConfigItemId: req.Id,
	}
	glcResp, err := s.client.DS.GetLatestCommit(grpcKit.RpcCtx(), glcReq)
	if err != nil {
		logs.Errorf("get config item latest commit failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// 2.2 if latest content sign equals request content sign,no need to commit
	if glcResp.Spec.Content.Signature == req.Sign {
		return resp, nil
	}

	// 2.3 if latest content sign not equals request content sign,create content and commit
	// validate if file content uploaded.
	metadata, err := s.client.provider.Metadata(grpcKit, req.Sign)
	if err != nil {
		logs.Errorf("validate file content uploaded failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	ccReq := &pbds.CreateContentReq{
		Attachment: &pbcontent.ContentAttachment{
			ConfigItemId: req.Id,
			BizId:        req.BizId,
			AppId:        req.AppId,
		},
		Spec: &pbcontent.ContentSpec{
			Signature: req.Sign,
			ByteSize:  req.ByteSize,
			Md5:       metadata.Md5,
		},
	}
	ccResp, err := s.client.DS.CreateContent(grpcKit.RpcCtx(), ccReq)
	if err != nil {
		logs.Errorf("create config item failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// 4. create commit
	ccmReq := &pbds.CreateCommitReq{
		Attachment: &pbcommit.CommitAttachment{
			BizId:        req.BizId,
			AppId:        req.AppId,
			ConfigItemId: req.Id,
		},
		ContentId: ccResp.Id,
		Memo:      req.Memo,
	}
	_, err = s.client.DS.CreateCommit(grpcKit.RpcCtx(), ccmReq)
	if err != nil {
		logs.Errorf("create config item failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	return resp, nil
}

// DeleteConfigItem delete config item with option
func (s *Service) DeleteConfigItem(ctx context.Context, req *pbcs.DeleteConfigItemReq) (
	*pbcs.DeleteConfigItemResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteConfigItemResp)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Update, ResourceID: req.AppId}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	r := &pbds.DeleteConfigItemReq{
		Id: req.Id,
		Attachment: &pbci.ConfigItemAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
	}
	_, err = s.client.DS.DeleteConfigItem(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("delete config item failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// BatchDeleteConfigItems is used to batch delete config items.
func (s *Service) BatchDeleteConfigItems(ctx context.Context, req *pbcs.BatchDeleteAppResourcesReq) (
	*pbcs.BatchDeleteResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Update, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	if len(req.GetIds()) == 0 {
		return nil, errf.Errorf(errf.InvalidArgument, i18n.T(grpcKit, "id is required"))
	}

	eg, egCtx := errgroup.WithContext(grpcKit.RpcCtx())
	eg.SetLimit(10)

	successfulIDs := []uint32{}
	failedIDs := []uint32{}
	var mux sync.Mutex

	// 使用 data-service 原子接口
	for _, v := range req.GetIds() {
		v := v
		eg.Go(func() error {
			r := &pbds.DeleteConfigItemReq{
				Id: v,
				Attachment: &pbci.ConfigItemAttachment{
					BizId: req.BizId,
					AppId: req.AppId,
				},
			}
			if _, err := s.client.DS.DeleteConfigItem(egCtx, r); err != nil {
				logs.Errorf("delete config item %d failed, err: %v, rid: %s", v, err, grpcKit.Rid)

				// 错误不返回异常，记录错误ID
				mux.Lock()
				failedIDs = append(failedIDs, v)
				mux.Unlock()
				return nil
			}

			mux.Lock()
			successfulIDs = append(successfulIDs, v)
			mux.Unlock()

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		logs.Errorf("batch delete config items failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, errf.Errorf(errf.Aborted, i18n.T(grpcKit, "batch delete config items failed"))
	}

	// 全部失败, 当前API视为失败
	if len(failedIDs) == len(req.Ids) {
		return nil, errf.Errorf(errf.Aborted, i18n.T(grpcKit, "batch delete config items failed"))
	}

	return &pbcs.BatchDeleteResp{SuccessfulIds: successfulIDs, FailedIds: failedIDs}, nil
}

// GetConfigItem get config item with content
func (s *Service) GetConfigItem(ctx context.Context, req *pbcs.GetConfigItemReq) (
	*pbcs.GetConfigItemResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	return s.getEditingConfigItem(grpcKit, req.ConfigItemId, grpcKit.BizID, req.AppId)
}

// getEditingConfigItem get edit config item
func (s *Service) getEditingConfigItem(grpcKit *kit.Kit, configItemID, bizID, appID uint32) (
	*pbcs.GetConfigItemResp, error) {
	// 1. get config item
	gciReq := &pbds.GetConfigItemReq{
		Id:    configItemID,
		BizId: bizID,
		AppId: appID,
	}
	gciResp, err := s.client.DS.GetConfigItem(grpcKit.RpcCtx(), gciReq)
	if err != nil {
		logs.Errorf("get config item failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// 2. get latest commit
	glcReq := &pbds.GetLatestCommitReq{
		BizId:        bizID,
		AppId:        appID,
		ConfigItemId: configItemID,
	}
	glcResp, err := s.client.DS.GetLatestCommit(grpcKit.RpcCtx(), glcReq)
	if err != nil {
		logs.Errorf("get config item latest commit failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// 3. get content
	gcReq := &pbds.GetContentReq{
		Id:    glcResp.Spec.ContentId,
		BizId: bizID,
		AppId: appID,
	}
	gcResp, err := s.client.DS.GetContent(grpcKit.RpcCtx(), gcReq)
	if err != nil {
		logs.Errorf("get config item content failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.GetConfigItemResp{
		ConfigItem: gciResp,
		Content:    gcResp.Spec,
	}
	return resp, nil
}

// GetReleasedConfigItem get released config item
func (s *Service) GetReleasedConfigItem(ctx context.Context, req *pbcs.GetReleasedConfigItemReq) (
	*pbcs.GetReleasedConfigItemResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	grciReq := &pbds.GetReleasedCIReq{
		BizId:        req.BizId,
		AppId:        req.AppId,
		ReleaseId:    req.ReleaseId,
		ConfigItemId: req.ConfigItemId,
	}
	releasedCI, err := s.client.DS.GetReleasedConfigItem(grpcKit.RpcCtx(), grciReq)
	if err != nil {
		logs.Errorf("get released config item failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.GetReleasedConfigItemResp{
		ConfigItem: releasedCI,
	}
	return resp, nil

}

// ListConfigItems list config item with filter
func (s *Service) ListConfigItems(ctx context.Context, req *pbcs.ListConfigItemsReq) (
	*pbcs.ListConfigItemsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	// Note: list the latest release and compare each config item exists and latest commit id to get changing status
	r := &pbds.ListConfigItemsReq{
		BizId:        grpcKit.BizID,
		AppId:        req.AppId,
		SearchFields: req.SearchFields,
		SearchValue:  req.SearchValue,
		Start:        req.Start,
		Limit:        req.Limit,
		All:          req.All,
		Ids:          req.Ids,
		WithStatus:   req.WithStatus,
		Status:       req.Status,
	}
	rp, err := s.client.DS.ListConfigItems(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list config items failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// 对比模板配置, 检测是否存在冲突
	trc, err := s.client.DS.ListAppBoundTmplRevisions(grpcKit.RpcCtx(), &pbds.ListAppBoundTmplRevisionsReq{
		BizId: grpcKit.BizID,
		AppId: req.AppId,
		All:   true,
	})
	if err != nil {
		logs.Errorf("list app template revisions failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	existingPaths := []string{}
	for _, v := range rp.GetDetails() {
		if v.FileState != constant.FileStateDelete {
			existingPaths = append(existingPaths, path.Join(v.Spec.Path, v.Spec.Name))
		}
	}
	for _, v := range trc.GetDetails() {
		if v.FileState != constant.FileStateDelete {
			existingPaths = append(existingPaths, path.Join(v.Path, v.Name))
		}
	}

	conflictNums, conflictPaths := checkExistingPathConflict(existingPaths)
	for _, v := range rp.GetDetails() {
		if v.FileState != constant.FileStateDelete {
			v.IsConflict = conflictPaths[path.Join(v.Spec.Path, v.Spec.Name)]
		}
	}

	resp := &pbcs.ListConfigItemsResp{
		Count:          rp.Count,
		Details:        rp.Details,
		ConflictNumber: conflictNums,
	}
	return resp, nil
}

// ListReleasedConfigItems list released config items
func (s *Service) ListReleasedConfigItems(ctx context.Context,
	req *pbcs.ListReleasedConfigItemsReq) (
	*pbcs.ListReleasedConfigItemsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	if req.ReleaseId <= 0 {
		return nil, fmt.Errorf("invalid release id %d, it must bigger than 0", req.ReleaseId)
	}

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListReleasedConfigItemsReq{
		BizId:        req.BizId,
		AppId:        req.AppId,
		ReleaseId:    req.ReleaseId,
		SearchFields: req.SearchFields,
		SearchValue:  req.SearchValue,
		Start:        req.Start,
		Limit:        req.Limit,
		All:          true,
	}

	rp, err := s.client.DS.ListReleasedConfigItems(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list released config items failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListReleasedConfigItemsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListConfigItemCount get config item count number
func (s *Service) ListConfigItemCount(ctx context.Context, req *pbcs.ListConfigItemCountReq) (
	*pbcs.ListConfigItemCountResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	r := &pbds.ListConfigItemCountReq{
		BizId: req.BizId,
		AppId: req.AppId,
	}
	rp, err := s.client.DS.ListConfigItemCount(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list config items failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListConfigItemCountResp{
		Details: rp.Details,
	}
	return resp, nil
}

// ListConfigItemByTuple 按照多个字段in查询
func (s *Service) ListConfigItemByTuple(ctx context.Context, req *pbcs.ListConfigItemByTupleReq) (
	*pbcs.ListConfigItemByTupleResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	data := []*pbds.ListConfigItemByTupleReq_Item{}

	for _, item := range req.Items {
		data = append(data, &pbds.ListConfigItemByTupleReq_Item{
			BizId: req.BizId,
			AppId: req.AppId,
			Name:  item.Name,
			Path:  item.Path,
		})
	}

	tuple, err := s.client.DS.ListConfigItemByTuple(grpcKit.RpcCtx(), &pbds.ListConfigItemByTupleReq{
		Items: data,
	})
	if err != nil {
		return nil, err
	}
	resp := &pbcs.ListConfigItemByTupleResp{Details: tuple.GetConfigItems()}
	return resp, nil
}

// UnDeleteConfigItem 配置项未命名版本恢复
func (s *Service) UnDeleteConfigItem(ctx context.Context, req *pbcs.UnDeleteConfigItemReq) (
	*pbcs.UnDeleteConfigItemResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Update, ResourceID: req.AppId}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	_, err = s.client.DS.UnDeleteConfigItem(grpcKit.RpcCtx(), &pbds.UnDeleteConfigItemReq{
		Id: req.Id,
		Attachment: &pbci.ConfigItemAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
	})
	if err != nil {
		return nil, err
	}

	return &pbcs.UnDeleteConfigItemResp{}, nil
}

// UndoConfigItem 撤消配置项
func (s *Service) UndoConfigItem(ctx context.Context, req *pbcs.UndoConfigItemReq) (*pbcs.UndoConfigItemResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Update, ResourceID: req.AppId}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	_, err = s.client.DS.UndoConfigItem(grpcKit.RpcCtx(), &pbds.UndoConfigItemReq{
		Id: req.Id,
		Attachment: &pbci.ConfigItemAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
	})
	if err != nil {
		return nil, err
	}
	return &pbcs.UndoConfigItemResp{}, nil
}

// checkExistingPathConflict Check existing path collections for conflicts.
func checkExistingPathConflict(existing []string) (uint32, map[string]bool) {
	conflictPaths := make(map[string]bool, len(existing))
	var conflictNums uint32
	conflictMap := make(map[string]bool, 0)
	// 遍历每一个路径
	for i := 0; i < len(existing); i++ {
		// 检查当前路径与后续路径之间是否存在冲突
		for j := i + 1; j < len(existing); j++ {
			if strings.HasPrefix(existing[j]+"/", existing[i]+"/") || strings.HasPrefix(existing[i]+"/", existing[j]+"/") {
				// 相等也算冲突
				if len(existing[j]) == len(existing[i]) {
					conflictNums++
				} else if len(existing[j]) < len(existing[i]) {
					conflictMap[existing[j]] = true
				} else {
					conflictMap[existing[i]] = true
				}

				conflictPaths[existing[i]] = true
				conflictPaths[existing[j]] = true
			}
		}
	}

	return uint32(len(conflictMap)) + conflictNums, conflictPaths
}

// CompareConfigItemConflicts compare config item version conflicts
func (s *Service) CompareConfigItemConflicts(ctx context.Context, req *pbcs.CompareConfigItemConflictsReq) ( // nolint
	*pbcs.CompareConfigItemConflictsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Update, ResourceID: req.AppId}, BizID: req.BizId},
	}

	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	// 获取该服务未发布的版本
	ci, err := s.client.DS.ListConfigItems(grpcKit.RpcCtx(), &pbds.ListConfigItemsReq{
		BizId:      grpcKit.BizID,
		AppId:      req.AppId,
		All:        true,
		WithStatus: true,
		Status:     []string{constant.FileStateAdd, constant.FileStateRevise, constant.FileStateUnchange},
	})
	if err != nil {
		logs.Errorf("list config items failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// 从服务获取发布的版本
	rci, err := s.client.DS.ListReleasedConfigItems(grpcKit.RpcCtx(), &pbds.ListReleasedConfigItemsReq{
		BizId:     req.BizId,
		AppId:     req.OtherAppId,
		ReleaseId: req.ReleaseId,
		All:       true,
	})
	if err != nil {
		logs.Errorf("list released config items failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	conflicts := make(map[string]bool)
	for _, v := range ci.GetDetails() {
		conflicts[path.Join(v.Spec.Path, v.Spec.Name)] = true
	}

	// 获取已生成版本的引用变量
	resf, err := s.client.DS.GetReleasedAppTmplVariableRefs(grpcKit.RpcCtx(), &pbds.GetReleasedAppTmplVariableRefsReq{
		BizId:     req.BizId,
		AppId:     req.OtherAppId,
		ReleaseId: req.ReleaseId,
	})
	if err != nil {
		return nil, err
	}
	resfMap := make(map[string][]string, 0)
	for _, v := range resf.GetDetails() {
		for _, ref := range v.GetReferences() {
			filePath := path.Join(ref.Path, ref.Name)
			resfMap[filePath] = append(resfMap[filePath], v.GetVariableName())
		}
	}

	// 获取已生成版本的变量值、描述等
	vars, err := s.client.DS.ListReleasedAppTmplVariables(grpcKit.RpcCtx(), &pbds.ListReleasedAppTmplVariablesReq{
		BizId:     req.BizId,
		AppId:     req.OtherAppId,
		ReleaseId: req.ReleaseId,
	})
	if err != nil {
		return nil, err
	}

	varsMap := make(map[string][]*pbcs.CompareConfigItemConflictsResp_Variable, 0)
	for _, v := range vars.GetDetails() {
		for key, name := range resfMap {
			for _, n := range name {
				if v.Name == n {
					varsMap[key] = append(varsMap[key], &pbcs.CompareConfigItemConflictsResp_Variable{
						Name:       n,
						Type:       v.Type,
						DefaultVal: v.DefaultVal,
						Memo:       v.Memo,
					})
				}
			}
		}
	}

	newConfigItem := func(v *pbrci.ReleasedConfigItem) *pbcs.CompareConfigItemConflictsResp_ConfigItem {
		return &pbcs.CompareConfigItemConflictsResp_ConfigItem{
			Id:        v.Id,
			Name:      v.Spec.Name,
			Path:      v.Spec.Path,
			FileType:  v.Spec.FileType,
			FileMode:  v.Spec.FileMode,
			Memo:      v.Spec.Memo,
			User:      v.Spec.Permission.User,
			UserGroup: v.Spec.Permission.UserGroup,
			Privilege: v.Spec.Permission.Privilege,
			Sign:      v.CommitSpec.Content.OriginSignature,
			ByteSize:  v.CommitSpec.Content.OriginByteSize,
			Variables: varsMap[path.Join(v.Spec.Path, v.Spec.Name)],
		}
	}

	exist := make([]*pbcs.CompareConfigItemConflictsResp_ConfigItem, 0)
	nonExist := make([]*pbcs.CompareConfigItemConflictsResp_ConfigItem, 0)
	for _, v := range rci.GetDetails() {
		if conflicts[path.Join(v.Spec.Path, v.Spec.Name)] {
			exist = append(exist, newConfigItem(v))
		} else {
			nonExist = append(nonExist, newConfigItem(v))
		}
	}

	return &pbcs.CompareConfigItemConflictsResp{Exist: exist, NonExist: nonExist}, nil
}
