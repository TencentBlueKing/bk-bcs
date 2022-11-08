package clustercredential

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// Get 获取集群
func (c *ClusterCredentialMgr) Get(req manager.GetClusterCredentialReq) (resp manager.GetClusterCredentialResp, err error) {
	servResp, err := c.client.GetClusterCredential(c.ctx, &clustermanager.GetClusterCredentialReq{
		ServerKey: req.ServerKey,
	})
	if err != nil {
		return
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	resp = manager.GetClusterCredentialResp{
		Data: manager.ClusterCredential{
			ServerKey:     servResp.Data.ServerKey,
			ClusterID:     servResp.Data.ClusterID,
			ClientModule:  servResp.Data.ClientModule,
			ServerAddress: servResp.Data.ServerAddress,
			CaCertData:    servResp.Data.CaCertData,
			UserToken:     servResp.Data.UserToken,
			ClusterDomain: servResp.Data.ClusterDomain,
			ConnectMode:   servResp.Data.ConnectMode,
			CreateTime:    servResp.Data.CreateTime,
			UpdateTime:    servResp.Data.UpdateTime,
			ClientCert:    servResp.Data.ClientCert,
			ClientKey:     servResp.Data.ClientKey,
		},
	}

	return
}
