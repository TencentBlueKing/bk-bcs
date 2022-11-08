package clustercredential

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	clustercredentialMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cluster_credential"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "get cluster credential from bcs-cluster-manager",
		Run:   get,
	}

	cmd.Flags().StringVarP(&serverKey, "serverKey", "s", "", `server key`)
	cmd.MarkFlagRequired("serverKey")

	return cmd
}

func get(cmd *cobra.Command, args []string) {
	resp, err := clustercredentialMgr.New(context.Background()).Get(manager.GetClusterCredentialReq{
		ServerKey: "",
	})
	if err != nil {
		klog.Fatalf("get cluster credential failed: %v", err)
	}

	klog.Infof("%+v", resp.Data)
}
