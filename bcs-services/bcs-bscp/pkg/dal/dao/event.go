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
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/orm"
	"bscp.io/pkg/dal/sharding"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/types"
)

// Event defines all the event related operations.
type Event interface {
	// Eventf initialize an event decorator instance to fire the event and
	// update the event's state
	Eventf(kt *kit.Kit) *EDecorator

	// List events with options.
	List(kt *kit.Kit, opts *types.ListEventsOption) (*types.ListEventDetails, error)

	// ListConsumedEvents list events with options that is handle event by cache service.
	ListConsumedEvents(kt *kit.Kit, opts *types.ListEventsOption) (*types.ListEventDetails, error)

	// LatestCursor get the latest event cursor which is the last already
	// consumed event's id.
	// Note:
	// if the returned cursor(event id) is 0, this means no event has been
	// consumed.
	LatestCursor(kt *kit.Kit) (uint32, error)

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
	orm   orm.Interface
	sd    *sharding.Sharding
	idGen IDGenInterface
}

// Eventf initialize an event decorator instance to fire the event and
// update the event's state.
func (ed *eventDao) Eventf(kt *kit.Kit) *EDecorator {
	return &EDecorator{
		kt:    kt,
		orm:   ed.orm,
		sd:    ed.sd,
		idGen: ed.idGen,
	}
}

// EDecorator is the event decorator which is used to fire the event and then
// update the final state after the previous related resource operate db
// transaction is finished.
type EDecorator struct {
	kt *kit.Kit
	// the event id list.
	idList []uint32
	orm    orm.Interface
	sd     *sharding.Sharding
	idGen  IDGenInterface
}

// Fire a resource's operate(Create/Update/Delete) event.
// After fire the event success, user should call the Finalizer function to do
// the event finalize work.
// Note:
// 1. Eventf must be called *AFTER* all the logical operation has
// already been finished within the same transaction, which means
// the related resource's has already been created or updated or
// deleted.
// 2. Make sure that if the Eventf execute failed, then the former
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

	ids, err := ef.idGen.Batch(ef.kt, table.EventTable, num)
	if err != nil {
		return errf.New(errf.DBOpFailed, "generate event id failed, err: "+err.Error())
	}

	list := make([]table.Event, num)
	for idx := range es {
		one := es[idx]
		list[idx] = table.Event{
			ID:   ids[idx],
			Spec: one.Spec,
			State: &table.EventState{
				FinalStatus: table.UnknownFS,
			},
			Attachment: one.Attachment,
			Revision:   one.Revision,
		}
	}

	one := ef.sd.Event()
	if err := one.Err(); err != nil {
		return errf.New(errf.Aborted, "insert events, but get event db failed, err: "+err.Error())
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "INSERT INTO ", table.EventTable.Name(), " (", table.EventColumns.ColumnExpr(), ") VALUES(", table.EventColumns.ColonNameExpr(), ")")
	sql := filter.SqlJoint(sqlSentence)
	if err := ef.orm.Do(one.DB()).BulkInsert(ef.kt.Ctx, sql, list); err != nil {
		return errf.New(errf.InvalidParameter, "insert events failed, err: "+err.Error())
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

	in := make([]string, len(ef.idList))
	for idx := range ef.idList {
		in[idx] = strconv.FormatUint(uint64(ef.idList[idx]), 10)
	}
	joined := strings.Join(in, ",")

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "UPDATE ", table.EventTable.Name(), " SET final_status = ", strconv.Itoa(int(state)), " WHERE id IN(", joined, ")")
	sql := filter.SqlJoint(sqlSentence)
	_, err := ef.orm.Do(ef.sd.Event().DB()).Exec(context.TODO(), sql)
	if err != nil {
		logs.ErrorDepthf(1, "update event final state to %d failed, id list: %s, err: %v, rid: %s", state, joined,
			err, ef.kt.Rid)
		return
	}

	return
}

