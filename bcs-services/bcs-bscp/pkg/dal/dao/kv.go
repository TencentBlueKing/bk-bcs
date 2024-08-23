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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/utils"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// Kv supplies all the kv related operations.
type Kv interface {
	// Create one kv instance
	Create(kit *kit.Kit, kv *table.Kv) (uint32, error)
	// Update one kv's info
	Update(kit *kit.Kit, kv *table.Kv) error
	// List kv with options.
	List(kit *kit.Kit, opt *types.ListKvOption) ([]*table.Kv, int64, error)
	// ListAllKvByKey lists all key-value pairs based on keys
	ListAllKvByKey(kit *kit.Kit, appID uint32, bizID uint32, keys []string, kvState []string) ([]*table.Kv, error)
	// Delete ..
	Delete(kit *kit.Kit, kv *table.Kv) error
	// DeleteWithTx delete kv instance with transaction.
	DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, kv *table.Kv) error
	// UpdateWithTx update kv instance with transaction.
	UpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, kv *table.Kv) error
	// GetByKvState get kv by KvState.
	GetByKvState(kit *kit.Kit, bizID, appID uint32, key string, kvState []string) (*table.Kv, error)
	// GetByID get kv by id.
	GetByID(kit *kit.Kit, bizID, appID, id uint32) (*table.Kv, error)
	// BatchCreateWithTx batch create content instances with transaction.
	BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, kvs []*table.Kv) error
	// BatchUpdateWithTx batch create content instances with transaction.
	BatchUpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, kvs []*table.Kv) error
	// BatchDeleteWithTx batch delete content instances with transaction.
	BatchDeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID, appID uint32, ids []uint32) error
	// ListAllByAppID list all Kv by appID
	ListAllByAppID(kit *kit.Kit, appID uint32, bizID uint32, kvState []string) ([]*table.Kv, error)
	// GetCount bizID config count
	GetCount(kit *kit.Kit, bizID uint32, appId []uint32) ([]*table.ListConfigItemCounts, error)
	// UpdateSelectedKVStates updates the states of selected kv pairs using a transaction
	UpdateSelectedKVStates(kit *kit.Kit, tx *gen.QueryTx, bizID, appID uint32, targetKVStates []string,
		newKVStates table.KvState) error
	// DeleteByStateWithTx deletes kv pairs with a specific state using a transaction
	DeleteByStateWithTx(kit *kit.Kit, tx *gen.QueryTx, kv *table.Kv) error
	// FetchIDsExcluding 获取指定ID后排除的ID
	FetchIDsExcluding(kit *kit.Kit, bizID uint32, appID uint32, ids []uint32) ([]uint32, error)
	// CountNumberUnDeleted 统计未删除的数量
	CountNumberUnDeleted(kit *kit.Kit, bizID uint32, opt *types.ListKvOption) (int64, error)
}

var _ Kv = new(kvDao)

type kvDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// CountNumberUnDeleted 统计未删除的数量
func (dao *kvDao) CountNumberUnDeleted(kit *kit.Kit, bizID uint32, opt *types.ListKvOption) (int64, error) {
	m := dao.genQ.Kv
	q := dao.genQ.Kv.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID), m.AppID.Eq(opt.AppID))
	if opt.SearchKey != "" {
		searchKey := "(?i)" + opt.SearchKey
		q = q.Where(q.Where(q.Or(m.Key.Regexp(searchKey)).Or(m.Creator.Regexp(searchKey)).Or(
			m.Reviser.Regexp(searchKey))))
	}

	if len(opt.Status) != 0 {
		q = q.Where(m.KvState.In(opt.Status...))
	}

	return q.Where(m.BizID.Eq(opt.BizID)).Where(m.KvState.Neq(table.KvStateDelete.String())).Count()
}

