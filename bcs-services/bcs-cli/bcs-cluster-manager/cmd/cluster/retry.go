package cluster

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	clusterMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cluster"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

var (
	retryClusterID string
)

func newRetryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "retry",
		Short: "retry create cluster from bcs-cluster-manager",
		Run:   retry,
	}

	cmd.Flags().StringVarP(&retryClusterID, "clusterID", "i", "", "")

	return cmd
}

func retry(cmd *cobra.Command, args []string) {
	resp, err := clusterMgr.New(context.Background()).Retry(manager.RetryClusterReq{
		ClusterID: retryClusterID,
	})
	if err != nil {
		klog.Fatalf("retry create cluster failed: %v", err)
	}

	klog.Infof("retry create cluster succeed: clusterID: %v, taskID: %v", resp.ClusterID, resp.TaskID)
}
