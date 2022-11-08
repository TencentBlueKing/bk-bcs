package namespacequota

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	namespaceQuotaMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/namespace_quota"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete namespace from bcs-cluster-manager",
		Run:   delete,
	}

	cmd.Flags().StringVarP(&clusterID, "clusterID", "c", "", `cluster ID`)
	cmd.MarkFlagRequired("clusterID")
	cmd.Flags().StringVarP(&namespace, "namespace", "c", "", `namespace`)
	cmd.MarkFlagRequired("namespace")
	cmd.Flags().StringVarP(&federationClusterID, "federationClusterID", "f", "", `federation cluster ID`)
	cmd.MarkFlagRequired("federationClusterID")

	return cmd
}

func delete(cmd *cobra.Command, args []string) {
	err := namespaceQuotaMgr.New(context.Background()).Delete(manager.DeleteNamespaceQuotaReq{
		ClusterID:           clusterID,
		Namespace:           namespace,
		FederationClusterID: federationClusterID,
	})
	if err != nil {
		klog.Fatalf("delete namespace quota failed: %v", err)
	}

	klog.Infof("delete namespace quota succeed")
}
