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
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"

	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/orm"
	"bscp.io/pkg/dal/sharding"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/types"

	"github.com/jmoiron/sqlx"
)

// App supplies all the app related operations.
type App interface {
	// Create one app instance
	Create(kit *kit.Kit, app *table.App) (uint32, error)
	// Update one app's info
	Update(kit *kit.Kit, app *table.App) error
	// get app with id.
	Get(kit *kit.Kit, bizID, appID uint32) (*table.App, error)
	// get app only with id.
	GetByID(kit *kit.Kit, appID uint32) (*table.App, error)
	// get app by name.
	GetByName(kit *kit.Kit, bizID uint32, name string) (*table.App, error)
	// List apps with options.
	List(kit *kit.Kit, opts *types.ListAppsOption) (*types.ListAppDetails, error)
	// ListAppsByGroupID list apps by group id.
	ListAppsByGroupID(kit *kit.Kit, groupID, bizID uint32) ([]*table.App, error)
	// Delete one app instance.
	Delete(kit *kit.Kit, app *table.App) error
	// ListAppMetaForCache list app's basic meta info.
	ListAppMetaForCache(kt *kit.Kit, bizID uint32, appID []uint32) (map[ /*appID*/ uint32]*types.AppCacheMeta, error)
}

var _ App = new(appDao)

type appDao struct {
	orm      orm.Interface
	sd       *sharding.Sharding
	idGen    IDGenInterface
	auditDao AuditDao
	event    Event
}

// List app's detail info with the filter's expression.
func (ap *appDao) List(kit *kit.Kit, opts *types.ListAppsOption) (*types.ListAppDetails, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list app options is nil")
	}

	po := &types.PageOption{
		EnableUnlimitedLimit: true,
		DisabledSort:         false,
	}

	if err := opts.Validate(po); err != nil {
		return nil, err
	}

	sqlOpt := &filter.SQLWhereOption{
		Priority: filter.Priority{"id", "biz_id"},
		CrownedOption: &filter.CrownedOption{
			CrownedOp: filter.And,
			Rules:     []filter.RuleFactory{},
		},
	}

	// 导航查询场景
	if len(opts.BizList) > 1 {
		sqlOpt.CrownedOption.Rules = []filter.RuleFactory{
			&filter.AtomRule{
				Field: "biz_id",
				Op:    filter.OpFactory(filter.In),
				Value: opts.BizList,
			},
		}
	} else {
		sqlOpt.CrownedOption.Rules = []filter.RuleFactory{
			&filter.AtomRule{
				Field: "biz_id",
				Op:    filter.OpFactory(filter.Equal),
				Value: opts.BizID,
			},
		}
	}
	whereExpr, args, err := opts.Filter.SQLWhereExpr(sqlOpt)
	if err != nil {
		return nil, err
	}

	// 如果 app 有分库分表, 跨 spaces 查询将不可用
	// do count operation only.
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT COUNT(*) FROM ", table.AppTable.Name(), whereExpr)
	countSql := filter.SqlJoint(sqlSentence)
	count, err := ap.orm.Do(ap.sd.ShardingOne(opts.BizID).DB()).Count(kit.Ctx, countSql, args...)
	if err != nil {
		return nil, err
	}

	// query app list for now.
	pageExpr, err := opts.Page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	var sqlQuery []string
	sqlQuery = append(sqlQuery, "SELECT ", table.AppColumns.NamedExpr(), " FROM ", table.AppTable.Name(), whereExpr, pageExpr)
	querySql := filter.SqlJoint(sqlQuery)

	list := make([]*table.App, 0)
	err = ap.orm.Do(ap.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, querySql, args...)
	if err != nil {
		return nil, err
	}

	return &types.ListAppDetails{Count: count, Details: list}, nil
}

