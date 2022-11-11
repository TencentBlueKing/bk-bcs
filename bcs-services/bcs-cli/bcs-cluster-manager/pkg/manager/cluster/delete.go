package cluster

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// Delete 删除集群
func (c *ClusterMgr) Delete(req manager.DeleteClusterReq) (resp manager.DeleteClusterResp, err error) {
	servResp, err := c.client.DeleteCluster(c.ctx, &clustermanager.DeleteClusterReq{ClusterID: req.ClusterID})
	if err != nil {
		return
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	resp.ClusterID = servResp.Data.ClusterID
	resp.TaskID = servResp.Task.TaskID

	return
}
