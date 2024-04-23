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

// Package argocd defines the argocd command
package argocd

import (
	"github.com/argoproj/argo-cd/v2/cmd/argocd/commands"
	"github.com/argoproj/argo-cd/v2/cmd/argocd/commands/initialize"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// NewArgoCmd create the argo command
func NewArgoCmd() *cobra.Command {
	argo := commands.NewCommand()
	argo.Short = "Controls a Argo CD server"
	for _, item := range argo.Commands() {
		if item.Use == "app" {
			item.AddCommand(initialize.InitCommand(applicationDryRun()))
			item.AddCommand(initialize.InitCommand(applicationDiff()))
			item.AddCommand(initialize.InitCommand(applicationHistory()))
		}
		if item.Use == "appset" {
			item.AddCommand(initialize.InitCommand(appsetGenerate()))
		}
	}
	return argo
}

func applicationDryRun() *cobra.Command {
	c := &cobra.Command{
		Use: "dry-run",
		Short: color.YellowString("Extender feature. Dry-run specify application with revisions, " +
			"or dry-run with application-manifest which not create yet."),
		Run: func(cmd *cobra.Command, args []string) {},
	}
	return c
}

func applicationDiff() *cobra.Command {
	c := &cobra.Command{
		Use:   "diff-revision",
		Short: color.YellowString("Extender feature. Diff the specified revision with current live state"),
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	return c
}

func applicationHistory() *cobra.Command {
	c := &cobra.Command{
		Use:   "history-manifests",
		Short: color.YellowString("Extender feature. Print the history manifests"),
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	return c
}

func appsetGenerate() *cobra.Command {
	c := &cobra.Command{
		Use:   "geerate",
		Short: color.YellowString("Extender feature. Generate the applcationset before create with dry-run"),
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	return c
}
