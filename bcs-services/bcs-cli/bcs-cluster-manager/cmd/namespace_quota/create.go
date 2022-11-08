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

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create namespace from bcs-cluster-manager",
		Run:   create,
	}

	cmd.Flags().StringVarP(&file, "file", "f", "./create_namespace_quota.json", `create namespace quota from json file. file template: 
	{"namespace":"test","federationClusterID":"","resourceQuota":""`)
	cmd.MarkFlagRequired("file")

	return cmd
}

func create(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(file)
	if err != nil {
		klog.Fatalf("read json file failed: %v", err)
	}

	req := manager.CreateNamespaceQuotaReq{}
	err = json.Unmarshal(data, &req)
	if err != nil {
		klog.Fatalf("unmarshal json file failed: %v", err)
	}

	resp, err := namespaceQuotaMgr.New(context.Background()).Create(req)
	if err != nil {
		klog.Fatalf("create namespace failed: %v", err)
	}

	klog.Infof("%+v", resp)
}
