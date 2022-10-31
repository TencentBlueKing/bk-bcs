package cluster

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// Delete 删除集群
func (c *ClusterMgr) Delete(req manager.DeleteClusterReq) error {
	resp, err := c.client.DeleteCluster(c.ctx, &clustermanager.DeleteClusterReq{ClusterID: req.ClusterID})
	if err != nil {
		return err
	}

	if resp != nil && resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return nil
}
