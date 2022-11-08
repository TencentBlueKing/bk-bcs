package cloudvpc

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	cloudvpcMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cloud_vpc"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newListCloudRegionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "listCloudRegions",
		Short: "list cloud regions from bcs-cluster-manager",
		Run:   listCloudRegions,
	}

	cmd.Flags().StringVarP(&cloudID, "cloudID", "c", "", `cloud ID`)
	cmd.MarkFlagRequired("cloudID")

	return cmd
}

func listCloudRegions(cmd *cobra.Command, args []string) {
	resp, err := cloudvpcMgr.New(context.Background()).ListCloudRegions(manager.ListCloudRegionsReq{
		CloudID: cloudID,
	})
	if err != nil {
		klog.Fatalf("list cloud regions failed: %v", err)
	}

	klog.Infof("%+v", resp.Data)
}
