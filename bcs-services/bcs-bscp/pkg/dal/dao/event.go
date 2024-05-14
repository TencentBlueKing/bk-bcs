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

	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// Event defines all the event related operations.
type Event interface {
	// Eventf initialize an event decorator instance to fire the event and
	// update the event's state
	Eventf(kit *kit.Kit) *EDecorator

	// List events with options.
	List(kit *kit.Kit, startCursor uint32, opt *types.BasePage) ([]*table.Event, int64, error)

	// ListConsumedEvents list events with options that is handle event by cache service.
	ListConsumedEvents(kit *kit.Kit, startCursor uint32, opt *types.BasePage) ([]*table.Event, int64, error)

	// LatestCursor get the latest event cursor which is the last already
	// consumed event's id.
	// Note:
	// if the returned cursor(event id) is 0, this means no event has been
	// consumed.
	LatestCursor(kit *kit.Kit) (uint32, error)

	// RecordCursor is used to record the cursor which describe where
	// the event has already been consumed with event id, so that the
	// event can be consumed continuously without being skipped or lost.
	RecordCursor(kit *kit.Kit, eventID uint32) error

	// Purge is used to try to remove number of consumed events, which is
	// created before, from the oldest to now order by event id.
	Purge(kit *kit.Kit, daysAgo uint) error
}

var _ Event = new(eventDao)

type eventDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// Eventf initialize an event decorator instance to fire the event and
// update the event's state.
func (dao *eventDao) Eventf(kit *kit.Kit) *EDecorator {
	return &EDecorator{
		kit:   kit,
		idGen: dao.idGen,
		genQ:  dao.genQ,
	}
}

// EDecorator is the event decorator which is used to fire the event and then
// update the final state after the previous related resource operate db
// transaction is finished.
type EDecorator struct {
	kit *kit.Kit
	// the event id list.
	idList []uint32
	idGen  IDGenInterface
	genQ   *gen.Query
}

// Fire a resource's operate(Create/Update/Delete) event.
// After fire the event success, user should call the Finalizer function to do
// the event finalize work.
// Note:
// 1. Fire must be called *AFTER* all the logical operation has
// already been finished within the same transaction, which means
// the related resource's has already been created or updated or
// deleted.
// 2. Make sure that if the Fire execute failed, then the former
// resource's operations must be failed at the same time, which
// means the transaction will be aborted or rollback.
// 3. It's accepted that if a resource operation is failed but its
// event is recorded successfully, because when an event is consumed,
// its unique id will be used to check if the resource is exists or
// not:
//
//	(1) if not, this event will be ignored.
//	(2) if yes, this event will be consumed event even if it is not
//	a real event(because the according operation may have already
//	been aborted).
func (ef *EDecorator) Fire(es ...types.Event) error {
	if len(es) == 0 {
		return nil
	}
	for _, one := range es {
		if err := one.Validate(); err != nil {
			return err
		}
	}
	num := len(es)

	ids, err := ef.idGen.Batch(ef.kit, table.EventTable, num)
	if err != nil {
		return errf.New(errf.DBOpFailed, "generate event id failed, err: "+err.Error())
	}

	list := make([]*table.Event, num)
	for idx := range es {
		one := es[idx]
		list[idx] = &table.Event{
			ID:   ids[idx],
			Spec: one.Spec,
			State: &table.EventState{
				FinalStatus: table.UnknownFS,
			},
			Attachment: one.Attachment,
			Revision:   one.Revision,
		}
	}
	batchSize := 100

	q := ef.genQ.Event.WithContext(ef.kit.Ctx)
	if err := q.CreateInBatches(list, batchSize); err != nil {
		return fmt.Errorf("insert events failed, err: %v", err)
	}

	// remember the event id list for the following finalize operation use.
	ef.idList = ids

	return nil
}

// FireWithTx is used to fire the event with the given transaction.
// Note: FireWithTx would make event to success state directly
func (ef *EDecorator) FireWithTx(tx *gen.QueryTx, es ...types.Event) error {
	if len(es) == 0 {
		return nil
	}
	for _, one := range es {
		if err := one.Validate(); err != nil {
			return err
		}
	}
	num := len(es)

	ids, err := ef.idGen.Batch(ef.kit, table.EventTable, num)
	if err != nil {
		return errf.New(errf.DBOpFailed, "generate event id failed, err: "+err.Error())
	}

	list := make([]*table.Event, num)
	for idx := range es {
		one := es[idx]
		list[idx] = &table.Event{
			ID:   ids[idx],
			Spec: one.Spec,
			State: &table.EventState{
				FinalStatus: table.SuccessFS,
			},
			Attachment: one.Attachment,
			Revision:   one.Revision,
		}
	}
	batchSize := 100

	q := tx.Event.WithContext(ef.kit.Ctx)
	if err := q.CreateInBatches(list, batchSize); err != nil {
		return fmt.Errorf("insert events failed, err: %v", err)
	}

	// remember the event id list for the following finalize operation use.
	ef.idList = ids

	return nil
}

// Finalizer do the event finalize work, if the txnError is nil, then update the
// related events final state with success, otherwise update it with failed.
func (ef *EDecorator) Finalizer(txnError error) {

	if len(ef.idList) == 0 {
		return
	}

	state := table.SuccessFS
	if txnError != nil {
		state = table.FailedFS
	}

	m := ef.genQ.Event
	q := ef.genQ.Event.WithContext(ef.kit.Ctx)
	if _, err := q.Where(m.ID.In(ef.idList...)).
		Select(m.FinalStatus).Update(m.FinalStatus, state); err != nil {
		logs.ErrorDepthf(1, "update event final state to %d failed, id list: %s, err: %v, rid: %s", state, ef.idList,
			err, ef.kit.Rid)
	}
}

