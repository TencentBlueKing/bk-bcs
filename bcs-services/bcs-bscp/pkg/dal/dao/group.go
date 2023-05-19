/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package dao

import (
	"fmt"
	"strconv"

	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/orm"
	"bscp.io/pkg/dal/sharding"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/types"
)

// Group supplies all the group related operations.
type Group interface {
	// CreateWithTx Create one group instance with transaction.
	CreateWithTx(kit *kit.Kit, tx *sharding.Tx, group *table.Group) (uint32, error)
	// UpdateWithTx Update one group instance with transaction.
	UpdateWithTx(kit *kit.Kit, tx *sharding.Tx, group *table.Group) error
	// Get group by id.
	Get(kit *kit.Kit, id, bizID uint32) (*table.Group, error)
	// GetByName get group by name.
	GetByName(kit *kit.Kit, bizID uint32, name string) (*table.Group, error)
	// List groups with options.
	List(kit *kit.Kit, opts *types.ListGroupsOption) (*types.ListGroupDetails, error)
	// DeleteWithTx delete one group instance with transaction.
	DeleteWithTx(kit *kit.Kit, tx *sharding.Tx, group *table.Group) error
	// ListAppGroups list all the groups of the app.
	ListAppGroups(kit *kit.Kit, bizID, appID uint32) ([]*table.Group, error)
	// ListGroupRleasesdApps list all the released apps of the group.
	ListGroupRleasesdApps(kit *kit.Kit, opts *types.ListGroupRleasesdAppsOption) (
		*types.ListGroupRleasesdAppsDetails, error)
}

var _ Group = new(groupDao)

type groupDao struct {
	orm      orm.Interface
	sd       *sharding.Sharding
	idGen    IDGenInterface
	auditDao AuditDao
	lock     LockDao
}

// CreateWithTx Create one group instance with transaction.
func (dao *groupDao) CreateWithTx(kit *kit.Kit, tx *sharding.Tx, g *table.Group) (uint32, error) {

	if g == nil {
		return 0, errf.New(errf.InvalidParameter, "group is nil")
	}

	if err := g.ValidateCreate(); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	// generate a group id and update to group.
	id, err := dao.idGen.One(kit, table.GroupTable)
	if err != nil {
		return 0, err
	}

	g.ID = id
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "INSERT INTO ", table.GroupTable.Name(),
		" (", table.GroupColumns.ColumnExpr(), ")  VALUES(", table.GroupColumns.ColonNameExpr(), ")")

	sql := filter.SqlJoint(sqlSentence)

	if iErr := dao.orm.Txn(tx.Tx()).Insert(kit.Ctx, sql, g); iErr != nil {
		return 0, iErr
	}

	// audit this to be created group details.
	au := &AuditOption{Txn: tx.Tx(), ResShardingUid: tx.ShardingUid()}
	if err = dao.auditDao.Decorator(kit, g.Attachment.BizID,
		enumor.Group).AuditCreate(g, au); err != nil {
		return 0, fmt.Errorf("audit create group failed, err: %v", err)
	}

	if err != nil {
		logs.Errorf("create group, but do auto txn failed, err: %v, rid: %s", err, kit.Rid)
		return 0, fmt.Errorf("create group, but auto run txn failed, err: %v", err)
	}

	return id, nil
}

// UpdateWithTx Update one group instance with transaction.
func (dao *groupDao) UpdateWithTx(kit *kit.Kit, tx *sharding.Tx, g *table.Group) error {

	if g == nil {
		return errf.New(errf.InvalidParameter, "group is nil")
	}

	if err := g.ValidateUpdate(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	opts := orm.NewFieldOptions().AddIgnoredFields("id", "biz_id")
	expr, toUpdate, err := orm.RearrangeSQLDataWithOption(g, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql expr failed, err: %v", err)
	}

	ab := dao.auditDao.Decorator(kit, g.Attachment.BizID, enumor.Group).PrepareUpdate(g)

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "UPDATE ", table.GroupTable.Name(), " SET ", expr,
		" WHERE id = ", strconv.Itoa(int(g.ID)), " AND biz_id = ", strconv.Itoa(int(g.Attachment.BizID)))
	sql := filter.SqlJoint(sqlSentence)

	var effected int64
	effected, err = dao.orm.Txn(tx.Tx()).Update(kit.Ctx, sql, toUpdate)
	if err != nil {
		logs.Errorf("update group: %d failed, err: %v, rid: %v", g.ID, err, kit.Rid)
		return err
	}

	if effected == 0 {
		logs.Errorf("update one group: %d, but record not found, rid: %v", g.ID, kit.Rid)
		return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
	}

	if effected > 1 {
		logs.Errorf("update one group: %d, but got updated group count: %d, rid: %v", g.ID,
			effected, kit.Rid)
		return fmt.Errorf("matched group count %d is not as excepted", effected)
	}

	// do audit
	if dErr := ab.Do(&AuditOption{Txn: tx.Tx(), ResShardingUid: tx.ShardingUid()}); dErr != nil {
		return fmt.Errorf("do group update audit failed, err: %v", dErr)
	}

	if err != nil {
		return err
	}

	return nil
}