// FetchIDsExcluding 获取指定ID后排除的ID
func (dao *kvDao) FetchIDsExcluding(kit *kit.Kit, bizID uint32, appID uint32, ids []uint32) ([]uint32, error) {
	m := dao.genQ.Kv
	q := dao.genQ.Kv.WithContext(kit.Ctx)

	var result []uint32
	if err := q.Select(m.ID).
		Where(m.BizID.Eq(bizID), m.AppID.Eq(appID), m.ID.NotIn(ids...)).
		Pluck(m.ID, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// BatchDeleteWithTx batch delete content instances with transaction.
func (dao *kvDao) BatchDeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID uint32,
	appID uint32, ids []uint32) error {
	m := dao.genQ.Kv
	_, err := dao.genQ.Kv.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID), m.AppID.Eq(appID), m.ID.In(ids...)).Delete()
	return err
}

// UpdateWithTx update kv instance with transaction.
func (dao *kvDao) UpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, kv *table.Kv) error {
	// 参数校验
	if err := kv.ValidateUpdate(); err != nil {
		return err
	}

	// 编辑操作, 获取当前记录做审计
	m := tx.Kv
	q := tx.Kv.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(kv.ID), m.BizID.Eq(kv.Attachment.BizID), m.AppID.Eq(kv.Attachment.AppID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, kv.Attachment.BizID).PrepareUpdate(kv, oldOne)

	_, err = q.Where(m.BizID.Eq(kv.Attachment.BizID), m.ID.Eq(kv.ID)).Select(m.Version, m.UpdatedAt,
		m.Reviser, m.KvState, m.Signature, m.Md5, m.ByteSize).Updates(kv)
	if err != nil {
		return err
	}

	return ad.Do(tx.Query)
}

