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

	"github.com/pkg/errors"

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/orm"
	"bscp.io/pkg/dal/sharding"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/types"
)

// Release supplies all the release related operations.
type Release interface {
	// CreateWithTx one release instance with tx.
	CreateWithTx(kit *kit.Kit, tx *sharding.Tx, release *table.Release) (uint32, error)
	// List releases with options.
	List(kit *kit.Kit, opts *types.ListReleasesOption) (*types.ListReleaseDetails, error)
	// GetByName ..
	GetByName(kit *kit.Kit, bizID uint32, appID uint32, name string) (*table.Release, error)
}

var _ Release = new(releaseDao)

type releaseDao struct {
	orm      orm.Interface
	genQ     *gen.Query
	sd       *sharding.Sharding
	idGen    IDGenInterface
	auditDao AuditDao
}

// CreateWithTx one release instance with tx.
func (dao *releaseDao) CreateWithTx(kit *kit.Kit, tx *sharding.Tx, release *table.Release) (uint32, error) {
	if release == nil {
		return 0, errf.New(errf.InvalidParameter, "release is nil")
	}

	if err := release.ValidateCreate(); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	if err := dao.validateAttachmentResExist(kit, release.Attachment); err != nil {
		return 0, err
	}

	// generate an release id and update to release.
	id, err := dao.idGen.One(kit, table.ReleaseTable)
	if err != nil {
		return 0, err
	}
	release.ID = id

	ad := dao.auditDao.DecoratorV2(kit, release.Attachment.BizID).PrepareCreate(release)
	createTx := func(tx *gen.Query) error {
		q := tx.Release.WithContext(kit.Ctx)
		if err := q.Create(release); err != nil {
			return err
		}

		if err := ad.Do(tx); err != nil {
			return err
		}

		return nil
	}
	if err := dao.genQ.Transaction(createTx); err != nil {
		return 0, nil
	}

	return release.ID, nil

}

// GetByName 通过名称获取, 可以做唯一性校验
func (dao *releaseDao) GetByName(kit *kit.Kit, bizID uint32, appID uint32, name string) (*table.Release, error) {
	one := new(table.Release)
	sql := table.SelectSQL(table.ReleaseColumns, one, "biz_id = ? AND app_id = ? AND name = ?")
	err := dao.orm.Do(dao.sd.Admin().DB()).Get(kit.Ctx, one, sql, bizID, appID, name)
	if err != nil {
		return nil, errors.Wrapf(err, "get release name")
	}

	return one, nil
}

// List releases with options.
func (dao *releaseDao) List(kit *kit.Kit, opts *types.ListReleasesOption) (
	*types.ListReleaseDetails, error) {

	m := dao.genQ.Release
	q := dao.genQ.Release.WithContext(kit.Ctx)

	result, count, err := q.Where(
		m.BizID.Eq(opts.BizID),
		m.AppID.Eq(opts.AppID),
		m.Deprecated.Is(opts.Deprecated)).
		FindByPage(opts.Page.Offset(), opts.Page.LimitInt())
	if err != nil {
		return nil, err
	}

	ListReleaseDetails := &types.ListReleaseDetails{
		Count:   uint32(count),
		Details: result,
	}

	return ListReleaseDetails, nil
}

// validateAttachmentResExist validate if attachment resource exists before creating release.
func (dao *releaseDao) validateAttachmentResExist(kit *kit.Kit, am *table.ReleaseAttachment) error {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, " WHERE id = ", strconv.Itoa(int(am.AppID)), " AND biz_id = ", strconv.Itoa(int(am.BizID)))
	sql := filter.SqlJoint(sqlSentence)
	exist, err := isResExist(kit, dao.orm, dao.sd.ShardingOne(am.BizID), table.AppTable, sql)
	if err != nil {
		return err
	}

	if !exist {
		return errf.New(errf.RelatedResNotExist, fmt.Sprintf("release attached app %d is not exist", am.AppID))
	}

	return nil
}
