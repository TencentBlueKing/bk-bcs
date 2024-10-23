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

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/sharding"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// Release supplies all the release related operations.
type Release interface {
	// CreateWithTx create one release instance with tx.
	CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, release *table.Release) (uint32, error)
	// List releases with options.
	List(kit *kit.Kit, opts *types.ListReleasesOption) (*types.ListReleaseDetails, error)
	// ListAllByIDs list all releases by releaseIDs.
	ListAllByIDs(kit *kit.Kit, ids []uint32, bizID uint32) ([]*table.Release, error)
	// GetByName get release by name
	GetByName(kit *kit.Kit, bizID uint32, appID uint32, name string) (*table.Release, error)
	// Get get release by id
	Get(kit *kit.Kit, bizID, appID, releaseID uint32) (*table.Release, error)
	// UpdateDeprecated update release deprecated status.
	UpdateDeprecated(kit *kit.Kit, bizID, appID, releaseID uint32, deprecated bool) error
	// DeleteWithTx delete release with tx.
	DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID, appID, releaseID uint32) error
	// GetReleaseLately get release lately info
	GetReleaseLately(kit *kit.Kit, bizID uint32, appID uint32) (*table.Release, error)
	// ListReleaseStrategies list release strategie the latest three pieces of data published
	ListReleaseStrategies(kit *kit.Kit, bizID uint32, appID uint32) ([]*types.ListReleasesStrategies, error)
}

var _ Release = new(releaseDao)

type releaseDao struct {
	genQ     *gen.Query
	sd       *sharding.Sharding
	idGen    IDGenInterface
	auditDao AuditDao
}

// GetReleaseLately get release lately info
func (dao *releaseDao) GetReleaseLately(kit *kit.Kit, bizID uint32, appID uint32) (*table.Release, error) {
	m := dao.genQ.Release
	return m.WithContext(kit.Ctx).Where(m.AppID.Eq(appID), m.BizID.Eq(bizID)).Order(m.ID.Desc()).Take()
}

// CreateWithTx create one release instance with tx.
func (dao *releaseDao) CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.Release) (uint32, error) {
	if g == nil {
		return 0, errors.New("release is nil")
	}

	if err := g.ValidateCreate(kit); err != nil {
		return 0, err
	}

	if err := dao.validateAttachmentResExist(kit, g.Attachment); err != nil {
		return 0, err
	}

	// generate an release id and update to release.
	id, err := dao.idGen.One(kit, table.ReleaseTable)
	if err != nil {
		return 0, err
	}
	g.ID = id

	q := tx.Release.WithContext(kit.Ctx)
	if err := q.Create(g); err != nil {
		return 0, err
	}

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)
	if err := ad.Do(tx.Query); err != nil {
		return 0, err
	}

	return g.ID, nil
}

// GetByName 通过名称获取, 可以做唯一性校验
func (dao *releaseDao) Get(kit *kit.Kit, bizID uint32, appID, releaseID uint32) (*table.Release, error) {
	m := dao.genQ.Release
	return m.WithContext(kit.Ctx).Where(m.ID.Eq(releaseID), m.AppID.Eq(appID), m.BizID.Eq(bizID)).Take()
}

// GetByName 通过名称获取, 可以做唯一性校验
func (dao *releaseDao) GetByName(kit *kit.Kit, bizID uint32, appID uint32, name string) (*table.Release, error) {
	m := dao.genQ.Release
	return m.WithContext(kit.Ctx).Where(m.Name.Eq(name), m.AppID.Eq(appID), m.BizID.Eq(bizID)).Take()
}

