package namespace

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *NamespaceMgr) Delete(req manager.DeleteNamespaceReq) (err error) {
	resp, err := c.client.DeleteNamespace(c.ctx, &clustermanager.DeleteNamespaceReq{
		Name:                req.Name,
		FederationClusterID: req.FederationClusterID,
		IsForced:            req.IsForced,
	})
	if err != nil {
		return
	}

	if resp != nil && resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return
}
