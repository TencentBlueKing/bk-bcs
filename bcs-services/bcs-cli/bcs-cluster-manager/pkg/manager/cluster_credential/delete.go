package clustercredential

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// Get 获取集群
func (c *ClusterCredentialMgr) Delete(req manager.DeleteClusterCredentialReq) (err error) {
	resp, err := c.client.DeleteClusterCredential(c.ctx, &clustermanager.DeleteClusterCredentialReq{
		ServerKey: req.ServerKey,
	})
	if err != nil {
		return
	}

	if resp != nil && resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return
}
