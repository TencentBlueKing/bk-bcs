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

package dao

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/selector"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// Publish defines all the publish operation related operations.
type Publish interface {
	// Publish publish an app's release with its strategy.
	// once an app's strategy along with its release id is published,
	// all its released config items are effected immediately.
	Publish(kit *kit.Kit, opt *types.PublishOption) (id uint32, err error)

	PublishWithTx(kit *kit.Kit, tx *gen.QueryTx, opt *types.PublishOption) (id uint32, err error)

	SubmitWithTx(kit *kit.Kit, tx *gen.QueryTx, opt *types.PublishOption) (id uint32, err error)

	UpsertPublishWithTx(kit *kit.Kit, tx *gen.QueryTx, opt *types.PublishOption, stg *table.Strategy) error
}

var _ Publish = new(pubDao)

type pubDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
	event    Event
}

// Publish publish an app's release with its strategy.
// once an app's strategy along with its release id is published,
// all its released config items are effected immediately.
// return the published strategy history record id.
func (dao *pubDao) Publish(kit *kit.Kit, opt *types.PublishOption) (uint32, error) {
	if opt == nil {
		return 0, errors.New("publish strategy option is nil")
	}

	if err := opt.Validate(); err != nil {
		return 0, err
	}

	// 手动事务处理
	tx := dao.genQ.Begin()
	if err := dao.validatePublishGroups(kit, tx, opt); err != nil {
		if e := tx.Rollback(); e != nil {
			logs.Errorf("rollback publish transaction failed, err: %v, rid: %s", e, kit.Rid)
		}
		return 0, err
	}
	stgID, err := func() (uint32, error) {
		groups := make([]*table.Group, 0, len(opt.Groups))
		var err error
		if len(opt.Groups) > 0 {
			m := tx.Group
			q := tx.Group.WithContext(kit.Ctx)
			groups, err = q.Where(m.ID.In(opt.Groups...), m.BizID.Eq(opt.BizID)).Find()
			if err != nil {
				logs.Errorf("get to be published groups(%s) failed, err: %v, rid: %s",
					tools.JoinUint32(opt.Groups, ","), err, kit.Rid)
				return 0, err
			}
		}
		return dao.publish(kit, tx, opt, groups)
	}()
	if err != nil {
		if e := tx.Rollback(); e != nil {
			logs.Errorf("rollback publish transaction failed, err: %v, rid: %s", e, kit.Rid)
		}
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		logs.Errorf("commit publish transaction failed, err: %v, rid: %s", err, kit.Rid)
		return 0, err
	}

	return stgID, nil
}

func (dao *pubDao) validatePublishGroups(kt *kit.Kit, tx *gen.QueryTx, opt *types.PublishOption) error {
	for _, groupID := range opt.Groups {
		// frontend would set groupID 0 as default.
		if groupID == 0 {
			opt.Default = true
			continue
		}
		gm := tx.Group
		group, e := gm.WithContext(kt.Ctx).Where(gm.ID.Eq(groupID), gm.BizID.Eq(opt.BizID)).Take()
		if e != nil {
			if e == gorm.ErrRecordNotFound {
				return fmt.Errorf("group %d not exists", groupID)
			}
			return e
		}
		if group.Spec.Public {
			continue
		}

		gam := tx.GroupAppBind
		if _, e := gam.WithContext(kt.Ctx).Where(
			gam.GroupID.Eq(groupID), gam.AppID.Eq(opt.AppID), gam.BizID.Eq(opt.BizID)).Take(); e != nil {
			if e == gorm.ErrRecordNotFound {
				return fmt.Errorf("group %d not bind app %d", groupID, opt.AppID)
			}
			return e
		}
	}
	return nil
}

func genStrategy(kit *kit.Kit, opt *types.PublishOption, stgID uint32, groups []*table.Group) *table.Strategy {
	now := time.Now()
	state := table.Publishing
	if opt.PubState != "" {
		state = table.PublishState(opt.PubState)
	}

	return &table.Strategy{
		ID: stgID,
		Spec: &table.StrategySpec{
			Name:      now.Format(time.RFC3339),
			ReleaseID: opt.ReleaseID,
			AsDefault: opt.Default,
			Scope: &table.Scope{
				Groups: groups,
			},
			Memo:             opt.Memo,
			PublishType:      opt.PublishType,
			PublishTime:      opt.PublishTime,
			PublishStatus:    opt.PublishStatus,
			RejectReason:     opt.RejectReason,
			Approver:         opt.Approver,
			ApproverProgress: opt.ApproverProgress,
		},
		State: &table.StrategyState{
			PubState: state,
		},
		Attachment: &table.StrategyAttachment{
			BizID: opt.BizID,
			AppID: opt.AppID,
		},
		Revision: &table.Revision{
			Creator: kit.User,
			Reviser: kit.User,
		},
	}
}

