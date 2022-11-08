package node

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *NodeMgr) Drain(req manager.DrainNodeReq) (resp manager.DrainNodeResp, err error) {
	servResp, err := c.client.DrainNode(c.ctx, &clustermanager.DrainNodeRequest{
		InnerIPs:  req.InnerIPs,
		ClusterID: req.ClusterID,
		Updater:   "bcs",
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
