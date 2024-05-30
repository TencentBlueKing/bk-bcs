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

package edit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	klog "k8s.io/klog/v2"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/cmd/util/editor"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/cmd/printer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/pkg"
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

func editClustersNamespace() *cobra.Command { // nolint
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
				klog.Infoln("Project code (English abbreviation), global unique, " +
					"the length cannot exceed 64 characters")
				return
			}
			client := pkg.NewClientWithConfiguration(context.Background())
			// 获取集群下所有命名空间
			namespaceResp, err := client.ListNamespaces(&pkg.ListNamespacesRequest{
				ProjectCode: projectCode,
				ClusterID:   clusterID,
			})
			if err != nil {
				klog.Infof("list variable definitions failed: %v", err)
				return
			}

			// 查找返回编辑的命名
			data, err := editData(projectCode, namespaceResp)
			if err != nil {
				klog.Infoln(err)
				return
			}

			// 原内容
			marshal, err := json.Marshal(data)
			if err != nil {
				klog.Infof("[namespace] deserialize failed: %v", err)
				return
			}
			// 把json转成yaml
			original, err := yaml.JSONToYAML(marshal)
			if err != nil {
				klog.Infof("json to yaml failed: %v", err)
				return
			}
			edit := editor.NewDefaultEditor([]string{})
			// 编辑后的
			edited, path, err := edit.LaunchTempFile(fmt.Sprintf("%s-edit-", filepath.Base(os.Args[0])),
				".yaml", bytes.NewBufferString(string(original)))
			if err != nil {
				klog.Infof("unexpected error: %v", err)
				return
			}
			if _, err = os.Stat(path); err != nil {
				klog.Infof("no temp file: %s", path)
				return
			}
			// 对比原内容是否更改
			if bytes.Equal(cmdutil.StripComments(original), cmdutil.StripComments(edited)) {
				klog.Infoln("Edit canceled, no valid changes were saved.")
				return
			}

			// 对比修改前后的数据
			editAfter, err := contrast(marshal, edited)
			if err != nil {
				klog.Infoln(err)
				return
			}

			// 需要提交编辑的数据
			updateData := &pkg.UpdateNamespaceRequest{
				ProjectCode: projectCode,
				ClusterID:   clusterID,
				Name:        name,
				Quota:       editAfter.Quota,
				Variables:   editAfter.Variables,
			}

			resp, err := client.UpdateNamespace(updateData, projectCode, clusterID, name)
			if err != nil {
				klog.Infof("update project failed: %v", err)
				return
			}
			printer.PrintInJSON(resp)
		},
	}

	cmd.Flags().StringVarP(&clusterID, "cluster-id", "", "",
		"cluster ID, required")
	cmd.Flags().StringVarP(&name, "name", "", "",
		"Namespace name, length cannot exceed 63 characters, can only contain lowercase letters, numbers, "+
			"and '-', must start with a letter and cannot end with '-'")

	return cmd
}

func editData(projectCode string, namespaceData *bcsproject.ListNamespacesResponse) (
	*pkg.UpdateNamespaceRequest, error) {
	// 从列表通过名称查找需要编辑的命名空间
	namespaceList := make(map[string]*bcsproject.NamespaceData, 0)
	for _, item := range namespaceData.Data {
		namespaceList[item.Name] = &bcsproject.NamespaceData{
			Name:             item.GetName(),
			Status:           item.GetStatus(),
			CreateTime:       item.GetCreateTime(),
			Quota:            item.GetQuota(),
			Used:             item.GetUsed(),
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
		err := fmt.Errorf("no namespace with that name found: %v", name)
		return nil, err
	}

	// 处理变量variable和Quota值为空时显示 [] {}
	variableValue := make([]pkg.VariableValue, 0)
	if len(namespace.Variables) != 0 {
		for _, item := range namespace.Variables {
			variableValue = append(variableValue, pkg.VariableValue{
				Id:          item.Id,
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

	// 需要用编辑器打开的数据
	updateNamespace := &pkg.UpdateNamespaceRequest{
		ProjectCode: projectCode,
		ClusterID:   clusterID,
		Name:        namespace.Name,
		Quota:       quotaVal,
		Variables:   variableValue,
	}
	return updateNamespace, nil
}

func contrast(original, edited []byte) (*pkg.UpdateNamespaceRequest, error) {

	// 把编辑后的内容yaml转成json
	editedJson, err := yaml.YAMLToJSON(edited)
	if err != nil {
		err = fmt.Errorf("json to yaml failed: %v", name)
		return nil, err
	}

	var (
		editBefore *pkg.UpdateNamespaceRequest
		editAfter  *pkg.UpdateNamespaceRequest
	)

	// 生成编辑前数据和编辑后数据做对比
	{
		err = json.Unmarshal(editedJson, &editAfter)
		if err != nil {
			err = fmt.Errorf("[edit after] deserialize failed: %v", name)
			return nil, err
		}

		err = json.Unmarshal(original, &editBefore)
		if err != nil {
			err = fmt.Errorf("[edit before] deserialize failed: %v", name)
			return nil, err
		}

		// 一定要放在在赋值前
		// 获取编辑前Variables值 除去能编辑的字段
		variablesBefore := make([]pkg.Variable, 0)
		if len(editBefore.Variables) != 0 {
			for _, item := range editBefore.Variables {
				variablesBefore = append(variablesBefore, pkg.Variable{
					ID:          item.Id,
					Key:         item.Key,
					Name:        item.Name,
					ClusterID:   item.ClusterID,
					ClusterName: item.ClusterName,
					Namespace:   item.Namespace,
					Scope:       item.Scope,
				})
			}
		}

		// 获取编辑后Variables值 除去能编辑的字段
		variablesAfter := make([]pkg.Variable, 0)
		if len(editAfter.Variables) != 0 {
			for _, item := range editAfter.Variables {
				variablesAfter = append(variablesAfter, pkg.Variable{
					ID:          item.Id,
					Key:         item.Key,
					Name:        item.Name,
					ClusterID:   item.ClusterID,
					ClusterName: item.ClusterName,
					Namespace:   item.Namespace,
					Scope:       item.Scope,
				})
			}
		}

		// 对比前后数据是否一直(已经过滤掉能编辑的值)
		if !reflect.DeepEqual(variablesBefore, variablesAfter) {
			err = fmt.Errorf("variables can only modify values")
			return nil, err
		}

		// 把能更改的值赋值到编辑前数据
		editBefore.Quota = editAfter.Quota
		editBefore.Variables = editAfter.Variables
	}

	// 对比整个前后数据是否一直
	if !reflect.DeepEqual(editBefore, editAfter) {
		err = fmt.Errorf("only edit desc and default value")
		return nil, err
	}
	return editBefore, nil
}