// List events with options.
func (dao *eventDao) List(kit *kit.Kit, startCursor uint32, opt *types.BasePage) ([]*table.Event, int64, error) {
	m := dao.genQ.Event
	q := dao.genQ.Event.WithContext(kit.Ctx)

	orderCol, ok := m.GetFieldByName(opt.Sort)
	if !ok {
		return nil, 0, fmt.Errorf("talbe events doesn't contains column %s", opt.Sort)
	}

	var orderCond field.Expr
	if opt.Order == types.Ascending {
		orderCond = orderCol
	} else {
		orderCond = orderCol.Desc()
	}

	result, count, err := q.Where(m.ID.Gt(startCursor), m.Resource.Neq(string(table.CursorReminder))).Order(orderCond).
		FindByPage(opt.Offset(), opt.LimitInt())
	if err != nil {
		return nil, 0, err
	}

	return result, count, nil
}

// ListConsumedEvents list events with options that is handle event by cache service.
func (dao *eventDao) ListConsumedEvents(kit *kit.Kit, startCursor uint32, opt *types.BasePage) (
	[]*table.Event, int64, error) {
	m := dao.genQ.Event
	q := dao.genQ.Event.WithContext(kit.Ctx)

	orderCol, ok := m.GetFieldByName(opt.Sort)
	if !ok {
		return nil, 0, fmt.Errorf("talbe events doesn't contains column %s", opt.Sort)
	}

	var orderCond field.Expr
	if opt.Order == types.Ascending {
		orderCond = orderCol
	} else {
		orderCond = orderCol.Desc()
	}

	result, count, err := q.Where(
		m.ID.Gt(startCursor),
		m.Resource.Neq(string(table.CursorReminder)),
		q.Columns(m.ID).Lte(q.Select(m.ResourceID).Where(m.Resource.Eq(string(table.CursorReminder))))).
		Order(orderCond).FindByPage(opt.Offset(), opt.LimitInt())
	if err != nil {
		return nil, 0, err
	}

	return result, count, nil
}

// LatestCursor get the latest event cursor which is the last already
// consumed event's id.
// Note:
// if the returned cursor(event id) is 0, this means no event has been
// consumed.
func (dao *eventDao) LatestCursor(kit *kit.Kit) (uint32, error) {
	m := dao.genQ.Event
	q := dao.genQ.Event.WithContext(kit.Ctx)

	var cursor uint32
	if err := q.Select(m.ResourceID).Where(m.Resource.Eq(string(table.CursorReminder))).Limit(1).
		Scan(&cursor); err != nil {
		return 0, err
	}

	return cursor, nil
}

const eventUser = "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp"

// RecordCursor is used to record the cursor which describe where
// the event has already been consumed with event id.
// Please reference its Interface for more detail description.
// eventID is the last event's id being consumed.
//
// Note:
// 1. the cursor is recorded with a special event resource which is
// table.CursorReminder.
// 2. this is an upsert operation.
func (dao *eventDao) RecordCursor(kit *kit.Kit, eventID uint32) error {
	if eventID <= 0 {
		return errors.New("invalid event id to record the cursor")
	}

	m := dao.genQ.Event
	q := dao.genQ.Event.WithContext(kit.Ctx)

	result, err := q.Where(m.ID.Eq(uint32(table.EventCursorReminderPrimaryID))).
		Select(m.ResourceID, m.CreatedAt).
		UpdateSimple(m.ResourceID.Value(eventID), m.CreatedAt.Value(time.Now().UTC()))
	if err != nil {
		return err
	}
	if result.RowsAffected == 1 {
		return nil
	}

	// the cursorReminder event is lost, insert it now, normally this can not happen, because
	// this event is initialed when the database is created.
	g := &table.Event{
		ID: uint32(table.EventCursorReminderPrimaryID),
		Spec: &table.EventSpec{
			Resource:    table.CursorReminder,
			ResourceID:  eventID,
			ResourceUid: "",
			OpType:      "",
		},
		Attachment: &table.EventAttachment{
			BizID: 0,
			AppID: 0,
		},
		Revision: &table.CreatedRevision{
			Creator: eventUser,
		},
	}
	if err := q.Create(g); err != nil {
		return err
	}

	return nil
}

// Purge is used to try to remove number of consumed events, which is
// created before, from the oldest to now order by event id.
func (dao *eventDao) Purge(kit *kit.Kit, daysAgo uint) error {
	m := dao.genQ.Event
	q := dao.genQ.Event.WithContext(kit.Ctx)

	var lastID uint32
	oldDate := time.Now().UTC().AddDate(0, 0, -int(daysAgo))
	if err := q.Select(m.ID).Where(m.CreatedAt.Lte(oldDate)).Order(m.ID.Desc()).Limit(1).Scan(&lastID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("get last event id to purge failed, err: %v", err)
	}
	logs.Infof("start to delete events less than or equal last id %d, rid: %s", lastID, kit.Rid)

	const step = 100
	for {
		result, err := q.Where(m.ID.Lte(lastID), m.ID.Neq(uint32(table.EventCursorReminderPrimaryID))).Order(m.ID).
			Limit(step).Delete()
		if err != nil {
			return fmt.Errorf("delete event failed, last id: %d, err: %v", lastID, err)
		}
		logs.Infof("deleted %d events successfully, rid: %s", result.RowsAffected, kit.Rid)

		if result.RowsAffected < step {
			return nil
		}

		// sleep a while before next delete step
		time.Sleep(10 * time.Second)
	}
}
