package cluster

import "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"

// DeleteNodes 从集群中删除节点
func (c *ClusterMgr) DeleteNodes(manager.DeleteNodesClusterReq) (manager.DeleteNodesClusterResp, error) {
	return manager.DeleteNodesClusterResp{}, nil
}
