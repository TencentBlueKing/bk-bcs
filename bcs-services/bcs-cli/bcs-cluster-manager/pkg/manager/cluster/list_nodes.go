package cluster

import "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"

// ListNodes 查询集群下所有节点列表
func (c *ClusterMgr) ListNodes(manager.ListClusterNodesReq) (manager.ListClusterNodesResp, error) {
	return manager.ListClusterNodesResp{}, nil
}
