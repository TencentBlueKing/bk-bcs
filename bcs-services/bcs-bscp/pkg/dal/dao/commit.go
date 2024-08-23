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

	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// Commit supplies all the commit related operations.
type Commit interface {
	// Create one commit instance.
	Create(kit *kit.Kit, commit *table.Commit) (uint32, error)
	// CreateWithTx create one commit instance with transaction
	CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, commit *table.Commit) (uint32, error)
	// BatchCreateWithTx batch create commit instances with transaction.
	BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, commits []*table.Commit) error
	// BatchListLatestCommits batch list config itmes' latest commit.
	BatchListLatestCommits(kit *kit.Kit, bizID, appID uint32, ids []uint32) ([]*table.Commit, error)
	// GetLatestCommit get config item's latest commit.
	GetLatestCommit(kit *kit.Kit, bizID, appID, configItemID uint32) (*table.Commit, error)
	// ListAppLatestCommits list app config items' latest commit.
	ListAppLatestCommits(kit *kit.Kit, bizID, appID uint32) ([]*table.Commit, error)
	// BatchDeleteWithTx batch delete commit data instance with transaction.
	BatchDeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, commitIDs []uint32) error
	// ListCommitsByGtID get list data by greater than ID.
	ListCommitsByGtID(kit *kit.Kit, commitID, bizID, appID, configItemID uint32) ([]*table.Commit, error)
}

var _ Commit = new(commitDao)

type commitDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// ListCommitsByGtID get list data by greater than ID.
func (dao *commitDao) ListCommitsByGtID(kit *kit.Kit, commitID, bizID, appID, configItemID uint32) (
	[]*table.Commit, error) {

	m := dao.genQ.Commit

	return dao.genQ.Commit.WithContext(kit.Ctx).
		Where(m.ID.Gt(commitID), m.BizID.Eq(bizID), m.AppID.Eq(appID),
			m.ConfigItemID.Eq(configItemID)).Find()
}

// BatchDeleteWithTx batch delete commit data instance with transaction.
func (dao *commitDao) BatchDeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, commitIDs []uint32) error {

	m := tx.Query.Commit
	q := tx.Query.Commit.WithContext(kit.Ctx)

	_, err := q.Where(m.ID.In(commitIDs...)).Delete()
	if err != nil {
		return err
	}
	return nil
}

// Create one commit instance.
func (dao *commitDao) Create(kit *kit.Kit, commit *table.Commit) (uint32, error) {

	if commit == nil {
		return 0, errf.New(errf.InvalidParameter, "commit is nil")
	}

	if err := commit.ValidateCreate(kit); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	if err := dao.validateAttachmentResExist(kit, commit.Attachment); err != nil {
		return 0, err
	}

	// generate a commit id and update to commit.
	id, err := dao.idGen.One(kit, commit.TableName())
	if err != nil {
		return 0, err
	}

	commit.ID = id

	ad := dao.auditDao.DecoratorV2(kit, commit.Attachment.BizID).PrepareCreate(commit)

	createTx := func(tx *gen.Query) error {
		q := tx.Commit.WithContext(kit.Ctx)
		if err = q.Create(commit); err != nil {
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

// CreateWithTx create one commit instance with transaction
func (dao *commitDao) CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, commit *table.Commit) (uint32, error) {

	if commit == nil {
		return 0, errf.New(errf.InvalidParameter, "commit is nil")
	}

	if err := commit.ValidateCreate(kit); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	// generate an commit id and update to commit.
	id, err := dao.idGen.One(kit, table.CommitsTable)

	if err != nil {
		return 0, err
	}

	commit.ID = id
	if err := tx.Query.Commit.WithContext(kit.Ctx).Create(commit); err != nil {
		return 0, err
	}

	ad := dao.auditDao.DecoratorV2(kit, commit.Attachment.BizID).PrepareCreate(commit)
	if err := ad.Do(tx.Query); err != nil {
		return 0, fmt.Errorf("audit create commit failed, err: %v", err)
	}

	return id, nil
}

// BatchCreateWithTx batch create commit instances with transaction.
// NOTE: 1. this method won't audit, because it's batch operation.
// 2. this method won't validate attachment resource exist, because it's batch operation.
func (dao *commitDao) BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, commits []*table.Commit) error {
	if len(commits) == 0 {
		return nil
	}
	ids, err := dao.idGen.Batch(kit, table.CommitsTable, len(commits))
	if err != nil {
		return err
	}
	for i, commit := range commits {
		if err := commit.ValidateCreate(kit); err != nil {
			return err
		}
		commit.ID = ids[i]
	}
	return tx.Query.Commit.WithContext(kit.Ctx).Save(commits...)
}

// BatchListLatestCommits batch list config items' latest commit.
func (dao *commitDao) BatchListLatestCommits(kit *kit.Kit, bizID, appID uint32, ids []uint32) ([]*table.Commit, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	m := dao.genQ.Commit
	q := dao.genQ.Commit.WithContext(kit.Ctx)
	subQuery := q.Select(m.ID.Max().As("commit_id")).Where(
		m.BizID.Eq(bizID), m.AppID.Eq(appID), m.ConfigItemID.In(ids...)).Group(m.ConfigItemID)
	return q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appID), q.Columns(m.ID).In(subQuery)).Find()
}

// ListAppLatestCommits list app config items' latest commit.
func (dao *commitDao) ListAppLatestCommits(kit *kit.Kit, bizID, appID uint32) ([]*table.Commit, error) {
	m := dao.genQ.Commit
	q := dao.genQ.Commit.WithContext(kit.Ctx)
	subQuery := q.Select(m.ID.Max().As("commit_id")).Where(
		m.BizID.Eq(bizID), m.AppID.Eq(appID)).Group(m.ConfigItemID)
	return q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appID), q.Columns(m.ID).In(subQuery)).Find()
}

// GetLatestCommit get config item's latest commit.
func (dao *commitDao) GetLatestCommit(kit *kit.Kit, bizID, appID, configItemID uint32) (*table.Commit, error) {
	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is 0")
	}
	if appID == 0 {
		return nil, errf.New(errf.InvalidParameter, "app id is 0")
	}
	if configItemID == 0 {
		return nil, errf.New(errf.InvalidParameter, "config item id is 0")
	}
	m := dao.genQ.Commit
	return m.WithContext(kit.Ctx).
		Where(m.ConfigItemID.Eq(configItemID), m.AppID.Eq(appID), m.BizID.Eq(bizID)).
		Order(m.ID.Desc()).First()
}

// validateAttachmentResExist validate if attachment resource exists before creating commit.
func (dao *commitDao) validateAttachmentResExist(kit *kit.Kit, am *table.CommitAttachment) error {

	appQ := dao.genQ.App
	// validate if commit attached app exists.
	if _, err := appQ.WithContext(kit.Ctx).
		Where(appQ.ID.Eq(am.AppID), appQ.BizID.Eq(am.BizID)).Take(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("commit attached app %d not exist", am.AppID)
		}
		return fmt.Errorf("get commit attached app %d failed", am.AppID)
	}

	ciQ := dao.genQ.ConfigItem
	// validate if commit attached config item exists.
	if _, err := ciQ.WithContext(kit.Ctx).Where(
		ciQ.BizID.Eq(am.BizID), ciQ.AppID.Eq(am.AppID), ciQ.ID.Eq(am.ConfigItemID)).Take(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("commit attached config item %d not exist", am.ConfigItemID)
		}
		return fmt.Errorf("get commit attached config item %d failed", am.ConfigItemID)
	}

	return nil
}
