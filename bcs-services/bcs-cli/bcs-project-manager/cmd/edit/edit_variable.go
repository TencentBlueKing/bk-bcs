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
)

var (
	editVariableLong = templates.LongDesc(i18n.T(`
		Edit a project variable from the default editor.`))

	editVariableExample = templates.Examples(i18n.T(`
		# Edit a project variable by key
		kubectl-bcs-project-manager edit variable --key=key`))

	searchKey string
)

// 定义不能编辑的变量参数
type readOnlyVariableParam struct {
	ID           string `json:"id"`
	Key          string `json:"key"`
	Name         string `json:"name"`
	DefaultValue string `json:"defaultValue"`
	Scope        string `json:"scope"`
	ScopeName    string `json:"scopeName"`
	Category     string `json:"category"`
	CategoryName string `json:"categoryName"`
	Desc         string `json:"desc"`
	Created      string `json:"created"`
	Updated      string `json:"updated"`
	Creator      string `json:"creator"`
	Updater      string `json:"updater"`
}

func editVariable() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "variable --key=key",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Edit a project variable by key"),
		Long:                  editVariableLong,
		Example:               editVariableExample,
		Run: func(cmd *cobra.Command, args []string) {
			projectCode := viper.GetString("bcs.project_code")
			if len(projectCode) == 0 {
				klog.Fatalf("Project code (English abbreviation), global unique, the length cannot exceed 64 characters")
			}

			client := pkg.NewClientWithConfiguration(context.Background())
			// 通过key获取当前项目下变量
			variableListResp, err := client.ListVariableDefinitions(&pkg.ListVariableDefinitionsRequest{
				SearchKey: searchKey,
				All:       true,
			}, projectCode)
			if err != nil {
				klog.Fatalf("list variable definitions failed: %v", err)
			}

			variableList := make(map[string]readOnlyVariableParam, 0)
			for _, item := range variableListResp.Data.Results {
				variableList[item.Key] = readOnlyVariableParam{
					ID:           item.GetId(),
					Key:          item.GetKey(),
					Name:         item.GetName(),
					DefaultValue: item.GetDefaultValue(),
					Scope:        item.GetScope(),
					ScopeName:    item.GetScopeName(),
					Category:     item.GetCategory(),
					CategoryName: item.GetCategoryName(),
					Desc:         item.GetDesc(),
					Created:      item.GetCreated(),
					Updated:      item.GetUpdated(),
					Creator:      item.GetCreator(),
					Updater:      item.GetUpdater(),
				}
			}
			variable, ok := variableList[searchKey]
			if !ok {
				klog.Fatalf("No variable found for key: %v", searchKey)
			}
			// 原内容
			marshal, err := json.Marshal(variable)
			if err != nil {
				klog.Fatal("[variable] deserialize failed: %v", err)
			}
			// 把json转成yaml
			original, err := yaml.JSONToYAML(marshal)
			if err != nil {
				klog.Fatal("json to yaml failed: %v", err)
			}
			edit := editor.NewDefaultEditor([]string{})
			// 编辑后的
			edited, path, err := edit.LaunchTempFile(fmt.Sprintf("%s-edit-", filepath.Base(os.Args[0])), ".yaml", bytes.NewBufferString(string(original)))
			if err != nil {
				klog.Fatalf("unexpected error: %v", err)
			}
			if _, err := os.Stat(path); err != nil {
				klog.Fatalf("no temp file: %s", path)
			}
			// 对比原内容是否更改
			if bytes.Equal(cmdutil.StripComments(original), cmdutil.StripComments(edited)) {
				klog.Fatalf("Edit cancelled, no valid changes were saved.")
			}
			// 把编辑后的内容yaml转成json
			editedJson, err := yaml.YAMLToJSON(edited)
			if err != nil {
				klog.Fatal("json to yaml failed: %v", err)
			}

			var (
				editBefore readOnlyVariableParam
				editAfter  readOnlyVariableParam
			)

			{
				err = json.Unmarshal(editedJson, &editBefore)
				if err != nil {
					klog.Fatal("[edit before] deserialize failed: %v", err)
				}
				err = json.Unmarshal(marshal, &editAfter)
				if err != nil {
					klog.Fatal("[edit after] deserialize failed: %v", err)
				}
				editAfter.DefaultValue = editBefore.DefaultValue
				editAfter.Desc = editBefore.Desc
			}
			if editBefore != editAfter {
				klog.Fatal("only edit desc and default value")
			}
			// 保证只修改描述和默认值
			updateData := &pkg.UpdateVariableRequest{
				ProjectCode: projectCode,
				VariableID:  variable.ID,
				Name:        variable.Name,
				Key:         variable.Key,
				Scope:       variable.Scope,
				Default:     editBefore.DefaultValue,
				Desc:        editBefore.Desc,
			}

			resp, err := client.UpdateVariable(updateData)
			if err != nil {
				klog.Fatalf("update project variable failed: %v", err)
			}
			printer.PrintInJSON(resp)
		},
	}

	cmd.Flags().StringVarP(&searchKey, "key", "", "",
		"Variable key, through this field fuzzy query item variable")

	return cmd
}
