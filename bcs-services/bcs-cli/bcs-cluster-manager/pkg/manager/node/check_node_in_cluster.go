package node

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *NodeMgr) CheckNodeInCluster(req manager.CheckNodeInClusterReq) (resp manager.CheckNodeInClusterResp, err error) {
	servResp, err := c.client.CheckNodeInCluster(c.ctx, &clustermanager.CheckNodesRequest{
		InnerIPs: req.InnerIPs,
	})
	if err != nil {
		return
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	for k, v := range servResp.Data {
		resp.Data[k] = manager.NodeResult{
			IsExist:     v.IsExist,
			ClusterID:   v.ClusterID,
			ClusterName: v.ClusterName,
		}
	}

	return
}
