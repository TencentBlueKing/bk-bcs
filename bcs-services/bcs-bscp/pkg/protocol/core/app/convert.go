/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package pbapp

import (
	"bscp.io/pkg/dal/table"
	pbbase "bscp.io/pkg/protocol/core/base"
)

// AppSpec convert pb AppSpec to table AppSpec
func (m *AppSpec) AppSpec() *table.AppSpec {
	if m == nil {
		return nil
	}

	return &table.AppSpec{
		Name:       m.Name,
		ConfigType: table.ConfigType(m.ConfigType),
		Mode:       table.AppMode(m.Mode),
		Memo:       m.Memo,
		Reload:     m.Reload.Reload(),
	}
}

// PbAppSpec convert table AppSpec to pb AppSpec
func PbAppSpec(spec *table.AppSpec) *AppSpec {
	if spec == nil {
		return nil
	}

	return &AppSpec{
		Name:       spec.Name,
		ConfigType: string(spec.ConfigType),
		Mode:       string(spec.Mode),
		Memo:       spec.Memo,
		Reload:     PbReload(spec.Reload),
	}
}

// Reload convert pb Reload to table Reload
func (r *Reload) Reload() *table.Reload {
	if r == nil {
		return nil
	}

	return &table.Reload{
		ReloadType:     table.AppReloadType(r.ReloadType),
		FileReloadSpec: r.FileReloadSpec.FileReloadSpec(),
	}
}

// PbReload convert table Reload to pb Reload
func PbReload(spec *table.Reload) *Reload {
	if spec == nil {
		return nil
	}

	return &Reload{
		ReloadType:     string(spec.ReloadType),
		FileReloadSpec: PbFileReloadSpec(spec.FileReloadSpec),
	}
}

// FileReloadSpec convert pb FileReloadSpec to table FileReloadSpec
func (f *FileReloadSpec) FileReloadSpec() *table.FileReloadSpec {
	if f == nil {
		return nil
	}

	return &table.FileReloadSpec{
		ReloadFilePath: f.ReloadFilePath,
	}
}

// PbFileReloadSpec convert table FileReloadSpec to pb FileReloadSpec
func PbFileReloadSpec(spec *table.FileReloadSpec) *FileReloadSpec {
	if spec == nil {
		return nil
	}

	return &FileReloadSpec{
		ReloadFilePath: spec.ReloadFilePath,
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
