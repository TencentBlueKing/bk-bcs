package namespacequota

import (
	"context"
	"encoding/json"
	"os"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	namespaceQuotaMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/namespace_quota"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update namespace from bcs-cluster-manager",
		Run:   update,
	}

	cmd.Flags().StringVarP(&file, "file", "f", "./update_namespace_quota.json", `update namespace quota from json file. file template: 
	{"namespace":"test","federationClusterID":"","clusterID":"","resourceQuota":""`)
	cmd.MarkFlagRequired("file")

	return cmd
}

func update(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(file)
	if err != nil {
		klog.Fatalf("read json file failed: %v", err)
	}

	req := manager.UpdateNamespaceQuotaReq{}
	err = json.Unmarshal(data, &req)
	if err != nil {
		klog.Fatalf("unmarshal json file failed: %v", err)
	}

	err = namespaceQuotaMgr.New(context.Background()).Update(req)
	if err != nil {
		klog.Fatalf("update namespace quota failed: %v", err)
	}

	klog.Infof("update namespace quota succeed")
}
