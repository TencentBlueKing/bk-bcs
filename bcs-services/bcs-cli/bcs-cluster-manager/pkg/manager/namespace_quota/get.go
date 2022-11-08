package namespacequota

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *NamespaceQuotaMgr) Get(req manager.GetNamespaceQuotaReq) (resp manager.GetNamespaceQuotaResp, err error) {
	servResp, err := c.client.GetNamespaceQuota(c.ctx, &clustermanager.GetNamespaceQuotaReq{
		Namespace:           req.Namespace,
		FederationClusterID: req.FederationClusterID,
		ClusterID:           req.ClusterID,
	})
	if err != nil {
		return
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	resp = manager.GetNamespaceQuotaResp{
		Data: manager.ResourceQuota{
			Namespace:           servResp.Data.Namespace,
			FederationClusterID: servResp.Data.FederationClusterID,
			ClusterID:           servResp.Data.ClusterID,
			ResourceQuota:       servResp.Data.ResourceQuota,
			Region:              servResp.Data.Region,
			CreateTime:          servResp.Data.CreateTime,
			UpdateTime:          servResp.Data.UpdateTime,
			Status:              servResp.Data.Status,
			Message:             servResp.Data.Message,
		},
	}

	return
}
