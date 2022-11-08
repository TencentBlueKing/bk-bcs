package node

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *NodeMgr) Update(req manager.UpdateNodeReq) error {
	resp, err := c.client.UpdateNode(c.ctx, &clustermanager.UpdateNodeRequest{
		InnerIPs:    req.InnerIPs,
		Status:      req.Status,
		NodeGroupID: req.NodeGroupID,
		ClusterID:   req.ClusterID,
		Updater:     "bcs",
	})
	if err != nil {
		return err
	}

	if resp != nil && resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return nil
}
