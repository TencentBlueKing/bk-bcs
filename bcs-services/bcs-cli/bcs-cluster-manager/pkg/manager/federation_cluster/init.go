package federationcluster

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *FederationClusterMgr) Init(req manager.InitFederationClusterReq) error {
	_, err := c.client.InitFederationCluster(c.ctx, &clustermanager.InitFederationClusterReq{})
	if err != nil {
		return err
	}

	return nil
}
