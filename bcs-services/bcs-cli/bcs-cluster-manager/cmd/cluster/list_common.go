package cluster

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	clusterMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cluster"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newListCommonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "listCommon",
		Short: "list common cluster from bcs-cluster-manager",
		Run:   listCommon,
	}

	return cmd
}

func listCommon(cmd *cobra.Command, args []string) {
	resp, err := clusterMgr.New(context.Background()).ListCommon(manager.ListCommonClusterReq{})
	if err != nil {
		klog.Fatalf("list common cluster failed: %v", err)
	}

	klog.Infof("%+v", resp.Data)
}
