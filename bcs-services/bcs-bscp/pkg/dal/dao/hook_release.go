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
	"fmt"

	"github.com/jmoiron/sqlx"
	"gorm.io/gorm"

	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/types"
)

// HookRelease supplies all the hook release related operations.
type HookRelease interface {
	// Create one hook instance.
	Create(kit *kit.Kit, hook *table.HookRelease) (uint32, error)
	// CreateWithTx ...
	CreateWithTx(kit *kit.Kit, tx *gen.Query, g *table.HookRelease) (uint32, error)
	// Get hook release by id
	Get(kit *kit.Kit, bizID, hookID, id uint32) (*table.HookRelease, error)
	// GetByName get HookRelease by name
	GetByName(kit *kit.Kit, bizID, hookID uint32, name string) (*table.HookRelease, error)
	// List hooks with options.
	List(kit *kit.Kit, opt *types.ListHookReleasesOption) ([]*table.HookRelease, int64, error)
	// Delete one strategy instance.
	Delete(kit *kit.Kit, strategy *table.HookRelease) error
	// Publish
	// Only one version of each script is allowed to go online.
	// If an earlier version is online,
	// set the earlier version to offline
	Publish(kit *kit.Kit, g *table.HookRelease) error
	// GetByPubState hook release by PubState
	GetByPubState(kit *kit.Kit, opt *types.GetByPubStateOption) (*table.HookRelease, error)
	// DeleteByHookIDWithTx  delete release revision with transaction
	DeleteByHookIDWithTx(kit *kit.Kit, tx *gen.Query, g *table.HookRelease) error
	// PublishNumPlusOneWithTx PublishNum +1 revision with transaction
	PublishNumPlusOneWithTx(kit *kit.Kit, tx *gen.Query) error
}

// HookReleaseOption defines all the needed infos to HookRelease a resource.
type HookReleaseOption struct {
	// resource's transaction infos.
	Txn *sqlx.Tx
	// ResShardingUid is the resource's sharding instance.
	ResShardingUid string
	genQ           *gen.Query
}

var _ HookRelease = new(hookReleaseDao)

type hookReleaseDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// Create one hook instance.
func (dao *hookReleaseDao) Create(kit *kit.Kit, g *table.HookRelease) (uint32, error) {

	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}

	// generate a HookRelease id and update to HookRelease.
	id, err := dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, err
	}
	g.ID = id

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)

	// 多个使用事务处理
	createTx := func(tx *gen.Query) error {
		q := tx.HookRelease.WithContext(kit.Ctx)
		if err := q.Create(g); err != nil {
			return err
		}

		if err := ad.Do(tx); err != nil {
			return err
		}

		return nil
	}
	if err := dao.genQ.Transaction(createTx); err != nil {
		return 0, err
	}

	return g.ID, nil

}

// NewHookReleaseDao create the HookRelease DAO
func NewHookReleaseDao(db *gorm.DB, idGen IDGenInterface, auditDao AuditDao) (HookRelease, error) {
	return &hookReleaseDao{
		genQ:     gen.Use(db),
		idGen:    idGen,
		auditDao: auditDao,
	}, nil
}

// CreateWithTx ....
func (dao *hookReleaseDao) CreateWithTx(kit *kit.Kit, tx *gen.Query, g *table.HookRelease) (uint32, error) {

	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}

	// generate a HookRelease id and update to HookRelease.
	id, err := dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, err
	}
	g.ID = id

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)

	err = tx.HookRelease.WithContext(kit.Ctx).Create(g)
	if err != nil {
		return 0, err
	}
	err = ad.Do(tx)
	if err != nil {
		return 0, err
	}

	return g.ID, nil
}

// Get hookRelease by id
func (dao *hookReleaseDao) Get(kit *kit.Kit, bizID, hookID, id uint32) (*table.HookRelease, error) {

	m := dao.genQ.HookRelease
	q := dao.genQ.HookRelease.WithContext(kit.Ctx)

	tplSpace, err := q.Where(m.BizID.Eq(bizID), m.HookID.Eq(hookID), m.ID.Eq(id)).Take()
	if err != nil {
		return nil, fmt.Errorf("get hookRelease failed, err: %v", err)
	}

	return tplSpace, nil
}

// GetByName get HookRelease by name
func (dao *hookReleaseDao) GetByName(kit *kit.Kit, bizID, hookID uint32, name string) (*table.HookRelease, error) {
	m := dao.genQ.HookRelease
	q := dao.genQ.HookRelease.WithContext(kit.Ctx)

	tplSpace, err := q.Where(m.BizID.Eq(bizID), m.HookID.Eq(hookID), m.Name.Eq(name)).Take()
	if err != nil {
		return nil, fmt.Errorf("get hookRelease failed, err: %v", err)
	}

	return tplSpace, nil
}

