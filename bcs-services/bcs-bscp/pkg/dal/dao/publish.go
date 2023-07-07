/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package dao

import (
	"errors"
	"time"

	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/selector"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/types"
)

// Publish defines all the publish operation related operations.
type Publish interface {
	// Publish publish an app's release with its strategy.
	// once an app's strategy along with its release id is published,
	// all its released config items are effected immediately.
	Publish(kit *kit.Kit, opt *types.PublishOption) (id uint32, err error)

	PublishWithTx(kit *kit.Kit, tx *gen.QueryTx, opt *types.PublishOption) (id uint32, err error)
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

	eDecorator := dao.event.Eventf(kit)
	var pubID uint32

	// 多个使用事务处理
	createTx := func(tx *gen.Query) error {
		groupIDs := opt.Groups
		if opt.All {
			m := dao.genQ.ReleasedGroup
			q := dao.genQ.ReleasedGroup.WithContext(kit.Ctx)
			err := q.Select(m.GroupID).Where(m.BizID.Eq(opt.BizID), m.AppID.Eq(opt.AppID),
				m.GroupID.Neq(0)).Scan(&groupIDs)
			if err != nil {
				logs.Errorf("get to be published groups(all) failed, err: %v, rid: %s", err, kit.Rid)
				return err
			}
			opt.Default = true
		}

		groups := make([]*table.Group, 0, len(groupIDs))
		var err error
		if len(groupIDs) > 0 {
			m := dao.genQ.Group
			q := dao.genQ.Group.WithContext(kit.Ctx)
			groups, err = q.Where(m.ID.In(groupIDs...), m.BizID.Eq(opt.BizID)).Find()
			if err != nil {
				logs.Errorf("get to be published groups(%s) failed, err: %v, rid: %s",
					tools.JoinUint32(groupIDs, ","), err, kit.Rid)
				return err
			}
		}

		// create strategy to publish it later
		stgID, err := dao.idGen.One(kit, table.StrategyTable)
		if err != nil {
			logs.Errorf("generate strategy id failed, err: %v, rid: %s", err, kit.Rid)
			return err
		}
		pubID = stgID
		stg := genStrategy(kit, opt, stgID, groups)

		sq := tx.Strategy.WithContext(kit.Ctx)
		if err := sq.Create(stg); err != nil {
			return err
		}

		// audit this to create strategy details
		ad := dao.auditDao.DecoratorV2(kit, opt.BizID).PrepareCreate(stg)
		if err := ad.Do(tx); err != nil {
			return err
		}
		// audit this to publish details
		ad = dao.auditDao.DecoratorV2(kit, opt.BizID).PreparePublish(stg)
		if err := ad.Do(tx); err != nil {
			return err
		}

		// add release publish num
		if err := dao.increaseReleasePublishNum(kit, tx, stg.Spec.ReleaseID); err != nil {
			logs.Errorf("increate release publish num failed, err: %v, rid: %s", err, kit.Rid)
			return err
		}

		if err := dao.upsertReleasedGroups(kit, tx, opt, stg); err != nil {
			logs.Errorf("upsert group current releases failed, err: %v, rid: %s", err, kit.Rid)
			return err
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
		if err := eDecorator.Fire(one); err != nil {
			logs.Errorf("fire publish strategy event failed, err: %v, rid: %s", err, kit.Rid)
			return errors.New("fire event failed, " + err.Error())
		}

		return nil
	}
	err := dao.genQ.Transaction(createTx)

	eDecorator.Finalizer(err)

	if err != nil {
		return 0, err
	}

	return pubID, nil
}

func genStrategy(kit *kit.Kit, opt *types.PublishOption, stgID uint32, groups []*table.Group) *table.Strategy {
	now := time.Now()
	return &table.Strategy{
		ID: stgID,
		Spec: &table.StrategySpec{
			Name:      now.Format(time.RFC3339),
			ReleaseID: opt.ReleaseID,
			AsDefault: opt.Default,
			Scope: &table.Scope{
				Groups: groups,
			},
			Mode: table.Normal,
			Memo: opt.Memo,
		},
		State: &table.StrategyState{
			PubState: table.Publishing,
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

	eDecorator := dao.event.Eventf(kit)
	var pubID uint32

	groupIDs := opt.Groups
	if opt.All {
		m := tx.ReleasedGroup
		q := tx.ReleasedGroup.WithContext(kit.Ctx)
		err := q.Select(m.GroupID).Where(m.BizID.Eq(opt.BizID), m.AppID.Eq(opt.AppID),
			m.GroupID.Neq(0)).Scan(&groupIDs)
		if err != nil {
			logs.Errorf("get to be published groups(all) failed, err: %v, rid: %s", err, kit.Rid)
			return 0, err
		}
		opt.Default = true
	}

	groups := make([]*table.Group, 0, len(groupIDs))
	var err error
	// list groups if gray release
	if len(groupIDs) > 0 {
		m := tx.Group
		q := tx.Group.WithContext(kit.Ctx)
		groups, err = q.Where(m.ID.In(groupIDs...)).Find()
		if err != nil {
			logs.Errorf("get to be published groups(%s) failed, err: %v, rid: %s",
				tools.JoinUint32(opt.Groups, ","), err, kit.Rid)
			return 0, err
		}
	}

	// create strategy to publish it later
	stgID, err := dao.idGen.One(kit, table.StrategyTable)
	if err != nil {
		logs.Errorf("generate strategy id failed, err: %v, rid: %s", err, kit.Rid)
		return 0, err
	}
	pubID = stgID
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
	if err := dao.increaseReleasePublishNum(kit, tx.Query, stg.Spec.ReleaseID); err != nil {
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

	return pubID, nil
}

// increaseReleasePublishNum increase release publish num by 1
func (dao *pubDao) increaseReleasePublishNum(kit *kit.Kit, tx *gen.Query, releaseID uint32) error {
	m := tx.Release
	q := tx.Release.WithContext(kit.Ctx)
	if _, err := q.Where(m.ID.Eq(releaseID)).UpdateSimple(m.PublishNum.Add(1)); err != nil {
		logs.Errorf("increase release publish num failed, err: %v, rid: %s", err, kit.Rid)
		return err
	}
	return nil
}

func (dao *pubDao) upsertReleasedGroups(kit *kit.Kit, tx *gen.Query, opt *types.PublishOption,
	stg *table.Strategy) error {
	groups := stg.Spec.Scope.Groups
	if opt.Default {
		groups = append(groups, &table.Group{
			ID: 0,
			Spec: &table.GroupSpec{
				Name:     "默认分组",
				Mode:     table.Default,
				Public:   true,
				Selector: new(selector.Selector),
				UID:      "",
			},
		})
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
