package project

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	projectMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/project"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "get project from bcs-cluster-manager",
		Run:   get,
	}

	cmd.Flags().StringVarP(&projectID, "projectID", "p", "", `project ID`)
	cmd.MarkFlagRequired("projectID")

	return cmd
}

func get(cmd *cobra.Command, args []string) {
	resp, err := projectMgr.New(context.Background()).Get(manager.GetProjectReq{
		ProjectID: projectID,
	})
	if err != nil {
		klog.Fatalf("get project failed: %v", err)
	}

	klog.Infof("%+v", resp.Data)
}
