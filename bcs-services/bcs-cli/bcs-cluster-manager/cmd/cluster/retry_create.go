package cluster

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	clusterMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cluster"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newRetryCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "retryCreate",
		Short: "retry create cluster from bcs-cluster-manager",
		Run:   retry,
	}

	cmd.Flags().StringVarP(&clusterID, "clusterID", "c", "", "cluster ID (required)")
	cmd.MarkFlagRequired("clusterID")

	return cmd
}

func retry(cmd *cobra.Command, args []string) {
	resp, err := clusterMgr.New(context.Background()).RetryCreate(manager.RetryCreateClusterReq{
		ClusterID: clusterID,
	})
	if err != nil {
		klog.Fatalf("retry create cluster failed: %v", err)
	}

	klog.Infof("retry create cluster succeed: clusterID: %v, taskID: %v", resp.ClusterID, resp.TaskID)
}
