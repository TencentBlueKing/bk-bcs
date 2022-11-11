package cloudvpc

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *CloudVPCMgr) Create(req manager.CreateCloudVPCReq) (err error) {
	resp, err := c.client.CreateCloudVPC(c.ctx, &clustermanager.CreateCloudVPCRequest{
		CloudID:     req.CloudID,
		NetworkType: req.NetworkType,
		Region:      req.Region,
		VpcName:     req.VPCName,
		VpcID:       req.VPCID,
		Creator:     "bcs",
	})
	if err != nil {
		return
	}

	if resp != nil && resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return
}
