package namespacequota

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *NamespaceQuotaMgr) CreateNamespaceWithQuota(req manager.CreateNamespaceWithQuotaReq) (resp manager.CreateNamespaceWithQuotaResp, err error) {
	servResp, err := c.client.CreateNamespaceWithQuota(c.ctx, &clustermanager.CreateNamespaceWithQuotaReq{
		Name:                req.Name,
		FederationClusterID: req.FederationClusterID,
		ProjectID:           req.ProjectID,
		BusinessID:          req.BusinessID,
		Labels:              req.Labels,
		Region:              req.Region,
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