// List releases with options.
func (dao *releaseDao) List(kit *kit.Kit, opts *types.ListReleasesOption) (*types.ListReleaseDetails, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list releases options null")
	}

	po := &types.PageOption{
		EnableUnlimitedLimit: true,
		DisabledSort:         false,
	}

	if err := opts.Validate(po); err != nil {
		return nil, err
	}

	m := dao.genQ.Release
	q := m.WithContext(kit.Ctx)
	if opts.SearchKey == "" {
		q = q.Where(m.BizID.Eq(opts.BizID), m.AppID.Eq(opts.AppID), m.Deprecated.Is(opts.Deprecated))
	} else {
		searchKey := "(?i)" + opts.SearchKey
		q = q.Where(m.BizID.Eq(opts.BizID), m.AppID.Eq(opts.AppID), m.Deprecated.Is(opts.Deprecated)).Where(
			q.Where(m.Name.Regexp(searchKey)).Or(m.Memo.Regexp(searchKey)).Or(m.Creator.Regexp(searchKey)))
	}
	q = q.Order(m.ID.Desc())

	var list []*table.Release
	var count int64
	var err error
	if opts.Page.Start == 0 && opts.Page.Limit == 0 {
		list, err = q.Find()
		if err != nil {
			return nil, err
		}
		count = int64(len(list))
	} else {
		list, count, err = q.FindByPage(opts.Page.Offset(), opts.Page.LimitInt())
		if err != nil {
			return nil, err
		}
	}
	return &types.ListReleaseDetails{Count: uint32(count), Details: list}, nil

}

// ListAllByIDs list all releases by releaseIDs.
func (dao *releaseDao) ListAllByIDs(kit *kit.Kit, ids []uint32, bizID uint32) ([]*table.Release, error) {

	if len(ids) == 0 {
		return nil, nil
	}

	m := dao.genQ.Release
	return m.WithContext(kit.Ctx).Where(m.ID.In(ids...), m.BizID.Eq(bizID)).Find()
}

// validateAttachmentResExist validate if attachment resource exists before creating release.
func (dao *releaseDao) validateAttachmentResExist(kit *kit.Kit, am *table.ReleaseAttachment) error {
	m := dao.genQ.App
	// validate if release attached app exists.
	if _, err := m.WithContext(kit.Ctx).Where(m.ID.Eq(am.AppID), m.BizID.Eq(am.BizID)).Take(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("release attached app %d not exist", am.AppID)
		}
		return fmt.Errorf("get release attached app %d failed", am.AppID)
	}
	return nil
}

func (dao *releaseDao) UpdateDeprecated(kit *kit.Kit, bizID, appID, releaseID uint32, deprecated bool) error {
	m := dao.genQ.Release
	_, err := m.WithContext(kit.Ctx).
		Where(m.ID.Eq(releaseID), m.AppID.Eq(appID), m.BizID.Eq(bizID)).
		Update(m.Deprecated, deprecated)
	return err
}

// DeleteWithTx delete release with tx.
func (dao *releaseDao) DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID, appID, releaseID uint32) error {
	m := tx.Release
	_, err := m.WithContext(kit.Ctx).Where(m.ID.Eq(releaseID), m.AppID.Eq(appID), m.BizID.Eq(bizID)).Delete()
	return err
}

// ListReleaseStrategies list release strategie the latest three pieces of data published
func (dao *releaseDao) ListReleaseStrategies(
	kit *kit.Kit, bizID uint32, appID uint32) ([]*types.ListReleasesStrategies, error) {
	var publishRecords []*types.ListReleasesStrategies
	r := dao.genQ.Release
	s := dao.genQ.Strategy
	err := r.WithContext(kit.Ctx).Select(s.PublishTime, r.Name, s.Scope, r.Creator, r.FullyReleased).
		Join(s, s.ReleaseID.EqCol(r.ID), s.AppID.EqCol(r.AppID), s.BizID.EqCol(r.BizID)).
		Where(s.PublishStatus.Eq(string(table.AlreadyPublish)), s.AppID.Eq(appID), s.BizID.Eq(bizID)).
		Order(s.PublishTime.Desc()).Limit(3).Offset(0).
		Scan(&publishRecords)

	if err != nil {
		return publishRecords, err
	}
	return publishRecords, nil
}
