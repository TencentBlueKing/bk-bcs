package cloudvpc

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *CloudVPCMgr) List() (resp manager.ListCloudVPCResp, err error) {
	servResp, err := c.client.ListCloudVPC(c.ctx, &clustermanager.ListCloudVPCRequest{})
	if err != nil {
		return
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	for _, v := range servResp.Data {
		resp.Data = append(resp.Data, &manager.CloudVPC{
			CloudID:     v.CloudID,
			Region:      v.Region,
			RegionName:  v.RegionName,
			NetworkType: v.NetworkType,
			VPCID:       v.VpcID,
			VPCName:     v.VpcName,
			Available:   v.Available,
			Extra:       v.Extra,
			Creator:     v.Creator,
			Updater:     v.Updater,
			CreatTime:   v.CreatTime,
			UpdateTime:  v.UpdateTime,
		})
	}

	return
}