// PublishWithTx publish with transaction
func (dao *pubDao) PublishWithTx(kit *kit.Kit, tx *gen.QueryTx, opt *types.PublishOption) (uint32, error) {
	if opt == nil {
		return 0, errors.New("publish strategy option is nil")
	}

	if err := opt.Validate(); err != nil {
		return 0, err
	}

	if err := dao.validatePublishGroups(kit, tx, opt); err != nil {
		if e := tx.Rollback(); e != nil {
			logs.Errorf("rollback publish transaction failed, err: %v, rid: %s", e, kit.Rid)
		}
		return 0, err
	}

	groups := make([]*table.Group, 0, len(opt.Groups))
	var err error
	if len(opt.Groups) > 0 {
		m := tx.Group
		q := tx.Group.WithContext(kit.Ctx)
		groups, err = q.Where(m.ID.In(opt.Groups...), m.BizID.Eq(opt.BizID)).Find()
		if err != nil {
			logs.Errorf("get to be published groups(%s) failed, err: %v, rid: %s",
				tools.JoinUint32(opt.Groups, ","), err, kit.Rid)
			return 0, err
		}
	}
	return dao.publish(kit, tx, opt, groups)
}

func (dao *pubDao) publish(kit *kit.Kit, tx *gen.QueryTx, opt *types.PublishOption, groups []*table.Group) (
	uint32, error) {
	eDecorator := dao.event.Eventf(kit)
	// create strategy to publish it later
	stgID, err := dao.idGen.One(kit, table.StrategyTable)
	if err != nil {
		logs.Errorf("generate strategy id failed, err: %v, rid: %s", err, kit.Rid)
		return 0, err
	}
	stg := genStrategy(kit, opt, stgID, groups)

	sq := tx.Strategy.WithContext(kit.Ctx)
	if err := sq.Create(stg); err != nil {
		return 0, err
	}

	// audit this to create strategy details
	ad := dao.auditDao.DecoratorV2(kit, opt.BizID).PrepareCreate(stg)
	if err := ad.Do(tx.Query); err != nil {
		return 0, err
	}
	// audit this to publish details
	ad = dao.auditDao.DecoratorV2(kit, opt.BizID).PreparePublish(stg)
	if err := ad.Do(tx.Query); err != nil {
		return 0, err
	}

	// add release publish num
	if err := dao.updateReleasePublishInfo(kit, tx.Query, opt); err != nil {
		logs.Errorf("increate release publish num failed, err: %v, rid: %s", err, kit.Rid)
		return 0, err
	}

	if err := dao.upsertReleasedGroups(kit, tx.Query, opt, stg); err != nil {
		logs.Errorf("upsert group current releases failed, err: %v, rid: %s", err, kit.Rid)
		return 0, err
	}

	// fire the event with txn to ensure the if save the event failed then the business logic is failed anyway.
	one := types.Event{
		Spec: &table.EventSpec{
			Resource: table.Publish,
			// use the published strategy history id, which represent a real publish operation.
			ResourceID: opt.ReleaseID,
			OpType:     table.InsertOp,
		},
		Attachment: &table.EventAttachment{BizID: opt.BizID, AppID: opt.AppID},
		Revision:   &table.CreatedRevision{Creator: kit.User},
	}
	if err := eDecorator.FireWithTx(tx, one); err != nil {
		logs.Errorf("fire publish strategy event failed, err: %v, rid: %s", err, kit.Rid)
		return 0, errors.New("fire event failed, " + err.Error())
	}

	return stgID, nil
}

