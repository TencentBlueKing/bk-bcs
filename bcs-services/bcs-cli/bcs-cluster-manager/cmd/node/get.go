package node

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	nodeMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/node"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "get node from bcs-cluster-manager",
		Run:   get,
	}

	cmd.Flags().StringVarP(&innerIP, "innerIP", "i", "", "")
	cmd.MarkFlagRequired("innerIP")

	return cmd
}

func get(cmd *cobra.Command, args []string) {
	resp, err := nodeMgr.New(context.Background()).Get(manager.GetNodeReq{
		InnerIP: innerIP,
	})
	if err != nil {
		klog.Fatalf("get node failed: %v", err)
	}

	klog.Infof("%+v", resp.Data)
}
