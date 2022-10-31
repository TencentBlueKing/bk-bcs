package cluster

import (
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// DeleteNodes 从集群中删除节点
func (c *ClusterMgr) DeleteNodes(req manager.DeleteNodesClusterReq) (manager.DeleteNodesClusterResp, error) {
	c.client.DeleteNodesFromCluster(c.ctx, &clustermanager.DeleteNodesRequest{
		ClusterID:      req.ClusterID,
		Nodes:          strings.Join(req.Nodes, ","),
		DeleteMode:     "",
		IsForce:        false,
		Operator:       "bcs",
		OnlyDeleteInfo: false,
	})
	return manager.DeleteNodesClusterResp{}, nil
}
