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
	"encoding/json"
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd"
	"github.com/argoproj/argo-cd/v2/cmd/argocd/commands"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	argocdcmd "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/pkg/argocd"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/pkg/utils"
)

// NewArgoCmd create the argo command
func NewArgoCmd() *cobra.Command {
	argo := commands.NewCommand()
	argo.Short = "Controls a Argo CD server"
	for _, item := range argo.Commands() {
		if item.Use == "app" {
			item.AddCommand(applicationDryRun())
			item.AddCommand(applicationDiff())
			item.AddCommand(applicationHistory())
			for _, c := range item.Commands() {
				if c.Name() != "sync" {
					continue
				}
				changeApplicationSync(c)
			}
		}
		if item.Use == "appset" {
			item.AddCommand(appsetGenerate())
		}
	}
	return argo
}

type dryRunOrDiffRequest struct {
	Name            string    `json:"name"`
	Revision        string    `json:"revision"`
	Revisions       *[]string `json:"revisions"`
	AppManifestPath string    `json:"appManifestPath"`
	ShowDetails     bool      `json:"showDetails"`
}

var (
	dryRunReq = &dryRunOrDiffRequest{}
)

// applicationDryRun dry-run application with manifests or specify revision(s)
// nolint
func applicationDryRun() *cobra.Command {
	c := &cobra.Command{
		Use: "dry-run APPNAME",
		Short: color.YellowString("Extender feature. Dry-run specify application with revisions, " +
			"or dry-run with application-manifest which not create yet."),
		Example: `  # Dry-Run application without set revision(s)
  powerapp argocd app dry-run bcs-test

  # Dry-Run application with revision(s)
  powerapp argocd app dry-run bcs-test --revision ee1afa42

  # Dry-Run application with application manifest file
  powerapp argocd app dry-run --manifests ./testapp.yaml`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 0 {
				dryRunReq.Name = args[0]
			}
			if dryRunReq.Name == "" && dryRunReq.AppManifestPath == "" {
				utils.ExitError("'--appname' or '--manifests' must specify one")
			}
			req := &argocd.ApplicationDryRunRequest{
				Revision: dryRunReq.Revision,
			}
			if dryRunReq.Revisions != nil {
				req.Revisions = *dryRunReq.Revisions
			}
			if dryRunReq.Name != "" {
				req.ApplicationName = dryRunReq.Name
			}
			if dryRunReq.AppManifestPath != "" {
				bs, err := os.ReadFile(dryRunReq.AppManifestPath)
				if err != nil {
					utils.ExitError(fmt.Sprintf("read manifest file '%s' failed: %s",
						dryRunReq.AppManifestPath, err.Error()))
				}
				bs = utils.CheckStringJsonOrYaml(bs)
				req.ApplicationManifests = string(bs)
			}
			h := argocdcmd.NewHandler()
			h.DryRun(cmd.Context(), req, dryRunReq.ShowDetails)
		},
	}
	c.Flags().StringVar(&dryRunReq.Revision, "revision", "", "Specify the revision of application")
	dryRunReq.Revisions = c.Flags().StringSlice("revisions", nil, "Specify the revisions of application")
	c.Flags().StringVar(&dryRunReq.AppManifestPath, "manifests", "", "File path of application manifests")
	c.Flags().BoolVar(&dryRunReq.ShowDetails, "show-details", false, "Show the dry-run details")
	return c
}

var (
	diffReq            = &dryRunOrDiffRequest{}
	diffOnlyShowTarget bool
	diffOnlyShowLive   bool
)

// applicationDiff diff the application live state with target revision
// nolint
func applicationDiff() *cobra.Command {
	c := &cobra.Command{
		Use:   "diff-revision APPNAME",
		Short: color.YellowString("Extender feature. Diff the specified revision with current live state"),
		Example: `  ## Diff target revision with live state
  powerapp argocd app diff-revision bcs-test --revision 6e44efaf --show-details

  ## Diff target revision and only show target manifests
  powerapp argocd app diff-revision bcs-test --revision 6e44efaf --show-target
`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 0 {
				diffReq.Name = args[0]
			} else {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			req := &argocd.ApplicationDiffRequest{
				ApplicationName: diffReq.Name,
				Revision:        diffReq.Revision,
			}
			if diffReq.Revisions != nil {
				req.Revisions = *diffReq.Revisions
			}
			if req.ApplicationName == "" {
				utils.ExitError("'--appname' need specify")
			}
			h := argocdcmd.NewHandler()
			h.Diff(cmd.Context(), req, diffReq.ShowDetails, diffOnlyShowTarget, diffOnlyShowLive)
		},
	}
	c.Flags().StringVar(&diffReq.Name, "appname", "", "Specify the application name")
	c.Flags().StringVar(&diffReq.Revision, "revision", "", "Specify the revision of application")
	diffReq.Revisions = c.Flags().StringSlice("revisions", nil, "Specify the revisions of application")
	c.Flags().BoolVar(&diffReq.ShowDetails, "show-details", false, "Show the diff details between target revision and live state")
	c.Flags().BoolVar(&diffOnlyShowTarget, "show-target", false, "Only show the target revision manifests")
	c.Flags().BoolVar(&diffOnlyShowLive, "show-live", false, "Only show the live state manifests")
	return c
}

func applicationHistory() *cobra.Command {
	c := &cobra.Command{
		Use:   "history-manifests",
		Short: color.YellowString("Extender feature. Print the history manifests"),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("not implement now")
			os.Exit(1)
		},
	}
	return c
}

var (
	appSetPath        string
	appSetShowDetails bool
)

// appsetGenerate generate applications by applicationset's manifest
// nolint
func appsetGenerate() *cobra.Command {
	c := &cobra.Command{
		Use:   "generate",
		Short: color.YellowString("Extender feature. Generate the applcationset before create with dry-run"),
		Run: func(cmd *cobra.Command, args []string) {
			bs, err := os.ReadFile(appSetPath)
			if err != nil {
				utils.ExitError(fmt.Sprintf("read file '%s' failed: %s", appSetPath, err.Error()))
			}
			jsonBS := utils.CheckStringJsonOrYaml(bs)
			appset := new(v1alpha1.ApplicationSet)
			if err = json.Unmarshal(jsonBS, appset); err != nil {
				utils.ExitError(fmt.Sprintf("unmarshal to argocd applicationset failed: %s", err.Error()))
			}
			h := argocdcmd.NewHandler()
			h.AppSetGenerate(cmd.Context(), appset, appSetShowDetails)
		},
	}
	c.Flags().StringVar(&appSetPath, "manifests", "", "File path of application-set manifest")
	c.Flags().BoolVar(&appSetShowDetails, "show-details", false, "Show the diff details between target revision and live state")
	return c
}
