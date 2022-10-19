/**
 * @Author: Ambition
 * @Description:
 * @File: update
 * @Version: 1.0.0
 * @Date: 2022/10/17 16:08
 */

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/cmd/printer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-project-manager/pkg"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/spf13/cobra"
	"k8s.io/klog"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/cmd/util/editor"
	"os"
	"sigs.k8s.io/yaml"
)

func newUpdateCmd() *cobra.Command {
	editCmd := &cobra.Command{
		Use:   "edit",
		Short: "",
		Long:  "edit project",
	}
	editCmd.AddCommand(updateProject())
	return editCmd
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
			edited, path, err := edit.LaunchTempFile("", "someprefix", bytes.NewBufferString(string(original)))
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
			param, err := yaml.YAMLToJSON(edited)
			if err != nil {
				klog.Fatal("json to yaml failed: %v", err)
			}
			var request bcsproject.UpdateProjectRequest
			json.Unmarshal(param, &request)
			useBKRes := new(wrappers.BoolValue)
			useBKRes.Value = getProjectResp.Data.UseBKRes

			isOffline := new(wrappers.BoolValue)
			isOffline.Value = projectInfo.IsOffline

			isSecret := new(wrappers.BoolValue)
			isSecret.Value = projectInfo.IsSecret

			// 保证只修改描述和业务ID
			updateData := &bcsproject.UpdateProjectRequest{
				ProjectID:   projectInfo.ProjectID,
				Name:        projectInfo.Name,
				UseBKRes:    useBKRes,
				Description: request.Description,
				IsOffline:   isOffline,
				Kind:        projectInfo.Kind,
				BusinessID:  request.BusinessID,
				IsSecret:    isSecret,
				DeployType:  projectInfo.DeployType,
				ProjectType: projectInfo.ProjectType,
				BGID:        projectInfo.BGID,
				BGName:      projectInfo.BGName,
				DeptID:      projectInfo.DeptID,
				DeptName:    projectInfo.DeptName,
				CenterID:    projectInfo.CenterID,
				CenterName:  projectInfo.CenterName,
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
