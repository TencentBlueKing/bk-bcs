package namespace

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *NamespaceMgr) Update(req manager.UpdateNamespaceReq) (err error) {
	resp, err := c.client.UpdateNamespace(c.ctx, &clustermanager.UpdateNamespaceReq{
		Name:                req.Name,
		FederationClusterID: req.FederationClusterID,
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
