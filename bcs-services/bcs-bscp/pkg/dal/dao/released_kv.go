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
	"fmt"

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/types"
)

// ReleasedKv supplies all the released kv related operations.
type ReleasedKv interface {
	// BulkCreateWithTx bulk create released kv with tx.
	BulkCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, kvs []*table.ReleasedKv) error
	// Get released kv.
	Get(kit *kit.Kit, bizID, appID, releasedID uint32, key string) (*table.ReleasedKv, error)
	// List released kv with options.
	List(kit *kit.Kit, opt *types.ListRKvOption) ([]*table.ReleasedKv, int64, error)
	// ListAllByReleaseIDs batch list released kvs by releaseIDs.
	ListAllByReleaseIDs(kit *kit.Kit, releasedIDs []uint32, bizID uint32) ([]*table.ReleasedKv, error)
	// GetReleasedLately get released kv lately
	GetReleasedLately(kit *kit.Kit, bizID, appID uint32) ([]*table.ReleasedKv, error)
	// GetReleasedLatelyByKey get released kv lately by key
	GetReleasedLatelyByKey(kit *kit.Kit, bizID, appID uint32, key string) (*table.ReleasedKv, error)
}

var _ ReleasedKv = new(releasedKvDao)

type releasedKvDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// BulkCreateWithTx bulk create released kv with tx.
func (dao *releasedKvDao) BulkCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, kvs []*table.ReleasedKv) error {

	if len(kvs) == 0 {
		return nil
	}

	// validate released kv field.
	for _, kv := range kvs {
		if err := kv.ValidateCreate(); err != nil {
			return err
		}
	}

	// generate released config items id.
	ids, err := dao.idGen.Batch(kit, table.ReleasedKvTable, len(kvs))
	if err != nil {
		return err
	}

	start := 0
	for _, kv := range kvs {
		kv.ID = ids[start]
		start++
	}
	batchSize := 100

	q := tx.ReleasedKv.WithContext(kit.Ctx)
	if err := q.CreateInBatches(kvs, batchSize); err != nil {
		return fmt.Errorf("create released kv in batch failed, err: %v", err)
	}

	ad := dao.auditDao.DecoratorV2(kit, kvs[0].Attachment.BizID).PrepareCreate(table.RkvList(kvs))
	if err := ad.Do(tx.Query); err != nil {
		return err
	}

	return nil

}

// Get released kv.
func (dao *releasedKvDao) Get(kit *kit.Kit, bizID, appID, releasedID uint32, key string) (*table.ReleasedKv, error) {
	m := dao.genQ.ReleasedKv
	return m.WithContext(kit.Ctx).Where(
		m.BizID.Eq(bizID), m.AppID.Eq(appID), m.ReleaseID.Eq(releasedID), m.Key.Eq(key)).Take()
}

// List released kv with options.
func (dao *releasedKvDao) List(kit *kit.Kit, opt *types.ListRKvOption) ([]*table.ReleasedKv, int64, error) {

	m := dao.genQ.ReleasedKv
	q := dao.genQ.ReleasedKv.WithContext(kit.Ctx).Where(m.BizID.Eq(opt.BizID), m.AppID.Eq(opt.AppID),
		m.ReleaseID.Eq(opt.ReleaseID)).Or(m.Key.Eq(opt.Key)).
		Order(m.Key)

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

// ListAllByReleaseIDs batch list released kvs by releaseIDs.
func (dao *releasedKvDao) ListAllByReleaseIDs(kit *kit.Kit, releasedIDs []uint32, bizID uint32) ([]*table.ReleasedKv,
	error) {
	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "biz_id can not be 0")
	}
	m := dao.genQ.ReleasedKv
	return m.WithContext(kit.Ctx).Where(m.ReleaseID.In(releasedIDs...), m.BizID.Eq(bizID)).Find()
}

// GetReleasedLately get released kv lately
func (dao *releasedKvDao) GetReleasedLately(kit *kit.Kit, bizID, appID uint32) ([]*table.ReleasedKv, error) {

	m := dao.genQ.ReleasedKv
	q := dao.genQ.ReleasedKv.WithContext(kit.Ctx)

	query := q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appID))
	subQuery := q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appID)).Order(m.ReleaseID.Desc()).Limit(1).Select(m.ReleaseID)
	return query.Where(q.Columns(m.ReleaseID).Eq(subQuery)).Find()

}

// GetReleasedLatelyByKey get released kv lately by key
func (dao *releasedKvDao) GetReleasedLatelyByKey(kit *kit.Kit, bizID, appID uint32, key string) (*table.ReleasedKv, error) {
	m := dao.genQ.ReleasedKv
	return m.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID), m.AppID.Eq(appID), m.Key.Eq(key)).Take()
}
