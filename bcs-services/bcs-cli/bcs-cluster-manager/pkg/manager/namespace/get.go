package namespace

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *NamespaceMgr) Get(req manager.GetNamespaceReq) (resp manager.GetNamespaceResp, err error) {
	servResp, err := c.client.GetNamespace(c.ctx, &clustermanager.GetNamespaceReq{
		Name:                req.Name,
		FederationClusterID: req.FederationClusterID,
	})
	if err != nil {
		return
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	resp = manager.GetNamespaceResp{
		Data: &manager.Namespace{
			Name:                servResp.Data.Name,
			FederationClusterID: servResp.Data.FederationClusterID,
			ProjectID:           servResp.Data.ProjectID,
			BusinessID:          servResp.Data.BusinessID,
			Labels:              servResp.Data.Labels,
		},
	}

	return
}
