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

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create cluster from bcs-cluster-manager",
		Run:   create,
	}

	cmd.Flags().StringVarP(&file, "file", "f", "./create_cluster.json", `create cluster from json file. file template: 
{"projectID":"3e11f4212ca2444d92a869c26fcbd4a9","businessID":"100148","engineType":"k8s","isExclusive":true,"clusterType":"single",
"creator":"evanxinli","manageType":"INDEPENDENT_CLUSTER","clusterName":"ceshi","environment":"stag","provider":"tencentCloud",
"description":"fsd","clusterBasicSettings":{"version":"1.20.6"},"networkType":"overlay","region":"ap-tokyo","vpcID":"vpc-8iple1iq",
"networkSettings":{"cidrStep":2048,"maxNodePodNum":32,"maxServiceNum":128},"master":["11.143.254.20","11.143.254.2"]}`)
	cmd.MarkFlagRequired("file")

	return cmd
}

func create(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(file)
	if err != nil {
		klog.Fatalf("read json file failed: %v", err)
	}

	req := manager.CreateClusterReq{}
	err = json.Unmarshal(data, &req)
	if err != nil {
		klog.Fatalf("unmarshal json file failed: %v", err)
	}

	resp, err := clusterMgr.New(context.Background()).Create(req)
	if err != nil {
		klog.Fatalf("create cluster failed: %v", err)
	}

	klog.Infof("create cluster succeed: clusterID: %v, taskID: %v", resp.ClusterID, resp.TaskID)
}
