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

// Package workflow xxx
package workflow

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/pkg/workflow"
)

// NewWorkflowCmd create workflow command
func NewWorkflowCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "workflow",
		Short: "Manage workflow resources",
	}
	command.AddCommand(list())
	command.AddCommand(create())
	command.AddCommand(getDetail())
	command.AddCommand(deleteWorkflow())
	command.AddCommand(update())
	command.AddCommand(execute())
	command.AddCommand(listHistories())
	command.AddCommand(historyGet())
	return command
}

var (
	listProjects = &([]string{})
)

func list() *cobra.Command {
	c := &cobra.Command{
		Use:   "list",
		Short: "List workflows",
		Run: func(cmd *cobra.Command, args []string) {
			h := workflow.NewHandler()
			h.List(cmd.Context(), listProjects)
		},
	}
	listProjects = c.Flags().StringSliceP("projects", "p", nil, "Filter by project name")
	return c
}

var (
	createFile string
)

func create() *cobra.Command {
	c := &cobra.Command{
		Use:   "create",
		Short: "Create workflow with setting files",
		Run: func(cmd *cobra.Command, args []string) {
			h := workflow.NewHandler()
			h.Create(cmd.Context(), createFile)
		},
	}
	c.PersistentFlags().StringVarP(&createFile, "file", "f", "",
		"The files that contain the configurations of a workflow")
	return c
}

var (
	workflowID string
)

func getDetail() *cobra.Command {
	c := &cobra.Command{
		Use:   "get ID",
		Short: "Get workflow details",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			getID := args[0]
			h := workflow.NewHandler()
			h.GetDetail(cmd.Context(), getID)
		},
	}
	return c
}

func deleteWorkflow() *cobra.Command {
	c := &cobra.Command{
		Use:   "delete ID",
		Short: "Delete workflow with ID",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			deleteID := args[0]
			h := workflow.NewHandler()
			h.Delete(cmd.Context(), deleteID)
		},
	}
	return c
}

var (
	updateFile string
)

func update() *cobra.Command {
	c := &cobra.Command{
		Use:   "update",
		Short: "Update specified workflow with setting files",
		Run: func(cmd *cobra.Command, args []string) {
			h := workflow.NewHandler()
			h.Update(cmd.Context(), updateFile)
		},
	}
	c.PersistentFlags().StringVarP(&updateFile, "file", "f", "",
		"The files that contain the configurations of a workflow")
	return c
}

func execute() *cobra.Command {
	c := &cobra.Command{
		Use:   "execute ID",
		Short: "Execute workflow with ID",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			executeID := args[0]
			h := workflow.NewHandler()
			h.Execute(cmd.Context(), executeID)
		},
	}
	return c
}

func listHistories() *cobra.Command {
	c := &cobra.Command{
		Use:   "history-list ID",
		Short: "List histories by workflow",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			wfid := args[0]
			h := workflow.NewHandler()
			h.ListHistories(cmd.Context(), wfid)
		},
	}
	return c
}

var (
	showHistoryDetails bool
)

func historyGet() *cobra.Command {
	c := &cobra.Command{
		Use:   "history-get HISTORY-ID",
		Short: "Get history info by history id",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			historyID := args[0]
			h := workflow.NewHandler()
			h.HistoryDetail(cmd.Context(), historyID, showHistoryDetails)
		},
	}
	c.PersistentFlags().BoolVar(&showHistoryDetails, "show-details", false,
		"Whether show the history details")
	return c
}
