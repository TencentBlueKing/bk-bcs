package project

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	projectMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/project"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list project from bcs-cluster-manager",
		Run:   list,
	}

	return cmd
}

func list(cmd *cobra.Command, args []string) {
	resp, err := projectMgr.New(context.Background()).List(manager.ListProjectReq{})
	if err != nil {
		klog.Fatalf("list project failed: %v", err)
	}

	klog.Infof("%+v", resp.Data)
}