// ListAppsByGroupID list apps by group id.
func (ap *appDao) ListAppsByGroupID(kit *kit.Kit, groupID, bizID uint32) ([]*table.App, error) {

	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is 0")
	}
	if groupID == 0 {
		return nil, errf.New(errf.InvalidParameter, "group id is 0")
	}

	group := &table.Group{}
	var sqlBuf bytes.Buffer
	sqlBuf.WriteString("SELECT ")
	sqlBuf.WriteString(table.GroupColumns.NamedExpr())
	sqlBuf.WriteString(" FROM ")
	sqlBuf.WriteString(table.GroupTable.Name())
	sqlBuf.WriteString(" WHERE biz_id = ? AND id = ?")

	err := ap.orm.Do(ap.sd.ShardingOne(bizID).DB()).Get(kit.Ctx, group, sqlBuf.String(), bizID, groupID)
	if err != nil {
		return nil, err
	}
	list := make([]*table.App, 0)

	if group.Spec.Public {
		var sqlSentence []string
		sqlSentence = append(sqlSentence, "SELECT ", table.AppColumns.NamedExpr(), " FROM ", table.AppTable.Name(),
			" WHERE biz_id = ?")
		sql := filter.SqlJoint(sqlSentence)

		if err := ap.orm.Do(ap.sd.ShardingOne(bizID).DB()).Select(kit.Ctx, &list, sql, bizID); err != nil {
			return nil, err
		}
	} else {
		var sqlSentence []string
		sqlSentence = append(sqlSentence, "SELECT ", table.AppColumns.NamedExpr(), " FROM ", table.AppTable.Name(),
			" WHERE biz_id = ? AND id IN (SELECT app_id FROM ", table.GroupAppBindTable.Name(), " WHERE group_id = ?)")
		sql := filter.SqlJoint(sqlSentence)

		err := ap.orm.Do(ap.sd.ShardingOne(bizID).DB()).Select(kit.Ctx, &list, sql, bizID, groupID)
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}

// Create one app instance
func (ap *appDao) Create(kit *kit.Kit, app *table.App) (uint32, error) {

	if app == nil {
		return 0, errf.New(errf.InvalidParameter, "app is nil")
	}

	if err := app.ValidateCreate(); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	// generate an app id and update to app.
	id, err := ap.idGen.One(kit, table.AppTable)
	if err != nil {
		return 0, err
	}

	app.ID = id

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "INSERT INTO ", table.AppTable.Name(),
		" (", table.AppColumns.ColumnExpr(), ") ", "VALUES(", table.AppColumns.ColonNameExpr(), ")")
	sql := filter.SqlJoint(sqlSentence)
	eDecorator := ap.event.Eventf(kit)
	err = ap.sd.ShardingOne(app.BizID).AutoTxn(kit, func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
		if err := ap.orm.Txn(txn).Insert(kit.Ctx, sql, app); err != nil {
			return err
		}

		// audit this to be create app details.
		au := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
		if err = ap.auditDao.Decorator(kit, app.BizID, enumor.App).AuditCreate(app, au); err != nil {
			return fmt.Errorf("audit create app failed, err: %v", err)
		}

		// fire the event with txn to ensure the if save the event failed then the business logic is failed anyway.
		one := types.Event{
			Spec: &table.EventSpec{
				Resource:   table.Application,
				ResourceID: app.ID,
				OpType:     table.InsertOp,
			},
			Attachment: &table.EventAttachment{BizID: app.BizID, AppID: app.ID},
			Revision:   &table.CreatedRevision{Creator: kit.User, CreatedAt: time.Now()},
		}
		if err = eDecorator.Fire(one); err != nil {
			logs.Errorf("fire create app: %s event failed, err: %v, rid: %s", app.ID, err, kit.Rid)
			return errf.New(errf.DBOpFailed, "fire event failed, "+err.Error())
		}
		return nil
	})

	eDecorator.Finalizer(err)

	if err != nil {
		logs.Errorf("create app, but do auto txn failed, err: %v, rid: %s", err, kit.Rid)
		return 0, fmt.Errorf("create app, but auto run txn failed, err: %v", err)
	}

	return id, nil
}

