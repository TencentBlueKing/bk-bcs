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

// Package cmd provides operations for init and upgrading the bk-apigateway resources and
// register the system 'bk-bscp' into bk-notice
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/script/migrations/itsm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/script/migrations/tenant"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "bcs-project-manager migrations tool",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

// migrateInitITSMCmd for auto register itsm service and write config into database
var migrateInitITSMCmd = &cobra.Command{
	Use:   "init-itsm",
	Short: "Register bcs related services into itsm",
	Run: func(cmd *cobra.Command, args []string) {

		if _, err := config.LoadConfig(configPath); err != nil {
			fmt.Printf("load config failed, err: %s\n", err.Error())
			os.Exit(1)
		}

		if !config.GlobalConf.ITSM.Enable || !config.GlobalConf.ITSM.AutoRegister {
			fmt.Println("itsm is disabled or auto register is disabled, skip migration")
			return
		}

		if err := itsm.InitServices(); err != nil {
			fmt.Printf("init itsm services failed, err: %s\n", err.Error())
			os.Exit(1)
		}
	},
}

// migrateInitTenantCmd for init projects tenantInfo(default) when env not enable multi tenant
var migrateInitTenantCmd = &cobra.Command{
	Use:   "init-tenant",
	Short: "init bcs project manager tenant info",
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := config.LoadConfig(configPath); err != nil {
			fmt.Printf("load config failed, err: %s\n", err.Error())
			os.Exit(1)
		}

		// 内部环境 且 不开启多租户
		if !config.GlobalConf.EnableMultiTenant {
			fmt.Printf("multi tenant is disabled or auto register is disabled, skip migration")
			return
		}
		if err := tenant.InitProject(); err != nil {
			fmt.Printf("init tenant project failed, err: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Printf("init tenant project successfully\n")
	},
}

func init() {
	migrateCmd.AddCommand(migrateInitITSMCmd)
	migrateCmd.AddCommand(migrateInitTenantCmd)
	rootCmd.AddCommand(migrateCmd)
}
