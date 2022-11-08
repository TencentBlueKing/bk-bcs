package clustercredential

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// Get 获取集群
func (c *ClusterCredentialMgr) Update(req manager.UpdateClusterCredentialReq) (err error) {
	resp, err := c.client.UpdateClusterCredential(c.ctx, &clustermanager.UpdateClusterCredentialReq{
		ClusterID:     req.ClusterID,
		ClientModule:  req.ClientModule,
		ServerAddress: req.ServerAddress,
		CaCertData:    req.CaCertData,
		UserToken:     req.UserToken,
		ClusterDomain: req.ClusterDomain,
	})
	if err != nil {
		return
	}

	if resp != nil && resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return
}
