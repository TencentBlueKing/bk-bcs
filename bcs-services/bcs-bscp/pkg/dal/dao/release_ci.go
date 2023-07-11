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
	"errors"
	"fmt"

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/types"
)

// ReleasedCI supplies all the released config item related operations.
type ReleasedCI interface {
	// BulkCreateWithTx bulk create released config items with tx.
	BulkCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, items []*table.ReleasedConfigItem) error
	// Get released config item by id and released id
	Get(kit *kit.Kit, id, bizID, releasedID uint32) (*table.ReleasedConfigItem, error)
	// GetReleasedLately released config item by app id and biz id
	GetReleasedLately(kit *kit.Kit, appId, bizID uint32, searchKey string) ([]*table.ReleasedConfigItem, error)
	// List released config items with options.
	List(kit *kit.Kit, opts *types.ListReleasedCIsOption) (*types.ListReleasedCIsDetails, error)
	// ListAll list all released config items in biz.
	ListAll(kit *kit.Kit, bizID uint32) ([]*table.ReleasedConfigItem, error)
	// ListAllByAppID list all released config items by appID.
	ListAllByAppID(kit *kit.Kit, appID, bizID uint32) ([]*table.ReleasedConfigItem, error)
	// ListAllByAppIDs batch list released config items by appIDs.
	ListAllByAppIDs(kit *kit.Kit, appIDs []uint32, bizID uint32) ([]*table.ReleasedConfigItem, error)
	// ListAllByReleaseIDs batch list released config items by releaseIDs.
	ListAllByReleaseIDs(kit *kit.Kit, releasedIDs []uint32, bizID uint32) ([]*table.ReleasedConfigItem, error)
}

var _ ReleasedCI = new(releasedCIDao)

type releasedCIDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// BulkCreateWithTx bulk create released config items.
func (dao *releasedCIDao) BulkCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, items []*table.ReleasedConfigItem) error {
	if len(items) == 0 {
		return errors.New("released config items is empty")
	}

	// validate released config item field.
	for _, item := range items {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	// generate released config items id.
	ids, err := dao.idGen.Batch(kit, table.ReleasedConfigItemTable, len(items))
	if err != nil {
		return err
	}

	start := 0
	for _, item := range items {
		item.ID = ids[start]
		start++
	}
	batchSize := 100

	q := tx.ReleasedConfigItem.WithContext(kit.Ctx)
	if err := q.CreateInBatches(items, batchSize); err != nil {
		return fmt.Errorf("insert events failed, err: %v", err)
	}

	ad := dao.auditDao.DecoratorV2(kit, items[0].Attachment.BizID).PrepareCreate(table.RciList(items))
	if err := ad.Do(tx.Query); err != nil {
		return err
	}

	return nil
}

// Get released config item by ID and config item id and release id.
func (dao *releasedCIDao) Get(kit *kit.Kit, configItemID, bizID, releaseID uint32) (*table.ReleasedConfigItem, error) {

	if configItemID == 0 {
		return nil, errf.New(errf.InvalidParameter, "config item id can not be 0")
	}

	if releaseID == 0 {
		return nil, errf.New(errf.InvalidParameter, "release id can not be 0")
	}

	m := dao.genQ.ReleasedConfigItem
	return m.WithContext(kit.Ctx).Where(
		m.ConfigItemID.Eq(configItemID), m.ReleaseID.Eq(releaseID), m.BizID.Eq(bizID)).Take()
}

// List released config items with options.
func (dao *releasedCIDao) List(kit *kit.Kit, opts *types.ListReleasedCIsOption) (
	*types.ListReleasedCIsDetails, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list released config items options null")
	}

	po := &types.PageOption{
		// allows list released ci without page
		EnableUnlimitedLimit: true,
		DisabledSort:         false,
	}
	if err := opts.Validate(po); err != nil {
		return nil, err
	}

	m := dao.genQ.ReleasedConfigItem

	query := m.WithContext(kit.Ctx).Where(m.ReleaseID.Eq(opts.ReleaseID), m.BizID.Eq(opts.BizID))
	if opts.SearchKey != "" {
		searchKey := "%" + opts.SearchKey + "%"
		query = query.Where(m.Name.Like(searchKey)).Or(m.Creator.Like(searchKey)).Or(m.Reviser.Like(searchKey))
	}

	var list []*table.ReleasedConfigItem
	var count int64
	var err error
	if opts.Page.Start == 0 && opts.Page.Limit == 0 {
		list, err = query.Find()
		if err != nil {
			return nil, err
		}
		count = int64(len(list))
	} else {
		list, count, err = query.FindByPage(opts.Page.Offset(), opts.Page.LimitInt())
		if err != nil {
			return nil, err
		}
	}
	return &types.ListReleasedCIsDetails{Count: uint32(count), Details: list}, nil
}

// ListAll list all released config items in biz.
func (dao *releasedCIDao) ListAll(kit *kit.Kit, bizID uint32) ([]*table.ReleasedConfigItem, error) {
	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "biz_id can not be 0")
	}

	m := dao.genQ.ReleasedConfigItem
	return m.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID)).Find()
}

// ListAll list all released config items in biz.
func (dao *releasedCIDao) ListAllByAppID(kit *kit.Kit, appID, bizID uint32) ([]*table.ReleasedConfigItem, error) {
	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "biz_id can not be 0")
	}

	m := dao.genQ.ReleasedConfigItem
	return m.WithContext(kit.Ctx).Where(m.AppID.Eq(appID), m.BizID.Eq(bizID)).Find()
}

// ListAllByAppIDs list all released config items by appIDs.
func (dao *releasedCIDao) ListAllByAppIDs(kit *kit.Kit,
	appIDs []uint32, bizID uint32) ([]*table.ReleasedConfigItem, error) {
	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "biz_id can not be 0")
	}

	m := dao.genQ.ReleasedConfigItem
	return m.WithContext(kit.Ctx).Where(m.AppID.In(appIDs...), m.BizID.Eq(bizID)).Find()
}

// ListAllByReleaseIDs list all released config items by releaseIDs.
func (dao *releasedCIDao) ListAllByReleaseIDs(kit *kit.Kit,
	releaseIDs []uint32, bizID uint32) ([]*table.ReleasedConfigItem, error) {
	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "biz_id can not be 0")
	}

	m := dao.genQ.ReleasedConfigItem
	return m.WithContext(kit.Ctx).Where(m.ReleaseID.In(releaseIDs...), m.BizID.Eq(bizID)).Find()
}

// GetReleasedLately
func (dao *releasedCIDao) GetReleasedLately(kit *kit.Kit, appId, bizID uint32, searchKey string) (
	[]*table.ReleasedConfigItem, error) {
	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "biz_id can not be 0")
	}

	m := dao.genQ.ReleasedConfigItem
	q := dao.genQ.ReleasedConfigItem.WithContext(kit.Ctx)
	query := q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appId))
	if searchKey != "" {
		param := "%" + searchKey + "%"
		query = q.Where(q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appId)),
			q.Where(m.Name.Like(param)).Or(m.Creator.Like(param)).Or(m.Reviser.Like(param)))
	}
	subQuery := q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appId)).Order(m.ReleaseID.Desc()).Limit(1).Select(m.ReleaseID)
	return query.Where(q.Columns(m.ReleaseID).Eq(subQuery)).Find()
}
