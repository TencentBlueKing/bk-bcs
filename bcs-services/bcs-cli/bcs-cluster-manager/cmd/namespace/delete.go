package namespace

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	namespaceMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/namespace"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete namespace from bcs-cluster-manager",
		Run:   delete,
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", `namespace name`)
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVarP(&federationClusterID, "federationClusterID", "f", "", `federation cluster ID`)
	cmd.MarkFlagRequired("name")
	cmd.Flags().BoolVarP(&isForced, "isForced", "i", false, `forcibly delete or not`)

	return cmd
}

func delete(cmd *cobra.Command, args []string) {
	err := namespaceMgr.New(context.Background()).Delete(manager.DeleteNamespaceReq{
		Name:                name,
		FederationClusterID: federationClusterID,
		IsForced:            isForced,
	})
	if err != nil {
		klog.Fatalf("delete namespace failed: %v", err)
	}

	klog.Infof("delete namespace succeed")
}
