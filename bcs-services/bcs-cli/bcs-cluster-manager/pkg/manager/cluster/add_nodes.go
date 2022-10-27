package cluster

import "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"

// AddNodes 添加节点到集群
func (c *ClusterMgr) AddNodes(manager.AddNodesClusterReq) (manager.AddNodesClusterResp, error) {
	return manager.AddNodesClusterResp{}, nil
}
