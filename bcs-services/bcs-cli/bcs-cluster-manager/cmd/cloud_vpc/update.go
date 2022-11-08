package cloudvpc

import (
	"context"
	"encoding/json"
	"os"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	cloudvpcMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cloud_vpc"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update cloud vpc from bcs-cluster-manager",
		Run:   update,
	}

	cmd.Flags().StringVarP(&file, "file", "f", "./update_cloudvpc.json", `update cloud vpc from json file. file template: 
	{"cloudID":"","networkType":"","region":"","regionName":"","vpcName":"test","vpcID":"","available":true}`)
	cmd.MarkFlagRequired("file")

	return cmd
}

func update(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(file)
	if err != nil {
		klog.Fatalf("read json file failed: %v", err)
	}

	req := manager.UpdateCloudVPCReq{}
	err = json.Unmarshal(data, &req)
	if err != nil {
		klog.Fatalf("unmarshal json file failed: %v", err)
	}

	err = cloudvpcMgr.New(context.Background()).Update(req)
	if err != nil {
		klog.Fatalf("update cloud vpc failed: %v", err)
	}

	klog.Infof("update cloud vpc succeed")
}
