package cluster

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	clusterMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cluster"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "create cluster from bcs-cluster-manager",
		Run:   create,
	}
}

func create(cmd *cobra.Command, args []string) {
	resp, err := clusterMgr.New(context.Background()).Create(manager.CreateClusterReq{})
	if err != nil {
		klog.Fatalf("create cluster failed: %v", err)
	}

	klog.Infof("create cluster succeed: clusterID: %v, taskID: %v", resp.ClusterID, resp.TaskID)
}
