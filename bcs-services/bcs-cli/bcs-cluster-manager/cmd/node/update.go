package node

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	nodeMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/node"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update node from bcs-cluster-manager",
		Run:   update,
	}

	cmd.Flags().StringSliceVarP(&innerIPs, "innerIPs", "i", []string{}, "node inner ip, for example: -i 47.43.47.103 -i 244.87.232.48")
	cmd.MarkFlagRequired("innerIPs")

	cmd.Flags().StringVarP(&status, "status", "s", "", "更新节点状态(INITIALIZATION/RUNNING/DELETING/ADD-FAILURE/REMOVE-FAILURE)")
	cmd.Flags().StringVarP(&nodeGroupID, "nodeGroupID", "n", "", "更新节点所属的node group ID")
	cmd.Flags().StringVarP(&clusterID, "clusterID", "c", "", "更新节点所属的clusterID")

	return cmd
}

func update(cmd *cobra.Command, args []string) {
	err := nodeMgr.New(context.Background()).Update(manager.UpdateNodeReq{
		InnerIPs:    innerIPs,
		Status:      status,
		NodeGroupID: nodeGroupID,
		ClusterID:   clusterID,
	})
	if err != nil {
		klog.Fatalf("get node failed: %v", err)
	}

	klog.Infof("update node succeed")
}
