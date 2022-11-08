package namespace

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	namespaceMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/namespace"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "get namespace from bcs-cluster-manager",
		Run:   get,
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", `namespace name`)
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVarP(&federationClusterID, "federationClusterID", "f", "", `federation cluster ID`)
	cmd.MarkFlagRequired("federationClusterID")

	return cmd
}

func get(cmd *cobra.Command, args []string) {
	resp, err := namespaceMgr.New(context.Background()).Get(manager.GetNamespaceReq{
		Name:                name,
		FederationClusterID: federationClusterID,
	})
	if err != nil {
		klog.Fatalf("get namespace failed: %v", err)
	}

	klog.Infof("%+v", resp.Data)
}
