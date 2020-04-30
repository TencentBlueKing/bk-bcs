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
	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var cmdList []*cobra.Command

//init all resource create sub command
func init() {
	lockerCmd := &cobra.Command{
		Use:   "lock",
		Short: "lock resource",
		Long:  "lock specified resource, such ConfigSet",
	}
	lockerCmd.AddCommand(lockConfigSetCmd())
	cmdList = append(cmdList, lockerCmd)

	unlockCmd := &cobra.Command{
		Use:   "unlock",
		Short: "unlock resource",
		Long:  "unlock specified resource, such ConfigSet",
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
		Short:   "lock configset",
		Long:    "lock ConfigSet by specified ID",
		Example: `
	bscp-client lock configset --Id specifiedID
	or
	bscp-client lock cfgset --name configsetname --app appName
		`,
		RunE: func(c *cobra.Command, args []string) error {
			operator := service.NewOperator(option.GlobalOptions)
			if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
				return err
			}
			//todo(DeveloperJim): add configSet name support
			ID, _ := c.Flags().GetString("Id")
			if len(ID) == 0 {
				//go by name and AppName
				name, _ := c.Flags().GetString("name")
				appName, _ := c.Flags().GetString("app")
				if len(name) == 0 || len(appName) == 0 {
					return fmt.Errorf("parameter name and app are required")
				}
				configset, err := operator.GetConfigSet(context.TODO(), appName, name)
				if err != nil {
					return err
				}
				if configset == nil {
					return fmt.Errorf("No relative ConfigSet %s", name)
				}
				ID = configset.Cfgsetid
			}
			if err := operator.LockConfigSet(context.TODO(), ID); err != nil {
				return err
			}
			c.Printf("Lock ConfigSet successfully: %s\n", ID)
			return nil
		},
	}
	// --Id is required
	cmd.Flags().String("Id", "", "specified ConfigSet ID")
	cmd.Flags().StringP("name", "n", "", "specified ConfigSet name")
	cmd.Flags().StringP("app", "a", "", "specified Application name for filter")
	return cmd
}

//unlockConfigSetCmd: client lock configset --id xxxxx
func unlockConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configset",
		Aliases: []string{"cfgset"},
		Short:   "unlock configset",
		Long:    "lock ConfigSet by specified ID",
		Example: `
	bscp-client unlock configset --Id specifiedID
	or 
	bscp-client unlock cfgset --name cfgname --app gameserver
		`,
		RunE: func(c *cobra.Command, args []string) error {
			operator := service.NewOperator(option.GlobalOptions)
			if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
				return err
			}
			//todo(DeveloperJim): add configSet name support
			ID, _ := c.Flags().GetString("Id")
			if len(ID) == 0 {
				//go by name and AppName
				name, _ := c.Flags().GetString("name")
				appName, _ := c.Flags().GetString("app")
				if len(name) == 0 || len(appName) == 0 {
					return fmt.Errorf("parameter name and app are required")
				}
				configset, err := operator.GetConfigSet(context.TODO(), appName, name)
				if err != nil {
					return err
				}
				if configset == nil {
					return fmt.Errorf("No relative ConfigSet %s", name)
				}
				ID = configset.Cfgsetid
			}
			if err := operator.UnLockConfigSet(context.TODO(), ID); err != nil {
				return err
			}
			c.Printf("unLock ConfigSet successfully: %s\n", ID)
			return nil
		},
	}
	// --Id is required
	cmd.Flags().String("Id", "", "specified ConfigSet ID")
	cmd.Flags().StringP("name", "n", "", "specified ConfigSet name")
	cmd.Flags().StringP("app", "a", "", "specified Application name for filter")
	return cmd
}
