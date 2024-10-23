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

// Package main is the entry of the gorm/gen code generator.
package main

import (
	"gorm.io/gen"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath:       "./pkg/dal/gen",
		Mode:          gen.WithDefaultQuery | gen.WithQueryInterface,
		FieldNullable: true,
	})

	// 需要 Gen 的模型这里添加
	g.ApplyBasic(
		table.IDGenerator{},
		table.Audit{},
		table.App{},
		table.ArchivedApp{},
		table.ConfigItem{},
		table.ReleasedConfigItem{},
		table.Commit{},
		table.Content{},
		table.ResourceLock{},
		table.Event{},
		table.Credential{},
		table.CredentialScope{},
		table.Strategy{},
		table.Group{},
		table.ReleasedGroup{},
		table.GroupAppBind{},
		table.Release{},
		table.ReleasedConfigItem{},
		table.Hook{},
		table.HookRevision{},
		table.ReleasedHook{},
		table.TemplateSpace{},
		table.Template{},
		table.TemplateSet{},
		table.TemplateRevision{},
		table.TemplateVariable{},
		table.AppTemplateBinding{},
		table.ReleasedAppTemplate{},
		table.AppTemplateVariable{},
		table.ReleasedAppTemplateVariable{},
		table.Kv{},
		table.ReleasedKv{},
		table.Client{},
		table.ClientEvent{},
		table.ClientQuery{},
		table.ItsmConfig{},
	)

	g.Execute()
}
