package cloudvpc

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	cloudvpcMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cloud_vpc"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete cloud vpc from bcs-cluster-manager",
		Run:   delete,
	}

	cmd.Flags().StringVarP(&cloudID, "cloudID", "c", "", `cloud ID (required)`)
	cmd.MarkFlagRequired("cloudID")
	cmd.Flags().StringVarP(&vpcID, "vpcID", "v", "", `VPC ID (required)`)
	cmd.MarkFlagRequired("vpcID")

	return cmd
}

func delete(cmd *cobra.Command, args []string) {
	err := cloudvpcMgr.New(context.Background()).Update(manager.UpdateCloudVPCReq{
		CloudID: cloudID,
		VPCID:   vpcID,
	})
	if err != nil {
		klog.Fatalf("delete cloud vpc failed: %v", err)
	}

	klog.Infof("delete cloud vpc succeed")
}