// List events with options.
func (ed *eventDao) List(kt *kit.Kit, opts *types.ListEventsOption) (*types.ListEventDetails, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list events options is nil")
	}

	if err := opts.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	sqlOpt := &filter.SQLWhereOption{
		Priority: filter.Priority{"id", "resource", "biz_id"},
		CrownedOption: &filter.CrownedOption{
			CrownedOp: filter.And,
			Rules: []filter.RuleFactory{
				// can not query
				&filter.AtomRule{
					Field: "resource",
					Op:    filter.NotEqual.Factory(),
					Value: table.CursorReminder,
				}},
		},
	}

	whereExpr, args, err := opts.Filter.SQLWhereExpr(sqlOpt)
	if err != nil {
		return nil, err
	}

	var sqlSentence []string
	if opts.Page.Count {
		// this is a count request, then do count operation only.
		sqlSentence = append(sqlSentence, "SELECT COUNT(*) FROM ", table.EventTable.Name(), whereExpr)
		sql := filter.SqlJoint(sqlSentence)
		var count uint32
		count, err = ed.orm.Do(ed.sd.Event().DB()).Count(kt.Ctx, sql, args...)
		if err != nil {
			return nil, err
		}

		return &types.ListEventDetails{Count: count, Details: make([]*table.Event, 0)}, nil
	}

	pageOption := &types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}}
	pageExpr, err := opts.Page.SQLExpr(pageOption)
	if err != nil {
		return nil, err
	}

	sqlSentence = append(sqlSentence, "SELECT ", table.EventColumns.NamedExpr(), " FROM ", table.EventTable.Name(), whereExpr, pageExpr)
	sql := filter.SqlJoint(sqlSentence)

	list := make([]*table.Event, 0)
	err = ed.orm.Do(ed.sd.Event().DB()).Select(kt.Ctx, &list, sql, args...)
	if err != nil {
		return nil, err
	}

	return &types.ListEventDetails{Count: 0, Details: list}, nil
}

// ListConsumedEvents list events with options that is handle event by cache service.
func (ed *eventDao) ListConsumedEvents(kt *kit.Kit, opts *types.ListEventsOption) (*types.ListEventDetails, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list consumed events options is nil")
	}

	if opts.Page.Count {
		return nil, errf.New(errf.InvalidParameter, "list consumed events not support count")
	}

	if err := opts.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	sqlOpt := &filter.SQLWhereOption{
		Priority: filter.Priority{"id", "resource", "biz_id"},
		CrownedOption: &filter.CrownedOption{
			CrownedOp: filter.And,
			Rules: []filter.RuleFactory{
				// can not query
				&filter.AtomRule{
					Field: "resource",
					Op:    filter.NotEqual.Factory(),
					Value: table.CursorReminder,
				}},
		},
	}

	whereExpr, args, err := opts.Filter.SQLWhereExpr(sqlOpt)
	if err != nil {
		return nil, err
	}

	pageOption := &types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}}
	pageExpr, err := opts.Page.SQLExpr(pageOption)
	if err != nil {
		return nil, err
	}

	var sqlSentenceSub []string
	sqlSentenceSub = append(sqlSentenceSub, "SELECT resource_id FROM ", table.EventTable.Name(), " WHERE resource = '", string(table.CursorReminder), "'")
	subSql := filter.SqlJoint(sqlSentenceSub)
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.EventColumns.NamedExpr(), " FROM ", table.EventTable.Name(), whereExpr,
		" AND id <= (", subSql, ") ", pageExpr)
	sql := filter.SqlJoint(sqlSentence)

	list := make([]*table.Event, 0)
	err = ed.orm.Do(ed.sd.Event().DB()).Select(kt.Ctx, &list, sql, args...)
	if err != nil {
		return nil, err
	}

	return &types.ListEventDetails{Count: 0, Details: list}, nil
}

// LatestCursor get the latest event cursor which is the last already
// consumed event's id.
// Note:
// if the returned cursor(event id) is 0, this means no event has been
// consumed.
func (ed *eventDao) LatestCursor(kt *kit.Kit) (uint32, error) {

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT resource_id FROM ", table.EventTable.Name(), " WHERE resource = '", string(table.CursorReminder), "' limit 1")
	sql := filter.SqlJoint(sqlSentence)

	cursor := uint32(0)
	if err := ed.orm.Do(ed.sd.Event().DB()).Get(kt.Ctx, &cursor, sql); err != nil {
		if err == orm.ErrRecordNotFound {
			return 0, nil
		}

		return 0, errf.New(errf.DBOpFailed, err.Error())
	}

	return cursor, nil
}

