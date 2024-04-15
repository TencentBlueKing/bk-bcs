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

// Package terraform defines the terraform command
package terraform

import (
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/apis/httpapi"
	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/pkg/terraform"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/pkg/utils"
)

// NewTerraformCmd create the terraform cmd instance
func NewTerraformCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "terraform",
		Short: "Operate the terraform resources",
	}
	command.AddCommand(apply())
	command.AddCommand(list())
	command.AddCommand(create())
	command.AddCommand(get())
	command.AddCommand(del())
	command.AddCommand(getDiff())
	command.AddCommand(getApply())
	command.AddCommand(sync())
	return command
}

var (
	applyFile string
)

func apply() *cobra.Command {
	c := &cobra.Command{
		Use:   "apply",
		Short: "Apply the terraform resource configuration",
		Run: func(cmd *cobra.Command, args []string) {
			blog.SetV(int32(options.LogV))
			h := terraform.NewHandler()
			bs, err := os.ReadFile(applyFile)
			if err != nil {
				utils.ExitError(fmt.Sprintf("read file '%s' failed: %s", applyFile, err.Error()))
			}
			h.Apply(cmd.Context(), bs)
		},
	}
	c.PersistentFlags().StringVarP(&applyFile, "file", "f", "",
		"The files that contain the configurations to apply")
	return c
}

var (
	listProjects = &([]string{})
)

func list() *cobra.Command {
	c := &cobra.Command{
		Use:   "list",
		Short: "List terraforms",
		Run: func(cmd *cobra.Command, args []string) {
			blog.SetV(int32(options.LogV))
			h := terraform.NewHandler()
			h.List(cmd.Context(), listProjects)
		},
	}
	listProjects = c.Flags().StringSliceP("projects", "p", nil, "Filter by project name")
	return c
}

var (
	createReq = &httpapi.TerraformCreateRequest{}
)

func create() *cobra.Command {
	c := &cobra.Command{
		Use:   "create",
		Short: "Create terraform by command line params (User also can use apply to replace this command)",
		Run: func(cmd *cobra.Command, args []string) {
			blog.SetV(int32(options.LogV))
			h := terraform.NewHandler()
			h.Create(cmd.Context(), createReq)
		},
	}
	c.Flags().StringVar(&createReq.Name, "name", "", "Resource name(must start with project name)")
	c.Flags().BoolVar(&createReq.Destroy, "destroy", false, "Whether to destroy resources when delete")
	c.Flags().StringVar(&createReq.Project, "project", "", "Terraform project name")
	c.Flags().StringVar(&createReq.Repo, "repo", "", "Repository URL(argocd repo)")
	c.Flags().StringVar(&createReq.Path, "path", "", "Path in repository to the terraform directory")
	c.Flags().StringVar(&createReq.Revision, "revision", "",
		"The tracking source branch, tag, commit the terraform will sync to")
	c.Flags().StringVar(&createReq.SyncPolicy, "sync-policy", "",
		"Set the sync policy (one of: manual, auto-sync)")
	return c
}

var (
	getName   string
	getOutput string
)

func del() *cobra.Command {
	c := &cobra.Command{
		Use:   "delete",
		Short: "Delete resource by name",
		Run: func(cmd *cobra.Command, args []string) {
			blog.SetV(int32(options.LogV))
			h := terraform.NewHandler()
			h.Delete(cmd.Context(), &getName)
		},
	}
	c.Flags().StringVar(&getName, "name", "", "Resource name")
	return c
}

func get() *cobra.Command {
	c := &cobra.Command{
		Use:   "get",
		Short: "Get terraform details",
		Run: func(cmd *cobra.Command, args []string) {
			blog.SetV(int32(options.LogV))
			h := terraform.NewHandler()
			h.Get(cmd.Context(), &getName, &getOutput)
		},
	}
	c.Flags().StringVar(&getName, "name", "", "Resource name")
	c.Flags().StringVarP(&getOutput, "output", "o", "", "Output format. One of: (json, yaml)")
	return c
}

func getDiff() *cobra.Command {
	c := &cobra.Command{
		Use:   "get-diff",
		Short: "Perform a diff against the target and live state (terraform plan result)",
		Run: func(cmd *cobra.Command, args []string) {
			blog.SetV(int32(options.LogV))
			h := terraform.NewHandler()
			h.GetDiff(cmd.Context(), &getName)
		},
	}
	c.Flags().StringVar(&getName, "name", "", "Resource name")
	return c
}

func getApply() *cobra.Command {
	c := &cobra.Command{
		Use:   "get-apply",
		Short: "Get the latest apply result of terraform",
		Run: func(cmd *cobra.Command, args []string) {
			blog.SetV(int32(options.LogV))
			h := terraform.NewHandler()
			h.GetApply(cmd.Context(), &getName)
		},
	}
	c.Flags().StringVar(&getName, "name", "", "Resource name")
	return c
}

func sync() *cobra.Command {
	c := &cobra.Command{
		Use:   "sync",
		Short: "Trigger the sync operation",
		Run: func(cmd *cobra.Command, args []string) {
			blog.SetV(int32(options.LogV))
			h := terraform.NewHandler()
			h.Sync(cmd.Context(), &getName)
		},
	}
	c.Flags().StringVar(&getName, "name", "", "Resource name")
	return c
}
