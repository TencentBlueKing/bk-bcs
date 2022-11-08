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

func newCreateNamespaceWithQuotaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "createNamespaceWithQuota",
		Short: "create namespace with quota from bcs-cluster-manager",
		Run:   createNamespaceWithQuota,
	}

	cmd.Flags().StringVarP(&file, "file", "f", "./create_namespace_with_quota.json", `create namespace with quota from json file. file template: 
	{"name":"test","federationClusterID":"","projectID":"","businessID":"","labels":{"xxx":"xxx","xxx":"xxx"}},"region":"","resourceQuota":""`)
	cmd.MarkFlagRequired("file")

	return cmd
}

func createNamespaceWithQuota(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(file)
	if err != nil {
		klog.Fatalf("read json file failed: %v", err)
	}

	req := manager.CreateNamespaceWithQuotaReq{}
	err = json.Unmarshal(data, &req)
	if err != nil {
		klog.Fatalf("unmarshal json file failed: %v", err)
	}

	resp, err := namespaceQuotaMgr.New(context.Background()).CreateNamespaceWithQuota(req)
	if err != nil {
		klog.Fatalf("create namespace failed: %v", err)
	}

	klog.Infof("%+v", resp)
}
