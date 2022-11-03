package node

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *NodeMgr) Cordon(req manager.CordonNodeReq) (resp manager.CordonNodeResp, err error) {
	servResp, err := c.client.CordonNode(c.ctx, &clustermanager.CordonNodeRequest{
		InnerIPs:  req.InnerIPs,
		ClusterID: req.ClusterID,
	})
	if err != nil {
		return
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	resp.Data = servResp.Fail

	return
}
