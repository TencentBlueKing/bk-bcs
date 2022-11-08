package clustercredential

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *ClusterCredentialMgr) List(req manager.ListClusterCredentialReq) (resp manager.ListClusterCredentialResp, err error) {
	servResp, err := c.client.ListClusterCredential(c.ctx, &clustermanager.ListClusterCredentialReq{
		Offset: req.Offset,
		Limit:  req.Limit,
	})
	if err != nil {
		return
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	for _, v := range servResp.Data {
		resp.Data = append(resp.Data, &manager.ClusterCredential{
			ServerKey:     v.ServerKey,
			ClusterID:     v.ClusterID,
			ClientModule:  v.ClientModule,
			ServerAddress: v.ServerAddress,
			CaCertData:    v.CaCertData,
			UserToken:     v.UserToken,
			ClusterDomain: v.ClusterDomain,
			ConnectMode:   v.ConnectMode,
			CreateTime:    v.CreateTime,
			UpdateTime:    v.UpdateTime,
			ClientCert:    v.ClientCert,
			ClientKey:     v.ClientKey,
		})
	}

	return
}
