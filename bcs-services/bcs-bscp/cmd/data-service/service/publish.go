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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/structpb"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/components/itsm"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/enumor"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/cache-service"
	pbgroup "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/group"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/selector"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// Publish exec publish strategy.
// nolint: funlen
func (s *Service) Publish(ctx context.Context, req *pbds.PublishReq) (*pbds.PublishResp, error) {
	// 只给流水线插件做兼容，该接口暂时还不能去除
	grpcKit := kit.FromGrpcContext(ctx)

	app, err := s.dao.App().Get(grpcKit, req.BizId, req.AppId)
	if err != nil {
		return nil, err
	}
	// 要么不审批立即上线，要么审批后自动上线
	publishType := table.Immediately
	if app.Spec.IsApprove {
		publishType = table.Automatically
	}
	return s.SubmitPublishApprove(ctx, &pbds.SubmitPublishApproveReq{
		BizId:           req.BizId,
		AppId:           req.AppId,
		ReleaseId:       req.ReleaseId,
		Memo:            req.Memo,
		All:             req.All,
		GrayPublishMode: req.GrayPublishMode,
		Default:         req.Default,
		Groups:          req.Groups,
		Labels:          req.Labels,
		GroupName:       req.GroupName,
		PublishType:     string(publishType),
		PublishTime:     "",
		IsCompare:       false,
	})
}

