package cluster

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// Get 获取集群
func (c *ClusterMgr) Get(req manager.GetClusterReq) (resp manager.GetClusterResp, err error) {
	servResp, err := c.client.GetCluster(c.ctx, &clustermanager.GetClusterReq{ClusterID: req.ClusterID})
	if err != nil {
		return
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	resp = manager.GetClusterResp{
		Data: manager.Cluster{
			ClusterID:   servResp.Data.ClusterID,
			ProjectID:   servResp.Data.ProjectID,
			BusinessID:  servResp.Data.BusinessID,
			EngineType:  servResp.Data.EngineType,
			IsExclusive: servResp.Data.IsExclusive,
			ClusterType: servResp.Data.ClusterType,
			Creator:     servResp.Data.Creator,
			Updater:     servResp.Data.Updater,
			ManageType:  servResp.Data.ManageType,
			ClusterName: servResp.Data.ClusterName,
			Environment: servResp.Data.Environment,
			Provider:    servResp.Data.Provider,
			Description: servResp.Data.Description,
			ClusterBasicSettings: manager.ClusterBasicSettings{
				Version: servResp.Data.ClusterBasicSettings.Version,
			},
			NetworkType: servResp.Data.NetworkType,
			Region:      servResp.Data.Region,
			VpcID:       servResp.Data.VpcID,
			NetworkSettings: manager.NetworkSettings{
				CidrStep:      servResp.Data.NetworkSettings.CidrStep,
				MaxNodePodNum: servResp.Data.NetworkSettings.MaxNodePodNum,
				MaxServiceNum: servResp.Data.NetworkSettings.MaxServiceNum,
			},
		},
	}

	return
}
