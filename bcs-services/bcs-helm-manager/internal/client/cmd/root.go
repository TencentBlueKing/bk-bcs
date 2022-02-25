/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const defaultCfgFile = "/etc/bcs/helmctl.yaml"

var (
	cfgFile  string
	jsonData string
	jsonFile string

	rootCMD = &cobra.Command{
		Use:   "helmctl",
		Short: "kubectl plugin for bcs helm manager",
		Long:  "kubectl plugin for bcs helm manager",
	}
)

// Execute is the entrance for cmd tools
func Execute() {
	if err := rootCMD.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func ensureConfig() {
	if cfgFile == "" {
		cfgFile = defaultCfgFile
	}

	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("read config from %s failed, %s\n", cfgFile, err.Error())
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(ensureConfig)
	rootCMD.AddCommand(availableCMD)
	rootCMD.AddCommand(getCMD)
	rootCMD.AddCommand(createCMD)
	rootCMD.AddCommand(updateCMD)
	rootCMD.AddCommand(deleteCMD)
	rootCMD.AddCommand(installCMD)
	rootCMD.AddCommand(uninstallCMD)
	rootCMD.AddCommand(upgradeCMD)
	rootCMD.AddCommand(rollbackCMD)
	rootCMD.PersistentFlags().StringVarP(
		&cfgFile, "config", "c", defaultCfgFile, "config file")
}
