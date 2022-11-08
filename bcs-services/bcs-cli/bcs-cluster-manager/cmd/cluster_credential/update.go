package clustercredential

import (
	"context"
	"encoding/json"
	"os"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	clustercredentialMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cluster_credential"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update cluster credential from bcs-cluster-manager",
		Run:   update,
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", `update cluster credential json file`)
	cmd.MarkFlagRequired("file")

	return cmd
}

func update(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(file)
	if err != nil {
		klog.Fatalf("read json file failed: %v", err)
	}

	req := manager.UpdateClusterCredentialReq{}
	err = json.Unmarshal(data, &req)
	if err != nil {
		klog.Fatalf("unmarshal json file failed: %v", err)
	}

	err = clustercredentialMgr.New(context.Background()).Update(manager.UpdateClusterCredentialReq{
		ClusterID:     req.ClusterID,
		ClientModule:  req.ClientModule,
		ServerAddress: req.ServerAddress,
		CaCertData:    req.CaCertData,
		UserToken:     req.UserToken,
		ClusterDomain: req.ClusterDomain,
	})
	if err != nil {
		klog.Fatalf("update cluster credential failed: %v", err)
	}

	klog.Infof("update cluster credential succeed")
}
