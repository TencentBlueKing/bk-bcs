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

package main

import (
	"gorm.io/gen"

	"bscp.io/pkg/dal/table"
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
		table.TemplateSpace{},
		table.Hook{},
		table.HookRelease{},
		table.Release{},
		table.ConfigHook{},
		table.App{},
	)

	g.Execute()
}
