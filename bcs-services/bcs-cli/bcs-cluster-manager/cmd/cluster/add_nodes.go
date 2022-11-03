package cluster

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	clusterMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cluster"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newAddNodesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addNodes",
		Short: "add nodes to cluster from bcs-cluster-manager",
		Run:   addNodes,
	}

	cmd.Flags().StringVarP(&clusterID, "clusterID", "i", "", "cluster ID (required)")
	cmd.MarkFlagRequired("clusterID")

	cmd.Flags().StringSliceVarP(&nodes, "node", "n", []string{}, "node ip, for example: -n 47.43.47.103 -n 244.87.232.48")
	cmd.MarkFlagRequired("node")

	return cmd
}

func addNodes(cmd *cobra.Command, args []string) {
	resp, err := clusterMgr.New(context.Background()).AddNodes(manager.AddNodesClusterReq{
		ClusterID: clusterID,
		Nodes:     nodes,
	})
	if err != nil {
		klog.Fatalf("add nodes to cluster failed: %v", err)
	}

	klog.Infof("add nodes to cluster succeed: taskID: %v", resp.TaskID)
}
