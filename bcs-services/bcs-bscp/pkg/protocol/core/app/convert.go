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

// Package pbapp provides application core protocol struct and convert functions.
package pbapp

import (
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
)

// AppSpec convert pb AppSpec to table AppSpec
func (m *AppSpec) AppSpec() *table.AppSpec {
	if m == nil {
		return nil
	}

	return &table.AppSpec{
		Name:        m.Name,
		ConfigType:  table.ConfigType(m.ConfigType),
		Memo:        m.Memo,
		Alias:       m.Alias,
		DataType:    table.DataType(m.DataType),
		ApproveType: table.ApproveType(m.ApproveType),
		IsApprove:   m.IsApprove,
		Approver:    m.Approver,
	}
}

// PbAppSpec convert table AppSpec to pb AppSpec
func PbAppSpec(spec *table.AppSpec) *AppSpec { //nolint:revive
	if spec == nil {
		return nil
	}

	return &AppSpec{
		Name:        spec.Name,
		ConfigType:  string(spec.ConfigType),
		Memo:        spec.Memo,
		Alias:       spec.Alias,
		DataType:    string(spec.DataType),
		IsApprove:   spec.IsApprove,
		ApproveType: string(spec.ApproveType),
		Approver:    spec.Approver,
	}
}

// PbApps convert table Apps to pb Apps
func PbApps(apps []*table.App) []*App {
	if apps == nil {
		return make([]*App, 0)
	}

	result := make([]*App, 0)
	for _, app := range apps {
		result = append(result, PbApp(app))
	}

	return result
}

// PbApp convert table App to pb App
func PbApp(app *table.App) *App {
	if app == nil {
		return nil
	}

	return &App{
		Id:       app.ID,
		BizId:    app.BizID,
		Spec:     PbAppSpec(app.Spec),
		Revision: pbbase.PbRevision(app.Revision),
	}
}
