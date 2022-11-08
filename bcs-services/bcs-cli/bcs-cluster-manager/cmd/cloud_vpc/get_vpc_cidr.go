package cloudvpc

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	cloudvpcMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cloud_vpc"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newGetVPCCidrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "getVPCCidr",
		Short: "list VPC Cidr from bcs-cluster-manager",
		Run:   getVPCCidr,
	}

	cmd.Flags().StringVarP(&vpcID, "vpcID", "c", "", `VPC ID`)
	cmd.MarkFlagRequired("vpcID")

	return cmd
}

func getVPCCidr(cmd *cobra.Command, args []string) {
	resp, err := cloudvpcMgr.New(context.Background()).GetVPCCidr(manager.GetVPCCidrReq{
		VPCID: vpcID,
	})
	if err != nil {
		klog.Fatalf("list VPC Cidr failed: %v", err)
	}

	klog.Infof("%+v", resp.Data)
}