// updateReleasePublishInfo update release publish info, include publish num and fully released status.
func (dao *pubDao) updateReleasePublishInfo(kit *kit.Kit, tx *gen.Query, opt *types.PublishOption) error {
	m := tx.Release
	q := tx.Release.WithContext(kit.Ctx)
	// if publish all or publish default group, then set fully released to true.
	if opt.All || opt.Default {
		if _, err := q.Where(m.ID.Eq(opt.ReleaseID)).UpdateSimple(m.PublishNum.Add(1),
			m.FullyReleased.Value(true)); err != nil {
			logs.Errorf("update release publish info failed, err: %v, rid: %s", err, kit.Rid)
			return err
		}
	} else {
		if _, err := q.Where(m.ID.Eq(opt.ReleaseID)).UpdateSimple(m.PublishNum.Add(1)); err != nil {
			logs.Errorf("update release publish info failed, err: %v, rid: %s", err, kit.Rid)
			return err
		}
	}
	return nil
}

// nolint: funlen
func (dao *pubDao) upsertReleasedGroups(kit *kit.Kit, tx *gen.Query, opt *types.PublishOption,
	stg *table.Strategy) error {
	defaultGroup := &table.Group{
		ID: 0,
		Spec: &table.GroupSpec{
			Name:     "默认分组",
			Mode:     table.GroupModeDefault,
			Public:   true,
			Selector: new(selector.Selector),
			UID:      "",
		},
	}
	if opt.All {
		// 1. delete all released groups
		m := tx.ReleasedGroup
		if _, err := m.WithContext(kit.Ctx).Where(m.BizID.Eq(opt.BizID), m.AppID.Eq(opt.AppID)).Delete(); err != nil {
			logs.Errorf("delete all released groups failed, err: %v, rid: %s", err, kit.Rid)
			return err
		}
		// 2. insert default group
		rgID, err := dao.idGen.One(kit, table.ReleasedGroupTable)
		if err != nil {
			logs.Errorf("generate released group id failed, err: %v, rid: %s", err, kit.Rid)
			return err
		}
		rg := &table.ReleasedGroup{
			ID:         rgID,
			GroupID:    defaultGroup.ID,
			AppID:      opt.AppID,
			ReleaseID:  opt.ReleaseID,
			StrategyID: stg.ID,
			Mode:       defaultGroup.Spec.Mode,
			Selector:   defaultGroup.Spec.Selector,
			UID:        defaultGroup.Spec.UID,
			Edited:     false,
			BizID:      opt.BizID,
			Reviser:    kit.User,
		}
		if err := tx.ReleasedGroup.WithContext(kit.Ctx).Create(rg); err != nil {
			logs.Errorf("insert default released group failed, err: %v, rid: %s", err, kit.Rid)
			return err
		}
		return nil
	}

	groups := stg.Spec.Scope.Groups
	if opt.Default {
		// 1. delete other released groups in this release.
		m := tx.ReleasedGroup
		if _, err := m.WithContext(kit.Ctx).
			Where(m.BizID.Eq(opt.BizID), m.AppID.Eq(opt.AppID), m.ReleaseID.Eq(opt.ReleaseID)).
			Delete(); err != nil {
			logs.Errorf("delete other released groups in release failed, err: %v, rid: %s", err, kit.Rid)
			return err
		}
		// 2. delete others groups in other releases.
		groupIDs := make([]uint32, 0, len(groups))
		for _, group := range groups {
			groupIDs = append(groupIDs, group.ID)
		}
		if _, err := m.WithContext(kit.Ctx).
			Where(m.BizID.Eq(opt.BizID), m.AppID.Eq(opt.AppID), m.GroupID.In(groupIDs...)).
			Delete(); err != nil {
			logs.Errorf("delete other released groups in other releases failed, err: %v, rid: %s", err, kit.Rid)
			return err
		}
		// 3. only publish default group
		groups = []*table.Group{defaultGroup}
	}
	for _, group := range groups {
		rg := &table.ReleasedGroup{
			GroupID:    group.ID,
			AppID:      opt.AppID,
			ReleaseID:  opt.ReleaseID,
			StrategyID: stg.ID,
			Mode:       group.Spec.Mode,
			Selector:   group.Spec.Selector,
			UID:        group.Spec.UID,
			Edited:     false,
			BizID:      opt.BizID,
			Reviser:    kit.User,
		}

		m := tx.ReleasedGroup
		q := tx.ReleasedGroup.WithContext(kit.Ctx)

		result, err := q.Where(m.BizID.Eq(opt.BizID), m.AppID.Eq(opt.AppID), m.GroupID.Eq(group.ID)).
			Omit(m.ID).Updates(rg)
		if err != nil {
			return err
		}
		if result.RowsAffected == 1 {
			continue
		}

		id, err := dao.idGen.One(kit, table.ReleasedGroupTable)
		if err != nil {
			return err
		}
		rg.ID = id

		if err := q.Create(rg); err != nil {
			return err
		}
	}
	return nil
}

