package cluster

import "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"

// Retry 重试创建集群
func (c *ClusterMgr) Retry(manager.RetryClusterReq) (manager.RetryClusterResp, error) {
	return manager.RetryClusterResp{}, nil
}
