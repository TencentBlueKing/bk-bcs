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

// Package cmd provides command-line operations for upgrading and rolling back database table structures.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gorm.io/gorm"

	// run the init function to add migrations
	_ "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/data-service/db-migration/migrations"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/data-service/db-migration/migrator"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/scripts/migrations/itsm"
)

// cmd for migration
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "database migrations tool",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

// sub migrate cmd to create migration
var migrateCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new empty migrations file",
	Run: func(cmd *cobra.Command, args []string) {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Println("Unable to read flag `name`, err:", err)
			return
		}

		var mode string
		mode, err = cmd.Flags().GetString("mode")
		if err != nil {
			fmt.Println("Unable to read flag `mode`, err:", err)
			return
		}

		if err = migrator.Create(name, mode); err != nil {
			fmt.Println("Unable to create migration, err:", err)
			return
		}
	},
}

// sub migrate cmd to execute up migration
var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "run up migrations",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			err   error
			step  int
			debug bool
			db    *gorm.DB
			mig   *migrator.Migrator
		)

		step, err = cmd.Flags().GetInt("step")
		if err != nil {
			fmt.Println("Unable to read flag `step`, err:", err)
			return
		}
		debug, err = cmd.Flags().GetBool("debug")
		if err != nil {
			fmt.Println("Unable to read flag `debug`, err:", err)
			return
		}

		err = cc.LoadSettings(SysOpt)
		if err != nil {
			fmt.Println("load settings from config files failed, err:", err)
			return
		}

		logs.InitLogger(cc.DataService().Log.Logs())

		db, err = migrator.NewDB(debug)
		if err != nil {
			fmt.Println("Unable to new db migrator, err:", err)
			return
		}

		mig, err = migrator.Init(db)
		if err != nil {
			fmt.Println("Unable to fetch migrator, err:", err)
			return
		}

		err = mig.Up(step)
		if err != nil {
			fmt.Println("Unable to run `up` migrations, err:", err)
			return
		}

	},
}

// sub migrate cmd to execute down migration
var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "run down migrations",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			err   error
			step  int
			debug bool
			db    *gorm.DB
			mig   *migrator.Migrator
		)

		step, err = cmd.Flags().GetInt("step")
		if err != nil {
			fmt.Println("Unable to read flag `step`, err:", err)
			return
		}
		debug, err = cmd.Flags().GetBool("debug")
		if err != nil {
			fmt.Println("Unable to read flag `debug`, err:", err)
			return
		}

		if err = cc.LoadSettings(SysOpt); err != nil {
			fmt.Println("load settings from config files failed, err:", err)
			return
		}

		logs.InitLogger(cc.DataService().Log.Logs())

		db, err = migrator.NewDB(debug)
		if err != nil {
			fmt.Println("Unable to new db migrator, err:", err)
			return
		}

		mig, err = migrator.Init(db)
		if err != nil {
			fmt.Println("Unable to fetch migrator, err:", err)
			return
		}

		err = mig.Down(step)
		if err != nil {
			fmt.Println("Unable to run `down` migrations, err:", err)
			return
		}
	},
}

// sub migrate cmd to get migration status
var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "display status of each migrations",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			err   error
			debug bool
			db    *gorm.DB
			mig   *migrator.Migrator
		)

		debug, err = cmd.Flags().GetBool("debug")
		if err != nil {
			fmt.Println("Unable to read flag `debug`, err:", err)
			return
		}

		if err = cc.LoadSettings(SysOpt); err != nil {
			fmt.Println("load settings from config files failed, err:", err)
			return
		}

		logs.InitLogger(cc.DataService().Log.Logs())

		db, err = migrator.NewDB(debug)
		if err != nil {
			fmt.Println("Unable to new db migrator, err:", err)
			return
		}

		mig, err = migrator.Init(db)
		if err != nil {
			fmt.Println("Unable to fetch migrator, err:", err)
			return
		}

		if err = mig.MigrationStatus(); err != nil {
			fmt.Println("Unable to fetch migration status, err:", err)
			return
		}

	},
}

var migrateInitITSMCmd = &cobra.Command{
	Use:   "init-itsm",
	Short: "Register bcsp approve services into itsm",
	Run: func(cmd *cobra.Command, args []string) {

		if err := cc.LoadSettings(SysOpt); err != nil {
			fmt.Printf("load config failed, err: %s\n", err.Error())
			os.Exit(1)
		}

		if !cc.DataService().ITSM.Enable || !cc.DataService().ITSM.AutoRegister {
			fmt.Println("itsm is disabled or auto register is disabled, skip migration")
			return
		}

		if err := itsm.InitServices(); err != nil {
			fmt.Printf("init itsm services failed, err: %s\n", err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	// Add "--debug" flag to all migrate sub commands
	migrateCmd.PersistentFlags().BoolP("debug", "d", false, "whether debug gorm to print sql, default is false")

	// Add "--name" flag to "create" command
	migrateCreateCmd.Flags().StringP("name", "n", "", "Name for the migration")

	// Add "--gorm" flag to "create" command
	migrateCreateCmd.Flags().StringP("mode", "m", "gorm", "mode of the migration:sql or gorm, default is gorm")

	// Add "--step" flag to both "up" and "down" command
	migrateUpCmd.Flags().IntP("step", "s", 0, "Number of migrations to execute")
	migrateDownCmd.Flags().IntP("step", "s", 0, "Number of migrations to execute")

	// Add "create", "up" and "down" commands to the "migrate" command
	migrateCmd.AddCommand(migrateUpCmd, migrateDownCmd, migrateCreateCmd, migrateStatusCmd)
	migrateCmd.AddCommand(migrateInitITSMCmd)

	// Add "migrate" command to the root command
	rootCmd.AddCommand(migrateCmd)
}
