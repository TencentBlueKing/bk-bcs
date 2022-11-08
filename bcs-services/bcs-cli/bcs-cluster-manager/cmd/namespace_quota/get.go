package namespacequota

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	namespaceQuotaMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/namespace_quota"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "get namespace from bcs-cluster-manager",
		Run:   get,
	}

	cmd.Flags().StringVarP(&clusterID, "clusterID", "c", "", `cluster ID`)
	cmd.MarkFlagRequired("clusterID")
	cmd.Flags().StringVarP(&namespace, "namespace", "c", "", `namespace`)
	cmd.MarkFlagRequired("namespace")
	cmd.Flags().StringVarP(&federationClusterID, "federationClusterID", "f", "", `federation cluster ID`)
	cmd.MarkFlagRequired("federationClusterID")

	return cmd
}

func get(cmd *cobra.Command, args []string) {
	resp, err := namespaceQuotaMgr.New(context.Background()).Get(manager.GetNamespaceQuotaReq{
		ClusterID:           clusterID,
		Namespace:           namespace,
		FederationClusterID: federationClusterID,
	})
	if err != nil {
		klog.Fatalf("create namespace failed: %v", err)
	}

	klog.Infof("%+v", resp)
}
