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

package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog"

	"github.com/Tencent/bk-bcs/bcs-common/common/version"
)

const (
	defaultCfgFile = "/etc/bcs/bcs-project-manager.yaml"
)

var (
	cfgFile    string
	flagOutput string
	rootCmd    = &cobra.Command{
		Use:   "kubectl-bcs-project-manager",
		Short: "kubectl-bcs-project-manager used to operate bcs-project-manager service",
		Long: `
kubectl-bcs-project-manager allows operators to get project info from bcs-project-manager.
`,
	}
)

// Execute is the entrance for cmd tools
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		klog.Fatalf(err.Error())
	}
}

func ensureConfig() {
	if cfgFile == "" {
		cfgFile = defaultCfgFile
	}
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
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
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newUpdateCmd())
	rootCmd.PersistentFlags().StringVarP(
		&cfgFile, "config", "c", defaultCfgFile, "config file")
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "wide",
		"optional parameter: json/wide, json will print the json string to stdout")
}
