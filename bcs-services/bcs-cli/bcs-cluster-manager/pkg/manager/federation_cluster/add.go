package federationcluster

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *FederationClusterMgr) Add(req manager.AddFederatedClusterReq) error {
	resp, err := c.client.AddFederatedCluster(c.ctx, &clustermanager.AddFederatedClusterReq{
		FederationClusterID: req.FederationClusterID,
		ClusterID:           req.ClusterID,
	})
	if err != nil {
		return err
	}

	if resp != nil && resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return nil
}
