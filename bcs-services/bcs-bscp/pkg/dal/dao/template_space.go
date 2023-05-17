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

	"github.com/jmoiron/sqlx"

	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/orm"
	"bscp.io/pkg/dal/sharding"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/filter"
)

// TemplateSpace supplies all the TemplateSpace related operations.
type TemplateSpace interface {
	// Create one TemplateSpace instance.
	Create(kit *kit.Kit, TemplateSpace *table.TemplateSpace) (uint32, error)
	// Update one TemplateSpace's info.
	Update(kit *kit.Kit, TemplateSpace *table.TemplateSpace) error
	// List TemplateSpaces with options.
	List(kit *kit.Kit, bizID uint32, offset, limit int) ([]*templateSpaceDao, int64, error)
	// Delete one strategy instance.
	Delete(kit *kit.Kit, strategy *table.TemplateSpace) error
	// GetByName get templateSpace by name.
	GetByName(kit *kit.Kit, bizID uint32, name string) (*table.TemplateSpace, error)
}

var _ TemplateSpace = new(templateSpaceDao)

type templateSpaceDao struct {
	orm      orm.Interface
	sd       *sharding.Sharding
	idGen    IDGenInterface
	auditDao AuditDao
	genM     *gen.Query
}

// Create one TemplateSpace instance.
func (dao *templateSpaceDao) Create(kit *kit.Kit, g *table.TemplateSpace) (uint32, error) {
	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}

	// generate a TemplateSpace id and update to TemplateSpace.
	id, err := dao.idGen.One(kit, table.TemplateSpaceTable)
	if err != nil {
		return 0, err
	}

	g.ID = id

	m := gen.TemplateSpace
	q := gen.Q.WithContext(kit.Ctx)

	q.TemplateSpace.Create(g)

	err = dao.sd.ShardingOne(g.Attachment.BizID).AutoTxn(kit,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			if err := dao.orm.Txn(txn).Insert(kit.Ctx, sql, g); err != nil {
				return err
			}

			// audit this to be created TemplateSpace details.
			au := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
			if err = dao.auditDao.Decorator(kit, g.Attachment.BizID, enumor.TemplateSpace).AuditCreate(g, au); err != nil {
				return fmt.Errorf("audit create TemplateSpace failed, err: %v", err)
			}

			return nil
		})

	if err != nil {
		logs.Errorf("create TemplateSpace, but do auto txn failed, err: %v, rid: %s", err, kit.Rid)
		return 0, fmt.Errorf("create TemplateSpace, but auto run txn failed, err: %v", err)
	}

	return id, nil
}

// Update one TemplateSpace instance.
func (dao *templateSpaceDao) Update(kit *kit.Kit, g *table.TemplateSpace) error {

	if g == nil {
		return errf.New(errf.InvalidParameter, "TemplateSpace is nil")
	}

	if err := g.ValidateUpdate(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	opts := orm.NewFieldOptions().AddIgnoredFields(
		"id", "biz_id")
	expr, toUpdate, err := orm.RearrangeSQLDataWithOption(g, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql expr failed, err: %v", err)
	}

	ab := dao.auditDao.Decorator(kit, g.Attachment.BizID, enumor.TemplateSpace).PrepareUpdate(g)

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "UPDATE ", table.TemplateSpaceTable.Name(), " SET ", expr, " WHERE id = ", strconv.Itoa(int(g.ID)),
		" AND biz_id = ", strconv.Itoa(int(g.Attachment.BizID)))
	sql := filter.SqlJoint(sqlSentence)

	err = dao.sd.ShardingOne(g.Attachment.BizID).AutoTxn(kit,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			var effected int64
			effected, err = dao.orm.Txn(txn).Update(kit.Ctx, sql, toUpdate)
			if err != nil {
				logs.Errorf("update TemplateSpace: %d failed, err: %v, rid: %v", g.ID, err, kit.Rid)
				return err
			}

			if effected == 0 {
				logs.Errorf("update one TemplateSpace: %d, but record not found, rid: %v", g.ID, kit.Rid)
				return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
			}

			if effected > 1 {
				logs.Errorf("update one TemplateSpace: %d, but got updated TemplateSpace count: %d, rid: %v", g.ID,
					effected, kit.Rid)
				return fmt.Errorf("matched TemplateSpace count %d is not as excepted", effected)
			}

			// do audit
			if err := ab.Do(&AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}); err != nil {
				return fmt.Errorf("do TemplateSpace update audit failed, err: %v", err)
			}

			return nil
		})

	if err != nil {
		return err
	}

	return nil
}

// List TemplateSpaces with options.
func (dao *templateSpaceDao) List(kit *kit.Kit, bizID uint32, offset, limit int) ([]*table.TemplateSpace, int64, error) {
	m := dao.genM.TemplateSpace
	q := m.WithContext(kit.Ctx)

	result, count, err := q.Where(m.BizID.Eq(bizID)).FindByPage(offset, limit)
	if err != nil {
		return nil, 0, err
	}

	return result, count, nil
}

// Delete one TemplateSpace instance.
func (dao *templateSpaceDao) Delete(kit *kit.Kit, g *table.TemplateSpace) error {
	if err := g.ValidateDelete(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	ab := dao.auditDao.Decorator(kit, g.Attachment.BizID, enumor.TemplateSpace).PrepareDelete(g.ID)

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "DELETE FROM ", table.TemplateSpaceTable.Name(), " WHERE id = ", strconv.Itoa(int(g.ID)),
		" AND biz_id = ", strconv.Itoa(int(g.Attachment.BizID)))
	expr := filter.SqlJoint(sqlSentence)

	err := dao.sd.ShardingOne(g.Attachment.BizID).AutoTxn(kit, func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
		// delete the TemplateSpace at first.
		err := dao.orm.Txn(txn).Delete(kit.Ctx, expr)
		if err != nil {
			return err
		}

		// audit this delete TemplateSpace details.
		auditOpt := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
		if err := ab.Do(auditOpt); err != nil {
			return fmt.Errorf("audit delete TemplateSpace failed, err: %v", err)
		}

		return nil
	})

	if err != nil {
		logs.Errorf("delete TemplateSpace: %d failed, err: %v, rid: %v", g.ID, err, kit.Rid)
		return fmt.Errorf("delete TemplateSpace, but run txn failed, err: %v", err)
	}

	return nil
}

// GetByName get by name
func (dao *templateSpaceDao) GetByName(kit *kit.Kit, bizID uint32, name string) (*table.TemplateSpace, error) {
	m := dao.genM.TemplateSpace
	q := m.WithContext(kit.Ctx)

	tplSpace, err := q.Where(m.BizID.Eq(bizID), m.Name.Eq(name)).Take()
	if err != nil {
		return nil, fmt.Errorf("get templateSpace failed, err: %v", err)
	}

	return tplSpace, nil
}
