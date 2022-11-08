package namespace

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *NamespaceMgr) List(req manager.ListNamespaceReq) (resp manager.ListNamespaceResp, err error) {
	servResp, err := c.client.ListNamespace(c.ctx, &clustermanager.ListNamespaceReq{
		FederationClusterID: req.FederationClusterID,
		ProjectID:           req.ProjectID,
		BusinessID:          req.BusinessID,
		Offset:              req.Offset,
		Limit:               req.Limit,
	})
	if err != nil {
		return
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	for _, v := range servResp.Data {
		resp.Data = append(resp.Data, &manager.Namespace{
			Name:                v.Name,
			FederationClusterID: v.FederationClusterID,
			ProjectID:           v.ProjectID,
			BusinessID:          v.BusinessID,
			Labels:              v.Labels,
		})
	}

	return
}