func (dao *kvDao) Create(kit *kit.Kit, kv *table.Kv) (uint32, error) {
	if kv == nil {
		return 0, fmt.Errorf("kv is nil")
	}

	if err := kv.ValidateCreate(kit); err != nil {
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

// Update one kv's info
func (dao *kvDao) Update(kit *kit.Kit, kv *table.Kv) error {

	if err := kv.ValidateUpdate(); err != nil {
		return err
	}

	// 更新操作, 获取当前记录做审计
	m := dao.genQ.Kv
	q := dao.genQ.Kv.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(kv.ID), m.BizID.Eq(kv.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, kv.Attachment.BizID).PrepareUpdate(kv, oldOne)

	// 多个使用事务处理
	updateTx := func(tx *gen.Query) error {
		q = tx.Kv.WithContext(kit.Ctx)
		if _, e := q.Where(m.BizID.Eq(kv.Attachment.BizID), m.ID.Eq(kv.ID)).Select(m.Version, m.UpdatedAt,
			m.Reviser, m.KvState, m.Signature, m.Md5, m.ByteSize, m.Memo).Updates(kv); e != nil {
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

// List kv with options.
func (dao *kvDao) List(kit *kit.Kit, opt *types.ListKvOption) ([]*table.Kv, int64, error) {

	m := dao.genQ.Kv
	q := dao.genQ.Kv.WithContext(kit.Ctx)

	orderCol, ok := m.GetFieldByName(opt.Page.Sort)
	if !ok {
		return nil, 0, errors.New("user doesn't contains orderColStr")
	}

	orderStr := "CASE WHEN kv_state = 'ADD' THEN 1 WHEN kv_state = 'REVISE' THEN 2 " +
		"WHEN kv_state = 'DELETE' THEN 3 WHEN kv_state = 'UNCHANGE' THEN 4 END,`key` asc"

	if opt.Page.Order == types.Descending {
		q = q.Order(orderCol.Desc())
	} else if len(opt.TopIDs) != 0 {
		q = q.Order(utils.NewCustomExpr("CASE WHEN id IN (?) THEN 0 ELSE 1 END, "+orderStr, []interface{}{opt.TopIDs}))
	} else {
		q = q.Order(utils.NewCustomExpr(orderStr, nil))
	}

	if opt.SearchKey != "" {
		searchKey := "(?i)" + opt.SearchKey
		q = q.Where(q.Where(q.Or(m.Key.Regexp(searchKey)).Or(m.Creator.Regexp(searchKey)).Or(
			m.Reviser.Regexp(searchKey))))
	}

	if len(opt.Status) != 0 {
		q = q.Where(m.KvState.In(opt.Status...))
	}
	q = q.Where(m.BizID.Eq(opt.BizID)).Where(m.AppID.Eq(opt.AppID))

	if len(opt.KvType) > 0 {
		q = q.Where(m.KvType.In(opt.KvType...))
	}
	if len(opt.Key) > 0 {
		q = q.Where(m.Key.In(opt.Key...))
	}

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

// Delete ..
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
		if _, e := q.Where(m.BizID.Eq(kv.Attachment.BizID), m.ID.Eq(kv.ID)).Delete(kv); e != nil {
			return e
		}

		if e := ad.Do(tx); e != nil {
			return e
		}
		return nil
	}
	if e := dao.genQ.Transaction(deleteTx); e != nil {
		return e
	}

	return nil

}

// DeleteWithTx delete kv instance with transaction.
func (dao *kvDao) DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, kv *table.Kv) error {
	// 参数校验
	if err := kv.ValidateDelete(); err != nil {
		return err
	}

	// 删除操作, 获取当前记录做审计
	m := tx.Kv
	q := tx.Kv.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(kv.ID), m.BizID.Eq(kv.Attachment.BizID), m.AppID.Eq(kv.Attachment.AppID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, kv.Attachment.BizID).PrepareDelete(oldOne)

	_, err = q.Where(m.BizID.Eq(kv.Attachment.BizID), m.ID.Eq(kv.ID)).Delete(kv)
	if err != nil {
		return err
	}

	if e := ad.Do(tx.Query); e != nil {
		return e
	}

	return nil

}

// DeleteByStateWithTx deletes kv pairs with a specific state using a transaction
func (dao *kvDao) DeleteByStateWithTx(kit *kit.Kit, tx *gen.QueryTx, kv *table.Kv) error {
	// 参数校验

	if kv.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if kv.Attachment.AppID <= 0 {
		return errors.New("app id should be set")
	}

	if kv.KvState == "" {
		return errors.New("KvState should be set")
	}

	// 删除操作, 获取当前记录做审计
	m := tx.Kv
	q := tx.Kv.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.BizID.Eq(kv.Attachment.BizID), m.AppID.Eq(kv.Attachment.AppID),
		m.KvState.Eq(string(kv.KvState))).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, kv.Attachment.BizID).PrepareDelete(oldOne)

	_, err = q.Where(m.BizID.Eq(kv.Attachment.BizID), m.AppID.Eq(kv.Attachment.AppID),
		m.KvState.Eq(string(kv.KvState))).Delete(kv)
	if err != nil {
		return err
	}

	if e := ad.Do(tx.Query); e != nil {
		return e
	}

	return nil

}

// GetByKvState get kv by KvState.
func (dao *kvDao) GetByKvState(kit *kit.Kit, bizID, appID uint32, key string, kvState []string) (*table.Kv, error) {
	m := dao.genQ.Kv
	q := dao.genQ.Kv.WithContext(kit.Ctx)

	kv, err := q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appID), m.Key.Eq(key), m.KvState.In(kvState...)).
		Take()
	if err != nil {
		return nil, err
	}

	return kv, nil
}

// GetByID get kv by id.
func (dao *kvDao) GetByID(kit *kit.Kit, bizID, appID, id uint32) (*table.Kv, error) {
	m := dao.genQ.Kv
	q := dao.genQ.Kv.WithContext(kit.Ctx)

	kv, err := q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appID), m.ID.Eq(id)).Take()
	if err != nil {
		return nil, err
	}

	return kv, nil
}

