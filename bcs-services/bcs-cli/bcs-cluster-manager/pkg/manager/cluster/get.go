package cluster

import "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"

// Get 获取集群
func (c *ClusterMgr) Get(manager.GetClusterReq) (manager.GetClusterResp, error) {
	return manager.GetClusterResp{}, nil
}
