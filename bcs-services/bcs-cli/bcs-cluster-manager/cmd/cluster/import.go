package cluster

import (
	"context"
	"encoding/json"
	"os"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	clusterMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cluster"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "import cluster from bcs-cluster-manager",
		Run:   importFunc,
	}

	cmd.Flags().StringVarP(&file, "file", "f", "./import.json", `create cluster from json file. file template: 
	{"clusterID":"","projectID":"3e11f4212ca2444d92a869c26fcbd4a9","businessID":"100148","engineType":"k8s",
	"isExclusive":false,"clusterType":"single","clusterName":"ceshi","environment":"stag","provider":"tencentCloud"}`)
	cmd.MarkFlagRequired("file")

	return cmd
}

func importFunc(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(file)
	if err != nil {
		klog.Fatalf("read file failed: %v", err)
	}

	req := manager.ImportClusterReq{}
	err = json.Unmarshal(data, &req)
	if err != nil {
		klog.Fatalf("unmarshal json file failed: %v", err)
	}

	err = clusterMgr.New(context.Background()).Import(req)
	if err != nil {
		klog.Fatalf("import cluster failed: %v", err)
	}

	klog.Infof("import cluster succeed")
}
