package cluster

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// Retry 重试创建集群
func (c *ClusterMgr) Retry(req manager.RetryClusterReq) (resp manager.RetryClusterResp, err error) {
	servResp, err := c.client.RetryCreateClusterTask(c.ctx, &clustermanager.RetryCreateClusterReq{
		ClusterID: req.ClusterID,
		Operator:  "bcs",
	})
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
