package node

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	nodeMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/node"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newDrainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drain",
		Short: "drain node from bcs-cluster-manager",
		Run:   drain,
	}

	cmd.Flags().StringSliceVarP(&innerIPs, "innerIPs", "i", []string{}, "node inner ip, for example: -i 47.43.47.103 -i 244.87.232.48")
	cmd.MarkFlagRequired("innerIPs")

	cmd.Flags().StringVarP(&clusterID, "clusterID", "c", "", "更新节点所属的clusterID")

	return cmd
}

func drain(cmd *cobra.Command, args []string) {
	resp, err := nodeMgr.New(context.Background()).Drain(manager.DrainNodeReq{
		InnerIPs:  innerIPs,
		ClusterID: clusterID,
	})
	if err != nil {
		klog.Fatalf("cordon node failed: %v", err)
	}

	klog.Infof("%+v", resp.Data)
}
