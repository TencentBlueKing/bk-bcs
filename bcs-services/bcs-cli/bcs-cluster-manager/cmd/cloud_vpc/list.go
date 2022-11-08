package cloudvpc

import (
	"context"

	cloudvpcMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cloud_vpc"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list cloud vpc from bcs-cluster-manager",
		Run:   list,
	}

	return cmd
}

func list(cmd *cobra.Command, args []string) {
	resp, err := cloudvpcMgr.New(context.Background()).List()
	if err != nil {
		klog.Fatalf("list cloud vpc failed: %v", err)
	}

	klog.Infof("%+v", resp.Data)
}
