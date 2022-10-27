package cluster

import "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"

// ListCommon 查询公共集群及公共集群所属权限
func (c *ClusterMgr) ListCommon(manager.ListCommonClusterReq) (manager.ListCommonClusterResp, error) {
	return manager.ListCommonClusterResp{}, nil
}
