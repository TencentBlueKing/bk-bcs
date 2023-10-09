package dao

import (
	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/types"
	"fmt"
)

// Kv supplies all the kv related operations.
type Kv interface {
	// Create one kv instance
	Create(kit *kit.Kit, kv *table.Kv) (uint32, error)
	// Update one kv's info
	Update(kit *kit.Kit, kv *table.Kv) error
	// List kv with options.
	List(kit *kit.Kit, opt *types.ListKvOption) ([]*table.Kv, int64, error)
	// Delete ..
	Delete(kit *kit.Kit, kv *table.Kv) error
	// GetByUniqueKey get kv by unique key.
	GetByUniqueKey(kit *kit.Kit, bizID, appID uint32, name string) (*table.Kv, error)
}

var _ Kv = new(kvDao)

type kvDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

func (dao *kvDao) Create(kit *kit.Kit, kv *table.Kv) (uint32, error) {
	if kv == nil {
		return 0, fmt.Errorf("kv is nil")
	}

	if err := kv.ValidateCreate(); err != nil {
		return 0, err
	}

	// generate an commit id and update to commit.
	id, err := dao.idGen.One(kit, table.Name(kv.TableName()))
	if err != nil {
		return 0, err
	}
	kv.ID = id

	ad := dao.auditDao.DecoratorV2(kit, kv.Attachment.BizID).PrepareCreate(kv)

	createTx := func(tx *gen.Query) error {
		q := tx.Kv.WithContext(kit.Ctx)
		if err = q.Create(kv); err != nil {
			return err
		}
		if err = ad.Do(tx); err != nil {
			return err
		}

		return nil
	}
	if err = dao.genQ.Transaction(createTx); err != nil {
		return 0, err
	}

	return id, nil
}

func (dao *kvDao) Update(kit *kit.Kit, kv *table.Kv) error {
	if err := kv.ValidateUpdate(); err != nil {
		return err
	}

	// 更新操作, 获取当前记录做审计
	m := dao.genQ.Credential
	q := dao.genQ.Credential.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(kv.ID), m.BizID.Eq(kv.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, kv.Attachment.BizID).PrepareUpdate(kv, oldOne)

	// 多个使用事务处理
	updateTx := func(tx *gen.Query) error {
		q = tx.Credential.WithContext(kit.Ctx)
		if _, e := q.Where(m.BizID.Eq(kv.Attachment.BizID), m.ID.Eq(kv.ID)).
			Select(m.Memo, m.Enable).Updates(kv); e != nil {
			return e
		}

		if e := ad.Do(tx); e != nil {
			return e
		}
		return nil
	}
	if e := dao.genQ.Transaction(updateTx); e != nil {
		return e
	}

	return nil

}

func (dao *kvDao) List(kit *kit.Kit, opt *types.ListKvOption) ([]*table.Kv, int64, error) {

	m := dao.genQ.Kv
	q := dao.genQ.Kv.WithContext(kit.Ctx).Where(m.BizID.Eq(opt.BizID), m.AppID.Eq(opt.AppID)).Order(m.ID.Desc())

	if opt.Page.Start == 0 && opt.Page.Limit == 0 {
		result, err := q.Find()
		if err != nil {
			return nil, 0, err
		}

		return result, int64(len(result)), err

	}

	result, count, err := q.FindByPage(opt.Page.Offset(), opt.Page.LimitInt())
	if err != nil {
		return nil, 0, err
	}

	return result, count, err

}

func (dao *kvDao) Delete(kit *kit.Kit, kv *table.Kv) error {

	// 参数校验
	if err := kv.ValidateDelete(); err != nil {
		return err
	}

	// 删除操作, 获取当前记录做审计
	m := dao.genQ.Kv
	q := dao.genQ.Kv.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(kv.ID), m.BizID.Eq(kv.Attachment.BizID), m.AppID.Eq(kv.Attachment.AppID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, kv.Attachment.BizID).PrepareDelete(oldOne)

	// 多个使用事务处理
	deleteTx := func(tx *gen.Query) error {
		q = tx.Kv.WithContext(kit.Ctx)
		if _, err := q.Where(m.BizID.Eq(kv.Attachment.BizID), m.ID.Eq(kv.ID)).Delete(kv); err != nil {
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

// GetByUniqueKey get kv by unique key.
func (dao *kvDao) GetByUniqueKey(kit *kit.Kit, bizID, appID uint32, name string) (*table.Kv, error) {
	m := dao.genQ.Kv
	q := dao.genQ.Kv.WithContext(kit.Ctx)

	kv, err := q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appID), m.Name.Eq(name)).Take()
	if err != nil {
		return nil, fmt.Errorf("get kv failed, err: %v", err)
	}

	return kv, nil
}
