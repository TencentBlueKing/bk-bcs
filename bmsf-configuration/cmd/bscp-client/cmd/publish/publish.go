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

package publish

import (
	"context"

	"github.com/spf13/cobra"

	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
)

//InitCommands init all create commands
func InitCommands() []*cobra.Command {
	return []*cobra.Command{publishCmd()}
}

//confirmReleaseCmd: client confirm release
func publishCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "publish",
		Aliases: []string{"pb"},
		Short:   "Publish release",
		Long:    "Publish release to make the configuration effective",
		Example: `
	bk-bscp-client publish --id <releaseid>
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			operator := service.NewOperator(option.GlobalOptions)
			if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
				return err
			}
			ID, _ := cmd.Flags().GetString("id")
			err := operator.ConfirmMultiRelease(context.TODO(), ID)
			if err != nil {
				return err
			}
			cmd.Printf("Publish successfully: %s\n", ID)
			cmd.Println()
			cmd.Printf("\tuse \"bk-bscp-client get release --id <releaseid>\" to get release detail\n\n")
			return nil
		},
	}
	// --Id is required
	cmd.Flags().StringP("id", "i", "", "the id of release")
	cmd.MarkFlagRequired("id")
	return cmd
}
