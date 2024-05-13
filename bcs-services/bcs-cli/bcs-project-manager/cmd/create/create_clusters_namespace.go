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

package create

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	klog "k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/cmd/util/editor"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"sigs.k8s.io/yaml"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/cmd/printer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/pkg"
	GenerateReq "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/pkg/create"
)

var (
	clusterID                   string
	createClustersNamespaceLong = templates.LongDesc(i18n.T(`
		Create a project namespace using the specified file or standard input.`))

	createClustersNamespaceExample = templates.Examples(i18n.T(`
		# Create a project namespace with a project code and clusterID
		kubectl-bcs-project-manager create namespace --cluster-id=clusterID
		# Create a project namespace with a file
		kubectl-bcs-project-manager create namespace --filename=file-address`))
)

func createClustersNamespace() *cobra.Command { // nolint
	cmd := &cobra.Command{
		Use:                   "namespace [--cluster-id=clusterID | --filename=file-address]",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"n"},
		Short: i18n.T("Create a project namespace using the " +
			"specified file or standard input"),
		Long:    createClustersNamespaceLong,
		Example: createClustersNamespaceExample,
		Run: func(cmd *cobra.Command, args []string) {
			projectCode := viper.GetString("bcs.project_code")
			if len(projectCode) == 0 {
				klog.Infoln("Project code (English abbreviation), " +
					"global unique, the length cannot exceed 64 characters")
				return
			}
			client := pkg.NewClientWithConfiguration(context.Background())
			// 获取当前集群信息
			cluster, err := client.GetCluster(&pkg.GetClusterRequest{ClusterID: clusterID})
			if err != nil {
				klog.Infoln(err)
				return
			}

			var quota pkg.Quota
			// 查询是否是共享集群 true是 false否
			if cluster.Data.IsShared {
				quota = pkg.Quota{
					CPURequests:    "1",
					MemoryRequests: "1Gi",
					CPULimits:      "1",
					MemoryLimits:   "1Gi",
				}
			}

			parameters := &pkg.CreateNamespaceRequest{}
			parameters.ClusterID = clusterID
			parameters.ProjectCode = projectCode
			parameters.Name = ""
			parameters.Quota = struct {
				CPURequests    string `json:"cpuRequests"`
				MemoryRequests string `json:"memoryRequests"`
				CPULimits      string `json:"cpuLimits"`
				MemoryLimits   string `json:"memoryLimits"`
			}(quota)
			{
				var (
					requestParam interface{}
					marshal      []byte
					created      []byte
					createdJson  []byte
					path         string
				)
				// 判断是否有文件路径
				if filename != "" {
					requestParam, err = GenerateReq.GenerateStruct(filename)
					if err != nil {
						klog.Infoln(err)
						return
					}
					marshal, err = json.Marshal(requestParam)
					if err != nil {
						klog.Infof("[requestParam] deserialize failed: %v", err)
						return
					}
					err = json.Unmarshal(marshal, &parameters)
					if err != nil {
						klog.Infof("[parameters] deserialize failed: %v", err)
						return
					}
					// 判断环境变量Project Code和文件里面是否一致
					if projectCode != parameters.ProjectCode {
						klog.Infoln("The environment variable project code is " +
							"inconsistent with the file project code")
						return
					}
					clusterID = parameters.ClusterID
				} else {
					// 处理模板数据
					marshal, err = json.Marshal(parameters)
					if err != nil {
						klog.Infof("[CreateNamespaceRequest] deserialize failed: %v", err)
						return
					}
					// 把json转成yaml
					original, formatErr := yaml.JSONToYAML(marshal)
					if err != nil {
						klog.Infof("json to yaml failed: %v", formatErr)
						return
					}
					create := editor.NewDefaultEditor([]string{})

					created, path, err = create.LaunchTempFile(fmt.Sprintf("%s-create-", filepath.Base(os.Args[0])),
						".yaml", bytes.NewBufferString(string(original)))

					if err != nil {
						klog.Infof("unexpected error: %v", err)
						return
					}
					if _, err = os.Stat(path); err != nil {
						klog.Infof("no temp file: %s", path)
						return
					}
					// 把创建的内容yaml转成json
					createdJson, err = yaml.YAMLToJSON(created)
					if err != nil {
						klog.Infof("json to yaml failed: %v", err)
						return
					}
					err = json.Unmarshal(createdJson, &parameters)
					if err != nil {
						klog.Infof("[YAMLToJSON] deserialize failed: %v", err)
						return
					}
				}
			}
			if parameters.Name == "" {
				klog.Infoln("Namespace name is required, The length cannot exceed 63 characters, " +
					"can only contain lowercase letters, numbers, " +
					"and '-', must start with a letter, and cannot end with '-'")
				return
			}
			resp, err := client.CreateNamespace(parameters, projectCode, clusterID)
			if err != nil {
				klog.Infof("create namespace failed: %v", err)
				return
			}
			printer.PrintInJSON(resp)
		},
	}
	cmd.Flags().StringVarP(&clusterID, "cluster-id", "", "",
		"cluster ID, If you specify to create a file, cluster id you can leave it blank, "+
			"otherwise it is required")

	return cmd
}