// ListAllKvByKey lists all key-value pairs based on keys
func (dao *kvDao) ListAllKvByKey(kit *kit.Kit, appID uint32, bizID uint32, keys []string,
	kvState []string) ([]*table.Kv, error) {

	if appID == 0 {
		return nil, fmt.Errorf("appID can not be 0")
	}
	if bizID == 0 {
		return nil, fmt.Errorf("bizID can not be 0")
	}

	m := dao.genQ.Kv

	q := dao.genQ.Kv.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID), m.AppID.Eq(appID), m.Key.In(keys...))
	if len(kvState) > 0 {
		q.Where(m.KvState.In(kvState...))
	}

	return q.Find()
}

// BatchCreateWithTx batch create content instances with transaction.
func (dao *kvDao) BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, kvs []*table.Kv) error {

	// generate an kv id and update to kv.
	if len(kvs) == 0 {
		return nil
	}
	ids, err := dao.idGen.Batch(kit, table.KvTable, len(kvs))
	if err != nil {
		return err
	}
	for i, kv := range kvs {
		if e := kv.ValidateCreate(kit); e != nil {
			return e
		}
		kv.ID = ids[i]
	}
	if e := tx.Kv.WithContext(kit.Ctx).Save(kvs...); e != nil {
		return e
	}
	return nil

}

// BatchUpdateWithTx batch create content instances with transaction.
func (dao *kvDao) BatchUpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, kvs []*table.Kv) error {
	if len(kvs) == 0 {
		return nil
	}
	if err := tx.Kv.WithContext(kit.Ctx).Save(kvs...); err != nil {
		return err
	}
	return nil
}

// UpdateSelectedKVStates 批量更新kv状态
func (dao *kvDao) UpdateSelectedKVStates(kit *kit.Kit, tx *gen.QueryTx, bizID, appID uint32,
	targetKVStates []string, newKVStates table.KvState) error {

	if bizID <= 0 {
		return errors.New("biz id should be set")
	}

	if appID <= 0 {
		return errors.New("app id should be set")
	}
	if len(targetKVStates) == 0 {
		return nil
	}
	if newKVStates == "" {
		return errors.New("newKVStates cannot be empty")
	}

	m := tx.Kv

	if _, err := tx.Kv.WithContext(kit.Ctx).
		Where(m.BizID.Eq(bizID), m.AppID.Eq(appID), m.KvState.In(targetKVStates...)).
		Select(m.KvState).
		Omit(m.UpdatedAt).
		Update(m.KvState, newKVStates); err != nil {
		return err
	}

	return nil
}

// ListAllByAppID list all Kv by appID
func (dao *kvDao) ListAllByAppID(kit *kit.Kit, appID uint32, bizID uint32, kvState []string) ([]*table.Kv, error) {
	if appID == 0 {
		return nil, errf.New(errf.InvalidParameter, i18n.T(kit, "appID can not be 0"))
	}
	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "bizID can not be 0")
	}
	m := dao.genQ.Kv
	return dao.genQ.Kv.WithContext(kit.Ctx).Where(m.AppID.Eq(appID), m.BizID.Eq(bizID), m.KvState.In(kvState...)).Find()
}

// GetCount get bizID kv count
func (dao *kvDao) GetCount(kit *kit.Kit, bizID uint32, appId []uint32) ([]*table.ListConfigItemCounts, error) {

	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "config item biz id can not be 0")
	}

	configItem := make([]*table.ListConfigItemCounts, 0)

	kvState := []string{
		string(table.KvStateAdd),
		string(table.KvStateRevise),
		string(table.KvStateUnchange),
	}
	m := dao.genQ.Kv
	q := dao.genQ.Kv.WithContext(kit.Ctx)
	if err := q.Select(m.AppID, m.ID.Count().As("count"), m.UpdatedAt.Max().As("updated_at")).
		Where(m.BizID.Eq(bizID), m.AppID.In(appId...),
			m.KvState.In(kvState...)).Group(m.AppID).Scan(&configItem); err != nil {
		return nil, err
	}

	return configItem, nil
}
