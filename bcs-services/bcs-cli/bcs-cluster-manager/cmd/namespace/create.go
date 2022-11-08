package namespace

import (
	"context"
	"encoding/json"
	"os"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	namespaceMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/namespace"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create namespace from bcs-cluster-manager",
		Run:   create,
	}

	cmd.Flags().StringVarP(&file, "file", "f", "./create_namespace.json", `create namespace from json file. file template: 
	{"name":"test","federationClusterID":"","projectID":"","businessID":"","labels":{"xxx":"xxx","xxx":"xxx"}}`)
	cmd.MarkFlagRequired("file")

	return cmd
}

func create(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(file)
	if err != nil {
		klog.Fatalf("read json file failed: %v", err)
	}

	req := manager.CreateNamespaceReq{}
	err = json.Unmarshal(data, &req)
	if err != nil {
		klog.Fatalf("unmarshal json file failed: %v", err)
	}

	err = namespaceMgr.New(context.Background()).Create(req)
	if err != nil {
		klog.Fatalf("create namespace failed: %v", err)
	}

	klog.Infof("create namespace succeed")
}