// Get group by id.
func (dao *groupDao) Get(kit *kit.Kit, id, bizID uint32) (*table.Group, error) {

	if bizID == 0 || id == 0 {
		return nil, errf.New(errf.InvalidParameter, "bizID or id is 0")
	}

	if id == 0 {
		return nil, errf.New(errf.InvalidParameter, "group id can not be 0")
	}
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.GroupColumns.NamedExpr(), " FROM ", table.GroupTable.Name(),
		" WHERE id = ", strconv.Itoa(int(id)))
	sql := filter.SqlJoint(sqlSentence)

	group := &table.Group{}
	if err := dao.orm.Do(dao.sd.ShardingOne(bizID).DB()).Get(kit.Ctx, group, sql); err != nil {
		return nil, err
	}
	return group, nil
}

// GetByName get group by name.
func (dao *groupDao) GetByName(kit *kit.Kit, bizID uint32, name string) (*table.Group, error) {

	if bizID == 0 || name == "" {
		return nil, errf.New(errf.InvalidParameter, "biz id or name is empty")
	}
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.GroupColumns.NamedExpr(), " FROM ", table.GroupTable.Name(),
		" WHERE name = '", name, "' AND biz_id = ", strconv.Itoa(int(bizID)))
	sql := filter.SqlJoint(sqlSentence)

	group := &table.Group{}
	if err := dao.orm.Do(dao.sd.ShardingOne(bizID).DB()).Get(kit.Ctx, group, sql); err != nil {
		return nil, err
	}
	return group, nil
}

// List groups with options.
func (dao *groupDao) List(kit *kit.Kit, opts *types.ListGroupsOption) (
	*types.ListGroupDetails, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list group options null")
	}

	po := &types.PageOption{
		EnableUnlimitedLimit: true,
		DisabledSort:         false,
	}

	if err := opts.Validate(po); err != nil {
		return nil, err
	}

	sqlOpt := &filter.SQLWhereOption{
		Priority: filter.Priority{"id", "biz_id", "app_id"},
		CrownedOption: &filter.CrownedOption{
			CrownedOp: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "biz_id",
					Op:    filter.Equal.Factory(),
					Value: opts.BizID,
				},
			},
		},
	}
	whereExpr, args, err := opts.Filter.SQLWhereExpr(sqlOpt)
	if err != nil {
		return nil, err
	}
	var sqlSentenceCount []string
	sqlSentenceCount = append(sqlSentenceCount, "SELECT COUNT(*) FROM ", table.GroupTable.Name(), whereExpr)
	countSql := filter.SqlJoint(sqlSentenceCount)
	count, err := dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Count(kit.Ctx, countSql, args...)
	if err != nil {
		return nil, err
	}

	// query group list for now.
	pageExpr, err := opts.Page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.GroupColumns.NamedExpr(), " FROM ", table.GroupTable.Name(), whereExpr, pageExpr)
	sql := filter.SqlJoint(sqlSentence)

	list := make([]*table.Group, 0)
	err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, sql, args...)
	if err != nil {
		return nil, err
	}

	return &types.ListGroupDetails{Count: count, Details: list}, nil
}

