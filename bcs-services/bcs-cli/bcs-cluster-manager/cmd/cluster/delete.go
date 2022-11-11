package cluster

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	clusterMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cluster"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete cluster from bcs-cluster-manager",
		Run:   delete,
	}

	cmd.Flags().StringVarP(&clusterID, "clusterID", "c", "", "cluster ID (required)")
	cmd.MarkFlagRequired("clusterID")

	return cmd
}

func delete(cmd *cobra.Command, args []string) {
	resp, err := clusterMgr.New(context.Background()).Delete(manager.DeleteClusterReq{ClusterID: clusterID})
	if err != nil {
		klog.Fatalf("delete cluster failed: %v", err)
	}

	klog.Infof("create cluster succeed: clusterID: %v, taskID: %v", resp.ClusterID, resp.TaskID)
}