// List hooks with options.
func (dao *hookReleaseDao) List(kit *kit.Kit,
	opt *types.ListHookReleasesOption) ([]*table.HookRelease, int64, error) {

	m := dao.genQ.HookRelease
	q := dao.genQ.HookRelease.WithContext(kit.Ctx)

	result, count, err := q.Where(
		m.BizID.Eq(opt.BizID),
		m.HookID.Eq(opt.HookID),
		m.Name.Like(fmt.Sprintf("%%%s%%", opt.SearchKey))).
		FindByPage(opt.Page.Offset(), opt.Page.LimitInt())
	if err != nil {
		return nil, 0, err
	}

	return result, count, nil

}

// Delete one strategy instance.
func (dao *hookReleaseDao) Delete(kit *kit.Kit, g *table.HookRelease) error {

	// 参数校验
	if err := g.ValidateDelete(); err != nil {
		return err
	}

	// 删除操作, 获取当前记录做审计
	m := dao.genQ.HookRelease
	q := dao.genQ.HookRelease.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareDelete(oldOne)

	// 多个使用事务处理
	deleteTx := func(tx *gen.Query) error {
		q = tx.HookRelease.WithContext(kit.Ctx)
		if _, err := q.Where(m.BizID.Eq(g.Attachment.BizID), m.ID.Eq(g.ID)).Delete(g); err != nil {
			return err
		}

		if err := ad.Do(tx); err != nil {
			return err
		}
		return nil
	}
	if err := dao.genQ.Transaction(deleteTx); err != nil {
		return err
	}

	return nil
}

// DeleteByHookIDWithTx  delete release revision with transaction
func (dao *hookReleaseDao) DeleteByHookIDWithTx(kit *kit.Kit, tx *gen.Query, g *table.HookRelease) error {

	// 参数校验
	if err := g.ValidateDeleteByHookID(); err != nil {
		return err
	}

	// 删除操作, 获取当前记录做审计
	m := dao.genQ.HookRelease
	q := dao.genQ.HookRelease.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.HookID.Eq(g.Attachment.HookID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareDelete(oldOne)

	if _, err := tx.HookRelease.WithContext(kit.Ctx).Where(m.BizID.Eq(g.Attachment.BizID), m.ID.Eq(g.ID)).Delete(g); err != nil {
		return err
	}

	if err := ad.Do(tx); err != nil {
		return err
	}

	return nil
}

// Publish
// Only one version of each script is allowed to go online.
// If an earlier version is online,
// set the earlier version to offline
func (dao *hookReleaseDao) Publish(kit *kit.Kit, g *table.HookRelease) error {

	if err := g.ValidatePublish(); err != nil {
		return err
	}

	m := dao.genQ.HookRelease
	q := dao.genQ.HookRelease.WithContext(kit.Ctx)

	currentRelease, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	g.Spec.PubState = table.PartialReleased

	oldReleased, err := q.Where(
		m.ID.Eq(g.ID),
		m.BizID.Eq(g.Attachment.BizID),
		m.PubState.Eq(table.PartialReleased.String())).
		Take()
	if err == nil {
		oldReleased.Revision = g.Revision
		oldReleased.Spec.PubState = table.FullReleased
	}

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareUpdate(g, currentRelease)

	// 多个使用事务处理
	updateTx := func(tx *gen.Query) error {
		q = tx.HookRelease.WithContext(kit.Ctx)

		if _, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Select(m.PubState, m.Reviser).Updates(g); err != nil {
			return err
		}

		if oldReleased != nil {
			if _, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Select(m.PubState, m.Reviser).
				Updates(oldReleased); err != nil {
				return err
			}
		}

		if err := ad.Do(tx); err != nil {
			return err
		}
		return nil

	}

	if err := dao.genQ.Transaction(updateTx); err != nil {
		return err
	}

	return nil
}

// GetByPubState hook release by PubState
func (dao *hookReleaseDao) GetByPubState(kit *kit.Kit,
	opt *types.GetByPubStateOption) (*table.HookRelease, error) {

	// 参数校验
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	m := dao.genQ.HookRelease
	q := dao.genQ.HookRelease.WithContext(kit.Ctx)

	release, err := q.Where(
		m.BizID.Eq(opt.BizID),
		m.HookID.Eq(opt.HookID),
		m.PubState.Eq(opt.State.String()),
	).Take()
	if err != nil {
		return nil, fmt.Errorf("get pubState(%s) hook(%d) failed, err: %v", opt.State, opt.HookID, err)
	}

	return release, nil

}

// PublishNumPlusOneWithTx PublishNum +1 revision with transaction
func (dao *hookReleaseDao) PublishNumPlusOneWithTx(kit *kit.Kit, tx *gen.Query) error {

	m := tx.HookRelease
	_, err := tx.WithContext(kit.Ctx).Hook.Update(m.PublishNum, gorm.Expr("publish_num + ?", 1))
	return err
}