const eventUser = "bscp.io"

// RecordCursor is used to record the cursor which describe where
// the event has already been consumed with event id.
// Please reference its Interface for more detail description.
// eventID is the last event's id being consumed.
//
// Note:
// 1. the cursor is recorded with a special event resource which is
// table.CursorReminder.
// 2. this is an upsert operation.
func (ed *eventDao) RecordCursor(kt *kit.Kit, eventID uint32) error {

	if eventID <= 0 {
		return errors.New("invalid event id to record the cursor")
	}

	at := time.Now().Format(constant.TimeStdFormat)
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "UPDATE ", table.EventTable.Name(), " SET resource_id = ", strconv.Itoa(int(eventID)),
		", created_at = '", at, "' WHERE id = ", strconv.Itoa(table.EventCursorReminderPrimaryID))
	sql := filter.SqlJoint(sqlSentence)

	one := ed.sd.Event()
	if err := one.Err(); err != nil {
		return fmt.Errorf("get event db failed, err: %v", err)
	}

	cnt, err := ed.orm.Do(one.DB()).Exec(kt.Ctx, sql)
	if err != nil {
		return err
	}

	if cnt == 1 {
		return nil
	}

	// the cursorReminder event is lost, insert it now, normally this can not happen, because
	// this event is initialed when the database is created.
	var sqlSentenceInsert []string
	sqlSentenceInsert = append(sqlSentenceInsert, "INSERT INTO ", table.EventTable.Name(), " (id, biz_id, app_id, op_type, resource, resource_id, creator, created_at) ",
		"VALUES(", strconv.Itoa(table.EventCursorReminderPrimaryID), ", 0, 0, '', '", string(table.CursorReminder), "', ", strconv.Itoa(int(eventID)), ", '", eventUser, "', '", at, "')")
	sql = filter.SqlJoint(sqlSentenceInsert)

	cnt, err = ed.orm.Do(one.DB()).Exec(kt.Ctx, sql)
	if err != nil {
		return err
	}

	if cnt != 1 {
		return fmt.Errorf("record event cursor failed, effected rows: %d", cnt)
	}

	return nil
}

// Purge is used to try to remove number of consumed events, which is
// created before, from the oldest to now order by event id.
func (ed *eventDao) Purge(kt *kit.Kit, daysAgo uint) error {
	one := ed.sd.Event()
	if err := one.Err(); err != nil {
		return fmt.Errorf("get event db failed, err: %v", err)
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT id FROM ", table.EventTable.Name(), " WHERE created_at <= (NOW() - INTERVAL ", strconv.Itoa(int(daysAgo)), " DAY) ORDER BY id DESC LIMIT 1")
	sql := filter.SqlJoint(sqlSentence)

	lastID := uint32(0)
	if err := ed.orm.Do(one.DB()).Get(kt.Ctx, &lastID, sql); err != nil {
		if err == orm.ErrRecordNotFound {
			return nil
		}

		return fmt.Errorf("get last event id to purge failed, err: %v", err)
	}
	logs.Infof("start to delete events less than or equal last id %d, rid: %s", lastID, kt.Rid)

	const step = 100

	for {
		var sqlSentenceDel []string
		sqlSentenceDel = append(sqlSentenceDel, "DELETE FROM ", table.EventTable.Name(), " WHERE id <= ", strconv.Itoa(int(lastID)), " ORDER BY id ASC LIMIT ", strconv.Itoa(int(step)))
		sql = filter.SqlJoint(sqlSentenceDel)
		cnt, err := ed.orm.Do(one.DB()).Exec(kt.Ctx, sql)
		if err != nil {
			return fmt.Errorf("delete event failed, last id: %d, err: %v", lastID, err)
		}

		logs.Infof("deleted %d events successfully, rid: %s", cnt, kt.Rid)

		if cnt < step {
			return nil
		}

		// sleep a while before next delete step
		time.Sleep(10 * time.Second)
	}
}
