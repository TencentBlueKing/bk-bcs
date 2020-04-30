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

package confirm

import (
	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
	"context"

	"github.com/spf13/cobra"
)

var confirm *cobra.Command

//init all resource create sub command
func init() {
	confirm = &cobra.Command{
		Use:   "confirm",
		Short: "confirm resource like commit/release",
		Long:  "confirm resource and make it affective, we only support Commit/Release Confirm now.",
	}
}

//InitCommands init all create commands
func InitCommands() []*cobra.Command {
	confirm.AddCommand(confirmCommitCmd())
	confirm.AddCommand(confirmReleaseCmd())
	return []*cobra.Command{confirm}
}

//lockConfigSetCmd: client lock configset --id xxxxx
func confirmCommitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "commit",
		Aliases: []string{"ci"},
		Short:   "confirm specified Commit",
		Long:    "confirm specified Commit, make it affective or to generate from templates",
		Example: `
		bscp-client confirm commit --business somebusiness --Id specifiedID
		bscp-client confirm ci --business somebusiness --Id specifiedID
		`,
		RunE: func(c *cobra.Command, args []string) error {
			operator := service.NewOperator(option.GlobalOptions)
			if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
				return err
			}
			ID, _ := c.Flags().GetString("Id")
			err := operator.ConfirmCommit(context.TODO(), ID)
			if err == nil {
				c.Printf("Confirm Commit successfully: %s\n", ID)
			}
			return err
		},
	}
	// --Id is required
	cmd.Flags().String("Id", "", "specified Commit ID")
	cmd.MarkFlagRequired("Id")
	return cmd
}

//confirmReleaseCmd: client confirm release
func confirmReleaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "release",
		Aliases: []string{"rel"},
		Short:   "confirm specified Release, publish to endpoints",
		Long:    "confirm specified release, ready to publish to all endpoints",
		Example: `
		bscp-client confirm release --business some --Id xxx
		bscp-client confirm rel --business some --Id xxx
		`,
		RunE: func(c *cobra.Command, args []string) error {
			operator := service.NewOperator(option.GlobalOptions)
			if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
				return err
			}
			ID, _ := c.Flags().GetString("Id")
			err := operator.ConfirmRelease(context.TODO(), ID)
			if err == nil {
				c.Printf("release Id %s confirm to publish successfully\n", ID)
				return nil
			}
			return err
		},
	}
	// --Id is required
	cmd.Flags().String("Id", "", "specified Release ID")
	cmd.MarkFlagRequired("Id")
	return cmd
}
