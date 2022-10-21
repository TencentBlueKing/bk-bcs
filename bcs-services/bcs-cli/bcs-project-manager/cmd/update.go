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

package cmd

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
	"sigs.k8s.io/yaml"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/cmd/printer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/pkg"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

func newUpdateCmd() *cobra.Command {
	editCmd := &cobra.Command{
		Use:   "edit",
		Short: "edit project from bcs-project-manager",
		Long:  "edit project from bcs-project-manager",
	}
	editCmd.AddCommand(updateProject())
	return editCmd
}

//定义不能编辑的参数
type readOnlyParam struct {
	CreateTime  string `json:"createTime"`
	UpdateTime  string `json:"updateTime"`
	Creator     string `json:"creator"`
	Updater     string `json:"updater"`
	Managers    string `json:"managers"`
	ProjectID   string `json:"projectID"`
	Name        string `json:"name"`
	ProjectCode string `json:"projectCode"`
	UseBKRes    bool   `json:"useBKRes"`
	IsOffline   bool   `json:"isOffline"`
	Kind        string `json:"kind"`
	IsSecret    bool   `json:"isSecret"`
	ProjectType uint32 `json:"projectType"`
	DeployType  uint32 `json:"deployType"`
	BGID        string `json:"BGID"`
	BGName      string `json:"BGName"`
	DeptID      string `json:"deptID"`
	DeptName    string `json:"deptName"`
	CenterID    string `json:"centerID"`
	CenterName  string `json:"centerName"`
	BusinessID  string `json:"businessID"`
	Description string `json:"description"`
}

func updateProject() *cobra.Command {
	subCmd := &cobra.Command{
		Use:                   "project (ID/CODE)",
		DisableFlagsInUseLine: true,
		Short:                 "",
		Long:                  "edit infos from bcs-project-manager",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				klog.Fatalf("edit project requires project ID or code")
			}
			ctx, cancel := context.WithCancel(context.Background())
			client, cliCtx, err := pkg.NewClientWithConfiguration(ctx)
			if err != nil {
				klog.Fatalf("init client failed: %v", err.Error())
			}
			defer cancel()
			// 先获取当前项目
			getProjectResp, err := client.GetProject(cliCtx, &bcsproject.GetProjectRequest{ProjectIDOrCode: args[0]})
			if err != nil {
				klog.Fatalf("get project failed: %v", err)
			}
			if getProjectResp != nil && getProjectResp.Code != 0 {
				klog.Fatal("get project response code not 0 but %d: %s", getProjectResp.Code, getProjectResp.Message)
			}
			projectInfo := getProjectResp.Data
			// 原内容
			marshal, err := json.Marshal(projectInfo)
			if err != nil {
				klog.Fatal("json marshal failed: %v", err)
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
				editBefore readOnlyParam
				editAfter  readOnlyParam
			)

			{
				err = json.Unmarshal(editedJson, &editBefore)
				if err != nil {
					klog.Fatal("[edited] deserialize failed: %v", err)
				}
				err = json.Unmarshal(marshal, &editAfter)
				if err != nil {
					klog.Fatal("[project info] deserialize failed: %v", err)
				}
				editAfter.BusinessID = editBefore.BusinessID
				editAfter.Description = editBefore.Description
			}

			if editBefore != editAfter {
				klog.Fatal("only edit description and project ID")
			}

			useBKRes := new(wrappers.BoolValue)
			useBKRes.Value = projectInfo.GetUseBKRes()

			isOffline := new(wrappers.BoolValue)
			isOffline.Value = projectInfo.GetIsOffline()

			isSecret := new(wrappers.BoolValue)
			isSecret.Value = projectInfo.GetIsSecret()

			// 保证只修改描述和业务ID
			updateData := &bcsproject.UpdateProjectRequest{
				ProjectID:   projectInfo.GetProjectID(),
				Name:        projectInfo.GetName(),
				UseBKRes:    useBKRes,
				Description: editBefore.Description,
				IsOffline:   isOffline,
				Kind:        projectInfo.GetKind(),
				BusinessID:  editBefore.BusinessID,
				IsSecret:    isSecret,
				DeployType:  projectInfo.GetDeployType(),
				ProjectType: projectInfo.GetProjectType(),
				BGID:        projectInfo.GetBGID(),
				BGName:      projectInfo.GetBGName(),
				DeptID:      projectInfo.GetDeptID(),
				DeptName:    projectInfo.GetDeptName(),
				CenterID:    projectInfo.GetCenterID(),
				CenterName:  projectInfo.GetCenterName(),
			}
			updateProjectResp, err := client.UpdateProject(cliCtx, updateData)
			if err != nil {
				klog.Fatalf("update project failed: %v", err)
			}
			if updateProjectResp != nil && updateProjectResp.Code != 0 {
				klog.Fatal("update project response code not 0 but %d: %s", updateProjectResp.Code, updateProjectResp.Message)
			}
			printer.PrintUpdateProjectInJSON(updateProjectResp)
		},
	}

	return subCmd
}

func getProjectInfo() {

}
