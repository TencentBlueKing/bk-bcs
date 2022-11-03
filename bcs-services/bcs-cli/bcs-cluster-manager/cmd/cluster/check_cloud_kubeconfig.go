package cluster

import (
	"context"
	"os"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	clusterMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cluster"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newCheckCloudKubeConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkCloudKubeconfig",
		Short: "delete cluster from bcs-cluster-manager",
		Run:   checkCloudKubeconfig,
	}

	cmd.Flags().StringVarP(&file, "file", "f", "./config", "kubeconfig file (required)")
	cmd.MarkFlagRequired("file")

	return cmd
}

func checkCloudKubeconfig(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(file)
	if err != nil {
		klog.Fatalf("read file failed: %v", err)
	}

	err = clusterMgr.New(context.Background()).CheckCloudKubeconfig(manager.CheckCloudKubeconfigReq{
		Kubeconfig: string(data),
	})
	if err != nil {
		klog.Fatalf("check cloud kubeconfig failed: %v", err)
	}

	klog.Infof("check cloud kubeconfig succeed")
}
