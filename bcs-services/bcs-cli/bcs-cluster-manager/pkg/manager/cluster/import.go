package cluster

import "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"

// Import 导入用户集群(支持多云集群导入功能: 集群ID/kubeConfig)
func (c *ClusterMgr) Import(manager.ImportClusterReq) error {
	return nil
}
