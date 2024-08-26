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
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbapp "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/app"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/thirdparty/esb/cmdb"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/version"
)

// CreateApp create application.
func (s *Service) CreateApp(ctx context.Context, req *pbds.CreateAppReq) (*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	// validate biz exist when user is not for test
	if !strings.HasPrefix(kt.User, constant.BKUserForTestPrefix) {
		if err := s.validateBizExist(kt, req.BizId); err != nil {
			logs.Errorf("validate biz exist failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	if _, err := s.dao.App().GetByName(kt, req.BizId, req.Spec.Name); err == nil {
		return nil, errf.Errorf(errf.InvalidRequest, i18n.T(kt, "app name %s already exists", req.Spec.Name))
	}

	if _, err := s.dao.App().GetByAlias(kt, req.BizId, req.Spec.Alias); err == nil {
		return nil, errf.Errorf(errf.InvalidRequest, i18n.T(kt, "app alias %s already exists", req.Spec.Alias))
	}

	app := &table.App{
		BizID: req.BizId,
		Spec:  req.Spec.AppSpec(),
		Revision: &table.Revision{
			Creator: kt.User,
			Reviser: kt.User,
		},
	}

	id, err := s.dao.App().Create(kt, app)
	if err != nil {
		logs.Errorf("create app failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// UpdateApp update application.
func (s *Service) UpdateApp(ctx context.Context, req *pbds.UpdateAppReq) (*pbapp.App, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	old, err := s.dao.App().GetByAlias(grpcKit, req.BizId, req.Spec.Alias)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logs.Errorf("get app failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, errf.Errorf(errf.DBOpFailed, i18n.T(grpcKit, "get app failed, err: %v", err))
	}
	if !errors.Is(gorm.ErrRecordNotFound, err) && old.ID != req.Id {
		return nil, errf.Errorf(errf.InvalidRequest, "app alias %s already exists", req.Spec.Alias)
	}

	app, err := s.dao.App().Get(grpcKit, req.BizId, req.Id)
	if err != nil {
		logs.Errorf("get app failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, errf.Errorf(errf.DBOpFailed, i18n.T(grpcKit, "get app failed, err: %v", err))
	}
	if app.Spec.ConfigType == table.KV {
		if e := s.checkUpdateAppDataType(grpcKit, req, app); e != nil {
			return nil, e
		}
	}

	app = &table.App{
		ID:    req.Id,
		BizID: req.BizId,
		Spec:  req.Spec.AppSpec(),
		Revision: &table.Revision{
			Reviser: grpcKit.User,
		},
	}
	if err = s.dao.App().Update(grpcKit, app); err != nil {
		logs.Errorf("update app failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	app, err = s.dao.App().Get(grpcKit, req.BizId, req.Id)
	if err != nil {
		logs.Errorf("updating the app was successful, but retrieving the app failed, err: %v, rid: %s",
			err, grpcKit.Rid)
		return nil, err
	}

	return pbapp.PbApp(app), nil
}

func (s *Service) checkUpdateAppDataType(kt *kit.Kit, req *pbds.UpdateAppReq, app *table.App) error {

	if app.Spec.DataType == table.DataType(req.Spec.DataType) {
		return nil
	}

	if req.Spec.DataType == string(table.KvAny) {
		return nil
	}

	// 获取所有的kv
	kvState := []string{
		string(table.KvStateAdd),
		string(table.KvStateRevise),
		string(table.KvStateUnchange),
		string(table.KvStateDelete),
	}
	kvList, err := s.dao.Kv().ListAllByAppID(kt, app.ID, req.BizId, kvState)
	if err != nil {
		return err
	}
	if len(kvList) == 0 {
		return nil
	}

	for _, kv := range kvList {
		kvType, _, err := s.getKv(kt, req.BizId, kv.Attachment.AppID, kv.Spec.Version, kv.Spec.Key)
		if err != nil {
			return err
		}

		if string(kvType) != req.Spec.DataType {
			return errf.Errorf(errf.InvalidArgument, i18n.T(kt, "the specified type does not match the actual configuration"))
		}
	}

	return nil
}

// DeleteApp delete application.
func (s *Service) DeleteApp(ctx context.Context, req *pbds.DeleteAppReq) (*pbbase.EmptyResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	app := &table.App{
		ID:    req.Id,
		BizID: req.BizId,
	}

	tx := s.dao.GenQuery().Begin()

	// 1. delete app related resources
	if err := s.deleteAppRelatedResources(grpcKit, req, tx); err != nil {
		logs.Errorf("delete app related resources failed, err: %v, rid: %s", err, grpcKit.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, errf.Errorf(errf.DBOpFailed,
			i18n.T(grpcKit, "delete app related resources failed, err: %v", err))
	}

	// 2. delete app
	if err := s.dao.App().DeleteWithTx(grpcKit, tx, app); err != nil {
		logs.Errorf("delete app failed, err: %v, rid: %s", err, grpcKit.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, errf.Errorf(errf.DBOpFailed,
			i18n.T(grpcKit, "delete app failed, err: %v", err))
	}

	if err := tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, errf.Errorf(errf.DBOpFailed,
			i18n.T(grpcKit, "delete app failed, err: %v", err))
	}

	return new(pbbase.EmptyResp), nil
}

func (s *Service) deleteAppRelatedResources(grpcKit *kit.Kit, req *pbds.DeleteAppReq, tx *gen.QueryTx) error {
	// delete app template binding
	if err := s.dao.AppTemplateBinding().DeleteByAppIDWithTx(grpcKit, tx, req.GetBizId(), req.Id); err != nil {
		logs.Errorf("delete app template binding failed, err: %v, rid: %s", err, grpcKit.Rid)
		return err
	}

	// delete group app binding
	if err := s.dao.GroupAppBind().BatchDeleteByAppIDWithTx(grpcKit, tx, req.Id, req.BizId); err != nil {
		logs.Errorf("delete group app binding failed, err: %v, rid: %s", err, grpcKit.Rid)
		return err
	}

	// delete released group
	if err := s.dao.ReleasedGroup().BatchDeleteByAppIDWithTx(grpcKit, tx, req.Id, req.BizId); err != nil {
		logs.Errorf("delete group app binding failed, err: %v, rid: %s", err, grpcKit.Rid)
		return err
	}

	// delete app template binding
	if err := s.dao.ReleasedAppTemplate().BatchDeleteByAppIDWithTx(grpcKit, tx, req.Id, req.BizId); err != nil {
		logs.Errorf("delete released app template failed, err: %v, rid: %s", err, grpcKit.Rid)
		return err
	}

	// delete released app template binding
	if err := s.dao.ReleasedAppTemplate().BatchDeleteByAppIDWithTx(grpcKit, tx, req.Id, req.BizId); err != nil {
		logs.Errorf("delete released app template failed, err: %v, rid: %s", err, grpcKit.Rid)
		return err
	}

	// delete released app template variables
	if err := s.dao.ReleasedAppTemplateVariable().BatchDeleteByAppIDWithTx(grpcKit, tx, req.Id, req.BizId); err != nil {
		logs.Errorf("delete released app template variables failed, err: %v, rid: %s", err, grpcKit.Rid)
		return err
	}

	// delete released hook
	if err := s.dao.ReleasedHook().DeleteByAppIDWithTx(grpcKit, tx, req.Id, req.BizId); err != nil {
		logs.Errorf("delete released hooks failed, err: %v, rid: %s", err, grpcKit.Rid)
		return err
	}

	// delete related credential scopes and update credentials
	if err := s.updateRelatedCredentials(grpcKit, tx, req.Id, req.BizId); err != nil {
		return err
	}

	return nil
}

// updateRelatedCredentials delete related credential scopes and update credentials to emit event.
func (s *Service) updateRelatedCredentials(grpcKit *kit.Kit, tx *gen.QueryTx, appID, bizID uint32) error {
	app, err := s.dao.App().Get(grpcKit, bizID, appID)
	if err != nil {
		logs.Errorf("get app failed, err: %v, rid: %s", err, grpcKit.Rid)
		return err
	}
	matchedScopeIDs := make([]uint32, 0)
	matchedCredentialIDs := make([]uint32, 0)
	// delete related credential scopes
	scopes, err := s.dao.CredentialScope().ListAll(grpcKit, bizID)
	if err != nil {
		return err
	}
	for _, scope := range scopes {
		appName, _, _ := scope.Spec.CredentialScope.Split()
		if appName == app.Spec.Name {
			matchedScopeIDs = append(matchedScopeIDs, scope.ID)
			matchedCredentialIDs = append(matchedCredentialIDs, scope.Attachment.CredentialId)
		}
	}
	if e := s.dao.CredentialScope().BatchDeleteWithTx(grpcKit, tx, bizID, matchedScopeIDs); e != nil {
		logs.Errorf("delete credential scopes failed, err: %v, rid: %s", e, grpcKit.Rid)
		return e
	}
	// update credentials
	matchedCredentialIDs = tools.RemoveDuplicates(matchedCredentialIDs)
	for _, credentialID := range matchedCredentialIDs {
		if e := s.dao.Credential().UpdateRevisionWithTx(grpcKit, tx, bizID, credentialID); e != nil {
			logs.Errorf("update credential revision failed, err: %v, rid: %s", e, grpcKit.Rid)
			return e
		}
	}

	return nil
}

// GetApp get apps by app id.
func (s *Service) GetApp(ctx context.Context, req *pbds.GetAppReq) (*pbapp.App, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	app, err := s.dao.App().Get(grpcKit, req.BizId, req.AppId)
	if err != nil {
		logs.Errorf("get app failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return pbapp.PbApp(app), nil
}

// GetAppByID get apps by only by app id.
func (s *Service) GetAppByID(ctx context.Context, req *pbds.GetAppByIDReq) (*pbapp.App, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	app, err := s.dao.App().GetByID(grpcKit, req.GetAppId())
	if err != nil {
		logs.Errorf("get app by id failed, err: %v, rid: %s", err, grpcKit.Rid)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errf.Errorf(errf.AppNotExists, i18n.T(grpcKit, "app %d not found", req.AppId))
		}
		return nil, errors.Wrapf(err, "query app by id %d", req.GetAppId())
	}

	return pbapp.PbApp(app), nil
}

// GetAppByName get app by app name.
func (s *Service) GetAppByName(ctx context.Context, req *pbds.GetAppByNameReq) (*pbapp.App, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	app, err := s.dao.App().GetByName(grpcKit, req.GetBizId(), req.GetAppName())
	if err != nil {
		logs.Errorf("get app by name failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, errors.Wrapf(err, "query app by name %s failed", req.GetAppName())
	}

	return pbapp.PbApp(app), nil
}

// ListAppsRest list apps by query condition.
func (s *Service) ListAppsRest(ctx context.Context, req *pbds.ListAppsRestReq) (*pbds.ListAppsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	// 默认分页
	limit := uint(req.Limit)
	if limit == 0 {
		limit = 50
	}

	opt := &types.BasePage{
		Start: req.Start,
		Limit: limit,
		All:   req.All,
	}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	bizList, err := tools.GetUint32List(req.BizId)
	if err != nil {
		return nil, err
	}
	if len(bizList) == 0 {
		return nil, fmt.Errorf("bizList is empty")
	}

	details, count, err := s.dao.App().List(kt, bizList, req.Name, req.Operator, opt)
	if err != nil {
		logs.Errorf("list apps failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListAppsResp{
		Count:   uint32(count),
		Details: pbapp.PbApps(details),
	}
	return resp, nil
}

// ListAppsByIDs list apps by query condition.
func (s *Service) ListAppsByIDs(ctx context.Context, req *pbds.ListAppsByIDsReq) (*pbds.ListAppsByIDsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if len(req.Ids) == 0 {
		return nil, fmt.Errorf("app ids is empty")
	}

	details, err := s.dao.App().ListAppsByIDs(kt, req.Ids)
	if err != nil {
		logs.Errorf("list apps failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListAppsByIDsResp{
		Details: pbapp.PbApps(details),
	}
	return resp, nil
}

// validateBizExist validate if biz exists in cmdb before create app.
func (s *Service) validateBizExist(kt *kit.Kit, bizID uint32) error {
	// if build version is debug mode, not need to validate biz exist in cmdb.
	if version.Debug() {
		return nil
	}

	searchBizParams := &cmdb.SearchBizParams{
		Fields: []string{"bk_biz_id"},
		Page:   cmdb.BasePage{Limit: 1},
		BizPropertyFilter: &cmdb.QueryFilter{
			Rule: cmdb.CombinedRule{
				Condition: cmdb.ConditionAnd,
				Rules: []cmdb.Rule{
					cmdb.AtomRule{
						Field:    cmdb.BizIDField,
						Operator: cmdb.OperatorEqual,
						Value:    bizID,
					}},
			}},
	}

	bizResp, err := s.esb.Cmdb().SearchBusiness(kt.Ctx, searchBizParams)
	if err != nil {
		return errf.Errorf(errf.InvalidRequest, i18n.T(kt, "business query failed, err: %v", err))
	}

	if bizResp.Count == 0 {
		return errf.Errorf(errf.RelatedResNotExist, i18n.T(kt, "app related biz %d is not exist", bizID))
	}

	return nil
}
