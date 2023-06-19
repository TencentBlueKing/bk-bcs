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
 *
 */

package main

import (
	"context"

	"github.com/oklog/run"
	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/migration"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage"
)

// MigrateCmd :
func MigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "migrate monitor data",
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		runCmd(cmd, runMigrate)
	}

	return cmd
}

// 运行 migrate
func runMigrate(ctx context.Context, g *run.Group, opt *option) error {
	// init storage
	storage.InitStorage()
	g.Add(migration.MigrateLogRule, func(err error) {})
	return nil
}
