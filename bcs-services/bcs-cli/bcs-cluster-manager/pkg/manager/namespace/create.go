package namespace

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *NamespaceMgr) Create(req manager.CreateNamespaceReq) (err error) {
	resp, err := c.client.CreateNamespace(c.ctx, &clustermanager.CreateNamespaceReq{
		Name:                req.Name,
		FederationClusterID: req.FederationClusterID,
		ProjectID:           req.ProjectID,
		BusinessID:          req.BusinessID,
		Labels:              req.Labels,
	})
	if err != nil {
		return
	}

	if resp != nil && resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return
}
