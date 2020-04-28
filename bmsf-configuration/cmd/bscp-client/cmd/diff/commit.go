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

package diff

import (
	"context"

	"github.com/spf13/cobra"

	"bk-bscp/cmd/bscp-client/cmd/utils"
	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
)

// diffCommitCmd: client diff commit.
func diffCommitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "commit",
		Aliases: []string{"ci"},
		Short:   "diff commit details",
		Long:    "diff commit detail information",
		Example: `
	bscp-client diff commit commitid-1 commitid-2
	bscp-client diff ci commitid-1 commitid-2
		`,
		RunE: handleDiffCommit,
	}

	cmd.Flags().String("mode", "text", "diff view mode, 'text' or 'git'")
	return cmd
}

func handleDiffCommit(cmd *cobra.Command, args []string) error {
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}

	// check options.
	if len(args) != 2 {
		cmd.Usage()
		return nil
	}
	commitID1 := args[0]
	commitID2 := args[1]

	// query target commits.
	commit1, err := operator.GetCommit(context.TODO(), commitID1)
	if err != nil {
		return err
	}
	if commit1 == nil {
		cmd.Printf("Found no Commit[%s] resource.\n", commitID1)
		return nil
	}
	commit2, err := operator.GetCommit(context.TODO(), commitID2)
	if err != nil {
		return err
	}
	if commit2 == nil {
		cmd.Printf("Found no Commit[%s] resource.\n", commitID2)
		return nil
	}

	mode, _ := cmd.Flags().GetString("mode")
	if mode == "git" {
		utils.PrintCommitDiffsGitMode(commit1, commit2)
	} else {
		utils.PrintCommitDiffsTextMode(commit1, commit2)
	}
	return nil
}