// SubmitPublishApprove submit publish strategy.
// nolint funlen
func (s *Service) SubmitPublishApprove(
	ctx context.Context, req *pbds.SubmitPublishApproveReq) (*pbds.PublishResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	app, err := s.dao.App().Get(grpcKit, req.BizId, req.AppId)
	if err != nil {
		return nil, err
	}

	release, err := s.dao.Release().Get(grpcKit, req.BizId, req.AppId, req.ReleaseId)
	if err != nil {
		return nil, err
	}
	if release.Spec.Deprecated {
		return nil, fmt.Errorf("release %s is deprecated, can not be submited", release.Spec.Name)
	}

	// 获取最近的上线版本
	strategy, err := s.dao.Strategy().GetLast(grpcKit, req.BizId, req.AppId, 0, 0)
	if err != nil {
		return nil, err
	}

	// 有在上线的版本则提示不能上线
	if strategy.Spec.PublishStatus == table.PendApproval || strategy.Spec.PublishStatus == table.PendPublish {
		return nil, errors.New("there is a version in publishing currently")
	}

	isRollback := true
	tx := s.dao.GenQuery().Begin()
	defer func() {
		if isRollback {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
			}
		}
	}()

	// group name
	var groupIDs []uint32
	var groupName []string
	// group 解析处理, 通过label创建
	groupIDs, groupName, err = s.parseGroup(grpcKit, req, tx)
	if err != nil {
		logs.Errorf("parse group failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// parse publish option
	opt := s.parsePublishOption(req, app)
	opt.Groups = groupIDs
	opt.Revision = &table.CreatedRevision{
		Creator: grpcKit.User,
	}

	pshID, err := s.dao.Publish().SubmitWithTx(grpcKit, tx, opt)
	if err != nil {
		logs.Errorf("publish strategy failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	if req.All {
		groupName = []string{"all"}
	}

	resInstance := fmt.Sprintf("releases_name: %s\ngroup: %s", release.Spec.Name, strings.Join(groupName, ","))

	// audit this to create strategy details
	ad := s.dao.AuditDao().DecoratorV3(grpcKit, opt.BizID, &table.AuditField{
		OperateWay:       grpcKit.OperateWay,
		Action:           enumor.PublishVersionConfig,
		ResourceInstance: resInstance,
		Status:           enumor.AuditStatus(opt.PublishStatus),
		AppId:            app.AppID(),
		StrategyId:       pshID,
		IsCompare:        req.IsCompare,
	}).PrepareCreateByInstance(pshID, req)
	if err = ad.Do(tx.Query); err != nil {
		return nil, err
	}

	// 定时上线
	err = s.setPublishTime(grpcKit, pshID, req)
	if err != nil {
		return nil, err
	}

	// itsm流程创建ticket
	if app.Spec.IsApprove {
		scope := strings.Join(groupName, ",")
		ticketData, errCreate := s.submitCreateApproveTicket(
			grpcKit, app, release.Spec.Name, scope, ad.GetAuditID(), release.ID)
		if errCreate != nil {
			logs.Errorf("submit create approve ticket, err: %v, rid: %s", errCreate, grpcKit.Rid)
			return nil, errCreate
		}

		err = s.dao.Strategy().UpdateByID(grpcKit, tx, pshID, map[string]interface{}{
			"itsm_ticket_type":     constant.ItsmTicketTypeCreate,
			"itsm_ticket_url":      ticketData.TicketURL,
			"itsm_ticket_sn":       ticketData.SN,
			"itsm_ticket_status":   constant.ItsmTicketStatusCreated,
			"itsm_ticket_state_id": ticketData.StateID,
		})

		if err != nil {
			logs.Errorf("update strategy by id err: %v, rid: %s", err, grpcKit.Rid)
			return nil, err
		}
	}

	// 不是空值表示被客户端拉取过
	var havePull bool
	if app.Spec.LastConsumedTime != nil {
		havePull = true
	}

	haveCredentials, err := s.checkAppHaveCredentials(grpcKit, req.BizId, req.AppId)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	isRollback = false

	resp := &pbds.PublishResp{
		PublishedStrategyHistoryId: pshID,
		HaveCredentials:            haveCredentials,
		HavePull:                   havePull,
	}
	return resp, nil
}

// Approve publish approve.
// nolint funlen
func (s *Service) Approve(ctx context.Context, req *pbds.ApproveReq) (*pbds.ApproveResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	logs.Infof("start approve operateway: %s, user: %s, req: %v", grpcKit.OperateWay, grpcKit.User, req)

	release, err := s.dao.Release().Get(grpcKit, req.BizId, req.AppId, req.ReleaseId)
	if err != nil {
		return nil, err
	}
	if release.Spec.Deprecated {
		return nil, fmt.Errorf("release %s is deprecated, can not be revoke", release.Spec.Name)
	}

	strategy, err := s.dao.Strategy().GetLast(grpcKit, req.BizId, req.AppId, req.ReleaseId, req.StrategyId)
	if err != nil {
		return nil, err
	}

	// 从itsm回调的，如果状态跟数据库一样直接返回结果
	if grpcKit.OperateWay == "" && strategy.Spec.PublishStatus == table.PublishStatus(req.PublishStatus) {
		return &pbds.ApproveResp{}, nil
	}

	var message string
	// 获取itsm ticket状态，不审批的不查
	// message 不为空的情况：itsm操作后数据不正常的message皆不为空，但数据库需要更新
	if strategy.Spec.Approver != "" {
		req, message, err = checkTicketStatus(grpcKit,
			strategy.Spec.ItsmTicketSn, strategy.Spec.ItsmTicketStateID, req)
		if err != nil {
			return nil, err
		}
	}

	// 默认要回滚，除非已经提交
	isRollback := true
	tx := s.dao.GenQuery().Begin()
	defer func() {
		if isRollback {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
			}
		}
	}()

	var updateContent map[string]interface{}
	itsmUpdata := make(map[string]interface{})
	switch req.PublishStatus {
	case string(table.RevokedPublish):
		updateContent, err = s.revokeApprove(grpcKit, req, strategy)
		if err != nil {
			return nil, err
		}
		itsmUpdata = map[string]interface{}{
			"sn":             strategy.Spec.ItsmTicketSn,
			"operator":       grpcKit.User,
			"action_type":    "WITHDRAW",
			"action_message": fmt.Sprintf("BSCP 代理用户 %s 撤回: %s", grpcKit.User, req.Reason),
		}
	case string(table.RejectedApproval):
		updateContent, err = s.rejectApprove(grpcKit, req, strategy)
		if err != nil {
			return nil, err
		}
		itsmUpdata = map[string]interface{}{
			"sn":       strategy.Spec.ItsmTicketSn,
			"state_id": strategy.Spec.ItsmTicketStateID,
			"approver": grpcKit.User,
			"action":   "false",
			"remark":   req.Reason,
		}
	case string(table.PendPublish):
		updateContent, err = s.passApprove(grpcKit, tx, req, strategy)
		if err != nil {
			return nil, err
		}
		itsmUpdata = map[string]interface{}{
			"sn":       strategy.Spec.ItsmTicketSn,
			"state_id": strategy.Spec.ItsmTicketStateID,
			"approver": grpcKit.User,
			"action":   "true",
		}
	case string(table.AlreadyPublish):
		updateContent, err = s.publishApprove(grpcKit, tx, req, strategy)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid publish_status: %s", req.PublishStatus)
	}

	updateContent["reviser"] = grpcKit.User
	err = s.dao.Strategy().UpdateByID(grpcKit, tx, strategy.ID, updateContent)
	if err != nil {
		return nil, err
	}

	// update audit details
	err = s.dao.AuditDao().UpdateByStrategyID(grpcKit, tx, strategy.ID, map[string]interface{}{
		"status": updateContent["publish_status"],
	})
	if err != nil {
		return nil, err
	}

	// 从页面进来且需要审批的数据则同步itsm
	if strategy.Spec.Approver != "" && message == "" {
		// 撤销状态下，直接撤销
		if req.PublishStatus == string(table.RevokedPublish) {
			err = itsm.WithdrawTicket(grpcKit.Ctx, itsmUpdata)
			if err != nil {
				return nil, err
			}
		}

		if req.PublishStatus == string(table.RejectedApproval) || req.PublishStatus == string(table.PendPublish) {
			err = itsm.UpdateTicketByApporver(grpcKit.Ctx, itsmUpdata)
			if err != nil {
				return nil, err
			}
		}
	}

	haveCredentials, err := s.checkAppHaveCredentials(grpcKit, req.BizId, req.AppId)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	isRollback = false
	// 在itsm操作后数据不正常的皆以报错形式返回
	// 但如果是回调地址调用则正常返回
	if message != "" && grpcKit.OperateWay != "" {
		return nil, errors.New(message)
	}
	return &pbds.ApproveResp{
		HaveCredentials: haveCredentials,
	}, nil
}

// GenerateReleaseAndPublish generate release and publish.
// nolint: funlen
func (s *Service) GenerateReleaseAndPublish(ctx context.Context, req *pbds.GenerateReleaseAndPublishReq) (
	*pbds.PublishResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	app, err := s.dao.App().GetByID(grpcKit, req.AppId)
	if err != nil {
		logs.Errorf("get app failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	if _, e := s.dao.Release().GetByName(grpcKit, req.BizId, req.AppId, req.ReleaseName); e == nil {
		return nil, fmt.Errorf("release name %s already exists", req.ReleaseName)
	}

	// 获取最近的上线版本
	strategy, err := s.dao.Strategy().GetLast(grpcKit, req.BizId, req.AppId, 0, 0)
	if err != nil {
		return nil, err
	}

	// 有在上线的版本则提示不能上线
	if strategy.Spec.PublishStatus == table.PendApproval || strategy.Spec.PublishStatus == table.PendPublish {
		return nil, errors.New("there is a version in publishing currently")
	}

	// 默认要回滚，除非已经提交
	isRollback := true
	tx := s.dao.GenQuery().Begin()
	defer func() {
		if isRollback {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
			}
		}
	}()

	groupIDs, groupName, err := s.genReleaseAndPublishGroupID(grpcKit, tx, req)
	if err != nil {
		return nil, err
	}

	// create release.
	release := &table.Release{
		Spec: &table.ReleaseSpec{
			Name: req.ReleaseName,
			Memo: req.ReleaseMemo,
		},
		Attachment: &table.ReleaseAttachment{
			BizID: req.BizId,
			AppID: req.AppId,
		},
		Revision: &table.CreatedRevision{
			Creator: grpcKit.User,
		},
	}
	releaseID, err := s.dao.Release().CreateWithTx(grpcKit, tx, release)
	if err != nil {
		logs.Errorf("create release failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	// create released hook.
	if err = s.createReleasedHook(grpcKit, tx, req.BizId, req.AppId, releaseID); err != nil {
		logs.Errorf("create released hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	switch app.Spec.ConfigType {
	case table.File:

		// Note: need to change batch operator to query config item and it's commit.
		// query app's all config items.
		cfgItems, e := s.getAppConfigItems(grpcKit)
		if e != nil {
			logs.Errorf("query app config item list failed, err: %v, rid: %s", e, grpcKit.Rid)
			return nil, e
		}

		// get app template revisions which are template config items
		tmplRevisions, e := s.getAppTmplRevisions(grpcKit)
		if e != nil {
			logs.Errorf("get app template revisions failed, err: %v, rid: %s", e, grpcKit.Rid)
			return nil, e
		}

		// if no config item, return directly.
		if len(cfgItems) == 0 && len(tmplRevisions) == 0 {
			return nil, errors.New("app config items is empty")
		}

		// do template and non-template config item related operations for create release.
		if err = s.doConfigItemOperations(grpcKit, req.Variables, tx, release.ID, tmplRevisions, cfgItems); err != nil {
			logs.Errorf("do template action for create release failed, err: %v, rid: %s", err, grpcKit.Rid)
			return nil, err
		}
	case table.KV:
		if err = s.doKvOperations(grpcKit, tx, req.AppId, req.BizId, release.ID); err != nil {
			logs.Errorf("do kv action for create release failed, err: %v, rid: %s", err, grpcKit.Rid)
			return nil, err
		}
	}

	// publish with transaction.
	kt := kit.FromGrpcContext(ctx)

	opt := &types.PublishOption{
		BizID:     req.BizId,
		AppID:     req.AppId,
		ReleaseID: releaseID,
		All:       req.All,
		Memo:      req.ReleaseMemo,
		Groups:    groupIDs,
		Revision: &table.CreatedRevision{
			Creator: kt.User,
		},
		PublishType:   table.Immediately,
		PublishStatus: table.AlreadyPublish,
		PubState:      string(table.Publishing),
	}

	// if approval required, current approver required, pub_state unpublished
	if app.Spec.IsApprove {
		opt.PublishType = table.Automatically
		opt.PublishStatus = table.PendApproval
		opt.Approver = app.Spec.Approver
		opt.ApproverProgress = app.Spec.Approver
		opt.PubState = string(table.Unpublished)
	}

	// 后续app改审批方式的时候可以判断是或签还是会签
	if app.Spec.ApproveType == table.OrSign {
		opt.Approver = app.Spec.Approver
		approver := strings.Split(app.Spec.Approver, ",")
		opt.Approver = strings.Join(approver, "|")
	}

	pshID, err := s.dao.Publish().SubmitWithTx(grpcKit, tx, opt)
	if err != nil {
		logs.Errorf("submit with tx failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	if req.All {
		groupName = []string{"all"}
	}

	resInstance := fmt.Sprintf("releases_name: %s\ngroup: %s", release.Spec.Name, strings.Join(groupName, ","))

	// audit this to create strategy details
	ad := s.dao.AuditDao().DecoratorV3(grpcKit, opt.BizID, &table.AuditField{
		OperateWay:       grpcKit.OperateWay,
		Action:           enumor.PublishVersionConfig,
		ResourceInstance: resInstance,
		Status:           enumor.AuditStatus(opt.PublishStatus),
		AppId:            app.AppID(),
		StrategyId:       pshID,
	}).PrepareCreateByInstance(pshID, req)
	if err = ad.Do(tx.Query); err != nil {
		return nil, err
	}

	// itsm流程创建ticket
	if app.Spec.IsApprove {
		scope := strings.Join(groupName, ",")
		ticketData, errCreate := s.submitCreateApproveTicket(
			grpcKit, app, release.Spec.Name, scope, ad.GetAuditID(), release.ID)
		if errCreate != nil {
			logs.Errorf("submit create approve ticket, err: %v, rid: %s", errCreate, grpcKit.Rid)
			return nil, errCreate
		}

		err = s.dao.Strategy().UpdateByID(grpcKit, tx, pshID, map[string]interface{}{
			"itsm_ticket_type":     constant.ItsmTicketTypeCreate,
			"itsm_ticket_url":      ticketData.TicketURL,
			"itsm_ticket_sn":       ticketData.SN,
			"itsm_ticket_status":   constant.ItsmTicketStatusCreated,
			"itsm_ticket_state_id": ticketData.StateID,
		})

		if err != nil {
			logs.Errorf("update strategy by id err: %v, rid: %s", err, grpcKit.Rid)
			return nil, err
		}
	}
	// commit transaction.
	if err = tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	isRollback = false
	return &pbds.PublishResp{PublishedStrategyHistoryId: pshID}, nil
}

// revokeApprove revoke publish approve.
func (s *Service) revokeApprove(
	kit *kit.Kit, req *pbds.ApproveReq, strategy *table.Strategy) (map[string]interface{}, error) {

	// 提单的人才能撤销
	if strategy.Revision.Creator != kit.User && kit.OperateWay == string(enumor.WebUI) {
		return nil, errors.New("no permission to revoke")
	}

	// 只有待上线以及待审批的类型才允许撤回
	if strategy.Spec.PublishStatus != table.PendPublish && strategy.Spec.PublishStatus != table.PendApproval {
		return nil, fmt.Errorf("revoked not allowed, current publish status is: %s", strategy.Spec.PublishStatus)
	}

	return map[string]interface{}{
		"publish_status":     table.RevokedPublish,
		"reject_reason":      req.Reason,
		"approver_progress":  strategy.Revision.Creator,
		"itsm_ticket_status": constant.ItsmTicketStatusRevoked,
	}, nil
}

// rejectApprove reject publish approve.
func (s *Service) rejectApprove(
	kit *kit.Kit, req *pbds.ApproveReq, strategy *table.Strategy) (map[string]interface{}, error) {

	if strategy.Spec.PublishStatus != table.PendApproval {
		return nil, fmt.Errorf("rejected not allowed, current publish status is: %s", strategy.Spec.PublishStatus)
	}

	if req.Reason == "" {
		return nil, errors.New("reason can not empty")
	}

	var rejector string
	// 判断是否在审批人队列
	users := strings.Split(strategy.Spec.ApproverProgress, ",")
	for _, v := range users {
		if v == kit.User {
			rejector = v
			break
		}
		for _, vv := range req.ApprovedBy {
			if v == vv {
				rejector = vv
				break
			}
		}
	}

	// 需要审批但不是审批人的情况返回无权限审批
	if rejector == "" {
		return nil, errors.New("no permission to approve")
	}

	return map[string]interface{}{
		"publish_status":     table.RejectedApproval,
		"reject_reason":      req.Reason,
		"approver_progress":  rejector,
		"itsm_ticket_status": constant.ItsmTicketStatusRejected,
	}, nil
}

// passApprove pass publish approve.
func (s *Service) passApprove(
	kit *kit.Kit, tx *gen.QueryTx, req *pbds.ApproveReq, strategy *table.Strategy) (map[string]interface{}, error) {

	if strategy.Spec.PublishStatus != table.PendApproval {
		return nil, fmt.Errorf("pass not allowed, current publish status is: %s", strategy.Spec.PublishStatus)
	}

	// 存在app更改成不审批的情况，要根据审批人来确定是会签还是或签
	// 判断是否在审批人队列
	isApprover := false
	progressUsers := strings.Split(strategy.Spec.ApproverProgress, ",")
	// 新的审批人列表
	var newProgressUsers []string
	for _, v := range progressUsers {
		isRemove := false
		// 与itsm已经通过的审批人列表做对比
		for _, vv := range req.ApprovedBy {
			if vv == v {
				isRemove = true
				break
			}
		}
		if v == kit.User {
			isApprover = true
			isRemove = true
		}

		// 不需要移除的审批人列表
		if !isRemove {
			newProgressUsers = append(newProgressUsers, v)
		}
	}

	// 页面过来的数据不是审批人的情况返回无权限审批
	if !isApprover && kit.OperateWay == string(enumor.WebUI) {
		return nil, errors.New("no permission to approve")
	}

	result := make(map[string]interface{})
	publishStatus := table.PendApproval
	// 或签通过或者是只有一个审批人的情况
	if strings.Contains(strategy.Spec.Approver, "|") || strategy.Spec.Approver == kit.User {
		publishStatus = table.PendPublish
		result["approver_progress"] = kit.User // 需要更新下给前端展示
		result["itsm_ticket_status"] = constant.ItsmTicketStatusPassed
	} else {
		// 会签通过
		// 最后一个的情况下，直接待上线
		if len(newProgressUsers) == 0 || kit.OperateWay == "" {
			publishStatus = table.PendPublish
			result["approver_progress"] = strategy.Spec.Approver
			result["itsm_ticket_status"] = constant.ItsmTicketStatusPassed
		} else {
			// 审批人列表更新
			result["approver_progress"] = strings.Join(newProgressUsers, ",")
		}
	}

	// 自动上线则直接上线
	if publishStatus == table.PendPublish && strategy.Spec.PublishType == table.Automatically {
		opt := types.PublishOption{
			BizID:     req.BizId,
			AppID:     req.AppId,
			ReleaseID: req.ReleaseId,
			All:       false,
		}

		if len(strategy.Spec.Scope.Groups) == 0 {
			opt.All = true
		}

		err := s.dao.Publish().UpsertPublishWithTx(kit, tx, &opt, strategy)

		if err != nil {
			return nil, err
		}
		publishStatus = table.AlreadyPublish
	}

	result["publish_status"] = publishStatus
	return result, nil
}

// publishApprove publish approve.
func (s *Service) publishApprove(
	kit *kit.Kit, tx *gen.QueryTx, req *pbds.ApproveReq, strategy *table.Strategy) (map[string]interface{}, error) {

	if strategy.Spec.PublishStatus != table.PendPublish {
		return nil, fmt.Errorf("publish not allowed, current publish status is: %s", strategy.Spec.PublishStatus)
	}

	opt := types.PublishOption{
		BizID:     req.BizId,
		AppID:     req.AppId,
		ReleaseID: req.ReleaseId,
		All:       false,
	}

	if len(strategy.Spec.Scope.Groups) == 0 {
		opt.All = true
	}

	err := s.dao.Publish().UpsertPublishWithTx(kit, tx, &opt, strategy)

	if err != nil {
		return nil, err
	}
	publishStatus := table.AlreadyPublish

	return map[string]interface{}{
		"pub_state":      table.Publishing,
		"publish_status": publishStatus,
	}, nil
}

// parse publish option
func (s *Service) parsePublishOption(req *pbds.SubmitPublishApproveReq, app *table.App) *types.PublishOption {

	opt := &types.PublishOption{
		BizID:         req.BizId,
		AppID:         req.AppId,
		ReleaseID:     req.ReleaseId,
		All:           req.All,
		Default:       req.Default,
		Memo:          req.Memo,
		PublishType:   table.PublishType(req.PublishType),
		PublishTime:   req.PublishTime,
		PublishStatus: table.PendPublish,
		PubState:      string(table.Publishing),
	}

	// if approval required, current approver required, pub_state unpublished
	if app.Spec.IsApprove {
		opt.PublishStatus = table.PendApproval
		opt.Approver = app.Spec.Approver
		opt.ApproverProgress = app.Spec.Approver
		opt.PubState = string(table.Unpublished)
	}

	// 后续app改审批方式的时候可以判断是或签还是会签
	if app.Spec.ApproveType == table.OrSign {
		opt.Approver = app.Spec.Approver
		approver := strings.Split(app.Spec.Approver, ",")
		opt.Approver = strings.Join(approver, "|")
	}

	// publish immediately
	if req.PublishType == string(table.Immediately) {
		opt.PublishStatus = table.AlreadyPublish
	}

	return opt
}

// checkAppHaveCredentials check if there is available credential for app.
// 1. credential scope can match app name.
// 2. credential is enabled.
func (s *Service) checkAppHaveCredentials(grpcKit *kit.Kit, bizID, appID uint32) (bool, error) {
	app, err := s.dao.App().Get(grpcKit, bizID, appID)
	if err != nil {
		return false, err
	}
	matchedCredentials := make([]uint32, 0)
	scopes, err := s.dao.CredentialScope().ListAll(grpcKit, bizID)
	if err != nil {
		return false, err
	}
	if len(scopes) == 0 {
		return false, nil
	}
	for _, scope := range scopes {
		match, e := scope.Spec.CredentialScope.MatchApp(app.Spec.Name)
		if e != nil {
			return false, e
		}
		if match {
			matchedCredentials = append(matchedCredentials, scope.Attachment.CredentialId)
		}
	}
	credentials, e := s.dao.Credential().BatchListByIDs(grpcKit, bizID, matchedCredentials)
	if e != nil {
		return false, e
	}
	for _, credential := range credentials {
		if credential.Spec.Enable {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) genReleaseAndPublishGroupID(grpcKit *kit.Kit, tx *gen.QueryTx,
	req *pbds.GenerateReleaseAndPublishReq) ([]uint32, []string, error) {

	groupIDs := make([]uint32, 0)
	groupNames := make([]string, 0)

	if !req.All {
		if req.GrayPublishMode == "" {
			// !NOTE: Compatible with previous pipelined plugins version
			req.GrayPublishMode = table.PublishByGroups.String()
		}
		publishMode := table.GrayPublishMode(req.GrayPublishMode)
		if e := publishMode.Validate(); e != nil {
			return groupIDs, groupNames, e
		}
		// validate and query group ids.
		if publishMode == table.PublishByGroups {
			for _, name := range req.Groups {
				group, e := s.dao.Group().GetByName(grpcKit, req.BizId, name)
				if e != nil {
					return groupIDs, groupNames, fmt.Errorf("group %s not exist", name)
				}
				groupIDs = append(groupIDs, group.ID)
				groupNames = append(groupNames, group.Spec.Name)
			}
		}
		if publishMode == table.PublishByLabels {
			groupID, e := s.getOrCreateGroupByLabels(grpcKit, tx, req.BizId, req.AppId, req.GroupName, req.Labels)
			if e != nil {
				logs.Errorf("create group by labels failed, err: %v, rid: %s", e, grpcKit.Rid)
				return groupIDs, groupNames, e
			}
			groupIDs = append(groupIDs, groupID)
			groupNames = append(groupNames, req.GroupName)
		}
	}

	return groupIDs, groupNames, nil
}

func (s *Service) getOrCreateGroupByLabels(grpcKit *kit.Kit, tx *gen.QueryTx, bizID, appID uint32, groupName string,
	labels []*structpb.Struct) (uint32, error) {
	elements := make([]selector.Element, 0)
	for _, label := range labels {
		element, err := pbgroup.UnmarshalElement(label)
		if err != nil {
			return 0, fmt.Errorf("unmarshal group label failed, err: %v", err)
		}
		elements = append(elements, *element)
	}
	sel := &selector.Selector{
		LabelsAnd: elements,
	}
	groups, err := s.dao.Group().ListAppValidGroups(grpcKit, bizID, appID)
	if err != nil {
		return 0, err
	}
	exists := make([]*table.Group, 0)
	for _, group := range groups {
		if group.Spec.Selector.Equal(sel) {
			exists = append(exists, group)
		}
	}
	// if same labels group exists, return it's id.
	if len(exists) > 0 {
		return exists[0].ID, nil
	}
	// else create new one.
	if groupName != "" {
		// if group name is not empty, use it as group name.
		_, err = s.dao.Group().GetByName(grpcKit, bizID, groupName)
		// if group name already exists, return error.
		if err == nil {
			return 0, fmt.Errorf("group %s already exists", groupName)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, err
		}
	} else {
		// generate group name by time.
		groupName = time.Now().Format("20060102150405.000")
		groupName = fmt.Sprintf("g_%s", strings.ReplaceAll(groupName, ".", ""))
	}
	group := table.Group{
		Spec: &table.GroupSpec{
			Name:     groupName,
			Public:   false,
			Mode:     table.GroupModeCustom,
			Selector: sel,
		},
		Attachment: &table.GroupAttachment{
			BizID: bizID,
		},
		Revision: &table.Revision{
			Creator: grpcKit.User,
			Reviser: grpcKit.User,
		},
	}
	groupID, err := s.dao.Group().CreateWithTx(grpcKit, tx, &group)
	if err != nil {
		return 0, err
	}
	if err := s.dao.GroupAppBind().BatchCreateWithTx(grpcKit, tx, []*table.GroupAppBind{
		{
			GroupID: groupID,
			AppID:   appID,
			BizID:   bizID,
		},
	}); err != nil {
		return 0, err
	}
	return groupID, nil
}

func (s *Service) createReleasedHook(grpcKit *kit.Kit, tx *gen.QueryTx, bizID, appID, releaseID uint32) error {
	pre, err := s.dao.ReleasedHook().Get(grpcKit, bizID, appID, 0, table.PreHook)
	if err == nil {
		pre.ID = 0
		pre.ReleaseID = releaseID
		if _, e := s.dao.ReleasedHook().CreateWithTx(grpcKit, tx, pre); e != nil {
			logs.Errorf("create released pre-hook failed, err: %v, rid: %s", e, grpcKit.Rid)
			return e
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logs.Errorf("query released pre-hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		return err
	}
	post, err := s.dao.ReleasedHook().Get(grpcKit, bizID, appID, 0, table.PostHook)
	if err == nil {
		post.ID = 0
		post.ReleaseID = releaseID
		if _, e := s.dao.ReleasedHook().CreateWithTx(grpcKit, tx, post); e != nil {
			logs.Errorf("create released post-hook failed, err: %v, rid: %s", e, grpcKit.Rid)
			return e
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logs.Errorf("query released post-hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		return err
	}
	return nil
}

// submitCreateApproveTicket create new itsm create approve ticket
func (s *Service) submitCreateApproveTicket(
	kt *kit.Kit, app *table.App, releaseName, scope string, aduitId, releaseID uint32) (*itsm.CreateTicketData, error) {

	// 或签和会签是不同的模板
	getConfigKey := constant.CreateOrSignApproveItsmServiceID
	if app.Spec.ApproveType == table.CountSign {
		getConfigKey = constant.CreateCountSignApproveItsmServiceID
	}
	itsmConfig, err := s.dao.ItsmConfig().GetConfig(kt, getConfigKey)
	if err != nil {
		return nil, err
	}

	// 获取所有的业务信息
	bizList, err := s.esb.Cmdb().ListAllBusiness(kt.Ctx)
	if err != nil {
		return nil, err
	}

	if len(bizList.Info) == 0 {
		return nil, fmt.Errorf("biz list is empty")
	}

	var bizName string
	for _, biz := range bizList.Info {
		if biz.BizID == int64(app.BizID) {
			bizName = biz.BizName
			break
		}
	}

	fields := []map[string]interface{}{
		{
			"key":   "title",
			"value": "服务配置中心(BSCP)版本上线审批",
		}, {
			"key":   "BIZ",
			"value": fmt.Sprintf(bizName+"(%d)", app.BizID),
		}, {
			"key":   "APP",
			"value": app.Spec.Name,
		}, {
			"key":   "VERSION_NAME",
			"value": releaseName,
		}, {
			"key":   "SCOPE",
			"value": scope,
		}, {
			"key":   "COMPARE",
			"value": fmt.Sprintf("%s/space/2/records/all?id=%d", cc.DataService().Esb.BscpHost, aduitId),
		}, {
			"key":   "BIZ_ID",
			"value": app.BizID,
		}, {
			"key":   "APP_ID",
			"value": app.ID,
		}, {
			"key":   "RELEASE_ID",
			"value": releaseID,
		},
	}

	stateApproveId := strconv.Itoa(itsmConfig.StateApproveId)
	reqData := map[string]interface{}{
		"creator":    kt.User,
		"service_id": itsmConfig.Value,
		"fields":     fields,
		"meta": map[string]interface{}{
			"state_processors": map[string]interface{}{
				stateApproveId: app.Spec.Approver,
			}},
	}

	resp, err := itsm.CreateTicket(kt.Ctx, reqData)
	if err != nil {
		return nil, err
	}

	resp.StateID = itsmConfig.StateApproveId
	return resp, nil
}

// 定时上线
func (s *Service) setPublishTime(kt *kit.Kit, pshID uint32, req *pbds.SubmitPublishApproveReq) error {
	if req.PublishType == string(table.Periodically) {
		// 通过上海时区计算unix
		location, err := time.LoadLocation("Asia/Shanghai")
		if err != nil {
			logs.Errorf("load location failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		publishTime, err := time.ParseInLocation(time.DateTime, req.PublishTime, location)
		if err != nil {
			logs.Errorf("parse time failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		_, err = s.cs.SetPublishTime(kt.Ctx, &pbcs.SetPublishTimeReq{
			BizId:       req.BizId,
			StrategyId:  pshID,
			PublishTime: publishTime.Unix(),
			AppId:       req.AppId,
		})
		if err != nil {
			logs.Errorf("set publish time failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}
	return nil
}

// group 解析处理, 通过label创建
func (s *Service) parseGroup(
	grpcKit *kit.Kit, req *pbds.SubmitPublishApproveReq, tx *gen.QueryTx) ([]uint32, []string, error) {
	// group name
	groupIDs := make([]uint32, 0)
	groupName := []string{}
	if !req.All {
		if req.GrayPublishMode == "" {
			// !NOTE: Compatible with previous pipelined plugins version
			req.GrayPublishMode = table.PublishByGroups.String()
		}
		publishMode := table.GrayPublishMode(req.GrayPublishMode)
		if e := publishMode.Validate(); e != nil {
			return groupIDs, groupName, e
		}
		// validate and query group ids.
		if publishMode == table.PublishByGroups {
			for _, groupID := range req.Groups {
				if groupID == 0 {
					groupIDs = append(groupIDs, groupID)
					continue
				}
				group, e := s.dao.Group().Get(grpcKit, groupID, req.BizId)
				if e != nil {
					return groupIDs, groupName, fmt.Errorf("group %d not exist", groupID)
				}
				groupIDs = append(groupIDs, group.ID)
				groupName = append(groupName, group.Spec.Name)
			}
		}
		if publishMode == table.PublishByLabels {
			groupID, gErr := s.getOrCreateGroupByLabels(grpcKit, tx, req.BizId, req.AppId, req.GroupName, req.Labels)
			if gErr != nil {
				logs.Errorf("create group by labels failed, err: %v, rid: %s", gErr, grpcKit.Rid)
				return groupIDs, groupName, fmt.Errorf("get group by labels failed: %s", gErr)
			}
			groupIDs = append(groupIDs, groupID)
			groupName = append(groupName, req.GroupName)
		}
	}
	return groupIDs, groupName, nil
}

// 检查ticket status
func checkTicketStatus(grpcKit *kit.Kit,
	sn string, stateID int, req *pbds.ApproveReq) (*pbds.ApproveReq, string, error) {
	var message string
	// 先获取tikect status
	ticketStatus, err := itsm.GetTicketStatus(grpcKit.Ctx, sn)
	if err != nil {
		return req, message, err
	}

	switch ticketStatus.Data.CurrentStatus {
	case constant.TicketRunningStatu:
		// 如果从页面来的是撤回，直接返回，itsm撤销不会回调
		if req.PublishStatus == string(table.RevokedPublish) {
			return req, message, err
		}
		// 发生这种情况就是审批通过/驳回没有回调，手动更新
		if len(ticketStatus.Data.CurrentSteps) == 0 {
			GetApproveNodeResultData, err := itsm.GetApproveNodeResult(grpcKit.Ctx, sn, stateID)
			if err != nil {
				return req, message, err
			}

			// 审批人列表，驳回的时候只有一个，会签通过时会以逗号分隔
			req.ApprovedBy = strings.Split(GetApproveNodeResultData.Data.Processeduser, ",")
			// 审批通过，非待上线的情况
			if GetApproveNodeResultData.Data.ApproveResult && req.PublishStatus != string(table.AlreadyPublish) {
				req.PublishStatus = string(table.PendPublish)
				return req,
					fmt.Sprintf("approval has been passed by itsm person: %s",
						GetApproveNodeResultData.Data.Processeduser), nil
			}
			// itsm驳回
			if !GetApproveNodeResultData.Data.ApproveResult {
				req.PublishStatus = string(table.RejectedApproval)
				req.Reason = GetApproveNodeResultData.Data.ApproveRemark
				return req,
					fmt.Sprintf("approval has been rejected by itsm person: %s",
						GetApproveNodeResultData.Data.Processeduser), nil
			}
		} else {
			// 统计itsm有多少人已经审批通过
			passApprover, err := itsm.GetTicketLogsByPass(grpcKit.Ctx, sn)
			if err != nil {
				return req, message, err
			}
			req.ApprovedBy = passApprover
		}
		return req, message, nil

	case constant.TicketRevokedStatu:
		req.PublishStatus = string(table.RevokedPublish)
		return req, "approval has been revoked by itsm person", nil
	case constant.TicketFinishedStatu:
		// 单据已结束证明数据是正常的，直接报错返回
		return req, message, errors.New("approval has been finished by itsm person")
	default:
		// 其他状态一律当撤销
		req.PublishStatus = string(table.RevokedPublish)
		req.Reason = "invalid tikcet status: " + ticketStatus.Data.CurrentStatus
		return req,
			fmt.Sprintf("approval has been revoked, invalid tikcet status: %s", ticketStatus.Data.CurrentStatus), nil
	}
}
