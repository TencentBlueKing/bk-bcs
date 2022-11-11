package cluster

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	clusterMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cluster"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newListNodesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "listNodes",
		Short: "list cluster nodes from bcs-cluster-manager",
		Run:   listNodes,
	}

	cmd.Flags().Uint32VarP(&offset, "offset", "o", 0, "offset number of queries")
	cmd.Flags().Uint32VarP(&limit, "limit", "l", 10, "limit number of queries")

	return cmd
}

func listNodes(cmd *cobra.Command, args []string) {
	resp, err := clusterMgr.New(context.Background()).ListNodes(manager.ListClusterNodesReq{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		klog.Fatalf("list cluster nodes failed: %v", err)
	}

	klog.Infof("%+v", resp.Data)
}
