package namespacequota

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *NamespaceQuotaMgr) Create(req manager.CreateNamespaceQuotaReq) (resp manager.CreateNamespaceQuotaResp, err error) {
	servResp, err := c.client.CreateNamespaceQuota(c.ctx, &clustermanager.CreateNamespaceQuotaReq{
		Namespace:           req.Namespace,
		FederationClusterID: req.FederationClusterID,
		ResourceQuota:       req.ResourceQuota,
	})
	if err != nil {
		return
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	resp.ClusterID = servResp.Data.ClusterID

	return
}
