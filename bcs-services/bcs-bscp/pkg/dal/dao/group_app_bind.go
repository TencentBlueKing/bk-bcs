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
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// GroupAppBind supplies all the group related operations.
type GroupAppBind interface {
	// BatchCreateWithTx batch create group app with transaction.
	BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, items []*table.GroupAppBind) error
	// BatchDeleteByGroupIDWithTx batch delete group app by group id with transaction.
	BatchDeleteByGroupIDWithTx(kit *kit.Kit, tx *gen.QueryTx, groupID, bizID uint32) error
	// BatchDeleteByAppIDWithTx batch delete group app by app id with transaction.
	BatchDeleteByAppIDWithTx(kit *kit.Kit, tx *gen.QueryTx, appID, bizID uint32) error
	// BatchListByGroupIDs batch list group app by group ids.
	BatchListByGroupIDs(kit *kit.Kit, bizID uint32, groupIDs []uint32) ([]*table.GroupAppBind, error)
	// Get get GroupAppBind by group id and app id.
	Get(kit *kit.Kit, groupID, appID, bizID uint32) (*table.GroupAppBind, error)
}

var _ GroupAppBind = new(groupAppDao)

type groupAppDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
	lock     LockDao
}

// BatchCreateWithTx batch create group app with transaction.
func (dao *groupAppDao) BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, items []*table.GroupAppBind) error {
	if len(items) == 0 {
		return nil
	}

	// generate released config items id.
	ids, err := dao.idGen.Batch(kit, table.GroupAppBindTable, len(items))
	if err != nil {
		return err
	}

	for i, item := range items {
		// validate released config item field.
		if err := item.ValidateCreate(); err != nil {
			return err
		}
		item.ID = ids[i]
	}

	return tx.Query.GroupAppBind.WithContext(kit.Ctx).Save(items...)
}

// BatchDeleteByGroupIDWithTx batch delete group app by group id with transaction.
func (dao *groupAppDao) BatchDeleteByGroupIDWithTx(kit *kit.Kit, tx *gen.QueryTx, groupID, bizID uint32) error {

	if groupID == 0 {
		return errf.New(errf.InvalidParameter, "group id is 0")
	}

	if bizID == 0 {
		return errf.New(errf.InvalidParameter, "biz id is 0")
	}

	m := tx.Query.GroupAppBind
	if _, err := tx.Query.GroupAppBind.WithContext(kit.Ctx).Where(
		m.GroupID.Eq(groupID), m.BizID.Eq(bizID)).Delete(); err != nil {
		return err
	}

	return nil
}

// BatchDeleteByAppIDWithTx batch delete group app by app id with transaction.
func (dao *groupAppDao) BatchDeleteByAppIDWithTx(kit *kit.Kit, tx *gen.QueryTx, appID, bizID uint32) error {

	if appID == 0 {
		return errf.New(errf.InvalidParameter, "app id is 0")
	}

	if bizID == 0 {
		return errf.New(errf.InvalidParameter, "biz id is 0")
	}

	m := tx.GroupAppBind
	_, err := m.WithContext(kit.Ctx).Where(m.AppID.Eq(appID), m.BizID.Eq(bizID)).Delete()

	return err
}

// BatchListByGroupIDs batch list group app by group ids.
func (dao *groupAppDao) BatchListByGroupIDs(kit *kit.Kit,
	bizID uint32, groupIDs []uint32) ([]*table.GroupAppBind, error) {

	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is 0")
	}

	if len(groupIDs) == 0 {
		return nil, nil
	}

	m := dao.genQ.GroupAppBind
	return m.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID), m.GroupID.In(groupIDs...)).Find()

}

// Get get GroupAppBind by group id and app id.
func (dao *groupAppDao) Get(kit *kit.Kit, groupID, appID, bizID uint32) (*table.GroupAppBind, error) {
	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is 0")
	}
	if groupID == 0 {
		return nil, errf.New(errf.InvalidParameter, "group id is 0")
	}
	if appID == 0 {
		return nil, errf.New(errf.InvalidParameter, "app id is 0")
	}

	m := dao.genQ.GroupAppBind
	return m.WithContext(kit.Ctx).Where(m.GroupID.Eq(groupID), m.AppID.Eq(appID), m.BizID.Eq(bizID)).Take()
}
