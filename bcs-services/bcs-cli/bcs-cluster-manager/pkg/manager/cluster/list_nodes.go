package cluster

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// ListNodes 查询集群下所有节点列表
func (c *ClusterMgr) ListNodes(req manager.ListClusterNodesReq) (resp manager.ListClusterNodesResp, err error) {
	servResp, err := c.client.ListNodesInCluster(c.ctx, &clustermanager.ListNodesInClusterRequest{
		Offset: req.Offset,
		Limit:  req.Limit,
	})
	if err != nil {
		return
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	for _, v := range servResp.Data {
		resp.Data = append(resp.Data, &manager.ClusterNode{
			NodeID:  v.NodeID,
			InnerIP: v.InnerIP,
		})
	}

	return
}
