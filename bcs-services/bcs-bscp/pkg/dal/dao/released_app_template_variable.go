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
	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
)

// ReleasedAppTemplateVariable supplies all the template revision related operations.
type ReleasedAppTemplateVariable interface {
	// CreateWithTx create one app template variable instance with transaction.
	CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, variable *table.ReleasedAppTemplateVariable) (uint32, error)
	// ListVariables lists all variables in released app template variable
	ListVariables(kit *kit.Kit, bizID, appID, releaseID uint32) ([]*table.TemplateVariableSpec, error)
}

var _ ReleasedAppTemplateVariable = new(releasedAppTemplateVariableDao)

type releasedAppTemplateVariableDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// CreateWithTx create one app template variable instance with transaction.
func (dao *releasedAppTemplateVariableDao) CreateWithTx(
	kit *kit.Kit, tx *gen.QueryTx, g *table.ReleasedAppTemplateVariable) (uint32, error) {
	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}

	// generate a Template id and update to Template.
	id, err := dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, err
	}
	g.ID = id

	q := tx.ReleasedAppTemplateVariable.WithContext(kit.Ctx)
	if err := q.Create(g); err != nil {
		return 0, err
	}

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)
	if err := ad.Do(tx.Query); err != nil {
		return 0, err
	}

	return g.ID, nil
}

// ListVariables lists all variables in released app template variable
func (dao *releasedAppTemplateVariableDao) ListVariables(kit *kit.Kit, bizID, appID, releaseID uint32) (
	[]*table.TemplateVariableSpec, error) {
	m := dao.genQ.ReleasedAppTemplateVariable
	q := dao.genQ.ReleasedAppTemplateVariable.WithContext(kit.Ctx)
	appVars, err := q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appID), m.ReleaseID.Eq(releaseID)).
		Find()
	if err != nil {
		return nil, err
	}
	if len(appVars) == 0 {
		return []*table.TemplateVariableSpec{}, nil
	}
	return appVars[0].Spec.Variables, nil
}
