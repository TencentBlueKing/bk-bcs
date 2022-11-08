package namespacequota

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *NamespaceQuotaMgr) Delete(req manager.DeleteNamespaceQuotaReq) error {
	resp, err := c.client.DeleteNamespaceQuota(c.ctx, &clustermanager.DeleteNamespaceQuotaReq{
		Namespace:           req.Namespace,
		FederationClusterID: req.FederationClusterID,
		ClusterID:           req.ClusterID,
		IsForced:            false,
	})
	if err != nil {
		return err
	}

	if resp != nil && resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return nil
}
