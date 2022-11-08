package cloudvpc

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *CloudVPCMgr) Delete(req manager.DeleteCloudVPCReq) (err error) {
	resp, err := c.client.DeleteCloudVPC(c.ctx, &clustermanager.DeleteCloudVPCRequest{
		CloudID: req.CloudID,
		VpcID:   req.VPCID,
	})
	if err != nil {
		return
	}

	if resp != nil && resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return
}
