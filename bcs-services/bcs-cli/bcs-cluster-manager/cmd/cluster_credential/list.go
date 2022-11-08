package clustercredential

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	clustercredentialMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cluster_credential"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list cluster credential from bcs-cluster-manager",
		Run:   list,
	}

	cmd.Flags().Uint32VarP(&offset, "offset", "o", 0, `offset`)
	cmd.Flags().Uint32VarP(&limit, "limit", "l", 10, `limit`)

	return cmd
}

func list(cmd *cobra.Command, args []string) {
	resp, err := clustercredentialMgr.New(context.Background()).List(manager.ListClusterCredentialReq{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		klog.Fatalf("list cluster credential failed: %v", err)
	}

	klog.Infof("%+v", resp.Data)
}
