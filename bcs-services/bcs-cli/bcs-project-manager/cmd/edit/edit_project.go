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

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/spf13/cobra"
	"k8s.io/klog"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/cmd/util/editor"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"sigs.k8s.io/yaml"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/cmd/printer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/pkg"
)

var (
	editProjectLong = templates.LongDesc(i18n.T(`
		Edit a project from the default editor.`))

	editProjectExample = templates.Examples(i18n.T(`
		# Edit project by PROJECT_ID or PROJECT_CODE
		kubectl-bcs-project-manager edit project PROJECT_ID/PROJECT_CODE`))
)

func editProject() *cobra.Command { // nolint
	cmd := &cobra.Command{
		Use:                   "project (ID/CODE)",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Edit project by ID or CODE"),
		Long:                  editProjectLong,
		Example:               editProjectExample,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				klog.Infoln("edit project requires project ID or code")
				return
			}
			client := pkg.NewClientWithConfiguration(context.Background())
			resp, err := client.GetProject(args[0])
			if err != nil {
				klog.Infof("get project failed: %v", err)
				return
			}

			projectInfo := resp.Data
			// 原内容
			marshal, err := json.Marshal(projectInfo)
			if err != nil {
				klog.Infof("json marshal failed: %v", err)
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
			}
			if _, err = os.Stat(path); err != nil {
				klog.Infof("no temp file: %s", path)
			}
			// 对比原内容是否更改
			if bytes.Equal(cmdutil.StripComments(original), cmdutil.StripComments(edited)) {
				klog.Infoln("Edit canceled, no valid changes were saved.")
			}
			// 把编辑后的内容yaml转成json
			editedJson, err := yaml.YAMLToJSON(edited)
			if err != nil {
				klog.Infof("json to yaml failed: %v", err)
				return
			}

			var (
				editBefore pkg.UpdateProjectRequest
				editAfter  pkg.UpdateProjectRequest
			)

			{
				err = json.Unmarshal(editedJson, &editBefore)
				if err != nil {
					klog.Infof("[edited] deserialize failed: %v", err)
					return
				}
				err = json.Unmarshal(marshal, &editAfter)
				if err != nil {
					klog.Infof("[project info] deserialize failed: %v", err)
					return
				}
				editAfter.BusinessID = editBefore.BusinessID
				editAfter.Description = editBefore.Description
			}

			if editBefore != editAfter {
				klog.Infoln("only edit description and project ID")
				return
			}

			useBKRes := new(wrappers.BoolValue)
			useBKRes.Value = projectInfo.GetUseBKRes()

			isOffline := new(wrappers.BoolValue)
			isOffline.Value = projectInfo.GetIsOffline()

			// 保证只修改描述和业务ID
			updateData := &pkg.UpdateProjectRequest{
				BusinessID:  editBefore.BusinessID,
				CreateTime:  projectInfo.GetCreateTime(),
				Creator:     projectInfo.GetCreator(),
				Description: editBefore.Description,
				Kind:        projectInfo.GetKind(),
				Managers:    projectInfo.GetManagers(),
				Name:        projectInfo.GetName(),
				ProjectCode: projectInfo.GetProjectCode(),
				ProjectID:   projectInfo.GetProjectID(),
				UpdateTime:  projectInfo.GetUpdateTime(),
			}
			updateProjectResp, err := client.UpdateProject(updateData)
			if err != nil {
				klog.Infof("update project failed: %v", err)
				return
			}
			printer.PrintInJSON(updateProjectResp)
		},
	}

	return cmd
}
