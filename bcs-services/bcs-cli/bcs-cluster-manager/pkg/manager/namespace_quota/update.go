package namespacequota

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *NamespaceQuotaMgr) Update(req manager.UpdateNamespaceQuotaReq) error {
	resp, err := c.client.UpdateNamespaceQuota(c.ctx, &clustermanager.UpdateNamespaceQuotaReq{
		Namespace:           req.Namespace,
		FederationClusterID: req.FederationClusterID,
		ClusterID:           req.ClusterID,
		ResourceQuota:       req.ResourceQuota,
	})
	if err != nil {
		return err
	}

	if resp != nil && resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return nil
}
