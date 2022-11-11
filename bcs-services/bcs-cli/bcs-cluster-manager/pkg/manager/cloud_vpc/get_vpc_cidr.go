package cloudvpc

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *CloudVPCMgr) GetVPCCidr(req manager.GetVPCCidrReq) (resp manager.GetVPCCidrResp, err error) {
	servResp, err := c.client.GetVPCCidr(c.ctx, &clustermanager.GetVPCCidrRequest{
		VpcID: req.VPCID,
	})
	if err != nil {
		return
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	for _, v := range servResp.Data {
		resp.Data = append(resp.Data, &manager.VPCCidr{
			VPC:      v.Vpc,
			Cidr:     v.Cidr,
			IPNumber: v.IPNumber,
			Status:   v.Status,
		})
	}

	return
}
