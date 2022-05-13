/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const defaultCfgFile = "/etc/bcs/bcs-data-manager.yaml"

var (
	cfgFile string

	rootCMD = &cobra.Command{
		Use:   "kubectl-bcs-data-manager",
		Short: "kubectl plugin for bcs data manager",
		Long:  "kubectl plugin for bcs data manager",
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
	rootCMD.AddCommand(getCMD)
	rootCMD.AddCommand(listCMD)
	rootCMD.PersistentFlags().StringVarP(
		&cfgFile, "config", "c", defaultCfgFile, "config file")
}
