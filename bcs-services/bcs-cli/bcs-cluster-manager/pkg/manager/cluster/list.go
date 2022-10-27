package cluster

import "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"

// List 获取集群列表
func (c *ClusterMgr) List(manager.ListClusterReq) (manager.ListClusterResp, error) {
	return manager.ListClusterResp{}, nil
}
