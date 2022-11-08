package node

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *NodeMgr) Get(req manager.GetNodeReq) (resp manager.GetNodeResp, err error) {
	servResp, err := c.client.GetNode(c.ctx, &clustermanager.GetNodeRequest{
		InnerIP: req.InnerIP,
	})
	if err != nil {
		return
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	for _, v := range servResp.Data {
		resp.Data = append(resp.Data, &manager.Node{
			NodeID:       v.NodeID,
			InnerIP:      v.InnerIP,
			InstanceType: v.InstanceType,
			CPU:          v.CPU,
			Mem:          v.Mem,
			GPU:          v.GPU,
			Status:       v.Status,
			ZoneID:       v.ZoneID,
			NodeGroupID:  v.NodeGroupID,
			ClusterID:    v.ClusterID,
			VPC:          v.VPC,
			Region:       v.Region,
			Passwd:       v.Passwd,
			Zone:         v.Zone,
			DeviceID:     v.DeviceID,
		})
	}

	return
}
