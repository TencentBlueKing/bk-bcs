package node

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	nodeMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/node"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newCheckNodeInClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkNodeInCluster",
		Short: "check node in cluster from bcs-cluster-manager",
		Run:   checkNodeInCluster,
	}

	cmd.Flags().StringSliceVarP(&innerIPs, "innerIPs", "i", []string{}, "node inner ip, for example: -i 47.43.47.103 -i 244.87.232.48")
	cmd.MarkFlagRequired("innerIPs")

	return cmd
}

func checkNodeInCluster(cmd *cobra.Command, args []string) {
	resp, err := nodeMgr.New(context.Background()).CheckNodeInCluster(manager.CheckNodeInClusterReq{
		InnerIPs: innerIPs,
	})
	if err != nil {
		klog.Fatalf("check node in cluster failed: %v", err)
	}

	klog.Infof("%+v", resp.Data)
}
