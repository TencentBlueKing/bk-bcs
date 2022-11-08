package project

import (
	"context"
	"encoding/json"
	"os"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	projectMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/project"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update project from bcs-cluster-manager",
		Run:   update,
	}

	cmd.Flags().StringVarP(&file, "file", "f", "./update_project.json", `update project from json file. file template: 
	{"projectID":"","name":"","updater":"","projectType":0,"useBKRes":false,"businessID":"","description":"",
	"isOffline":false,"kind":"","deployType":0,"bgID":"","bgName":"","deptID":"","deptName":"","centerID":"",
	"centerName":"","isSecret":false,"credentials":{"test":{"key":"","secret":"","subscriptionID":"",
	"tenantID":"","resourceGroupName":"","clientID":"","clientSecret":""}}}`)
	cmd.MarkFlagRequired("file")

	return cmd
}

func update(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(file)
	if err != nil {
		klog.Fatalf("read json file failed: %v", err)
	}

	req := manager.UpdateProjectReq{}
	err = json.Unmarshal(data, &req)
	if err != nil {
		klog.Fatalf("unmarshal json file failed: %v", err)
	}

	err = projectMgr.New(context.Background()).Update(req)
	if err != nil {
		klog.Fatalf("create project failed: %v", err)
	}

	klog.Infof("create project succeed")
}
