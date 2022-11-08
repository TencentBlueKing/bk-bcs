package namespace

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	namespaceMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/namespace"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list namespace from bcs-cluster-manager",
		Run:   list,
	}

	cmd.Flags().StringVarP(&federationClusterID, "federationClusterID", "f", "", `federation cluster ID`)
	cmd.MarkFlagRequired("federationClusterID")
	cmd.Flags().StringVarP(&projectID, "projectID", "p", "", `project ID`)
	cmd.MarkFlagRequired("projectID")
	cmd.Flags().StringVarP(&businessID, "businessID", "b", "", `business ID`)
	cmd.MarkFlagRequired("businessID")
	cmd.Flags().Uint32VarP(&offset, "offset", "o", 0, `offset`)
	cmd.Flags().Uint32VarP(&limit, "limit", "l", 10, `limit`)

	return cmd
}

func list(cmd *cobra.Command, args []string) {
	resp, err := namespaceMgr.New(context.Background()).List(manager.ListNamespaceReq{
		FederationClusterID: federationClusterID,
		ProjectID:           projectID,
		BusinessID:          businessID,
		Offset:              offset,
		Limit:               limit,
	})
	if err != nil {
		klog.Fatalf("list namespace failed: %v", err)
	}

	klog.Infof("%+v", resp.Data)
}
