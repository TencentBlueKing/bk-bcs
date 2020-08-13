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

package lock

import (
	"context"
	"fmt"
	"path"

	"github.com/spf13/cobra"

	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
	"bk-bscp/internal/protocol/common"
)

var cmdList []*cobra.Command

//init all resource create sub command
func init() {
	lockerCmd := &cobra.Command{
		Use:   "lock",
		Short: "Lock resource",
		Long:  "Lock specified resource, such ConfigSet",
	}
	lockerCmd.AddCommand(lockConfigSetCmd())
	cmdList = append(cmdList, lockerCmd)

	unlockCmd := &cobra.Command{
		Use:   "unlock",
		Short: "Unlock resource",
		Long:  "Unlock specified resource, such ConfigSet",
	}
	unlockCmd.AddCommand(unlockConfigSetCmd())
	cmdList = append(cmdList, unlockCmd)
}

//InitCommands init all create commands
func InitCommands() []*cobra.Command {
	return cmdList
}

//lockConfigSetCmd: client lock configset --id xxxxx
func lockConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configset",
		Aliases: []string{"cfgset"},
		Short:   "Lock ConfigSet",
		Long:    "Lock ConfigSet by specified ConfigSet ID or ConfigSet",
		Example: `
	bk-bscp-client lock configset --id configsetId
	bk-bscp-client lock cfgset --cfgset /etc/server.yaml
		`,
		RunE: func(c *cobra.Command, args []string) error {
			option.SetGlobalVarByName(c, "app") // not judge input
			operator := service.NewOperator(option.GlobalOptions)
			if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
				return err
			}
			//todo(DeveloperJim): add configSet name support
			ID, _ := c.Flags().GetString("id")
			cfgset, _ := c.Flags().GetString("cfgset")
			appName, _ := c.Flags().GetString("app")
			if len(appName) == 0 {
				return fmt.Errorf("parameter are required")
			}
			cfgsetFpath, cfgsetName := path.Split(cfgset)
			query := &common.ConfigSet{
				Cfgsetid: ID,
				Name:     cfgsetName,
				Fpath:    cfgsetFpath,
			}
			configset, err := operator.GetConfigSet(context.TODO(), appName, query)
			if err != nil {
				return err
			}
			if configset == nil {
				return fmt.Errorf("No relative ConfigSet by your parameter")
			}
			ID = configset.Cfgsetid

			if err := operator.LockConfigSet(context.TODO(), ID); err != nil {
				return err
			}
			c.Printf("Lock ConfigSet successfully: %s\n\n", ID)
			return nil
		},
	}
	// --Id is required
	cmd.Flags().StringP("id", "i", "", "specified ConfigSet ID")
	cmd.Flags().StringP("cfgset", "c", "", "specified ConfigSet")
	cmd.Flags().StringP("app", "a", "", "specified Application name for filter")
	return cmd
}

//unlockConfigSetCmd: client lock configset --id xxxxx
func unlockConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configset",
		Aliases: []string{"cfgset"},
		Short:   "Unlock configset",
		Long:    "Unlock ConfigSet by specified ConfigSet ID or ConfigSet",
		Example: `
	bk-bscp-client unlock configset --Id configsetId
	bk-bscp-client unlock cfgset --cfgset /etc/server.yaml
		`,
		RunE: func(c *cobra.Command, args []string) error {
			err := option.SetGlobalVarByName(c, "app")
			if err != nil {
				return err
			}
			operator := service.NewOperator(option.GlobalOptions)
			if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
				return err
			}
			//todo(DeveloperJim): add configSet name support
			ID, _ := c.Flags().GetString("id")
			cfgset, _ := c.Flags().GetString("cfgset")
			appName, _ := c.Flags().GetString("app")
			if len(appName) == 0 {
				return fmt.Errorf("parameter app are required")
			}
			cfgsetFpath, cfgsetName := path.Split(cfgset)
			query := &common.ConfigSet{
				Cfgsetid: ID,
				Name:     cfgsetName,
				Fpath:    cfgsetFpath,
			}
			configset, err := operator.GetConfigSet(context.TODO(), appName, query)
			if err != nil {
				return err
			}
			if configset == nil {
				return fmt.Errorf("No relative ConfigSet by your parameter")
			}
			ID = configset.Cfgsetid

			if err := operator.UnLockConfigSet(context.TODO(), ID); err != nil {
				return err
			}
			c.Printf("unLock ConfigSet successfully: %s\n\n", ID)
			return nil
		},
	}
	// --Id is required
	cmd.Flags().StringP("id", "i", "", "specified ConfigSet ID")
	cmd.Flags().StringP("cfgset", "c", "", "specified ConfigSet")
	cmd.Flags().StringP("app", "a", "", "specified Application name for filter")
	return cmd
}