// Update an app instance.
func (ap *appDao) Update(kit *kit.Kit, app *table.App) error {

	if app == nil {
		return errf.New(errf.InvalidParameter, "app is nil")
	}

	updateApp, err := ap.Get(kit, app.BizID, app.ID)
	if err != nil {
		return fmt.Errorf("get update app failed, err: %v", err)
	}

	if err := app.ValidateUpdate(updateApp.Spec.ConfigType); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	opts := orm.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields("id", "biz_id")
	expr, toUpdate, err := orm.RearrangeSQLDataWithOption(app, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql expr failed, err: %v", err)
	}

	ab := ap.auditDao.Decorator(kit, app.BizID, enumor.App).PrepareUpdate(app)

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "UPDATE ", table.AppTable.Name(), " SET ", expr, " WHERE id = ",
		strconv.Itoa(int(app.ID)), " and biz_id = ", strconv.Itoa(int(app.BizID)))
	sql := filter.SqlJoint(sqlSentence)

	eDecorator := ap.event.Eventf(kit)
	err = ap.sd.ShardingOne(app.BizID).AutoTxn(kit, func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
		effected, err := ap.orm.Txn(txn).Update(kit.Ctx, sql, toUpdate)
		if err != nil {
			logs.Errorf("update app: %d failed, err: %v, rid: %v", app.ID, err, kit.Rid)
			return err
		}

		if effected == 0 {
			logs.Errorf("update one app: %d, but record not found, rid: %v", app.ID, kit.Rid)
			return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
		}

		if effected > 1 {
			logs.Errorf("update one app: %d, but got updated app count: %d, rid: %v", app.ID, effected, kit.Rid)
			return fmt.Errorf("matched app count %d is not as excepted", effected)
		}

		// do audit
		if err := ab.Do(&AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}); err != nil {
			return fmt.Errorf("do app update audit failed, err: %v", err)
		}

		// fire the event with txn to ensure the if save the event failed then the business logic is failed anyway.
		one := types.Event{
			Spec: &table.EventSpec{
				Resource:   table.Application,
				ResourceID: app.ID,
				OpType:     table.UpdateOp,
			},
			Attachment: &table.EventAttachment{BizID: app.BizID, AppID: app.ID},
			Revision:   &table.CreatedRevision{Creator: kit.User, CreatedAt: time.Now()},
		}
		if err := eDecorator.Fire(one); err != nil {
			logs.Errorf("fire update app: %s event failed, err: %v, rid: %s", app.ID, err, kit.Rid)
			return errf.New(errf.DBOpFailed, "fire event failed, "+err.Error())
		}
		return nil
	})

	eDecorator.Finalizer(err)

	if err != nil {
		return err
	}

	return nil
}

// Delete an app instance.
func (ap *appDao) Delete(kit *kit.Kit, app *table.App) error {

	if app == nil {
		return errf.New(errf.InvalidParameter, "app is nil")
	}

	if err := app.ValidateDelete(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	ab := ap.auditDao.Decorator(kit, app.BizID, enumor.App).PrepareDelete(app.ID)

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "DELETE FROM ", table.AppTable.Name(), " WHERE id = ",
		strconv.Itoa(int(app.ID)), " AND biz_id = ", strconv.Itoa(int(app.BizID)))
	sql := filter.SqlJoint(sqlSentence)

	eDecorator := ap.event.Eventf(kit)
	err := ap.sd.ShardingOne(app.BizID).AutoTxn(kit, func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
		oldApp, err := ap.Get(kit, app.BizID, app.ID)
		if err != nil {
			return fmt.Errorf("get pre app failed, err: %v", err)
		}

		// delete the app at first.
		err = ap.orm.Txn(txn).Delete(kit.Ctx, sql)
		if err != nil {
			return err
		}

		// archived this deleted app to archive table.
		if err := ap.archiveApp(kit, txn, oldApp); err != nil {
			return err
		}

		// audit this delete app details.
		auditOpt := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
		if err := ab.Do(auditOpt); err != nil {
			return fmt.Errorf("audit delete app failed, err: %v", err)
		}

		// fire the event with txn to ensure the if save the event failed then the business logic is failed anyway.
		one := types.Event{
			Spec: &table.EventSpec{
				Resource:   table.Application,
				ResourceID: app.ID,
				OpType:     table.DeleteOp,
			},
			Attachment: &table.EventAttachment{BizID: app.BizID, AppID: app.ID},
			Revision:   &table.CreatedRevision{Creator: kit.User, CreatedAt: time.Now()},
		}
		if err := eDecorator.Fire(one); err != nil {
			logs.Errorf("fire delete app: %s event failed, err: %v, rid: %s", app.ID, err, kit.Rid)
			return errf.New(errf.DBOpFailed, "fire event failed, "+err.Error())
		}

		return nil
	})

	eDecorator.Finalizer(err)

	if err != nil {
		logs.Errorf("delete app: %d failed, err: %v, rid: %v", app.ID, err, kit.Rid)
		return fmt.Errorf("delete app, but run txn failed, err: %v", err)
	}

	return nil
}

func (ap *appDao) Get(kit *kit.Kit, bizID uint32, appID uint32) (*table.App, error) {

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.AppColumns.NamedExpr(), " FROM ",
		table.AppTable.Name(), " WHERE id = ", strconv.Itoa(int(appID)), " AND biz_id = ", strconv.Itoa(int(bizID)))
	sql := filter.SqlJoint(sqlSentence)

	one := new(table.App)
	err := ap.orm.Do(ap.sd.MustSharding(bizID)).Get(kit.Ctx, one, sql)
	if err != nil {
		return nil, fmt.Errorf("get app details failed, err: %v", err)
	}

	return one, nil
}

