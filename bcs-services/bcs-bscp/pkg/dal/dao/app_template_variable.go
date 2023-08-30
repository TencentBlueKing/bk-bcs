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
	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
)

// AppTemplateVariable supplies all the app template binding related operations.
type AppTemplateVariable interface {
	// Get gets app template variables
	Get(kit *kit.Kit, bizID, appID uint32) (*table.AppTemplateVariable, error)
	// ListVariables lists all variables in app template variable
	ListVariables(kit *kit.Kit, bizID, appID uint32) ([]*table.TemplateVariableSpec, error)
}

var _ AppTemplateVariable = new(appTemplateVariableDao)

type appTemplateVariableDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// Get gets app template variables.
func (dao *appTemplateVariableDao) Get(kit *kit.Kit, bizID, appID uint32) (*table.AppTemplateVariable, error) {
	m := dao.genQ.AppTemplateVariable
	q := dao.genQ.AppTemplateVariable.WithContext(kit.Ctx)
	appVars, err := q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appID)).Find()
	if err != nil {
		return nil, err
	}
	if len(appVars) == 0 {
		return nil, nil
	}
	return appVars[0], nil
}

// ListVariables lists all variables in app template variable
func (dao *appTemplateVariableDao) ListVariables(kit *kit.Kit, bizID, appID uint32) (
	[]*table.TemplateVariableSpec, error) {
	m := dao.genQ.AppTemplateVariable
	q := dao.genQ.AppTemplateVariable.WithContext(kit.Ctx)
	appVars, err := q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appID)).Find()
	if err != nil {
		return nil, err
	}
	if len(appVars) == 0 {
		return []*table.TemplateVariableSpec{}, nil
	}
	return appVars[0].Spec.Variables, nil
}
