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
	"sort"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd"
	kubectlcmd "k8s.io/kubectl/pkg/cmd"
	"k8s.io/kubectl/pkg/cmd/plugin"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/internal/clusterset"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/internal/selectui"
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
	kc.AddCommand(projectCommand())
	kc.AddCommand(clusterCommand())
	kc.AddCommand(setCommand())
	kc.AddCommand(setGlobalCommand())
	kc.AddCommand(infoCommand())
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

func projectCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "projs",
		Short: color.YellowString("Extender feature. List all projects which user have view permission"),
		Run: func(cmd *cobra.Command, args []string) {
			resultProjects := make([]*argoProject, 0)
			projects := listProjects(cmd.Context())
			for i := range projects.Items {
				proj := projects.Items[i]
				resultProjects = append(resultProjects, &argoProject{
					Name:       proj.Name,
					AliaName:   proj.Annotations[common.ProjectAliaName],
					BusinessID: proj.Annotations[common.ProjectBusinessIDKey],
				})
			}

			tw := utils.DefaultTableWriter()
			tw.SetHeader(func() []string {
				return []string{
					"项目名称", "项目Code", "业务ID",
				}
			}())
			for _, pj := range resultProjects {
				tw.Append(func() []string {
					return []string{
						pj.AliaName, pj.Name, pj.BusinessID,
					}
				}())
			}
			tw.Render()
		},
	}
	return c
}

func clusterCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "clusters",
		Short: color.YellowString("Extender feature. List projects' clusters which user have view permission"),
		Run: func(cmd *cobra.Command, args []string) {
			clusters := listClusters(cmd.Context(), "")
			projClusters := make(map[string][]*v1alpha1.Cluster)
			for i := range clusters.Items {
				cls := clusters.Items[i]
				projClusters[cls.Project] = append(projClusters[cls.Project], &cls)
			}

			tw := utils.DefaultTableWriter()
			tw.SetHeader(func() []string {
				return []string{
					"集群ID", "连接状态", "集群名称", "项目名称",
				}
			}())
			projects := listProjects(cmd.Context())
			for i := range projects.Items {
				proj := projects.Items[i]
				projAliaName := proj.Annotations[common.ProjectAliaName]
				clses, ok := projClusters[proj.Name]
				if !ok {
					continue
				}
				for _, cls := range clses {
					var status string
					if cls.Info.ConnectionState.Status == "Failed" {
						status = "Failed"
					} else {
						status = "Success"
					}
					tw.Append(func() []string {
						return []string{
							cls.Name, status, cls.Annotations[common.ClusterAliaName], projAliaName + "/" + proj.Name,
						}
					}())
				}
			}
			tw.Render()
		},
	}
	return c
}

var switchCluster bool

// setCluster set cluster common
func setCluster(cmd *cobra.Command, args []string) *clusterset.ClusterInfo {
	var clsInfo *clusterset.ClusterInfo
	if !switchCluster {
		var clusterID string
		if len(args) != 0 {
			clusterID = args[0]
		}
		if clusterID == "" {
			utils.ExitError("must set cluster-id or use '--switch' param")
		}
		clsList := listClusters(cmd.Context(), "")
		var existCls *v1alpha1.Cluster
		for i := range clsList.Items {
			cls := clsList.Items[i]
			if cls.Name == clusterID {
				existCls = &cls
				break
			}
		}
		clsInfo = &clusterset.ClusterInfo{ClusterID: clusterID}
		if existCls != nil {
			clsInfo.Project = existCls.Project
			clsInfo.ClusterName = existCls.Annotations[common.ClusterAliaName]
		}
	} else {
		clsList := listClusters(cmd.Context(), "")
		selectClusters := make([]selectui.Needle, 0, len(clsList.Items))
		for i := range clsList.Items {
			cls := clsList.Items[i]
			selectClusters = append(selectClusters, selectui.Needle{
				Name:      cls.Annotations[common.ClusterAliaName],
				Project:   cls.Project,
				ClusterID: cls.Name,
			})
		}
		sort.Slice(selectClusters, func(i, j int) bool {
			if selectClusters[i].Project == selectClusters[j].Project {
				return selectClusters[i].ClusterID < selectClusters[j].ClusterID
			}
			return selectClusters[i].Project < selectClusters[j].Project
		})
		index, err := selectui.SelectUI(selectClusters, "Select BCS Cluster:")
		if err != nil {
			utils.ExitError(fmt.Sprintf("select ui failed: %s", err.Error()))
		}
		item := selectClusters[index]
		clsInfo = &clusterset.ClusterInfo{
			ClusterID:   item.ClusterID,
			ClusterName: item.Name,
			Project:     item.Project,
		}
	}
	return clsInfo
}