// GetByID 通过 AppId 查询
func (ap *appDao) GetByID(kit *kit.Kit, appID uint32) (*table.App, error) {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.AppColumns.NamedExpr(), " FROM ", table.AppTable.Name(),
		" WHERE id = ", strconv.Itoa(int(appID)))
	expr := filter.SqlJoint(sqlSentence)
	one := new(table.App)
	err := ap.orm.Do(ap.sd.Admin().DB()).Get(kit.Ctx, one, expr)
	if err != nil {
		return nil, fmt.Errorf("get app details failed, err: %v", err)
	}

	return one, nil
}

// GetByName 通过 name 查询
func (ap *appDao) GetByName(kit *kit.Kit, bizID uint32, name string) (*table.App, error) {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.AppColumns.NamedExpr(), " FROM ", table.AppTable.Name(),
		" WHERE name = '", name, "' AND biz_id = ", strconv.Itoa(int(bizID)))
	expr := filter.SqlJoint(sqlSentence)
	one := new(table.App)
	err := ap.orm.Do(ap.sd.Admin().DB()).Get(kit.Ctx, one, expr)
	if err != nil {
		return nil, fmt.Errorf("get app details failed, err: %v", err)
	}

	return one, nil
}

func getAppMode(kit *kit.Kit, orm orm.Interface, sd *sharding.Sharding, bizID, appID uint32) (table.AppMode, error) {

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.AppColumns.NamedExpr(), " FROM ", table.AppTable.Name(),
		" WHERE id = ", strconv.Itoa(int(appID)), " AND biz_id = ", strconv.Itoa(int(bizID)))
	sql := filter.SqlJoint(sqlSentence)
	one := new(table.App)
	err := orm.Do(sd.MustSharding(bizID)).Get(kit.Ctx, one, sql)
	if err != nil {
		return "", errf.New(errf.DBOpFailed, fmt.Sprintf("get app mode failed, err: %v", err))
	}

	if err := one.Spec.Mode.Validate(); err != nil {
		return "", errf.New(errf.InvalidParameter, err.Error())
	}

	return one.Spec.Mode, nil
}

func (ap *appDao) archiveApp(kit *kit.Kit, txn *sqlx.Tx, app *table.App) error {

	id, err := ap.idGen.One(kit, table.ArchivedAppTable)
	if err != nil {
		return err
	}

	archivedApp := &table.ArchivedApp{
		ID:        id,
		AppID:     app.ID,
		BizID:     app.BizID,
		CreatedAt: time.Now(),
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "INSERT INTO ", table.ArchivedAppTable.Name(),
		" (", table.ArchivedAppColumns.ColumnExpr(), ") ", "VALUES(", table.ArchivedAppColumns.ColonNameExpr(), ")")
	sql := filter.SqlJoint(sqlSentence)
	err = ap.orm.Txn(txn).Insert(kit.Ctx, sql, archivedApp)
	if err != nil {
		return fmt.Errorf("archived delete app failed, err: %v", err)
	}

	return nil
}

// ListAppMetaForCache list app's basic meta info.
func (ap *appDao) ListAppMetaForCache(kt *kit.Kit, bizID uint32, appIDs []uint32) (
	map[uint32]*types.AppCacheMeta, error) {

	if bizID <= 0 || len(appIDs) == 0 {
		return nil, errors.New("invalid biz id or app id list")
	}

	appIDList := tools.JoinUint32(appIDs, ",")
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT id, name AS 'spec.name', config_type AS 'spec.config_type', mode AS 'spec.mode', reload_type AS ",
		"'spec.reload.reload_type', reload_file_path AS 'spec.reload.file_reload_spec.reload_file_path' ",
		"FROM ", table.AppTable.Name(), " WHERE id IN (", appIDList, ") AND biz_id = ", strconv.Itoa(int(bizID)))
	sql := filter.SqlJoint(sqlSentence)
	appList := make([]*table.App, 0)
	if err := ap.orm.Do(ap.sd.MustSharding(bizID)).Select(kt.Ctx, &appList, sql); err != nil {
		return nil, fmt.Errorf("query db with app failed, err: %v", err)
	}

	meta := make(map[uint32]*types.AppCacheMeta)
	for _, one := range appList {
		meta[one.ID] = &types.AppCacheMeta{
			Name:       one.Spec.Name,
			ConfigType: one.Spec.ConfigType,
			Mode:       one.Spec.Mode,
			Reload:     one.Spec.Reload,
		}
	}

	return meta, nil
}
