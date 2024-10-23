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
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// ItsmConfig supplies all the itsm config related operations.
type ItsmConfig interface {
	// GetConfig Get itsm config.
	GetConfig(kit *kit.Kit, key string) (*table.ItsmConfig, error)
	// SetConfig Set itsm config.
	SetConfig(kit *kit.Kit, itsmConfig *table.ItsmConfig) error
	// UpdateConfig update itsm config.
	UpdateConfig(kit *kit.Kit, itsmConfig *table.ItsmConfig) error
}

var _ ItsmConfig = new(itsmConfigDao)

type itsmConfigDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao // nolint
}

// GetConfig Get itsm config.
func (dao *itsmConfigDao) GetConfig(kit *kit.Kit, key string) (*table.ItsmConfig, error) {
	m := dao.genQ.ItsmConfig
	return m.WithContext(kit.Ctx).Where(
		m.Key.Eq(key)).Take()
}

// SetConfig Set itsm config.
func (dao *itsmConfigDao) SetConfig(kit *kit.Kit, itsmConfig *table.ItsmConfig) error {

	// generate an content id and update to content.
	id, err := dao.idGen.One(kit, table.ItsmConfigTable)
	if err != nil {
		return err
	}
	itsmConfig.ID = id

	return dao.genQ.ItsmConfig.WithContext(kit.Ctx).Create(itsmConfig)
}

// UpdateConfig update itsm config.
func (dao *itsmConfigDao) UpdateConfig(kit *kit.Kit, itsmConfig *table.ItsmConfig) error {

	i := dao.genQ.ItsmConfig

	_, err := i.WithContext(kit.Ctx).Where(i.Key.Eq(itsmConfig.Key)).
		Select(i.Value, i.WorkflowId, i.StateApproveId).Updates(itsmConfig)
	if err != nil {
		return err
	}

	return nil
}
