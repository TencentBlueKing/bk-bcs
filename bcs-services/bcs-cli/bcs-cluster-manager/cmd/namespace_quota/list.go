package namespacequota

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	namespaceQuotaMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/namespace_quota"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list namespace from bcs-cluster-manager",
		Run:   list,
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "c", "", `namespace`)
	cmd.MarkFlagRequired("namespace")
	cmd.Flags().StringVarP(&federationClusterID, "federationClusterID", "f", "", `federation cluster ID`)
	cmd.MarkFlagRequired("federationClusterID")

	cmd.Flags().Uint32VarP(&offset, "offset", "o", 0, `offset`)
	cmd.MarkFlagRequired("offset")
	cmd.Flags().Uint32VarP(&limit, "limit", "l", 0, `limit`)
	cmd.MarkFlagRequired("limit")

	return cmd
}

func list(cmd *cobra.Command, args []string) {
	resp, err := namespaceQuotaMgr.New(context.Background()).List(manager.ListNamespaceQuotaReq{
		Namespace:           namespace,
		FederationClusterID: federationClusterID,
		Offset:              offset,
		Limit:               limit,
	})
	if err != nil {
		klog.Fatalf("create namespace failed: %v", err)
	}

	klog.Infof("%+v", resp)
}
