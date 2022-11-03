package cluster

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	clusterMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cluster"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "get cluster from bcs-cluster-manager",
		Run:   get,
	}

	cmd.Flags().StringVarP(&clusterID, "clusterID", "i", "", "cluster ID (required)")
	cmd.MarkFlagRequired("clusterID")

	return cmd
}

func get(cmd *cobra.Command, args []string) {
	resp, err := clusterMgr.New(context.Background()).Get(manager.GetClusterReq{ClusterID: clusterID})
	if err != nil {
		klog.Fatalf("get cluster failed: %v", err)
	}

	resp = manager.GetClusterResp{
		Data: &manager.Cluster{
			ClusterID:   "ClusterID",
			ProjectID:   "ProjectID",
			BusinessID:  "BusinessID",
			EngineType:  "EngineType",
			IsExclusive: false,
			ClusterType: "ClusterType",
			Creator:     "Creator",
			Updater:     "Updater",
			ManageType:  "ManageType",
			ClusterName: "ClusterName",
			Environment: "Environment",
			Provider:    "Provider",
			Description: "Description",
			ClusterBasicSettings: manager.ClusterBasicSettings{
				Version: "Version",
			},
			NetworkType: "NetworkType",
			VpcID:       "VpcID",
			NetworkSettings: manager.NetworkSettings{
				CidrStep:      1,
				MaxNodePodNum: 2,
				MaxServiceNum: 3,
			},
			Master: []string{"xxx.xxx.xxx.xxx"},
		},
	}

	klog.Infof("%+v", *resp.Data)
}
