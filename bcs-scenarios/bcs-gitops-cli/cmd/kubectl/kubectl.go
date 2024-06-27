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

// Package kubectl xx
package kubectl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd"
	kubectlcmd "k8s.io/kubectl/pkg/cmd"
	"k8s.io/kubectl/pkg/cmd/plugin"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/pkg/clusterset"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/pkg/httputils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/pkg/utils"
)

// KubectlCommand defines the kubectl command
type KubectlCommand struct {
	command *cobra.Command
	configs *genericclioptions.ConfigFlags
}

// NewKubectlCmd create kubectl command instance
func NewKubectlCmd() *KubectlCommand {
	defaultConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag().
		WithDiscoveryBurst(300).WithDiscoveryQPS(50.0)

	kc := cmd.NewKubectlCommand(kubectlcmd.KubectlOptions{
		PluginHandler: kubectlcmd.NewDefaultPluginHandler(plugin.ValidPluginFilenamePrefixes),
		Arguments:     os.Args,
		ConfigFlags:   defaultConfigFlags,
		IOStreams:     genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
	})
	kc.Use = "k"
	kc.AddCommand(clusterCommand())
	return &KubectlCommand{
		command: kc,
		configs: defaultConfigFlags,
	}
}

// GetCommand return the command
func (c *KubectlCommand) GetCommand() *cobra.Command {
	return c.command
}

// GetConfigs return the confis
func (c *KubectlCommand) GetConfigs() *genericclioptions.ConfigFlags {
	return c.configs
}

func clusterCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "cluster",
		Short: color.YellowString("Extender feature. Managed the cluster context"),
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	c.AddCommand(clusterList())
	c.AddCommand(clusterSet())
	c.AddCommand(clusterInfo())
	return c
}

type argoProject struct {
	Name       string
	AliaName   string
	BusinessID string
}

type argoCluster struct {
	Name     string
	AliaName string
	Status   string
}

func clusterList() *cobra.Command {
	c := &cobra.Command{
		Use:   "list",
		Short: "Show the clusters that managed in gitops which user have permissions",
		Run: func(cmd *cobra.Command, args []string) {
			resultProjects := make([]*argoProject, 0)
			resultClusters := make(map[string][]*argoCluster)
			projects := listProjects(cmd.Context())
			for i := range projects.Items {
				proj := projects.Items[i]
				resultProjects = append(resultProjects, &argoProject{
					Name:       proj.Name,
					AliaName:   proj.Annotations[common.ProjectAliaName],
					BusinessID: proj.Annotations[common.ProjectBusinessIDKey],
				})

				clusters := listClusters(cmd.Context(), proj.Name)
				for j := range clusters.Items {
					cluster := clusters.Items[j]
					argoCls := &argoCluster{
						Name:     cluster.Name,
						AliaName: cluster.Annotations[common.ClusterAliaName],
					}
					if cluster.Info.ConnectionState.Status == "Failed" {
						argoCls.Status = "Failed"
					} else {
						argoCls.Status = "Success"
					}
					resultClusters[proj.Name] = append(resultClusters[proj.Name], argoCls)
				}
			}

			tw := utils.DefaultTableWriter()
			tw.SetHeader(func() []string {
				return []string{
					"项目名称", "项目Code", "业务ID", "集群ID", "连接状态", "集群名称",
				}
			}())
			for _, pj := range resultProjects {
				v, ok := resultClusters[pj.Name]
				if !ok || len(v) == 0 {
					continue
				}
				for ans := 0; ans < len(v); ans++ {
					tw.Append(func() []string {
						return []string{
							pj.AliaName, pj.Name, pj.BusinessID,
							v[ans].Name, v[ans].Status, v[ans].AliaName,
						}
					}())
				}
			}
			tw.Render()
		},
	}
	return c
}

func clusterSet() *cobra.Command {
	c := &cobra.Command{
		Use:   "set BCS-CLUSTER-ID",
		Short: `Set the cluster for kubectl(Use "export CLUSTER=BCS-K8S-12345 for current shell session)"`,
		Run: func(cmd *cobra.Command, args []string) {
			var clusterID string
			if len(args) != 0 {
				clusterID = args[0]
			}
			if clusterID == "" {
				utils.ExitError("should set CLUSTER-ID")
			}
			setter := clusterset.Setter{}
			if err := setter.SetCluster(clusterID); err != nil {
				utils.ExitError(fmt.Sprintf("set global cluster '%s' failed: %s", clusterID, err.Error()))
			}
		},
	}
	return c
}

func clusterInfo() *cobra.Command {
	c := &cobra.Command{
		Use:   "info",
		Short: "Return the cluster had set",
		Run: func(cmd *cobra.Command, args []string) {
			setter := clusterset.Setter{}
			clusterID, err := setter.GetCurrentCluster()
			if err != nil {
				utils.ExitError(fmt.Sprintf("get current cluster failed: %s", err.Error()))
			}
			fmt.Println(clusterID)
		},
	}
	return c
}

func listProjects(ctx context.Context) *v1alpha1.AppProjectList {
	projects := new(v1alpha1.AppProjectList)
	body := httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   "/api/v1/projects",
		Method: http.MethodGet,
	})
	if err := json.Unmarshal(body, projects); err != nil {
		utils.ExitError(fmt.Sprintf("unmarshal projects failed: %s", err.Error()))
	}
	return projects
}

func listClusters(ctx context.Context, proj string) *v1alpha1.ClusterList {
	clusters := new(v1alpha1.ClusterList)
	body := httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   "/api/v1/clusters",
		Method: http.MethodGet,
		QueryParams: map[string][]string{
			"projects": {proj},
		},
	})
	if err := json.Unmarshal(body, clusters); err != nil {
		utils.ExitError(fmt.Sprintf("unmarshal clusters failed: %s", err.Error()))
	}
	return clusters
}
