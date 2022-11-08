package clustercredential

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	clustercredentialMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cluster_credential"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete cluster credential from bcs-cluster-manager",
		Run:   delete,
	}

	cmd.Flags().StringVarP(&serverKey, "serverKey", "s", "", `server key`)
	cmd.MarkFlagRequired("serverKey")

	return cmd
}

func delete(cmd *cobra.Command, args []string) {
	err := clustercredentialMgr.New(context.Background()).Delete(manager.DeleteClusterCredentialReq{
		ServerKey: serverKey,
	})
	if err != nil {
		klog.Fatalf("delete cluster credential failed: %v", err)
	}

	klog.Infof("delete cluster credential succeed")
}