// DeleteWithTx delete group with transaction.
func (dao *groupDao) DeleteWithTx(kit *kit.Kit, tx *sharding.Tx, g *table.Group) error {

	if g == nil {
		return errf.New(errf.InvalidParameter, "group is nil")
	}

	if err := g.ValidateDelete(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	ab := dao.auditDao.Decorator(kit, g.Attachment.BizID, enumor.Group).PrepareDelete(g.ID)
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "DELETE FROM ", table.GroupTable.Name(), " WHERE id = ", strconv.Itoa(int(g.ID)),
		" AND biz_id = ", strconv.Itoa(int(g.Attachment.BizID)))
	expr := filter.SqlJoint(sqlSentence)

	// delete the group at first.
	err := dao.orm.Txn(tx.Tx()).Delete(kit.Ctx, expr)
	if err != nil {
		return err
	}

	// audit this delete group details.
	auditOpt := &AuditOption{Txn: tx.Tx(), ResShardingUid: tx.ShardingUid()}
	if err := ab.Do(auditOpt); err != nil {
		return fmt.Errorf("audit delete group failed, err: %v", err)
	}
	return nil
}

// ListAppGroups list groups by app id.
func (dao *groupDao) ListAppGroups(kit *kit.Kit, bizID, appID uint32) ([]*table.Group, error) {

	if bizID == 0 || appID == 0 {
		return nil, errf.New(errf.InvalidParameter, "bizID or appID is 0")
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.GroupColumns.NamedExpr(), " FROM ", table.GroupTable.Name(),
		" WHERE biz_id = ", strconv.Itoa(int(bizID)), " AND (public = true OR id IN ",
		"(SELECT group_id FROM group_app_binds WHERE biz_id = ", strconv.Itoa(int(bizID)),
		" AND app_id =", strconv.Itoa(int(appID)), "))")
	sql := filter.SqlJoint(sqlSentence)

	list := make([]*table.Group, 0)
	err := dao.orm.Do(dao.sd.ShardingOne(bizID).DB()).Select(kit.Ctx, &list, sql)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// ListGroupRleasesdApps list group released apps and their latest release info.
func (dao *groupDao) ListGroupRleasesdApps(kit *kit.Kit, opts *types.ListGroupRleasesdAppsOption) (
	*types.ListGroupRleasesdAppsDetails, error) {
	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list group released apps options null")
	}
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	var countSqlSentence []string
	countSqlSentence = append(countSqlSentence, "SELECT COUNT(*) FROM ", table.AppTable.Name(), " a JOIN ",
		table.ReleaseTable.Name(), " r ON a.id = r.app_id JOIN ", table.ReleasedGroupTable.Name(),
		" g ON r.id = g.release_id AND a.id = g.app_id ", " WHERE g.group_id = ", strconv.Itoa(int(opts.GroupID)),
		" AND a.biz_id = ", strconv.Itoa(int(opts.BizID)), " AND r.biz_id = ", strconv.Itoa(int(opts.BizID)),
		" AND g.biz_id = ", strconv.Itoa(int(opts.BizID)),
	)
	countSql := filter.SqlJoint(countSqlSentence)
	count, err := dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Count(kit.Ctx, countSql)
	if err != nil {
		return nil, err
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT a.id AS app_id, a.name AS app_name, r.id AS release_id, ",
		"r.name AS release_name, g.edited as edited ", " FROM ", table.AppTable.Name(), " a JOIN ",
		table.ReleaseTable.Name(), " r ON a.id = r.app_id JOIN ", table.ReleasedGroupTable.Name(),
		" g ON r.id = g.release_id AND a.id = g.app_id ", " WHERE g.group_id = ", strconv.Itoa(int(opts.GroupID)),
		" AND a.biz_id = ", strconv.Itoa(int(opts.BizID)), " AND r.biz_id = ", strconv.Itoa(int(opts.BizID)),
		" AND g.biz_id = ", strconv.Itoa(int(opts.BizID)),
		" LIMIT ", strconv.Itoa(int(opts.Start)), ", ", strconv.Itoa(int(opts.Limit)),
	)
	sql := filter.SqlJoint(sqlSentence)

	list := make([]*types.ListGroupRleasesdAppsData, 0)
	if err := dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, sql); err != nil {
		return nil, err
	}
	return &types.ListGroupRleasesdAppsDetails{
		Count:   count,
		Details: list,
	}, nil
}
