package project

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	projectMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/project"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete project from bcs-cluster-manager",
		Run:   delete,
	}

	cmd.Flags().StringVarP(&projectID, "projectID", "p", "", `project ID`)
	cmd.MarkFlagRequired("projectID")

	return cmd
}

func delete(cmd *cobra.Command, args []string) {
	err := projectMgr.New(context.Background()).Delete(manager.DeleteProjectReq{
		ProjectID: projectID,
		IsForce:   false,
	})
	if err != nil {
		klog.Fatalf("create project failed: %v", err)
	}

	klog.Infof("create project succeed")
}