// setCommand set cluster for current session
func setCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "setc",
		Short: color.YellowString("Extender feature. Set cluster-id for current session"),
		Run: func(cmd *cobra.Command, args []string) {
			clsInfo := setCluster(cmd, args)
			setter := &clusterset.ClusterSetter{}
			if err := setter.SetCluster(clsInfo); err != nil {
				utils.ExitError(fmt.Sprintf("set cluster for current session failed: %s", err.Error()))
			}
			fmt.Printf("cluster '%s' is set for current-session\n", clsInfo.ClusterID)
		},
	}
	c.PersistentFlags().BoolVar(&switchCluster, "switch", false, "switch cluster with bash ui")
	return c
}

// setGlobalCommand set cluster for global context
func setGlobalCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "setg",
		Short: color.YellowString("Extender feature. Set cluster-id for global context"),
		Run: func(cmd *cobra.Command, args []string) {
			clsInfo := setCluster(cmd, args)
			setter := &clusterset.ClusterSetter{}
			if err := setter.SetClusterGlobal(clsInfo); err != nil {
				utils.ExitError(fmt.Sprintf("set cluster for current session failed: %s", err.Error()))
			}
			fmt.Printf("cluster '%s' is set for global context\n", clsInfo.ClusterID)
		},
	}
	c.PersistentFlags().BoolVar(&switchCluster, "switch", false, "switch cluster with bash ui")
	return c
}

// infoCommand return cluster info
func infoCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "info",
		Short: color.YellowString("Extender feature. Show the cluster set"),
		Run: func(cmd *cobra.Command, args []string) {
			setter := &clusterset.ClusterSetter{}
			clusters, err := setter.ReturnClusterInfo()
			if err != nil {
				utils.ExitError(fmt.Sprintf("get cluster-info failed: %s", err.Error()))
			}
			tw := utils.DefaultTableWriter()
			for i, cls := range clusters {
				if i == 0 {
					tw.Append(func() []string {
						return []string{"*", color.YellowString(cls.Status), color.YellowString(cls.ClusterID),
							color.YellowString(cls.ClusterName), color.YellowString(cls.Project)}
					}())
				} else {
					tw.Append(func() []string {
						return []string{"", cls.Status, cls.ClusterID, cls.ClusterName, cls.Project}
					}())
				}
			}
			tw.Render()
		},
	}
	return c
}

type argoProject struct {
	Name       string
	AliaName   string
	BusinessID string
}

// listProjects list projects
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

// listClusters list clusters
func listClusters(ctx context.Context, proj string) *v1alpha1.ClusterList {
	req := &httputils.HTTPRequest{
		Path:   "/api/v1/clusters",
		Method: http.MethodGet,
	}
	if proj != "" {
		req.QueryParams = map[string][]string{
			"projects": {proj},
		}
	}
	body := httputils.DoRequest(ctx, req)
	clusters := new(v1alpha1.ClusterList)
	if err := json.Unmarshal(body, clusters); err != nil {
		utils.ExitError(fmt.Sprintf("unmarshal clusters failed: %s", err.Error()))
	}
	return clusters
}
