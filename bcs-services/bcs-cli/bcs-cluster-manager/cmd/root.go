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
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	klog "k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/add"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/check"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/clean"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/cordon"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/create"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/delete"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/disable"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/drain"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/enable"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/get"
	imported "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/import"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/list"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/move"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/remove"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/retry"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/uncordon"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/update"
)

const (
	defaultCfgFile = "/etc/bcs/bcs.yaml"
)

var (
	cfgFile    string
	flagOutput string // nolint
)

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
}

// NewRootCommand returns the rootCmd instance
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "kubectl-bcs-cluster-manager",
		Short: "kubectl-bcs-cluster-manager used to operate bcs-cluster-manager service",
		Long:  `kubectl-bcs-cluster-manager allows operators to get project info from bcs-cluster-manager`,
	}
	cobra.OnInitialize(ensureConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "cfg", defaultCfgFile, "config file")

	rootCmd.AddCommand(create.NewCreateCmd())
	rootCmd.AddCommand(delete.NewDeleteCmd())
	rootCmd.AddCommand(update.NewUpdateCmd())
	rootCmd.AddCommand(list.NewListCmd())
	rootCmd.AddCommand(get.NewGetCmd())
	rootCmd.AddCommand(add.NewAddCmd())
	rootCmd.AddCommand(check.NewCheckCmd())
	rootCmd.AddCommand(clean.NewCleanCmd())
	rootCmd.AddCommand(cordon.NewCordonCmd())
	rootCmd.AddCommand(disable.NewDisableCmd())
	rootCmd.AddCommand(drain.NewDrainCmd())
	rootCmd.AddCommand(enable.NewEnableCmd())
	rootCmd.AddCommand(imported.NewImportCmd())
	rootCmd.AddCommand(move.NewMoveCmd())
	rootCmd.AddCommand(remove.NewRemoveCmd())
	rootCmd.AddCommand(retry.NewRetryCmd())
	rootCmd.AddCommand(uncordon.NewUncordonCmd())

	return rootCmd
}
