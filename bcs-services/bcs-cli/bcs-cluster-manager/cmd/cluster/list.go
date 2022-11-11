package cluster

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	clusterMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cluster"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list cluster from bcs-cluster-manager",
		Run:   list,
	}

	cmd.Flags().Uint32VarP(&offset, "offset", "o", 0, "offset number of queries")
	cmd.Flags().Uint32VarP(&limit, "limit", "l", 10, "limit number of queries")

	return cmd
}

func list(cmd *cobra.Command, args []string) {
	resp, err := clusterMgr.New(context.Background()).List(manager.ListClusterReq{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		klog.Fatalf("list cluster failed: %v", err)
	}

	klog.Infof("%+v", resp.Data)
}
