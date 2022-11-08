package federationcluster

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	federationClusterMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/federation_cluster"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "add federation cluster from bcs-cluster-manager",
		Run:   add,
	}

	cmd.Flags().StringVarP(&clusterID, "clusterID", "c", "", `cluster ID`)
	cmd.MarkFlagRequired("clusterID")
	cmd.Flags().StringVarP(&federationClusterID, "federationClusterID", "f", "", `federation cluster ID`)
	cmd.MarkFlagRequired("federationClusterID")

	return cmd
}

func add(cmd *cobra.Command, args []string) {
	err := federationClusterMgr.New(context.Background()).Add(manager.AddFederatedClusterReq{
		FederationClusterID: clusterID,
		ClusterID:           federationClusterID,
	})
	if err != nil {
		klog.Fatalf("add federation cluster failed: %v", err)
	}

	klog.Infof("add federation cluster succeed")
}
