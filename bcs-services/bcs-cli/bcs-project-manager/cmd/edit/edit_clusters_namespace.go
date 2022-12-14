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

package edit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/cmd/util/editor"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/cmd/printer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/pkg"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

var (
	clusterID                 string
	name                      string
	editClustersNamespaceLong = templates.LongDesc(i18n.T(`
		Edit a project namespace from the default editor.`))

	editClustersNamespaceExample = templates.Examples(i18n.T(`
		# Edit project namespace by clusterID and name
		kubectl-bcs-project-manager edit namespace --cluster-id=clusterID --name=name`))
)

func editClustersNamespace() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "namespace --cluster-id=clusterID",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"n"},
		Short:                 i18n.T("Edit a project namespace from the default editor"),
		Long:                  editClustersNamespaceLong,
		Example:               editClustersNamespaceExample,
		Run: func(cmd *cobra.Command, args []string) {
			projectCode := viper.GetString("bcs.project_code")
			if len(projectCode) == 0 {
				klog.Infoln("Project code (English abbreviation), global unique, the length cannot exceed 64 characters")
				return
			}
			client := pkg.NewClientWithConfiguration(context.Background())
			// ?????????????????????????????????
			namespaceResp, err := client.ListNamespaces(&pkg.ListNamespacesRequest{
				ProjectCode: projectCode,
				ClusterID:   clusterID,
			})
			if err != nil {
				klog.Infoln("list variable definitions failed: %v", err)
				return
			}

			// ??????????????????????????????????????????????????????
			namespaceList := make(map[string]*bcsproject.NamespaceData, 0)
			for _, item := range namespaceResp.Data {
				namespaceList[item.Name] = &bcsproject.NamespaceData{
					Name:             item.GetName(),
					Status:           item.GetStatus(),
					CreateTime:       item.GetCreateTime(),
					Quota:            item.GetQuota(),
					Labels:           item.GetLabels(),
					Annotations:      item.GetAnnotations(),
					Variables:        item.GetVariables(),
					ItsmTicketSN:     item.GetItsmTicketSN(),
					ItsmTicketStatus: item.GetItsmTicketStatus(),
					ItsmTicketURL:    item.GetItsmTicketURL(),
					ItsmTicketType:   item.GetItsmTicketType(),
				}
			}
			namespace, ok := namespaceList[name]
			if !ok {
				klog.Infoln("No namespace with that name found: %v", name)
				return
			}

			// ????????????variable???Quota?????????????????? [] {}
			variableValue := make([]pkg.Data, 0)
			if len(namespace.Variables) != 0 {
				for _, item := range namespace.Variables {
					variableValue = append(variableValue, pkg.Data{
						ID:          item.Id,
						Key:         item.Key,
						Name:        item.Name,
						ClusterID:   item.ClusterID,
						ClusterName: item.ClusterName,
						Namespace:   item.Namespace,
						Value:       item.Value,
						Scope:       item.Scope,
					})
				}
			}
			quotaVal := pkg.Quota{}
			if namespace.Quota != nil {
				quotaVal = pkg.Quota{
					CPURequests:    namespace.Quota.CpuRequests,
					MemoryRequests: namespace.Quota.MemoryRequests,
					CPULimits:      namespace.Quota.CpuLimits,
					MemoryLimits:   namespace.Quota.MemoryLimits,
				}
			}

			updateNamespace := pkg.UpdateNamespaceTemplate{
				UpdateNamespaceRequest: pkg.UpdateNamespaceRequest{
					ProjectCode: projectCode,
					ClusterID:   clusterID,
					Name:        namespace.Name,
					Quota:       quotaVal,
				},
				Variable: variableValue,
			}

			// ?????????
			marshal, err := json.Marshal(updateNamespace)
			if err != nil {
				klog.Infoln("[namespace] deserialize failed: %v", err)
				return
			}
			// ???json??????yaml
			original, err := yaml.JSONToYAML(marshal)
			if err != nil {
				klog.Infoln("json to yaml failed: %v", err)
				return
			}
			edit := editor.NewDefaultEditor([]string{})
			// ????????????
			edited, path, err := edit.LaunchTempFile(fmt.Sprintf("%s-edit-", filepath.Base(os.Args[0])), ".yaml", bytes.NewBufferString(string(original)))
			if err != nil {
				klog.Infoln("unexpected error: %v", err)
				return
			}
			if _, err := os.Stat(path); err != nil {
				klog.Infoln("no temp file: %s", path)
				return
			}
			// ???????????????????????????
			if bytes.Equal(cmdutil.StripComments(original), cmdutil.StripComments(edited)) {
				klog.Infoln("Edit cancelled, no valid changes were saved.")
				return
			}
			// ?????????????????????yaml??????json
			editedJson, err := yaml.YAMLToJSON(edited)
			if err != nil {
				klog.Infoln("json to yaml failed: %v", err)
				return
			}

			var (
				editBefore pkg.UpdateNamespaceTemplate
				editAfter  pkg.UpdateNamespaceTemplate
			)

			// ????????????????????????????????????????????????
			{
				err = json.Unmarshal(editedJson, &editAfter)
				if err != nil {
					klog.Infoln("[edit after] deserialize failed: %v", err)
					return
				}

				err = json.Unmarshal(marshal, &editBefore)
				if err != nil {
					klog.Infoln("[edit before] deserialize failed: %v", err)
					return
				}

				// ???????????????????????????
				// ???????????????Variables??? ????????????????????????
				variablesBefore := make([]pkg.Variable, 0)
				if len(editBefore.Variable) != 0 {
					for _, item := range editBefore.Variable {
						variablesBefore = append(variablesBefore, pkg.Variable{
							ID:          item.ID,
							Key:         item.Key,
							Name:        item.Name,
							ClusterID:   item.ClusterID,
							ClusterName: item.ClusterName,
							Namespace:   item.Namespace,
							Scope:       item.Scope,
						})
					}
				}

				// ???????????????Variables??? ????????????????????????
				variablesAfter := make([]pkg.Variable, 0)
				if len(editAfter.Variable) != 0 {
					for _, item := range editAfter.Variable {
						variablesAfter = append(variablesAfter, pkg.Variable{
							ID:          item.ID,
							Key:         item.Key,
							Name:        item.Name,
							ClusterID:   item.ClusterID,
							ClusterName: item.ClusterName,
							Namespace:   item.Namespace,
							Scope:       item.Scope,
						})
					}
				}

				// ??????????????????????????????(??????????????????????????????)
				if !reflect.DeepEqual(variablesBefore, variablesAfter) {
					klog.Infoln("Variables can only modify values")
					return
				}

				// ??????????????????????????????????????????
				editBefore.Quota = editAfter.Quota
				editBefore.Variable = editAfter.Variable
			}

			// ????????????????????????????????????
			if !reflect.DeepEqual(editBefore, editAfter) {
				klog.Infoln("only edit desc and default value")
				return
			}

			// ????????????????????????
			updateNamespaceData := &pkg.UpdateNamespaceRequest{
				ProjectCode: projectCode,
				ClusterID:   clusterID,
				Name:        name,
				Quota:       editAfter.Quota,
			}

			resp, err := client.UpdateNamespace(updateNamespaceData, projectCode, clusterID, name)
			if err != nil {
				klog.Infoln("update namespace failed: %v", err)
				return
			}

			// ??????????????????????????????
			updateNamespaceVariableData := &pkg.UpdateNamespaceVariablesReq{
				ProjectCode: projectCode,
				ClusterID:   clusterID,
				Namespace:   name,
				Data:        editAfter.Variable,
			}
			res, err := client.UpdateNamespaceVariables(updateNamespaceVariableData)
			if err != nil {
				klog.Infoln("update namespace variables failed: %v", err)
				return
			}
			if res.Code != 0 && resp.Code != 0 {
				klog.Infoln("Failed to update the namespace and variables: %v", err)
				return
			}
			printer.PrintInJSON(nil)
		},
	}

	cmd.Flags().StringVarP(&clusterID, "cluster-id", "", "",
		"cluster ID, required")
	cmd.Flags().StringVarP(&name, "name", "", "",
		"Namespace name, length cannot exceed 63 characters, can only contain lowercase letters, numbers, and '-', must start with a letter and cannot end with '-'")

	return cmd
}
