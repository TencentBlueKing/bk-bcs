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

package cmd

import (
	"flag"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/klog"
)

const defaultCfgFile = "/etc/bcs/helmctl.yaml"

var (
	cfgFile  string
	jsonData string
	jsonFile string

	rootCMD = &cobra.Command{
		Use:   "helmctl",
		Short: "helm for bcs",
		Long:  "helm for bcs",
	}
)

// Execute is the entrance for cmd tools
func Execute() {
	if err := rootCMD.Execute(); err != nil {
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
		klog.Fatalf("read config from %s failed, %s", cfgFile, err.Error())
	}
}

func init() {
	cobra.OnInitialize(ensureConfig)
	rootCMD.AddCommand(availableCMD)
	rootCMD.AddCommand(getCMD)
	rootCMD.AddCommand(deleteCMD)
	rootCMD.AddCommand(historyCMD)
	rootCMD.AddCommand(installCMD)
	rootCMD.AddCommand(uninstallCMD)
	rootCMD.AddCommand(upgradeCMD)
	rootCMD.AddCommand(rollbackCMD)
	rootCMD.AddCommand(pushCMD)
	rootCMD.AddCommand(diffCMD)
	flag.StringVar(&cfgFile, "config", defaultCfgFile, "config file, yaml mode")
	rootCMD.PersistentFlags().StringVarP(
		&flagOutput, "output", "o", "", "output format, one of json|wide")
	klog.InitFlags(nil)
	flag.Parse()
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
}
