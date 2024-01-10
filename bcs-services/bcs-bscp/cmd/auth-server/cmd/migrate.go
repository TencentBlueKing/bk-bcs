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

// Package cmd provides operations for upgrading the permission model.
package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/client"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/sys"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/metrics"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "auth-server migrations tool",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var migrateInitCmd = &cobra.Command{
	Use:   "init-iam",
	Short: "Initializes the authority center model",
	Run: func(cmd *cobra.Command, args []string) {

		if err := cc.LoadSettings(SysOpt.Sys); err != nil {
			fmt.Println("load settings from config files failed, err:", err)
			return
		}

		iamSys, err := NewIamSys()
		if err != nil {
			fmt.Printf("new iam sys failed, err: %v\n", err)
		}

		if err := iamSys.Register(context.Background(), cc.AuthServer().IAM.Host); err != nil {
			fmt.Printf("initialize service failed, err: %v\n", err)
			return
		}

	},
}

// NewIamSys new a iamSystem
func NewIamSys() (*sys.Sys, error) {

	iamSettings := cc.AuthServer().IAM
	tlsConfig := new(tools.TLSConfig)
	if iamSettings.TLS.Enable() {
		tlsConfig = &tools.TLSConfig{
			InsecureSkipVerify: iamSettings.TLS.InsecureSkipVerify,
			CertFile:           iamSettings.TLS.CertFile,
			KeyFile:            iamSettings.TLS.KeyFile,
			CAFile:             iamSettings.TLS.CAFile,
			Password:           iamSettings.TLS.Password,
		}
	}
	cfg := &client.Config{
		Address:   []string{iamSettings.APIURL},
		AppCode:   iamSettings.AppCode,
		AppSecret: iamSettings.AppSecret,
		SystemID:  sys.SystemIDBSCP,
		TLS:       tlsConfig,
	}
	iamCli, err := client.NewClient(cfg, metrics.Register())
	if err != nil {
		return nil, err
	}

	iamSys, err := sys.NewSys(iamCli)
	if err != nil {
		return nil, fmt.Errorf("new iam sys failed, err: %v", err)
	}
	logs.Infof("initialize iam sys success.")

	return iamSys, nil

}

func init() {

	// Add "--debug" flag to all migrate sub commands
	migrateCmd.PersistentFlags().BoolP("debug", "d", false,
		"whether to debug output the execution process,, default is false")

	migrateCmd.AddCommand(migrateInitCmd)

	// Add "migrate" command to the root command
	rootCmd.AddCommand(migrateCmd)
}
