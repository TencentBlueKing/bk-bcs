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

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create project from bcs-cluster-manager",
		Run:   create,
	}

	cmd.Flags().StringVarP(&file, "file", "f", "./create_project.json", `create project from json file. file template: 
	{"name":"test","englishName":"","projectType":0,"useBKRes":false,"businessID":"","kind":"","deployType":0}`)
	cmd.MarkFlagRequired("file")

	return cmd
}

func create(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(file)
	if err != nil {
		klog.Fatalf("read json file failed: %v", err)
	}

	req := manager.CreateProjectReq{}
	err = json.Unmarshal(data, &req)
	if err != nil {
		klog.Fatalf("unmarshal json file failed: %v", err)
	}

	err = projectMgr.New(context.Background()).Create(req)
	if err != nil {
		klog.Fatalf("create project failed: %v", err)
	}

	klog.Infof("create project succeed")
}
