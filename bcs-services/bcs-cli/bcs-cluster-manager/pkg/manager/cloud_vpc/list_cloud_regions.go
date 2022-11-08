package cloudvpc

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *CloudVPCMgr) ListCloudRegions(req manager.ListCloudRegionsReq) (resp manager.ListCloudRegionsResp, err error) {
	servResp, err := c.client.ListCloudRegions(c.ctx, &clustermanager.ListCloudRegionsRequest{
		CloudID: req.CloudID,
	})
	if err != nil {
		return
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	for _, v := range servResp.Data {
		resp.Data = append(resp.Data, &manager.CloudRegion{
			CloudID:    v.CloudID,
			Region:     v.Region,
			RegionName: v.RegionName,
		})
	}

	return
}
