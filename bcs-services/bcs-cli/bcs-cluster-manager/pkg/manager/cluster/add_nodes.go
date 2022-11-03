package cluster

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// AddNodes 添加节点到集群
func (c *ClusterMgr) AddNodes(req manager.AddNodesClusterReq) (resp manager.AddNodesClusterResp, err error) {
	servResp, err := c.client.AddNodesToCluster(c.ctx, &clustermanager.AddNodesRequest{
		ClusterID:         req.ClusterID,
		Nodes:             req.Nodes,
		InitLoginPassword: "123456",
		NodeGroupID:       "",
		OnlyCreateInfo:    false,
		Operator:          "bcs",
	})
	if err != nil {
		return
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	resp.TaskID = servResp.Data.TaskID

	return
}
