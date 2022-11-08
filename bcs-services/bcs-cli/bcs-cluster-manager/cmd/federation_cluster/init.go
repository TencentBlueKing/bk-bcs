package federationcluster

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	federationClusterMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/federation_cluster"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "init federation cluster from bcs-cluster-manager",
		Run:   initFunc,
	}

	return cmd
}

func initFunc(cmd *cobra.Command, args []string) {
	err := federationClusterMgr.New(context.Background()).Init(manager.InitFederationClusterReq{})
	if err != nil {
		klog.Fatalf("init federation cluster failed: %v", err)
	}

	klog.Infof("init federation cluster succeed")
}