// SubmitWithTx submit with transaction
func (dao *pubDao) SubmitWithTx(kit *kit.Kit, tx *gen.QueryTx, opt *types.PublishOption) (uint32, error) {
	if opt == nil {
		return 0, errors.New("submit strategy option is nil")
	}

	if err := opt.Validate(); err != nil {
		return 0, err
	}

	if err := dao.validatePublishGroups(kit, tx, opt); err != nil {
		if e := tx.Rollback(); e != nil {
			logs.Errorf("rollback publish transaction failed, err: %v, rid: %s", e, kit.Rid)
		}
		return 0, err
	}

	groups := make([]*table.Group, 0, len(opt.Groups))
	var err error
	if len(opt.Groups) > 0 {
		m := tx.Group
		q := tx.Group.WithContext(kit.Ctx)
		groups, err = q.Where(m.ID.In(opt.Groups...), m.BizID.Eq(opt.BizID)).Find()
		if err != nil {
			logs.Errorf("get to be published groups(%s) failed, err: %v, rid: %s",
				tools.JoinUint32(opt.Groups, ","), err, kit.Rid)
			return 0, err
		}
	}

	return dao.submit(kit, tx, opt, groups)
}

// create relate table
func (dao *pubDao) submit(kit *kit.Kit, tx *gen.QueryTx, opt *types.PublishOption, groups []*table.Group) (
	uint32, error) {
	eDecorator := dao.event.Eventf(kit)
	// create strategy to publish it later
	stgID, err := dao.idGen.One(kit, table.StrategyTable)
	if err != nil {
		logs.Errorf("generate strategy id failed, err: %v, rid: %s", err, kit.Rid)
		return 0, err
	}
	if opt.PublishType == table.Immediately {
		// publish immediately auto generation time
		opt.PublishTime = time.Now().Format(time.DateTime)
	}
	stg := genStrategy(kit, opt, stgID, groups)

	err = stg.Spec.ValidateSubmitPublishContent()
	if err != nil {
		logs.Errorf("validate publish content failed, err: %v", err)
		return 0, err
	}

	sq := tx.Strategy.WithContext(kit.Ctx)
	if err := sq.Create(stg); err != nil {
		return 0, err
	}

	// publish immediately and update the table directly
	if opt.PublishType == table.Immediately {

		// add release publish num
		if err := dao.updateReleasePublishInfo(kit, tx.Query, opt); err != nil {
			logs.Errorf("increate release publish num failed, err: %v, rid: %s", err, kit.Rid)
			return 0, err
		}

		if err := dao.upsertReleasedGroups(kit, tx.Query, opt, stg); err != nil {
			logs.Errorf("upsert group current releases failed, err: %v, rid: %s", err, kit.Rid)
			return 0, err
		}
	}

	// fire the event with txn to ensure the if save the event failed then the business logic is failed anyway.
	one := types.Event{
		Spec: &table.EventSpec{
			Resource: table.Publish,
			// use the published strategy history id, which represent a real publish operation.
			ResourceID: opt.ReleaseID,
			OpType:     table.InsertOp,
		},
		Attachment: &table.EventAttachment{BizID: opt.BizID, AppID: opt.AppID},
		Revision:   &table.CreatedRevision{Creator: kit.User},
	}
	if err := eDecorator.FireWithTx(tx, one); err != nil {
		logs.Errorf("fire publish strategy event failed, err: %v, rid: %s", err, kit.Rid)
		return 0, errors.New("fire event failed, " + err.Error())
	}

	return stgID, nil
}

// UpsertPublishWithTx publish with transaction
func (dao *pubDao) UpsertPublishWithTx(
	kit *kit.Kit, tx *gen.QueryTx, opt *types.PublishOption, stg *table.Strategy) error {
	// add release publish num
	if err := dao.updateReleasePublishInfo(kit, tx.Query, opt); err != nil {
		logs.Errorf("increate release publish num failed, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	if err := dao.upsertReleasedGroups(kit, tx.Query, opt, stg); err != nil {
		logs.Errorf("upsert group current releases failed, err: %v, rid: %s", err, kit.Rid)
		return err
	}
	return nil
}
