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

package rollback

import (
	"context"

	"github.com/spf13/cobra"

	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
)

// InitCommands init all cancel commands.
func InitCommands() []*cobra.Command {
	// init all sub resource command.
	return []*cobra.Command{rollbackReleaseCmd()}
}

// rollbackMultiReleaseCmd: client rollback multi-commit.
func rollbackReleaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rollback",
		Aliases: []string{"rb"},
		Short:   "Rollback release",
		Long:    "RollBack to the specified release version",
		Example: `
	bk-bscp-client rollback --id xxxxxxxx
		`,
		RunE: handRollbackRelease,
	}

	cmd.Flags().StringP("id", "i", "", "the id of release")
	cmd.MarkFlagRequired("id")
	return cmd
}

func handRollbackRelease(cmd *cobra.Command, args []string) error {
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	multiCommitId, _ := cmd.Flags().GetString("id")

	err := operator.RollbackMultiReleaseById(context.TODO(), multiCommitId)
	if err != nil {
		return err
	}
	cmd.Printf("Rollback successfully: %s\n\n", multiCommitId)
	return nil
}
