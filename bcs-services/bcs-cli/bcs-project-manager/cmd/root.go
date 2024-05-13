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

// Package cmd xxx
package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	klog "k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/cmd/create"
	delete2 "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/cmd/delete"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/cmd/edit"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/cmd/get"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/cmd/render"
)

var (
	cfgFile = "/etc/bcs/bcs.yaml"
	rootCmd = &cobra.Command{
		Use:   "kubectl-bcs-project-manager",
		Short: "kubectl-bcs-project-manager used to operate bcs-project-manager service",
		Long: `
kubectl-bcs-project-manager allows operators to get project info from bcs-project-manager.
`,
	}
	debug bool
)

// Execute is the entrance for cmd tools
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		klog.Infoln(err.Error())
		return
	}
}

func ensureConfig() {
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	err := viper.ReadInConfig()
	if cfgFile != "" && err != nil {
		klog.Fatalf("read config from '%s' failed,err: %s", cfgFile, err.Error())
	}
}

func init() {
	log.SetFlags(0)
	cobra.OnInitialize(ensureConfig)
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "print the version detail info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.GetVersion())
		},
	})
	rootCmd.AddCommand(get.NewCmdGet())
	rootCmd.AddCommand(edit.NewCmdEdit())
	rootCmd.AddCommand(create.NewCmdCreate())
	rootCmd.AddCommand(render.NewCmdRender())
	rootCmd.AddCommand(delete2.NewCmdDelete())
	rootCmd.PersistentFlags().StringVarP(
		&cfgFile, "config", "c", cfgFile, "config file, optional")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "v", false, "Debug mode")
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}
